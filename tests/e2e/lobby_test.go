package e2e

import (
	"fmt"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2ELobby(t *testing.T) {
	t.Parallel()

	playerNum := 2
	t.Run("Should not be able to join game that doesn't exist", func(t *testing.T) {
		t.Parallel()
		playerPages, err := setupTest(t, playerNum)
		require.NoError(t, err)

		hostPlayerPage := playerPages[0]
		otherPlayerPage := playerPages[1]

		err = hostPlayerPage.GetByPlaceholder("Enter your nickname").Fill("HostPlayer")
		require.NoError(t, err)
		err = hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Start"}).Click()
		require.NoError(t, err)
		err = otherPlayerPage.GetByPlaceholder("Enter your nickname").Fill("OtherPlayer")
		require.NoError(t, err)
		err = otherPlayerPage.GetByPlaceholder("ABC12").Fill("FAKE_CODE")
		require.NoError(t, err)
		err = otherPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Join"}).Click()
		require.NoError(t, err)

		err = expect.Locator(otherPlayerPage.Locator("text=failed to join room")).ToBeVisible()
		require.NoError(t, err)
	})

	t.Run("Should be able to join game using the join URL", func(t *testing.T) {
		t.Parallel()
		playerPages, err := setupTest(t, playerNum)
		require.NoError(t, err)
		hostPlayerPage := playerPages[0]
		otherPlayerPage := playerPages[1]

		err = hostPlayerPage.GetByPlaceholder("Enter your nickname").Fill("HostPlayer")
		require.NoError(t, err)
		err = hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Start"}).Click()
		require.NoError(t, err)
		err = hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Copy Join Link"}).Click()
		require.NoError(t, err)

		locator := hostPlayerPage.Locator("input[name='room_code']")
		code, err := locator.InputValue()
		require.NoError(t, err)

		url, err := hostPlayerPage.Evaluate("navigator.clipboard.readText()")
		require.NoError(t, err)

		expectedURL := fmt.Sprintf("%sjoin/%s", hostPlayerPage.URL(), code)
		assert.Equal(t, expectedURL, url)

		_, err = otherPlayerPage.Goto(expectedURL)
		require.NoError(t, err)

		err = otherPlayerPage.GetByPlaceholder("Enter your nickname").Fill("OtherPlayer")
		require.NoError(t, err)

		err = otherPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Join"}).Click()
		require.NoError(t, err)

		expect.Locator(otherPlayerPage.GetByText(code)).ToBeVisible()
		require.NoError(t, err)
	})

	t.Run("Should be able to update nickname and avatar in lobby", func(t *testing.T) {
		t.Parallel()
		playerPages, err := setupTest(t, playerNum)
		require.NoError(t, err)

		hostPlayerPage := playerPages[0]
		otherPlayerPage := playerPages[1:]

		err = joinRoom(hostPlayerPage, otherPlayerPage)
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

		err = hostPlayerPage.Locator("#update_nickname_form").
			Locator(`input[name="player_nickname"]`).
			Fill("test_nickname")
		require.NoError(t, err)
		err = hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Update Nickname"}).Click()
		require.NoError(t, err)

		newNickname := hostPlayerPage.Locator("#update_nickname_form").Locator(`input[name="player_nickname"]`)
		err = expect.Locator(newNickname).ToHaveValue("test_nickname")
		require.NoError(t, err)
	})

	t.Run("Should be able to kick player in lobby", func(t *testing.T) {
		t.Parallel()
		playerPages, err := setupTest(t, playerNum)
		require.NoError(t, err)
		hostPlayerPage := playerPages[0]
		otherPlayerPage := playerPages[1]

		err = joinRoom(hostPlayerPage, playerPages[1:])
		require.NoError(t, err)

		err = hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Kick Player"}).Click()
		require.NoError(t, err)

		err = expect.Locator(otherPlayerPage.GetByText("you have been kicked from the room")).ToBeVisible()
		require.NoError(t, err)
	})
}
