package websockets

import (
	"bytes"
	"context"
	"fmt"

	"gitlab.com/hmajid2301/banterbus/internal/entities"
	"gitlab.com/hmajid2301/banterbus/internal/transport/websockets/views"
)

// TODO: refactor to another package
type RoomServicer interface {
	Create(ctx context.Context, gameName string, player entities.NewHostPlayer) (entities.Room, error)
	Join(ctx context.Context, roomCode string, playerID string, playerNickname string) (entities.Room, error)
	Start(ctx context.Context, roomCode string, playerID string) (entities.Room, error)
}

type CreateRoom struct {
	GameName       string `json:"game_name"`
	PlayerNickname string `json:"player_nickname"`
}

type JoinRoom struct {
	PlayerNickname string `json:"player_nickname"`
	RoomCode       string `json:"room_code"`
}

type StartGame struct {
	RoomCode string `json:"room_code"`
	PlayerID string `json:"player_id"`
}

func (h *CreateRoom) Handle(ctx context.Context, client *client, sub *Subscriber) error {
	newPlayer := entities.NewHostPlayer{
		ID:       client.playerID,
		Nickname: h.PlayerNickname,
	}
	newRoom, err := sub.roomServicer.Create(ctx, h.GameName, newPlayer)
	if err != nil {
		errStr := "failed to create room"
		clientErr := sub.updateClientAboutErr(ctx, client, errStr)
		return fmt.Errorf("%w: %w", err, clientErr)
	}

	room := NewRoom()

	room.addClient(client)
	sub.rooms[newRoom.Code] = room

	go room.runRoom()

	err = sub.updateClientsRoom(ctx, newRoom)
	return err
}

func (h *JoinRoom) Handle(ctx context.Context, client *client, sub *Subscriber) error {
	room, ok := sub.rooms[h.RoomCode]
	if !ok {
		err := fmt.Errorf("room with code %s does not exist", h.RoomCode)
		clientErr := sub.updateClientAboutErr(ctx, client, err.Error())
		return fmt.Errorf("%w: %w", err, clientErr)
	}
	room.addClient(client)

	updatedRoom, err := sub.roomServicer.Join(ctx, h.RoomCode, client.playerID, h.PlayerNickname)
	if err != nil {
		errStr := "failed to join room"
		if err == entities.ErrNicknameExists {
			errStr = err.Error()
		}
		clientErr := sub.updateClientAboutErr(ctx, client, errStr)
		return fmt.Errorf("%w: %w", err, clientErr)
	}

	err = sub.updateClientsRoom(ctx, updatedRoom)
	return err
}

func (h *StartGame) Handle(ctx context.Context, client *client, sub *Subscriber) error {
	updatedRoom, err := sub.roomServicer.Start(ctx, h.RoomCode, client.playerID)
	if err != nil {
		errStr := "failed to start game"
		clientErr := sub.updateClientAboutErr(ctx, client, errStr)
		return fmt.Errorf("%w: %w", err, clientErr)
	}

	err = sub.updateClientsGame(ctx, updatedRoom)
	return err
}

// TODO: refactor to another file
// TODO: rename Room to lobby
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
