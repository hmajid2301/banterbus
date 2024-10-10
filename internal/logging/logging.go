package logging

import (
	"log/slog"
	"os"

	slogctx "github.com/veqryn/slog-context"
)

func New(logLevel slog.Level, defaultAttrs []slog.Attr) *slog.Logger {
	var handler slog.Handler
	opts := slog.HandlerOptions{
		AddSource: true,
		Level:     logLevel,
	}

	if os.Getenv("BANTERBUS_ENVIRONMENT") == "production" {
		handler = slog.NewJSONHandler(os.Stdout, &opts).WithAttrs(defaultAttrs)
	} else {
		handler = NewPrettyHandler(os.Stdout, PrettyHandlerOptions{SlogOpts: opts})
	}

	customHandler := slogctx.NewHandler(handler, nil)
	logger := slog.New(customHandler)
	return logger
}
