package telemetry

import (
	"context"
	"sync/atomic"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var activeConnectionsCount int64

func IncrementSubscribers(ctx context.Context) error {
	m := otel.Meter("gitlab.com/hmajid2301/banterbus")

	counter, err := m.Int64Counter("websocket.connections.total",
		metric.WithDescription("Total number of WebSocket connections established"),
		metric.WithUnit("{call}"),
	)
	if err != nil {
		return err
	}

	counter.Add(ctx, 1)
	return nil
}

func ActiveConnections(_ context.Context) error {
	m := otel.Meter("gitlab.com/hmajid2301/banterbus")

	gauge, err := m.Int64ObservableGauge("websocket.connections.active",
		metric.WithDescription("Total number of active websocket connections."),
		metric.WithUnit("1"),
	)
	if err != nil {
		return err
	}

	_, err = m.RegisterCallback(
		func(_ context.Context, o metric.Observer) error {
			o.ObserveInt64(gauge, atomic.LoadInt64(&activeConnectionsCount))
			return nil
		},
		gauge,
	)
	if err != nil {
		return err
	}

	return nil
}

func IncrementActiveConnections() {
	atomic.AddInt64(&activeConnectionsCount, 1)
}

func DecrementActiveConnections() {
	atomic.AddInt64(&activeConnectionsCount, -1)
}

func RecordConnectionDuration(ctx context.Context, time float64) error {
	m := otel.Meter("gitlab.com/hmajid2301/banterbus")

	histogram, err := m.Float64Histogram("websocket.connection.duration",
		metric.WithDescription("Time taken for a subscription to be closed."),
		metric.WithUnit("s"),
	)
	if err != nil {
		return err
	}

	histogram.Record(ctx, time)
	return nil
}

func RecordRequestLatency(ctx context.Context, latency float64, messageType string, status string) error {
	m := otel.Meter("gitlab.com/hmajid2301/banterbus")

	histogram, err := m.Float64Histogram("message.processing.duration",
		metric.WithDescription("Time taken to handle message from client."),
		metric.WithUnit("ms"),
		metric.WithExplicitBucketBoundaries([]float64{0.1, 0.5, 1, 2, 5}...),
	)
	if err != nil {
		return err
	}

	histogram.Record(
		ctx,
		latency,
		metric.WithAttributes(attribute.String("message_type", messageType), attribute.String("status", status)),
	)
	return nil
}

func RecordMessageSendLatency(ctx context.Context, latency float64) error {
	m := otel.Meter("gitlab.com/hmajid2301/banterbus")

	histogram, err := m.Float64Histogram("message.sent.duration",
		metric.WithDescription("Time taken to send message from client."),
		metric.WithUnit("ms"),
		metric.WithExplicitBucketBoundaries([]float64{0.1, 0.5, 1, 2, 5}...),
	)
	if err != nil {
		return err
	}

	histogram.Record(
		ctx,
		latency,
	)
	return nil
}

func IncrementMessageSent(ctx context.Context) error {
	m := otel.Meter("gitlab.com/hmajid2301/banterbus")

	counter, err := m.Int64Counter("message.sent.total",
		metric.WithDescription("Total number of messages sent."),
		metric.WithUnit("{call}"),
	)
	if err != nil {
		return err
	}

	counter.Add(ctx, 1)
	return nil
}

func IncrementMessageSentError(ctx context.Context) error {
	m := otel.Meter("gitlab.com/hmajid2301/banterbus")

	counter, err := m.Int64Counter("message.sent.errors.total",
		metric.WithDescription("Total number of messages sent which throw an error."),
		metric.WithUnit("{call}"),
	)
	if err != nil {
		return err
	}

	counter.Add(ctx, 1)
	return nil
}

func IncrementMessageReceived(ctx context.Context, messageType string) error {
	m := otel.Meter("gitlab.com/hmajid2301/banterbus")

	counter, err := m.Int64Counter("message.received.total",
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

	counter, err := m.Int64Counter("message.received.errors.total",
		metric.WithDescription("Total number of messages received which throw an error."),
		metric.WithUnit("{call}"),
	)
	if err != nil {
		return err
	}

	counter.Add(ctx, 1)
	return nil
}

func IncrementReconnectionCount(ctx context.Context) error {
	m := otel.Meter("gitlab.com/hmajid2301/banterbus")

	counter, err := m.Int64Counter("connection.reconnect.total",
		metric.WithDescription("Total number of clients which reconnected."),
		metric.WithUnit("1"),
	)
	if err != nil {
		return err
	}

	counter.Add(ctx, 1)
	return nil
}

func IncrementHandshakeFailures(ctx context.Context, reason string) error {
	m := otel.Meter("gitlab.com/hmajid2301/banterbus")

	counter, err := m.Int64Counter(
		"handshake.failure.total",
		metric.WithDescription(
			"Total number of clients which failed the handshake to upgrade to websocker connection.",
		),
		metric.WithUnit("1"),
	)
	if err != nil {
		return err
	}

	counter.Add(ctx, 1, metric.WithAttributes(attribute.String("reason", reason)))
	return nil
}
