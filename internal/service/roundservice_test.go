package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gitlab.com/hmajid2301/banterbus/internal/service"
	mockService "gitlab.com/hmajid2301/banterbus/internal/service/mocks"
	sqlc "gitlab.com/hmajid2301/banterbus/internal/store/db"
)

func TestRoundServiceSubmitAnswer(t *testing.T) {
	roomID := "fbb75599-9f7a-4392-b523-fd433b3208ea"
	roundID := "222222-fbb75599-9f7a-4392-b523-fd433b3208ea"
	playerID := "33333-9f7a-4392-b523-fd433b3208ea"

	t.Run("Should successfully submit answer", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		now := time.Now()
		ctx := context.Background()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(sqlc.Room{
			ID:        roomID,
			RoomState: sqlc.ROOMSTATE_PLAYING.String(),
		}, nil)
		mockStore.EXPECT().GetLatestRoundByPlayerID(ctx, playerID).Return(
			sqlc.FibbingItRound{
				ID:             roundID,
				SubmitDeadline: now.Add(1 * time.Hour),
			}, nil)

		mockRandomizer.EXPECT().GetID().Return("12345")
		mockStore.EXPECT().AddFibbingItAnswer(ctx, sqlc.AddFibbingItAnswerParams{
			ID:       "12345",
			RoundID:  roundID,
			PlayerID: playerID,
			Answer:   "My answer",
		}).Return(sqlc.FibbingItAnswer{}, nil)

		err := srv.SubmitAnswer(ctx, playerID, "My answer", now)
		assert.NoError(t, err)
	})

	t.Run("Should fail to submit answer because we fail to get room details", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		now := time.Now()
		ctx := context.Background()

		mockStore.EXPECT().
			GetRoomByPlayerID(ctx, playerID).
			Return(sqlc.Room{}, fmt.Errorf("failed to get room details"))
		err := srv.SubmitAnswer(ctx, playerID, "My answer", now)
		assert.Error(t, err)
	})

	t.Run("Should fail to submit answer because room not in PLAYING state", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		now := time.Now()
		ctx := context.Background()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(sqlc.Room{
			ID:        roomID,
			RoomState: sqlc.ROOMSTATE_CREATED.String(),
		}, nil)

		err := srv.SubmitAnswer(ctx, playerID, "My answer", now)
		assert.ErrorContains(t, err, "room is not in PLAYING state")
	})

	t.Run("Should fail to submit answer because failed to get latest round", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		service := service.NewRoundService(mockStore, mockRandomizer)

		now := time.Now()
		ctx := context.Background()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(sqlc.Room{
			ID:        roomID,
			RoomState: sqlc.ROOMSTATE_PLAYING.String(),
		}, nil)
		mockStore.EXPECT().GetLatestRoundByPlayerID(ctx, playerID).Return(
			sqlc.FibbingItRound{},
			fmt.Errorf("failed to get latest round"),
		)

		err := service.SubmitAnswer(ctx, playerID, "My answer", now)
		assert.Error(t, err)
	})

	t.Run("Should fail to submit answer because after submit time", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		now := time.Now()
		ctx := context.Background()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(sqlc.Room{
			ID:        roomID,
			RoomState: sqlc.ROOMSTATE_PLAYING.String(),
		}, nil)
		mockStore.EXPECT().GetLatestRoundByPlayerID(ctx, playerID).Return(
			sqlc.FibbingItRound{
				ID:             roundID,
				SubmitDeadline: time.Now().Add(-1 * time.Hour),
			}, nil)

		err := srv.SubmitAnswer(ctx, playerID, "My answer", now)
		assert.ErrorContains(t, err, "answer submission deadline has passed")
	})

	t.Run("Should fail to submit answer because failed to add answer to DB", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		now := time.Now()
		ctx := context.Background()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(sqlc.Room{
			ID:        roomID,
			RoomState: sqlc.ROOMSTATE_PLAYING.String(),
		}, nil)
		mockStore.EXPECT().GetLatestRoundByPlayerID(ctx, playerID).Return(
			sqlc.FibbingItRound{
				ID:             roundID,
				SubmitDeadline: now.Add(1 * time.Hour),
			}, nil)

		mockRandomizer.EXPECT().GetID().Return("12345")
		mockStore.EXPECT().AddFibbingItAnswer(ctx, sqlc.AddFibbingItAnswerParams{
			ID:       "12345",
			RoundID:  roundID,
			PlayerID: playerID,
			Answer:   "My answer",
		}).Return(sqlc.FibbingItAnswer{}, fmt.Errorf("failed to add answer to DB"))

		err := srv.SubmitAnswer(ctx, playerID, "My answer", now)
		assert.Error(t, err)
	})
}
