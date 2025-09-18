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
