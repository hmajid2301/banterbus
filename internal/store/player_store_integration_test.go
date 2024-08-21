package store_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/hmajid2301/banterbus/internal/entities"
	"gitlab.com/hmajid2301/banterbus/internal/store"
)

func TestIntegrationUpdateNickname(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("Should update player nickname in DB successfully", func(t *testing.T) {
		db, teardown := setupSubtest(t)
		defer teardown()

		myStore, err := store.NewStore(db)
		assert.NoError(t, err)

		ctx := context.Background()
		newPlayer := entities.NewPlayer{
			ID:       "fbb75599-9f7a-4392-b523-fd433b3208ea",
			Nickname: "Majiy00",
			Avatar:   []byte(""),
		}

		newRoom := entities.NewRoom{
			GameName: "fibbing_it",
		}

		_, err = myStore.CreateRoom(ctx, newPlayer, newRoom)
		assert.NoError(t, err)

		players, err := myStore.UpdateNickname(ctx, "Majiy01", newPlayer.ID)
		assert.Equal(t, "Majiy01", players[0].Nickname)
		assert.NoError(t, err)
	})

	t.Run("Should fail to update nickname not in CREATED state", func(t *testing.T) {
		db, teardown := setupSubtest(t)
		defer teardown()

		myStore, err := store.NewStore(db)
		assert.NoError(t, err)

		newPlayer := entities.NewPlayer{
			ID:       "fbb75599-9f7a-4392-b523-fd433b3208ea",
			Nickname: "Majiy00",
			Avatar:   []byte(""),
		}

		newRoom := entities.NewRoom{
			GameName: "fibbing_it",
		}

		ctx := context.Background()
		roomCode, err := myStore.CreateRoom(ctx, newPlayer, newRoom)
		assert.NoError(t, err)

		_, err = db.ExecContext(
			ctx,
			"UPDATE rooms SET room_state = 'PLAYING' WHERE room_code = ?",
			roomCode,
		)
		assert.NoError(t, err)

		_, err = myStore.UpdateNickname(ctx, "Majiy01", newPlayer.ID)
		assert.Error(t, err)
	})
}

func TestIntegrationUpdatePlayer(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("Should update player avatar in DB successfully", func(t *testing.T) {
		db, teardown := setupSubtest(t)
		defer teardown()

		myStore, err := store.NewStore(db)
		assert.NoError(t, err)

		ctx := context.Background()
		newPlayer := entities.NewPlayer{
			ID:       "fbb75599-9f7a-4392-b523-fd433b3208ea",
			Nickname: "Majiy00",
			Avatar:   []byte("1234"),
		}

		newRoom := entities.NewRoom{
			GameName: "fibbing_it",
		}

		_, err = myStore.CreateRoom(ctx, newPlayer, newRoom)
		assert.NoError(t, err)

		players, err := myStore.UpdateAvatar(ctx, []byte("123456"), newPlayer.ID)
		assert.Equal(t, []byte("123456"), players[0].Avatar)
		assert.NoError(t, err)
	})

	t.Run("Should fail to update avatar not in CREATED state", func(t *testing.T) {
		db, teardown := setupSubtest(t)
		defer teardown()

		myStore, err := store.NewStore(db)
		assert.NoError(t, err)

		newPlayer := entities.NewPlayer{
			ID:       "fbb75599-9f7a-4392-b523-fd433b3208ea",
			Nickname: "Majiy00",
			Avatar:   []byte(""),
		}

		newRoom := entities.NewRoom{
			GameName: "fibbing_it",
		}

		ctx := context.Background()
		roomCode, err := myStore.CreateRoom(ctx, newPlayer, newRoom)
		assert.NoError(t, err)

		_, err = db.ExecContext(
			ctx,
			"UPDATE rooms SET room_state = 'PLAYING' WHERE room_code = ?",
			roomCode,
		)
		assert.NoError(t, err)

		_, err = myStore.UpdateAvatar(ctx, []byte("123456"), newPlayer.ID)
		assert.Error(t, err)
	})
}
