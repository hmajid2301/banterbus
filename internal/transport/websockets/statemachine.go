package websockets

import (
	"context"
	"log/slog"
	"strings"
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

type State interface {
	Start(ctx context.Context)
}

type QuestionState struct {
	GameStateID uuid.UUID
	Subscriber  Subscriber
	NextRound   bool
}

var tracer = otel.Tracer("banterbus-game-state")

func (q *QuestionState) Start(ctx context.Context) {
	const spanName = "game.question_state.process"
	start := time.Now()

	ctx, span := telemetry.StartInternalSpan(ctx, tracer, spanName,
		attribute.String("game.state_id", q.GameStateID.String()),
		attribute.String("game.state", "question"),
		attribute.Bool("question_state.next_round", q.NextRound),
		attribute.Int64("question_state.configured_duration_ms",
			q.Subscriber.config.Timings.ShowQuestionScreenFor.Milliseconds()),
		attribute.String("component", "game-state-machine"),
	)
	defer span.End()

	telemetry.AddGameStateTransition(ctx, "", "FibbingITQuestion", "timer_start", &q.GameStateID)
	telemetry.AddTimingAttributes(ctx, "question_timer",
		q.Subscriber.config.Timings.ShowQuestionScreenFor.String(),
		q.Subscriber.config.Timings.ShowQuestionScreenFor.String(), false)

	err := telemetry.IncrementStateCount(ctx, "question")
	if err != nil {
		q.Subscriber.logger.WarnContext(ctx, "failed to increment state counter",
			slog.Any("error", err))
	}

	defer func() {
		duration := float64(time.Since(start).Seconds())
		if err := telemetry.RecordStateDuration(ctx, duration, "question"); err != nil {
			q.Subscriber.logger.WarnContext(ctx, "failed to record state duration",
				slog.Any("error", err))
		}
	}()

	deadline := time.Now().UTC().Add(q.Subscriber.config.Timings.ShowQuestionScreenFor)

	questionState, err := q.Subscriber.roundService.UpdateStateToQuestion(ctx, q.GameStateID, deadline, q.NextRound)
	if err != nil {
		if err.Error() == "game completed - no more round types available" {
			span.AddEvent("state_transition", trace.WithAttributes(
				attribute.String("next_state", "winner"),
				attribute.String("transition_reason", "all_round_types_completed"),
			))

			q.Subscriber.logger.InfoContext(ctx, "all round types completed, transitioning to winner state",
				slog.String("game_state_id", q.GameStateID.String()))

			w := &WinnerState{GameStateID: q.GameStateID, Subscriber: q.Subscriber}
			go w.Start(telemetry.PropagateContext(ctx))
			return
		}

		span.SetStatus(codes.Error, "failed to update state")
		span.RecordError(err, trace.WithAttributes(
			attribute.String("error.type", "state_update_failure"),
		))

		if err.Error() == "no fibber questions available" || err.Error() == "no normal questions available" {
			q.Subscriber.logger.ErrorContext(
				ctx,
				"question availability issue during testing, game likely in cleanup",
				slog.Any("error", err),
				slog.String("game_state_id", q.GameStateID.String()),
			)
		} else if strings.Contains(err.Error(), "current state: FibbingITQuestion") {
			q.Subscriber.logger.ErrorContext(
				ctx,
				"race condition detected: already in question state, ignoring duplicate transition",
				slog.Any("error", err),
				slog.String("game_state_id", q.GameStateID.String()),
			)
		} else if err.Error() == "no rows in result set" {
			// Check if context is cancelled, which indicates intentional cleanup
			select {
			case <-ctx.Done():
				q.Subscriber.logger.ErrorContext(
					ctx,
					"game state deleted during context cancellation, stopping state machine",
					slog.Any("error", err),
					slog.String("game_state_id", q.GameStateID.String()),
				)
				return
			default:
				// Context not cancelled, this might be a transient database issue
				q.Subscriber.logger.WarnContext(
					ctx,
					"temporary database issue detected, retrying after delay",
					slog.Any("error", err),
					slog.String("game_state_id", q.GameStateID.String()),
				)
				// Add a small delay before continuing
				time.Sleep(100 * time.Millisecond)
			}
		} else {
			q.Subscriber.logger.ErrorContext(
				ctx,
				"failed to update game state to question",
				slog.Any("error", err),
				slog.String("game_state_id", q.GameStateID.String()),
			)
		}

		_ = telemetry.IncrementStateOperationError(ctx, "question", "state_update")
		return
	}

	showModal := q.NextRound
	if err := q.Subscriber.updateClientsAboutQuestion(ctx, questionState, showModal); err != nil {
		span.SetStatus(codes.Error, "client update failed")
		span.RecordError(err, trace.WithAttributes(
			attribute.String("error.type", "client_update_failure"),
		))

		q.Subscriber.logger.ErrorContext(
			ctx,
			"failed to update clients to question screen",
			slog.Any("error", err),
			slog.String("game_state_id", q.GameStateID.String()),
		)

		_ = telemetry.IncrementStateOperationError(ctx, "question", "client_update")
		return
	}

	// Use a timer that can be cancelled with context
	timer := time.NewTimer(time.Until(deadline))
	defer timer.Stop()

	select {
	case <-timer.C:
		currentGameState, err := q.Subscriber.roundService.GetGameState(ctx, q.GameStateID)
		if err != nil {
			if err.Error() == "no rows in result set" {
				// Check if context is cancelled, which indicates intentional cleanup
				select {
				case <-ctx.Done():
					q.Subscriber.logger.ErrorContext(
						ctx,
						"game state deleted during context cancellation, stopping state machine",
						slog.Any("error", err),
						slog.String("game_state_id", q.GameStateID.String()),
					)
					return
				default:
					// Context not cancelled, this might be a transient database issue
					q.Subscriber.logger.WarnContext(
						ctx,
						"temporary database issue during voting transition, retrying with voting state",
						slog.Any("error", err),
						slog.String("game_state_id", q.GameStateID.String()),
					)
					// Proceed to start voting state anyway - it will handle the retry logic
					time.Sleep(1 * time.Second)
					v := &VotingState{GameStateID: q.GameStateID, Subscriber: q.Subscriber}
					go v.Start(telemetry.PropagateContext(ctx))
					return
				}
			} else {
				q.Subscriber.logger.ErrorContext(
					ctx,
					"failed to get game state before voting transition",
					slog.Any("error", err),
					slog.String("game_state_id", q.GameStateID.String()),
				)
			}
			return
		}

		if currentGameState == db.FibbingITQuestion {
			span.AddEvent("state_transition", trace.WithAttributes(
				attribute.String("next_state", "voting"),
				attribute.String("transition_reason", "timeout"),
			))

			v := &VotingState{GameStateID: q.GameStateID, Subscriber: q.Subscriber}
			go v.Start(ctx)
		} else {
			q.Subscriber.logger.InfoContext(ctx, "game state already transitioned from question state", slog.String("current_state", currentGameState.String()))
		}
	case <-ctx.Done():
		q.Subscriber.logger.InfoContext(ctx, "question state cancelled before timeout",
			slog.String("game_state_id", q.GameStateID.String()))
		return
	}
}

type VotingState struct {
	Subscriber  Subscriber
	GameStateID uuid.UUID
}

func (v *VotingState) Start(ctx context.Context) {
	const spanName = "fibbing_it.voting_state.process"
	start := time.Now()

	ctx, span := tracer.Start(
		ctx,
		spanName,
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(
			attribute.String("game.id", v.GameStateID.String()),
			attribute.Int64("voting_state.configured_duration_ms",
				v.Subscriber.config.Timings.ShowVotingScreenFor.Milliseconds()),
		),
	)
	defer span.End()

	if err := telemetry.IncrementStateCount(ctx, "voting"); err != nil {
		v.Subscriber.logger.WarnContext(ctx, "failed to increment state counter",
			slog.Any("error", err))
	}

	defer func() {
		duration := float64(time.Since(start).Seconds())
		if err := telemetry.RecordStateDuration(ctx, duration, "voting"); err != nil {
			v.Subscriber.logger.WarnContext(ctx, "failed to record state duration",
				slog.Any("error", err))
		}
	}()

	deadline := time.Now().UTC().Add(v.Subscriber.config.Timings.ShowVotingScreenFor)

	votingState, err := v.Subscriber.roundService.UpdateStateToVoting(ctx, v.GameStateID, deadline)
	if err != nil {
		span.SetStatus(codes.Error, "failed to update state")
		span.RecordError(err, trace.WithAttributes(
			attribute.String("error.type", "state_update_failure"),
		))

		if err.Error() == "no rows in result set" {
			// Check if context is cancelled, which indicates intentional cleanup
			select {
			case <-ctx.Done():
				v.Subscriber.logger.ErrorContext(
					ctx,
					"game state deleted during context cancellation, stopping voting state",
					slog.Any("error", err),
					slog.String("game_state_id", v.GameStateID.String()),
				)
				return
			default:
				// Context not cancelled, this might be a transient database issue
				v.Subscriber.logger.WarnContext(
					ctx,
					"temporary database issue during voting state start, retrying after delay",
					slog.Any("error", err),
					slog.String("game_state_id", v.GameStateID.String()),
				)
				time.Sleep(1 * time.Second)
				// Retry by starting a new voting state
				v2 := &VotingState{GameStateID: v.GameStateID, Subscriber: v.Subscriber}
				go v2.Start(ctx)
				return
			}
		} else if err.Error() == "game state is not in FIBBING_IT_QUESTION state" {
			v.Subscriber.logger.WarnContext(
				ctx,
				"state transition race condition detected, game already transitioned",
				slog.Any("error", err),
				slog.String("game_state_id", v.GameStateID.String()),
			)
		} else {
			v.Subscriber.logger.ErrorContext(
				ctx,
				"failed to update game state to voting",
				slog.Any("error", err),
				slog.String("game_state_id", v.GameStateID.String()),
			)
		}

		_ = telemetry.IncrementStateOperationError(ctx, "voting", "state_update")
		return
	}

	if err := v.Subscriber.updateClientsAboutVoting(ctx, votingState); err != nil {
		span.SetStatus(codes.Error, "client update failed")
		span.RecordError(err, trace.WithAttributes(
			attribute.String("error.type", "client_update_failure"),
		))

		v.Subscriber.logger.ErrorContext(
			ctx,
			"failed to update clients to voting screen",
			slog.Any("error", err),
			slog.String("game_state_id", v.GameStateID.String()),
		)

		_ = telemetry.IncrementStateOperationError(ctx, "voting", "client_update")
		return
	}

	// Check periodically if all players have voted to transition early
	ticker := time.NewTicker(1 * time.Second) // Check every second
	defer ticker.Stop()

	// Use a proper timer that respects context cancellation
	timer := time.NewTimer(time.Until(deadline))
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			// Timeout reached, transition to reveal
			span.AddEvent("state_transition", trace.WithAttributes(
				attribute.String("next_state", "reveal"),
				attribute.String("transition_reason", "timeout"),
			))
			goto transitionToReveal
		case <-ticker.C:
			// Check if all players have voted
			allReady, err := v.Subscriber.roundService.AreAllPlayersVotingReady(ctx, v.GameStateID)
			if err != nil {
				v.Subscriber.logger.WarnContext(ctx, "failed to check if all players are ready for voting",
					slog.Any("error", err),
					slog.String("game_state_id", v.GameStateID.String()))
				continue // Continue checking on error
			}

			if allReady {
				// All players have voted, transition early
				span.AddEvent("state_transition", trace.WithAttributes(
					attribute.String("next_state", "reveal"),
					attribute.String("transition_reason", "all_players_voted"),
				))
				v.Subscriber.logger.InfoContext(ctx, "all players have voted, transitioning early to reveal",
					slog.String("game_state_id", v.GameStateID.String()))
				goto transitionToReveal
			}
		case <-ctx.Done():
			v.Subscriber.logger.InfoContext(ctx, "voting state cancelled before completion",
				slog.String("game_state_id", v.GameStateID.String()))
			return
		}
	}

