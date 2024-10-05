package websockets

import (
	"context"
	"fmt"

	"gitlab.com/hmajid2301/banterbus/internal/entities"
)

type PlayerServicer interface {
	UpdateNickname(ctx context.Context, nickname string, playerID string) (entities.Room, error)
	GenerateNewAvatar(ctx context.Context, playerID string) (entities.Room, error)
	TogglePlayerIsReady(ctx context.Context, playerID string) (entities.Room, error)
}

func (h *UpdateNickname) Handle(ctx context.Context, client *client, sub *Subscriber) error {
	updatedRoom, err := sub.playerService.UpdateNickname(ctx, h.PlayerNickname, client.playerID)
	if err != nil {
		errStr := "failed to update nickname"
		if err == entities.ErrNicknameExists {
			errStr = err.Error()
		}
		clientErr := sub.updateClientAboutErr(ctx, client, errStr)
		return fmt.Errorf("%w: %w", err, clientErr)
	}

	err = sub.updateClientsRoom(ctx, updatedRoom)
	return err
}

func (h *GenerateNewAvatar) Handle(
	ctx context.Context,
	client *client,
	sub *Subscriber,
) error {
	updatedRoom, err := sub.playerService.GenerateNewAvatar(ctx, client.playerID)
	if err != nil {
		errStr := "failed to generate new avatar"
		clientErr := sub.updateClientAboutErr(ctx, client, errStr)
		return fmt.Errorf("%w: %w", err, clientErr)
	}

	err = sub.updateClientsRoom(ctx, updatedRoom)
	return err
}

func (h *TogglePlayerIsReady) Handle(
	ctx context.Context,
	client *client,
	sub *Subscriber,
) error {
	updatedRoom, err := sub.playerService.TogglePlayerIsReady(ctx, client.playerID)
	if err != nil {
		errStr := "failed to update ready status"
		clientErr := sub.updateClientAboutErr(ctx, client, errStr)
		return fmt.Errorf("%w: %w", err, clientErr)
	}

	err = sub.updateClientsRoom(ctx, updatedRoom)
	return err
}
