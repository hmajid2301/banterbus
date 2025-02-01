package telemetry

import (
	"context"
	"errors"
	"fmt"
	"time"

	hostMetrics "go.opentelemetry.io/contrib/instrumentation/host"
	runtimeMetrics "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func SetupOTelSDK(ctx context.Context, environment string) (shutdown func(context.Context) error, err error) {
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

	// TODO: use ldflags to make these dynamic
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

	tracerProvider, err := newTraceProvider(ctx, res)
	if err != nil {
		handleErr(err)
		return shutdown, err
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	meterProvider, err := newMeterProvider(ctx, res)
	if err != nil {
		handleErr(err)
		return shutdown, err
	}

	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)
	//
	// logProvider, err := newLogProvider(ctx, res)
	// if err != nil {
	// 	handleErr(err)
	// 	return shutdown, err
	// }
	//
	// shutdownFuncs = append(shutdownFuncs, logProvider.Shutdown)
	// global.SetLoggerProvider(logProvider)

	return shutdown, err
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTraceProvider(ctx context.Context, res *resource.Resource) (*trace.TracerProvider, error) {
	traceExporter, err := otlptracehttp.New(ctx)

	if err != nil {
		return nil, err
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter,
			trace.WithBatchTimeout(time.Second),
		),
		trace.WithResource(res),
	)
	otel.SetTracerProvider(traceProvider)
	return traceProvider, nil
}

func newMeterProvider(ctx context.Context, res *resource.Resource) (*metric.MeterProvider, error) {
	metricExporter, err := otlpmetrichttp.New(ctx)
	if err != nil {
		return nil, err
	}

	reader := metric.NewPeriodicReader(metricExporter, metric.WithProducer(runtimeMetrics.NewProducer()))
	meterProvider := metric.NewMeterProvider(metric.WithReader(reader), metric.WithResource(res))

	if err = runtimeMetrics.Start(runtimeMetrics.WithMeterProvider(meterProvider)); err != nil {
		return nil, fmt.Errorf("failed to start runtime metrics: %v", err)
	}

	if err = hostMetrics.Start(hostMetrics.WithMeterProvider(meterProvider)); err != nil {
		return nil, fmt.Errorf("failed to start host metrics: %v", err)
	}

	otel.SetMeterProvider(meterProvider)
	return meterProvider, nil
}

// TODO: enable when the slogotel logger has more control over say log level adding source etc
// func newLogProvider(ctx context.Context, res *resource.Resource) (*log.LoggerProvider, error) {
// 	logExporter, err := otlploghttp.New(ctx)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	loggerProvider := log.NewLoggerProvider(
// 		log.WithProcessor(log.NewBatchProcessor(logExporter)),
// 		log.WithResource(res),
// 	)
// 	return loggerProvider, nil
// }
