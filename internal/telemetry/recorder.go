package telemetry

import (
	"context"

	"gitlab.com/hmajid2301/banterbus/internal/telemetry/metrics"
)

type Recorder struct{}

func NewRecorder() *Recorder {
	return &Recorder{}
}

func (r *Recorder) RecordGameCompletion(ctx context.Context, success bool, duration float64, playerCount int) error {
	return metrics.RecordGameCompletion(ctx, success, duration, playerCount)
}

func (r *Recorder) RecordLobbyOperation(ctx context.Context, operation string, success bool) error {
	return metrics.RecordLobbyHealth(ctx, operation, success)
}

func (r *Recorder) RecordPlayerExperience(ctx context.Context, satisfaction string) error {
	return metrics.RecordPlayerExperience(ctx, satisfaction)
}

func (r *Recorder) UpdateActiveGameCounts(ctx context.Context, lobbies, games int64) error {
	return metrics.UpdateActiveGameCounts(ctx, lobbies, games)
}

func (r *Recorder) RecordPlayerDisconnection(ctx context.Context, reason string, gameState string) error {
	return metrics.IncrementPlayerDisconnections(ctx, reason, gameState)
}
