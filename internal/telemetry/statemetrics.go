package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func IncrementStateCount(ctx context.Context, state string) error {
	m := otel.Meter("gitlab.com/hmajid2301/banterbus")

	counter, err := m.Int64Counter("total_state_count",
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
