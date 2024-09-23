package logger

import (
	"log/slog"
	"os"

	slogotel "github.com/remychantenay/slog-otel"
)

func New() *slog.Logger {
	var handler slog.Handler
	if os.Getenv("BANTERBUS_ENVIRONMENT") == "production" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
		})
	} else {
		handler = NewPrettyHandler(os.Stdout, PrettyHandlerOptions{})
	}
	slog.SetDefault(slog.New(slogotel.OtelHandler{
		Next: handler,
	}))

	logger := slog.Default()
	return logger
}
