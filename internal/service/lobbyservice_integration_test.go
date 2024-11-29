package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/hmajid2301/banterbus/internal/service"
	"gitlab.com/hmajid2301/banterbus/internal/service/randomizer"
	sqlc "gitlab.com/hmajid2301/banterbus/internal/store/db"
)

const defaultHostPlayerID = "123"
const defaultHostNickname = "host_player"
const defaultOtherPlayerID = "456"
const defaultOtherPlayerNickname = "another_player"

func TestIntegrationLobbyCreate(t *testing.T) {
	t.Run("Should successfully create room", func(t *testing.T) {
		db, teardown := setupSubtest(t)
		defer teardown()

		str, err := sqlc.NewDB(db)
		assert.NoError(t, err)
		randomizer := randomizer.NewUserRandomizer()

		newPlayer := service.NewHostPlayer{
			ID: "123",
		}
		srv := service.NewLobbyService(str, randomizer)

		ctx := context.Background()
		lobby, err := srv.Create(ctx, "fibbing_it", newPlayer)

		assert.NoError(t, err)
		assert.Len(t, roomCode, 5)
		assert.Len(t, lobby.Players, 1, "There should be 1 player in the room when it is created")
		assert.Equal(t, "123", lobby.Players[0].ID, "Player ID should match the ID of player who created the room")
		assert.NotEmpty(t, lobby.Players[0].Nickname)
		assert.True(t, lobby.Players[0].IsHost, "Player should be the host")
		assert.False(t, lobby.Players[0].IsReady, "Player should not be ready when room is created")
	})

	t.Run("Should create room successfully, when player specifies their own nickname", func(t *testing.T) {
		db, teardown := setupSubtest(t)
		defer teardown()

		str, err := sqlc.NewDB(db)
		assert.NoError(t, err)
		randomizer := randomizer.NewUserRandomizer()

		newPlayer := service.NewHostPlayer{
			ID:       "123",
			Nickname: "Majiy00",
		}
		srv := service.NewLobbyService(str, randomizer)

		ctx := context.Background()
		lobby, err := srv.Create(ctx, "fibbing_it", newPlayer)

		assert.NoError(t, err)
		assert.Equal(t, "123", lobby.Players[0].ID)
		assert.Equal(t, "Majiy00", lobby.Players[0].Nickname)
	})
}

func TestIntegrationLobbyJoin(t *testing.T) {
	t.Run("Should successfully join room", func(t *testing.T) {
		db, teardown := setupSubtest(t)
		defer teardown()

		str, err := sqlc.NewDB(db)
		assert.NoError(t, err)
		randomizer := randomizer.NewUserRandomizer()

		srv := service.NewLobbyService(str, randomizer)

		ctx := context.Background()
		lobby, err := createRoom(ctx, srv)
		assert.NoError(t, err)

		lobby, err = srv.Join(ctx, lobby.Code, "456", "")
		assert.NoError(t, err)

		joinedPlayer := lobby.Players[1]

		assert.Len(t, lobby.Players, 2)
		assert.Equal(t, "456", joinedPlayer.ID, "Player ID should match the ID of player who joined the room")
		assert.NotEmpty(t, joinedPlayer.Nickname)
		assert.False(t, joinedPlayer.IsHost, "Player who joined should not be the host")
		assert.False(t, joinedPlayer.IsReady, "Player who joined should not be ready")
	})

	t.Run("Should successfully join room, with nickname", func(t *testing.T) {
		db, teardown := setupSubtest(t)
		defer teardown()

		str, err := sqlc.NewDB(db)
		assert.NoError(t, err)
		randomizer := randomizer.NewUserRandomizer()

		srv := service.NewLobbyService(str, randomizer)

		ctx := context.Background()
		lobby, err := createRoom(ctx, srv)
		assert.NoError(t, err)

		lobby, err = srv.Join(ctx, lobby.Code, "456", "nickname")
		assert.NoError(t, err)

		joinedPlayer := lobby.Players[1]
		assert.Len(t, lobby.Players, 2)
		assert.Equal(t, "nickname", joinedPlayer.Nickname)
	})

	t.Run("Should fail to join room where room code doesn't exist", func(t *testing.T) {
		db, teardown := setupSubtest(t)
		defer teardown()

		str, err := sqlc.NewDB(db)
		assert.NoError(t, err)
		randomizer := randomizer.NewUserRandomizer()

		srv := service.NewLobbyService(str, randomizer)

		ctx := context.Background()
		_, err = createRoom(ctx, srv)
		assert.NoError(t, err)

		_, err = srv.Join(ctx, "ABC12", "456", "nickname")
		assert.Error(t, err)
	})

	t.Run("Should fail to join room where not in CREATED state", func(t *testing.T) {
		db, teardown := setupSubtest(t)
		defer teardown()

		str, err := sqlc.NewDB(db)
		assert.NoError(t, err)
		randomizer := randomizer.NewUserRandomizer()

		srv := service.NewLobbyService(str, randomizer)

		ctx := context.Background()
		lobby, err := createRoom(ctx, srv)
		assert.NoError(t, err)

		_, err = db.ExecContext(
			ctx,
			"UPDATE rooms SET room_state = 'PLAYING' WHERE room_code = ?",
			lobby.Code,
		)
		assert.NoError(t, err)

		_, err = srv.Join(ctx, lobby.Code, "456", "nickname")
		assert.ErrorContains(t, err, "room is not in CREATED state")
	})
}

