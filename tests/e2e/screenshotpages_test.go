package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestE2EScreenshotPages(t *testing.T) {
	t.Run("Should successfully screenshot every page", func(t *testing.T) {
		t.Cleanup(ResetBrowserContexts)
		hostPlayerPage := pages[0]
		otherPlayerPage := pages[1]

		hostPlayerPage.Screenshot(playwright.PageScreenshotOptions{
			Path: playwright.String("home.png"),
		})

		err := joinRoom(hostPlayerPage, otherPlayerPage)
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

		err = hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Not Ready"}).Click()
		require.NoError(t, err)

		err = otherPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Close"}).Click()
		require.NoError(t, err)

		err = otherPlayerPage.Locator("#submit_answer_form").
			Locator(`input[name="answer"]`).
			Fill("this is a another answer")
		require.NoError(t, err)

		err = otherPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Submit Answer"}).Click()
		require.NoError(t, err)

		err = otherPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Not Ready"}).Click()
		require.NoError(t, err)

		votesText := hostPlayerPage.GetByText("Votes:")
		expect.Locator(votesText).ToBeVisible()

		// TODO: seems to intermittently fail here around here
		// err = hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Submit Vote"}).Click()
		// require.NoError(t, err)
		//
		// hostPlayerPage.Screenshot(playwright.PageScreenshotOptions{
		// 	Path: playwright.String("votes.png"),
		// })
		//
		// votedFor := hostPlayerPage.GetByText("You all voted for")
		// playwright.Locator.WaitFor(votedFor)
		// hostPlayerPage.Screenshot(playwright.PageScreenshotOptions{
		// 	Path: playwright.String("reveal.png"),
		// })
		//
		// scoring := hostPlayerPage.GetByText("100")
		// playwright.Locator.WaitFor(scoring)
		// hostPlayerPage.Screenshot(playwright.PageScreenshotOptions{
		// 	Path: playwright.String("scoring.png"),
		// })
	})
}
