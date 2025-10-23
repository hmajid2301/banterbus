package db

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/invopop/ctxi18n"
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

		roundID, err := uuid.NewV7()
		if err != nil {
			return err
		}
		round, err := q.AddFibbingItRound(ctx, AddFibbingItRoundParams{
			ID:               roundID,
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

			roleID, err := uuid.NewV7()
			if err != nil {
				return err
			}
			_, err = q.AddFibbingItRole(ctx, AddFibbingItRoleParams{
				ID:         roleID,
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
		roundID, err := uuid.NewV7()
		if err != nil {
			return err
		}
		newRound, err := q.AddFibbingItRound(ctx, AddFibbingItRoundParams{
			ID:               roundID,
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

			roleID, err := uuid.NewV7()
			if err != nil {
				return err
			}
			_, err = q.AddFibbingItRole(ctx, AddFibbingItRoleParams{
				ID:         roleID,
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

func (s *DB) AddScores(ctx context.Context, arg NewScoresArgs) error {
	return s.TransactionWithRetry(ctx, func(q *Queries) error {
		for _, player := range arg.Players {
			scoreID, err := uuid.NewV7()
			if err != nil {
				return err
			}
			_, err = q.AddFibbingItScore(ctx, AddFibbingItScoreParams{
				ID:       scoreID,
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

func (s *DB) CreateQuestionWithTranslation(ctx context.Context, arg CreateQuestionArgs) (uuid.UUID, error) {
	var questionID uuid.UUID

	err := s.TransactionWithRetry(ctx, func(q *Queries) error {
		questionGroup, err := q.GetGroupByName(ctx, arg.GroupName)
		if err != nil {
			return err
		}

		qID, err := uuid.NewV7()
		if err != nil {
			return err
		}
		newQuestion, err := q.AddQuestion(ctx, AddQuestionParams{
			ID:        qID,
			GameName:  arg.GameName,
			RoundType: arg.RoundType,
			GroupID:   questionGroup.ID,
		})
		if err != nil {
			return err
		}

		translationID, err := uuid.NewV7()
		if err != nil {
			return err
		}
		_, err = q.AddQuestionTranslation(ctx, AddQuestionTranslationParams{
			ID:         translationID,
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

type UpdateNicknameArgs struct {
	PlayerID uuid.UUID
	Nickname string
}

type UpdateNicknameResult struct {
	Players []GetAllPlayersInRoomRow
}

func (s *DB) UpdateNicknameWithPlayers(ctx context.Context, arg UpdateNicknameArgs) (UpdateNicknameResult, error) {
	var result UpdateNicknameResult

	err := s.TransactionWithRetry(ctx, func(q *Queries) error {
		room, err := q.GetRoomByPlayerIDForUpdate(ctx, arg.PlayerID)
		if err != nil {
			if IsLockConflict(err) {
				return errors.New("room is currently being modified, please try again")
			}
			return err
		}

		if room.RoomState != Created.String() {
			return errors.New("room is not in CREATED state")
		}

		playersInRoom, err := q.GetAllPlayersInRoom(ctx, arg.PlayerID)
		if err != nil {
			return err
		}

		for _, player := range playersInRoom {
			if player.Nickname == arg.Nickname {
				return errors.New("nickname already exists")
			}
		}

		_, err = q.UpdateNickname(ctx, UpdateNicknameParams{
			Nickname: arg.Nickname,
			ID:       arg.PlayerID,
		})
		if err != nil {
			return err
		}

		players, err := q.GetAllPlayersInRoom(ctx, arg.PlayerID)
		if err != nil {
			return err
		}

		if len(players) == 0 {
			return errors.New("no players found in room")
		}

		result.Players = players
		return nil
	})

	return result, err
}

type UpdateStateToQuestionArgs struct {
	GameStateID       uuid.UUID
	Deadline          time.Time
	NextRound         bool
	NormalsQuestionID uuid.UUID
	FibberQuestionID  uuid.UUID
	RoundType         string
	RoundNumber       int32
	Players           []GetAllPlayersByGameStateIDRow
	FibberLoc         int
}

type UpdateStateToQuestionResult struct {
	RoundID     uuid.UUID
	RoundType   string
	RoundNumber int32
	Players     []GetAllPlayersByGameStateIDRow
}

func (s *DB) UpdateStateToQuestion(ctx context.Context, arg UpdateStateToQuestionArgs) (UpdateStateToQuestionResult, error) {
	var result UpdateStateToQuestionResult

	err := s.TransactionWithRetry(ctx, func(q *Queries) error {
		game, err := q.GetGameStateForUpdate(ctx, arg.GameStateID)
		if err != nil {
			if IsLockConflict(err) {
				currentGame, err := q.GetGameState(ctx, arg.GameStateID)
				if err != nil {
					return err
				}
				currentState, err := ParseFibbingItGameState(currentGame.State)
				if err != nil {
					return err
				}
				if currentState == FibbingITQuestion {
					round, err := q.GetLatestRoundByGameStateID(ctx, arg.GameStateID)
					if err != nil {
						return err
					}
					players, err := q.GetAllPlayersByGameStateID(ctx, arg.GameStateID)
					if err != nil {
						return err
					}
					result.RoundID = round.ID
					result.RoundType = round.RoundType
					result.RoundNumber = round.Round
					result.Players = players
					return nil
				}
				return errors.New("game state is already in FIBBING_IT_QUESTION state")
			}
			return err
		}

		gameState, err := ParseFibbingItGameState(game.State)
		if err != nil {
			return err
		}

		if gameState == FibbingITQuestion {
			round, err := q.GetLatestRoundByGameStateID(ctx, arg.GameStateID)
			if err != nil {
				return err
			}
			players, err := q.GetAllPlayersByGameStateID(ctx, arg.GameStateID)
			if err != nil {
				return err
			}
			result.RoundID = round.ID
			result.RoundType = round.RoundType
			result.RoundNumber = round.Round
			result.Players = players
			return nil
		}

		if gameState != FibbingItReveal && gameState != FibbingItScoring {
			return errors.New("game state is not in FIBBING_IT_REVEAL state or FIBBING_IT_SCORING state")
		}

		_, err = q.UpdateGameState(ctx, UpdateGameStateParams{
			ID:             arg.GameStateID,
			SubmitDeadline: pgtype.Timestamp{Time: arg.Deadline, Valid: true},
			State:          FibbingITQuestion.String(),
		})
		if err != nil {
			return err
		}

		roundID, err := uuid.NewV7()
		if err != nil {
			return err
		}
		_, err = q.AddFibbingItRound(ctx, AddFibbingItRoundParams{
			ID:               roundID,
			RoundType:        arg.RoundType,
			Round:            arg.RoundNumber,
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

			roleID, err := uuid.NewV7()
			if err != nil {
				return err
			}
			_, err = q.AddFibbingItRole(ctx, AddFibbingItRoleParams{
				ID:         roleID,
				RoundID:    roundID,
				PlayerID:   player.ID,
				PlayerRole: role,
			})
			if err != nil {
				return err
			}
		}

		result.RoundID = roundID
		result.RoundType = arg.RoundType
		result.RoundNumber = arg.RoundNumber
		result.Players = arg.Players
		return nil
	})

	return result, err
}

type GenerateNewAvatarArgs struct {
	PlayerID uuid.UUID
	Avatar   string
}

type GenerateNewAvatarResult struct {
	Players []GetAllPlayersInRoomRow
}

func (s *DB) GenerateNewAvatarWithPlayers(ctx context.Context, arg GenerateNewAvatarArgs) (GenerateNewAvatarResult, error) {
	var result GenerateNewAvatarResult

	err := s.TransactionWithRetry(ctx, func(q *Queries) error {
		room, err := q.GetRoomByPlayerIDForUpdate(ctx, arg.PlayerID)
		if err != nil {
			if IsLockConflict(err) {
				return errors.New("room is currently being modified, please try again")
			}
			return err
		}

		if room.RoomState != Created.String() {
			return errors.New("room is not in CREATED state")
		}

		_, err = q.UpdateAvatar(ctx, UpdateAvatarParams{
			Avatar: arg.Avatar,
			ID:     arg.PlayerID,
		})
		if err != nil {
			return err
		}

		players, err := q.GetAllPlayersInRoom(ctx, arg.PlayerID)
		if err != nil {
			return err
		}

		if len(players) == 0 {
			return errors.New("no players found in room")
		}

		result.Players = players
		return nil
	})

	return result, err
}

type TogglePlayerIsReadyArgs struct {
	PlayerID uuid.UUID
}

type TogglePlayerIsReadyResult struct {
	Players []GetAllPlayersInRoomRow
}

func (s *DB) TogglePlayerReadyWithPlayers(ctx context.Context, arg TogglePlayerIsReadyArgs) (TogglePlayerIsReadyResult, error) {
	var result TogglePlayerIsReadyResult

	err := s.TransactionWithRetry(ctx, func(q *Queries) error {
		room, err := q.GetRoomByPlayerIDForUpdate(ctx, arg.PlayerID)
		if err != nil {
			if IsLockConflict(err) {
				return errors.New("room is currently being modified, please try again")
			}
			return err
		}

		if room.RoomState != Created.String() {
			return errors.New("room is not in CREATED state")
		}

		_, err = q.TogglePlayerIsReady(ctx, arg.PlayerID)
		if err != nil {
			return err
		}

		players, err := q.GetAllPlayersInRoom(ctx, arg.PlayerID)
		if err != nil {
			return err
		}

		if len(players) == 0 {
			return errors.New("no players found in room")
		}

		result.Players = players
		return nil
	})

	return result, err
}

type JoinRoomArgs struct {
	RoomCode string
	PlayerID uuid.UUID
	Nickname string
	Avatar   string
}

type JoinRoomResult struct {
	Players []GetAllPlayersInRoomRow
}

func (s *DB) JoinRoom(ctx context.Context, arg JoinRoomArgs) (JoinRoomResult, error) {
	var result JoinRoomResult

	err := s.TransactionWithRepeatableRead(ctx, func(q *Queries) error {
		room, err := q.GetRoomByCode(ctx, arg.RoomCode)
		if err != nil {
			return err
		}

		if room.RoomState != Created.String() {
			return errors.New("room is not in CREATED state")
		}

		playersInRoom, err := q.GetAllPlayerByRoomCode(ctx, arg.RoomCode)
		if err != nil {
			return err
		}

		for _, p := range playersInRoom {
			if p.Nickname == arg.Nickname {
				return errors.New("nickname already exists")
			}
		}

		locale := ctxi18n.Locale(ctx).Code().String()
		_, err = q.AddPlayer(ctx, AddPlayerParams{
			ID:       arg.PlayerID,
			Avatar:   arg.Avatar,
			Nickname: arg.Nickname,
			Locale:   pgtype.Text{String: locale},
		})
		if err != nil {
			return err
		}

		_, err = q.AddRoomPlayer(ctx, AddRoomPlayerParams{
			RoomID:   room.ID,
			PlayerID: arg.PlayerID,
		})
		if err != nil {
			return err
		}

		players, err := q.GetAllPlayersInRoom(ctx, arg.PlayerID)
		if err != nil {
			return err
		}

		result.Players = players
		return nil
	})

	return result, err
}

type UpdateStateToScoreArgs struct {
	GameStateID uuid.UUID
	Deadline    time.Time
	Scores      []AddFibbingItScoreParams
}

type UpdateStateToScoreResult struct {
	Success bool
}

func (s *DB) UpdateStateToScore(ctx context.Context, arg UpdateStateToScoreArgs) (UpdateStateToScoreResult, error) {
	var result UpdateStateToScoreResult

	err := s.TransactionWithRetry(ctx, func(q *Queries) error {
		game, err := q.GetGameStateForUpdate(ctx, arg.GameStateID)
		if err != nil {
			if IsLockConflict(err) {
				currentGame, err := q.GetGameState(ctx, arg.GameStateID)
				if err != nil {
					return err
				}
				currentState, err := ParseFibbingItGameState(currentGame.State)
				if err != nil {
					return err
				}
				if currentState == FibbingItScoring {
					result.Success = true
					return nil
				}
				return errors.New("game state is not in FIBBING_IT_REVEAL state")
			}
			return err
		}

		gameState, err := ParseFibbingItGameState(game.State)
		if err != nil {
			return err
		}

		if gameState == FibbingItScoring {
			result.Success = true
			return nil
		}

		if gameState != FibbingItReveal {
			return errors.New("game state is not in FIBBING_IT_REVEAL state")
		}

		_, err = q.UpdateGameStateIfInState(ctx, UpdateGameStateIfInStateParams{
			State:          FibbingItScoring.String(),
			SubmitDeadline: pgtype.Timestamp{Time: arg.Deadline, Valid: true},
			ID:             arg.GameStateID,
			State_2:        FibbingItReveal.String(),
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errors.New("game state is not in FIBBING_IT_REVEAL state")
			}
			return err
		}

		for _, score := range arg.Scores {
			_, err = q.AddFibbingItScore(ctx, score)
			if err != nil {
				return err
			}
		}

		result.Success = true
		return nil
	})

	return result, err
}

func (s *DB) ExecuteTransactionWithRetry(ctx context.Context, fn func(*Queries) error) error {
	return s.TransactionWithRetry(ctx, fn)
}

type UpdateStateToVotingArgs struct {
	GameStateID uuid.UUID
	Deadline    time.Time
}

type UpdateStateToVotingResult struct {
	Round   int32
	RoundID uuid.UUID
}

func (s *DB) UpdateStateToVoting(ctx context.Context, arg UpdateStateToVotingArgs) (UpdateStateToVotingResult, error) {
	var result UpdateStateToVotingResult

	err := s.TransactionWithRetry(ctx, func(q *Queries) error {
		game, err := q.GetGameStateForUpdate(ctx, arg.GameStateID)
		if err != nil {
			if IsLockConflict(err) {
				currentGame, err := q.GetGameState(ctx, arg.GameStateID)
				if err != nil {
					return err
				}
				currentState, err := ParseFibbingItGameState(currentGame.State)
				if err != nil {
					return err
				}
				if currentState == FibbingItVoting {
					round, err := q.GetLatestRoundByGameStateID(ctx, arg.GameStateID)
					if err != nil {
						return err
					}
					result.Round = round.Round
					result.RoundID = round.ID
					return nil
				}
				return errors.New("game state is not in FIBBING_IT_QUESTION state")
			}
			return err
		}

		gameState, err := ParseFibbingItGameState(game.State)
		if err != nil {
			return err
		}

		if gameState == FibbingItVoting {
			round, err := q.GetLatestRoundByGameStateID(ctx, arg.GameStateID)
			if err != nil {
				return err
			}
			result.Round = round.Round
			result.RoundID = round.ID
			return nil
		}

		if gameState != FibbingITQuestion {
			return errors.New("game state is not in FIBBING_IT_QUESTION state")
		}

		_, err = q.UpdateGameStateIfInState(ctx, UpdateGameStateIfInStateParams{
			State:          FibbingItVoting.String(),
			SubmitDeadline: pgtype.Timestamp{Time: arg.Deadline, Valid: true},
			ID:             arg.GameStateID,
			State_2:        FibbingITQuestion.String(),
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errors.New("game state is not in FIBBING_IT_QUESTION state")
			}
			return err
		}

		round, err := q.GetLatestRoundByGameStateID(ctx, arg.GameStateID)
		if err != nil {
			return err
		}

		result.Round = round.Round
		result.RoundID = round.ID
		return nil
	})

	return result, err
}

type UpdateStateToRevealArgs struct {
	GameStateID uuid.UUID
	Deadline    time.Time
}

type UpdateStateToRevealResult struct {
	Round   int32
	RoundID uuid.UUID
}

func (s *DB) UpdateStateToReveal(ctx context.Context, arg UpdateStateToRevealArgs) (UpdateStateToRevealResult, error) {
	var result UpdateStateToRevealResult

	err := s.TransactionWithRetry(ctx, func(q *Queries) error {
		game, err := q.GetGameStateForUpdate(ctx, arg.GameStateID)
		if err != nil {
			if IsLockConflict(err) {
				currentGame, err := q.GetGameState(ctx, arg.GameStateID)
				if err != nil {
					return err
				}
				currentState, err := ParseFibbingItGameState(currentGame.State)
				if err != nil {
					return err
				}
				if currentState == FibbingItReveal {
					round, err := q.GetLatestRoundByGameStateID(ctx, arg.GameStateID)
					if err != nil {
						return err
					}
					result.Round = round.Round
					result.RoundID = round.ID
					return nil
				}
				return errors.New("game state is not in FIBBING_IT_VOTING state")
			}
			return err
		}

		gameState, err := ParseFibbingItGameState(game.State)
		if err != nil {
			return err
		}

		if gameState == FibbingItReveal {
			round, err := q.GetLatestRoundByGameStateID(ctx, arg.GameStateID)
			if err != nil {
				return err
			}
			result.Round = round.Round
			result.RoundID = round.ID
			return nil
		}

		if gameState != FibbingItVoting {
			return errors.New("game state is not in FIBBING_IT_VOTING state")
		}

		_, err = q.UpdateGameStateIfInState(ctx, UpdateGameStateIfInStateParams{
			State:          FibbingItReveal.String(),
			SubmitDeadline: pgtype.Timestamp{Time: arg.Deadline, Valid: true},
			ID:             arg.GameStateID,
			State_2:        FibbingItVoting.String(),
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errors.New("game state is not in FIBBING_IT_VOTING state")
			}
			return err
		}

		round, err := q.GetLatestRoundByGameStateID(ctx, arg.GameStateID)
		if err != nil {
			return err
		}

		result.Round = round.Round
		result.RoundID = round.ID
		return nil
	})

	return result, err
}
