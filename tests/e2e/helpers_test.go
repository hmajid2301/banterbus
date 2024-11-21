package e2e

import (
	"fmt"
	"time"

	"github.com/playwright-community/playwright-go"
)

func joinRoom(hostPlayerPage playwright.Page, otherPlayerPage playwright.Page) error {
	err := hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Start"}).Click()
	if err != nil {
		return err
	}

	locator := hostPlayerPage.Locator("input[name='room_code']")

	// Retry mechanism to wait for the room code to be available
	var code string
	for i := 0; i < 5; i++ {
		code, err = locator.InputValue()
		if err != nil {
			return err
		}

		if code != "" {
			break
		}

		time.Sleep(100 * time.Millisecond)
	}

	if code == "" {
		return fmt.Errorf("room code is empty")
	}

	err = otherPlayerPage.GetByPlaceholder("ABC12").Fill(code)
	if err != nil {
		return err
	}

	err = otherPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Join"}).Click()
	return err
}
