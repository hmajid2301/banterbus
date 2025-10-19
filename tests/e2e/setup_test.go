package e2e

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

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
		return fmt.Errorf("could not start Playwright: %v", err)
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

	headless := true
	if os.Getenv("BANTERBUS_PLAYWRIGHT_SHOW_BROWSER") != "" {
		headless = false
	}

	browser, err = browserType.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(headless),
		Args: []string{
			"--no-sandbox",
			"--disable-setuid-sandbox",
			"--disable-dev-shm-usage",
			"--disable-background-timer-throttling",
			"--disable-backgrounding-occluded-windows",
			"--disable-renderer-backgrounding",
			"--disable-web-security",
			"--disable-features=VizDisplayCompositor",
			"--disable-gpu",
		},
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		return fmt.Errorf("could not start browser: %v", err)
	}

	expect = playwright.NewPlaywrightAssertions(10000)

	if webappURL == "" {
		webappURL = "http://localhost:8081"
	}

	return nil
}

func sanitizeTestName(testName string) string {
	reg := regexp.MustCompile(`[<>:"/\\|?*]`)
	sanitized := reg.ReplaceAllString(testName, "_")

	sanitized = strings.ReplaceAll(sanitized, " ", "_")

	if len(sanitized) > 200 {
		sanitized = sanitized[:200]
	}

	return sanitized
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
	contexts := []playwright.BrowserContext{}

	testName := sanitizeTestName(t.Name())

	for i := range playerNum {
		if i > 0 {
			time.Sleep(100 * time.Millisecond)
		}
		tempVideoDir := filepath.Join("videos", "temp", fmt.Sprintf("%s_%d", testName, i))
		if err := os.MkdirAll(tempVideoDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create temp video dir: %w", err)
		}

		context, err := browser.NewContext(playwright.BrowserNewContextOptions{
			RecordVideo:       &playwright.RecordVideo{Dir: tempVideoDir},
			Permissions:       []string{"clipboard-read", "clipboard-write"},
			IgnoreHttpsErrors: playwright.Bool(true),
			ColorScheme:       playwright.ColorSchemeLight,
		})
		if err != nil {
			return nil, fmt.Errorf("context creation failed: %w", err)
		}
		contexts = append(contexts, context)

		page, err := context.NewPage()
		if err != nil {
			return nil, fmt.Errorf("page creation failed: %w", err)
		}

		page.SetDefaultTimeout(8000)

		var navErr error
		for retry := 0; retry < 2; retry++ {
			_, navErr = page.Goto(webappURL, playwright.PageGotoOptions{
				WaitUntil: playwright.WaitUntilStateDomcontentloaded,
				Timeout:   playwright.Float(5000),
			})
			if navErr == nil {
				break
			}
			if retry < 1 {
				time.Sleep(500 * time.Millisecond)
			}
		}
		if navErr != nil {
			return nil, fmt.Errorf("failed to go to URL after retries: %w", navErr)
		}

		_, err = page.WaitForSelector("input[name='player_nickname']", playwright.PageWaitForSelectorOptions{
			Timeout: playwright.Float(3000),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to wait for page elements: %w", err)
		}

		_, err = page.WaitForFunction("() => window.htmx && window.htmx.trigger", playwright.PageWaitForFunctionOptions{
			Timeout: playwright.Float(3000),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to wait for htmx/websocket setup: %w", err)
		}

		testNameLocator := page.Locator(`input[name="test_name"]`)
		count, err := testNameLocator.Count()
		if err == nil && count > 0 {
			err = testNameLocator.Fill(testName)
			if err != nil {
				log.Printf("Warning: failed to set test name: %v", err)
			}
		}

		pages = append(pages, page)
	}

	t.Cleanup(func() {
		statusDir := "passed"
		if t.Failed() {
			statusDir = "failed"
		}
		finalBaseDir := filepath.Join("videos", statusDir)

		for i := range contexts {
			tempVideoDir := filepath.Join("videos", "temp", fmt.Sprintf("%s_%d", testName, i))
			finalVideoPath := filepath.Join(finalBaseDir, fmt.Sprintf("%s_player_%d.webm", testName, i+1))

			files, err := filepath.Glob(filepath.Join(tempVideoDir, "*.webm"))
			if err != nil {
				log.Printf("Error finding video files: %v", err)
				continue
			}

			if len(files) > 0 {
				if err := os.MkdirAll(finalBaseDir, 0755); err != nil {
					log.Printf("Error creating final video directory: %v", err)
					continue
				}

				if err := os.Rename(files[0], finalVideoPath); err != nil {
					log.Printf("Error renaming video file: %v", err)
				} else {
					log.Printf("Video saved as: %s", finalVideoPath)
				}
			}

			if err := os.RemoveAll(tempVideoDir); err != nil {
				log.Printf("Error removing temp directory: %v", err)
			}
		}
	})

	return pages, nil
}
