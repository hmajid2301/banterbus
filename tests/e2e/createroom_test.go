package e2e

import (
	"log"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NOTE: is this an integration test? should i be testing from main.go
// Start the server then run the playwright tests? This is doing 90% of the work.
func TestE2ECreateRoom(t *testing.T) {
	t.Run("Should create room with a player nickname", func(t *testing.T) {
		page, err := browser.NewPage()
		require.NoError(t, err)

		// TODO: don't hardcode take the address
		// TODO: https://github.com/playwright-community/playwright-go/blob/v0.2000.0/tests/tracing_test.go
		_, err = page.Goto(serverAddress)
		require.NoError(t, err)

		err = page.Locator("#create_room_form").GetByPlaceholder("Enter your nickname here").Click()
		require.NoError(t, err)
		err = page.Locator("#create_room_form").GetByPlaceholder("Enter your nickname here").Fill("test_nickname")
		require.NoError(t, err)
		err = page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Create Room"}).Click()
		require.NoError(t, err)

		err = page.Locator("img").Click()
		require.NoError(t, err)

		count, err := page.GetByText("Code:").Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		if _, err = page.Screenshot(playwright.PageScreenshotOptions{
			Path: playwright.String("./foo.png"),
		}); err != nil {
			log.Fatalf("could not create screenshot: %v", err)
		}
	})
}
