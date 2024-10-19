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
	transport "gitlab.com/hmajid2301/banterbus/internal/transport/http"
	"gitlab.com/hmajid2301/banterbus/internal/transport/websockets"
)

var (
	pw              *playwright.Playwright
	browser         playwright.Browser
	browserContexts []playwright.BrowserContext
	pages           []playwright.Page
	expect          playwright.PlaywrightAssertions
	headless        = os.Getenv("HEADFUL") == ""
	isChromium      bool
	isFirefox       bool
	isWebKit        bool
	browserName     = getBrowserName()
	browserType     playwright.BrowserType
	serverAddress   string
	testUserNum     = 2
)

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

	// launch browser, headless or not depending on HEADFUL env
	browser, err = browserType.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(headless),
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

	ResetBrowserContexts()
	return server
}

// AfterAll does cleanup, e.g. stop playwright driver
func AfterAll(server *httptest.Server) {
	server.Close()
	for i := 0; i < testUserNum; i++ {
		browserContexts[i].Close()
	}
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
	roomServicer := service.NewLobbyService(myStore, userRandomizer)
	playerServicer := service.NewPlayerService(myStore, userRandomizer)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	subscriber := websockets.NewSubscriber(roomServicer, playerServicer, logger)

	staticFS := http.Dir("../../static")
	srv := transport.NewServer(subscriber, logger, staticFS)
	server := httptest.NewServer(srv.Srv.Handler)

	serverAddress = server.Listener.Addr().String()
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
		})
		if err != nil {
			log.Fatalf("could not create a new browser context: %v", err)
		}
		page, err := browserContexts[i].NewPage()
		if err != nil {
			log.Fatalf("could not create page: %v", err)
		}

		_, err = page.Goto(serverAddress)
		if err != nil {
			log.Fatalf("could not go to page: %v", err)
		}

		pages[i] = page
	}
}
