package websockets

import (
	"context"
	"fmt"

	"gitlab.com/hmajid2301/banterbus/internal/entities"
)

type LobbyServicer interface {
	Create(ctx context.Context, gameName string, player entities.NewHostPlayer) (entities.Room, error)
	Join(ctx context.Context, roomCode string, playerID string, playerNickname string) (entities.Room, error)
	Start(ctx context.Context, roomCode string, playerID string) (entities.Room, error)
}

func (h *CreateRoom) Handle(ctx context.Context, client *client, sub *Subscriber) error {
	newPlayer := entities.NewHostPlayer{
		ID:       client.playerID,
		Nickname: h.PlayerNickname,
	}
	newRoom, err := sub.lobbyService.Create(ctx, h.GameName, newPlayer)
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

func (h *JoinLobby) Handle(ctx context.Context, client *client, sub *Subscriber) error {
	room, ok := sub.rooms[h.RoomCode]
	if !ok {
		err := fmt.Errorf("room with code %s does not exist", h.RoomCode)
		clientErr := sub.updateClientAboutErr(ctx, client, err.Error())
		return fmt.Errorf("%w: %w", err, clientErr)
	}
	room.addClient(client)

	updatedRoom, err := sub.lobbyService.Join(ctx, h.RoomCode, client.playerID, h.PlayerNickname)
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
	updatedRoom, err := sub.lobbyService.Start(ctx, h.RoomCode, client.playerID)
	if err != nil {
		errStr := "failed to start game"
		clientErr := sub.updateClientAboutErr(ctx, client, errStr)
		return fmt.Errorf("%w: %w", err, clientErr)
	}

	err = sub.updateClientsGame(ctx, updatedRoom)
	return err
}
