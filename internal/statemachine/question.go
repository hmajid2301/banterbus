package statemachine

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/gofrs/uuid/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"gitlab.com/hmajid2301/banterbus/internal/service"
	"gitlab.com/hmajid2301/banterbus/internal/store/db"
	"gitlab.com/hmajid2301/banterbus/internal/telemetry"
)

type QuestionState struct {
	GameStateID  uuid.UUID
	NextRound    bool
	Dependencies *StateDependencies
}

func NewQuestionState(gameStateID uuid.UUID, nextRound bool, deps *StateDependencies) (*QuestionState, error) {
	if deps == nil {
		return nil, errors.New("dependencies cannot be nil")
	}
	return &QuestionState{
		GameStateID:  gameStateID,
		NextRound:    nextRound,
		Dependencies: deps,
	}, nil
}

var tracer = otel.Tracer("banterbus-game-state")

func (q *QuestionState) Start(ctx context.Context) error {
	stateCtx, cleanup := startStateExecution(
		ctx,
		"question",
		q.GameStateID,
		q.Dependencies.Logger,
		q.Dependencies.Timings.ShowQuestionScreenFor.Milliseconds(),
		attribute.Bool("question_state.next_round", q.NextRound),
	)
	defer cleanup()

	telemetry.AddGameStateTransition(stateCtx.ctx, "", "FibbingITQuestion", "timer_start", &q.GameStateID)
	telemetry.AddTimingAttributes(stateCtx.ctx, "question_timer",
		q.Dependencies.Timings.ShowQuestionScreenFor.String(),
		q.Dependencies.Timings.ShowQuestionScreenFor.String(), false)

	deadline := time.Now().UTC().Add(q.Dependencies.Timings.ShowQuestionScreenFor)

	questionState, err := q.Dependencies.RoundService.UpdateStateToQuestion(stateCtx.ctx, q.GameStateID, deadline, q.NextRound)
	if err != nil {
		q.handleQuestionUpdateError(stateCtx, err)
		return nil
	}

	showModal := q.NextRound
	if err := q.Dependencies.ClientUpdater.UpdateClientsAboutQuestion(stateCtx.ctx, questionState, showModal); err != nil {
		stateCtx.recordClientUpdateError(err)
	}

	timer := time.NewTimer(time.Until(deadline))
	defer timer.Stop()

	select {
	case <-timer.C:
		q.transitionToVoting(stateCtx)
	case <-stateCtx.ctx.Done():
		stateCtx.logger.InfoContext(stateCtx.ctx, "question state cancelled",
			slog.String("game_state_id", q.GameStateID.String()))
	}

	return nil
}

func (q *QuestionState) handleQuestionUpdateError(stateCtx *stateExecutionContext, err error) {
	if errors.Is(err, service.ErrGameCompleted) {
		stateCtx.addTransition("winner", "all_round_types_completed")
		stateCtx.logger.InfoContext(stateCtx.ctx, "all round types completed, transitioning to winner state",
			slog.String("game_state_id", q.GameStateID.String()))

		w, err := NewWinnerState(q.GameStateID, q.Dependencies)
		if err != nil {
			stateCtx.logger.ErrorContext(stateCtx.ctx, "failed to create winner state",
				slog.Any("error", err),
				slog.String("game_state_id", q.GameStateID.String()))
			return
		}
		q.Dependencies.Transitioner.StartStateMachine(stateCtx.ctx, q.GameStateID, w)
		return
	}

	stateCtx.span.SetStatus(codes.Error, "failed to update state")
	stateCtx.span.RecordError(err, trace.WithAttributes(
		attribute.String("error.type", "state_update_failure"),
	))

	if errors.Is(err, service.ErrNoNormalQuestions) || errors.Is(err, service.ErrNoFibberQuestions) {
		stateCtx.logger.ErrorContext(stateCtx.ctx,
			"question availability issue during testing, game likely in cleanup",
			slog.Any("error", err),
			slog.String("game_state_id", q.GameStateID.String()))
	} else if errors.Is(err, service.ErrAlreadyInQuestionState) {
		stateCtx.logger.ErrorContext(stateCtx.ctx,
			"race condition detected: already in question state, ignoring duplicate transition",
			slog.Any("error", err),
			slog.String("game_state_id", q.GameStateID.String()))
	} else if errors.Is(err, sql.ErrNoRows) {
		select {
		case <-stateCtx.ctx.Done():
			stateCtx.logger.ErrorContext(stateCtx.ctx,
				"game state deleted during context cancellation, stopping state machine",
				slog.Any("error", err),
				slog.String("game_state_id", q.GameStateID.String()))
		default:
			stateCtx.logger.DebugContext(stateCtx.ctx,
				"temporary database issue detected, stopping question state",
				slog.Any("error", err),
				slog.String("game_state_id", q.GameStateID.String()))
		}
		return
	} else {
		stateCtx.logger.ErrorContext(stateCtx.ctx,
			"failed to update game state to question",
			slog.Any("error", err),
			slog.String("game_state_id", q.GameStateID.String()))
	}

	_ = telemetry.IncrementStateOperationError(stateCtx.ctx, "question", "state_update")
}

func (q *QuestionState) transitionToVoting(stateCtx *stateExecutionContext) {
	stateCtx.addTransition("voting", "timeout_or_all_ready")

	currentGameState, err := q.Dependencies.RoundService.GetGameState(stateCtx.ctx, q.GameStateID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			select {
			case <-stateCtx.ctx.Done():
				stateCtx.logger.ErrorContext(stateCtx.ctx,
					"game state deleted during context cancellation, stopping state machine",
					slog.Any("error", err),
					slog.String("game_state_id", q.GameStateID.String()))
			default:
				stateCtx.logger.DebugContext(stateCtx.ctx,
					"temporary database issue during voting transition, retrying with voting state",
					slog.Any("error", err),
					slog.String("game_state_id", q.GameStateID.String()))
				time.Sleep(1 * time.Second)
				v, err := NewVotingState(q.GameStateID, q.Dependencies)
				if err != nil {
					stateCtx.logger.ErrorContext(stateCtx.ctx, "failed to create voting state",
						slog.Any("error", err),
						slog.String("game_state_id", q.GameStateID.String()))
					return
				}
				q.Dependencies.Transitioner.StartStateMachine(stateCtx.ctx, q.GameStateID, v)
			}
			return
		}

		stateCtx.logger.ErrorContext(stateCtx.ctx,
			"failed to get game state before voting transition",
			slog.Any("error", err),
			slog.String("game_state_id", q.GameStateID.String()))
		return
	}

	if currentGameState == db.FibbingITQuestion {
		v, err := NewVotingState(q.GameStateID, q.Dependencies)
		if err != nil {
			stateCtx.logger.ErrorContext(stateCtx.ctx, "failed to create voting state",
				slog.Any("error", err),
				slog.String("game_state_id", q.GameStateID.String()))
			return
		}
		q.Dependencies.Transitioner.StartStateMachine(stateCtx.ctx, q.GameStateID, v)
	} else {
		stateCtx.logger.InfoContext(stateCtx.ctx, "game state already transitioned from question state",
			slog.String("current_state", currentGameState.String()))
	}
}
