package statemachine

import (
	"context"
	"log/slog"
	"time"

	"github.com/gofrs/uuid/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"gitlab.com/hmajid2301/banterbus/internal/telemetry"
)

const (
	DefaultTickerInterval = 2 * time.Second
)

type stateExecutionContext struct {
	ctx         context.Context
	span        trace.Span
	stateName   string
	gameStateID uuid.UUID
	logger      *slog.Logger
	startTime   time.Time
}

func startStateExecution(
	ctx context.Context,
	stateName string,
	gameStateID uuid.UUID,
	logger *slog.Logger,
	durationMs int64,
	extraAttrs ...attribute.KeyValue,
) (*stateExecutionContext, func()) {
	logger.DebugContext(ctx, stateName+" state starting",
		slog.String("game_state_id", gameStateID.String()))

	attrs := []attribute.KeyValue{
		attribute.String("game.state_id", gameStateID.String()),
		attribute.String("game.state", stateName),
		attribute.Int64(stateName+"_state.configured_duration_ms", durationMs),
	}
	attrs = append(attrs, extraAttrs...)

	ctx, span := tracer.Start(ctx, "game."+stateName+"_state.process",
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(attrs...))

	if err := telemetry.IncrementStateCount(ctx, stateName); err != nil {
		logger.WarnContext(ctx, "failed to increment state counter", slog.Any("error", err))
	}

	stateCtx := &stateExecutionContext{
		ctx:         ctx,
		span:        span,
		stateName:   stateName,
		gameStateID: gameStateID,
		logger:      logger,
		startTime:   time.Now(),
	}

	cleanup := func() {
		span.End()
		duration := float64(time.Since(stateCtx.startTime).Seconds())
		if err := telemetry.RecordStateDuration(ctx, duration, stateName); err != nil {
			logger.WarnContext(ctx, "failed to record state duration", slog.Any("error", err))
		}
	}

	return stateCtx, cleanup
}

func (e *stateExecutionContext) recordClientUpdateError(err error) {
	e.span.SetStatus(codes.Error, "client update failed")
	e.span.RecordError(err, trace.WithAttributes(
		attribute.String("error.type", "client_update_failure"),
	))

	e.logger.ErrorContext(e.ctx,
		"failed to update clients to "+e.stateName+" screen",
		slog.Any("error", err),
		slog.String("game_state_id", e.gameStateID.String()))

	_ = telemetry.IncrementStateOperationError(e.ctx, e.stateName, "client_update")
}

func (e *stateExecutionContext) addTransition(nextState, reason string, extraAttrs ...attribute.KeyValue) {
	attrs := []attribute.KeyValue{
		attribute.String("next_state", nextState),
		attribute.String("transition_reason", reason),
	}
	attrs = append(attrs, extraAttrs...)
	e.span.AddEvent("state_transition", trace.WithAttributes(attrs...))
}

func (e *stateExecutionContext) recordStateUpdateError(err error, operation string) {
	e.span.SetStatus(codes.Error, "failed to update state")
	e.span.RecordError(err, trace.WithAttributes(
		attribute.String("error.type", "state_update_failure"),
	))

	e.logger.ErrorContext(e.ctx,
		"failed to update state",
		slog.Any("error", err),
		slog.String("operation", operation),
		slog.String("game_state_id", e.gameStateID.String()))

	_ = telemetry.IncrementStateOperationError(e.ctx, e.stateName, "state_update")
}
