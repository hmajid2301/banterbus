package websockets

import (
	"context"
	"fmt"

	"gitlab.com/hmajid2301/banterbus/internal/entities"
)

type LobbyServicer interface {
	Create(ctx context.Context, gameName string, player entities.NewHostPlayer) (entities.Lobby, error)
	Join(ctx context.Context, roomCode string, playerID string, playerNickname string) (entities.Lobby, error)
	Start(ctx context.Context, roomCode string, playerID string) (entities.GameState, error)
	KickPlayer(
		ctx context.Context,
		roomCode string,
		playerID string,
		playerNicknameToKick string,
	) (entities.Lobby, string, error)
}

func (c *CreateRoom) Handle(ctx context.Context, client *client, sub *Subscriber) error {
	newPlayer := entities.NewHostPlayer{
		ID:       client.playerID,
		Nickname: c.PlayerNickname,
	}
	lobby, err := sub.lobbyService.Create(ctx, c.GameName, newPlayer)
	if err != nil {
		errStr := "failed to create room"
		clientErr := sub.updateClientAboutErr(ctx, client, errStr)
		return fmt.Errorf("%w: %w", err, clientErr)
	}

	room := newRoom()

	room.addClient(client)
	sub.rooms[lobby.Code] = room

	go room.runRoom()

	err = sub.updateClientsAboutLobby(ctx, lobby)
	return err
}

func (j *JoinLobby) Handle(ctx context.Context, client *client, sub *Subscriber) error {
	room, ok := sub.rooms[j.RoomCode]
	if !ok {
		err := fmt.Errorf("room with code %s does not exist", j.RoomCode)
		clientErr := sub.updateClientAboutErr(ctx, client, err.Error())
		return fmt.Errorf("%w: %w", err, clientErr)
	}
	room.addClient(client)

	updatedRoom, err := sub.lobbyService.Join(ctx, j.RoomCode, client.playerID, j.PlayerNickname)
	if err != nil {
		errStr := "failed to join room"
		if err == entities.ErrNicknameExists {
			errStr = err.Error()
		}
		clientErr := sub.updateClientAboutErr(ctx, client, errStr)
		return fmt.Errorf("%w: %w", err, clientErr)
	}

	err = sub.updateClientsAboutLobby(ctx, updatedRoom)
	return err
}

func (s *StartGame) Handle(ctx context.Context, client *client, sub *Subscriber) error {
	updatedRoom, err := sub.lobbyService.Start(ctx, s.RoomCode, client.playerID)
	if err != nil {
		errStr := "failed to start game"
		clientErr := sub.updateClientAboutErr(ctx, client, errStr)
		return fmt.Errorf("%w: %w", err, clientErr)
	}

	err = sub.updateClientsAboutGame(ctx, updatedRoom)
	return err
}

func (k *KickPlayer) Handle(ctx context.Context, client *client, sub *Subscriber) error {
	updatedRoom, playerToKickID, err := sub.lobbyService.KickPlayer(
		ctx,
		k.RoomCode,
		client.playerID,
		k.PlayerNicknameToKick,
	)
	if err != nil {
		errStr := "failed to kick player"
		clientErr := sub.updateClientAboutErr(ctx, client, errStr)
		return fmt.Errorf("%w: %w", err, clientErr)
	}

	room, ok := sub.rooms[k.RoomCode]
	if !ok {
		err := fmt.Errorf("room with code %s does not exist", k.RoomCode)
		clientErr := sub.updateClientAboutErr(ctx, client, err.Error())
		return fmt.Errorf("%w: %w", err, clientErr)
	}

	playerKickedClient, err := room.getClient(playerToKickID)
	if err != nil {
		errStr := "failed to kick player"
		clientErr := sub.updateClientAboutErr(ctx, client, errStr)
		return fmt.Errorf("%w: %w", err, clientErr)
	}

	room.removeClient(playerToKickID)
	err = sub.updateClientsAboutLobby(ctx, updatedRoom)
	if err != nil {
		return fmt.Errorf("failed to send kick error message to player: %w", err)
	}

	// TODO: take user back to home page instead of just an error
	err = sub.updateClientAboutErr(ctx, playerKickedClient, "you have been kicked from the room")
	return err
}
