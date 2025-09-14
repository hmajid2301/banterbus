package service_test

import (
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/banterbus/banterbus/internal/service"
	"gitlab.com/banterbus/banterbus/internal/service/randomizer"
	"gitlab.com/banterbus/banterbus/internal/store/db"
)

func TestIntegrationPlayerUpdateNickname(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully update nickname", func(t *testing.T) {
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

		ctx, err := getI18nCtx()
		require.NoError(t, err)

		lobbyService := service.NewLobbyService(str, randomizer, "en-GB")
		lobbyService.Create(ctx, "fibbing_it", newPlayer)

		srv := service.NewPlayerService(str, randomizer)
		lobby, err := srv.UpdateNickname(ctx, "majiy01", id)
		assert.NoError(t, err)
		assert.Equal(t, "majiy01", lobby.Players[0].Nickname)
	})

	t.Run("Should fail to update nickname, because room is not in CREATED state", func(t *testing.T) {
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
			Nickname: "majiy01",
		}

		ctx, err := getI18nCtx()
		require.NoError(t, err)

		lobbyService := service.NewLobbyService(str, randomizer, "en-GB")
		lobby, err := lobbyService.Create(ctx, "fibbing_it", newPlayer)
		require.NoError(t, err)

		_, err = pool.Exec(
			ctx,
			"UPDATE rooms SET room_state = 'PLAYING' WHERE room_code = $1",
			lobby.Lobby.Code,
		)
		require.NoError(t, err)

		srv := service.NewPlayerService(str, randomizer)
		_, err = srv.UpdateNickname(ctx, "majiy01", id)
		assert.ErrorContains(t, err, "room is not in CREATED state")
	})

	t.Run("Should fail to update nickname, because nickname already exists", func(t *testing.T) {
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
			Nickname: "majiy01",
		}

		ctx, err := getI18nCtx()
		require.NoError(t, err)

		lobbyService := service.NewLobbyService(str, randomizer, "en-GB")
		lobbyService.Create(ctx, "fibbing_it", newPlayer)

		srv := service.NewPlayerService(str, randomizer)
		_, err = srv.UpdateNickname(ctx, "majiy01", id)
		assert.ErrorContains(t, err, "nickname already exists")
	})
}

func TestIntegrationPlayerGenerateNewAvatar(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully update avatar", func(t *testing.T) {
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

		ctx, err := getI18nCtx()
		require.NoError(t, err)

		lobbyService := service.NewLobbyService(str, randomizer, "en-GB")
		lobby, err := lobbyService.Create(ctx, "fibbing_it", newPlayer)
		require.NoError(t, err)
		oldAvatar := lobby.Lobby.Players[0].Avatar

		srv := service.NewPlayerService(str, randomizer)
		updatedLobby, err := srv.GenerateNewAvatar(ctx, id)
		assert.NoError(t, err)
		newAvatar := updatedLobby.Players[0].Avatar
		assert.NotEqual(t, oldAvatar, newAvatar)
	})

	t.Run("Should fail to update avatar, because room is not in CREATED state", func(t *testing.T) {
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

		ctx, err := getI18nCtx()
		require.NoError(t, err)

		lobbyService := service.NewLobbyService(str, randomizer, "en-GB")
		lobby, err := lobbyService.Create(ctx, "fibbing_it", newPlayer)
		require.NoError(t, err)

		_, err = pool.Exec(
			ctx,
			"UPDATE rooms SET room_state = 'PLAYING' WHERE room_code = $1",
			lobby.Lobby.Code,
		)
		require.NoError(t, err)

		srv := service.NewPlayerService(str, randomizer)
		_, err = srv.GenerateNewAvatar(ctx, id)
		assert.ErrorContains(t, err, "room is not in CREATED state")
	})
}

func TestIntegrationToggleIsReady(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully toggle not ready -> ready", func(t *testing.T) {
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

		ctx, err := getI18nCtx()
		require.NoError(t, err)

		lobbyService := service.NewLobbyService(str, randomizer, "en-GB")
		_, err = lobbyService.Create(ctx, "fibbing_it", newPlayer)
		require.NoError(t, err)

		srv := service.NewPlayerService(str, randomizer)
		lobby, err := srv.TogglePlayerIsReady(ctx, id)
		assert.NoError(t, err)
		assert.True(t, lobby.Players[0].IsReady)
	})

	t.Run("Should successfully toggle ready -> notready", func(t *testing.T) {
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

		ctx, err := getI18nCtx()
		require.NoError(t, err)

		lobbyService := service.NewLobbyService(str, randomizer, "en-GB")
		_, err = lobbyService.Create(ctx, "fibbing_it", newPlayer)
		require.NoError(t, err)

		srv := service.NewPlayerService(str, randomizer)
		_, err = srv.TogglePlayerIsReady(ctx, id)
		assert.NoError(t, err)
		lobby, err := srv.TogglePlayerIsReady(ctx, id)
		assert.NoError(t, err)
		assert.False(t, lobby.Players[0].IsReady)
	})

	t.Run("Should fail to update avatar, because player id doesn't exist", func(t *testing.T) {
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

		ctx, err := getI18nCtx()
		require.NoError(t, err)

		lobbyService := service.NewLobbyService(str, randomizer, "en-GB")
		_, err = lobbyService.Create(ctx, "fibbing_it", newPlayer)
		require.NoError(t, err)

		srv := service.NewPlayerService(str, randomizer)
		_, err = srv.TogglePlayerIsReady(ctx, uuid.Must(uuid.NewV4()))
		assert.ErrorContains(t, err, "no rows in result set")
	})

	t.Run("Should fail to update avatar, because room is not in CREATED state", func(t *testing.T) {
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

		ctx, err := getI18nCtx()
		require.NoError(t, err)

		lobbyService := service.NewLobbyService(str, randomizer, "en-GB")
		lobby, err := lobbyService.Create(ctx, "fibbing_it", newPlayer)
		require.NoError(t, err)

		_, err = pool.Exec(
			ctx,
			"UPDATE rooms SET room_state = 'PLAYING' WHERE room_code = $1",
			lobby.Lobby.Code,
		)
		require.NoError(t, err)

		srv := service.NewPlayerService(str, randomizer)
		_, err = srv.TogglePlayerIsReady(ctx, id)
		assert.ErrorContains(t, err, "room is not in CREATED state")
	})
}
