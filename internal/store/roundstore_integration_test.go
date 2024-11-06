package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/hmajid2301/banterbus/internal/store"
)

func TestIntegrationSubmitAnswer(t *testing.T) {
	t.Run("Should successfully submit answer", func(t *testing.T) {
		db, teardown := setupSubtest(t)
		defer teardown()

		myStore, err := store.NewStore(db)
		require.NoError(t, err)

		ctx := context.Background()
		gameState, err := startGame(ctx, myStore)
		require.NoError(t, err)

		err = myStore.SubmitAnswer(ctx, gameState.Players[0].ID, "This is my answer", time.Now())
		assert.NoError(t, err)

		err = myStore.SubmitAnswer(ctx, gameState.Players[1].ID, "This is the other players answer", time.Now())
		assert.NoError(t, err)
	})

	t.Run("Should fail to submit answer, time has passed", func(t *testing.T) {
		db, teardown := setupSubtest(t)
		defer teardown()

		myStore, err := store.NewStore(db)
		require.NoError(t, err)

		ctx := context.Background()
		gameState, err := startGame(ctx, myStore)
		require.NoError(t, err)

		err = myStore.SubmitAnswer(ctx, gameState.Players[0].ID, "This is my answer", time.Now().Add(120*time.Second))
		assert.Error(t, err)
	})

	t.Run("Should fail to submit answer, player id doesn't belong to room", func(t *testing.T) {
		db, teardown := setupSubtest(t)
		defer teardown()

		myStore, err := store.NewStore(db)
		require.NoError(t, err)

		ctx := context.Background()
		_, err = startGame(ctx, myStore)
		require.NoError(t, err)

		err = myStore.SubmitAnswer(ctx, "7979797606", "This is my answer", time.Now().Add(120*time.Second))
		assert.Error(t, err)
	})
}
