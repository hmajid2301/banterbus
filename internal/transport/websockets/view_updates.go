package websockets

import (
	"bytes"
	"context"

	"gitlab.com/hmajid2301/banterbus/internal/entities"
	"gitlab.com/hmajid2301/banterbus/internal/transport/websockets/views"
)

func (s *Subscriber) updateClientsRoom(ctx context.Context, updatedRoom entities.Room) error {
	var buf bytes.Buffer
	clientsInRoom := s.rooms[updatedRoom.Code].clients
	for _, player := range updatedRoom.Players {
		client := clientsInRoom[player.ID]
		component := views.Room(updatedRoom.Code, updatedRoom.Players, player)
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
	component := views.Error(errStr)
	err := component.Render(ctx, &buf)
	if err != nil {
		return err
	}

	client.messages <- buf.Bytes()
	return nil
}

func (s *Subscriber) updateClientsGame(ctx context.Context, updatedRoom entities.Room) error {
	var buf bytes.Buffer
	clientsInRoom := s.rooms[updatedRoom.Code].clients
	for _, player := range updatedRoom.Players {
		client := clientsInRoom[player.ID]
		component := views.Game(updatedRoom.Players, player)
		err := component.Render(ctx, &buf)
		if err != nil {
			return err
		}
		client.messages <- buf.Bytes()

	}
	return nil
}
