package telemetry_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/hmajid2301/banterbus/internal/telemetry"
	"gitlab.com/hmajid2301/banterbus/internal/telemetry/metrics"
)

func TestTelemetryMetrics(t *testing.T) {
	t.Parallel()

	t.Run("Should initialize metrics middleware without error", func(t *testing.T) {
		t.Parallel()

		// Test that we can create middleware without panicking
		middleware, err := metrics.NewGameStateMiddleware()
		assert.NoError(t, err)
		assert.NotNil(t, middleware)
	})

	t.Run("Should record HTTP request metrics", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// Test that metrics recording doesn't panic
		err := metrics.RecordHTTPRequest(ctx, "/test", "GET", 200, 100.0)
		assert.NoError(t, err)
	})

	t.Run("Should record game state transition metrics", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// Test that state transition recording doesn't panic
		err := metrics.RecordGameStateTransition(ctx, "CREATED", "STARTED", true, 3)
		assert.NoError(t, err)
	})

	t.Run("Should initialize metrics without error", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// Test that metrics initialization doesn't panic
		err := metrics.InitializeMetrics(ctx)
		assert.NoError(t, err)
	})

	t.Run("Should record state duration metrics", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// Test state duration recording
		err := metrics.RecordStateDuration(ctx, 30.5, "LOBBY")
		assert.NoError(t, err)
	})

	t.Run("Should record WebSocket events", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// Test WebSocket event recording
		err := metrics.RecordWebSocketEvent(ctx, "message_sent", true)
		assert.NoError(t, err)
	})
}

func TestTelemetryRecorder(t *testing.T) {
	t.Parallel()

	t.Run("Should create new recorder", func(t *testing.T) {
		t.Parallel()

		recorder := telemetry.NewRecorder()
		assert.NotNil(t, recorder)
	})
}
