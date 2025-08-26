package websockets

import (
	"context"
	"errors"

	"github.com/gofrs/uuid/v5"

	"gitlab.com/hmajid2301/banterbus/internal/service"
	"gitlab.com/hmajid2301/banterbus/internal/telemetry"
)

type PlayerServicer interface {
	UpdateNickname(ctx context.Context, nickname string, playerID uuid.UUID) (service.Lobby, error)
	GenerateNewAvatar(ctx context.Context, playerID uuid.UUID) (service.Lobby, error)
	TogglePlayerIsReady(ctx context.Context, playerID uuid.UUID) (service.Lobby, error)
	UpdateLocale(ctx context.Context, playerID uuid.UUID, locale string) error
}

func (u *UpdateNickname) Handle(ctx context.Context, client *Client, sub *Subscriber) error {
	telemetry.AddPlayerActionAttributes(ctx, client.playerID.String(), "update_nickname", false, false)

	updatedRoom, err := sub.playerService.UpdateNickname(ctx, u.PlayerNickname, client.playerID)
	if err != nil {
		telemetry.RecordBusinessLogicError(ctx, "update_nickname", err.Error(), telemetry.GameContext{
			PlayerID: &client.playerID,
		})
		errStr := "Failed to update nickname"
		if err == service.ErrNicknameExists {
			errStr = err.Error()
		}
		clientErr := sub.updateClientAboutErr(ctx, client.playerID, errStr)
		return errors.Join(clientErr, err)
	}

	telemetry.AddRoomStateAttributes(ctx, "Created", updatedRoom.Code, len(updatedRoom.Players))

	err = sub.updateClientsAboutLobby(ctx, updatedRoom)
	return err
}

func (g *GenerateNewAvatar) Handle(ctx context.Context, client *Client, sub *Subscriber) error {
	updatedRoom, err := sub.playerService.GenerateNewAvatar(ctx, client.playerID)
	if err != nil {
		errStr := "Failed to generate new avatar"
		clientErr := sub.updateClientAboutErr(ctx, client.playerID, errStr)
		return errors.Join(clientErr, err)
	}

	err = sub.updateClientsAboutLobby(ctx, updatedRoom)
	return err
}

// TODO: join errors in all handlers rather than fmt them
func (t *TogglePlayerIsReady) Handle(ctx context.Context, client *Client, sub *Subscriber) error {
	telemetry.AddPlayerActionAttributes(ctx, client.playerID.String(), "toggle_ready", false, false)

	updatedRoom, err := sub.playerService.TogglePlayerIsReady(ctx, client.playerID)
	if err != nil {
		telemetry.RecordBusinessLogicError(ctx, "toggle_ready", err.Error(), telemetry.GameContext{
			PlayerID: &client.playerID,
		})
		errStr := "Failed to toggle ready status"
		clientErr := sub.updateClientAboutErr(ctx, client.playerID, errStr)
		return errors.Join(clientErr, err)
	}

	var isReady bool
	for _, player := range updatedRoom.Players {
		if player.ID == client.playerID {
			isReady = player.IsReady
			break
		}
	}

	telemetry.AddPlayerActionAttributes(ctx, client.playerID.String(), "toggle_ready_completed", false, isReady)
	telemetry.AddRoomStateAttributes(ctx, "Created", updatedRoom.Code, len(updatedRoom.Players))

	err = sub.updateClientsAboutLobby(ctx, updatedRoom)
	return err
}
