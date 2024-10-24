package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func IncrementMessageReceived(ctx context.Context, messageType string) error {
	m := otel.Meter("gitlab.com/hmajid2301/banterbus")

	counter, err := m.Int64Counter("message_type_count",
		metric.WithDescription("Total number of messages received."),
		metric.WithUnit("{call}"),
	)
	if err != nil {
		return err
	}

	counter.Add(ctx, 1, metric.WithAttributes(attribute.String("message_type", messageType)))
	return nil
}

func IncrementMessageReceivedError(ctx context.Context) error {
	m := otel.Meter("gitlab.com/hmajid2301/banterbus")

	counter, err := m.Int64Counter("message_type_error_count",
		metric.WithDescription("Total number of messages received which throw an error."),
		metric.WithUnit("{call}"),
	)
	if err != nil {
		return err
	}

	counter.Add(ctx, 1)
	return nil
}
