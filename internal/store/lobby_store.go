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

	if room.HostPlayer != playerID {
		return players, fmt.Errorf("player is not the host of the room")
	}

	if room.RoomState != CREATED.String() {
		return players, fmt.Errorf("room is not in CREATED state")
	}

	players, err = s.queries.WithTx(tx).GetAllPlayersInRoom(ctx, playerID)
	if err != nil {
		return players, err
	}

	if len(players) < 2 {
		return players, fmt.Errorf("not enough players to start the game")
	}

	for _, player := range players {
		if !player.IsReady.Bool {
			return players, fmt.Errorf("not all players are ready: %s", player.ID)
		}
	}

	_, err = s.queries.WithTx(tx).UpdateRoomState(ctx, sqlc.UpdateRoomStateParams{
		RoomState: PLAYING.String(),
		ID:        room.ID,
	})
	if err != nil {
		return players, err
	}

	return players, tx.Commit()
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
