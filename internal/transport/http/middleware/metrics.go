package middleware

import (
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

var (
	metricsOnce  sync.Once
	requestCount metric.Int64Counter
	latencyHist  metric.Float64Histogram
	requestSize  metric.Int64Histogram
	responseSize metric.Int64Histogram
	metricsRng   = rand.New(rand.NewSource(time.Now().UnixNano()))
)

type metricsResponseWriter struct {
	*responseWriter
	size int
}

func (mrw *metricsResponseWriter) Write(data []byte) (int, error) {
	size, err := mrw.responseWriter.Write(data)
	mrw.size += size
	return size, err
}

func wrapResponseWriterForMetrics(w http.ResponseWriter) *metricsResponseWriter {
	return &metricsResponseWriter{
		responseWriter: wrapResponseWriter(w),
		size:           0,
	}
}

func (m Middleware) Metrics(next http.Handler) http.Handler {
	metricsOnce.Do(func() {
		meter := otel.GetMeterProvider().Meter("http.metrics")

		var err error
		requestCount, err = meter.Int64Counter(
			"http.server.request_count",
			metric.WithUnit("1"),
			metric.WithDescription("Number of HTTP requests"),
		)
		if err != nil {
			otel.Handle(err)
		}

		latencyHist, err = meter.Float64Histogram(
			"http.server.duration",
			metric.WithUnit("ms"),
			metric.WithDescription("HTTP request duration"),
		)
		if err != nil {
			otel.Handle(err)
		}

		requestSize, err = meter.Int64Histogram(
			"http.server.request.size",
			metric.WithUnit("By"),
			metric.WithDescription("Request body size"),
		)
		if err != nil {
			otel.Handle(err)
		}

		responseSize, err = meter.Int64Histogram(
			"http.server.response.size",
			metric.WithUnit("By"),
			metric.WithDescription("Response body size"),
		)
		if err != nil {
			otel.Handle(err)
		}
	})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasPrefix(path, "/static") || path == "/readiness" || path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		ww := wrapResponseWriterForMetrics(w)
		next.ServeHTTP(ww, r)

		duration := time.Since(start).Milliseconds()
		statusCode := ww.Status()
		if statusCode == 0 {
			statusCode = http.StatusOK
		}

		attrs := []attribute.KeyValue{
			semconv.HTTPRequestMethodKey.String(r.Method),
			semconv.HTTPResponseStatusCode(statusCode),
			semconv.HTTPRoute(r.URL.EscapedPath()),
		}

		ctx := r.Context()
		if requestCount != nil {
			requestCount.Add(ctx, 1, metric.WithAttributes(attrs...))
		}
		if latencyHist != nil {
			latencyHist.Record(ctx, float64(duration), metric.WithAttributes(attrs...))
		}
		if requestSize != nil && r.ContentLength >= 0 {
			requestSize.Record(ctx, r.ContentLength, metric.WithAttributes(attrs...))
		}
		if responseSize != nil {
			responseSize.Record(ctx, int64(ww.size), metric.WithAttributes(attrs...))
		}
	})
}
