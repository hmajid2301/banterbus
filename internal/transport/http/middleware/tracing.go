package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

func (Middleware) Tracing(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		propagator := otel.GetTextMapPropagator()
		ctx = propagator.Extract(ctx, propagation.HeaderCarrier(r.Header))

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

		defer span.End()

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

		combinedCtx := txn.Context()
		h.ServeHTTP(w, r.WithContext(combinedCtx))
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
