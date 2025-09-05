//go:build integration

package telemetry

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/contrib/processors/minsev"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

func TestIntegrationSetup(t *testing.T) {
	t.Parallel()

	t.Run("Should setup telemetry in test environment", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		shutdown, err := Setup(ctx, "test", minsev.SeverityDebug)

		require.NoError(t, err)
		assert.NotNil(t, shutdown)

		err = shutdown(ctx)
		assert.NoError(t, err)
	})

	t.Run("Should setup telemetry in development environment", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		shutdown, err := Setup(ctx, "development", minsev.SeverityDebug)

		require.NoError(t, err)
		assert.NotNil(t, shutdown)

		err = shutdown(ctx)
		assert.NoError(t, err)
	})

	t.Run("Should setup telemetry in production environment", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		shutdown, err := Setup(ctx, "production", minsev.SeverityDebug)

		require.NoError(t, err)
		assert.NotNil(t, shutdown)

		err = shutdown(ctx)
		assert.NoError(t, err)
	})

	t.Run("Should handle context cancellation during setup", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		shutdown, err := Setup(ctx, "test", minsev.SeverityDebug)

		if err != nil {
			assert.Contains(t, err.Error(), "context")
		} else {
			assert.NotNil(t, shutdown)
			err = shutdown(t.Context())
			assert.NoError(t, err)
		}
	})
}

func TestIntegrationSetupShutdownBehavior(t *testing.T) {
	t.Parallel()

	t.Run("Should handle shutdown properly", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		shutdown, err := Setup(ctx, "test", minsev.SeverityDebug)
		require.NoError(t, err)

		// Call shutdown
		err = shutdown(ctx)
		assert.NoError(t, err)

		// Second shutdown should still work (idempotent)
		err = shutdown(ctx)
		assert.NoError(t, err)
	})

	t.Run("Should handle shutdown with timeout", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		shutdown, err := Setup(ctx, "test", minsev.SeverityDebug)
		require.NoError(t, err)

		// Test shutdown with timeout
		timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		err = shutdown(timeoutCtx)
		assert.NoError(t, err)
	})
}

func TestIntegrationSetupMultipleSetups(t *testing.T) {
	t.Parallel()

	t.Run("Should handle multiple setup calls", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// First setup
		shutdown1, err1 := Setup(ctx, "test", minsev.SeverityDebug)
		require.NoError(t, err1)

		// Second setup should also work
		shutdown2, err2 := Setup(ctx, "test", minsev.SeverityDebug)
		require.NoError(t, err2)

		// Both shutdowns should work
		err1 = shutdown1(ctx)
		assert.NoError(t, err1)

		err2 = shutdown2(ctx)
		assert.NoError(t, err2)
	})
}

func TestSetupWithTelemetryDisabled(t *testing.T) {
	t.Parallel()

	t.Run("Should disable telemetry successfully", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		shutdown, err := Setup(ctx, "test", true)
		require.NoError(t, err)

		// Shutdown should be callable and return no error
		err = shutdown(ctx)
		assert.NoError(t, err)

		// Multiple shutdowns should not cause issues
		err = shutdown(ctx)
		assert.NoError(t, err)
	})
}

func TestSetupWithTestEnvironment(t *testing.T) {
	t.Parallel()

	t.Run("Should setup telemetry for test environment", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		shutdown, err := Setup(ctx, "test", minsev.SeverityDebug)
		require.NoError(t, err)
		t.Cleanup(func() { shutdown(ctx) })

		// Verify propagator is set
		propagator := otel.GetTextMapPropagator()
		assert.NotNil(t, propagator)

		// Verify tracer provider is set and is the correct type
		tracerProvider := otel.GetTracerProvider()
		assert.NotNil(t, tracerProvider)
		_, ok := tracerProvider.(*trace.TracerProvider)
		assert.True(t, ok)

		// Verify meter provider is set and is the correct type
		meterProvider := otel.GetMeterProvider()
		assert.NotNil(t, meterProvider)
		_, ok = meterProvider.(*metric.MeterProvider)
		assert.True(t, ok)
	})
}

func TestSetupTracingFunctionality(t *testing.T) {
	t.Parallel()

	t.Run("Should create and manage traces correctly", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		shutdown, err := Setup(ctx, "test", minsev.SeverityDebug)
		require.NoError(t, err)
		t.Cleanup(func() { shutdown(ctx) })

		tracer := otel.Tracer("banterbus-test")

		// Test creating a tracer
		assert.NotNil(t, tracer)

		// Test creating a span
		ctx, span := tracer.Start(ctx, "test-operation")
		assert.NotNil(t, span)
		defer span.End()

		// Test creating a child span
		_, childSpan := tracer.Start(ctx, "child-operation")
		assert.NotNil(t, childSpan)
		defer childSpan.End()

		// Spans should be different
		assert.NotEqual(t, span, childSpan)

		// Test span context
		spanContext := span.SpanContext()
		assert.True(t, spanContext.IsValid())
		assert.True(t, spanContext.HasTraceID())
		assert.True(t, spanContext.HasSpanID())

		childSpanContext := childSpan.SpanContext()
		assert.True(t, childSpanContext.IsValid())
		assert.Equal(t, spanContext.TraceID(), childSpanContext.TraceID())
		assert.NotEqual(t, spanContext.SpanID(), childSpanContext.SpanID())
	})
}

func TestSetupMetricsFunctionality(t *testing.T) {
	t.Parallel()

	t.Run("Should create and use metrics correctly", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		shutdown, err := Setup(ctx, "test", minsev.SeverityDebug)
		require.NoError(t, err)
		t.Cleanup(func() { shutdown(ctx) })

		meter := otel.Meter("banterbus-test")

		// Test counter
		counter, err := meter.Int64Counter("test_requests_total")
		require.NoError(t, err)
		assert.NotNil(t, counter)
		counter.Add(ctx, 1)
		counter.Add(ctx, 5)

		// Test histogram
		histogram, err := meter.Float64Histogram("test_request_duration")
		require.NoError(t, err)
		assert.NotNil(t, histogram)
		histogram.Record(ctx, 0.1)
		histogram.Record(ctx, 0.25)

		// Test gauge
		gauge, err := meter.Int64UpDownCounter("test_active_connections")
		require.NoError(t, err)
		assert.NotNil(t, gauge)
		gauge.Add(ctx, 10)
		gauge.Add(ctx, -2)
	})
}

func TestSetupPropagatorConfiguration(t *testing.T) {
	t.Parallel()

	t.Run("Should configure propagator correctly", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		shutdown, err := Setup(ctx, "test", minsev.SeverityDebug)
		require.NoError(t, err)
		t.Cleanup(func() { shutdown(ctx) })

		propagator := otel.GetTextMapPropagator()
		assert.NotNil(t, propagator)

		// Test that it's a composite propagator with the expected fields
		fields := propagator.Fields()
		assert.Contains(t, fields, "traceparent")
		assert.Contains(t, fields, "baggage")
	})
}
