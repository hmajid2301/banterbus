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

		playerPages, err := setupTest(t, playerNum)
		require.NoError(t, err)

		hostPlayerPage := playerPages[0]
		otherPlayerPages := playerPages[1:]

		_, err = startGame(hostPlayerPage, otherPlayerPages)
		require.NoError(t, err)

		fibber, normals, err := getPlayerRoles(hostPlayerPage, otherPlayerPages)
		require.NoError(t, err)

		for _, player := range playerPages {
			closeButton := player.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Close"})
			err = closeButton.WaitFor(playwright.LocatorWaitForOptions{
				State:   playwright.WaitForSelectorStateVisible,
				Timeout: playwright.Float(5000),
			})
			require.NoError(t, err)

			err = closeButton.Click()
			require.NoError(t, err)
		}

		err = submitAnswerForPlayer(fibber, "I am not a fibber")
		require.NoError(t, err)

		err = fibber.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Ready"}).Click()
		require.NoError(t, err)

		for i, normal := range normals {
			err = submitAnswerForPlayer(normal, fmt.Sprintf("I am a normal player %d", i))
			require.NoError(t, err)

			err = normal.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Ready"}).Click()
			require.NoError(t, err)
		}

		fibberTest := hostPlayerPage.GetByText("I am not a fibber")
		err = expect.Locator(fibberTest).ToBeVisible()
		require.NoError(t, err)

		for _, normal := range normals {
			err = normal.GetByText("I am not a fibber").Click()
			require.NoError(t, err)
		}

		for _, player := range append(normals, fibber) {
			fibberCaughtText := player.GetByText("They were fibber")
			err = expect.Locator(fibberCaughtText).ToBeVisible()
			require.NoError(t, err)
		}

		scoreboardText := hostPlayerPage.GetByText("Scoreboard")
		err = expect.Locator(scoreboardText).ToBeVisible()
		require.NoError(t, err)

		roundText := hostPlayerPage.GetByText("Round")
		err = expect.Locator(roundText).ToBeVisible()
		require.NoError(t, err)

		fibber, normals, err = getPlayerRoles(hostPlayerPage, otherPlayerPages)
		require.NoError(t, err)

		for _, player := range playerPages {
			closeButton := player.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Close"})
			err = closeButton.WaitFor(playwright.LocatorWaitForOptions{
				State:   playwright.WaitForSelectorStateVisible,
				Timeout: playwright.Float(5000),
			})
			require.NoError(t, err)

			err = closeButton.Click()
			require.NoError(t, err)
		}

		for _, normal := range normals {
			agreeButton := normal.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Strongly Agree"})
			err = agreeButton.WaitFor(playwright.LocatorWaitForOptions{
				State:   playwright.WaitForSelectorStateVisible,
				Timeout: playwright.Float(5000),
			})
			require.NoError(t, err)

			err = agreeButton.Click()
			require.NoError(t, err)

			readyButton := normal.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Ready"})
			err = readyButton.WaitFor(playwright.LocatorWaitForOptions{
				State:   playwright.WaitForSelectorStateVisible,
				Timeout: playwright.Float(5000),
			})
			require.NoError(t, err)

			err = readyButton.Click()
			require.NoError(t, err)
		}

		err = fibber.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Strongly Disagree"}).Click()
		require.NoError(t, err)

		err = fibber.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Ready"}).Click()
		require.NoError(t, err)

		fibberTest = hostPlayerPage.GetByText("Strongly Disagree")
		err = expect.Locator(fibberTest).ToBeVisible()
		require.NoError(t, err)

		for _, normal := range normals {
			err = normal.GetByText("Strongly Disagree").Click()
			require.NoError(t, err)
		}

		for _, player := range append(normals, fibber) {
			fibberCaughtText := player.GetByText("They were fibber")
			err = expect.Locator(fibberCaughtText).ToBeVisible()
			require.NoError(t, err)
		}

		scoreboardText = hostPlayerPage.GetByText("Scoreboard")
		err = expect.Locator(scoreboardText).ToBeVisible()
		require.NoError(t, err)

		roundText = hostPlayerPage.GetByText("Round")
		err = expect.Locator(roundText).ToBeVisible()
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

		searchText := fmt.Sprintf("Answer: %s", fibberPlayerVotedName)
		fibberTest = hostPlayerPage.GetByText(searchText)

		err = expect.Locator(fibberTest).ToBeVisible()
		require.NoError(t, err)

		for _, normal := range normals {
			err = normal.GetByText(searchText, playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).Click()
			require.NoError(t, err)
		}

		for _, player := range append(normals, fibber) {
			fibberCaughtText := player.GetByText("They were fibber")
			err = expect.Locator(fibberCaughtText).ToBeVisible()
			require.NoError(t, err)
		}

		winnerText := hostPlayerPage.GetByText("The winner is")
		err = expect.Locator(winnerText).ToBeVisible()
		require.NoError(t, err)
	})

	t.Run("Should successfully complete an entire round without guessing the fibber", func(t *testing.T) {
		playerPages, err := setupTest(t, playerNum)
		require.NoError(t, err)

		hostPlayerPage := playerPages[0]
		otherPlayerPages := playerPages[1:]

		_, err = startGame(hostPlayerPage, otherPlayerPages)
		require.NoError(t, err)

		fibber, normals, err := getPlayerRoles(hostPlayerPage, otherPlayerPages)
		require.NoError(t, err)

		for _, player := range playerPages {
			closeButton := player.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Close"})
			err = closeButton.WaitFor(playwright.LocatorWaitForOptions{
				State:   playwright.WaitForSelectorStateVisible,
				Timeout: playwright.Float(5000),
			})
			require.NoError(t, err)
			err = closeButton.Click()
			require.NoError(t, err)
			player.WaitForTimeout(500)
		}

		answerForm := fibber.Locator("#submit_answer_form")
		err = answerForm.WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(5000),
		})
		require.NoError(t, err)

		for i := 0; i < 2; i++ {
			err = submitAnswerForPlayer(fibber, "I am not a fibber")
			require.NoError(t, err)

			fibber.WaitForTimeout(500)

			readyButton := fibber.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Ready"})
			err = readyButton.WaitFor(playwright.LocatorWaitForOptions{
				State:   playwright.WaitForSelectorStateVisible,
				Timeout: playwright.Float(5000),
			})
			require.NoError(t, err)

			err = readyButton.Click()
			require.NoError(t, err)

			for j, normal := range normals {
				err = submitAnswerForPlayer(normal, fmt.Sprintf("I am a normal player %d", j))
				require.NoError(t, err)

				normal.WaitForTimeout(500)

				readyButton := normal.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Ready"})
				err = readyButton.WaitFor(playwright.LocatorWaitForOptions{
					State:   playwright.WaitForSelectorStateVisible,
					Timeout: playwright.Float(5000),
				})
				require.NoError(t, err)

				err = readyButton.Click()
				require.NoError(t, err)
			}

			fibberTest := hostPlayerPage.GetByText("I am not a fibber")
			expect.Locator(fibberTest).ToBeVisible()
			for _, normal := range normals {
				err = normal.GetByText("I am a normal player 1").Click()
				require.NoError(t, err)
			}

			for _, player := range append(normals, fibber) {
				notCaught := player.GetByText("You failed to vote for a single player ...")
				expect.Locator(notCaught).ToBeVisible()
			}

			time.Sleep(time.Second * 1)
		}
	})
}
