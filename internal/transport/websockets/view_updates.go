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
	clientsInRoom := s.rooms[lobby.Code].clients
	for _, player := range lobby.Players {
		client := clientsInRoom[player.ID]
		component := sections.Lobby(lobby.Code, lobby.Players, player)
		err := component.Render(ctx, &buf)
		if err != nil {
			return err
		}
		client.messages <- buf.Bytes()
	}

	return nil
}

func (s *Subscriber) updateClientAboutErr(ctx context.Context, client *client, errStr string) error {
	var buf bytes.Buffer
	component := sections.Error(errStr)
	err := component.Render(ctx, &buf)
	if err != nil {
		return err
	}

	client.messages <- buf.Bytes()
	return nil
}

func (s *Subscriber) updateClientsAboutGame(ctx context.Context, gameState entities.GameState) error {
	var buf bytes.Buffer
	clientsInRoom := s.rooms[gameState.RoomCode].clients
	for _, player := range gameState.Players {
		client := clientsInRoom[player.ID]
		component := sections.Game(gameState, player)
		err := component.Render(ctx, &buf)
		if err != nil {
			return err
		}
		client.messages <- buf.Bytes()
	}

	return nil
}
