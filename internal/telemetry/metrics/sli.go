package metrics

import (
	"context"
	"sync/atomic"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	successfulGameCompletions int64
	totalGameAttempts         int64
	activeLobbies             int64
	gamesInProgress           int64
)

func RecordGameCompletion(ctx context.Context, success bool, duration float64, playerCount int) error {
	m := otel.Meter("gitlab.com/hmajid2301/banterbus")

	// Game completion tracking
	completionCounter, err := m.Int64Counter("game.completion.total",
		metric.WithDescription("Total game completion attempts"),
		metric.WithUnit("{game}"),
	)
	if err != nil {
		return err
	}

	status := "failure"
	if success {
		status = "success"
		atomic.AddInt64(&successfulGameCompletions, 1)
	}
	atomic.AddInt64(&totalGameAttempts, 1)

	completionCounter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("status", status),
		attribute.Int("player_count", playerCount),
	))

	// Game duration tracking
	durationHistogram, err := m.Float64Histogram("game.duration",
		metric.WithDescription("Game duration in seconds"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries([]float64{60, 300, 600, 900, 1800, 3600}...),
	)
	if err != nil {
		return err
	}

	durationHistogram.Record(ctx, duration, metric.WithAttributes(
		attribute.String("status", status),
		attribute.Int("player_count", playerCount),
	))

	return nil
}

// RecordPlayerExperience tracks player satisfaction
func RecordPlayerExperience(ctx context.Context, satisfaction string) error {
	m := otel.Meter("gitlab.com/hmajid2301/banterbus")

	counter, err := m.Int64Counter("player.experience",
		metric.WithDescription("Player experience satisfaction ratings"),
		metric.WithUnit("{response}"),
	)
	if err != nil {
		return err
	}

	counter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("satisfaction", satisfaction),
	))

	return nil
}

// RecordAvailability tracks service availability
func RecordAvailability(ctx context.Context, endpoint string, statusCode int, duration float64) error {
	m := otel.Meter("gitlab.com/hmajid2301/banterbus")

	// Request tracking
	availabilityCounter, err := m.Int64Counter("http.requests.total",
		metric.WithDescription("Total HTTP requests"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		return err
	}

	status := "success"
	if statusCode >= 500 {
		status = "failure"
	}

	availabilityCounter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("endpoint", endpoint),
		attribute.String("status", status),
		attribute.Int("status_code", statusCode),
	))

	// Response time tracking
	responseTimeHistogram, err := m.Float64Histogram("http.response_time",
		metric.WithDescription("HTTP response time"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries([]float64{0.1, 0.2, 0.5, 1.0, 2.0, 5.0}...),
	)
	if err != nil {
		return err
	}

	responseTimeHistogram.Record(ctx, duration, metric.WithAttributes(
		attribute.String("endpoint", endpoint),
		attribute.String("status", status),
	))

	return nil
}

// RecordLobbyHealth tracks lobby operations
func RecordLobbyHealth(ctx context.Context, operation string, success bool) error {
	m := otel.Meter("gitlab.com/hmajid2301/banterbus")

	counter, err := m.Int64Counter("lobby.operations.total",
		metric.WithDescription("Lobby operations"),
		metric.WithUnit("{operation}"),
	)
	if err != nil {
		return err
	}

	status := "failure"
	if success {
		status = "success"
	}

	counter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("operation", operation),
		attribute.String("status", status),
	))

	return nil
}

// UpdateActiveGameCounts updates gauges for active games and lobbies
func UpdateActiveGameCounts(ctx context.Context, lobbies, games int64) error {
	atomic.StoreInt64(&activeLobbies, lobbies)
	atomic.StoreInt64(&gamesInProgress, games)
	return nil
}

// RegisterMetrics registers gauge metrics for monitoring
func RegisterMetrics(ctx context.Context) error {
	m := otel.Meter("gitlab.com/hmajid2301/banterbus")

	// Game completion rate gauge
	gameCompletionRate, err := m.Float64ObservableGauge("game.completion_rate",
		metric.WithDescription("Current game completion rate"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return err
	}

	// Active lobbies gauge
	activeLobbyGauge, err := m.Int64ObservableGauge("lobby.active",
		metric.WithDescription("Number of active lobbies"),
		metric.WithUnit("{lobby}"),
	)
	if err != nil {
		return err
	}

	// Active games gauge
	activeGameGauge, err := m.Int64ObservableGauge("game.active",
		metric.WithDescription("Number of active games"),
		metric.WithUnit("{game}"),
	)
	if err != nil {
		return err
	}

	_, err = m.RegisterCallback(
		func(_ context.Context, o metric.Observer) error {
			total := atomic.LoadInt64(&totalGameAttempts)
			successful := atomic.LoadInt64(&successfulGameCompletions)
			var rate float64
			if total > 0 {
				rate = float64(successful) / float64(total)
			}

			o.ObserveFloat64(gameCompletionRate, rate)
			o.ObserveInt64(activeLobbyGauge, atomic.LoadInt64(&activeLobbies))
			o.ObserveInt64(activeGameGauge, atomic.LoadInt64(&gamesInProgress))
			return nil
		},
		gameCompletionRate,
		activeLobbyGauge,
		activeGameGauge,
	)

	return err
}

// RecordServiceCall tracks service operation calls
func RecordServiceCall(ctx context.Context, operation string, success bool, duration float64) error {
	m := otel.Meter("gitlab.com/hmajid2301/banterbus")

	counter, err := m.Int64Counter("service.calls.total",
		metric.WithDescription("Total service calls"),
		metric.WithUnit("{call}"),
	)
	if err != nil {
		return err
	}

	status := "failure"
	if success {
		status = "success"
	}

	counter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("operation", operation),
		attribute.String("status", status),
	))

	histogram, err := m.Float64Histogram("service.call.duration",
		metric.WithDescription("Service call duration"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries([]float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0, 5.0}...),
	)
	if err != nil {
		return err
	}

	histogram.Record(ctx, duration, metric.WithAttributes(
		attribute.String("operation", operation),
		attribute.String("status", status),
	))

	return nil
}
