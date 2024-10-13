package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand/v2"

	"github.com/google/uuid"

	"gitlab.com/hmajid2301/banterbus/internal/entities"
	sqlc "gitlab.com/hmajid2301/banterbus/internal/store/db"
)

type Store struct {
	db      *sql.DB
	queries *sqlc.Queries
}

type RoomState int

const (
	CREATED RoomState = iota
	PLAYING
	PAUSED
	FINISHED
	ABANDONED
)

func (rs RoomState) String() string {
	return [...]string{"CREATED", "PLAYING", "PAUSED", "FINISHED", "ABANDONED"}[rs]
}

func NewStore(db *sql.DB) (Store, error) {
	queries := sqlc.New(db)
	store := Store{
		db:      db,
		queries: queries,
	}

	return store, nil
}

func (s Store) CreateRoom(
	ctx context.Context,
	player entities.NewPlayer,
	room entities.NewRoom,
) (roomCode string, err error) {
	tx, err := s.db.Begin()
	if err != nil {
		return roomCode, err
	}

	defer func() {
		if err != nil {
			err = errors.Join(err, tx.Rollback())
		}
	}()

	for {
		roomCode = randomRoomCode()
		room, err := s.queries.WithTx(tx).GetRoomByCode(ctx, roomCode)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				break
			}

			return roomCode, err
		}

		if room.RoomState == FINISHED.String() || room.RoomState == ABANDONED.String() {
			break
		}
	}

	newPlayer, err := s.queries.WithTx(tx).AddPlayer(ctx, sqlc.AddPlayerParams{
		ID:       player.ID,
		Avatar:   player.Avatar,
		Nickname: player.Nickname,
	})
	if err != nil {
		return roomCode, err
	}

	u := uuid.Must(uuid.NewV7())
	newRoom, err := s.queries.WithTx(tx).AddRoom(ctx, sqlc.AddRoomParams{
		ID:         u.String(),
		GameName:   room.GameName,
		RoomCode:   roomCode,
		RoomState:  CREATED.String(),
		HostPlayer: newPlayer.ID,
	})
	if err != nil {
		return roomCode, err
	}

	_, err = s.queries.WithTx(tx).AddRoomPlayer(ctx, sqlc.AddRoomPlayerParams{
		RoomID:   newRoom.ID,
		PlayerID: newPlayer.ID,
	})
	if err != nil {
		return roomCode, err
	}
	return roomCode, tx.Commit()
}

func (s Store) AddPlayerToRoom(
	ctx context.Context,
	player entities.NewPlayer,
	roomCode string,
) (players []sqlc.GetAllPlayersInRoomRow, err error) {
	tx, err := s.db.Begin()
	if err != nil {
		return players, err
	}

	defer func() {
		if err != nil {
			err = errors.Join(err, tx.Rollback())
		}
	}()

	room, err := s.queries.WithTx(tx).GetRoomByCode(ctx, roomCode)
	if err != nil {
		return players, err
	}

	if room.RoomState != CREATED.String() {
		return players, fmt.Errorf("room is not in CREATED state")
	}

	playersInRoom, err := s.queries.WithTx(tx).GetAllPlayerByRoomCode(ctx, roomCode)
	if err != nil {
		return players, err
	}

	for _, p := range playersInRoom {
		if p.Nickname == player.Nickname {
			return players, entities.ErrNicknameExists
		}
	}

	newPlayer, err := s.queries.WithTx(tx).AddPlayer(ctx, sqlc.AddPlayerParams{
		ID:       player.ID,
		Avatar:   player.Avatar,
		Nickname: player.Nickname,
	})
	if err != nil {
		return players, err
	}

	_, err = s.queries.WithTx(tx).AddRoomPlayer(ctx, sqlc.AddRoomPlayerParams{
		RoomID:   room.ID,
		PlayerID: newPlayer.ID,
	})
	if err != nil {
		return players, err
	}

	players, err = s.queries.WithTx(tx).GetAllPlayersInRoom(ctx, player.ID)
	if err != nil {
		return players, err
	}

	return players, tx.Commit()
}

