package service

import (
	"context"
	"time"
)

type RoundService struct {
	store Storer
}

func NewRoundService(store Storer) *RoundService {
	return &RoundService{store: store}
}

func (r *RoundService) SubmitAnswer(ctx context.Context, playerID string, answer string, submittedAt time.Time) error {
	return r.store.SubmitAnswer(ctx, playerID, answer, submittedAt)
}
