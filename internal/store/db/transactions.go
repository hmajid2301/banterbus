package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateRoomArgs struct {
	Player     AddPlayerParams
	Room       AddRoomParams
	RoomPlayer AddRoomPlayerParams
}

func (s *DB) CreateRoom(ctx context.Context, arg CreateRoomArgs) error {
	return s.TransactionWithRetry(ctx, func(q *Queries) error {
		_, err := q.AddPlayer(ctx, arg.Player)
		if err != nil {
			return err
		}

		_, err = q.AddRoom(ctx, arg.Room)
		if err != nil {
			return err
		}

		_, err = q.AddRoomPlayer(ctx, arg.RoomPlayer)
		if err != nil {
			return err
		}

		return nil
	})
}

type AddPlayerToRoomArgs struct {
	Player     AddPlayerParams
	RoomPlayer AddRoomPlayerParams
}

func (s *DB) AddPlayerToRoom(ctx context.Context, arg AddPlayerToRoomArgs) error {
	return s.TransactionWithRetry(ctx, func(q *Queries) error {
		_, err := q.AddPlayer(ctx, arg.Player)
		if err != nil {
			return err
		}

		_, err = q.AddRoomPlayer(ctx, arg.RoomPlayer)
		return err
	})
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

func (s *DB) StartGame(ctx context.Context, arg StartGameArgs) error {
	return s.TransactionWithRetry(ctx, func(q *Queries) error {
		_, err := q.UpdateRoomState(ctx, UpdateRoomStateParams{
			RoomState: Playing.String(),
			ID:        arg.RoomID,
		})
		if err != nil {
			return err
		}

		_, err = q.AddGameState(ctx, AddGameStateParams{
			ID:             arg.GameStateID,
			RoomID:         arg.RoomID,
			State:          FibbingITQuestion.String(),
			SubmitDeadline: pgtype.Timestamp{Time: arg.Deadline, Valid: true},
		})
		if err != nil {
			return err
		}

		round, err := q.AddFibbingItRound(ctx, AddFibbingItRoundParams{
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

			_, err = q.AddFibbingItRole(ctx, AddFibbingItRoleParams{
				ID:         uuid.Must(uuid.NewV7()),
				RoundID:    round.ID,
				PlayerID:   player.ID,
				PlayerRole: role,
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}

type NewRoundArgs struct {
	GameStateID       uuid.UUID
	NormalsQuestionID uuid.UUID
	FibberQuestionID  uuid.UUID
	RoundType         string
	Round             int32
	Players           []GetAllPlayersByGameStateIDRow
	FibberLoc         int
}

func (s *DB) NewRound(ctx context.Context, arg NewRoundArgs) error {
	return s.TransactionWithRetry(ctx, func(q *Queries) error {
		newRound, err := q.AddFibbingItRound(ctx, AddFibbingItRoundParams{
			ID:               uuid.Must(uuid.NewV7()),
			RoundType:        arg.RoundType,
			Round:            arg.Round,
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

			_, err = q.AddFibbingItRole(ctx, AddFibbingItRoleParams{
				ID:         uuid.Must(uuid.NewV7()),
				RoundID:    newRound.ID,
				PlayerID:   player.ID,
				PlayerRole: role,
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}

type NewScoresArgs struct {
	Players []AddFibbingItScoreParams
}

func (s *DB) NewScores(ctx context.Context, arg NewScoresArgs) error {
	return s.TransactionWithRetry(ctx, func(q *Queries) error {
		for _, player := range arg.Players {
			_, err := q.AddFibbingItScore(ctx, AddFibbingItScoreParams{
				ID:       uuid.Must(uuid.NewV7()),
				PlayerID: player.PlayerID,
				RoundID:  player.RoundID,
				Score:    player.Score,
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}

type CreateQuestionArgs struct {
	GameName  string
	GroupName string
	RoundType string
	Text      string
	Locale    string
}

func (s *DB) CreateQuestion(ctx context.Context, arg CreateQuestionArgs) (uuid.UUID, error) {
	var questionID uuid.UUID

	err := s.TransactionWithRetry(ctx, func(q *Queries) error {
		questionGroup, err := q.GetGroupByName(ctx, arg.GroupName)
		if err != nil {
			return err
		}

		newQuestion, err := q.AddQuestion(ctx, AddQuestionParams{
			ID:        uuid.Must(uuid.NewV7()),
			GameName:  arg.GameName,
			RoundType: arg.RoundType,
			GroupID:   questionGroup.ID,
		})
		if err != nil {
			return err
		}

		_, err = q.AddQuestionTranslation(ctx, AddQuestionTranslationParams{
			ID:         uuid.Must(uuid.NewV7()),
			Question:   arg.Text,
			QuestionID: newQuestion.ID,
			Locale:     arg.Locale,
		})
		if err != nil {
			return err
		}

		questionID = newQuestion.ID
		return nil
	})

	return questionID, err
}
