package store_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/hmajid2301/banterbus/internal/entities"
	"gitlab.com/hmajid2301/banterbus/internal/store"
)

func TestIntegrationUpdateNickname(t *testing.T) {
	t.Run("Should update player nickname in DB successfully", func(t *testing.T) {
		db, teardown := setupSubtest(t)
		defer teardown()

		myStore, err := store.NewStore(db)
		require.NoError(t, err)

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
		require.NoError(t, err)

		players, err := myStore.UpdateNickname(ctx, "Majiy01", newPlayer.ID)
		assert.Equal(t, "Majiy01", players[0].Nickname)
		assert.NoError(t, err)
	})

	t.Run("Should fail to update nickname, nickname already exists", func(t *testing.T) {
		db, teardown := setupSubtest(t)
		defer teardown()

		myStore, err := store.NewStore(db)
		require.NoError(t, err)

		newPlayer := entities.NewPlayer{
			ID:       "fbb75599-9f7a-4392-b523-fd433b3208ea",
			Nickname: "Majiy00",
			Avatar:   []byte(""),
		}

		newRoom := entities.NewRoom{
			GameName: "fibbing_it",
		}

		ctx := context.Background()
		_, err = myStore.CreateRoom(ctx, newPlayer, newRoom)
		require.NoError(t, err)

		_, err = myStore.UpdateNickname(ctx, newPlayer.Nickname, newPlayer.ID)
		assert.Error(t, err)
	})

	t.Run("Should fail to update nickname not in CREATED state", func(t *testing.T) {
		db, teardown := setupSubtest(t)
		defer teardown()

		myStore, err := store.NewStore(db)
		require.NoError(t, err)

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
		require.NoError(t, err)

		_, err = db.ExecContext(
			ctx,
			"UPDATE rooms SET room_state = 'PLAYING' WHERE room_code = ?",
			roomCode,
		)
		require.NoError(t, err)

		_, err = myStore.UpdateNickname(ctx, "Majiy01", newPlayer.ID)
		assert.Error(t, err)
	})
}

func TestIntegrationUpdatePlayer(t *testing.T) {
	t.Run("Should update player avatar in DB successfully", func(t *testing.T) {
		db, teardown := setupSubtest(t)
		defer teardown()

		myStore, err := store.NewStore(db)
		require.NoError(t, err)

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
		require.NoError(t, err)

		players, err := myStore.UpdateAvatar(ctx, []byte("123456"), newPlayer.ID)
		assert.Equal(t, []byte("123456"), players[0].Avatar)
		assert.NoError(t, err)
	})

	t.Run("Should fail to update avatar not in CREATED state", func(t *testing.T) {
		db, teardown := setupSubtest(t)
		defer teardown()

		myStore, err := store.NewStore(db)
		require.NoError(t, err)

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
		require.NoError(t, err)

		_, err = db.ExecContext(
			ctx,
			"UPDATE rooms SET room_state = 'PLAYING' WHERE room_code = ?",
			roomCode,
		)
		require.NoError(t, err)

		_, err = myStore.UpdateAvatar(ctx, []byte("123456"), newPlayer.ID)
		assert.Error(t, err)
	})
}

func TestIntegrationToggleIsReady(t *testing.T) {
	t.Run("Should toggle player ready state in DB successfully", func(t *testing.T) {
		db, teardown := setupSubtest(t)
		defer teardown()

		myStore, err := store.NewStore(db)
		require.NoError(t, err)

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
		require.NoError(t, err)

		players, err := myStore.ToggleIsReady(ctx, newPlayer.ID)
		assert.NoError(t, err)
		assert.True(t, players[0].IsReady.Bool)
	})

	t.Run("Should fail to toggle is ready state not in CREATED state", func(t *testing.T) {
		db, teardown := setupSubtest(t)
		defer teardown()

		myStore, err := store.NewStore(db)
		require.NoError(t, err)

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
		require.NoError(t, err)

		_, err = db.ExecContext(
			ctx,
			"UPDATE rooms SET room_state = 'PLAYING' WHERE room_code = ?",
			roomCode,
		)
		require.NoError(t, err)

		_, err = myStore.ToggleIsReady(ctx, newPlayer.ID)
		assert.Error(t, err)
	})
}
func TestIntegrationGetLobbyByPlayerID(t *testing.T) {
	t.Run("Should get lobby successfully", func(t *testing.T) {
		db, teardown := setupSubtest(t)
		defer teardown()

		myStore, err := store.NewStore(db)
		require.NoError(t, err)

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
		require.NoError(t, err)

		players, err := myStore.GetLobbyByPlayerID(ctx, newPlayer.ID)
		assert.NoError(t, err)
		// TODO: improve assert
		assert.Equal(t, newPlayer.ID, players[0].HostPlayer)
	})

	t.Run("Should fail to get lobby when player not in lobby", func(t *testing.T) {
		db, teardown := setupSubtest(t)
		defer teardown()

		myStore, err := store.NewStore(db)
		require.NoError(t, err)

		ctx := context.Background()
		_, err = myStore.GetLobbyByPlayerID(ctx, "wrongID")
		assert.Error(t, err)
	})
}

func TestIntegrationGetGameStateByPlayerID(t *testing.T) {
	t.Run("Should get game state successfully", func(t *testing.T) {
		db, teardown := setupSubtest(t)
		defer teardown()

		myStore, err := store.NewStore(db)
		require.NoError(t, err)

		ctx := context.Background()
		newPlayer := entities.NewPlayer{
			ID:       "fbb75599-9f7a-4392-b523-fd433b3208ea",
			Nickname: "Majiy00",
			Avatar:   []byte("1234"),
		}

		newRoom := entities.NewRoom{
			GameName: "fibbing_it",
		}

		roomCode, err := myStore.CreateRoom(ctx, newPlayer, newRoom)
		require.NoError(t, err)

		otherPlayer := entities.NewPlayer{
			ID:       "123",
			Nickname: "AnotherPlayer",
			Avatar:   []byte(""),
		}
		_, err = myStore.AddPlayerToRoom(ctx, otherPlayer, roomCode)
		require.NoError(t, err)

		players, err := myStore.GetLobbyByPlayerID(ctx, otherPlayer.ID)
		require.NoError(t, err)

		for _, player := range players {
			_, err = myStore.ToggleIsReady(ctx, player.ID)
			require.NoError(t, err)
		}

		_, err = myStore.StartGame(ctx, roomCode, newPlayer.ID)
		require.NoError(t, err)

		gameState, err := myStore.GetGameStateByPlayerID(ctx, players[0].ID)
		assert.NoError(t, err)
		assert.Equal(t, 1, gameState.Round)
		assert.Equal(t, "free_form", gameState.RoundType)
	})

	t.Run("Should fail to get game state player not in game", func(t *testing.T) {
		db, teardown := setupSubtest(t)
		defer teardown()

		myStore, err := store.NewStore(db)
		require.NoError(t, err)

		ctx := context.Background()
		_, err = myStore.GetGameStateByPlayerID(ctx, "wrongID")
		assert.Error(t, err)
	})
}
