package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gitlab.com/hmajid2301/banterbus/internal/service"
	mockService "gitlab.com/hmajid2301/banterbus/internal/service/mocks"
)

func TestSubmitAnswer(t *testing.T) {
	t.Run("Should successfully submit answer", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		service := service.NewRoundService(mockStore)

		now := time.Now()
		ctx := context.Background()
		mockStore.EXPECT().SubmitAnswer(ctx, "fbb75599-9f7a-4392-b523-fd433b3208ea", "My answer", now).Return(nil)

		err := service.SubmitAnswer(ctx, "fbb75599-9f7a-4392-b523-fd433b3208ea", "My answer", now)
		assert.NoError(t, err)
	})

	t.Run("Should fail to submit answer, when database throws an error", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		service := service.NewRoundService(mockStore)

		now := time.Now()
		ctx := context.Background()
		mockStore.EXPECT().SubmitAnswer(ctx, "fbb75599-9f7a-4392-b523-fd433b3208ea", "My answer", now).
			Return(fmt.Errorf("failed to submit answer"))

		err := service.SubmitAnswer(ctx, "fbb75599-9f7a-4392-b523-fd433b3208ea", "My answer", now)
		assert.Error(t, err)
	})
}