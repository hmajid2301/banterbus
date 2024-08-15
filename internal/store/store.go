package store

import (
	"context"
	"database/sql"
	"io/fs"
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
		ID:       player.ID,
		Avatar:   player.Avatar,
		Nickname: player.Nickname,
	})
	if err != nil {
		return err
	}

	u := uuid.Must(uuid.NewV7())
	newRoom, err := s.queries.WithTx(tx).AddRoom(ctx, sqlc.AddRoomParams{
		ID:         u.String(),
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

func (s Store) AddPlayerToRoom(ctx context.Context, player entities.NewPlayer, roomCode string) (players []sqlc.GetAllPlayersInRoomRow, err error) {
	tx, err := s.db.Begin()
	if err != nil {
		return players, err
	}

	defer func() {
		if err != nil {
			err = tx.Rollback()
		}
	}()

	newPlayer, err := s.queries.WithTx(tx).AddPlayer(ctx, sqlc.AddPlayerParams{
		ID:       player.ID,
		Avatar:   player.Avatar,
		Nickname: player.Nickname,
	})
	if err != nil {
		return players, err
	}

	roomID, err := s.queries.WithTx(tx).GetRoomByCode(ctx, roomCode)
	if err != nil {
		return players, err
	}

	_, err = s.queries.WithTx(tx).AddRoomPlayer(ctx, sqlc.AddRoomPlayerParams{
		RoomID:   roomID,
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

func (s Store) UpdateNickname(ctx context.Context, nickname string, playerID string) (players []sqlc.GetAllPlayersInRoomRow, err error) {
	tx, err := s.db.Begin()
	if err != nil {
		return players, err
	}

	defer func() {
		if err != nil {
			err = tx.Rollback()
		}
	}()

	_, err = s.queries.WithTx(tx).UpdateNickname(ctx, sqlc.UpdateNicknameParams{
		Nickname: nickname,
		ID:       playerID,
	})
	if err != nil {
		return players, err
	}

	players, err = s.queries.WithTx(tx).GetAllPlayersInRoom(ctx, playerID)
	if err != nil {
		return players, err
	}

	return players, tx.Commit()
}

func (s Store) UpdateAvatar(ctx context.Context, avatar []byte, playerID string) (players []sqlc.GetAllPlayersInRoomRow, err error) {
	tx, err := s.db.Begin()
	if err != nil {
		return players, err
	}

	defer func() {
		if err != nil {
			err = tx.Rollback()
		}
	}()

	_, err = s.queries.WithTx(tx).UpdateAvatar(ctx, sqlc.UpdateAvatarParams{
		Avatar: avatar,
		ID:     playerID,
	})
	if err != nil {
		return players, err
	}

	players, err = s.queries.WithTx(tx).GetAllPlayersInRoom(ctx, playerID)
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
