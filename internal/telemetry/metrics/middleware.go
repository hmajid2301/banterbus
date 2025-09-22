package metrics

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type GameStateMiddleware struct {
	meter               metric.Meter
	stateTransitionHist metric.Float64Histogram
	stateErrorCounter   metric.Int64Counter
}

func NewGameStateMiddleware() (*GameStateMiddleware, error) {
	meter := otel.Meter("gitlab.com/hmajid2301/banterbus")

	stateTransitionHist, err := meter.Float64Histogram(
		"game.state.transition.duration",
		metric.WithDescription("Time taken for game state transitions"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries([]float64{0.1, 0.5, 1.0, 2.0, 5.0, 10.0, 30.0}...),
	)
	if err != nil {
		return nil, err
	}

	stateErrorCounter, err := meter.Int64Counter(
		"game.state.errors.total",
		metric.WithDescription("Total number of game state operation errors"),
		metric.WithUnit("{error}"),
	)
	if err != nil {
		return nil, err
	}

	return &GameStateMiddleware{
		meter:               meter,
		stateTransitionHist: stateTransitionHist,
		stateErrorCounter:   stateErrorCounter,
	}, nil
}

func (m *GameStateMiddleware) RecordStateTransition(
	ctx context.Context,
	fromState, toState string,
	duration time.Duration,
	gameID string,
	playerCount int,
	success bool,
) {
	status := "success"
	if !success {
		status = "failure"
		m.stateErrorCounter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("from_state", fromState),
			attribute.String("to_state", toState),
			attribute.String("reason", "transition_failed"),
		))
	}

	m.stateTransitionHist.Record(ctx, duration.Seconds(), metric.WithAttributes(
		attribute.String("from_state", fromState),
		attribute.String("to_state", toState),
		attribute.String("status", status),
		attribute.Int("player_count", playerCount),
	))

	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.AddEvent("game.state.transition", trace.WithAttributes(
			attribute.String("from_state", fromState),
			attribute.String("to_state", toState),
			attribute.String("game_id", gameID),
			attribute.String("duration", duration.String()),
			attribute.Bool("success", success),
			attribute.Int("player_count", playerCount),
		))
	}
}

func RecordHTTPRequest(ctx context.Context, endpoint, method string, statusCode int, duration float64) error {
	m := otel.Meter("gitlab.com/hmajid2301/banterbus")

	counter, err := m.Int64Counter("http.requests.total",
		metric.WithDescription("Total HTTP requests by endpoint"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		return err
	}

	status := "success"
	if statusCode >= 400 {
		status = "error"
	}

	counter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("endpoint", endpoint),
		attribute.String("method", method),
		attribute.String("status", status),
		attribute.Int("status_code", statusCode),
	))

	histogram, err := m.Float64Histogram("http.request.duration",
		metric.WithDescription("HTTP request duration by endpoint"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries([]float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0}...),
	)
	if err != nil {
		return err
	}

	histogram.Record(ctx, duration, metric.WithAttributes(
		attribute.String("endpoint", endpoint),
		attribute.String("method", method),
		attribute.String("status", status),
	))

	return nil
}

func InitializeMetrics(ctx context.Context) error {
	if err := RegisterMetrics(ctx); err != nil {
		return err
	}

	if err := ActiveConnections(ctx); err != nil {
		return err
	}

	return nil
}
