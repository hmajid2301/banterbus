package logging

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
	slogmulti "github.com/samber/slog-multi"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/trace"
)

func New() *slog.Logger {
	// INFO: print to stdout for fly.io logs quickly debugging without needing to go to otel.
	stdoutHandler := tint.NewHandler(os.Stdout, &tint.Options{
		AddSource:  true,
		TimeFormat: time.Kitchen,
	})
	otelHandler := otelslog.NewHandler("banterbus", otelslog.WithSource(true))

	traceStdoutHandler := &traceHandler{handler: stdoutHandler}
	traceOtelHandler := &traceHandler{handler: otelHandler}

	fanoutHandler := slogmulti.Fanout(traceStdoutHandler, traceOtelHandler)

	logger := slog.New(fanoutHandler)
	return logger
}

type traceHandler struct {
	handler slog.Handler
}

func (h *traceHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *traceHandler) Handle(ctx context.Context, record slog.Record) error {
	if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		traceID := span.SpanContext().TraceID().String()
		record.AddAttrs(slog.String("trace_id", traceID))
	}

	return h.handler.Handle(ctx, record)
}

func (h *traceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &traceHandler{handler: h.handler.WithAttrs(attrs)}
}

func (h *traceHandler) WithGroup(name string) slog.Handler {
	return &traceHandler{handler: h.handler.WithGroup(name)}
}
