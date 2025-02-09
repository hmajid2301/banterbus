package websockets

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
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

var tracer = otel.Tracer("")

func (q *QuestionState) Start(ctx context.Context) {
	const spanName = "fibbing_it.question_state.process"
	start := time.Now()

	ctx, span := tracer.Start(
		ctx,
		spanName,
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(
			attribute.String("game.id", q.GameStateID.String()),
			attribute.Bool("question_state.next_round", q.NextRound),
			attribute.Int64("question_state.configured_duration_ms",
				q.Subscriber.config.Timings.ShowQuestionScreenFor.Milliseconds()),
		),
	)
	defer span.End()

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
		span.SetStatus(codes.Error, "failed to update state")
		span.RecordError(err, trace.WithAttributes(
			attribute.String("error.type", "state_update_failure"),
		))

		q.Subscriber.logger.ErrorContext(
			ctx,
			"failed to update game state to question",
			slog.Any("error", err),
			slog.String("game_state_id", q.GameStateID.String()),
		)

		playerIDs := []uuid.UUID{}
		for _, player := range questionState.Players {
			playerIDs = append(playerIDs, player.ID)
		}
		err = q.Subscriber.updateClientsAboutErr(ctx, playerIDs, "Failed to move to question page.")
		if err != nil {
			q.Subscriber.logger.ErrorContext(
				ctx,
				"failed to update clients",
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

	time.Sleep(time.Until(deadline))
	span.AddEvent("state_transition", trace.WithAttributes(
		attribute.String("next_state", "voting"),
		attribute.String("transition_reason", "timeout"),
	))

	v := &VotingState{GameStateID: q.GameStateID, Subscriber: q.Subscriber}
	go v.Start(ctx)
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

		v.Subscriber.logger.ErrorContext(
			ctx,
			"failed to update game state to voting",
			slog.Any("error", err),
			slog.String("game_state_id", v.GameStateID.String()),
		)

		playerIDs := []uuid.UUID{}
		for _, player := range votingState.Players {
			playerIDs = append(playerIDs, player.ID)
		}

		err = v.Subscriber.updateClientsAboutErr(ctx, playerIDs, "Failed to move to voting page.")
		if err != nil {
			v.Subscriber.logger.ErrorContext(
				ctx,
				"failed to update clients",
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

	time.Sleep(time.Until(deadline))
	span.AddEvent("state_transition", trace.WithAttributes(
		attribute.String("next_state", "reveal"),
		attribute.String("transition_reason", "timeout"),
	))

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

		r.Subscriber.logger.ErrorContext(
			ctx,
			"failed to update game state to reveal",
			slog.Any("error", err),
			slog.String("game_state_id", r.GameStateID.String()),
		)

		err = r.Subscriber.updateClientsAboutErr(ctx, revealState.PlayerIDs, "Failed to move to reveal page.")
		if err != nil {
			r.Subscriber.logger.ErrorContext(
				ctx,
				"failed to update clients",
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
		// TODO: turn this string into say ROUND_3
		if revealState.RoundType == "most_likely" {
			nextState = db.FibbingItWinner
		}
	}

	span.AddEvent("state_transition", trace.WithAttributes(
		attribute.String("next_state", nextState.String()),
		attribute.Bool("final_round", finalRound),
		attribute.Bool("fibber_found", fibberFound),
	))

	time.Sleep(time.Until(deadline))

	switch nextState {
	case db.FibbingItWinner:
		w := &WinnerState{GameStateID: r.GameStateID, Subscriber: r.Subscriber}
		go w.Start(ctx)
	case db.FibbingItScoring:
		s := &ScoringState{GameStateID: r.GameStateID, Subscriber: r.Subscriber}
		go s.Start(ctx)
	default:
		q := &QuestionState{GameStateID: r.GameStateID, Subscriber: r.Subscriber, NextRound: false}
		go q.Start(ctx)
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

		r.Subscriber.logger.ErrorContext(
			ctx,
			"failed to update game state to scoring",
			slog.Any("error", err),
			slog.String("game_state_id", r.GameStateID.String()),
		)

		playerIDs := []uuid.UUID{}
		for _, player := range scoringState.Players {
			playerIDs = append(playerIDs, player.ID)
		}

		err = r.Subscriber.updateClientsAboutErr(ctx, playerIDs, "Failed to move to scoring page.")
		if err != nil {
			r.Subscriber.logger.ErrorContext(
				ctx,
				"failed to update clients",
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

	time.Sleep(time.Until(deadline))
	span.AddEvent("state_transition", trace.WithAttributes(
		attribute.String("next_state", "question"),
		attribute.String("transition_reason", "timeout"),
		attribute.Bool("next_round", true),
	))

	q := &QuestionState{GameStateID: r.GameStateID, Subscriber: r.Subscriber, NextRound: true}
	go q.Start(ctx)
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

		r.Subscriber.logger.ErrorContext(
			ctx,
			"failed to update game state to winner",
			slog.Any("error", err),
			slog.String("game_state_id", r.GameStateID.String()),
		)

		playerIDs := []uuid.UUID{}
		for _, player := range winnerState.Players {
			playerIDs = append(playerIDs, player.ID)
		}

		err = r.Subscriber.updateClientsAboutErr(ctx, playerIDs, "Failed to move to winner reveal page.")
		if err != nil {
			r.Subscriber.logger.ErrorContext(
				ctx,
				"failed to update clients",
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

	time.Sleep(time.Until(deadline))
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
}
