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

func NewLogger(logLevel slog.Level) *slog.Logger {
	stdoutHandler := tint.NewHandler(os.Stdout, &tint.Options{
		AddSource:  true,
		TimeFormat: time.Kitchen,
		Level:      logLevel,
	})
	otelHandler := otelslog.NewHandler("banterbus", otelslog.WithSource(true))
	fanoutHandler := slogmulti.Fanout(stdoutHandler, otelHandler)
	ctxHandler := slogctx.NewHandler(fanoutHandler, nil)
	logger := slog.New(ctxHandler)
	return logger
}

func NewLogLevelFromMinSev(severity minsev.Severity) slog.Level {
	switch severity {
	case minsev.SeverityError:
		return slog.LevelError
	case minsev.SeverityWarn:
		return slog.LevelWarn
	case minsev.SeverityInfo:
		return slog.LevelInfo
	case minsev.SeverityDebug:
		return slog.LevelDebug
	default:
		return slog.LevelInfo
	}
}
