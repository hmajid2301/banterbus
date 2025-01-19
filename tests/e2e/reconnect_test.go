package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2EReconnect(t *testing.T) {
	playerNum := 2
	t.Run("Should be able to reconnect to room with just one player", func(t *testing.T) {
		p := ResetBrowserContexts(playerNum)
		t.Cleanup(func() { ResetBrowserContexts(playerNum) })
		hostPlayerPage := p[0]

		err := hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Start"}).Click()
		require.NoError(t, err)

		nickname, err := hostPlayerPage.Locator("#update_nickname_form").
			Locator(`input[name="player_nickname"]`).InputValue()
		require.NoError(t, err)

		code, err := hostPlayerPage.Locator("input[name='room_code']").InputValue()
		require.NoError(t, err)

		_, err = hostPlayerPage.Reload()
		require.NoError(t, err)

		err = hostPlayerPage.GetByPlaceholder("ABC12").Fill(code)
		require.NoError(t, err)
		err = hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Join"}).Click()
		require.NoError(t, err)

		notReady := hostPlayerPage.GetByText("Not Ready")
		err = expect.Locator(notReady).ToBeVisible()
		assert.NoError(t, err)

		refreshNickname, err := hostPlayerPage.Locator("#update_nickname_form").
			Locator(`input[name="player_nickname"]`).InputValue()
		require.NoError(t, err)
		assert.Equal(t, nickname, refreshNickname)
	})

	t.Run("Should be able to reconnect with started game showing questions", func(t *testing.T) {
		p := ResetBrowserContexts(playerNum)
		t.Cleanup(func() { ResetBrowserContexts(playerNum) })
		hostPlayerPage := p[0]
		otherPlayerPage := p[1]

		err := joinRoom(hostPlayerPage, p[1:])
		require.NoError(t, err)

		code, err := hostPlayerPage.Locator("input[name='room_code']").InputValue()
		require.NoError(t, err)

		err = otherPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Ready"}).Click()
		require.NoError(t, err)
		err = hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Ready"}).Click()
		require.NoError(t, err)
		err = hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Start Game"}).Click()
		require.NoError(t, err)

		_, err = hostPlayerPage.Reload()
		require.NoError(t, err)

		// TODO: maybe refactor this connect code
		err = hostPlayerPage.GetByPlaceholder("ABC12").Fill(code)
		require.NoError(t, err)
		err = hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Join"}).Click()
		require.NoError(t, err)

		roundNum := hostPlayerPage.GetByText("Round 1 / 3")
		err = expect.Locator(roundNum).ToBeVisible()
		require.NoError(t, err)
	})
	// TODO: voting
	// TODO: reveal
	// TODO: scoring
}
