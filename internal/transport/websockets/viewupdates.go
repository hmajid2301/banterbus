package websockets

import (
	"bytes"
	"context"

	"gitlab.com/hmajid2301/banterbus/internal/entities"
	"gitlab.com/hmajid2301/banterbus/internal/views/sections"
)

// TODO: rename these funcs
func (s *Subscriber) updateClientsAboutLobby(ctx context.Context, lobby entities.Lobby) error {
	var buf bytes.Buffer
	for _, player := range lobby.Players {
		component := sections.Lobby(lobby.Code, lobby.Players, player)
		err := component.Render(ctx, &buf)
		if err != nil {
			return err
		}

		err = s.websocket.Publish(ctx, player.ID, buf.Bytes())
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Subscriber) updateClientAboutErr(ctx context.Context, playerID string, errStr string) error {
	var buf bytes.Buffer
	component := sections.Error(errStr)
	err := component.Render(ctx, &buf)
	if err != nil {
		return err
	}

	err = s.websocket.Publish(ctx, playerID, buf.Bytes())
	return err
}

func (s *Subscriber) updateClientsAboutGame(ctx context.Context, gameState entities.GameState) error {
	var buf bytes.Buffer
	for _, player := range gameState.Players {
		component := sections.Game(gameState, player)
		err := component.Render(ctx, &buf)
		if err != nil {
			return err
		}

		err = s.websocket.Publish(ctx, player.ID, buf.Bytes())
		if err != nil {
			return err
		}
	}

	return nil
}
