package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2EReconnect(t *testing.T) {
	t.Parallel()

	playerNum := 2
	t.Run("Should be able to reconnect to room with just one player", func(t *testing.T) {
		t.Parallel()
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
			Timeout: playwright.Float(30000),
		})
		require.NoError(t, err)

		code, err := roomCodeInput.InputValue()
		require.NoError(t, err)

		_, err = hostPlayerPage.Reload(playwright.PageReloadOptions{
			WaitUntil: playwright.WaitUntilStateDomcontentloaded,
			Timeout:   playwright.Float(60000),
		})
		require.NoError(t, err)

		err = hostPlayerPage.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
			State:   playwright.LoadStateNetworkidle,
			Timeout: playwright.Float(30000),
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
		t.Parallel()
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
			Timeout:   playwright.Float(60000),
		})
		require.NoError(t, err)

		err = hostPlayerPage.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
			State:   playwright.LoadStateNetworkidle,
			Timeout: playwright.Float(30000),
		})
		require.NoError(t, err)

		err = reconnectToRoom(hostPlayerPage, "HostPlayer", code)
		require.NoError(t, err)

		roundNum := hostPlayerPage.GetByText("Round 1 / 3")
		err = expect.Locator(roundNum).ToBeVisible()
		require.NoError(t, err)
	})

	// TODO: voting
	// TODO: reveal
	// TODO: scoring
	//
	//	t.Run("Should be able to reconnect during scoring phase", func(t *testing.T) {
	//		t.Parallel()
	//		playerPages, err := setupTest(t, 3)
	//		require.NoError(t, err)
	//
	//		hostPlayerPage := playerPages[0]
	//		otherPlayerPages := playerPages[1:]
	//
	//		code, err := startGame(hostPlayerPage, otherPlayerPages)
	//		require.NoError(t, err)
	//
	//		questionUI := hostPlayerPage.Locator("text=Round").First()
	//		err = expect.Locator(questionUI).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
	//			Timeout: playwright.Float(30000),
	//		})
	//		require.NoError(t, err)
	//
	//		for _, player := range playerPages {
	//			submitButton := player.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Submit Answer"})
	//			err = submitButton.Click()
	//			require.NoError(t, err)
	//		}
	//
	//		votingUI := hostPlayerPage.Locator("text=Votes").First()
	//		err = expect.Locator(votingUI).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
	//			Timeout: playwright.Float(60000),
	//		})
	//		require.NoError(t, err)
	//
	//		voteForm := hostPlayerPage.Locator("form#vote_for_player button").First()
	//		voteFormVisible := expect.Locator(voteForm).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
	//			Timeout: playwright.Float(5000),
	//		})
	//		if voteFormVisible == nil {
	//			for i, player := range playerPages {
	//				playerVoteForm := player.Locator("form#vote_for_player button").First()
	//				playerVoteForm.Click() // Best effort - ignore errors if it fails
	//				if i < len(playerPages)-1 {
	//					player.WaitForTimeout(2000)
	//				}
	//			}
	//		}
	//
	//		postVotingUI := hostPlayerPage.Locator("text=/Scoreboard|New Round!|You all voted for/").First()
	//		err = expect.Locator(postVotingUI).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
	//			Timeout: playwright.Float(75000), // Even longer timeout for CI environment
	//		})
	//		require.NoError(t, err)
	//
	//		_, err = hostPlayerPage.Reload(playwright.PageReloadOptions{
	//			WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	//			Timeout:   playwright.Float(60000),
	//		})
	//		require.NoError(t, err)
	//
	//		err = hostPlayerPage.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
	//			State:   playwright.LoadStateNetworkidle,
	//			Timeout: playwright.Float(30000),
	//		})
	//		require.NoError(t, err)
	//
	//		err = reconnectToRoom(hostPlayerPage, "HostPlayer", code)
	//		require.NoError(t, err)
	//
	//		reconnectedGameUI := hostPlayerPage.Locator("text=/Scoreboard|New Round!|You all voted for|Ready/").First()
	//		err = expect.Locator(reconnectedGameUI).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
	//			Timeout: playwright.Float(30000), // 30 second timeout for reconnection
	//		})
	//		require.NoError(t, err)
	//	})
}
