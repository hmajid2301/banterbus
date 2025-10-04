package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestE2EErrorHandling(t *testing.T) {
	t.Run("Should handle invalid room codes gracefully", func(t *testing.T) {
		playerPages, err := setupTest(t, 1)
		require.NoError(t, err)

		player := playerPages[0]

		// Try to join with valid nickname but invalid room code to test server-side validation
		err = player.Locator("input[name='player_nickname']").Fill("TestPlayer")
		require.NoError(t, err)

		err = player.Locator("input[name='room_code']").Fill("INVALID")
		require.NoError(t, err)

		err = player.GetByText("Join").Click()
		require.NoError(t, err)

		// Wait for websocket response and toast to appear
		err = expect.Locator(player.Locator("[role='alert']")).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
			Timeout: playwright.Float(8000),
		})
		require.NoError(t, err)

		// Should show validation error in toast notification (room not found)
		err = expect.Locator(player.Locator("[role='alert'] p")).
			ToContainText("Failed to join room", playwright.LocatorAssertionsToContainTextOptions{
				Timeout: playwright.Float(2000),
			})
		require.NoError(t, err)
	})

	t.Run("Should handle extremely long nicknames", func(t *testing.T) {
		playerPages, err := setupTest(t, 1)
		require.NoError(t, err)

		player := playerPages[0]

		// Try with extremely long nickname
		longNickname := "ThisIsAnExtremelyLongNicknameThantShouldBeRejectedByTheValidation"
		err = player.Locator("input[name='player_nickname']").Fill(longNickname)
		require.NoError(t, err)

		err = player.GetByText("Start").Click()
		require.NoError(t, err)

	})

	t.Run("Should handle special characters in nicknames", func(t *testing.T) {
		playerPages, err := setupTest(t, 1)
		require.NoError(t, err)

		player := playerPages[0]

		// Try with special characters
		specialNickname := "Player@#$%^&*()"
		err = player.Locator("input[name='player_nickname']").Fill(specialNickname)
		require.NoError(t, err)

		err = player.GetByText("Start").Click()
		require.NoError(t, err)

	})
}
