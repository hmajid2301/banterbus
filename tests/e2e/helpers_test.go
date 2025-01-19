package e2e

import (
	"fmt"
	"time"

	"github.com/mdobak/go-xerrors"
	"github.com/playwright-community/playwright-go"
)

func joinRoom(hostPlayerPage playwright.Page, otherPlayerPages []playwright.Page) error {
	err := hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Start"}).Click()
	if err != nil {
		return err
	}

	locator := hostPlayerPage.Locator("input[name='room_code']")

	// INFO: Retry mechanism to wait for the room code to be available
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
		return xerrors.New("room code is empty")
	}

	for _, player := range otherPlayerPages {
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

func startGame(hostPlayerPage playwright.Page, otherPlayerPages []playwright.Page) error {
	err := joinRoom(hostPlayerPage, otherPlayerPages)
	if err != nil {
		return err
	}

	for _, player := range append(otherPlayerPages, hostPlayerPage) {
		err = player.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Ready"}).Click()
		if err != nil {
			return err
		}
	}

	err = hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Start Game"}).Click()
	if err != nil {
		return err
	}

	return nil
}

func startAndGetRoles(
	hostPlayerPage playwright.Page,
	otherPlayerPages []playwright.Page,
) (playwright.Page, []playwright.Page, error) {
	var fibber playwright.Page
	normals := []playwright.Page{}
	err := startGame(hostPlayerPage, otherPlayerPages)
	if err != nil {
		return fibber, normals, err
	}

	submitButton := hostPlayerPage.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Submit Answer"})
	err = submitButton.WaitFor()
	if err != nil {
		return fibber, normals, fmt.Errorf("failed to find 'Submit Answer' button: %w", err)
	}

	fibberFound := false
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

	if !fibberFound {
		return fibber, normals, fmt.Errorf("failed to find fibber")
	}

	return fibber, normals, nil
}
