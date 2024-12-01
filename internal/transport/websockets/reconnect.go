package websockets

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/a-h/templ"

	"gitlab.com/hmajid2301/banterbus/internal/service"
	sqlc "gitlab.com/hmajid2301/banterbus/internal/store/db"
	"gitlab.com/hmajid2301/banterbus/internal/views/sections"
)

const errStr = "failed to reconnect to game"

func (s Subscriber) Reconnect(ctx context.Context, client *client) error {
	s.logger.DebugContext(ctx, "attempting to reconnect player", slog.String("player_id", client.playerID))
	roomState, err := s.playerService.GetRoomState(ctx, client.playerID)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	var component templ.Component
	switch roomState {
	case sqlc.ROOMSTATE_CREATED:
		lobby, err := s.playerService.GetLobby(ctx, client.playerID)
		if err != nil {
			clientErr := s.updateClientAboutErr(ctx, client.playerID, errStr)
			return errors.Join(clientErr, err)
		}

		var mePlayer service.LobbyPlayer
		for _, player := range lobby.Players {
			if player.ID == client.playerID {
				mePlayer = player
			}
		}

		component = sections.Lobby(lobby.Code, lobby.Players, mePlayer)
	case sqlc.ROOMSTATE_PLAYING:
		component, err = s.reconnectToPlayingGame(ctx, client)
		if err != nil {
			return err
		}
	case sqlc.ROOMSTATE_PAUSED:
		return fmt.Errorf("cannot reconnect game to paused game, as this is not implemented")
	case sqlc.ROOMSTATE_ABANDONED:
		return fmt.Errorf("cannot reconnect game is abandoned")
	case sqlc.ROOMSTATE_FINISHED:
		return fmt.Errorf("cannot reconnect game is finished")
	default:
		return fmt.Errorf("unknown room state: %s", roomState)
	}

	err = component.Render(ctx, &buf)
	if err != nil {
		return err
	}

	err = s.websocket.Publish(ctx, client.playerID, buf.Bytes())
	return err
}

func (s Subscriber) reconnectToPlayingGame(ctx context.Context, client *client) (templ.Component, error) {
	var component templ.Component
	gameState, err := s.playerService.GetGameState(ctx, client.playerID)
	if err != nil {
		clientErr := s.updateClientAboutErr(ctx, client.playerID, errStr)
		return component, errors.Join(clientErr, err)
	}

	switch gameState {
	case sqlc.GAMESTATE_FIBBING_IT_SHOW_QUESTION:
		question, err := s.playerService.GetQuestionState(ctx, client.playerID)
		if err != nil {
			clientErr := s.updateClientAboutErr(ctx, client.playerID, errStr)
			return component, errors.Join(clientErr, err)
		}
		component = sections.Question(question, question.Players[0])
	case sqlc.GAMESTATE_FIBBING_IT_VOTING:
		voting, err := s.playerService.GetVotingState(ctx, client.playerID)
		if err != nil {
			clientErr := s.updateClientAboutErr(ctx, client.playerID, errStr)
			return component, errors.Join(clientErr, err)
		}
		component = sections.Voting(voting, client.playerID)
	default:
		return component, fmt.Errorf("unknown game state: %s", gameState)
	}

	return component, nil
}
