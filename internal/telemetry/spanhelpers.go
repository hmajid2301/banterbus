package telemetry

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// BusinessOperationSpan creates a span for business logic operations
func BusinessOperationSpan(ctx context.Context, operation string, attrs ...attribute.KeyValue) (context.Context, trace.Span, func(error)) {
	tracer := otel.Tracer("banterbus-business")

	baseAttrs := []attribute.KeyValue{
		attribute.String("business.operation", operation),
		attribute.String("component", "business_logic"),
	}
	baseAttrs = append(baseAttrs, attrs...)

	ctx, span := tracer.Start(ctx, "business."+operation,
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(baseAttrs...),
	)

	start := time.Now()

	return ctx, span, func(err error) {
		duration := time.Since(start).Seconds()
		success := err == nil

		if err != nil {
			span.RecordError(err)
			span.SetAttributes(attribute.String("error.message", err.Error()))
		}

		span.SetAttributes(
			attribute.Float64("business.duration", duration),
			attribute.Bool("business.success", success),
		)

		RecordBusinessOperation(ctx, operation, success, duration)

		span.End()
	}
}

// GameStateSpan creates a span for game state transitions
func GameStateSpan(ctx context.Context, fromState, toState string, playerCount int) (context.Context, trace.Span, func(error)) {
	tracer := otel.Tracer("banterbus-game-state")

	ctx, span := tracer.Start(ctx, "game.state_transition",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(
			attribute.String("game.from_state", fromState),
			attribute.String("game.to_state", toState),
			attribute.Int("game.player_count", playerCount),
			attribute.String("component", "game_state"),
		),
	)

	return ctx, span, func(err error) {
		success := err == nil

		if err != nil {
			span.RecordError(err)
			span.SetAttributes(attribute.String("error.message", err.Error()))
		}

		span.SetAttributes(
			attribute.Bool("game.transition_success", success),
		)

		RecordGameStateTransition(ctx, fromState, toState, success, playerCount)

		span.End()
	}
}

// LobbyOperationSpan creates a span for lobby operations
func LobbyOperationSpan(ctx context.Context, operation string, attrs ...attribute.KeyValue) (context.Context, trace.Span, func(error)) {
	tracer := otel.Tracer("banterbus-lobby")

	baseAttrs := []attribute.KeyValue{
		attribute.String("lobby.operation", operation),
		attribute.String("component", "lobby"),
	}
	baseAttrs = append(baseAttrs, attrs...)

	ctx, span := tracer.Start(ctx, "lobby."+operation,
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(baseAttrs...),
	)

	start := time.Now()

	return ctx, span, func(err error) {
		duration := time.Since(start).Seconds()
		success := err == nil

		if err != nil {
			span.RecordError(err)
			span.SetAttributes(attribute.String("error.message", err.Error()))
		}

		span.SetAttributes(
			attribute.Float64("lobby.duration", duration),
			attribute.Bool("lobby.success", success),
		)

		RecordBusinessOperation(ctx, "lobby_"+operation, success, duration)

		span.End()
	}
}

// WebSocketMessageSpan creates a span for WebSocket message handling
func WebSocketMessageSpan(ctx context.Context, messageType string) (context.Context, trace.Span, func(error)) {
	tracer := otel.Tracer("banterbus-websocket")

	ctx, span := tracer.Start(ctx, "ws.message."+messageType,
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(
			attribute.String("ws.message_type", messageType),
			attribute.String("component", "websocket"),
		),
	)

	start := time.Now()

	return ctx, span, func(err error) {
		duration := time.Since(start).Seconds()
		success := err == nil

		if err != nil {
			span.RecordError(err)
			span.SetAttributes(attribute.String("error.message", err.Error()))
		}

		span.SetAttributes(
			attribute.Float64("ws.processing_duration", duration),
			attribute.Bool("ws.success", success),
		)

		RecordWebSocketEvent(ctx, messageType, success)

		span.End()
	}
}

// AuthenticationSpan creates a span for authentication operations
func AuthenticationSpan(ctx context.Context, operation string) (context.Context, trace.Span, func(error)) {
	tracer := otel.Tracer("banterbus-auth")

	ctx, span := tracer.Start(ctx, "auth."+operation,
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(
			attribute.String("auth.operation", operation),
			attribute.String("component", "authentication"),
		),
	)

	start := time.Now()

	return ctx, span, func(err error) {
		duration := time.Since(start).Seconds()
		success := err == nil

		if err != nil {
			span.RecordError(err)
			span.SetAttributes(attribute.String("error.message", err.Error()))
		}

		span.SetAttributes(
			attribute.Float64("auth.duration", duration),
			attribute.Bool("auth.success", success),
		)

		RecordBusinessOperation(ctx, "auth_"+operation, success, duration)

		span.End()
	}
}

// CacheOperationSpan creates a span for cache operations
func CacheOperationSpan(ctx context.Context, operation, key string) (context.Context, trace.Span, func(bool, error)) {
	tracer := otel.Tracer("banterbus-cache")

	ctx, span := tracer.Start(ctx, "cache."+operation,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("cache.operation", operation),
			attribute.String("cache.key", key),
			attribute.String("component", "cache"),
		),
	)

	start := time.Now()

	return ctx, span, func(hit bool, err error) {
		duration := time.Since(start).Seconds()
		success := err == nil

		if err != nil {
			span.RecordError(err)
			span.SetAttributes(attribute.String("error.message", err.Error()))
		}

		span.SetAttributes(
			attribute.Float64("cache.duration", duration),
			attribute.Bool("cache.success", success),
			attribute.Bool("cache.hit", hit),
		)

		RecordBusinessOperation(ctx, "cache_"+operation, success, duration)

		span.End()
	}
}
