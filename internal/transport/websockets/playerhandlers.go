package websockets

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"gitlab.com/hmajid2301/banterbus/internal/service"
)

type PlayerServicer interface {
	UpdateNickname(ctx context.Context, nickname string, playerID uuid.UUID) (service.Lobby, error)
	GenerateNewAvatar(ctx context.Context, playerID uuid.UUID) (service.Lobby, error)
	TogglePlayerIsReady(ctx context.Context, playerID uuid.UUID) (service.Lobby, error)
}

func (u *UpdateNickname) Handle(ctx context.Context, client *client, sub *Subscriber) error {
	updatedRoom, err := sub.playerService.UpdateNickname(ctx, u.PlayerNickname, client.playerID)
	if err != nil {
		errStr := "failed to update nickname"
		if err == service.ErrNicknameExists {
			errStr = err.Error()
		}
		clientErr := sub.updateClientAboutErr(ctx, client.playerID, errStr)
		return errors.Join(clientErr, err)
	}

	err = sub.updateClientsAboutLobby(ctx, updatedRoom)
	return err
}

func (g *GenerateNewAvatar) Handle(ctx context.Context, client *client, sub *Subscriber) error {
	updatedRoom, err := sub.playerService.GenerateNewAvatar(ctx, client.playerID)
	if err != nil {
		errStr := "failed to generate new avatar"
		clientErr := sub.updateClientAboutErr(ctx, client.playerID, errStr)
		return errors.Join(clientErr, err)
	}

	err = sub.updateClientsAboutLobby(ctx, updatedRoom)
	return err
}

// TODO: join errors in all handlers rather than fmt them
func (t *TogglePlayerIsReady) Handle(ctx context.Context, client *client, sub *Subscriber) error {
	updatedRoom, err := sub.playerService.TogglePlayerIsReady(ctx, client.playerID)
	if err != nil {
		errStr := "failed to update ready status"
		clientErr := sub.updateClientAboutErr(ctx, client.playerID, errStr)
		return errors.Join(clientErr, err)
	}

	err = sub.updateClientsAboutLobby(ctx, updatedRoom)
	return err
}
