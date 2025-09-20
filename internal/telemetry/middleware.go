package telemetry

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
	stuckGamesCounter   metric.Int64Counter
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

	stuckGamesCounter, err := meter.Int64Counter(
		"game.stuck.total",
		metric.WithDescription("Total number of stuck games detected"),
		metric.WithUnit("{game}"),
	)
	if err != nil {
		return nil, err
	}

	return &GameStateMiddleware{
		meter:               meter,
		stateTransitionHist: stateTransitionHist,
		stateErrorCounter:   stateErrorCounter,
		stuckGamesCounter:   stuckGamesCounter,
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

func (m *GameStateMiddleware) DetectStuckGame(
	ctx context.Context,
	currentState string,
	stateDuration time.Duration,
	playerCount int,
	expectedMaxDuration time.Duration,
) bool {
	if stateDuration > expectedMaxDuration {
		m.stuckGamesCounter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("stuck_state", currentState),
			attribute.String("duration", stateDuration.String()),
			attribute.Int("player_count", playerCount),
			attribute.String("reason", "timeout_exceeded"),
		))

		if err := RecordGameStuckDuration(ctx, currentState, stateDuration.Seconds()); err == nil {
			IncrementGameStuckCount(ctx, currentState, "timeout_exceeded")
		}

		return true
	}
	return false
}

type GameHealthChecker struct {
	meter                 metric.Meter
	healthCheckHist       metric.Float64Histogram
	unhealthyGamesCounter metric.Int64Counter
}

func NewGameHealthChecker() (*GameHealthChecker, error) {
	meter := otel.Meter("gitlab.com/hmajid2301/banterbus")

	healthCheckHist, err := meter.Float64Histogram(
		"game.health.check.duration",
		metric.WithDescription("Time taken for game health checks"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries([]float64{0.01, 0.05, 0.1, 0.5, 1.0}...),
	)
	if err != nil {
		return nil, err
	}

	unhealthyGamesCounter, err := meter.Int64Counter(
		"game.unhealthy.total",
		metric.WithDescription("Total number of unhealthy games detected"),
		metric.WithUnit("{game}"),
	)
	if err != nil {
		return nil, err
	}

	return &GameHealthChecker{
		meter:                 meter,
		healthCheckHist:       healthCheckHist,
		unhealthyGamesCounter: unhealthyGamesCounter,
	}, nil
}

// PerformHealthCheck runs comprehensive health checks on a game
func (hc *GameHealthChecker) PerformHealthCheck(
	ctx context.Context,
	checks map[string]func() (bool, string),
) map[string]bool {
	start := time.Now()
	results := make(map[string]bool)
	var hasFailures bool

	for checkName, checkFunc := range checks {
		passed, reason := checkFunc()
		results[checkName] = passed

		if !passed {
			hasFailures = true
			hc.unhealthyGamesCounter.Add(ctx, 1, metric.WithAttributes(
				attribute.String("check_type", checkName),
				attribute.String("failure_reason", reason),
			))

			RecordDataIntegrity(ctx, checkName, false, reason)
		} else {
			RecordDataIntegrity(ctx, checkName, true, "check_passed")
		}
	}

	status := "healthy"
	if hasFailures {
		status = "unhealthy"
	}

	hc.healthCheckHist.Record(ctx, time.Since(start).Seconds(), metric.WithAttributes(
		attribute.String("status", status),
		attribute.Int("checks_performed", len(checks)),
	))

	return results
}

type ConnectionQualityTracker struct {
	meter                 metric.Meter
	connectionLatencyHist metric.Float64Histogram
	connectionErrorRate   metric.Float64Histogram
	activeConnections     metric.Int64ObservableGauge
}

func NewConnectionQualityTracker() (*ConnectionQualityTracker, error) {
	meter := otel.Meter("gitlab.com/hmajid2301/banterbus")

	connectionLatencyHist, err := meter.Float64Histogram(
		"connection.quality.latency",
		metric.WithDescription("WebSocket connection latency"),
		metric.WithUnit("ms"),
		metric.WithExplicitBucketBoundaries([]float64{1, 5, 10, 25, 50, 100, 250, 500, 1000}...),
	)
	if err != nil {
		return nil, err
	}

	connectionErrorRate, err := meter.Float64Histogram(
		"connection.quality.error_rate",
		metric.WithDescription("WebSocket connection error rate"),
		metric.WithUnit("%"),
		metric.WithExplicitBucketBoundaries([]float64{0.1, 0.5, 1.0, 2.0, 5.0, 10.0}...),
	)
	if err != nil {
		return nil, err
	}

	activeConnections, err := meter.Int64ObservableGauge(
		"connection.quality.active_count",
		metric.WithDescription("Number of active high-quality connections"),
		metric.WithUnit("{connection}"),
	)
	if err != nil {
		return nil, err
	}

	return &ConnectionQualityTracker{
		meter:                 meter,
		connectionLatencyHist: connectionLatencyHist,
		connectionErrorRate:   connectionErrorRate,
		activeConnections:     activeConnections,
	}, nil
}

func (cqt *ConnectionQualityTracker) RecordConnectionQuality(
	ctx context.Context,
	playerID string,
	latency time.Duration,
	errorRate float64,
	messagesSent, messagesReceived int,
) {
	cqt.connectionLatencyHist.Record(ctx, float64(latency.Milliseconds()))

	cqt.connectionErrorRate.Record(ctx, errorRate)

	quality := "good"
	if latency > 500*time.Millisecond || errorRate > 5.0 {
		quality = "poor"
	} else if latency > 100*time.Millisecond || errorRate > 1.0 {
		quality = "fair"
	}

	RecordConnectionQuality(ctx, float64(latency.Milliseconds()), errorRate)

	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(
			attribute.String("connection.quality", quality),
			attribute.String("player_id", playerID),
			attribute.Int64("connection.latency_ms", latency.Milliseconds()),
			attribute.Float64("connection.error_rate", errorRate),
			attribute.Int("connection.messages_sent", messagesSent),
			attribute.Int("connection.messages_received", messagesReceived),
		)
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
	if err := RegisterSLOGauges(ctx); err != nil {
		return err
	}

	if err := ActiveConnections(ctx); err != nil {
		return err
	}

	return nil
}
