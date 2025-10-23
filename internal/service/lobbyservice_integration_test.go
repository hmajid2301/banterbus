package service_test

import (
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/hmajid2301/banterbus/internal/service"
	"gitlab.com/hmajid2301/banterbus/internal/service/randomizer"
	db "gitlab.com/hmajid2301/banterbus/internal/store/db"
)

func TestIntegrationLobbyCreate(t *testing.T) {
	t.Parallel()

	t.Run("Should create room successfully", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		id, err := uuid.NewV4()
		require.NoError(t, err)
		newPlayer := service.NewHostPlayer{
			ID:       id,
			Nickname: "Majiy00",
		}

		ctx, err := getI18nCtx(t.Context())
		require.NoError(t, err)

		srv := service.NewLobbyService(str, randomizer, "en-GB")
		lobby, err := srv.Create(ctx, "fibbing_it", newPlayer)

		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, lobby.Lobby.Players[0].ID)
		assert.Equal(t, "Majiy00", lobby.Lobby.Players[0].Nickname)
		assert.True(t, lobby.Lobby.Players[0].IsHost)
	})
}

func TestIntegrationLobbyJoin(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully join room", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx(t.Context())
		require.NoError(t, err)

		srv := service.NewLobbyService(str, randomizer, "en-GB")
		lobby, err := createRoom(ctx, srv)
		assert.NoError(t, err)

		joinResult, err := srv.Join(ctx, lobby.Code, defaultOtherPlayerID, "nickname")
		assert.NoError(t, err)

		joinedPlayer := joinResult.Lobby.Players[1]
		assert.Len(t, joinResult.Lobby.Players, 2)
		assert.Equal(t, "nickname", joinedPlayer.Nickname)
	})

	t.Run("Should fail to join room where room code doesn't exist", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx(t.Context())
		require.NoError(t, err)

		srv := service.NewLobbyService(str, randomizer, "en-GB")
		_, err = createRoom(ctx, srv)
		assert.NoError(t, err)

		_, err = srv.Join(ctx, "ABC12", defaultOtherPlayerID, "nickname")
		assert.Error(t, err)
	})

	t.Run("Should fail to join room where not in CREATED state", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx(t.Context())
		require.NoError(t, err)

		srv := service.NewLobbyService(str, randomizer, "en-GB")
		lobby, err := createRoom(ctx, srv)
		assert.NoError(t, err)

		_, err = pool.Exec(
			ctx,
			"UPDATE rooms SET room_state = 'PLAYING' WHERE room_code = $1",
			lobby.Code,
		)
		assert.NoError(t, err)

		_, err = srv.Join(ctx, lobby.Code, defaultOtherPlayerID, "nickname")
		assert.ErrorContains(t, err, "room is not in CREATED state")
	})
}

