package service

import (
	"context"
	"fmt"
	"time"

	sqlc "gitlab.com/hmajid2301/banterbus/internal/store/db"
)

type RoundService struct {
	store      Storer
	randomizer Randomizer
}

func NewRoundService(store Storer, randomizer Randomizer) *RoundService {
	return &RoundService{store: store, randomizer: randomizer}
}

func (r *RoundService) SubmitAnswer(ctx context.Context, playerID string, answer string, submittedAt time.Time) error {
	room, err := r.store.GetRoomByPlayerID(ctx, playerID)
	if err != nil {
		return err
	}

	if room.RoomState != sqlc.ROOMSTATE_PLAYING.String() {
		return fmt.Errorf("room is not in PLAYING state")
	}

	round, err := r.store.GetLatestRoundByPlayerID(ctx, playerID)
	if err != nil {
		return err
	}

	if submittedAt.After(round.SubmitDeadline) {
		return fmt.Errorf("answer submission deadline has passed")
	}

	_, err = r.store.AddFibbingItAnswer(ctx, sqlc.AddFibbingItAnswerParams{
		ID:       r.randomizer.GetID(),
		RoundID:  round.ID,
		PlayerID: playerID,
		Answer:   answer,
	})

	return err
}