transitionToReveal:

	r := &RevealState{GameStateID: v.GameStateID, Subscriber: v.Subscriber}
	go r.Start(ctx)
}

type RevealState struct {
	Subscriber  Subscriber
	GameStateID uuid.UUID
}

func (r *RevealState) Start(ctx context.Context) {
	const spanName = "fibbing_it.reveal_state.process"
	start := time.Now()

	ctx, span := tracer.Start(
		ctx,
		spanName,
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(
			attribute.String("game.id", r.GameStateID.String()),
			attribute.Int64("reveal_state.configured_duration_ms",
				r.Subscriber.config.Timings.ShowRevealScreenFor.Milliseconds()),
		),
	)
	defer span.End()

	if err := telemetry.IncrementStateCount(ctx, "reveal"); err != nil {
		r.Subscriber.logger.WarnContext(ctx, "failed to increment state counter",
			slog.Any("error", err))
	}

	defer func() {
		duration := float64(time.Since(start).Seconds())
		if err := telemetry.RecordStateDuration(ctx, duration, "reveal"); err != nil {
			r.Subscriber.logger.WarnContext(ctx, "failed to record state duration",
				slog.Any("error", err))
		}
	}()

	deadline := time.Now().UTC().Add(r.Subscriber.config.Timings.ShowRevealScreenFor)

	revealState, err := r.Subscriber.roundService.UpdateStateToReveal(ctx, r.GameStateID, deadline)
	if err != nil {
		span.SetStatus(codes.Error, "failed to update state")
		span.RecordError(err, trace.WithAttributes(
			attribute.String("error.type", "state_update_failure"),
		))

		if err.Error() == "no rows in result set" {
			// Check if context is cancelled, which indicates intentional cleanup
			select {
			case <-ctx.Done():
				r.Subscriber.logger.ErrorContext(
					ctx,
					"game state deleted during context cancellation, stopping reveal state",
					slog.Any("error", err),
					slog.String("game_state_id", r.GameStateID.String()),
				)
				return
			default:
				// Context not cancelled, this might be a transient database issue
				r.Subscriber.logger.WarnContext(
					ctx,
					"temporary database issue during reveal state start, retrying after delay",
					slog.Any("error", err),
					slog.String("game_state_id", r.GameStateID.String()),
				)
				time.Sleep(1 * time.Second)
				// Retry by starting a new reveal state
				r2 := &RevealState{GameStateID: r.GameStateID, Subscriber: r.Subscriber}
				go r2.Start(ctx)
				return
			}
		} else if err.Error() == "game state is not in FIBBING_IT_VOTING state" {
			r.Subscriber.logger.WarnContext(
				ctx,
				"state transition race condition detected, game already transitioned",
				slog.Any("error", err),
				slog.String("game_state_id", r.GameStateID.String()),
			)
		} else {
			r.Subscriber.logger.ErrorContext(
				ctx,
				"failed to update game state to reveal",
				slog.Any("error", err),
				slog.String("game_state_id", r.GameStateID.String()),
			)
		}

		_ = telemetry.IncrementStateOperationError(ctx, "reveal", "state_update")
		return
	}

	span.SetAttributes(
		attribute.Int("round.number", revealState.Round),
		attribute.String("round.type", revealState.RoundType),
	)

	if err := r.Subscriber.updateClientsAboutReveal(ctx, revealState); err != nil {
		span.SetStatus(codes.Error, "client update failed")
		span.RecordError(err, trace.WithAttributes(
			attribute.String("error.type", "client_update_failure"),
		))

		r.Subscriber.logger.ErrorContext(
			ctx,
			"failed to update clients to reveal screen",
			slog.Any("error", err),
			slog.String("game_state_id", r.GameStateID.String()),
		)

		_ = telemetry.IncrementStateOperationError(ctx, "reveal", "client_update")
		return
	}

	maxRounds := 3
	finalRound := revealState.Round == maxRounds
	fibberFound := revealState.ShouldReveal && revealState.VotedForPlayerRole == "fibber"
	nextState := db.FibbingITQuestion

	if finalRound || fibberFound {
		nextState = db.FibbingItScoring
		if revealState.RoundType == service.RoundTypeMostLikely {
			nextState = db.FibbingItWinner
		}
	}

	span.AddEvent("state_transition", trace.WithAttributes(
		attribute.String("next_state", nextState.String()),
		attribute.Bool("final_round", finalRound),
		attribute.Bool("fibber_found", fibberFound),
	))

	// Use a timer that respects context cancellation
	timer := time.NewTimer(time.Until(deadline))
	defer timer.Stop()

	select {
	case <-timer.C:
		switch nextState {
		case db.FibbingItWinner:
			w := &WinnerState{GameStateID: r.GameStateID, Subscriber: r.Subscriber}
			go w.Start(telemetry.PropagateContext(ctx))
		case db.FibbingItScoring:
			s := &ScoringState{GameStateID: r.GameStateID, Subscriber: r.Subscriber}
			go s.Start(ctx)
		default:
			q := &QuestionState{GameStateID: r.GameStateID, Subscriber: r.Subscriber, NextRound: false}
			go q.Start(ctx)
		}
	case <-ctx.Done():
		r.Subscriber.logger.InfoContext(ctx, "reveal state cancelled before completion",
			slog.String("game_state_id", r.GameStateID.String()))
		return
	}
}

