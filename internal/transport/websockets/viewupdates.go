package websockets

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/invopop/ctxi18n"
	"go.opentelemetry.io/otel/trace"

	"gitlab.com/hmajid2301/banterbus/internal/service"
	"gitlab.com/hmajid2301/banterbus/internal/statemachine"
	"gitlab.com/hmajid2301/banterbus/internal/views/sections"
)

type Toast struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

// TODO: This function makes a DB call for every player on every view update, which can cause performance issues.
// Consider implementing one of these optimizations:
// 1. Cache player locales in memory (with TTL or invalidation on locale updates)
// 2. Batch fetch all player locales at once before rendering
// 3. Pass locale information through the state objects instead of querying here
// 4. Store locale in Redis alongside player session data
func (s *Subscriber) getContextWithPlayerLocale(ctx context.Context, playerID uuid.UUID) context.Context {
	player, err := s.playerService.GetPlayerByID(ctx, playerID)
	if err != nil {
		if !errors.Is(err, service.ErrPlayerNotFound) {
			s.logger.DebugContext(ctx, "failed to get player for locale",
				slog.String("player_id", playerID.String()),
				slog.Any("error", err))
		}
		ctx, _ = ctxi18n.WithLocale(ctx, s.config.App.DefaultLocale.String())
		return ctx
	}

	if player.Locale.Valid && player.Locale.String != "" {
		localeCtx, err := ctxi18n.WithLocale(ctx, player.Locale.String)
		if err != nil {
			s.logger.DebugContext(ctx, "failed to set locale for player",
				slog.String("player_id", playerID.String()),
				slog.String("locale", player.Locale.String),
				slog.Any("error", err))
			ctx, _ = ctxi18n.WithLocale(ctx, s.config.App.DefaultLocale.String())
			return ctx
		}
		return localeCtx
	}

	ctx, _ = ctxi18n.WithLocale(ctx, s.config.App.DefaultLocale.String())
	return ctx
}

