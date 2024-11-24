package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	sqlc "gitlab.com/hmajid2301/banterbus/internal/store/db"
)

type RoundService struct {
	store Storer
}

func NewRoundService(store Storer) *RoundService {
	return &RoundService{store: store}
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
		ID:       uuid.Must(uuid.NewV7()).String(),
		RoundID:  round.ID,
		PlayerID: playerID,
		Answer:   answer,
	})

	return err
}
