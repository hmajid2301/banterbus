package logging

import (
	"log/slog"
	"os"
	"regexp"
	"time"

	"github.com/lmittmann/tint"
	slogctx "github.com/veqryn/slog-context"
)

func New(logLevel slog.Level, defaultAttrs []slog.Attr) *slog.Logger {
	var handler slog.Handler

	if os.Getenv("BANTERBUS_ENVIRONMENT") == "production" {
		opts := slog.HandlerOptions{
			AddSource: true,
			Level:     logLevel,
		}
		handler = slog.NewJSONHandler(os.Stdout, &opts).WithAttrs(defaultAttrs)
	} else {
		handler = tint.NewHandler(os.Stdout, &tint.Options{
			AddSource:  true,
			Level:      logLevel,
			TimeFormat: time.Kitchen,
		})
	}

	customHandler := slogctx.NewHandler(handler, nil)
	logger := slog.New(customHandler)
	return logger
}

func StripSVGData(message string) string {
	re := regexp.MustCompile(`data:image/svg\+xml;base64,[A-Za-z0-9+/=]+`)
	return re.ReplaceAllString(message, "")
}
