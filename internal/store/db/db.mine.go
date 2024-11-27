package sqlc

import (
	"database/sql"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

type DB struct {
	*Queries
	db *sql.DB
}

func NewDB(db *sql.DB) (DB, error) {
	queries := New(db)
	store := DB{
		db:      db,
		Queries: queries,
	}

	return store, nil
}

type RoomState int

const (
	ROOMSTATE_CREATED RoomState = iota
	ROOMSTATE_PLAYING
	ROOMSTATE_PAUSED
	ROOMSTATE_FINISHED
	ROOMSTATE_ABANDONED
)

func (rs RoomState) String() string {
	return [...]string{"CREATED", "PLAYING", "PAUSED", "FINISHED", "ABANDONED"}[rs]
}

func RoomStateFromString(s string) (RoomState, error) {
	stringToRoomState := map[string]RoomState{
		"CREATED":   ROOMSTATE_CREATED,
		"PLAYING":   ROOMSTATE_PLAYING,
		"PAUSED":    ROOMSTATE_PAUSED,
		"FINISHED":  ROOMSTATE_FINISHED,
		"ABANDONED": ROOMSTATE_ABANDONED,
	}

	if rs, ok := stringToRoomState[s]; ok {
		return rs, nil
	}
	return 0, errors.New("invalid RoomState string")
}

type GameStateEnum int

const (
	GAMESTATE_FIBBING_IT_SHOW_QUESTION GameStateEnum = iota
)

func (gs GameStateEnum) String() string {
	return [...]string{"GAMESTATE_FIBBING_IT_SHOW_QUESTION"}[gs]
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
