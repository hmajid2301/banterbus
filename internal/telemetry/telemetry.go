package telemetry

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

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
	

	"gitlab.com/hmajid2301/banterbus/internal/config"
)

func Setup(
	ctx context.Context,
	environment string,
	logLevel minsev.Severity,
	telemetryConfig config.Telemetry,
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

	tracerProvider, err := newTraceProvider(ctx, res, environment, telemetryConfig)
	if err != nil {
		handleErr(err)
		return shutdown, err
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	meterProvider, err := newMeterProvider(ctx, res, environment, telemetryConfig)
	if err != nil {
		handleErr(err)
		return shutdown, err
	}

	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	logProvider, err := newLoggerProvider(ctx, res, logLevel, telemetryConfig)
	if err != nil {
		handleErr(err)
		return shutdown, err
	}

	shutdownFuncs = append(shutdownFuncs, logProvider.Shutdown)
	global.SetLoggerProvider(logProvider)

	if environment != "test" {
		err = InitializeMetrics(ctx)
		if err != nil {
			handleErr(err)
			return shutdown, err
		}
	}

	return shutdown, err
}

func getOTLPEndpoint() string {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint != "" {
		// Remove protocol prefix for OTLP HTTP exporters
		endpoint = strings.TrimPrefix(endpoint, "http://")
		endpoint = strings.TrimPrefix(endpoint, "https://")
		return endpoint
	}
	// Return empty string if no endpoint is configured
	// This allows disabling OTEL by setting OTEL_EXPORTER_OTLP_ENDPOINT=""
	return ""
}



func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTraceProvider(
	ctx context.Context,
	res *resource.Resource,
	environment string,
	telemetryConfig config.Telemetry,
) (*trace.TracerProvider, error) {
	var tracerOptions []trace.TracerProviderOption
	tracerOptions = append(tracerOptions, trace.WithResource(res))

	if environment == "test" {
		tracerOptions = append(tracerOptions, trace.WithSampler(trace.AlwaysSample()))
	} else {
		endpoint := getOTLPEndpoint()
		if endpoint != "" {
			client := http.DefaultClient
			traceExporter, err := otlptracehttp.New(ctx,
				otlptracehttp.WithEndpoint(endpoint),
				otlptracehttp.WithHTTPClient(client),
			)
			if err != nil {
				return nil, err
			}
			tracerOptions = append(tracerOptions, trace.WithBatcher(traceExporter,
				trace.WithBatchTimeout(time.Second),
			))
		}
	}

	traceProvider := trace.NewTracerProvider(tracerOptions...)
	otel.SetTracerProvider(traceProvider)
	return traceProvider, nil
}

func newMeterProvider(
	ctx context.Context,
	res *resource.Resource,
	environment string,
	telemetryConfig config.Telemetry,
) (*metric.MeterProvider, error) {
	var meterProvider *metric.MeterProvider

	if environment == "test" {
		meterProvider = metric.NewMeterProvider(metric.WithResource(res))
	} else {
		endpoint := getOTLPEndpoint()
		if endpoint != "" {
			client := http.DefaultClient
			metricExporter, err := otlpmetrichttp.New(ctx,
				otlpmetrichttp.WithEndpoint(endpoint),
				otlpmetrichttp.WithHTTPClient(client),
			)
			if err != nil {
				return nil, err
			}

			reader := metric.NewPeriodicReader(metricExporter, metric.WithProducer(runtimeMetrics.NewProducer()))
			meterProvider = metric.NewMeterProvider(metric.WithReader(reader), metric.WithResource(res))

			if err = hostMetrics.Start(hostMetrics.WithMeterProvider(meterProvider)); err != nil {
				return nil, fmt.Errorf("failed to start host metrics: %w", err)
			}
		} else {
			meterProvider = metric.NewMeterProvider(metric.WithResource(res))
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
	telemetryConfig config.Telemetry,
) (*log.LoggerProvider, error) {
	endpoint := getOTLPEndpoint()
	if endpoint == "" {
		// Return a basic provider without OTLP exporter when no endpoint is configured
		provider := log.NewLoggerProvider(log.WithResource(res))
		return provider, nil
	}

	client := http.DefaultClient
	exporter, err := otlploghttp.New(ctx,
		otlploghttp.WithEndpoint(endpoint),
		otlploghttp.WithHTTPClient(client),
	)
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
