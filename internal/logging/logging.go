package logging

import (
	"log/slog"
	"os"
	"regexp"

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

func StripSVGData(message string) string {
	re := regexp.MustCompile(`data:image/svg\+xml;base64,[A-Za-z0-9+/=]+`)
	return re.ReplaceAllString(message, "")
}
