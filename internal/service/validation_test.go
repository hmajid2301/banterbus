package service_test

import (
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

		ctx := t.Context()
		playerID, err := uuid.NewV7()
		require.NoError(t, err)

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

		ctx := t.Context()
		playerID, err := uuid.NewV7()
		require.NoError(t, err)

		validAnswer := strings.Repeat("a", 500)

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(
			db.Room{}, errors.New("simulated db error"))

		err = srv.SubmitAnswer(ctx, playerID, validAnswer, time.Now())

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "simulated db error")
		assert.NotContains(t, err.Error(), "answer too long")
	})

	t.Run("Should reject empty answer", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		playerID, err := uuid.NewV7()
		require.NoError(t, err)

		err = srv.SubmitAnswer(ctx, playerID, "", time.Now())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "answer cannot be empty")
	})
}
