package websockets

import (
	"context"

	"gitlab.com/hmajid2301/banterbus/internal/entities"
)

type PlayerServicer interface {
	UpdateNickname(ctx context.Context, nickname string, playerID string) (entities.Room, error)
	GenerateNewAvatar(ctx context.Context, playerID string) (entities.Room, error)
}

type UpdateNickname struct {
	PlayerNickname string `mapstructure:"player_nickname"`
	PlayerID       string `mapstructure:"player_id"`
}

type GenerateNewAvatar struct {
	PlayerID string `mapstructure:"player_id"`
}

func (h *UpdateNickname) Handle(ctx context.Context, client *client, sub *Subscriber) error {
	updatedRoom, err := sub.playerServicer.UpdateNickname(ctx, h.PlayerNickname, h.PlayerID)
	if err != nil {
		return err
	}

	err = sub.updateClients(ctx, updatedRoom)
	return err
}

func (h *GenerateNewAvatar) Handle(
	ctx context.Context,
	client *client,
	sub *Subscriber,
) error {
	updatedRoom, err := sub.playerServicer.GenerateNewAvatar(ctx, h.PlayerID)
	if err != nil {
		return err
	}

	err = sub.updateClients(ctx, updatedRoom)
	return err
}
