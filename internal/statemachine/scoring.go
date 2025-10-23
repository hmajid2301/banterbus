package statemachine

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/gofrs/uuid/v5"
	"go.opentelemetry.io/otel/attribute"

	"gitlab.com/hmajid2301/banterbus/internal/service"
)

type ScoringState struct {
	GameStateID  uuid.UUID
	Dependencies *StateDependencies
}

func NewScoringState(gameStateID uuid.UUID, deps *StateDependencies) (*ScoringState, error) {
	if deps == nil {
		return nil, errors.New("dependencies cannot be nil")
	}
	return &ScoringState{
		GameStateID:  gameStateID,
		Dependencies: deps,
	}, nil
}

func (r *ScoringState) Start(ctx context.Context) error {
	stateCtx, cleanup := startStateExecution(
		ctx,
		"scoring",
		r.GameStateID,
		r.Dependencies.Logger,
		r.Dependencies.Timings.ShowScoreScreenFor.Milliseconds(),
		attribute.Int("scoring.guess_fibber_points", r.Dependencies.Scoring.GuessedFibber),
		attribute.Int("scoring.fibber_evade_points", r.Dependencies.Scoring.FibberEvadeCapture),
	)
	defer cleanup()

	deadline := time.Now().UTC().Add(r.Dependencies.Timings.ShowScoreScreenFor)

	scoringState, err := r.updateToScoringWithRetry(stateCtx, deadline)
	if err != nil {
		return err
	}

	if err := r.Dependencies.ClientUpdater.UpdateClientsAboutScore(stateCtx.ctx, scoringState); err != nil {
		stateCtx.recordClientUpdateError(err)
	}

	timer := time.NewTimer(time.Until(deadline))
	defer timer.Stop()

	select {
	case <-timer.C:
		r.transitionToNextState(stateCtx, scoringState)
	case <-stateCtx.ctx.Done():
		stateCtx.logger.InfoContext(stateCtx.ctx, "scoring state cancelled",
			slog.String("game_state_id", r.GameStateID.String()))
	}

	return nil
}

func (r *ScoringState) transitionToNextState(stateCtx *stateExecutionContext, scoringState service.ScoreState) {
	shouldEndGame := scoringState.TotalRounds >= 3 ||
		scoringState.RoundType == service.RoundTypeMostLikely

	stateCtx.logger.InfoContext(stateCtx.ctx, "scoring state transition decision",
		slog.Int("round_number", scoringState.RoundNumber),
		slog.Int("total_rounds", scoringState.TotalRounds),
		slog.Bool("fibber_caught", scoringState.FibberCaught),
		slog.String("round_type", scoringState.RoundType),
		slog.Bool("should_end_game", shouldEndGame),
		slog.String("game_state_id", r.GameStateID.String()))

	if shouldEndGame {
		stateCtx.addTransition("winner", "timeout", attribute.Bool("game_ended", true))
		w, err := NewWinnerState(r.GameStateID, r.Dependencies)
		if err != nil {
			stateCtx.logger.ErrorContext(stateCtx.ctx, "failed to create winner state",
				slog.Any("error", err),
				slog.String("game_state_id", r.GameStateID.String()))
			return
		}
		r.Dependencies.Transitioner.StartStateMachine(stateCtx.ctx, r.GameStateID, w)
	} else {
		stateCtx.addTransition("question", "timeout", attribute.Bool("next_round", true))
		q, err := NewQuestionState(r.GameStateID, true, r.Dependencies)
		if err != nil {
			stateCtx.logger.ErrorContext(stateCtx.ctx, "failed to create question state",
				slog.Any("error", err),
				slog.String("game_state_id", r.GameStateID.String()))
			return
		}
		r.Dependencies.Transitioner.StartStateMachine(stateCtx.ctx, r.GameStateID, q)
	}
}

func (r *ScoringState) updateToScoringWithRetry(stateCtx *stateExecutionContext, deadline time.Time) (service.ScoreState, error) {
	scoringState, err := r.Dependencies.RoundService.UpdateStateToScore(stateCtx.ctx, r.GameStateID, deadline, r.Dependencies.Scoring)
	if err != nil {
		if errors.Is(err, service.ErrNotInRevealState) {
			stateCtx.logger.WarnContext(stateCtx.ctx,
				"state transition race condition detected, game already transitioned",
				slog.Any("error", err),
				slog.String("game_state_id", r.GameStateID.String()))
		}
		stateCtx.recordStateUpdateError(err, "update_to_scoring")
		return service.ScoreState{}, err
	}

	return scoringState, nil
}
