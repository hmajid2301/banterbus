package logging

import (
	"log/slog"
	"os"

	slogotel "github.com/remychantenay/slog-otel"
)

func New(logLevel slog.Level) *slog.Logger {
	var handler slog.Handler
	if os.Getenv("BANTERBUS_ENVIRONMENT") == "production" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     logLevel,
		})
	} else {
		handler = NewPrettyHandler(os.Stdout, PrettyHandlerOptions{
			SlogOpts: slog.HandlerOptions{
				AddSource: true,
				Level:     logLevel,
			},
		})
	}
	slog.SetDefault(slog.New(slogotel.OtelHandler{
		Next: handler,
	}))

	logger := slog.Default()
	return logger
}
