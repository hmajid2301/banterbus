package e2e

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

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

	headless := os.Getenv("BANTERBUS_PLAYWRIGHT_HEADLESS") == "true"
	browser, err = browserType.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(headless),
		Args: []string{
			"--no-sandbox",
			"--disable-setuid-sandbox",
			"--disable-dev-shm-usage",
			"--disable-background-timer-throttling",
			"--disable-backgrounding-occluded-windows",
			"--disable-renderer-backgrounding",
		},
	})
	if err != nil {
		return fmt.Errorf("could not start browser: %v", err)
	}

	expect = playwright.NewPlaywrightAssertions(120000) // 2 minutes for CI

	// Set webappURL from environment or default to localhost
	if webappURL == "" {
		webappURL = "http://localhost:8081"
	}

	return nil
}

// sanitizeTestName converts a test name to a filename-safe string
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
		tempVideoDir := filepath.Join("videos", "temp", fmt.Sprintf("%s_%d", testName, i))
		if err := os.MkdirAll(tempVideoDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create temp video dir: %v", err)
		}

		context, err := browser.NewContext(playwright.BrowserNewContextOptions{
			RecordVideo: &playwright.RecordVideo{Dir: tempVideoDir},
			Viewport: &playwright.Size{
				Width:  960,
				Height: 1280,
			},
			Permissions: []string{"clipboard-read", "clipboard-write"},
		})
		if err != nil {
			return nil, fmt.Errorf("context creation failed: %v", err)
		}
		contexts = append(contexts, context)

		page, err := context.NewPage()
		if err != nil {
			return nil, fmt.Errorf("page creation failed: %v", err)
		}

		// Set longer timeout for page operations - increased for CI
		page.SetDefaultTimeout(120000) // 2 minutes for CI

		// Navigate to the URL with extended timeout for CI
		_, err = page.Goto(webappURL, playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateDomcontentloaded,
			Timeout:   playwright.Float(120000), // 2 minutes timeout for CI
		})
		if err != nil {
			return nil, fmt.Errorf("failed to go to URL: %v", err)
		}

		// Wait for page to be fully loaded and interactive
		err = page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
			State:   playwright.LoadStateNetworkidle,
			Timeout: playwright.Float(60000), // 1 minute for network idle
		})
		if err != nil {
			return nil, fmt.Errorf("failed to wait for page load: %v", err)
		}
		pages = append(pages, page)
	}

	// Setup cleanup that renames video files
	t.Cleanup(func() {
		statusDir := "passed"
		if t.Failed() {
			statusDir = "failed"
		}
		finalBaseDir := filepath.Join("videos", statusDir)

		for i, context := range contexts {
			if err := context.Close(); err != nil {
				log.Printf("Context close error: %v", err)
			}

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
