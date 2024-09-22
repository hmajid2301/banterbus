package e2e

import (
	"strings"

	"github.com/playwright-community/playwright-go"
)

func joinRoom(hostPlayerPage playwright.Page, otherPlayerPage playwright.Page) error {
	err := hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Create Room"}).Click()
	if err != nil {
		return err
	}

	locator := hostPlayerPage.Locator("text=Code:")
	textContent, err := locator.TextContent()
	if err != nil {
		return err
	}

	code := strings.Replace(textContent, "Code: ", "", 1)
	err = otherPlayerPage.Locator("#join_room_form").GetByPlaceholder("Enter your room code").Fill(code)
	if err != nil {
		return err
	}

	err = otherPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Join Room"}).Click()
	return err
}
