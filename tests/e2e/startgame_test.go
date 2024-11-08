package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestE2EStartGame(t *testing.T) {
	t.Run("Should start game with two players", func(t *testing.T) {
		t.Cleanup(ResetBrowserContexts)
		hostPlayerPage := pages[0]
		otherPlayerPage := pages[1]

		err := joinRoom(hostPlayerPage, otherPlayerPage)
		require.NoError(t, err)

		avatars := otherPlayerPage.GetByAltText("avatar")
		err = expect.Locator(avatars).ToHaveCount(2)
		require.NoError(t, err)

		err = otherPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Ready"}).Click()
		require.NoError(t, err)
		err = hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Ready"}).Click()
		require.NoError(t, err)
		err = hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Start Game"}).Click()
		require.NoError(t, err)

		roundNum := otherPlayerPage.GetByText("Round 1 / 3")
		err = expect.Locator(roundNum).ToBeVisible()
		require.NoError(t, err)
	})

	t.Run("Should submit answer", func(t *testing.T) {
		t.Cleanup(ResetBrowserContexts)
		hostPlayerPage := pages[0]
		otherPlayerPage := pages[1]

		err := joinRoom(hostPlayerPage, otherPlayerPage)
		require.NoError(t, err)

		avatars := otherPlayerPage.GetByAltText("avatar")
		err = expect.Locator(avatars).ToHaveCount(2)
		require.NoError(t, err)

		err = otherPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Ready"}).Click()
		require.NoError(t, err)
		err = hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Ready"}).Click()
		require.NoError(t, err)
		err = hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Start Game"}).Click()
		require.NoError(t, err)

		err = hostPlayerPage.Locator("#submit_answer_form").
			Locator(`input[name="answer"]`).
			Fill("this is a test answer")
		require.NoError(t, err)
		err = hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Submit"}).Click()
		require.NoError(t, err)
	})
}
