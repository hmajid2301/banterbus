package sqlc

import (
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	*Queries
	db *pgxpool.Pool
}

func NewDB(db *pgxpool.Pool) (DB, error) {
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
	GAMESTATE_FIBBING_IT_VOTING
)

func GameStateFromString(s string) (GameStateEnum, error) {
	stringToGameState := map[string]GameStateEnum{
		"FIBBING_IT_SHOW_QUESTION": GAMESTATE_FIBBING_IT_SHOW_QUESTION,
		"FIBBING_IT_VOTING":        GAMESTATE_FIBBING_IT_VOTING,
	}

	if rs, ok := stringToGameState[s]; ok {
		return rs, nil
	}
	return 0, errors.New("invalid GameState string")
}

func (gs GameStateEnum) String() string {
	return [...]string{"FIBBING_IT_SHOW_QUESTION", "FIBBING_IT_VOTING"}[gs]
}