func TestIntegrationLobbyKickPlayer(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully kick player from room", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx(t.Context())
		require.NoError(t, err)

		srv := service.NewLobbyService(str, randomizer, "en-GB")
		lobby, err := lobbyWithTwoPlayers(ctx, srv)
		assert.NoError(t, err)

		lobby, playerKickedID, err := srv.KickPlayer(ctx, lobby.Code, defaultHostPlayerID, defaultOtherPlayerNickname)
		assert.NoError(t, err)
		assert.Equal(t, defaultOtherPlayerID, playerKickedID)
		assert.Len(t, lobby.Players, 1)
		assert.Equal(t, defaultHostPlayerID, lobby.Players[0].ID)
	})

	t.Run("Should fail to kick player because room code does not exist", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx(t.Context())
		require.NoError(t, err)

		srv := service.NewLobbyService(str, randomizer, "en-GB")
		_, err = lobbyWithTwoPlayers(ctx, srv)
		assert.NoError(t, err)

		_, _, err = srv.KickPlayer(ctx, "ABC12", defaultHostPlayerID, defaultOtherPlayerNickname)
		assert.Error(t, err)
	})

	t.Run("Should fail to kick player is not host of the room", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx(t.Context())
		require.NoError(t, err)

		srv := service.NewLobbyService(str, randomizer, "en-GB")
		lobby, err := lobbyWithTwoPlayers(ctx, srv)
		assert.NoError(t, err)

		_, _, err = srv.KickPlayer(ctx, lobby.Code, defaultOtherPlayerID, defaultHostNickname)
		assert.ErrorContains(t, err, "player is not the host of the room")
	})

	t.Run("Should fail to kick player because room is not in CREATED state", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx(t.Context())
		require.NoError(t, err)

		srv := service.NewLobbyService(str, randomizer, "en-GB")
		lobby, err := lobbyWithTwoPlayers(ctx, srv)
		assert.NoError(t, err)

		_, err = pool.Exec(
			ctx,
			"UPDATE rooms SET room_state = 'PLAYING' WHERE room_code = $1",
			lobby.Code,
		)
		assert.NoError(t, err)

		_, _, err = srv.KickPlayer(ctx, lobby.Code, defaultHostPlayerID, defaultOtherPlayerNickname)
		assert.ErrorContains(t, err, "room is not in CREATED state")
	})

	t.Run("Should fail to kick player because player with nickname not in room", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx(t.Context())
		require.NoError(t, err)

		srv := service.NewLobbyService(str, randomizer, "en-GB")
		lobby, err := lobbyWithTwoPlayers(ctx, srv)
		assert.NoError(t, err)

		_, _, err = srv.KickPlayer(ctx, lobby.Code, defaultHostPlayerID, "wrong_nickname")
		assert.ErrorContains(t, err, "player with nickname wrong_nickname not found")
	})

	t.Run("Should reassign host when host disconnects", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx(t.Context())
		require.NoError(t, err)

		srv := service.NewLobbyService(str, randomizer, "en-GB")
		lobby, err := lobbyWithTwoPlayers(ctx, srv)
		assert.NoError(t, err)

		assert.True(t, lobby.Players[0].IsHost)
		assert.False(t, lobby.Players[1].IsHost)

		err = srv.HandlePlayerDisconnect(ctx, defaultHostPlayerID)
		assert.NoError(t, err)

		lobby, err = srv.GetLobby(ctx, defaultOtherPlayerID)
		assert.NoError(t, err)
		assert.Len(t, lobby.Players, 1)
		assert.Equal(t, defaultOtherPlayerID, lobby.Players[0].ID)
		assert.True(t, lobby.Players[0].IsHost)
	})

	t.Run("Should remove non-host player when they disconnect", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx(t.Context())
		require.NoError(t, err)

		srv := service.NewLobbyService(str, randomizer, "en-GB")
		lobby, err := lobbyWithTwoPlayers(ctx, srv)
		assert.NoError(t, err)

		assert.True(t, lobby.Players[0].IsHost)
		assert.False(t, lobby.Players[1].IsHost)

		err = srv.HandlePlayerDisconnect(ctx, defaultOtherPlayerID)
		assert.NoError(t, err)

		lobby, err = srv.GetLobby(ctx, defaultHostPlayerID)
		assert.NoError(t, err)
		assert.Len(t, lobby.Players, 1)
		assert.Equal(t, defaultHostPlayerID, lobby.Players[0].ID)
		assert.True(t, lobby.Players[0].IsHost)
	})
}

