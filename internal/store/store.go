package store

import (
	"context"
	"database/sql"
	"io/fs"
	"os"
	"path/filepath"

	"gitlab.com/hmajid2301/banterbus/internal/entities"
	sqlc "gitlab.com/hmajid2301/banterbus/internal/store/db"
)

type Store struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewStore(db *sql.DB) (Store, error) {
	queries := sqlc.New(db)
	store := Store{
		db:      db,
		queries: queries,
	}

	return store, nil
}

func (s Store) CreateRoom(ctx context.Context, player entities.NewPlayer, room entities.NewRoom) (err error) {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			err = tx.Rollback()
		}
	}()

	newPlayer, err := s.queries.WithTx(tx).AddPlayer(ctx, sqlc.AddPlayerParams{
		Avatar:          player.Avatar,
		Nickname:        player.Nickname,
		LatestSessionID: player.SessionID,
	})
	if err != nil {
		return err
	}

	newRoom, err := s.queries.WithTx(tx).AddRoom(ctx, sqlc.AddRoomParams{
		GameName:   room.GameName,
		RoomCode:   room.RoomCode,
		HostPlayer: newPlayer.ID,
	})
	if err != nil {
		return err
	}

	_, err = s.queries.WithTx(tx).AddRoomPlayer(ctx, sqlc.AddRoomPlayerParams{
		RoomID:   newRoom.ID,
		PlayerID: newPlayer.ID,
	})
	if err != nil {
		return err
	}
	return tx.Commit()
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
