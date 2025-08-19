package logging

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
	slogmulti "github.com/samber/slog-multi"
	"go.opentelemetry.io/contrib/bridges/otelslog"
)

func New() *slog.Logger {
	// INFO: print to stdout for fly.io logs quickly debugging without needing to go to otel.
	stdoutHandler := tint.NewHandler(os.Stdout, &tint.Options{
		AddSource:  true,
		TimeFormat: time.Kitchen,
	})
	otelHandler := otelslog.NewHandler("banterbus", otelslog.WithSource(true))
	fanoutHandler := slogmulti.Fanout(stdoutHandler, otelHandler)

	logger := slog.New(fanoutHandler)
	return logger
}
