package e2e

import (
	"fmt"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestE2ERounds(t *testing.T) {
	playerNum := 6
	players := ResetBrowserContexts(playerNum)

	t.Run("Should successfully complete an entire round type by voting for the fibber", func(t *testing.T) {
		t.Cleanup(func() { ResetBrowserContexts(playerNum) })
		hostPlayerPage := players[0]
		otherPlayerPages := players[1:]

		fibber, normals, err := startAndGetRoles(hostPlayerPage, otherPlayerPages)
		require.NoError(t, err)

		for _, player := range players {
			err = player.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Close"}).Click()
			require.NoError(t, err)
		}

		err = fibber.Locator("#submit_answer_form").Locator(`input[name="answer"]`).Fill("I am not a fibber")
		require.NoError(t, err)

		err = fibber.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Submit Answer"}).Click()
		require.NoError(t, err)

		err = fibber.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Not Ready"}).Click()
		require.NoError(t, err)

		for i, normal := range normals {
			err = normal.Locator("#submit_answer_form").
				Locator(`input[name="answer"]`).
				Fill(fmt.Sprintf("I am a normal player %d", i))
			require.NoError(t, err)

			err = normal.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Submit Answer"}).Click()
			require.NoError(t, err)

			err = normal.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Not Ready"}).Click()
			require.NoError(t, err)
		}

		fibberTest := hostPlayerPage.GetByText("I am not a fibber")
		expect.Locator(fibberTest).ToBeVisible()
		for _, normal := range normals {
			err = normal.GetByText("I am not a fibber").Click()
			require.NoError(t, err)
		}

		for _, player := range append(normals, fibber) {
			fibberCaughtText := player.GetByText("They were fibber")
			expect.Locator(fibberCaughtText).ToBeVisible()
		}

		// for _, player := range append(normals, fibber) {
		// 	scoreboardText := player.GetByText("Scoreboard")
		// 	expect.Locator(scoreboardText).ToBeVisible()
		//
		// 	maxScoreCount, err := player.GetByText("100").Count()
		// 	require.NoError(t, err)
		// 	assert.Equal(t, len(normals), maxScoreCount)
		// }
	})
}
