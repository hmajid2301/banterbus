package statemachine

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/gofrs/uuid/v5"

	"gitlab.com/hmajid2301/banterbus/internal/service"
)

type VotingState struct {
	GameStateID  uuid.UUID
	Dependencies *StateDependencies
}

func NewVotingState(gameStateID uuid.UUID, deps *StateDependencies) (*VotingState, error) {
	if deps == nil {
		return nil, errors.New("dependencies cannot be nil")
	}
	return &VotingState{
		GameStateID:  gameStateID,
		Dependencies: deps,
	}, nil
}

func (v *VotingState) Start(ctx context.Context) error {
	stateCtx, cleanup := startStateExecution(
		ctx,
		"voting",
		v.GameStateID,
		v.Dependencies.Logger,
		v.Dependencies.Timings.ShowVotingScreenFor.Milliseconds(),
	)
	defer cleanup()

	deadline := time.Now().UTC().Add(v.Dependencies.Timings.ShowVotingScreenFor)

	votingState, err := v.updateToVotingWithRetry(stateCtx, deadline)
	if err != nil {
		return nil
	}

	if err := v.Dependencies.ClientUpdater.UpdateClientsAboutVoting(stateCtx.ctx, votingState); err != nil {
		stateCtx.recordClientUpdateError(err)
	}

	timer := time.NewTimer(time.Until(deadline))
	defer timer.Stop()

	select {
	case <-timer.C:
		stateCtx.addTransition("reveal", "timeout_or_all_ready")
		r, err := NewRevealState(v.GameStateID, v.Dependencies)
		if err != nil {
			stateCtx.logger.ErrorContext(stateCtx.ctx, "failed to create reveal state",
				slog.Any("error", err),
				slog.String("game_state_id", v.GameStateID.String()))
			return nil
		}
		v.Dependencies.Transitioner.StartStateMachine(stateCtx.ctx, v.GameStateID, r)
	case <-stateCtx.ctx.Done():
		stateCtx.logger.InfoContext(stateCtx.ctx, "voting state cancelled",
			slog.String("game_state_id", v.GameStateID.String()))
	}

	return nil
}

func (v *VotingState) updateToVotingWithRetry(stateCtx *stateExecutionContext, deadline time.Time) (service.VotingState, error) {
	votingState, err := v.Dependencies.RoundService.UpdateStateToVoting(stateCtx.ctx, v.GameStateID, deadline)
	if err != nil {
		if errors.Is(err, service.ErrNotInQuestionState) {
			stateCtx.logger.WarnContext(stateCtx.ctx,
				"state transition race condition detected, game already transitioned",
				slog.Any("error", err),
				slog.String("game_state_id", v.GameStateID.String()))
		}
		stateCtx.recordStateUpdateError(err, "update_to_voting")
		return service.VotingState{}, err
	}

	return votingState, nil
}
