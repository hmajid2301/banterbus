package e2e

import (
	"fmt"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2EReconnect(t *testing.T) {

	playerNum := 2
	t.Run("Should be able to reconnect to room with just one player", func(t *testing.T) {

		playerPages, err := setupTest(t, playerNum)
		require.NoError(t, err)
		hostPlayerPage := playerPages[0]

		err = hostPlayerPage.GetByPlaceholder("Enter your nickname").Fill("HostPlayer")
		require.NoError(t, err)
		err = hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Start"}).Click()
		require.NoError(t, err)

		nickname, err := hostPlayerPage.Locator("#update_nickname_form").
			Locator(`input[name="player_nickname"]`).InputValue()
		require.NoError(t, err)

		roomCodeInput := hostPlayerPage.Locator("input[name='room_code']")
		err = roomCodeInput.WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(5000),
		})
		require.NoError(t, err)

		code, err := roomCodeInput.InputValue()
		require.NoError(t, err)

		_, err = hostPlayerPage.Reload(playwright.PageReloadOptions{
			WaitUntil: playwright.WaitUntilStateDomcontentloaded,
			Timeout:   playwright.Float(5000),
		})
		require.NoError(t, err)

		err = hostPlayerPage.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
			State:   playwright.LoadStateNetworkidle,
			Timeout: playwright.Float(5000),
		})
		require.NoError(t, err)

		err = hostPlayerPage.GetByPlaceholder("Enter your nickname").Fill("HostPlayer")
		require.NoError(t, err)
		err = hostPlayerPage.GetByPlaceholder("ABC12").Fill(code)
		require.NoError(t, err)
		err = hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Join"}).Click()
		require.NoError(t, err)

		readyButton := hostPlayerPage.Locator("button:has-text('Ready')")
		err = expect.Locator(readyButton).ToBeVisible()
		assert.NoError(t, err)

		refreshNickname, err := hostPlayerPage.Locator("#update_nickname_form").
			Locator(`input[name="player_nickname"]`).InputValue()
		require.NoError(t, err)
		assert.Equal(t, nickname, refreshNickname)
	})

	t.Run("Should be able to reconnect with started game showing questions", func(t *testing.T) {

		playerPages, err := setupTest(t, playerNum)
		require.NoError(t, err)

		hostPlayerPage := playerPages[0]
		otherPlayerPage := playerPages[1]

		err = joinRoom(hostPlayerPage, playerPages[1:])
		require.NoError(t, err)

		roomCodeInput := hostPlayerPage.Locator("input[name='room_code']").First()
		err = roomCodeInput.WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(10000),
		})
		require.NoError(t, err)

		code, err := roomCodeInput.InputValue()
		require.NoError(t, err)

		err = otherPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Ready"}).Click()
		require.NoError(t, err)
		err = hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Ready"}).Click()
		require.NoError(t, err)
		err = hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Start Game"}).Click()
		require.NoError(t, err)

		_, err = hostPlayerPage.Reload(playwright.PageReloadOptions{
			WaitUntil: playwright.WaitUntilStateDomcontentloaded,
			Timeout:   playwright.Float(5000),
		})
		require.NoError(t, err)

		err = hostPlayerPage.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
			State:   playwright.LoadStateNetworkidle,
			Timeout: playwright.Float(5000),
		})
		require.NoError(t, err)

		err = reconnectToRoom(hostPlayerPage, "HostPlayer", code)
		require.NoError(t, err)

		roundNum := hostPlayerPage.GetByText("Round 1 / 3")
		err = expect.Locator(roundNum).ToBeVisible()
		require.NoError(t, err)
	})

	t.Run("Should be able to reconnect during scoring phase", func(t *testing.T) {

		playerPages, err := setupTest(t, 3)
		require.NoError(t, err)

		hostPlayerPage := playerPages[0]
		otherPlayerPages := playerPages[1:]

		_, err = startGame(hostPlayerPage, otherPlayerPages)
		require.NoError(t, err)

		questionUI := hostPlayerPage.Locator("text=Round").First()
		err = expect.Locator(questionUI).ToBeVisible()
		require.NoError(t, err)

		fibber, normals, err := getPlayerRoles(hostPlayerPage, otherPlayerPages)
		require.NoError(t, err)

		for _, player := range playerPages {
			err = player.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Close"}).Click()
			require.NoError(t, err)
		}

		err = submitAnswerForPlayer(fibber, "I am the fibber")
		require.NoError(t, err)
		err = fibber.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Ready"}).Click()
		require.NoError(t, err)

		for _, normal := range normals {
			err = submitAnswerForPlayer(normal, "I am not a fibber")
			require.NoError(t, err)

			readyButton := normal.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Ready"})
			err = readyButton.Click()
			require.NoError(t, err)
		}

		votingIndicators := []string{"Votes", "voting.votes", "vote_for_player"}
		var votingUI playwright.Locator
		var lastErr error

		for _, indicator := range votingIndicators {
			if indicator == "vote_for_player" {
				votingUI = hostPlayerPage.Locator("form#vote_for_player").First()
			} else {
				votingUI = hostPlayerPage.Locator(fmt.Sprintf("text=%s", indicator)).First()
			}

			lastErr = expect.Locator(votingUI).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
				Timeout: playwright.Float(5000),
			})
			if lastErr == nil {
				break
			}
		}
		require.NoError(t, lastErr)

		fibberAnswer := hostPlayerPage.GetByText("I am the fibber")
		err = expect.Locator(fibberAnswer).ToBeVisible()
		require.NoError(t, err)

		for _, normal := range normals {
			err = normal.GetByText("I am the fibber").Click()
			require.NoError(t, err)
		}

		err = fibber.GetByText("I am not a fibber").First().Click()
		require.NoError(t, err)

		scoreboardUI := hostPlayerPage.Locator("text=Scoreboard").First()
		err = expect.Locator(scoreboardUI).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
			Timeout: playwright.Float(75000),
		})
		require.NoError(t, err)

		_, err = hostPlayerPage.Reload(playwright.PageReloadOptions{
			WaitUntil: playwright.WaitUntilStateDomcontentloaded,
			Timeout:   playwright.Float(5000),
		})
		require.NoError(t, err)

		err = hostPlayerPage.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
			State:   playwright.LoadStateNetworkidle,
			Timeout: playwright.Float(5000),
		})
		require.NoError(t, err)

		reconnectedGameUI := hostPlayerPage.Locator("text=/Scoreboard|New Round!|You all voted for|Ready/").First()
		err = expect.Locator(reconnectedGameUI).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
			Timeout: playwright.Float(5000),
		})
		require.NoError(t, err)
	})
}