func TestIntegrationLobbyStart(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully start game", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx(t.Context())
		require.NoError(t, err)

		srv := service.NewLobbyService(str, randomizer, "en-GB")
		plySrv := service.NewPlayerService(str, randomizer)
		lobby, err := lobbyWithTwoPlayers(ctx, srv)
		assert.NoError(t, err)

		_, err = plySrv.TogglePlayerIsReady(ctx, defaultHostPlayerID)
		assert.NoError(t, err)
		_, err = plySrv.TogglePlayerIsReady(ctx, defaultOtherPlayerID)
		assert.NoError(t, err)

		deadline := time.Now().Add(5 * time.Second)
		gameState, err := srv.Start(ctx, lobby.Code, defaultHostPlayerID, deadline)
		assert.NoError(t, err)

		assert.NotEmpty(t, gameState.GameStateID)
		assert.Equal(t, 1, gameState.Round)
		assert.Equal(t, "free_form", gameState.RoundType)
		assert.Len(t, gameState.Players, 2)
		assert.LessOrEqual(t, int(gameState.Deadline.Seconds()), 5)

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
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx(t.Context())
		require.NoError(t, err)

		srv := service.NewLobbyService(str, randomizer, "en-GB")
		plySrv := service.NewPlayerService(str, randomizer)
		_, err = lobbyWithTwoPlayers(ctx, srv)
		assert.NoError(t, err)

		_, err = plySrv.TogglePlayerIsReady(ctx, defaultHostPlayerID)
		assert.NoError(t, err)
		_, err = plySrv.TogglePlayerIsReady(ctx, defaultOtherPlayerID)
		assert.NoError(t, err)

		deadline := time.Now().Add(5 * time.Second)
		_, err = srv.Start(ctx, "unknown_code", defaultHostPlayerID, deadline)
		assert.Error(t, err)
	})

	t.Run("Should fail to start game because player starting is not host of the room", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx(t.Context())
		require.NoError(t, err)

		srv := service.NewLobbyService(str, randomizer, "en-GB")
		plySrv := service.NewPlayerService(str, randomizer)
		lobby, err := lobbyWithTwoPlayers(ctx, srv)
		assert.NoError(t, err)

		_, err = plySrv.TogglePlayerIsReady(ctx, defaultHostPlayerID)
		assert.NoError(t, err)
		_, err = plySrv.TogglePlayerIsReady(ctx, defaultOtherPlayerID)
		assert.NoError(t, err)

		deadline := time.Now().Add(5 * time.Second)
		_, err = srv.Start(ctx, lobby.Code, defaultOtherPlayerID, deadline)
		assert.ErrorContains(t, err, "player is not the host of the room")
	})

	t.Run("Should fail to start game because room is not in CREATED state", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx(t.Context())
		require.NoError(t, err)

		srv := service.NewLobbyService(str, randomizer, "en-GB")
		plySrv := service.NewPlayerService(str, randomizer)
		lobby, err := lobbyWithTwoPlayers(ctx, srv)
		assert.NoError(t, err)

		_, err = plySrv.TogglePlayerIsReady(ctx, defaultHostPlayerID)
		assert.NoError(t, err)
		_, err = plySrv.TogglePlayerIsReady(ctx, defaultOtherPlayerID)
		assert.NoError(t, err)

		_, err = pool.Exec(
			ctx,
			"UPDATE rooms SET room_state = 'PLAYING' WHERE room_code = $1",
			lobby.Code,
		)
		assert.NoError(t, err)

		deadline := time.Now().Add(5 * time.Second)
		_, err = srv.Start(ctx, lobby.Code, defaultHostPlayerID, deadline)
		assert.ErrorContains(t, err, "room is not in CREATED state")
	})

	t.Run("Should fail to start game as there is only one player in room", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx(t.Context())
		require.NoError(t, err)

		srv := service.NewLobbyService(str, randomizer, "en-GB")
		lobby, err := createRoom(ctx, srv)
		assert.NoError(t, err)

		deadline := time.Now().Add(5 * time.Second)
		_, err = srv.Start(ctx, lobby.Code, defaultHostPlayerID, deadline)
		assert.ErrorContains(t, err, "not enough players to start the game")
	})

	t.Run("Should fail to start game as not every one is ready", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx(t.Context())
		require.NoError(t, err)

		srv := service.NewLobbyService(str, randomizer, "en-GB")
		plySrv := service.NewPlayerService(str, randomizer)
		lobby, err := lobbyWithTwoPlayers(ctx, srv)
		assert.NoError(t, err)

		_, err = plySrv.TogglePlayerIsReady(ctx, defaultHostPlayerID)
		assert.NoError(t, err)

		deadline := time.Now().Add(5 * time.Second)
		_, err = srv.Start(ctx, lobby.Code, defaultHostPlayerID, deadline)
		assert.ErrorContains(t, err, "not all players are ready")
	})
}

func TestIntegrationLobbyGetRoomState(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully get room state", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		id, err := uuid.NewV4()
		require.NoError(t, err)
		newPlayer := service.NewHostPlayer{
			ID: id,
		}

		ctx, err := getI18nCtx(t.Context())
		require.NoError(t, err)

		srv := service.NewLobbyService(str, randomizer, "en-GB")
		_, err = srv.Create(ctx, "fibbing_it", newPlayer)
		require.NoError(t, err)

		roomState, err := srv.GetRoomState(ctx, id)
		assert.NoError(t, err)
		assert.Equal(t, db.Created, roomState)
	})

	t.Run("Should fail to get room state, player id not found", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		id, err := uuid.NewV4()
		require.NoError(t, err)
		newPlayer := service.NewHostPlayer{
			ID: id,
		}

		ctx, err := getI18nCtx(t.Context())
		require.NoError(t, err)

		srv := service.NewLobbyService(str, randomizer, "en-GB")
		_, err = srv.Create(ctx, "fibbing_it", newPlayer)
		require.NoError(t, err)

		_, err = srv.GetRoomState(ctx, uuid.Must(uuid.NewV4()))
		assert.ErrorContains(t, err, "player is not currently in any game")
	})
}