type ScoringState struct {
	Subscriber  Subscriber
	GameStateID uuid.UUID
}

func (r *ScoringState) Start(ctx context.Context) {
	const spanName = "fibbing_it.scoring_state.process"
	start := time.Now()

	ctx, span := tracer.Start(
		ctx,
		spanName,
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(
			attribute.String("game.id", r.GameStateID.String()),
			attribute.Int64("scoring_state.configured_duration_ms",
				r.Subscriber.config.Timings.ShowScoreScreenFor.Milliseconds()),
			attribute.Int("scoring.guess_fibber_points", r.Subscriber.config.Scoring.GuessFibber),
			attribute.Int("scoring.fibber_evade_points", r.Subscriber.config.Scoring.FibberEvadeCapture),
		),
	)
	defer span.End()

	if err := telemetry.IncrementStateCount(ctx, "scoring"); err != nil {
		r.Subscriber.logger.WarnContext(ctx, "failed to increment state counter",
			slog.Any("error", err))
	}

	defer func() {
		duration := float64(time.Since(start).Seconds())
		if err := telemetry.RecordStateDuration(ctx, duration, "scoring"); err != nil {
			r.Subscriber.logger.WarnContext(ctx, "failed to record state duration",
				slog.Any("error", err))
		}
	}()

	deadline := time.Now().UTC().Add(r.Subscriber.config.Timings.ShowScoreScreenFor)
	scoring := service.Scoring{
		GuessedFibber:      r.Subscriber.config.Scoring.GuessFibber,
		FibberEvadeCapture: r.Subscriber.config.Scoring.FibberEvadeCapture,
	}

	scoringState, err := r.Subscriber.roundService.UpdateStateToScore(ctx, r.GameStateID, deadline, scoring)
	if err != nil {
		span.SetStatus(codes.Error, "failed to update state")
		span.RecordError(err, trace.WithAttributes(
			attribute.String("error.type", "state_update_failure"),
		))

		if err.Error() == "no rows in result set" {
			// Check if context is cancelled, which indicates intentional cleanup
			select {
			case <-ctx.Done():
				r.Subscriber.logger.ErrorContext(
					ctx,
					"game state deleted during context cancellation, stopping scoring state",
					slog.Any("error", err),
					slog.String("game_state_id", r.GameStateID.String()),
				)
				return
			default:
				// Context not cancelled, this might be a transient database issue
				r.Subscriber.logger.WarnContext(
					ctx,
					"temporary database issue during scoring state start, retrying after delay",
					slog.Any("error", err),
					slog.String("game_state_id", r.GameStateID.String()),
				)
				time.Sleep(1 * time.Second)
				// Retry by starting a new scoring state
				r2 := &ScoringState{GameStateID: r.GameStateID, Subscriber: r.Subscriber}
				go r2.Start(ctx)
				return
			}
		} else if err.Error() == "game state is not in FIBBING_IT_REVEAL_ROLE state" {
			r.Subscriber.logger.WarnContext(
				ctx,
				"state transition race condition detected, game already transitioned",
				slog.Any("error", err),
				slog.String("game_state_id", r.GameStateID.String()),
			)
		} else {
			r.Subscriber.logger.ErrorContext(
				ctx,
				"failed to update game state to scoring",
				slog.Any("error", err),
				slog.String("game_state_id", r.GameStateID.String()),
			)
		}

		_ = telemetry.IncrementStateOperationError(ctx, "scoring", "state_update")
		return
	}

	if err := r.Subscriber.updateClientsAboutScore(ctx, scoringState); err != nil {
		span.SetStatus(codes.Error, "client update failed")
		span.RecordError(err, trace.WithAttributes(
			attribute.String("errorl_type", "client_update_failure"),
		))

		r.Subscriber.logger.ErrorContext(
			ctx,
			"failed to update clients to scoring screen",
			slog.Any("error", err),
			slog.String("game_state_id", r.GameStateID.String()),
		)

		_ = telemetry.IncrementStateOperationError(ctx, "scoring", "client_update")
		return
	}

	// Use a timer that respects context cancellation
	timer := time.NewTimer(time.Until(deadline))
	defer timer.Stop()

	select {
	case <-timer.C:
		span.AddEvent("state_transition", trace.WithAttributes(
			attribute.String("next_state", "question"),
			attribute.String("transition_reason", "timeout"),
			attribute.Bool("next_round", true),
		))

		q := &QuestionState{GameStateID: r.GameStateID, Subscriber: r.Subscriber, NextRound: true}
		go q.Start(ctx)
	case <-ctx.Done():
		r.Subscriber.logger.InfoContext(ctx, "scoring state cancelled before completion",
			slog.String("game_state_id", r.GameStateID.String()))
		return
	}
}

