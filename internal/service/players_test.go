package service_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	mockService "gitlab.com/hmajid2301/banterbus/internal/mocks/service"
	"gitlab.com/hmajid2301/banterbus/internal/service"
	sqlc "gitlab.com/hmajid2301/banterbus/internal/store/db"
)

func TestPlayerServiceUpdateNickname(t *testing.T) {
	t.Run("Should update nickname successfully", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		service := service.NewPlayerService(mockStore, mockRandom)

		ctx := context.Background()
		mockStore.EXPECT().
			UpdateNickname(ctx, "new_nickname", "fbb75599-9f7a-4392-b523-fd433b3208ea").
			Return([]sqlc.GetAllPlayersInRoomRow{
				{
					ID:       "b75599-9f7a-4392-b523-fd433b3208ea",
					Nickname: "new_nickname",
					Avatar:   []byte(""),
					RoomCode: "ABC12",
				},
			}, nil)

		room, err := service.UpdateNickname(
			ctx,
			"new_nickname",
			"fbb75599-9f7a-4392-b523-fd433b3208ea",
		)

		assert.NoError(t, err)
		assert.Equal(t, "ABC12", room.Code)
		assert.Len(t, room.Players, 1)
		assert.NotEmpty(t, room.Players[0].Nickname)
	})

	t.Run("Should throw error when fail to update nickname in DB", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		service := service.NewPlayerService(mockStore, mockRandom)

		ctx := context.Background()
		mockStore.EXPECT().
			UpdateNickname(ctx, "new_nickname", "fbb75599-9f7a-4392-b523-fd433b3208ea").
			Return([]sqlc.GetAllPlayersInRoomRow{}, fmt.Errorf("failed to update nickname"))
		_, err := service.UpdateNickname(
			ctx,
			"new_nickname",
			"fbb75599-9f7a-4392-b523-fd433b3208ea",
		)
		assert.Error(t, err)
	})
}

func TestPlayerServiceGenerateAvatar(t *testing.T) {
	t.Run("Should generate new avatar successfully", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		service := service.NewPlayerService(mockStore, mockRandom)

		ctx := context.Background()
		mockRandom.EXPECT().GetAvatar().Return([]byte("123"))
		mockStore.EXPECT().
			UpdateAvatar(ctx, []byte("123"), "fbb75599-9f7a-4392-b523-fd433b3208ea").
			Return([]sqlc.GetAllPlayersInRoomRow{
				{
					ID:       "fbb75599-9f7a-4392-b523-fd433b3208ea",
					Nickname: "Majiy00",
					Avatar:   []byte("123"),
					RoomCode: "ABC12",
				},
			}, nil)

		room, err := service.GenerateNewAvatar(ctx, "fbb75599-9f7a-4392-b523-fd433b3208ea")

		assert.NoError(t, err)
		assert.Equal(t, "ABC12", room.Code)
		assert.Len(t, room.Players, 1)
		assert.NotEmpty(t, room.Players[0].Avatar)
	})

	t.Run("Should throw error when fail to update nickname in DB", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		service := service.NewPlayerService(mockStore, mockRandom)

		ctx := context.Background()
		mockRandom.EXPECT().GetAvatar().Return([]byte("123"))
		mockStore.EXPECT().
			UpdateAvatar(ctx, []byte("123"), "fbb75599-9f7a-4392-b523-fd433b3208ea").
			Return([]sqlc.GetAllPlayersInRoomRow{}, fmt.Errorf("failed to generate new avatar"))
		_, err := service.GenerateNewAvatar(ctx, "fbb75599-9f7a-4392-b523-fd433b3208ea")
		assert.Error(t, err)
	})
}
