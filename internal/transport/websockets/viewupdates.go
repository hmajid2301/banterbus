package websockets

import (
	"bytes"
	"context"

	"gitlab.com/hmajid2301/banterbus/internal/service"
	"gitlab.com/hmajid2301/banterbus/internal/views/sections"
)

// TODO: rename these funcs
func (s *Subscriber) updateClientsAboutLobby(ctx context.Context, lobby service.Lobby) error {
	var buf bytes.Buffer
	for _, player := range lobby.Players {
		component := sections.Lobby(lobby.Code, lobby.Players, player)
		err := component.Render(ctx, &buf)
		if err != nil {
			return err
		}

		err = s.websocket.Publish(ctx, player.ID, buf.Bytes())
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Subscriber) updateClientAboutErr(ctx context.Context, playerID string, errStr string) error {
	var buf bytes.Buffer
	component := sections.Error(errStr)
	err := component.Render(ctx, &buf)
	if err != nil {
		return err
	}

	err = s.websocket.Publish(ctx, playerID, buf.Bytes())
	return err
}

func (s *Subscriber) updateClientsAboutQuestion(ctx context.Context, gameState service.GameState) error {
	var buf bytes.Buffer
	for _, player := range gameState.Players {
		component := sections.Question(gameState, player)
		err := component.Render(ctx, &buf)
		if err != nil {
			return err
		}

		err = s.websocket.Publish(ctx, player.ID, buf.Bytes())
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Subscriber) updateClientAboutVoting(ctx context.Context, players []service.VotingPlayer) error {
	var buf bytes.Buffer
	for _, player := range players {
		component := sections.Voting(players, player.ID)
		err := component.Render(ctx, &buf)
		if err != nil {
			return err
		}

		err = s.websocket.Publish(ctx, player.ID, buf.Bytes())
		if err != nil {
			return err
		}
	}

	return nil
}
