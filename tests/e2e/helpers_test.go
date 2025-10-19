package e2e

import (
	"errors"
	"fmt"
	"time"

	"github.com/playwright-community/playwright-go"
)

func joinRoom(hostPlayerPage playwright.Page, otherPlayerPages []playwright.Page) error {
	err := hostPlayerPage.Locator("input[name='player_nickname']").Fill("HostPlayer")
	if err != nil {
		return err
	}

	err = hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{
		Name: "Start",
	}).Click()
	if err != nil {
		return err
	}

	locator := hostPlayerPage.Locator("input[name='room_code']")
	err = expect.Locator(locator).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
		Timeout: playwright.Float(10000),
	})
	if err != nil {
		return fmt.Errorf("room code input not visible: %w", err)
	}

	var code string
	for i := 0; i < 10; i++ {
		code, err = locator.InputValue()
		if err != nil {
			fmt.Printf("failed to get room code (attempt %d): %v\n", i+1, err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		if code != "" {
			break
		}

		time.Sleep(500 * time.Millisecond)
	}

	if code == "" {
		return errors.New("room code is empty")
	}

	for i, player := range otherPlayerPages {
		err := player.Locator("input[name='player_nickname']").Fill(fmt.Sprintf("OtherPlayer%d", i))
		if err != nil {
			return err
		}

		err = player.Locator("input[name='room_code']").Fill(code)
		if err != nil {
			return err
		}

		err = player.GetByRole("button", playwright.PageGetByRoleOptions{
			Name: "Join",
		}).Click()
		if err != nil {
			return err
		}
	}
	return nil
}

func startGame(hostPlayerPage playwright.Page, otherPlayerPages []playwright.Page) (string, error) {
	err := joinRoom(hostPlayerPage, otherPlayerPages)
	if err != nil {
		return "", err
	}

	roomCodeInput := hostPlayerPage.Locator("input[name='room_code']")
	err = roomCodeInput.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(10000),
	})
	if err != nil {
		return "", fmt.Errorf("room code input not visible: %w", err)
	}

	code, err := roomCodeInput.InputValue()
	if err != nil {
		return "", fmt.Errorf("failed to get room code: %w", err)
	}

	for _, player := range append(otherPlayerPages, hostPlayerPage) {
		readyButton := player.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Ready"})
		err = expect.Locator(readyButton).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
			Timeout: playwright.Float(10000),
		})
		if err != nil {
			return "", fmt.Errorf("ready button not visible: %w", err)
		}

		err = readyButton.Click()
		if err != nil {
			return "", fmt.Errorf("failed to click ready: %w", err)
		}

		player.WaitForTimeout(1000)
	}

	startGameButton := hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Start Game"})
	err = expect.Locator(startGameButton).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
		Timeout: playwright.Float(10000),
	})
	if err != nil {
		return "", fmt.Errorf("start game button not visible: %w", err)
	}

	err = startGameButton.Click()
	if err != nil {
		return "", fmt.Errorf("failed to click start game: %w", err)
	}

	roundElement := hostPlayerPage.Locator(":has-text('Round')").First()
	err = roundElement.WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(5000),
		State:   playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		return "", fmt.Errorf("failed to wait for game to start: %w", err)
	}

	questionForm := hostPlayerPage.Locator("#submit_answer_form")
	err = questionForm.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(10000),
	})
	if err != nil {
		return "", fmt.Errorf("failed to wait for question form: %w", err)
	}

	hostPlayerPage.WaitForTimeout(2000)

	return code, nil
}

func submitAnswerForPlayer(player playwright.Page, answer string) error {
	answerForm := player.Locator("#submit_answer_form")
	err := answerForm.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(10000),
	})
	if err != nil {
		return fmt.Errorf("answer form not visible: %w", err)
	}

	inputField := player.Locator("#submit_answer_form").Locator(`input[name="answer"]`)
	err = inputField.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(2000),
	})

	if err != nil {
		answerButton := player.GetByRole("button", playwright.PageGetByRoleOptions{Name: answer})
		err = answerButton.WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(10000),
		})
		if err != nil {
			return fmt.Errorf("answer button '%s' not found: %w", answer, err)
		}
		return answerButton.Click()
	} else {
		err = inputField.Fill(answer)
		if err != nil {
			return fmt.Errorf("failed to fill answer: %w", err)
		}

		submitButton := player.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Submit Answer"})
		err = submitButton.WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(5000),
		})
		if err != nil {
			return fmt.Errorf("submit button not found: %w", err)
		}

		err = submitButton.Click()
		if err != nil {
			return fmt.Errorf("failed to click submit button: %w", err)
		}
		return nil
	}
}

func getPlayerRoles(
	hostPlayerPage playwright.Page,
	otherPlayerPages []playwright.Page,
) (playwright.Page, []playwright.Page, error) {
	var fibber playwright.Page
	normals := []playwright.Page{}

	submitButton := hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Ready"})
	err := submitButton.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		return fibber, normals, fmt.Errorf("failed to find Ready button: %w", err)
	}

	fibberFound := false
	for i := 0; i < 3; i++ {
		fibberFound = false
		for _, player := range append(otherPlayerPages, hostPlayerPage) {
			if fibberFound {
				normals = append(normals, player)
				continue
			}

			roleElement := player.Locator("[data-player-role]")
			err = roleElement.WaitFor(playwright.LocatorWaitForOptions{
				State:   playwright.WaitForSelectorStateVisible,
				Timeout: playwright.Float(5000),
			})
			if err != nil {
				return fibber, normals, fmt.Errorf("failed to find role element on page for player: %w", err)
			}

			roleAttribute, err := roleElement.GetAttribute("data-player-role")
			if err != nil {
				return fibber, normals, fmt.Errorf("failed to get role attribute: %w", err)
			}

			if roleAttribute == "fibber" {
				fibber = player
				fibberFound = true
			} else {
				normals = append(normals, player)
			}
		}

		if fibberFound {
			break
		}
	}

	if !fibberFound {
		return fibber, normals, fmt.Errorf("failed to find fibber")
	}

	return fibber, normals, nil
}

func reconnectToRoom(page playwright.Page, nickname, roomCode string) error {
	nicknameInput := page.GetByPlaceholder("Enter your nickname")
	err := nicknameInput.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(10000),
	})
	if err != nil {
		return fmt.Errorf("nickname input not visible: %w", err)
	}

	err = nicknameInput.Fill(nickname)
	if err != nil {
		return err
	}

	err = page.GetByPlaceholder("ABC12").Fill(roomCode)
	if err != nil {
		return err
	}

	err = page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Join"}).Click()
	if err != nil {
		return err
	}

	return nil
}
