package ws

import (
	"bytes"
	"context"
	"fmt"

	"gitlab.com/hmajid2301/banterbus/internal/entities"
	"gitlab.com/hmajid2301/banterbus/internal/views"
)

// TODO: refactor to another package
type RoomServicer interface {
	Create(ctx context.Context, gameName string, player entities.CreateRoomPlayer) (entities.Room, error)
	Join(ctx context.Context, roomCode string, playerID string, playerNickname string) (entities.Room, error)
}

type CreateRoomEvent struct {
	GameName       string `json:"game_name"`
	PlayerNickname string `json:"player_nickname"`
}

type JoinRoomEvent struct {
	PlayerNickname string `json:"player_nickname"`
	RoomCode       string `json:"room_code"`
}

func (h *CreateRoomEvent) Handle(ctx context.Context, client *client, sub *subscriber) error {
	newPlayer := entities.CreateRoomPlayer{
		ID:       client.playerID,
		Nickname: h.PlayerNickname,
	}
	newRoom, err := sub.roomServicer.Create(ctx, h.GameName, newPlayer)
	if err != nil {
		return err
	}

	room := NewRoom()

	room.addClient(client)
	sub.rooms[newRoom.Code] = room

	go room.runRoom()

	err = sub.updateClients(ctx, newRoom)
	return err
}

func (h *JoinRoomEvent) Handle(ctx context.Context, client *client, sub *subscriber) error {
	room, ok := sub.rooms[h.RoomCode]
	if !ok {
		return fmt.Errorf("room with code %s does not exist", h.RoomCode)
	}
	room.addClient(client)

	updatedRoom, err := sub.roomServicer.Join(ctx, h.RoomCode, client.playerID, h.PlayerNickname)
	if err != nil {
		return err
	}

	err = sub.updateClients(ctx, updatedRoom)
	return err
}

// TODO: refactor to
func (s *subscriber) updateClients(ctx context.Context, updatedRoom entities.Room) error {
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
