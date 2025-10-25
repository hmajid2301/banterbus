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

func TestIntegrationRoundServiceSubmitAnswer(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully submit answer", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx(t.Context())
		require.NoError(t, err)

		lobbyService := service.NewLobbyService(str, randomizer, "en-GB")
		playerService := service.NewPlayerService(str, randomizer)
		roundService := service.NewRoundService(str, randomizer, "en-GB")

		_, err = startGame(ctx, lobbyService, playerService)
		require.NoError(t, err)

		hostID := uuid.Must(uuid.FromString("0193a62a-4dff-774c-850a-b1fe78e2a8d1"))

		pauseStatus, err := roundService.PauseGame(ctx, hostID)
		require.NoError(t, err)
		assert.True(t, pauseStatus.IsPaused)
		assert.NotNil(t, pauseStatus.PausedAt)
		assert.LessOrEqual(t, pauseStatus.PauseTimeRemainingMs, int32(300000))

		time.Sleep(500 * time.Millisecond)

		pauseStatus2, err := roundService.ResumeGame(ctx, hostID)
		require.NoError(t, err)
		assert.False(t, pauseStatus2.IsPaused)
		assert.Nil(t, pauseStatus2.PausedAt)
		assert.Less(t, pauseStatus2.PauseTimeRemainingMs, pauseStatus.PauseTimeRemainingMs)
	})

	t.Run("Should prevent non-host from pausing", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx(t.Context())
		require.NoError(t, err)

		lobbyService := service.NewLobbyService(str, randomizer, "en-GB")
		playerService := service.NewPlayerService(str, randomizer)
		roundService := service.NewRoundService(str, randomizer, "en-GB")

		_, err = startGame(ctx, lobbyService, playerService)
		require.NoError(t, err)

		nonHostID := uuid.Must(uuid.FromString("0193a62a-4dff-774c-850a-b1fe78e2a8d2"))

		_, err = roundService.PauseGame(ctx, nonHostID)
		assert.ErrorIs(t, err, service.ErrNotHost)
	})

	t.Run("Should prevent pausing when already paused", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx(t.Context())
		require.NoError(t, err)

		lobbyService := service.NewLobbyService(str, randomizer, "en-GB")
		playerService := service.NewPlayerService(str, randomizer)
		roundService := service.NewRoundService(str, randomizer, "en-GB")

		_, err = startGame(ctx, lobbyService, playerService)
		require.NoError(t, err)

		hostID := uuid.Must(uuid.FromString("0193a62a-4dff-774c-850a-b1fe78e2a8d1"))

		_, err = roundService.PauseGame(ctx, hostID)
		require.NoError(t, err)

		_, err = roundService.PauseGame(ctx, hostID)
		assert.ErrorIs(t, err, service.ErrGameAlreadyPaused)
	})

	t.Run("Should prevent resuming when not paused", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx(t.Context())
		require.NoError(t, err)

		lobbyService := service.NewLobbyService(str, randomizer, "en-GB")
		playerService := service.NewPlayerService(str, randomizer)
		roundService := service.NewRoundService(str, randomizer, "en-GB")

		_, err = startGame(ctx, lobbyService, playerService)
		require.NoError(t, err)

		hostID := uuid.Must(uuid.FromString("0193a62a-4dff-774c-850a-b1fe78e2a8d1"))

		_, err = roundService.ResumeGame(ctx, hostID)
		assert.ErrorIs(t, err, service.ErrGameNotPaused)
	})

	t.Run("Should correctly track pause time budget", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx(t.Context())
		require.NoError(t, err)

		lobbyService := service.NewLobbyService(str, randomizer, "en-GB")
		playerService := service.NewPlayerService(str, randomizer)
		roundService := service.NewRoundService(str, randomizer, "en-GB")

		_, err = startGame(ctx, lobbyService, playerService)
		require.NoError(t, err)

		hostID := uuid.Must(uuid.FromString("0193a62a-4dff-774c-850a-b1fe78e2a8d1"))

		pauseStatus1, err := roundService.PauseGame(ctx, hostID)
		require.NoError(t, err)
		initialBudget := pauseStatus1.PauseTimeRemainingMs

		time.Sleep(500 * time.Millisecond)

		resumeStatus1, err := roundService.ResumeGame(ctx, hostID)
		require.NoError(t, err)
		budgetAfterFirstPause := resumeStatus1.PauseTimeRemainingMs

		assert.Less(t, budgetAfterFirstPause, initialBudget)

		pauseStatus2, err := roundService.PauseGame(ctx, hostID)
		require.NoError(t, err)
		assert.Equal(t, budgetAfterFirstPause, pauseStatus2.PauseTimeRemainingMs)
	})
}
