package websockets

import (
	"bytes"
	"context"

	"github.com/google/uuid"

	"gitlab.com/hmajid2301/banterbus/internal/service"
	"gitlab.com/hmajid2301/banterbus/internal/views/sections"
)

type Toast struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

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

func (s *Subscriber) updateClientAboutErr(ctx context.Context, playerID uuid.UUID, errStr string) error {
	var buf bytes.Buffer
	component := sections.Error(errStr)
	err := component.Render(ctx, &buf)
	if err != nil {
		return err
	}

	err = s.websocket.Publish(ctx, playerID, buf.Bytes())
	return err
}

func (s *Subscriber) updateClientsAboutQuestion(
	ctx context.Context,
	gameState service.QuestionState,
	showModal bool,
) error {
	var buf bytes.Buffer
	for _, player := range gameState.Players {
		component := sections.Question(gameState, player, showModal)
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

func (s *Subscriber) updateClientsAboutVoting(ctx context.Context, votingState service.VotingState) error {
	var buf bytes.Buffer
	for _, player := range votingState.Players {
		component := sections.Voting(votingState, player)
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

func (s *Subscriber) updateClientsAboutReveal(ctx context.Context, revealState service.RevealRoleState) error {
	var buf bytes.Buffer
	for _, id := range revealState.PlayerIDs {
		component := sections.Reveal(revealState)
		err := component.Render(ctx, &buf)
		if err != nil {
			return err
		}

		err = s.websocket.Publish(ctx, id, buf.Bytes())
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Subscriber) updateClientsAboutScore(ctx context.Context, scoreState service.ScoreState) error {
	var buf bytes.Buffer
	for _, player := range scoreState.Players {
		component := sections.Score(scoreState, player)
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

func (s *Subscriber) updateClientsAboutWinner(ctx context.Context, winnerState service.WinnerState) error {
	var buf bytes.Buffer
	for _, player := range winnerState.Players {
		component := sections.Winner(winnerState)
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
