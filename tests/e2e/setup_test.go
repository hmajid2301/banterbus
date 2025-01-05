package e2e

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/MicahParks/jwkset"
	"github.com/MicahParks/keyfunc/v3"
	"github.com/invopop/ctxi18n"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/mdobak/go-xerrors"
	"github.com/playwright-community/playwright-go"

	"gitlab.com/hmajid2301/banterbus/internal/banterbustest"
	"gitlab.com/hmajid2301/banterbus/internal/config"
	"gitlab.com/hmajid2301/banterbus/internal/service"
	"gitlab.com/hmajid2301/banterbus/internal/service/randomizer"
	"gitlab.com/hmajid2301/banterbus/internal/store/db"
	"gitlab.com/hmajid2301/banterbus/internal/store/pubsub"
	transporthttp "gitlab.com/hmajid2301/banterbus/internal/transport/http"
	"gitlab.com/hmajid2301/banterbus/internal/transport/websockets"
	"gitlab.com/hmajid2301/banterbus/internal/views"
)

var (
	pw              *playwright.Playwright
	browser         playwright.Browser
	browserContexts []playwright.BrowserContext
	pages           []playwright.Page
	expect          playwright.PlaywrightAssertions
	headless        = os.Getenv("BANTERBUS_PLAYWRIGHT_HEADLESS") == ""
	browserName     = getBrowserName()
	browserType     playwright.BrowserType
	webappURL       = os.Getenv("BANTERBUS_PLAYWRIGHT_URL")
	testUserNum     = 2
)

func TestMain(m *testing.M) {
	server, err := BeforeAll()
	if err != nil {
		log.Fatalf("could not start server: %v", err)
	}

	code := m.Run()
	AfterAll(server)
	os.Exit(code)
}

func BeforeAll() (*httptest.Server, error) {
	var err error
	pw, err = playwright.Run()
	if err != nil {
		log.Fatalf("could not start Playwright: %v", err)
	}
	switch browserName {
	case "chromium":
		browserType = pw.Chromium
	case "firefox":
		browserType = pw.Firefox
	case "webkit":
		browserType = pw.WebKit
	default:
		browserType = pw.Chromium
	}

	browser, err = browserType.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(headless),
	})
	if err != nil {
		return nil, xerrors.New("could not start browser: %v", err)
	}

	expect = playwright.NewPlaywrightAssertions(1000)

	// INFO: if no address passed start local server
	var server *httptest.Server
	if webappURL == "" {
		server, err = newTestServer()
		if err != nil {
			return nil, err
		}
	}

	ResetBrowserContexts()
	return server, nil
}

func AfterAll(server *httptest.Server) {
	if server != nil {
		server.Close()
	}

	for i := 0; i < testUserNum; i++ {
		browserContexts[i].Close()
	}
	if err := pw.Stop(); err != nil {
		log.Fatalf("could not stop Playwright: %v", err)
	}
}

func getBrowserName() string {
	browserName, hasEnv := os.LookupEnv("BROWSER")
	if hasEnv {
		return browserName
	}
	return "chromium"
}

func newTestServer() (*httptest.Server, error) {
	ctx := context.Background()
	pool, err := banterbustest.CreateDB(ctx)
	if err != nil {
		return nil, err
	}

	myStore, err := db.NewDB(pool)
	if err != nil {
		return nil, err
	}

	userRandomizer := randomizer.NewUserRandomizer()
	lobbyServicer := service.NewLobbyService(myStore, userRandomizer, "en-GB")
	playerServicer := service.NewPlayerService(myStore, userRandomizer)
	roundServicer := service.NewRoundService(myStore, userRandomizer, "en-GB")
	logger := setupLogger()

	redisAddr := os.Getenv("BANTERBUS_REDIS_ADDRESS")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisClient, err := pubsub.NewRedisClient(redisAddr)
	if err != nil {
		return nil, err
	}

	conf, err := config.LoadConfig(ctx)
	if err != nil {
		return nil, err
	}

	conf.Timings.ShowScoreScreenFor = time.Second * 2
	conf.App.AutoReconnect = false

	subscriber := websockets.NewSubscriber(lobbyServicer, playerServicer, roundServicer, logger, redisClient, conf)
	err = ctxi18n.LoadWithDefault(views.Locales, "en-GB")
	if err != nil {
		return nil, xerrors.New("error loading locales", err)
	}

	staticFS := http.Dir("../../static")
	serverConfig := transporthttp.ServerConfig{
		Host:          "localhost",
		Port:          8198,
		DefaultLocale: i18n.Code("en-GB"),
	}

	u, err := url.Parse(conf.JWT.JWSURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse jwks URL: %w", err)
	}

	storage, err := jwkset.NewStorageFromHTTP(u, jwkset.HTTPClientStorageOptions{Ctx: ctx})
	if err != nil {
		return nil, fmt.Errorf("failed to jwkset storage: %w", err)
	}

	k, err := keyfunc.New(keyfunc.Options{
		Storage: storage,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create keyfunc: %w", err)
	}
	srv := transporthttp.NewServer(subscriber, logger, staticFS, k.Keyfunc, serverConfig)
	server := httptest.NewServer(srv.Server.Handler)

	webappURL = server.Listener.Addr().String()
	return server, nil
}

func ResetBrowserContexts() {
	var err error

	browserContexts = make([]playwright.BrowserContext, testUserNum)
	pages = make([]playwright.Page, testUserNum)

	for i := 0; i < testUserNum; i++ {
		browserContexts[i], err = browser.NewContext(playwright.BrowserNewContextOptions{
			RecordVideo: &playwright.RecordVideo{
				Dir: "videos/",
			},
			Permissions: []string{"clipboard-read", "clipboard-write"},
		})

		if err != nil {
			log.Fatalf("could not create a new browser context: %v", err)
		}
		page, err := browserContexts[i].NewPage()
		if err != nil {
			log.Fatalf("could not create page: %v", err)
		}

		_, err = page.Goto(webappURL)
		if err != nil {
			log.Fatalf("could not go to page: %v", err)
		}

		pages[i] = page
	}
}

func setupLogger() *slog.Logger {
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "DEBUG"
	}

	var level slog.Level
	switch strings.ToUpper(logLevel) {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		log.Fatalf("unknown log level: %s", logLevel)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))

	if os.Getenv("LOG_DISABLED") == "true" {
		logger = slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
			Level: level,
		}))
	}

	return logger
}
