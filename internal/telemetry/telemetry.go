package telemetry

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gitlab.com/hmajid2301/banterbus/internal/telemetry/metrics"
	hostMetrics "go.opentelemetry.io/contrib/instrumentation/host"
	runtimeMetrics "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/contrib/processors/minsev"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func Setup(
	ctx context.Context,
	environment string,
	logLevel minsev.Severity,
) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	res, err := resource.New(
		ctx,
		resource.WithHost(),
		resource.WithContainerID(),
		resource.WithAttributes(
			semconv.ServiceNamespaceKey.String(environment),
			semconv.ServiceNameKey.String("banterbus"),
		),
		// resource.WithSchemaURL("https://gitlab.com/hmajid2301/banterbus"),
	)
	if err != nil {
		handleErr(err)
		return shutdown, err
	}

	tracerProvider, err := newTraceProvider(ctx, res, environment)
	if err != nil {
		handleErr(err)
		return shutdown, err
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	meterProvider, err := newMeterProvider(ctx, res, environment)
	if err != nil {
		handleErr(err)
		return shutdown, err
	}

	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	logProvider, err := newLoggerProvider(ctx, res, logLevel)
	if err != nil {
		handleErr(err)
		return shutdown, err
	}

	shutdownFuncs = append(shutdownFuncs, logProvider.Shutdown)
	global.SetLoggerProvider(logProvider)

	if environment != "test" {
		err = initializeMetrics(ctx)
		if err != nil {
			handleErr(err)
			return shutdown, err
		}
	}

	return shutdown, err
}

func initializeMetrics(ctx context.Context) error {
	if err := metrics.RegisterMetrics(ctx); err != nil {
		return err
	}

	if err := metrics.ActiveConnections(ctx); err != nil {
		return err
	}

	return nil
}

// RecordGameLogicError records a game logic error
func RecordGameLogicError(ctx context.Context) error {
	return metrics.RecordGameLogicError(ctx)
}

// IncrementStateCount increments state count
func IncrementStateCount(ctx context.Context, state string) error {
	return metrics.IncrementStateCount(ctx, state)
}

// RecordStateDuration records state duration
func RecordStateDuration(ctx context.Context, duration float64, state string) error {
	return metrics.RecordStateDuration(ctx, duration, state)
}

// IncrementStateOperationError increments state operation error
func IncrementStateOperationError(ctx context.Context, state string, errReason string) error {
	return metrics.IncrementStateOperationError(ctx, state, errReason)
}

// IncrementActiveConnections increments active connections
func IncrementActiveConnections() {
	metrics.IncrementActiveConnections()
}

// DecrementActiveConnections decrements active connections
func DecrementActiveConnections() {
	metrics.DecrementActiveConnections()
}

// RecordConnectionDuration records connection duration
func RecordConnectionDuration(ctx context.Context, time float64) error {
	return metrics.RecordConnectionDuration(ctx, time)
}

// IncrementReconnectionCount increments reconnection count
func IncrementReconnectionCount(ctx context.Context) error {
	return metrics.IncrementReconnectionCount(ctx)
}

// IncrementHandshakeFailures increments handshake failures
func IncrementHandshakeFailures(ctx context.Context, reason string) error {
	return metrics.IncrementHandshakeFailures(ctx, reason)
}

// IncrementSubscribers increments subscribers
func IncrementSubscribers(ctx context.Context) error {
	return metrics.IncrementSubscribers(ctx)
}

// IncrementMessageSentError increments message sent error
func IncrementMessageSentError(ctx context.Context, messageType ...string) error {
	msgType := "unknown"
	if len(messageType) > 0 {
		msgType = messageType[0]
	}
	return metrics.IncrementMessageSentError(ctx, msgType)
}

// IncrementMessageSent increments message sent
func IncrementMessageSent(ctx context.Context, messageType ...string) error {
	msgType := "unknown"
	if len(messageType) > 0 {
		msgType = messageType[0]
	}
	return metrics.IncrementMessageSent(ctx, msgType)
}

// RecordMessageSendLatency records message send latency
func RecordMessageSendLatency(ctx context.Context, latency float64, messageType ...string) error {
	msgType := "unknown"
	if len(messageType) > 0 {
		msgType = messageType[0]
	}
	return metrics.RecordMessageSendLatency(ctx, latency, msgType)
}

// IncrementMessageReceivedError increments message received error
func IncrementMessageReceivedError(ctx context.Context, messageType ...string) error {
	msgType := "unknown"
	if len(messageType) > 0 {
		msgType = messageType[0]
	}
	return metrics.IncrementMessageReceivedError(ctx, msgType)
}

// IncrementMessageReceived increments message received
func IncrementMessageReceived(ctx context.Context, messageType string) error {
	return metrics.IncrementMessageReceived(ctx, messageType)
}

// RecordRequestLatency records request latency
func RecordRequestLatency(ctx context.Context, latency float64, messageType string, status string) error {
	return metrics.RecordRequestLatency(ctx, latency, messageType, status)
}

// RecordWebSocketError records websocket error
func RecordWebSocketError(ctx context.Context) error {
	return metrics.RecordWebSocketError(ctx)
}

// RecordValidationErrorMetric records validation error
func RecordValidationErrorMetric(ctx context.Context) error {
	return metrics.RecordValidationErrorMetric(ctx)
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTraceProvider(ctx context.Context, res *resource.Resource, environment string) (*trace.TracerProvider, error) {
	var tracerOptions []trace.TracerProviderOption
	tracerOptions = append(tracerOptions, trace.WithResource(res))

	if environment == "test" {
		tracerOptions = append(tracerOptions, trace.WithSampler(trace.AlwaysSample()))
	} else {
		traceExporter, err := otlptracehttp.New(ctx)
		if err != nil {
			return nil, err
		}
		tracerOptions = append(tracerOptions, trace.WithBatcher(traceExporter,
			trace.WithBatchTimeout(time.Second),
		))
	}

	traceProvider := trace.NewTracerProvider(tracerOptions...)
	otel.SetTracerProvider(traceProvider)
	return traceProvider, nil
}

func newMeterProvider(ctx context.Context, res *resource.Resource, environment string) (*metric.MeterProvider, error) {
	var meterProvider *metric.MeterProvider

	if environment == "test" {
		meterProvider = metric.NewMeterProvider(metric.WithResource(res))
	} else {
		metricExporter, err := otlpmetrichttp.New(ctx)
		if err != nil {
			return nil, err
		}

		reader := metric.NewPeriodicReader(metricExporter, metric.WithProducer(runtimeMetrics.NewProducer()))
		meterProvider = metric.NewMeterProvider(metric.WithReader(reader), metric.WithResource(res))

		if err = hostMetrics.Start(hostMetrics.WithMeterProvider(meterProvider)); err != nil {
			return nil, fmt.Errorf("failed to start host metrics: %w", err)
		}
	}

	if err := runtimeMetrics.Start(runtimeMetrics.WithMeterProvider(meterProvider)); err != nil {
		return nil, fmt.Errorf("failed to start runtime metrics: %w", err)
	}

	otel.SetMeterProvider(meterProvider)
	return meterProvider, nil
}

func newLoggerProvider(
	ctx context.Context,
	res *resource.Resource,
	logLevel minsev.Severity,
) (*log.LoggerProvider, error) {
	exporter, err := otlploghttp.New(ctx)
	if err != nil {
		return nil, err
	}

	p := log.NewBatchProcessor(exporter)
	processor := minsev.NewLogProcessor(p, logLevel)
	provider := log.NewLoggerProvider(
		log.WithProcessor(processor),
		log.WithResource(res),
	)
	return provider, nil
}
