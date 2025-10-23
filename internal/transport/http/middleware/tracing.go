package middleware

import (
	"net/http"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
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

		tracer := otel.Tracer("banterbus-backend-http")
		spanName := getSpanNameForRequest(r)

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
