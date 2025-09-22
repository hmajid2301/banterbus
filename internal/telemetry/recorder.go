package telemetry

import (
	"context"

	"gitlab.com/hmajid2301/banterbus/internal/telemetry/metrics"
)

// Recorder provides simple metrics recording for services
type Recorder struct{}

func NewRecorder() *Recorder {
	return &Recorder{}
}

// RecordGameCompletion records game completion metrics
func (r *Recorder) RecordGameCompletion(ctx context.Context, success bool, duration float64, playerCount int) error {
	return metrics.RecordGameCompletion(ctx, success, duration, playerCount)
}

// RecordLobbyOperation records lobby operation metrics
func (r *Recorder) RecordLobbyOperation(ctx context.Context, operation string, success bool) error {
	return metrics.RecordLobbyHealth(ctx, operation, success)
}

// RecordPlayerExperience records player satisfaction
func (r *Recorder) RecordPlayerExperience(ctx context.Context, satisfaction string) error {
	return metrics.RecordPlayerExperience(ctx, satisfaction)
}

// UpdateActiveGameCounts updates active game and lobby counts
func (r *Recorder) UpdateActiveGameCounts(ctx context.Context, lobbies, games int64) error {
	return metrics.UpdateActiveGameCounts(ctx, lobbies, games)
}

// RecordPlayerDisconnection records player disconnection
func (r *Recorder) RecordPlayerDisconnection(ctx context.Context, reason string, gameState string) error {
	return metrics.IncrementPlayerDisconnections(ctx, reason, gameState)
}
