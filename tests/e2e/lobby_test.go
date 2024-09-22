package e2e

import (
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2ELobby(t *testing.T) {

	t.Run("Should not be able to join game that doesn't exist", func(t *testing.T) {
		t.Cleanup(ResetBrowserContexts)
		hostPlayerPage := pages[0]
		otherPlayerPage := pages[1]

		err := hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Create Room"}).Click()
		require.NoError(t, err)

		err = otherPlayerPage.Locator("#join_room_form").GetByPlaceholder("Enter your room code").Fill("FAKE_CODE")
		require.NoError(t, err)

		err = otherPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Join Room"}).Click()
		require.NoError(t, err)
	})

	t.Run("Should be able to update nickname and avatar in lobby", func(t *testing.T) {
		t.Cleanup(ResetBrowserContexts)
		hostPlayerPage := pages[0]
		otherPlayerPage := pages[1]

		err := joinRoom(hostPlayerPage, otherPlayerPage)
		require.NoError(t, err)

		hostAvatar := hostPlayerPage.GetByAltText("avatar").First()
		oldSrc, err := hostAvatar.GetAttribute("src")
		require.NoError(t, err)
		err = hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Update Avatar"}).Click()
		require.NoError(t, err)

		// TODO: work out how to wait for image to change
		time.Sleep(100 * time.Millisecond)

		hostAvatar = hostPlayerPage.GetByAltText("avatar").First()
		newSrc, err := hostAvatar.GetAttribute("src")
		require.NoError(t, err)
		assert.NotEqual(t, oldSrc, newSrc)

		err = hostPlayerPage.Locator("#update_nickname_form").Locator(`input[name="player_nickname"]`).Fill("test_nickname")
		require.NoError(t, err)
		err = hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Update Nickname"}).Click()
		require.NoError(t, err)

		newNickname := hostPlayerPage.Locator("#update_nickname_form").Locator(`input[name="player_nickname"]`)
		err = expect.Locator(newNickname).ToHaveValue("test_nickname")
		require.NoError(t, err)
	})
}
