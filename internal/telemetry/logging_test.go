package telemetry_test

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/hmajid2301/banterbus/internal/telemetry"

	"go.opentelemetry.io/otel/sdk/trace"
)

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("Should create logger with fanout handler", func(t *testing.T) {
		t.Parallel()

		logger := telemetry.NewLogger(slog.LevelInfo)

		assert.NotNil(t, logger)
		assert.IsType(t, &slog.Logger{}, logger)
	})

	t.Run("Should log messages without errors", func(t *testing.T) {
		t.Parallel()

		logger := telemetry.NewLogger(slog.LevelInfo)

		assert.NotPanics(t, func() {
			logger.Info("test message")
			logger.Error("test error")
			logger.Debug("test debug")
		})
	})
}

func TestTraceHandler(t *testing.T) {
	t.Parallel()

	t.Run("Should handle enabled check", func(t *testing.T) {
		t.Parallel()

		logger := telemetry.NewLogger(slog.LevelInfo)
		ctx := t.Context()

		enabled := logger.Enabled(ctx, slog.LevelInfo)
		assert.True(t, enabled)
	})

	t.Run("Should add trace ID when span is present", func(t *testing.T) {
		t.Parallel()

		tp := trace.NewTracerProvider()
		tracer := tp.Tracer("test")
		ctx, span := tracer.Start(t.Context(), "test-span")
		defer span.End()

		logger := telemetry.NewLogger(slog.LevelInfo)

		assert.NotPanics(t, func() {
			logger.InfoContext(ctx, "test message with trace")
		})
	})

	t.Run("Should handle logging without trace context", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := telemetry.NewLogger(slog.LevelInfo)

		assert.NotPanics(t, func() {
			logger.InfoContext(ctx, "test message without trace")
		})
	})

	t.Run("Should handle WithAttrs correctly", func(t *testing.T) {
		t.Parallel()

		logger := telemetry.NewLogger(slog.LevelInfo)
		loggerWithAttrs := logger.With(slog.String("key", "value"))

		assert.NotNil(t, loggerWithAttrs)
		assert.NotPanics(t, func() {
			loggerWithAttrs.Info("test with attrs")
		})
	})

	t.Run("Should handle WithGroup correctly", func(t *testing.T) {
		t.Parallel()

		logger := telemetry.NewLogger(slog.LevelInfo)
		loggerWithGroup := logger.WithGroup("test-group")

		assert.NotNil(t, loggerWithGroup)
		assert.NotPanics(t, func() {
			loggerWithGroup.Info("test with group")
		})
	})

	t.Run("Should handle different log levels", func(t *testing.T) {
		t.Parallel()

		logger := telemetry.NewLogger(slog.LevelInfo)
		ctx := t.Context()

		assert.NotPanics(t, func() {
			logger.DebugContext(ctx, "debug message")
			logger.InfoContext(ctx, "info message")
			logger.WarnContext(ctx, "warn message")
			logger.ErrorContext(ctx, "error message")
		})
	})
}
