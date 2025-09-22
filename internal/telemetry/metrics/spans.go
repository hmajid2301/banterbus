package metrics

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func RecordGameStateTransition(ctx context.Context, fromState, toState string, success bool, playerCount int) error {
	m := otel.Meter("gitlab.com/hmajid2301/banterbus")

	counter, err := m.Int64Counter("game.state_transitions.total",
		metric.WithDescription("Game state transitions by states"),
		metric.WithUnit("{transition}"),
	)
	if err != nil {
		return err
	}

	status := "failure"
	if success {
		status = "success"
	}

	playerBucket := getPlayerCountBucket(playerCount)

	counter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("from_state", fromState),
		attribute.String("to_state", toState),
		attribute.String("status", status),
		attribute.String("player_count_bucket", playerBucket),
	))

	return nil
}

func RecordErrorsByComponent(ctx context.Context, component, errorType string) error {
	m := otel.Meter("gitlab.com/hmajid2301/banterbus")

	counter, err := m.Int64Counter("errors.total",
		metric.WithDescription("Errors by component and type"),
		metric.WithUnit("{error}"),
	)
	if err != nil {
		return err
	}

	counter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("component", component),
		attribute.String("error_type", errorType),
	))

	return nil
}

func RecordWebSocketEvent(ctx context.Context, eventType string, success bool) error {
	m := otel.Meter("gitlab.com/hmajid2301/banterbus")

	counter, err := m.Int64Counter("websocket.events.total",
		metric.WithDescription("WebSocket events by type"),
		metric.WithUnit("{event}"),
	)
	if err != nil {
		return err
	}

	status := "failure"
	if success {
		status = "success"
	}

	messageCategory := getMessageCategory(eventType)

	counter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("event_category", messageCategory),
		attribute.String("status", status),
	))

	return nil
}

func getPlayerCountBucket(count int) string {
	switch {
	case count <= 2:
		return "1-2"
	case count <= 4:
		return "3-4"
	case count <= 8:
		return "5-8"
	default:
		return "8+"
	}
}

func getMessageCategory(messageType string) string {
	switch {
	case messageType == "join_game" || messageType == "leave_game" || messageType == "kick_player":
		return "lobby_management"
	case messageType == "start_game" || messageType == "next_round" || messageType == "end_game":
		return "game_control"
	case messageType == "submit_answer" || messageType == "vote":
		return "player_action"
	case messageType == "ping" || messageType == "heartbeat":
		return "connection_management"
	default:
		return "other"
	}
}