func (s *Subscriber) updateClientsAboutLobby(ctx context.Context, lobby service.Lobby) error {
	for _, player := range lobby.Players {
		playerCtx := s.getContextWithPlayerLocale(ctx, player.ID)

		var buf bytes.Buffer
		component := sections.Lobby(lobby.Code, lobby.Players, player, s.rules)
		err := component.Render(playerCtx, &buf)
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
	traceID := span.SpanContext().TraceID().String()

	errWithID := fmt.Sprintf("%s. Correleation ID: %s", errStr, traceID)

	t := Toast{Message: errWithID, Type: "failure"}
	toastJSON, err := json.Marshal(t)
	if err != nil {
		return err
	}

	err = s.websocket.Publish(ctx, playerID, toastJSON)
	return err
}

func (s *Subscriber) UpdateClientsAboutQuestion(
	ctx context.Context,
	gameState service.QuestionState,
	showModal bool,
) error {
	for _, player := range gameState.Players {
		playerCtx := s.getContextWithPlayerLocale(ctx, player.ID)

		var buf bytes.Buffer
		component := sections.Question(gameState, player, showModal)
		err := component.Render(playerCtx, &buf)
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

func (s *Subscriber) UpdateClientsAboutVoting(ctx context.Context, votingState service.VotingState) error {
	for _, player := range votingState.Players {
		playerCtx := s.getContextWithPlayerLocale(ctx, player.ID)

		var buf bytes.Buffer
		component := sections.Voting(votingState, player)
		err := component.Render(playerCtx, &buf)
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

func (s *Subscriber) UpdateClientsAboutReveal(ctx context.Context, revealState service.RevealRoleState) error {
	for _, id := range revealState.PlayerIDs {
		playerCtx := s.getContextWithPlayerLocale(ctx, id)

		var buf bytes.Buffer
		component := sections.Reveal(revealState)
		err := component.Render(playerCtx, &buf)
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

func (s *Subscriber) UpdateClientsAboutScore(ctx context.Context, scoreState service.ScoreState) error {
	maxScore := 0
	for _, player := range scoreState.Players {
		if player.Score > maxScore {
			maxScore = player.Score
		}
	}

	for _, player := range scoreState.Players {
		playerCtx := s.getContextWithPlayerLocale(ctx, player.ID)

		var buf bytes.Buffer
		component := sections.Score(scoreState, player, maxScore)
		err := component.Render(playerCtx, &buf)
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

func (s *Subscriber) UpdateClientsAboutWinner(ctx context.Context, winnerState service.WinnerState) error {
	maxScore := 0
	for _, player := range winnerState.Players {
		if player.Score > maxScore {
			maxScore = player.Score
		}
	}

	for _, player := range winnerState.Players {
		playerCtx := s.getContextWithPlayerLocale(ctx, player.ID)

		var buf bytes.Buffer
		component := sections.Winner(winnerState, maxScore)
		err := component.Render(playerCtx, &buf)
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

func (s *Subscriber) updateClientsAboutPause(ctx context.Context, pauseStatus service.PauseStatus, gameStateID uuid.UUID) error {
	players, err := s.roundService.GetAllPlayersByGameStateID(ctx, gameStateID)
	if err != nil {
		return err
	}

	for _, player := range players {
		message := "Game paused by host"
		err := s.updateClientAboutErr(ctx, player.ID, message)
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to send pause notification",
				slog.String("player_id", player.ID.String()),
				slog.Any("error", err))
		}
	}

	return nil
}

func (s *Subscriber) updateClientsAboutResume(ctx context.Context, pauseStatus service.PauseStatus, gameStateID uuid.UUID) error {
	players, err := s.roundService.GetAllPlayersByGameStateID(ctx, gameStateID)
	if err != nil {
		return err
	}

	for _, player := range players {
		message := "Game resumed"
		err := s.updateClientAboutErr(ctx, player.ID, message)
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to send resume notification",
				slog.String("player_id", player.ID.String()),
				slog.Any("error", err))
		}
	}

	return nil
}

func (s *Subscriber) restartStateMachineAfterResume(ctx context.Context, gameStateID uuid.UUID, state string, submitDeadline time.Time) error {
	s.logger.InfoContext(ctx, "restarting state machine after resume",
		slog.String("game_state_id", gameStateID.String()),
		slog.String("state", state),
		slog.Time("submit_deadline", submitDeadline))

	remaining := time.Until(submitDeadline)
	if remaining <= 0 {
		s.logger.WarnContext(ctx, "deadline already passed, not restarting state machine",
			slog.String("game_state_id", gameStateID.String()),
			slog.Duration("remaining", remaining))
		return nil
	}

	deps, err := s.NewStateDependencies()
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to build state dependencies",
			slog.Any("error", err),
			slog.String("game_state_id", gameStateID.String()))
		return err
	}

	switch state {
	case "FibbingITQuestion":
		deps.Timings.ShowQuestionScreenFor = remaining
	case "FibbingItVoting":
		deps.Timings.ShowVotingScreenFor = remaining
	case "FibbingItReveal":
		deps.Timings.ShowRevealScreenFor = remaining
	case "FibbingItScoring":
		deps.Timings.ShowScoreScreenFor = remaining
	case "FibbingItWinner":
		deps.Timings.ShowWinnerScreenFor = remaining
	}

	var stateMachine statemachine.State
	switch state {
	case "FibbingITQuestion":
		stateMachine, err = statemachine.NewQuestionState(gameStateID, false, deps)
	case "FibbingItVoting":
		stateMachine, err = statemachine.NewVotingState(gameStateID, deps)
	case "FibbingItReveal":
		stateMachine, err = statemachine.NewRevealState(gameStateID, deps)
	case "FibbingItScoring":
		stateMachine, err = statemachine.NewScoringState(gameStateID, deps)
	case "FibbingItWinner":
		stateMachine, err = statemachine.NewWinnerState(gameStateID, deps)
	default:
		s.logger.WarnContext(ctx, "cannot restart state machine for state",
			slog.String("state", state),
			slog.String("game_state_id", gameStateID.String()))
		return nil
	}

	if err != nil {
		s.logger.ErrorContext(ctx, "failed to create state machine",
			slog.Any("error", err),
			slog.String("state", state),
			slog.String("game_state_id", gameStateID.String()))
		return err
	}

	s.stateMachines.Start(ctx, gameStateID, stateMachine)
	s.logger.InfoContext(ctx, "state machine restarted after resume",
		slog.String("game_state_id", gameStateID.String()),
		slog.String("state", state),
		slog.Duration("remaining_time", remaining))

	return nil
}
