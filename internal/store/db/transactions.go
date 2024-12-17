package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateRoomArgs struct {
	Player     AddPlayerParams
	Room       AddRoomParams
	RoomPlayer AddRoomPlayerParams
}

func (s DB) CreateRoom(ctx context.Context, arg CreateRoomArgs) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	_, err = s.WithTx(tx).AddPlayer(ctx, arg.Player)
	if err != nil {
		return err
	}

	_, err = s.WithTx(tx).AddRoom(ctx, arg.Room)
	if err != nil {
		return err
	}

	_, err = s.WithTx(tx).AddRoomPlayer(ctx, arg.RoomPlayer)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

type AddPlayerToRoomArgs struct {
	Player     AddPlayerParams
	RoomPlayer AddRoomPlayerParams
}

func (s DB) AddPlayerToRoom(ctx context.Context, arg AddPlayerToRoomArgs) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	_, err = s.AddPlayer(ctx, arg.Player)
	if err != nil {
		return err
	}

	_, err = s.AddRoomPlayer(ctx, arg.RoomPlayer)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

type StartGameArgs struct {
	RoomID            uuid.UUID
	GameStateID       uuid.UUID
	NormalsQuestionID uuid.UUID
	FibberQuestionID  uuid.UUID
	Players           []GetAllPlayersInRoomRow
	FibberLoc         int
	Deadline          time.Time
}

func (s DB) StartGame(ctx context.Context, arg StartGameArgs) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)
	_, err = s.WithTx(tx).UpdateRoomState(ctx, UpdateRoomStateParams{
		RoomState: ROOMSTATE_PLAYING.String(),
		ID:        arg.RoomID,
	})
	if err != nil {
		return err
	}

	_, err = s.WithTx(tx).AddGameState(ctx, AddGameStateParams{
		ID:             arg.GameStateID,
		RoomID:         arg.RoomID,
		State:          GAMESTATE_FIBBING_IT_QUESTION.String(),
		SubmitDeadline: pgtype.Timestamp{Time: arg.Deadline, Valid: true},
	})
	if err != nil {
		return err
	}

	round, err := s.WithTx(tx).AddFibbingItRound(ctx, AddFibbingItRoundParams{
		ID:               uuid.Must(uuid.NewV7()),
		RoundType:        "free_form",
		Round:            1,
		FibberQuestionID: arg.FibberQuestionID,
		NormalQuestionID: arg.NormalsQuestionID,
		GameStateID:      arg.GameStateID,
	})
	if err != nil {
		return err
	}

	for i, player := range arg.Players {
		role := "normal"
		if i == arg.FibberLoc {
			role = "fibber"
		}

		_, err = s.WithTx(tx).AddFibbingItRole(ctx, AddFibbingItRoleParams{
			ID:         uuid.Must(uuid.NewV7()),
			RoundID:    round.ID,
			PlayerID:   player.ID,
			PlayerRole: role,
		})
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

type NewRoundArgs struct {
	GameStateID       uuid.UUID
	NormalsQuestionID uuid.UUID
	FibberQuestionID  uuid.UUID
	RoundType         string
	Round             int32
	PlayerIDs         []uuid.UUID
	FibberLoc         int
}

func (s DB) NewRound(ctx context.Context, arg NewRoundArgs) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	newRound, err := s.AddFibbingItRound(ctx, AddFibbingItRoundParams{
		ID:               uuid.Must(uuid.NewV7()),
		RoundType:        arg.RoundType,
		Round:            int32(arg.Round),
		FibberQuestionID: arg.FibberQuestionID,
		NormalQuestionID: arg.NormalsQuestionID,
		GameStateID:      arg.GameStateID,
	})
	if err != nil {
		return err
	}

	for i, id := range arg.PlayerIDs {
		role := "normal"
		if i == arg.FibberLoc {
			role = "fibber"
		}

		_, err = s.AddFibbingItRole(ctx, AddFibbingItRoleParams{
			ID:         uuid.Must(uuid.NewV7()),
			RoundID:    newRound.ID,
			PlayerID:   id,
			PlayerRole: role,
		})
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