func TestIntegrationLobbyKickPlayer(t *testing.T) {
	t.Run("Should successfully kick player from room", func(t *testing.T) {
		db, teardown := setupSubtest(t)
		defer teardown()

		str, err := sqlc.NewDB(db)
		assert.NoError(t, err)
		randomizer := randomizer.NewUserRandomizer()

		srv := service.NewLobbyService(str, randomizer)
		ctx := context.Background()
		lobby, err := lobbyWithTwoPlayers(ctx, srv)
		assert.NoError(t, err)

		lobby, playerKickedID, err := srv.KickPlayer(ctx, lobby.Code, defaultHostPlayerID, defaultOtherPlayerNickname)
		assert.NoError(t, err)
		assert.Equal(t, defaultOtherPlayerID, playerKickedID)
		assert.Len(t, lobby.Players, 1)
		assert.Equal(t, defaultHostPlayerID, lobby.Players[0].ID)
	})

	t.Run("Should fail to kick player because room code does not exist", func(t *testing.T) {
		db, teardown := setupSubtest(t)
		defer teardown()

		str, err := sqlc.NewDB(db)
		assert.NoError(t, err)
		randomizer := randomizer.NewUserRandomizer()

		srv := service.NewLobbyService(str, randomizer)
		ctx := context.Background()
		_, err = lobbyWithTwoPlayers(ctx, srv)
		assert.NoError(t, err)

		_, _, err = srv.KickPlayer(ctx, "ABC12", defaultHostPlayerID, defaultOtherPlayerNickname)
		assert.Error(t, err)
	})

	t.Run("Should fail to kick player is not host of the room", func(t *testing.T) {
		db, teardown := setupSubtest(t)
		defer teardown()

		str, err := sqlc.NewDB(db)
		assert.NoError(t, err)
		randomizer := randomizer.NewUserRandomizer()

		srv := service.NewLobbyService(str, randomizer)
		ctx := context.Background()
		lobby, err := lobbyWithTwoPlayers(ctx, srv)
		assert.NoError(t, err)

		_, _, err = srv.KickPlayer(ctx, lobby.Code, defaultOtherPlayerID, defaultHostNickname)
		assert.ErrorContains(t, err, "player is not the host of the room")
	})

	t.Run("Should fail to kick player because room is not in CREATED state", func(t *testing.T) {
		db, teardown := setupSubtest(t)
		defer teardown()

		str, err := sqlc.NewDB(db)
		assert.NoError(t, err)
		randomizer := randomizer.NewUserRandomizer()

		srv := service.NewLobbyService(str, randomizer)
		ctx := context.Background()
		lobby, err := lobbyWithTwoPlayers(ctx, srv)
		assert.NoError(t, err)

		_, err = db.ExecContext(
			ctx,
			"UPDATE rooms SET room_state = 'PLAYING' WHERE room_code = ?",
			lobby.Code,
		)
		assert.NoError(t, err)

		_, _, err = srv.KickPlayer(ctx, lobby.Code, defaultHostPlayerID, defaultOtherPlayerNickname)
		assert.ErrorContains(t, err, "room is not in CREATED state")
	})

	t.Run("Should fail to kick player because player with nickname not in room", func(t *testing.T) {
		db, teardown := setupSubtest(t)
		defer teardown()

		str, err := sqlc.NewDB(db)
		assert.NoError(t, err)
		randomizer := randomizer.NewUserRandomizer()

		srv := service.NewLobbyService(str, randomizer)
		ctx := context.Background()
		lobby, err := lobbyWithTwoPlayers(ctx, srv)
		assert.NoError(t, err)

		_, _, err = srv.KickPlayer(ctx, lobby.Code, defaultHostPlayerID, "wrong_nickname")
		assert.ErrorContains(t, err, "player with nickname wrong_nickname not found to kick")
	})
}

