package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"

	"gitlab.com/hmajid2301/banterbus/internal/telemetry"
)

func (Middleware) Tracing(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		propagator := otel.GetTextMapPropagator()
		ctx = propagator.Extract(ctx, propagation.HeaderCarrier(r.Header))

		// Extract test name from baggage first (preferred method)
		bag := baggage.FromContext(ctx)
		testNameFromBaggage := bag.Member("test_name").Value()

		// If not in baggage, check headers as fallback
		if testNameFromBaggage == "" {
			if testName := r.Header.Get("X-Test-Name"); testName != "" {
				ctx = telemetry.AddTestNameToBaggage(ctx, testName)
			}
		}

		// If still not found, check query parameters as final fallback
		if testNameFromBaggage == "" {
			if testName := r.URL.Query().Get("test_name"); testName != "" {
				ctx = telemetry.AddTestNameToBaggage(ctx, testName)
			}
		}

		tracer := otel.Tracer("banterbus-backend-http")
		spanName := getSpanNameForRequest(r)
		sentryTrace := r.Header.Get("sentry-trace")
		var sentryTraceID string
		if sentryTrace != "" {
			parts := strings.Split(sentryTrace, "-")
			if len(parts) > 0 {
				sentryTraceID = parts[0]
			}
		}

		ctx, span := tracer.Start(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				semconv.HTTPRoute(getHTTPRoute(r.URL.Path)),
				semconv.HTTPRequestMethodOriginal(r.Method),
				semconv.URLFull(r.URL.String()),
				semconv.URLPath(r.URL.Path),
				semconv.UserAgentOriginal(r.UserAgent()),
				semconv.NetworkProtocolName("http"),
				semconv.NetworkProtocolVersion("1.1"),
				semconv.ServerAddress(r.Host),
				attribute.String("component", "http-handler"),
			),
		)

		if sentryTraceID != "" {
			span.SetAttributes(attribute.String("sentry.trace_id", sentryTraceID))
		}

		start := time.Now()
		rw := wrapResponseWriter(w)

		defer func() {
			duration := time.Since(start).Seconds()
			statusCode := rw.Status()
			if statusCode == 0 {
				statusCode = 200
			}
			telemetry.RecordHTTPRequest(ctx, getHTTPRoute(r.URL.Path), r.Method, statusCode, duration)
			span.End()
		}()

		opts := []sentry.SpanOption{
			sentry.WithOpName("http.server"),
			sentry.ContinueFromRequest(r),
			sentry.WithTransactionSource(sentry.SourceURL),
		}
		txn := sentry.StartTransaction(ctx,
			fmt.Sprintf("%s %s", r.Method, r.URL.Path),
			opts...,
		)
		defer txn.Finish()

		h.ServeHTTP(rw, r.WithContext(ctx))
	})
}

func getSpanNameForRequest(r *http.Request) string {
	path := r.URL.Path
	method := r.Method

	if strings.HasPrefix(path, "/static/") {
		return "GET /static/*"
	}

	if r.Header.Get("Upgrade") == "websocket" {
		return "WebSocket /ws"
	}

	switch {
	case strings.HasPrefix(path, "/join/"):
		return "GET /join/{code}"
	case path == "/":
		return "GET /"
	case path == "/ws":
		return "WebSocket /ws"
	default:
		return method + " " + path
	}
}

func getHTTPRoute(path string) string {
	if strings.HasPrefix(path, "/static/") {
		return "/static/*"
	}

	switch {
	case strings.HasPrefix(path, "/join/"):
		return "/join/{code}"
	case path == "/":
		return "/"
	case path == "/ws":
		return "/ws"
	default:
		return path
	}
}
