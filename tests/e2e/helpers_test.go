package e2e

import (
	"errors"
	"fmt"
	"time"

	"github.com/playwright-community/playwright-go"
)

func joinRoom(hostPlayerPage playwright.Page, otherPlayerPages []playwright.Page) error {
	err := hostPlayerPage.GetByPlaceholder("Enter your nickname").Fill("HostPlayer")
	if err != nil {
		return err
	}

	err = hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Start"}).Click()
	if err != nil {
		return err
	}

	locator := hostPlayerPage.Locator("input[name='room_code']")
	err = expect.Locator(locator).ToBeVisible(playwright.LocatorAssertionsToBeVisibleOptions{
		Timeout: playwright.Float(30000), // 30 seconds for room creation
	})
	if err != nil {
		return fmt.Errorf("room code input not visible: %w", err)
	}

	// INFO: Retry mechanism to wait for the room code to be available
	var code string
	for i := 0; i < 10; i++ { // Increased attempts
		code, err = locator.InputValue()
		if err != nil {
			fmt.Printf("failed to get room code (attempt %d): %v\n", i+1, err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		if code != "" {
			break
		}

		time.Sleep(500 * time.Millisecond) // Increased wait time
	}

	if code == "" {
		return errors.New("room code is empty")
	}

	for i, player := range otherPlayerPages {
		err := player.GetByPlaceholder("Enter your nickname").Fill(fmt.Sprintf("OtherPlayer%d", i))
		if err != nil {
			return err
		}

		err = player.GetByPlaceholder("ABC12").Fill(code)
		if err != nil {
			return err
		}

		err = player.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Join"}).Click()
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

	// Get room code before starting the game
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
			Timeout: playwright.Float(30000), // 30 seconds
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
		Timeout: playwright.Float(30000), // 30 seconds
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
		Timeout: playwright.Float(60000), // 60 second timeout for game start in CI
		State:   playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		return "", fmt.Errorf("failed to wait for game to start: %w", err)
	}

	hostPlayerPage.WaitForTimeout(5000) // Increased for CI stability

	return code, nil
}

func getPlayerRoles(
	hostPlayerPage playwright.Page,
	otherPlayerPages []playwright.Page,
) (playwright.Page, []playwright.Page, error) {
	var fibber playwright.Page
	normals := []playwright.Page{}

	submitButton := hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Ready"})
	err := submitButton.WaitFor()
	if err != nil {
		return fibber, normals, fmt.Errorf("failed to find 'Submit Answer' button: %w", err)
	}

	fibberFound := false
	for i := 0; i < 3; i++ {
		fibberFound = false
		for _, player := range append(otherPlayerPages, hostPlayerPage) {
			if fibberFound {
				normals = append(normals, player)
				continue
			}

			fibberText := player.GetByText("You are fibber")
			isFibber, err := fibberText.IsVisible()
			if err != nil {
				return fibber, normals, err
			}

			if !isFibber {
				normals = append(normals, player)
				continue
			}

			fibber = player
			fibberFound = true
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
	err := page.GetByPlaceholder("Enter your nickname").Fill(nickname)
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
