package e2e

import (
	"fmt"
	"testing"

	"github.com/playwright-community/playwright-go"
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

		err = expect.Locator(otherPlayerPage.Locator("text=Room with code FAKE_CODE does not exist")).ToBeVisible()
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

		_, err = hostPlayerPage.WaitForFunction(fmt.Sprintf(`() => {
            const avatar = document.querySelector('img[alt="avatar"]');
            return avatar && avatar.src !== '%s';
        }`, oldSrc), nil)
		require.NoError(t, err)

		err = hostPlayerPage.Locator("#update_nickname_form").Locator(`input[name="player_nickname"]`).Fill("test_nickname")
		require.NoError(t, err)
		err = hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Update Nickname"}).Click()
		require.NoError(t, err)

		newNickname := hostPlayerPage.Locator("#update_nickname_form").Locator(`input[name="player_nickname"]`)
		err = expect.Locator(newNickname).ToHaveValue("test_nickname")
		require.NoError(t, err)
	})
}
