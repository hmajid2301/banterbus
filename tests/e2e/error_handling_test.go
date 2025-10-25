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

		err = player.Locator("input[name='player_nickname']").Fill("TestPlayer")
		require.NoError(t, err)

		err = player.Locator("input[name='room_code']").Fill("INVALID")
		require.NoError(t, err)

		err = player.GetByText("Join").Click()
		require.NoError(t, err)

		errorAlert := player.Locator("[role='alert']").Filter(playwright.LocatorFilterOptions{
			HasText: "Failed to join room",
		})
		err = expect.Locator(errorAlert).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
			Timeout: playwright.Float(8000),
		})
		require.NoError(t, err)
	})

	t.Run("Should handle extremely long nicknames", func(t *testing.T) {
		playerPages, err := setupTest(t, 1)
		require.NoError(t, err)

		player := playerPages[0]

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

		specialNickname := "Player@#$%^&*()"
		err = player.Locator("input[name='player_nickname']").Fill(specialNickname)
		require.NoError(t, err)

		err = player.GetByText("Start").Click()
		require.NoError(t, err)

	})
}
