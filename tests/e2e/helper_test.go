package e2e

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/playwright-community/playwright-go"

	"gitlab.com/hmajid2301/banterbus/internal/banterbustest"
	"gitlab.com/hmajid2301/banterbus/internal/service"
	"gitlab.com/hmajid2301/banterbus/internal/store"
	"gitlab.com/hmajid2301/banterbus/internal/transport/websockets"
)

var (
	pw      *playwright.Playwright
	browser playwright.Browser
	//  TODO: set this vars
	// browsercontext playwright.BrowserContext
	// page           playwright.Page
	expect        playwright.PlaywrightAssertions
	headless      = os.Getenv("HEADFUL") == ""
	isChromium    bool
	isFirefox     bool
	isWebKit      bool
	browserName   = getBrowserName()
	browserType   playwright.BrowserType
	serverAddress string
)

// default context options for most tests
var DEFAULT_CONTEXT_OPTIONS = playwright.BrowserNewContextOptions{
	AcceptDownloads: playwright.Bool(true),
	HasTouch:        playwright.Bool(true),
}

func TestMain(m *testing.M) {
	server := BeforeAll()
	code := m.Run()
	AfterAll(server)
	os.Exit(code)
}

func BeforeAll() *httptest.Server {
	var err error
	pw, err = playwright.Run()
	if err != nil {
		log.Fatalf("could not start Playwright: %v", err)
	}
	if browserName == "chromium" || browserName == "" {
		browserType = pw.Chromium
	} else if browserName == "firefox" {
		browserType = pw.Firefox
	} else if browserName == "webkit" {
		browserType = pw.WebKit
	}

	execPath := "/nix/store/7c9hmhxkkqxcvikz6g2zcy1gvkkg8zg3-playwright-browsers-chromium/chromium-1129/chrome-linux/chrome"
	// launch browser, headless or not depending on HEADFUL env
	browser, err = browserType.Launch(playwright.BrowserTypeLaunchOptions{
		Headless:       playwright.Bool(headless),
		ExecutablePath: &execPath,
	})
	if err != nil {
		log.Fatalf("could not launch: %v", err)
	}
	// init web-first assertions with 1s timeout instead of default 5s
	expect = playwright.NewPlaywrightAssertions(1000)
	isChromium = browserName == "chromium" || browserName == ""
	isFirefox = browserName == "firefox"
	isWebKit = browserName == "webkit"

	server, err := newTestServer()
	if err != nil {
		log.Fatalf("could not start server: %v", err)
	}
	return server
}

// AfterAll does cleanup, e.g. stop playwright driver
func AfterAll(server *httptest.Server) {
	server.Close()
	if err := pw.Stop(); err != nil {
		log.Fatalf("could not start Playwright: %v", err)
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
	db, err := banterbustest.CreateDB(ctx)
	if err != nil {
		return nil, err
	}

	myStore, err := store.NewStore(db)
	if err != nil {
		return nil, err
	}

	userRandomizer := service.NewUserRandomizer()
	roomServicer := service.NewRoomService(myStore, userRandomizer)
	playerServicer := service.NewPlayerService(myStore, userRandomizer)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ws" {
			subscriber := websockets.NewSubscriber(roomServicer, playerServicer, logger)
			_ = subscriber.Subscribe(context.Background(), r, w)
		} else if r.URL.Path == "/" {
			http.ServeFile(w, r, "../../static")
		} else {
			http.NotFound(w, r)

		}
	}))

	serverAddress = server.Listener.Addr().String()
	return server, nil

}
