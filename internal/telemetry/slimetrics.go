package telemetry

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

	// SLI: Game completion rate
	completionCounter, err := m.Int64Counter("sli.game.completion.total",
		metric.WithDescription("Total game completion attempts for SLI calculation"),
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

	// Game duration for performance SLI
	durationHistogram, err := m.Float64Histogram("sli.game.duration",
		metric.WithDescription("Game duration for performance SLI"),
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

// RecordPlayerExperience tracks player experience metrics for SLI
func RecordPlayerExperience(ctx context.Context, satisfaction string) error {
	m := otel.Meter("gitlab.com/hmajid2301/banterbus")

	counter, err := m.Int64Counter("sli.player.experience",
		metric.WithDescription("Player experience satisfaction for SLI"),
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

// RecordAvailability tracks service availability for SLO
func RecordAvailability(ctx context.Context, endpoint string, statusCode int, duration float64) error {
	m := otel.Meter("gitlab.com/hmajid2301/banterbus")

	// Availability SLI
	availabilityCounter, err := m.Int64Counter("sli.availability.requests.total",
		metric.WithDescription("Total requests for availability SLI calculation"),
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

	// Response time SLI
	responseTimeHistogram, err := m.Float64Histogram("sli.response_time",
		metric.WithDescription("Response time for latency SLI"),
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

// RecordLobbyHealth tracks lobby creation and join success rates
func RecordLobbyHealth(ctx context.Context, operation string, success bool) error {
	m := otel.Meter("gitlab.com/hmajid2301/banterbus")

	counter, err := m.Int64Counter("sli.lobby.operations.total",
		metric.WithDescription("Lobby operations for health SLI"),
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

// RegisterSLOGauges registers gauge metrics for SLO monitoring
func RegisterSLOGauges(ctx context.Context) error {
	m := otel.Meter("gitlab.com/hmajid2301/banterbus")

	// Game completion rate gauge for SLO monitoring
	gameCompletionRate, err := m.Float64ObservableGauge("slo.game.completion_rate",
		metric.WithDescription("Current game completion rate for SLO monitoring"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return err
	}

	// Active lobbies gauge
	activeLobbyGauge, err := m.Int64ObservableGauge("slo.lobbies.active",
		metric.WithDescription("Number of active lobbies"),
		metric.WithUnit("{lobby}"),
	)
	if err != nil {
		return err
	}

	// Active games gauge
	activeGameGauge, err := m.Int64ObservableGauge("slo.games.active",
		metric.WithDescription("Number of games in progress"),
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

func RecordDataIntegrity(ctx context.Context, checkType string, passed bool, details string) error {
	m := otel.Meter("gitlab.com/hmajid2301/banterbus")

	counter, err := m.Int64Counter("sli.data.integrity.checks",
		metric.WithDescription("Data integrity check results"),
		metric.WithUnit("{check}"),
	)
	if err != nil {
		return err
	}

	status := "pass"
	if !passed {
		status = "fail"
	}

	counter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("check_type", checkType),
		attribute.String("status", status),
		attribute.String("details", details),
	))

	return nil
}
