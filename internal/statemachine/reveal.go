package statemachine

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/gofrs/uuid/v5"
	"go.opentelemetry.io/otel/attribute"

	"gitlab.com/hmajid2301/banterbus/internal/service"
	"gitlab.com/hmajid2301/banterbus/internal/store/db"
)

type RevealState struct {
	GameStateID  uuid.UUID
	Dependencies *StateDependencies
}

func NewRevealState(gameStateID uuid.UUID, deps *StateDependencies) (*RevealState, error) {
	if deps == nil {
		return nil, errors.New("dependencies cannot be nil")
	}
	return &RevealState{
		GameStateID:  gameStateID,
		Dependencies: deps,
	}, nil
}

func (r *RevealState) Start(ctx context.Context) error {
	stateCtx, cleanup := startStateExecution(
		ctx,
		"reveal",
		r.GameStateID,
		r.Dependencies.Logger,
		r.Dependencies.Timings.ShowRevealScreenFor.Milliseconds(),
	)
	defer cleanup()

	deadline := time.Now().UTC().Add(r.Dependencies.Timings.ShowRevealScreenFor)

	revealState, err := r.updateToRevealWithRetry(stateCtx, deadline)
	if err != nil {
		return err
	}

	stateCtx.span.SetAttributes(
		attribute.Int("round.number", revealState.Round),
		attribute.String("round.type", revealState.RoundType),
	)

	if err := r.Dependencies.ClientUpdater.UpdateClientsAboutReveal(stateCtx.ctx, revealState); err != nil {
		stateCtx.recordClientUpdateError(err)
	}

	nextState := r.determineNextState(stateCtx, revealState)

	timer := time.NewTimer(time.Until(deadline))
	defer timer.Stop()

	select {
	case <-timer.C:
		r.transitionToNextState(stateCtx, nextState)
	case <-stateCtx.ctx.Done():
		stateCtx.logger.InfoContext(stateCtx.ctx, "reveal state cancelled",
			slog.String("game_state_id", r.GameStateID.String()))
	}

	return nil
}

func (r *RevealState) updateToRevealWithRetry(stateCtx *stateExecutionContext, deadline time.Time) (service.RevealRoleState, error) {
	revealState, err := r.Dependencies.RoundService.UpdateStateToReveal(stateCtx.ctx, r.GameStateID, deadline)
	if err != nil {
		if errors.Is(err, service.ErrNotInVotingState) {
			stateCtx.logger.WarnContext(stateCtx.ctx,
				"state transition race condition detected, game already transitioned",
				slog.Any("error", err),
				slog.String("game_state_id", r.GameStateID.String()))
		}
		stateCtx.recordStateUpdateError(err, "update_to_reveal")
		return service.RevealRoleState{}, err
	}

	return revealState, nil
}

func (r *RevealState) determineNextState(stateCtx *stateExecutionContext, revealState service.RevealRoleState) db.FibbingItGameState {
	maxRounds := 3
	finalRound := revealState.Round == maxRounds
	fibberFound := revealState.ShouldReveal && revealState.VotedForPlayerRole == service.FibberRole
	nextState := db.FibbingITQuestion

	if finalRound || fibberFound {
		nextState = db.FibbingItScoring
		if revealState.RoundType == service.RoundTypeMostLikely {
			nextState = db.FibbingItWinner
		}
	}

	stateCtx.logger.InfoContext(stateCtx.ctx, "reveal state transition decision",
		slog.String("next_state", nextState.String()),
		slog.Bool("final_round", finalRound),
		slog.Bool("fibber_found", fibberFound),
		slog.String("game_state_id", r.GameStateID.String()))

	stateCtx.addTransition(nextState.String(), "timeout",
		attribute.Bool("final_round", finalRound),
		attribute.Bool("fibber_found", fibberFound))

	return nextState
}

func (r *RevealState) transitionToNextState(stateCtx *stateExecutionContext, nextState db.FibbingItGameState) {
	stateCtx.logger.InfoContext(stateCtx.ctx, "reveal state timer expired, transitioning",
		slog.String("next_state", nextState.String()),
		slog.String("game_state_id", r.GameStateID.String()))

	currentGameState, err := r.Dependencies.RoundService.GetGameStateByID(stateCtx.ctx, r.GameStateID)
	if err != nil {
		stateCtx.logger.ErrorContext(stateCtx.ctx,
			"failed to verify game state before transition",
			slog.Any("error", err),
			slog.String("game_state_id", r.GameStateID.String()))
		return
	}

	if currentGameState != db.FibbingItReveal {
		stateCtx.logger.InfoContext(stateCtx.ctx, "game state already transitioned from reveal state",
			slog.String("current_state", currentGameState.String()),
			slog.String("expected_next_state", nextState.String()))
		return
	}

	switch nextState {
	case db.FibbingItWinner:
		w, err := NewWinnerState(r.GameStateID, r.Dependencies)
		if err != nil {
			stateCtx.logger.ErrorContext(stateCtx.ctx, "failed to create winner state",
				slog.Any("error", err),
				slog.String("game_state_id", r.GameStateID.String()))
			return
		}
		r.Dependencies.Transitioner.StartStateMachine(stateCtx.ctx, r.GameStateID, w)
	case db.FibbingItScoring:
		s, err := NewScoringState(r.GameStateID, r.Dependencies)
		if err != nil {
			stateCtx.logger.ErrorContext(stateCtx.ctx, "failed to create scoring state",
				slog.Any("error", err),
				slog.String("game_state_id", r.GameStateID.String()))
			return
		}
		r.Dependencies.Transitioner.StartStateMachine(stateCtx.ctx, r.GameStateID, s)
	default:
		q, err := NewQuestionState(r.GameStateID, false, r.Dependencies)
		if err != nil {
			stateCtx.logger.ErrorContext(stateCtx.ctx, "failed to create question state",
				slog.Any("error", err),
				slog.String("game_state_id", r.GameStateID.String()))
			return
		}
		r.Dependencies.Transitioner.StartStateMachine(stateCtx.ctx, r.GameStateID, q)
	}
}
