package store_test

import (
	"context"
	"database/sql"
	"testing"

	"gitlab.com/hmajid2301/banterbus/internal/banterbustest"
	"gitlab.com/hmajid2301/banterbus/internal/entities"
	"gitlab.com/hmajid2301/banterbus/internal/store"
)

func setupSubtest(t *testing.T) (*sql.DB, func()) {
	ctx := context.Background()
	db := banterbustest.CreateDB(ctx, t)

	return db, func() {
		db.Close()
	}
}

func createRoom(ctx context.Context, myStore store.Store) (string, error) {
	newPlayer := entities.NewPlayer{
		Nickname: "Majiy00",
		Avatar:   []byte(""),
	}

	newRoom := entities.NewRoom{
		GameName: "fibbing_it",
	}

	roomCode, err := myStore.CreateRoom(ctx, newPlayer, newRoom)
	return roomCode, err
}