func TestIntegrationLobbyStart(t *testing.T) {
	t.Run("Should successfully start game", func(t *testing.T) {
		db, teardown := setupSubtest(t)
		defer teardown()

		str, err := sqlc.NewDB(db)
		assert.NoError(t, err)
		randomizer := randomizer.NewUserRandomizer()

		srv := service.NewLobbyService(str, randomizer)
		plySrv := service.NewPlayerService(str, randomizer)
		ctx := context.Background()
		lobby, err := lobbyWithTwoPlayers(ctx, srv)
		assert.NoError(t, err)

		_, err = plySrv.TogglePlayerIsReady(ctx, defaultHostPlayerID)
		assert.NoError(t, err)
		_, err = plySrv.TogglePlayerIsReady(ctx, defaultOtherPlayerID)
		assert.NoError(t, err)

		gameState, err := srv.Start(ctx, lobby.Code, defaultHostPlayerID)
		assert.NoError(t, err)

		assert.NotEmpty(t, gameState.GameStateID)
		assert.Equal(t, 1, gameState.Round)
		assert.Equal(t, "free_form", gameState.RoundType)
		assert.Len(t, gameState.Players, 2)

		fibberCount := 0
		normalCount := 0
		for _, player := range gameState.Players {
			if player.Role == "fibber" {
				fibberCount++
			} else {
				normalCount++
			}
		}

		assert.Equal(t, 1, fibberCount, "There should be 1 fibber in the game")
		assert.Equal(t, 1, normalCount)
		assert.NotEqual(
			t,
			gameState.Players[0].Question,
			gameState.Players[1].Question,
			"Questions should be different between fibber and normals",
		)
	})

	t.Run("Should fail to start game because room not found", func(t *testing.T) {
		db, teardown := setupSubtest(t)
		defer teardown()

		str, err := sqlc.NewDB(db)
		assert.NoError(t, err)
		randomizer := randomizer.NewUserRandomizer()

		srv := service.NewLobbyService(str, randomizer)
		plySrv := service.NewPlayerService(str, randomizer)
		ctx := context.Background()
		_, err = lobbyWithTwoPlayers(ctx, srv)
		assert.NoError(t, err)

		_, err = plySrv.TogglePlayerIsReady(ctx, defaultHostPlayerID)
		assert.NoError(t, err)
		_, err = plySrv.TogglePlayerIsReady(ctx, defaultOtherPlayerID)
		assert.NoError(t, err)

		_, err = srv.Start(ctx, "unknown_code", defaultHostPlayerID)
		assert.Error(t, err)
	})

	t.Run("Should fail to start game because player starting is not host of the room", func(t *testing.T) {
		db, teardown := setupSubtest(t)
		defer teardown()

		str, err := sqlc.NewDB(db)
		assert.NoError(t, err)
		randomizer := randomizer.NewUserRandomizer()

		srv := service.NewLobbyService(str, randomizer)
		plySrv := service.NewPlayerService(str, randomizer)
		ctx := context.Background()
		lobby, err := lobbyWithTwoPlayers(ctx, srv)
		assert.NoError(t, err)

		_, err = plySrv.TogglePlayerIsReady(ctx, defaultHostPlayerID)
		assert.NoError(t, err)
		_, err = plySrv.TogglePlayerIsReady(ctx, defaultOtherPlayerID)
		assert.NoError(t, err)

		_, err = srv.Start(ctx, lobby.Code, defaultOtherPlayerID)
		assert.ErrorContains(t, err, "player is not the host of the room")
	})

	t.Run("Should fail to start game because room is not in CREATED state", func(t *testing.T) {
		db, teardown := setupSubtest(t)
		defer teardown()

		str, err := sqlc.NewDB(db)
		assert.NoError(t, err)
		randomizer := randomizer.NewUserRandomizer()

		srv := service.NewLobbyService(str, randomizer)
		plySrv := service.NewPlayerService(str, randomizer)
		ctx := context.Background()
		lobby, err := lobbyWithTwoPlayers(ctx, srv)
		assert.NoError(t, err)

		_, err = plySrv.TogglePlayerIsReady(ctx, defaultHostPlayerID)
		assert.NoError(t, err)
		_, err = plySrv.TogglePlayerIsReady(ctx, defaultOtherPlayerID)
		assert.NoError(t, err)

		_, err = db.ExecContext(
			ctx,
			"UPDATE rooms SET room_state = 'PLAYING' WHERE room_code = ?",
			lobby.Code,
		)

		assert.NoError(t, err)
		_, err = srv.Start(ctx, lobby.Code, defaultHostPlayerID)
		assert.ErrorContains(t, err, "room is not in CREATED state")
	})

	t.Run("Should fail to start game as there is only one player in room", func(t *testing.T) {
		db, teardown := setupSubtest(t)
		defer teardown()

		str, err := sqlc.NewDB(db)
		assert.NoError(t, err)
		randomizer := randomizer.NewUserRandomizer()

		srv := service.NewLobbyService(str, randomizer)
		ctx := context.Background()
		lobby, err := createRoom(ctx, srv)
		assert.NoError(t, err)

		_, err = srv.Start(ctx, lobby.Code, defaultHostPlayerID)
		assert.ErrorContains(t, err, "not enough players to start the game")
	})

	t.Run("Should fail to start game as not every one is ready", func(t *testing.T) {
		db, teardown := setupSubtest(t)
		defer teardown()

		str, err := sqlc.NewDB(db)
		assert.NoError(t, err)
		randomizer := randomizer.NewUserRandomizer()

		srv := service.NewLobbyService(str, randomizer)
		plySrv := service.NewPlayerService(str, randomizer)
		ctx := context.Background()
		lobby, err := lobbyWithTwoPlayers(ctx, srv)
		assert.NoError(t, err)

		_, err = plySrv.TogglePlayerIsReady(ctx, defaultHostPlayerID)
		assert.NoError(t, err)

		_, err = srv.Start(ctx, lobby.Code, defaultHostPlayerID)
		assert.ErrorContains(t, err, "not all players are ready")
	})
}
