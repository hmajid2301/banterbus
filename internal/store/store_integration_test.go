package store_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"

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

func TestIntegrationCreateRoom(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("Should create room to DB successfully", func(t *testing.T) {
		db, teardown := setupSubtest(t)
		defer teardown()

		myStore, err := store.NewStore(db)
		assert.NoError(t, err)

		ctx := context.Background()
		newPlayer := entities.NewPlayer{
			Nickname:  "Majiy00",
			Avatar:    []byte(""),
			SessionID: 1,
		}

		newRoom := entities.NewRoom{
			GameName: "fibbing_it",
			RoomCode: "1234",
		}

		err = myStore.CreateRoom(ctx, newPlayer, newRoom)
		assert.NoError(t, err)

		var count int
		err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM rooms").Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 1, count, "One entry should have been added to rooms table")

		err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM players").Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 1, count, "One entry should have been added to players table")

		err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM rooms_players").Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 1, count, "One entry should have been added to rooms_players table")
	})
}
