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
	expect    playwright.PlaywrightAssertions
	browser   playwright.Browser
	webappURL = os.Getenv("BANTERBUS_PLAYWRIGHT_URL")
)

func TestMain(m *testing.M) {
	pw, server, err := beforeAll()
	if err != nil {
		log.Fatalf("could not start server: %v", err)
	}

	code := m.Run()
	afterAll(pw, server)
	os.Exit(code)
}

func beforeAll() (*playwright.Playwright, *httptest.Server, error) {
	var err error
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not start Playwright: %v", err)
	}

	browserName, hasEnv := os.LookupEnv("BROWSER")
	if !hasEnv {
		browserName = "chromium"
	}

	var browserType playwright.BrowserType

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

	headless := os.Getenv("BANTERBUS_PLAYWRIGHT_HEADLESS") == ""
	browser, err = browserType.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(headless),
	})
	if err != nil {
		return &playwright.Playwright{}, nil, xerrors.New("could not start browser: %v", err)
	}

	expect = playwright.NewPlaywrightAssertions(1000)

	// INFO: if no address passed start local server
	var server *httptest.Server
	if webappURL == "" {
		server, err = banterbustest.NewTestServer()
		webappURL = server.Listener.Addr().String()
		if err != nil {
			return &playwright.Playwright{}, nil, err
		}
	}

	return pw, server, nil
}

func afterAll(pw *playwright.Playwright, server *httptest.Server) {
	if server != nil {
		server.Close()
	}

	if err := pw.Stop(); err != nil {
		log.Fatalf("could not stop Playwright: %v", err)
	}
}

func setupTest(playerNum int) ([]playwright.Page, func(pages []playwright.Page) error) {
	var err error

	contexts := make([]playwright.BrowserContext, playerNum)
	pages := make([]playwright.Page, playerNum)

	for i := 0; i < playerNum; i++ {
		contexts[i], err = browser.NewContext(playwright.BrowserNewContextOptions{
			RecordVideo: &playwright.RecordVideo{
				Dir: "videos/",
			},
			Permissions: []string{"clipboard-read", "clipboard-write"},
		})

		if err != nil {
			log.Fatalf("could not create a new browser context: %v", err)
		}
		page, err := contexts[i].NewPage()
		if err != nil {
			log.Fatalf("could not create page: %v", err)
		}

		_, err = page.Goto(webappURL)
		if err != nil {
			log.Fatalf("could not go to page: %v", err)
		}

		pages[i] = page
	}

	return pages, func(pages []playwright.Page) error {
		for _, page := range pages {
			err := page.Close()
			return err
		}
		return nil
	}
}
