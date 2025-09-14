package e2e

import (
	"fmt"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2ELobby(t *testing.T) {

	playerNum := 2
	t.Run("Should not be able to join game that doesn't exist", func(t *testing.T) {

		playerPages, err := setupTest(t, playerNum)
		require.NoError(t, err)

		hostPlayerPage := playerPages[0]
		otherPlayerPage := playerPages[1]

		err = hostPlayerPage.Locator("input[name='player_nickname']").Fill("HostPlayer")
		require.NoError(t, err)

		startButton := hostPlayerPage.GetByText("Start")
		err = expect.Locator(startButton).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
			Timeout: playwright.Float(5000),
		})
		require.NoError(t, err)

		err = startButton.Click()
		require.NoError(t, err)

		err = expect.Locator(hostPlayerPage.Locator("input[name='room_code']")).
			ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
				Timeout: playwright.Float(5000),
			})
		require.NoError(t, err)

		err = otherPlayerPage.Locator("input[name='player_nickname']").Fill("OtherPlayer")
		require.NoError(t, err)
		err = otherPlayerPage.Locator("input[name='room_code']").Fill("FAKE_CODE")
		require.NoError(t, err)
		err = otherPlayerPage.GetByText("Join").Click()
		require.NoError(t, err)

		err = expect.Locator(otherPlayerPage.Locator("text=failed to join room")).
			ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
				Timeout: playwright.Float(5000),
			})
		require.NoError(t, err)
	})

	t.Run("Should be able to join game using the join URL", func(t *testing.T) {

		playerPages, err := setupTest(t, playerNum)
		require.NoError(t, err)
		hostPlayerPage := playerPages[0]
		otherPlayerPage := playerPages[1]

		err = hostPlayerPage.Locator("input[name='player_nickname']").Fill("HostPlayer")
		require.NoError(t, err)
		err = hostPlayerPage.GetByText("Start").Click()
		require.NoError(t, err)

		copyButton := hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{
			Name: "Copy Join Link",
		})
		err = expect.Locator(copyButton).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
			Timeout: playwright.Float(5000),
		})
		require.NoError(t, err)

		err = copyButton.Click()
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

		err = otherPlayerPage.Locator("input[name='player_nickname']").Fill("OtherPlayer")
		require.NoError(t, err)

		err = otherPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{
			Name: "Join",
		}).Click()
		require.NoError(t, err)

		roomCodeInput := otherPlayerPage.Locator("input[name='room_code']")
		err = expect.Locator(roomCodeInput).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
			Timeout: playwright.Float(5000),
		})
		require.NoError(t, err)

		actualCode, err := roomCodeInput.InputValue()
		require.NoError(t, err)
		assert.Equal(t, code, actualCode)
	})

	t.Run("Should be able to update nickname and avatar in lobby", func(t *testing.T) {

		playerPages, err := setupTest(t, playerNum)
		require.NoError(t, err)

		hostPlayerPage := playerPages[0]
		otherPlayerPage := playerPages[1:]

		err = joinRoom(hostPlayerPage, otherPlayerPage)
		require.NoError(t, err)

		hostAvatar := hostPlayerPage.GetByAltText("avatar").First()
		oldSrc, err := hostAvatar.GetAttribute("src")
		require.NoError(t, err)
		updateAvatarButton := hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{
			Name: "Update Avatar",
		})
		err = updateAvatarButton.WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(5000),
		})
		require.NoError(t, err)

		err = updateAvatarButton.Click()
		require.NoError(t, err)

		_, err = hostPlayerPage.WaitForFunction(fmt.Sprintf(`() => {
            const avatar = document.querySelector('img[alt="avatar"]');
            return avatar && avatar.src !== '%s';
        }`, oldSrc), nil)
		require.NoError(t, err)

		nicknameInput := hostPlayerPage.Locator("#update_nickname_form").Locator(`input[name="player_nickname"]`)
		err = nicknameInput.Clear()
		require.NoError(t, err)

		err = nicknameInput.Fill("test_nickname")
		require.NoError(t, err)

		err = nicknameInput.Press("Enter")
		require.NoError(t, err)

		newNickname := hostPlayerPage.Locator("#update_nickname_form").Locator(`input[name="player_nickname"]`)
		err = expect.Locator(newNickname).
			ToHaveValue("test_nickname", playwright.LocatorAssertionsToHaveValueOptions{Timeout: playwright.Float(5000)})
		require.NoError(t, err)
	})

	t.Run("Should be able to kick player in lobby", func(t *testing.T) {

		playerPages, err := setupTest(t, playerNum)
		require.NoError(t, err)
		hostPlayerPage := playerPages[0]
		otherPlayerPage := playerPages[1]

		err = joinRoom(hostPlayerPage, playerPages[1:])
		require.NoError(t, err)

		readyButton := hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Ready"})
		err = expect.Locator(readyButton).
			ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
				Timeout: playwright.Float(5000),
			})
		require.NoError(t, err)

		kickButton := hostPlayerPage.Locator("button[aria-label='Kick Player']")
		err = expect.Locator(kickButton).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
			Timeout: playwright.Float(5000),
		})
		require.NoError(t, err)

		err = kickButton.Click()
		require.NoError(t, err)

		hostPlayerPage.WaitForTimeout(1500)

		modal := hostPlayerPage.Locator("#kick-player-modal")

		err = modal.WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateAttached,
			Timeout: playwright.Float(5000),
		})
		require.NoError(t, err)

		err = modal.WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(5000),
		})
		require.NoError(t, err)

		hostPlayerPage.WaitForTimeout(1000)

		modalConfirmButton := modal.Locator("button.bg-red")
		err = expect.Locator(modalConfirmButton).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
			Timeout: playwright.Float(5000),
		})
		require.NoError(t, err)

		err = modalConfirmButton.Click()
		require.NoError(t, err)

		toastMessage := otherPlayerPage.GetByText("you have been kicked from the room", playwright.PageGetByTextOptions{
			Exact: playwright.Bool(false),
		})
		err = expect.Locator(toastMessage).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
			Timeout: playwright.Float(5000),
		})
		require.NoError(t, err)
	})
}
