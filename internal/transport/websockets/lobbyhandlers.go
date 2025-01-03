package websockets

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/mdobak/go-xerrors"

	"gitlab.com/hmajid2301/banterbus/internal/service"
	"gitlab.com/hmajid2301/banterbus/internal/store/db"
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
	GetLobby(ctx context.Context, playerID uuid.UUID) (service.Lobby, error)
	GetRoomState(ctx context.Context, playerID uuid.UUID) (db.RoomState, error)
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
		if errors.Is(err, service.ErrPlayerAlreadyInRoom) {
			component, err := sub.Reconnect(ctx, client.playerID)
			// TODO: return error back to customer
			if err == nil {
				_ = sub.websocket.Publish(ctx, client.playerID, component.Bytes())
			}
		}

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
	deadline := time.Now().UTC().Add(sub.config.Timings.ShowQuestionScreenFor)
	questionState, err := sub.lobbyService.Start(ctx, s.RoomCode, client.playerID, deadline)
	if err != nil {
		errStr := "failed to start game"
		clientErr := sub.updateClientAboutErr(ctx, client.playerID, errStr)
		return xerrors.Append(clientErr, err)
	}

	showRole := true
	err = sub.updateClientsAboutQuestion(ctx, questionState, showRole)
	if err != nil {
		return err
	}

	go func() {
		v := VotingState{
			gameStateID: questionState.GameStateID,
			subscriber:  *sub,
		}

		deadline := time.Now().UTC().Add(sub.config.Timings.ShowQuestionScreenFor)
		time.Sleep(time.Until(deadline))
		go v.Start(ctx)
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
		return errors.Join(clientErr, err)
	}

	err = sub.updateClientsAboutLobby(ctx, updatedRoom)
	if err != nil {
		return xerrors.New("failed to send kick error message to player", err)
	}

	// TODO: take user back to home page instead of just an error
	err = sub.updateClientAboutErr(ctx, playerToKickID, "you have been kicked from the room")
	return err
}
