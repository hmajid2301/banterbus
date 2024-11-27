package websockets

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/hmajid2301/banterbus/internal/service"
)

const votingDelay = 60 * time.Second

type LobbyServicer interface {
	Create(ctx context.Context, gameName string, player service.NewHostPlayer) (service.Lobby, error)
	Join(ctx context.Context, roomCode string, playerID string, playerNickname string) (service.Lobby, error)
	Start(ctx context.Context, roomCode string, playerID string) (service.GameState, error)
	KickPlayer(
		ctx context.Context,
		roomCode string,
		playerID string,
		playerNicknameToKick string,
	) (service.Lobby, string, error)
}

func (c *CreateRoom) Handle(ctx context.Context, client *client, sub *Subscriber) error {
	newPlayer := service.NewHostPlayer{
		ID:       client.playerID,
		Nickname: c.PlayerNickname,
	}
	lobby, err := sub.lobbyService.Create(ctx, c.GameName, newPlayer)
	if err != nil {
		errStr := "failed to create room"
		clientErr := sub.updateClientAboutErr(ctx, client.playerID, errStr)
		return fmt.Errorf("%w: %w", err, clientErr)
	}

	err = sub.updateClientsAboutLobby(ctx, lobby)
	return err
}

func (j *JoinLobby) Handle(ctx context.Context, client *client, sub *Subscriber) error {
	updatedRoom, err := sub.lobbyService.Join(ctx, j.RoomCode, client.playerID, j.PlayerNickname)
	if err != nil {
		errStr := "failed to join room"
		if err == service.ErrNicknameExists {
			errStr = err.Error()
		}
		clientErr := sub.updateClientAboutErr(ctx, client.playerID, errStr)
		return fmt.Errorf("%w: %w", err, clientErr)
	}

	err = sub.updateClientsAboutLobby(ctx, updatedRoom)
	return err
}

func (s *StartGame) Handle(ctx context.Context, client *client, sub *Subscriber) error {
	updatedRoom, err := sub.lobbyService.Start(ctx, s.RoomCode, client.playerID)
	if err != nil {
		errStr := "failed to start game"
		clientErr := sub.updateClientAboutErr(ctx, client.playerID, errStr)
		return fmt.Errorf("%w: %w", err, clientErr)
	}
	err = sub.updateClientsAboutGame(ctx, updatedRoom)
	if err != nil {
		return err
	}

	// TODO: we want to start a state machine, as everything will be time based started by backen by backend.
	go func() {
		time.Sleep(votingDelay)

		var players []service.VotingPlayer
		for _, player := range updatedRoom.Players {
			players = append(players, service.VotingPlayer{
				ID:       player.ID,
				Nickname: player.Nickname,
				Avatar:   player.Avatar,
			})
		}

		err = sub.updateClientAboutVoting(ctx, players)
		if err != nil {
			// TODO: log error
			fmt.Printf("failed to update clients about game: %v\n", err)
		}
	}()

	return nil
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
		clientErr := sub.updateClientAboutErr(ctx, client.playerID, errStr)
		return fmt.Errorf("%w: %w", err, clientErr)
	}

	err = sub.updateClientsAboutLobby(ctx, updatedRoom)
	if err != nil {
		return fmt.Errorf("failed to send kick error message to player: %w", err)
	}

	// TODO: take user back to home page instead of just an error
	err = sub.updateClientAboutErr(ctx, playerToKickID, "you have been kicked from the room")
	return err
}
