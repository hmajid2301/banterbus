package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/hmajid2301/banterbus/internal/service"
	mockService "gitlab.com/hmajid2301/banterbus/internal/service/mocks"
	"gitlab.com/hmajid2301/banterbus/internal/store/db"
)

func TestAnswerLengthValidation(t *testing.T) {
	t.Parallel()

	t.Run("Should reject answer that exceeds maximum length", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := context.Background()
		playerID, err := uuid.NewV7()
		require.NoError(t, err)

		// Create an answer that exceeds the 500 character limit
		longAnswer := strings.Repeat("a", 501)

		err = srv.SubmitAnswer(ctx, playerID, longAnswer, time.Now())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "answer too long")
		assert.Contains(t, err.Error(), "501 characters")
		assert.Contains(t, err.Error(), "max 500")
	})

	t.Run("Should accept answer within length limits", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := context.Background()
		playerID, err := uuid.NewV7()
		require.NoError(t, err)

		// Create an answer that is exactly at the limit
		validAnswer := strings.Repeat("a", 500)

		// Mock the expected database calls to simulate a valid game state
		// The service will call GetRoomByPlayerID first
		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(
			// Return an error to fail the test after validation passes
			// This proves that validation logic completed successfully
			db.Room{}, errors.New("simulated db error"))

		err = srv.SubmitAnswer(ctx, playerID, validAnswer, time.Now())

		// The error should be our simulated DB error, NOT about length validation
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "simulated db error")
		assert.NotContains(t, err.Error(), "answer too long")
	})

	t.Run("Should reject empty answer", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := context.Background()
		playerID, err := uuid.NewV7()
		require.NoError(t, err)

		err = srv.SubmitAnswer(ctx, playerID, "", time.Now())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "answer cannot be empty")
	})
}
