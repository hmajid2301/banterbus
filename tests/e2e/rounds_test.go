package e2e

import (
	"fmt"
	"testing"

	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestE2ERounds(t *testing.T) {
	playerNum := 6

	t.Run("Should successfully complete an entire game where the fibber is caught first time", func(t *testing.T) {
		playerPages, teardown := setupTest(playerNum)
		t.Cleanup(func() { teardown(playerPages) })

		hostPlayerPage := playerPages[0]
		otherPlayerPages := playerPages[1:]

		err := startGame(hostPlayerPage, otherPlayerPages)
		require.NoError(t, err)

		fibber, normals, err := getPlayerRoles(hostPlayerPage, otherPlayerPages)
		require.NoError(t, err)

		for _, player := range playerPages {
			err = player.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Close"}).Click()
			require.NoError(t, err)
		}

		err = fibber.Locator("#submit_answer_form").Locator(`input[name="answer"]`).Fill("I am not a fibber")
		require.NoError(t, err)

		err = fibber.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Submit Answer"}).Click()
		require.NoError(t, err)

		err = fibber.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Ready"}).Click()
		require.NoError(t, err)

		for i, normal := range normals {
			err = normal.Locator("#submit_answer_form").
				Locator(`input[name="answer"]`).
				Fill(fmt.Sprintf("I am a normal player %d", i))
			require.NoError(t, err)

			err = normal.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Submit Answer"}).Click()
			require.NoError(t, err)

			err = normal.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Ready"}).Click()
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

		scoreboardText := hostPlayerPage.GetByText("Scoreboard")
		expect.Locator(scoreboardText).ToBeVisible()

		roleText := hostPlayerPage.GetByText("You are")
		expect.Locator(roleText).
			ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{Timeout: playwright.Float(35 * 1000)})
		require.NoError(t, err)

		fibber, normals, err = getPlayerRoles(hostPlayerPage, otherPlayerPages)
		require.NoError(t, err)

		for _, player := range playerPages {
			err = player.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Close"}).Click()
			require.NoError(t, err)
		}

		for _, normal := range normals {
			err = normal.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Strongly Agree"}).Click()
			require.NoError(t, err)
			err = normal.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Ready"}).Click()
			require.NoError(t, err)
		}

		err = fibber.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Strongly Disagree"}).Click()
		require.NoError(t, err)

		err = fibber.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Ready"}).Click()
		require.NoError(t, err)

		fibberTest = hostPlayerPage.GetByText("Strongly Disagree")
		expect.Locator(fibberTest).ToBeVisible()
		for _, normal := range normals {
			err = normal.GetByText("Strongly Disagree").Click()
			require.NoError(t, err)
		}

		for _, player := range append(normals, fibber) {
			fibberCaughtText := player.GetByText("They were fibber")
			expect.Locator(fibberCaughtText).ToBeVisible()
		}

		scoreboardText = hostPlayerPage.GetByText("Scoreboard")
		expect.Locator(scoreboardText).ToBeVisible()

		roleText = hostPlayerPage.GetByText("You are")
		expect.Locator(roleText).
			ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{Timeout: playwright.Float(35 * 1000)})
		require.NoError(t, err)

		fibber, normals, err = getPlayerRoles(hostPlayerPage, otherPlayerPages)
		require.NoError(t, err)

		for _, player := range playerPages {
			err = player.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Close"}).Click()
			require.NoError(t, err)
		}

		submitAnswer := fibber.GetByRole("button").Nth(5)
		fibberPlayerVotedName, err := submitAnswer.InnerText()
		require.NoError(t, err)
		err = submitAnswer.Click()
		require.NoError(t, err)
		err = fibber.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Ready"}).Click()
		require.NoError(t, err)

		for _, normal := range normals {
			err = normal.GetByRole("button").Nth(6).Click()
			require.NoError(t, err)

			err = normal.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Ready"}).Click()
			require.NoError(t, err)
		}

		searchText := fmt.Sprintf("Answer %s", fibberPlayerVotedName)
		fibberTest = hostPlayerPage.GetByText(searchText)
		expect.Locator(fibberTest).ToBeVisible()
		for _, normal := range normals {
			err = normal.GetByText(searchText, playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).Click()
			require.NoError(t, err)
		}

		scoreboardText = hostPlayerPage.GetByText("The winner is")
		expect.Locator(scoreboardText).ToBeVisible()

		time.Sleep(time.Second * 1)
	})
}
