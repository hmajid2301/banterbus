package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	sqlc "gitlab.com/hmajid2301/banterbus/internal/store/db"
)

func (s Store) SubmitAnswer(ctx context.Context, playerID string, answer string, submittedAt time.Time) (err error) {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			err = errors.Join(err, tx.Rollback())
		}
	}()

	room, err := s.queries.WithTx(tx).GetRoomByPlayerID(ctx, playerID)
	if err != nil {
		return err
	}

	if room.RoomState != PLAYING.String() {
		return fmt.Errorf("room is not in PLAYING state")
	}

	round, err := s.queries.WithTx(tx).GetLatestRoundByPlayerID(ctx, playerID)
	if err != nil {
		return err
	}

	if submittedAt.After(round.SubmitDeadline) {
		return fmt.Errorf("answer submission deadline has passed")
	}

	_, err = s.queries.WithTx(tx).AddFibbingItAnswer(ctx, sqlc.AddFibbingItAnswerParams{
		ID:       uuid.Must(uuid.NewV7()).String(),
		RoundID:  round.ID,
		PlayerID: playerID,
		Answer:   answer,
	})
	if err != nil {
		return err
	}

	return tx.Commit()
}
