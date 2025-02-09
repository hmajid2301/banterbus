package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestE2EScreenshotPages(t *testing.T) {
	playerNum := 2
	t.Run("Should successfully screenshot every page", func(t *testing.T) {
		playerPages, teardown := setupTest(playerNum)
		t.Cleanup(func() { teardown(playerPages) })

		hostPlayerPage := playerPages[0]
		otherPlayerPage := playerPages[1]

		hostPlayerPage.Screenshot(playwright.PageScreenshotOptions{
			Path: playwright.String("home.png"),
		})

		err := joinRoom(hostPlayerPage, playerPages[1:])
		require.NoError(t, err)

		hostPlayerPage.Screenshot(playwright.PageScreenshotOptions{
			Path: playwright.String("lobby.png"),
		})

		err = otherPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Ready"}).Click()
		require.NoError(t, err)

		err = hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Ready"}).Click()
		require.NoError(t, err)

		err = hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Start Game"}).Click()
		require.NoError(t, err)

		hostPlayerPage.Screenshot(playwright.PageScreenshotOptions{
			Path: playwright.String("question_modal.png"),
		})

		err = hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Close"}).Click()
		require.NoError(t, err)

		hostPlayerPage.Screenshot(playwright.PageScreenshotOptions{
			Path: playwright.String("question.png"),
		})

		err = hostPlayerPage.Locator("#submit_answer_form").
			Locator(`input[name="answer"]`).
			Fill("this is a test answer")
		require.NoError(t, err)

		err = hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Submit Answer"}).Click()
		require.NoError(t, err)

		err = hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Ready"}).Click()
		require.NoError(t, err)

		err = otherPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Close"}).Click()
		require.NoError(t, err)

		err = otherPlayerPage.Locator("#submit_answer_form").
			Locator(`input[name="answer"]`).
			Fill("this is a another answer")
		require.NoError(t, err)

		err = otherPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Submit Answer"}).Click()
		require.NoError(t, err)

		err = otherPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Ready"}).Click()
		require.NoError(t, err)

		votesText := hostPlayerPage.GetByText("Votes:")
		expect.Locator(votesText).ToBeVisible()

		err = hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Submit Vote"}).Click()
		require.NoError(t, err)

		hostPlayerPage.Screenshot(playwright.PageScreenshotOptions{
			Path: playwright.String("votes.png"),
		})

		votedFor := hostPlayerPage.GetByText("You all voted for")
		playwright.Locator.WaitFor(votedFor)
		hostPlayerPage.Screenshot(playwright.PageScreenshotOptions{
			Path: playwright.String("reveal.png"),
		})

		scoring := hostPlayerPage.GetByText("Scoreboard")
		playwright.Locator.WaitFor(scoring)
		hostPlayerPage.Screenshot(playwright.PageScreenshotOptions{
			Path: playwright.String("scoring.png"),
		})
	})
}
