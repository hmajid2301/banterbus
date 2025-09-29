package metrics

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
		metric.WithUnit("{connection}"),
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
		metric.WithDescription("WebSocket connection duration"),
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

	histogram, err := m.Float64Histogram("websocket.message.processing.duration",
		metric.WithDescription("Time taken to process WebSocket message"),
		metric.WithUnit("ms"),
		metric.WithExplicitBucketBoundaries([]float64{0.1, 0.5, 1, 2, 5}...),
	)
	if err != nil {
		return err
	}

	histogram.Record(
		ctx,
		latency,
		metric.WithAttributes(attribute.String("message.type", messageType), attribute.String("status", status)),
	)
	return nil
}

func RecordMessageSendLatency(ctx context.Context, latency float64, messageType string) error {
	m := otel.Meter("gitlab.com/hmajid2301/banterbus")

	histogram, err := m.Float64Histogram("websocket.message.sent.duration",
		metric.WithDescription("Time taken to send WebSocket message"),
		metric.WithUnit("ms"),
		metric.WithExplicitBucketBoundaries([]float64{0.1, 0.5, 1, 2, 5}...),
	)
	if err != nil {
		return err
	}

	histogram.Record(
		ctx,
		latency,
		metric.WithAttributes(attribute.String("message.type", messageType)),
	)
	return nil
}

func IncrementMessageSent(ctx context.Context, messageType string) error {
	m := otel.Meter("gitlab.com/hmajid2301/banterbus")

	counter, err := m.Int64Counter("websocket.messages.sent.total",
		metric.WithDescription("Total number of WebSocket messages sent"),
		metric.WithUnit("{message}"),
	)
	if err != nil {
		return err
	}

	counter.Add(ctx, 1, metric.WithAttributes(attribute.String("message.type", messageType)))
	return nil
}

func IncrementMessageSentError(ctx context.Context, messageType string) error {
	m := otel.Meter("gitlab.com/hmajid2301/banterbus")

	counter, err := m.Int64Counter("websocket.messages.sent.errors.total",
		metric.WithDescription("Total number of WebSocket message send errors"),
		metric.WithUnit("{error}"),
	)
	if err != nil {
		return err
	}

	counter.Add(ctx, 1, metric.WithAttributes(attribute.String("message.type", messageType)))
	return nil
}

func IncrementMessageReceived(ctx context.Context, messageType string) error {
	m := otel.Meter("gitlab.com/hmajid2301/banterbus")

	counter, err := m.Int64Counter("websocket.messages.received.total",
		metric.WithDescription("Total number of WebSocket messages received"),
		metric.WithUnit("{message}"),
	)
	if err != nil {
		return err
	}

	counter.Add(ctx, 1, metric.WithAttributes(attribute.String("message.type", messageType)))
	return nil
}

func IncrementMessageReceivedError(ctx context.Context, messageType string) error {
	m := otel.Meter("gitlab.com/hmajid2301/banterbus")

	counter, err := m.Int64Counter("websocket.messages.received.errors.total",
		metric.WithDescription("Total number of WebSocket message receive errors"),
		metric.WithUnit("{error}"),
	)
	if err != nil {
		return err
	}

	counter.Add(ctx, 1, metric.WithAttributes(attribute.String("message.type", messageType)))
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

func IncrementPlayerDisconnections(ctx context.Context, reason string, gameState string) error {
	m := otel.Meter("gitlab.com/hmajid2301/banterbus")

	counter, err := m.Int64Counter("player.disconnection.count",
		metric.WithDescription("Number of unexpected player disconnections"),
		metric.WithUnit("{disconnection}"),
	)
	if err != nil {
		return err
	}

	counter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("reason", reason),
		attribute.String("game_state", gameState),
	))
	return nil
}

func RecordError(ctx context.Context, errorType, component string) error {
	m := otel.Meter("gitlab.com/hmajid2301/banterbus")

	counter, err := m.Int64Counter("errors_total",
		metric.WithDescription("Total number of errors thrown by the system"),
		metric.WithUnit("{error}"),
	)
	if err != nil {
		return err
	}

	counter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("error_type", errorType),
		attribute.String("component", component),
	))
	return nil
}

func RecordDatabaseError(ctx context.Context) error {
	return RecordError(ctx, "database_error", "database")
}

func RecordWebSocketError(ctx context.Context) error {
	return RecordError(ctx, "websocket_error", "websocket")
}

func RecordGameLogicError(ctx context.Context) error {
	return RecordError(ctx, "game_logic_error", "game")
}

func RecordValidationErrorMetric(ctx context.Context) error {
	return RecordError(ctx, "validation_error", "validation")
}
