package metrics

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func IncrementStateCount(ctx context.Context, state string) error {
	m := otel.Meter("gitlab.com/hmajid2301/banterbus")

	counter, err := m.Int64Counter("state.count.total",
		metric.WithDescription("Total number of time this state was started."),
		metric.WithUnit("{call}"),
	)
	if err != nil {
		return err
	}

	counter.Add(ctx, 1, metric.WithAttributes(attribute.String("state", state)))
	return nil
}

func RecordStateDuration(ctx context.Context, duration float64, state string) error {
	m := otel.Meter("gitlab.com/hmajid2301/banterbus")

	histogram, err := m.Float64Histogram("state.duration",
		metric.WithDescription("The time we are in that state."),
		metric.WithUnit("s"),
	)
	if err != nil {
		return err
	}

	histogram.Record(ctx, duration, metric.WithAttributes(attribute.String("state", state)))
	return nil
}

func IncrementStateOperationError(ctx context.Context, state string, errReason string) error {
	m := otel.Meter("gitlab.com/hmajid2301/banterbus")

	counter, err := m.Int64Counter("state.operation.errors",
		metric.WithDescription("Count of errors during state operations"),
		metric.WithUnit("{error}"),
	)
	if err != nil {
		return err
	}

	counter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("state", state),
		attribute.String("error", errReason),
	))
	return nil
}

func RecordStateTransitionFailure(ctx context.Context, fromState, toState, reason string) error {
	m := otel.Meter("gitlab.com/hmajid2301/banterbus")

	counter, err := m.Int64Counter("state.transition.failures",
		metric.WithDescription("Count of failed state transitions"),
		metric.WithUnit("{failure}"),
	)
	if err != nil {
		return err
	}

	counter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("from_state", fromState),
		attribute.String("to_state", toState),
		attribute.String("reason", reason),
	))
	return nil
}

func RecordGameStuckDuration(ctx context.Context, state string, duration float64) error {
	m := otel.Meter("gitlab.com/hmajid2301/banterbus")

	histogram, err := m.Float64Histogram("game.stuck.duration",
		metric.WithDescription("Duration games remain stuck in a state"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries([]float64{60, 300, 600, 1800, 3600}...),
	)
	if err != nil {
		return err
	}

	histogram.Record(ctx, duration, metric.WithAttributes(
		attribute.String("state", state),
	))
	return nil
}

// IncrementGameStuckCount tracks the number of games that get stuck
func IncrementGameStuckCount(ctx context.Context, state string, stuckReason string) error {
	m := otel.Meter("gitlab.com/hmajid2301/banterbus")

	counter, err := m.Int64Counter("game.stuck.count",
		metric.WithDescription("Number of games stuck in specific states"),
		metric.WithUnit("{game}"),
	)
	if err != nil {
		return err
	}

	counter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("state", state),
		attribute.String("reason", stuckReason),
	))
	return nil
}

// RecordStateTimeoutCount tracks state timeouts
func RecordStateTimeoutCount(ctx context.Context, state string, timeoutType string) error {
	m := otel.Meter("gitlab.com/hmajid2301/banterbus")

	counter, err := m.Int64Counter("state.timeout.count",
		metric.WithDescription("Number of state timeouts"),
		metric.WithUnit("{timeout}"),
	)
	if err != nil {
		return err
	}

	counter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("state", state),
		attribute.String("timeout_type", timeoutType),
	))
	return nil
}

// RecordActiveGamesGauge tracks the number of active games
func RecordActiveGamesGauge(ctx context.Context) error {
	m := otel.Meter("gitlab.com/hmajid2301/banterbus")

	gauge, err := m.Int64ObservableGauge("games.active.count",
		metric.WithDescription("Number of active games"),
		metric.WithUnit("{game}"),
	)
	if err != nil {
		return err
	}

	_, err = m.RegisterCallback(
		func(_ context.Context, o metric.Observer) error {
			o.ObserveInt64(gauge, 0)
			return nil
		},
		gauge,
	)
	return err
}
