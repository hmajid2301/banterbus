package store_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/hmajid2301/banterbus/internal/entities"
	"gitlab.com/hmajid2301/banterbus/internal/store"
	sqlc "gitlab.com/hmajid2301/banterbus/internal/store/db"
)

func TestIntegrationCreateRoom(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("Should create room in DB successfully", func(t *testing.T) {
		db, teardown := setupSubtest(t)
		defer teardown()

		myStore, err := store.NewStore(db)
		assert.NoError(t, err)

		ctx := context.Background()
		newPlayer := entities.NewPlayer{
			Nickname: "Majiy00",
			Avatar:   []byte(""),
		}

		newRoom := entities.NewRoom{
			GameName: "fibbing_it",
		}

		roomCode, err := myStore.CreateRoom(ctx, newPlayer, newRoom)
		assert.NotEmpty(t, roomCode, "room code should not be empty")
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

	t.Run("Should create room in DB with correct state", func(t *testing.T) {
		db, teardown := setupSubtest(t)
		defer teardown()

		myStore, err := store.NewStore(db)
		assert.NoError(t, err)

		ctx := context.Background()
		newPlayer := entities.NewPlayer{
			Nickname: "Majiy00",
			Avatar:   []byte(""),
		}

		newRoom := entities.NewRoom{
			GameName: "fibbing_it",
		}

		roomCode, err := myStore.CreateRoom(ctx, newPlayer, newRoom)
		assert.NotEmpty(t, roomCode, "room code should not be empty")
		assert.NoError(t, err)

		var i sqlc.Room
		err = db.QueryRowContext(ctx, "SELECT * FROM rooms").Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.GameName,
			&i.HostPlayer,
			&i.RoomState,
			&i.RoomCode,
		)
		assert.NoError(t, err)
		assert.Equal(t, roomCode, i.RoomCode, "Room code returned should match room code in DB")
		assert.Equal(t, store.CREATED.String(), i.RoomState, "Room state should be CREATED")
		assert.Equal(t, newRoom.GameName, i.GameName, "Game name should be fibbing_it")
	})
}

func TestIntegrationAddPlayerToRoom(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("Should successfully join room", func(t *testing.T) {
		db, teardown := setupSubtest(t)
		defer teardown()

		myStore, err := store.NewStore(db)
		assert.NoError(t, err)

		ctx := context.Background()
		roomCode, err := createRoom(ctx, myStore)
		assert.NoError(t, err)

		newPlayer := entities.NewPlayer{
			ID:       "123",
			Nickname: "AnotherPlayer",
			Avatar:   []byte(""),
		}
		players, err := myStore.AddPlayerToRoom(ctx, newPlayer, roomCode)
		assert.Len(t, players, 2, "There should be 2 players in the room")
		assert.NoError(t, err)

		assert.Equal(
			t,
			roomCode,
			players[0].RoomCode,
			"Room code should returned match created room, room code",
		)
		assert.NoError(t, err)
	})

	t.Run("Should fail to join room not in CREATED state", func(t *testing.T) {
		db, teardown := setupSubtest(t)
		defer teardown()

		myStore, err := store.NewStore(db)
		assert.NoError(t, err)

		ctx := context.Background()
		roomCode, err := createRoom(ctx, myStore)
		assert.NoError(t, err)

		_, err = db.ExecContext(
			ctx,
			"UPDATE rooms SET room_state = 'PLAYING' WHERE room_code = ?",
			roomCode,
		)
		assert.NoError(t, err)

		newPlayer := entities.NewPlayer{
			ID:       "123",
			Nickname: "AnotherPlayer",
			Avatar:   []byte(""),
		}
		player, err := myStore.AddPlayerToRoom(ctx, newPlayer, roomCode)
		assert.Error(t, err)
		assert.Empty(t, player)
	})
}
