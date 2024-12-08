package websockets

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"gitlab.com/hmajid2301/banterbus/internal/config"
	"gitlab.com/hmajid2301/banterbus/internal/service"
)

type LobbyServicer interface {
	Create(ctx context.Context, gameName string, player service.NewHostPlayer) (service.Lobby, error)
	Join(ctx context.Context, roomCode string, playerID uuid.UUID, playerNickname string) (service.Lobby, error)
	Start(ctx context.Context, roomCode string, playerID uuid.UUID, deadline time.Time) (service.QuestionState, error)
	KickPlayer(
		ctx context.Context,
		roomCode string,
		playerID uuid.UUID,
		playerNicknameToKick string,
	) (service.Lobby, uuid.UUID, error)
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
		return errors.Join(clientErr, err)
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
		return errors.Join(clientErr, err)
	}

	err = sub.updateClientsAboutLobby(ctx, updatedRoom)
	return err
}

func (s *StartGame) Handle(ctx context.Context, client *client, sub *Subscriber) error {
	deadline := time.Now().UTC().Add(config.ShowQuestionScreenFor)
	questionState, err := sub.lobbyService.Start(ctx, s.RoomCode, client.playerID, deadline)
	if err != nil {
		errStr := "failed to start game"
		clientErr := sub.updateClientAboutErr(ctx, client.playerID, errStr)
		return errors.Join(clientErr, err)
	}

	showRole := true
	err = sub.updateClientsAboutQuestion(ctx, questionState, showRole)
	if err != nil {
		return err
	}

	time.Sleep(config.ShowQuestionScreenFor)

	// TODO: we want to start a state machine, as everything will be time based started by backen by backend.
	go MoveToVoting(ctx, sub, questionState.Players, questionState.GameStateID, questionState.Round)

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
		return errors.Join(clientErr, err)
	}

	err = sub.updateClientsAboutLobby(ctx, updatedRoom)
	if err != nil {
		return fmt.Errorf("failed to send kick error message to player: %w", err)
	}

	// TODO: take user back to home page instead of just an error
	err = sub.updateClientAboutErr(ctx, playerToKickID, "you have been kicked from the room")
	return err
}
