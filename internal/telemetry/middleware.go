package telemetry

import (
	"context"

	"gitlab.com/hmajid2301/banterbus/internal/telemetry/metrics"
)

// Re-export essential middleware functions from metrics package
func NewGameStateMiddleware() (*metrics.GameStateMiddleware, error) {
	return metrics.NewGameStateMiddleware()
}

func RecordHTTPRequest(ctx context.Context, endpoint, method string, statusCode int, duration float64) error {
	return metrics.RecordHTTPRequest(ctx, endpoint, method, statusCode, duration)
}

func InitializeMetrics(ctx context.Context) error {
	return metrics.InitializeMetrics(ctx)
}
