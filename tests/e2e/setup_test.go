package e2e

// E2E Test Tracing Guide
// ======================
//
// This package includes built-in support for correlating E2E tests with server logs using trace IDs.
//
// How it works:
// 1. Frontend generates trace IDs and propagates them to backend via HTTP headers and WebSocket messages
// 2. In test/dev environments, trace IDs are displayed in the top-right corner of each page
// 3. E2E tests can extract these trace IDs and log them for easy correlation
// 4. All server logs include the trace_id field, making it easy to filter logs for a specific test
//
// Usage in tests:
//   func TestSomething(t *testing.T) {
//       pages, err := setupTest(t, 2)
//       require.NoError(t, err)
//
//       // Extract and log trace ID at key points for debugging
//       logTraceIDForTest(t, pages[0], "After page load")
//
//       // ... test actions ...
//
//       logTraceIDForTest(t, pages[0], "After game creation")
//   }
//
// Log correlation:
//   1. Run your E2E test and note the trace ID from the test logs
//   2. Search server logs for that trace ID to see all related server activity
//   3. Example: `grep "trace_id\":\"abc123..." server.log`

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

func extractTraceIDFromPage(page playwright.Page) (string, error) {
	traceElement := page.Locator("#current-trace-id")
	if count, err := traceElement.Count(); err == nil && count > 0 {
		if traceID, err := traceElement.InnerText(); err == nil && traceID != "" {
			return traceID, nil
		}
	}

	return "", fmt.Errorf("trace ID element not found or empty")
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

	// Setup cleanup that renames video files and extracts trace info
	t.Cleanup(func() {
		statusDir := "passed"
		if t.Failed() {
			statusDir = "failed"
		}
		finalBaseDir := filepath.Join("videos", statusDir)

		// Extract trace ID from the first page if available
		var traceID string
		if len(pages) > 0 {
			if id, err := extractTraceIDFromPage(pages[0]); err == nil {
				traceID = id
				log.Printf("Test %s completed with trace ID: %s", testName, traceID)
			} else {
				log.Printf("Failed to extract trace ID from page: %v", err)
			}
		}

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

func logTraceIDForTest(t *testing.T, page playwright.Page, action string) {
	if traceID, err := extractTraceIDFromPage(page); err == nil {
		t.Logf("%s - Test: %s, Trace ID: %s", action, t.Name(), traceID)
	}
}
