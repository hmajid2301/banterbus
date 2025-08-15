package e2e

import (
	"log"
	"os"
	"testing"

	"github.com/mdobak/go-xerrors"
	"github.com/playwright-community/playwright-go"
)

var (
	expect    playwright.PlaywrightAssertions
	browser   playwright.Browser
	pw        *playwright.Playwright
	webappURL = os.Getenv("BANTERBUS_PLAYWRIGHT_URL")
)

func TestMain(m *testing.M) {
	code := 1
	defer func() {
		afterAll()
		os.Exit(code)
	}()

	if err := beforeAll(); err != nil {
		log.Fatalf("could not start server: %v", err)
	}

	code = m.Run()
}

func beforeAll() error {
	var err error
	pw, err = playwright.Run()
	if err != nil {
		return xerrors.New("could not start Playwright: %v", err)
	}

	browserName := os.Getenv("BROWSER")
	if browserName == "" {
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

	headless := os.Getenv("BANTERBUS_PLAYWRIGHT_HEADLESS") == "true"
	browser, err = browserType.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(headless),
	})
	if err != nil {
		return xerrors.New("could not start browser: %v", err)
	}

	expect = playwright.NewPlaywrightAssertions(1000)

	// Set webappURL from environment or default to localhost
	if webappURL == "" {
		webappURL = "http://localhost:8080"
	}

	return nil
}

func afterAll() {
	if browser != nil {
		if err := browser.Close(); err != nil {
			log.Printf("Browser close error: %v", err)
		}
	}
	if pw != nil {
		if err := pw.Stop(); err != nil {
			log.Printf("Playwright stop error: %v", err)
		}
	}
}

func setupTest(t *testing.T, playerNum int) ([]playwright.Page, error) {
	pages := []playwright.Page{}

	for range playerNum {
		context, err := browser.NewContext(playwright.BrowserNewContextOptions{
			RecordVideo: &playwright.RecordVideo{Dir: "videos"},
			Viewport: &playwright.Size{
				Width:  960,
				Height: 1280,
			},
			Permissions: []string{"clipboard-read", "clipboard-write"},
		})
		if err != nil {
			return nil, xerrors.New("context creation failed: %v", err)
		}

		page, err := context.NewPage()
		if err != nil {
			return nil, xerrors.New("page creation failed: %v", err)
		}

		_, err = page.Goto(webappURL)
		if err != nil {
			return nil, xerrors.New("failed to go to URL: %v", err)
		}
		pages = append(pages, page)
		cleanup := func() {
			if err := page.Close(); err != nil {
				log.Printf("Page close error: %v", err)
			}
			if err := context.Close(); err != nil {
				log.Printf("Context close error: %v", err)

			}
		}
		t.Cleanup(cleanup)

	}

	return pages, nil
}

func setupTestMultiple(playerNum int) ([]playwright.Page, func(pages []playwright.Page) error) {
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
