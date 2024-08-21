package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"math/rand/v2"
	"os"
	"path/filepath"

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
			rbErr := tx.Rollback()
			if rbErr != nil {
				err = fmt.Errorf(
					"failed to rollback: %w; while handling this error: %w",
					rbErr,
					err,
				)
			}
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
			rbErr := tx.Rollback()
			if rbErr != nil {
				err = fmt.Errorf(
					"failed to rollback: %w; while handling this error: %w",
					rbErr,
					err,
				)
			}
		}
	}()

	room, err := s.queries.WithTx(tx).GetRoomByCode(ctx, roomCode)
	if err != nil {
		return players, err
	}

	if room.RoomState != CREATED.String() {
		return players, fmt.Errorf("room is not in CREATED state")
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

func GetDB(dbFolder string) (*sql.DB, error) {
	if _, err := os.Stat(dbFolder); os.IsNotExist(err) {
		permissions := 0755
		err = os.Mkdir(dbFolder, fs.FileMode(permissions))
		if err != nil {
			return nil, err
		}
	}

	dbPath := filepath.Join(dbFolder, "banterbus.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("PRAGMA journal_mode=WAL")
	return db, err
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