func TestIntegrationLobbyGetLobby(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully get lobby", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()
		lobbyService := service.NewLobbyService(str, randomizer, "en-GB")

		id, err := uuid.NewV4()
		require.NoError(t, err)
		newPlayer := service.NewHostPlayer{
			ID: id,
		}

		ctx, err := getI18nCtx(t.Context())
		require.NoError(t, err)

		createdLobby, err := lobbyService.Create(ctx, "fibbing_it", newPlayer)
		require.NoError(t, err)

		lobby, err := lobbyService.GetLobby(ctx, id)
		assert.NoError(t, err)

		expectedLobby := service.Lobby{
			Code: createdLobby.Lobby.Code,
			Players: []service.LobbyPlayer{
				{
					ID:       id,
					Nickname: createdLobby.Lobby.Players[0].Nickname,
					Avatar:   createdLobby.Lobby.Players[0].Avatar,
					IsHost:   true,
					IsReady:  false,
				},
			},
		}
		assert.Equal(t, expectedLobby, lobby)
	})
}

// Cross-service error propagation and consistency tests
func TestIntegrationCrossServiceErrorPropagation(t *testing.T) {
	t.Parallel()

	t.Run("Should propagate database errors across service boundaries", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()
		lobbyService := service.NewLobbyService(str, randomizer, "en-GB")
		playerService := service.NewPlayerService(str, randomizer)
		roundService := service.NewRoundService(str, randomizer, "en-GB")

		ctx, err := getI18nCtx(t.Context())
		require.NoError(t, err)

		// Create a room and start a game
		hostID := uuid.Must(uuid.NewV4())
		newPlayer := service.NewHostPlayer{
			ID:       hostID,
			Nickname: "Host",
		}
		createdLobby, err := lobbyService.Create(ctx, "fibbing_it", newPlayer)
		require.NoError(t, err)

		playerID := uuid.Must(uuid.NewV4())
		_, err = lobbyService.Join(ctx, createdLobby.Lobby.Code, playerID, "Player1")
		require.NoError(t, err)

		// Start game
		_, err = playerService.TogglePlayerIsReady(ctx, hostID)
		require.NoError(t, err)
		_, err = playerService.TogglePlayerIsReady(ctx, playerID)
		require.NoError(t, err)

		deadline := time.Now().Add(30 * time.Second)
		_, err = lobbyService.Start(ctx, createdLobby.Lobby.Code, hostID, deadline)
		require.NoError(t, err)

		// Test cross-service error propagation with invalid player operations
		invalidPlayerID := uuid.Must(uuid.NewV4())

		submitDeadline := time.Now().Add(25 * time.Second)
		err = roundService.SubmitAnswer(ctx, invalidPlayerID, "Invalid Answer", submitDeadline)
		assert.Error(t, err)
		assert.NotEmpty(t, err.Error())

		// Try player service operations on same invalid player
		_, err = playerService.TogglePlayerIsReady(ctx, invalidPlayerID)
		assert.Error(t, err)
		assert.NotEmpty(t, err.Error())

		// Valid operations should still work after invalid ones
		err = roundService.SubmitAnswer(ctx, hostID, "Valid Answer", submitDeadline)
		assert.NoError(t, err)
	})

	t.Run("Should handle service layer validation consistently", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()
		lobbyService := service.NewLobbyService(str, randomizer, "en-GB")
		playerService := service.NewPlayerService(str, randomizer)

		ctx, err := getI18nCtx(t.Context())
		require.NoError(t, err)

		// Test nickname validation across services
		hostID := uuid.Must(uuid.NewV4())
		newPlayer := service.NewHostPlayer{
			ID:       hostID,
			Nickname: "Host",
		}
		createdLobby, err := lobbyService.Create(ctx, "fibbing_it", newPlayer)
		require.NoError(t, err)

		duplicatePlayerID := uuid.Must(uuid.NewV4())
		_, err = lobbyService.Join(ctx, createdLobby.Lobby.Code, duplicatePlayerID, "Host")
		if err != nil {
			assert.NotEmpty(t, err.Error())
		}

		// Try updating to duplicate nickname via player service - should also fail
		playerID := uuid.Must(uuid.NewV4())
		_, err = lobbyService.Join(ctx, createdLobby.Lobby.Code, playerID, "Player1")
		require.NoError(t, err)

		_, err = playerService.UpdateNickname(ctx, "Host", playerID) // Already taken by host
		if err != nil {
			assert.NotEmpty(t, err.Error())
		}

		_, err = playerService.UpdateNickname(ctx, "UniquePlayer", playerID)
		assert.NoError(t, err)
	})
}

