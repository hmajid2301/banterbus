package e2e

import (
	"log"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/mdobak/go-xerrors"
	"github.com/playwright-community/playwright-go"

	"gitlab.com/hmajid2301/banterbus/internal/banterbustest"
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
		server, err = banterbustest.NewTestServer()
		webappURL = server.Listener.Addr().String()
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
