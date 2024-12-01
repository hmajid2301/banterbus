package sqlc

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type CreateRoomParams struct {
	Player     AddPlayerParams
	Room       AddRoomParams
	RoomPlayer AddRoomPlayerParams
}

func (s DB) CreateRoom(ctx context.Context, arg CreateRoomParams) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

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

	return tx.Commit()
}

type AddPlayerToRoomArgs struct {
	Player     AddPlayerParams
	RoomPlayer AddRoomPlayerParams
}

func (s DB) AddPlayerToRoom(ctx context.Context, arg AddPlayerToRoomArgs) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	_, err = s.AddPlayer(ctx, arg.Player)
	if err != nil {
		return err
	}

	_, err = s.AddRoomPlayer(ctx, arg.RoomPlayer)
	if err != nil {
		return err
	}

	return tx.Commit()
}

type StartGameArgs struct {
	RoomID            string
	GameStateID       string
	NormalsQuestionID string
	FibberQuestionID  string
	Players           []GetAllPlayersInRoomRow
	FibberLoc         int
	Deadline          time.Time
}

func (s DB) StartGame(ctx context.Context, arg StartGameArgs) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()
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
		State:          GAMESTATE_FIBBING_IT_SHOW_QUESTION.String(),
		SubmitDeadline: arg.Deadline,
	})
	if err != nil {
		return err
	}

	round, err := s.WithTx(tx).AddFibbingItRound(ctx, AddFibbingItRoundParams{
		ID:               uuid.Must(uuid.NewV7()).String(),
		RoundType:        "free_form",
		Round:            1,
		FibberQuestionID: arg.FibberQuestionID,
		NormalQuestionID: arg.NormalsQuestionID,
		RoomID:           arg.RoomID,
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
			ID:         uuid.Must(uuid.NewV7()).String(),
			RoundID:    round.ID,
			PlayerID:   player.ID,
			PlayerRole: role,
		})
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
