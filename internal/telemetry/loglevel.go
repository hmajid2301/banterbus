package telemetry

import (
	"log/slog"

	"go.opentelemetry.io/contrib/processors/minsev"
)

func newLogLevelFromMinSev(severity minsev.Severity) slog.Level {
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