type NewRoundState struct {
	Subscriber  Subscriber
	GameStateID uuid.UUID
}

type WinnerState struct {
	Subscriber  Subscriber
	GameStateID uuid.UUID
}

func (r *WinnerState) Start(ctx context.Context) {
	const spanName = "fibbing_it.winner_state.process"
	start := time.Now()

	ctx, span := tracer.Start(
		ctx,
		spanName,
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(
			attribute.String("game.id", r.GameStateID.String()),
			attribute.Int64("winner_state.configured_duration_ms",
				r.Subscriber.config.Timings.ShowWinnerScreenFor.Milliseconds()),
		),
	)
	defer span.End()

	if err := telemetry.IncrementStateCount(ctx, "winner"); err != nil {
		r.Subscriber.logger.WarnContext(ctx, "failed to increment state counter",
			slog.Any("error", err))
	}

	defer func() {
		duration := float64(time.Since(start).Seconds())
		if err := telemetry.RecordStateDuration(ctx, duration, "winner"); err != nil {
			r.Subscriber.logger.WarnContext(ctx, "failed to record state duration",
				slog.Any("error", err))
		}
	}()

	deadline := time.Now().UTC().Add(r.Subscriber.config.Timings.ShowWinnerScreenFor)

	winnerState, err := r.Subscriber.roundService.UpdateStateToWinner(ctx, r.GameStateID, deadline)
	if err != nil {
		span.SetStatus(codes.Error, "failed to update state")
		span.RecordError(err, trace.WithAttributes(
			attribute.String("error.type", "state_update_failure"),
		))

		if err.Error() == "no rows in result set" {
			// Check if context is cancelled, which indicates intentional cleanup
			select {
			case <-ctx.Done():
				r.Subscriber.logger.ErrorContext(
					ctx,
					"game state deleted during context cancellation, stopping winner state",
					slog.Any("error", err),
					slog.String("game_state_id", r.GameStateID.String()),
				)
				return
			default:
				// Context not cancelled, this might be a transient database issue
				r.Subscriber.logger.WarnContext(
					ctx,
					"temporary database issue during winner state start, retrying after delay",
					slog.Any("error", err),
					slog.String("game_state_id", r.GameStateID.String()),
				)
				time.Sleep(1 * time.Second)
				// Retry by starting a new winner state
				r2 := &WinnerState{GameStateID: r.GameStateID, Subscriber: r.Subscriber}
				go r2.Start(ctx)
				return
			}
		} else if err.Error() == "game state is not in FIBBING_IT_SCORING_STATE state" {
			r.Subscriber.logger.WarnContext(
				ctx,
				"state transition race condition detected, game already transitioned",
				slog.Any("error", err),
				slog.String("game_state_id", r.GameStateID.String()),
			)
		} else {
			r.Subscriber.logger.ErrorContext(
				ctx,
				"failed to update game state to winner",
				slog.Any("error", err),
				slog.String("game_state_id", r.GameStateID.String()),
			)
		}

		_ = telemetry.IncrementStateOperationError(ctx, "winner", "state_update")
		return
	}

	if err := r.Subscriber.updateClientsAboutWinner(ctx, winnerState); err != nil {
		span.SetStatus(codes.Error, "client update failed")
		span.RecordError(err, trace.WithAttributes(
			attribute.String("error.type", "client_update_failure"),
		))

		r.Subscriber.logger.ErrorContext(
			ctx,
			"failed to update clients to winner screen",
			slog.Any("error", err),
			slog.String("game_state_id", r.GameStateID.String()),
		)

		_ = telemetry.IncrementStateOperationError(ctx, "winner", "client_update")
		return
	}

	// Use a timer that respects context cancellation
	timer := time.NewTimer(time.Until(deadline))
	defer timer.Stop()

	select {
	case <-timer.C:
		span.AddEvent("game_completion", trace.WithAttributes(
			attribute.String("completion_status", "success"),
		))

		if err := r.Subscriber.roundService.FinishGame(ctx, r.GameStateID); err != nil {
			span.SetStatus(codes.Error, "failed to finish game")
			span.RecordError(err, trace.WithAttributes(
				attribute.String("error.type", "game_cleanup_failure"),
			))

			r.Subscriber.logger.ErrorContext(
				ctx,
				"failed to finish game",
				slog.Any("error", err),
				slog.String("game_state_id", r.GameStateID.String()),
			)

			_ = telemetry.IncrementStateOperationError(ctx, "winner", "game_cleanup")
			return
		}

		span.AddEvent("game_terminated", trace.WithAttributes(
			attribute.String("termination_reason", "normal_completion"),
		))
	case <-ctx.Done():
		r.Subscriber.logger.InfoContext(ctx, "winner state cancelled before completion",
			slog.String("game_state_id", r.GameStateID.String()))
		return
	}
}