func (s Store) StartGame(
	ctx context.Context,
	roomCode string,
	playerID string,
) (gameState entities.GameState, err error) {
	tx, err := s.db.Begin()
	if err != nil {
		return gameState, err
	}

	defer func() {
		if err != nil {
			err = errors.Join(err, tx.Rollback())
		}
	}()

	room, err := s.queries.WithTx(tx).GetRoomByCode(ctx, roomCode)
	if err != nil {
		return gameState, err
	}

	if room.HostPlayer != playerID {
		return gameState, fmt.Errorf("player is not the host of the room")
	}

	if room.RoomState != CREATED.String() {
		return gameState, fmt.Errorf("room is not in CREATED state")
	}

	// TODO: rename the playersInRoom variable
	playersInRoom, err := s.queries.WithTx(tx).GetAllPlayersInRoom(ctx, playerID)
	if err != nil {
		return gameState, err
	}

	if len(playersInRoom) < 2 {
		return gameState, fmt.Errorf("not enough players to start the game")
	}

	for _, player := range playersInRoom {
		if !player.IsReady.Bool {
			return gameState, fmt.Errorf("not all players are ready: %s", player.ID)
		}
	}

	_, err = s.queries.WithTx(tx).UpdateRoomState(ctx, sqlc.UpdateRoomStateParams{
		RoomState: PLAYING.String(),
		ID:        room.ID,
	})
	if err != nil {
		return gameState, err
	}

	normalsQuestion, err := s.queries.WithTx(tx).GetRandomQuestionByRound(ctx, sqlc.GetRandomQuestionByRoundParams{
		GameName:     room.GameName,
		LanguageCode: "en-GB",
		// TODO: should we fetch a random round from the database?
		Round: "free_form",
	})
	if err != nil {
		return gameState, err
	}

	g, err := s.queries.WithTx(tx).AddGameState(ctx, sqlc.AddGameStateParams{
		ID:     uuid.Must(uuid.NewV7()).String(),
		RoomID: room.ID,
	})
	if err != nil {
		return gameState, err
	}

	fibberQuestion, err := s.queries.WithTx(tx).GetRandomQuestionInGroup(ctx, sqlc.GetRandomQuestionInGroupParams{
		GroupID: normalsQuestion.GroupID,
		ID:      normalsQuestion.ID,
	})
	if err != nil {
		return gameState, err
	}

	round, err := s.queries.WithTx(tx).AddFibbingItRound(ctx, sqlc.AddFibbingItRoundParams{
		ID:               uuid.Must(uuid.NewV7()).String(),
		RoundType:        "free_form",
		Round:            1,
		FibberQuestionID: fibberQuestion.ID,
		NormalQuestionID: normalsQuestion.ID,
		GameStateID:      g.ID,
	})
	if err != nil {
		return gameState, err
	}

	players := []entities.PlayerWithRole{}
	randomFibberLoc := rand.IntN(len(playersInRoom))
	for i, player := range playersInRoom {
		role := "normal"
		question := normalsQuestion.Question

		if i == randomFibberLoc {
			role = "fibber"
			question = fibberQuestion.Question
		}

		players = append(players, entities.PlayerWithRole{
			ID:       player.ID,
			Nickname: player.Nickname,
			Avatar:   player.Avatar,
			Role:     role,
			Question: question,
		})

		_, err = s.queries.WithTx(tx).AddFibbingItRole(ctx, sqlc.AddFibbingItRoleParams{
			ID:         uuid.Must(uuid.NewV7()).String(),
			RoundID:    round.ID,
			PlayerID:   player.ID,
			PlayerRole: role,
		})
		if err != nil {
			return gameState, err
		}

	}
	gameState = entities.GameState{
		Players:   players,
		Round:     1,
		RoundType: "free_form",
		RoomCode:  roomCode,
	}

	return gameState, tx.Commit()
}

func randomRoomCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	codeByte := make([]byte, 5)
	for i := range codeByte {
		codeByte[i] = charset[rand.IntN(len(charset))]
	}
	code := string(codeByte)
	return code
}