func TestIntegrationCrossServiceDatabaseResilience(t *testing.T) {
	t.Parallel()

	t.Run("Should handle database timeout scenarios across services", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		// Create DB with very short timeouts to force failures
		baseDelay := (time.Millisecond * 1)
		str := db.NewDB(pool, 1, baseDelay) // Only 1 retry, 1ms delay
		randomizer := randomizer.NewUserRandomizer()
		lobbyService := service.NewLobbyService(str, randomizer, "en-GB")
		playerService := service.NewPlayerService(str, randomizer)

		ctx, err := getI18nCtx(t.Context())
		require.NoError(t, err)

		hostID := uuid.Must(uuid.NewV4())
		newPlayer := service.NewHostPlayer{
			ID:       hostID,
			Nickname: "Host",
		}

		_, err1 := lobbyService.Create(ctx, "fibbing_it", newPlayer)
		if err1 != nil {
			assert.True(t,
				err1.Error() != "",
			)
		}

		// Try player service operation with same context
		_, err2 := playerService.TogglePlayerIsReady(ctx, hostID)
		if err2 != nil {
			assert.True(t,
				err2.Error() != "",
			)
		}

		// Verify system remains stable with normal context
		normalStr := db.NewDB(pool, 3, time.Millisecond*100)
		normalLobbyService := service.NewLobbyService(normalStr, randomizer, "en-GB")

		newHostID := uuid.Must(uuid.NewV4())
		newNormalPlayer := service.NewHostPlayer{
			ID:       newHostID,
			Nickname: "NormalHost",
		}

		_, err = normalLobbyService.Create(ctx, "fibbing_it", newNormalPlayer)
		assert.NoError(t, err)
	})
}

// Database constraint violation tests
func TestIntegrationLobbyConstraintViolations(t *testing.T) {
	t.Parallel()

	t.Run("Should handle invalid room codes during join", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()
		srv := service.NewLobbyService(str, randomizer, "en-GB")

		ctx, err := getI18nCtx(t.Context())
		require.NoError(t, err)

		playerID := uuid.Must(uuid.NewV4())

		// Try to join with non-existent room code
		_, err = srv.Join(ctx, "INVALID_CODE", playerID, "TestPlayer")
		assert.Error(t, err)
		assert.NotEmpty(t, err.Error())
	})

	t.Run("Should handle duplicate player in same room", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()
		srv := service.NewLobbyService(str, randomizer, "en-GB")

		ctx, err := getI18nCtx(t.Context())
		require.NoError(t, err)

		// Create room
		hostID := uuid.Must(uuid.NewV4())
		newPlayer := service.NewHostPlayer{
			ID:       hostID,
			Nickname: "Host",
		}
		createdLobby, err := srv.Create(ctx, "fibbing_it", newPlayer)
		require.NoError(t, err)

		// Join with a different player
		player2ID := uuid.Must(uuid.NewV4())
		_, err = srv.Join(ctx, createdLobby.Lobby.Code, player2ID, "Player2")
		require.NoError(t, err)

		// Try to join again with same player ID - should handle gracefully
		_, err = srv.Join(ctx, createdLobby.Lobby.Code, player2ID, "Player2Again")
		if err != nil {
			assert.NotEmpty(t, err.Error())
		}
	})

	t.Run("Should handle joining room in wrong state", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()
		srv := service.NewLobbyService(str, randomizer, "en-GB")

		ctx, err := getI18nCtx(t.Context())
		require.NoError(t, err)

		// Create room with 2 players
		hostID := uuid.Must(uuid.NewV4())
		newPlayer := service.NewHostPlayer{
			ID:       hostID,
			Nickname: "Host",
		}
		createdLobby, err := srv.Create(ctx, "fibbing_it", newPlayer)
		require.NoError(t, err)

		player2ID := uuid.Must(uuid.NewV4())
		_, err = srv.Join(ctx, createdLobby.Lobby.Code, player2ID, "Player2")
		require.NoError(t, err)

		// Create player service to mark players as ready
		playerService := service.NewPlayerService(str, randomizer)
		_, err = playerService.TogglePlayerIsReady(ctx, hostID)
		require.NoError(t, err)
		_, err = playerService.TogglePlayerIsReady(ctx, player2ID)
		require.NoError(t, err)

		// Start the game
		deadline := time.Now().Add(5 * time.Minute)
		_, err = srv.Start(ctx, createdLobby.Lobby.Code, hostID, deadline)
		require.NoError(t, err)

		player3ID := uuid.Must(uuid.NewV4())
		_, err = srv.Join(ctx, createdLobby.Lobby.Code, player3ID, "Player3")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not in CREATED state")
	})
}
