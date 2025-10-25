//go:build dev

package telemetry

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
	slogmulti "github.com/samber/slog-multi"
	slogctx "github.com/veqryn/slog-context"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/processors/minsev"
)

func NewLogger(logLevel minsev.Severity) *slog.Logger {
	level := newLogLevelFromMinSev(logLevel)
	stdoutHandler := tint.NewHandler(os.Stdout, &tint.Options{
		AddSource:  true,
		TimeFormat: time.Kitchen,
		Level:      level,
	})
	otelHandler := otelslog.NewHandler("banterbus", otelslog.WithSource(true))
	fanoutHandler := slogmulti.Fanout(stdoutHandler, otelHandler)
	ctxHandler := slogctx.NewHandler(fanoutHandler, nil)
	logger := slog.New(ctxHandler)
	return logger
}
