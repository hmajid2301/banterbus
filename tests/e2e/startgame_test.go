package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestE2EStartGame(t *testing.T) {
	t.Cleanup(ResetBrowserContexts)
	hostPlayerPage := pages[0]
	otherPlayerPage := pages[1]

	t.Run("Should start game with two players", func(t *testing.T) {
		err := joinRoom(hostPlayerPage, otherPlayerPage)
		require.NoError(t, err)

		avatars := otherPlayerPage.GetByAltText("avatar")
		err = expect.Locator(avatars).ToHaveCount(2)
		require.NoError(t, err)

		err = otherPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Toggle Ready"}).Click()
		require.NoError(t, err)
		err = hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Toggle Ready"}).Click()
		require.NoError(t, err)
		err = hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Start Game"}).Click()
		require.NoError(t, err)

		startText := otherPlayerPage.GetByText("Fibbing It Starting")
		err = expect.Locator(startText).ToBeVisible()
		require.NoError(t, err)
	})
}
