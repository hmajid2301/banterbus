package store_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/hmajid2301/banterbus/internal/banterbustest"
	"gitlab.com/hmajid2301/banterbus/internal/entities"
	"gitlab.com/hmajid2301/banterbus/internal/store"
	sqlc "gitlab.com/hmajid2301/banterbus/internal/store/db"
)

func setupSubtest(t *testing.T) (*sql.DB, func()) {
	ctx := context.Background()
	db, err := banterbustest.CreateDB(ctx)
	require.NoError(t, err)

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

func TestIntegrationCreateRoom(t *testing.T) {
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

	t.Run("Should fail to join room, nickname already exists", func(t *testing.T) {
		db, teardown := setupSubtest(t)
		defer teardown()

		myStore, err := store.NewStore(db)
		assert.NoError(t, err)

		ctx := context.Background()
		roomCode, err := createRoom(ctx, myStore)
		assert.NoError(t, err)

		newPlayer := entities.NewPlayer{
			ID:       "123",
			Nickname: "Majiy00",
			Avatar:   []byte(""),
		}
		player, err := myStore.AddPlayerToRoom(ctx, newPlayer, roomCode)
		assert.Error(t, err)
		assert.Empty(t, player)
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

func TestIntegrationStartGame(t *testing.T) {
	t.Run("Should successfully start game", func(t *testing.T) {
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
		assert.NoError(t, err)

		for _, player := range players {
			_, err = myStore.ToggleIsReady(ctx, player.ID)
			require.NoError(t, err)
		}

		// INFO: first player is the host, so only they can start the game
		gameState, err := myStore.StartGame(ctx, roomCode, players[0].ID)
		assert.NoError(t, err)
		assert.Len(t, gameState.Players, 2, "There should be 2 players in the room")

		var roomState string
		err = db.QueryRowContext(ctx, "SELECT room_state FROM rooms WHERE room_code = ?", roomCode).Scan(&roomState)
		assert.NoError(t, err)

		assert.Equal(t, store.PLAYING.String(), roomState, "Room should be in PLAYING state after starting game")
		assert.NoError(t, err)

		playerOne := gameState.Players[0]
		playerTwo := gameState.Players[1]

		assert.NotEqual(t, playerOne.Role, playerTwo.Role, "Players should have different roles")
		assert.NotEqual(t, playerOne.Question, playerTwo.Question, "Players should have differnt questions")
	})

	t.Run("Should fail to start game, player is not host", func(t *testing.T) {
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
		assert.NoError(t, err)

		for _, player := range players {
			_, err = myStore.ToggleIsReady(ctx, player.ID)
			require.NoError(t, err)
		}

		_, err = myStore.StartGame(ctx, roomCode, players[1].ID)
		assert.Error(t, err)
	})

	t.Run("Should fail to start game, game state is not CREATED", func(t *testing.T) {
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
		assert.NoError(t, err)

		for _, player := range players {
			_, err = myStore.ToggleIsReady(ctx, player.ID)
			require.NoError(t, err)
		}

		_, err = db.ExecContext(
			ctx,
			"UPDATE rooms SET room_state = 'PLAYING' WHERE room_code = ?",
			roomCode,
		)
		assert.NoError(t, err)

		_, err = myStore.StartGame(ctx, roomCode, players[0].ID)
		assert.Error(t, err)
	})

	t.Run("Should fail to start game, not enough players in the game", func(t *testing.T) {
		db, teardown := setupSubtest(t)
		defer teardown()

		myStore, err := store.NewStore(db)
		assert.NoError(t, err)

		ctx := context.Background()
		roomCode, err := createRoom(ctx, myStore)
		assert.NoError(t, err)

		var player sqlc.Player
		query := `
            SELECT p.id, p.created_at, p.updated_at, p.avatar, p.nickname, p.is_ready
            FROM players p
            JOIN rooms_players rp ON p.id = rp.player_id
            JOIN rooms r ON rp.room_id = r.id
            WHERE r.room_code = ? LIMIT 1;
        `
		err = db.QueryRowContext(ctx, query, roomCode).Scan(
			&player.ID,
			&player.CreatedAt,
			&player.UpdatedAt,
			&player.Avatar,
			&player.Nickname,
			&player.IsReady,
		)
		assert.NoError(t, err)

		_, err = myStore.StartGame(ctx, roomCode, player.ID)
		assert.Error(t, err)
	})

	t.Run("Should fail to start game, not all players are ready", func(t *testing.T) {
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
		assert.NoError(t, err)

		_, err = myStore.StartGame(ctx, roomCode, players[0].ID)
		assert.Error(t, err)
	})
}
