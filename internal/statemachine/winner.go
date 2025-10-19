package statemachine

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/gofrs/uuid/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"gitlab.com/hmajid2301/banterbus/internal/service"
	"gitlab.com/hmajid2301/banterbus/internal/telemetry"
)

type WinnerState struct {
	GameStateID  uuid.UUID
	Dependencies *StateDependencies
}

func NewWinnerState(gameStateID uuid.UUID, deps *StateDependencies) (*WinnerState, error) {
	if deps == nil {
		return nil, errors.New("dependencies cannot be nil")
	}
	return &WinnerState{
		GameStateID:  gameStateID,
		Dependencies: deps,
	}, nil
}

func (r *WinnerState) Start(ctx context.Context) error {
	stateCtx, cleanup := startStateExecution(
		ctx,
		"winner",
		r.GameStateID,
		r.Dependencies.Logger,
		r.Dependencies.Timings.ShowWinnerScreenFor.Milliseconds(),
	)
	defer cleanup()

	deadline := time.Now().UTC().Add(r.Dependencies.Timings.ShowWinnerScreenFor)

	winnerState, err := r.updateToWinnerWithRetry(stateCtx, deadline)
	if err != nil {
		return err
	}

	if err := r.Dependencies.ClientUpdater.UpdateClientsAboutWinner(stateCtx.ctx, winnerState); err != nil {
		stateCtx.recordClientUpdateError(err)
	}

	timer := time.NewTimer(time.Until(deadline))
	defer timer.Stop()

	select {
	case <-timer.C:
		stateCtx.span.AddEvent("game_completion", trace.WithAttributes(
			attribute.String("completion_status", "success"),
		))

		if err := r.Dependencies.RoundService.FinishGame(stateCtx.ctx, r.GameStateID); err != nil {
			stateCtx.span.SetStatus(codes.Error, "failed to finish game")
			stateCtx.span.RecordError(err, trace.WithAttributes(
				attribute.String("error.type", "game_cleanup_failure"),
			))

			stateCtx.logger.ErrorContext(stateCtx.ctx,
				"failed to finish game",
				slog.Any("error", err),
				slog.String("game_state_id", r.GameStateID.String()))

			_ = telemetry.IncrementStateOperationError(stateCtx.ctx, "winner", "game_cleanup")
			return nil
		}

		stateCtx.span.AddEvent("game_terminated", trace.WithAttributes(
			attribute.String("termination_reason", "normal_completion"),
		))
	case <-stateCtx.ctx.Done():
		stateCtx.logger.InfoContext(stateCtx.ctx, "winner state cancelled",
			slog.String("game_state_id", r.GameStateID.String()))
	}

	return nil
}

func (r *WinnerState) updateToWinnerWithRetry(stateCtx *stateExecutionContext, deadline time.Time) (service.WinnerState, error) {
	winnerState, err := r.Dependencies.RoundService.UpdateStateToWinner(stateCtx.ctx, r.GameStateID, deadline)
	if err != nil {
		if errors.Is(err, service.ErrNotInScoringState) {
			stateCtx.logger.WarnContext(stateCtx.ctx,
				"state transition race condition detected, game already transitioned",
				slog.Any("error", err),
				slog.String("game_state_id", r.GameStateID.String()))
		}
		stateCtx.recordStateUpdateError(err, "update_to_winner")
		return service.WinnerState{}, err
	}

	return winnerState, nil
}
