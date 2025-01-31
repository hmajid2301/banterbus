package websockets

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"

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
		component := sections.Lobby(lobby.Code, lobby.Players, player, s.rules)
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

func (s *Subscriber) updateClientsAboutErr(ctx context.Context, playerIDs []uuid.UUID, errStr string) error {
	var err error
	for _, playerID := range playerIDs {
		clientErr := s.updateClientAboutErr(ctx, playerID, errStr)
		err = errors.Join(err, clientErr)
	}

	return err
}

func (s *Subscriber) updateClientAboutErr(ctx context.Context, playerID uuid.UUID, errStr string) error {
	span := trace.SpanFromContext(ctx)
	spanID := span.SpanContext().SpanID().String()

	errWithID := fmt.Sprintf("%s. Correleation ID: %s", errStr, spanID)

	t := Toast{Message: errWithID, Type: "failure"}
	toastJSON, err := json.Marshal(t)
	if err != nil {
		return err
	}

	err = s.websocket.Publish(ctx, playerID, toastJSON)
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

	// TODO: should this be moved else where?
	maxScore := 0
	for _, player := range scoreState.Players {
		if player.Score > maxScore {
			maxScore = player.Score
		}
	}

	for _, player := range scoreState.Players {
		component := sections.Score(scoreState, player, maxScore)
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

	maxScore := 0
	for _, player := range winnerState.Players {
		if player.Score > maxScore {
			maxScore = player.Score
		}
	}

	for _, player := range winnerState.Players {
		component := sections.Winner(winnerState, maxScore)
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
