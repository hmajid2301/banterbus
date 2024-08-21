package service_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/hmajid2301/banterbus/internal/entities"
	mockService "gitlab.com/hmajid2301/banterbus/internal/mocks/service"
	"gitlab.com/hmajid2301/banterbus/internal/service"
	sqlc "gitlab.com/hmajid2301/banterbus/internal/store/db"
)

func TestRoomServiceCreate(t *testing.T) {
	t.Run("Should create room in DB successfully", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)

		service := service.NewRoomService(mockStore, mockRandom)

		newPlayer := entities.NewHostPlayer{
			ID: "fbb75599-9f7a-4392-b523-fd433b3208ea",
		}

		newCreatedPlayer := entities.NewPlayer{
			ID:       newPlayer.ID,
			Nickname: "Majiy00",
			Avatar:   []byte(""),
		}

		ctx := context.Background()
		mockRandom.EXPECT().GetNickname().Return(newCreatedPlayer.Nickname)
		mockRandom.EXPECT().GetAvatar().Return(newCreatedPlayer.Avatar)
		mockStore.EXPECT().
			CreateRoom(ctx, newCreatedPlayer, entities.NewRoom{GameName: "fibbing_it"}).
			Return("ABC12", nil)
		room, err := service.Create(ctx, "fibbing_it", newPlayer)

		assert.NoError(t, err)
		assert.Equal(t, "ABC12", room.Code)
		assert.Len(t, room.Players, 1)
		assert.NotEmpty(t, room.Players[0].Nickname)
	})

	t.Run("Should create room in DB successfully, when nickname is passed", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)

		service := service.NewRoomService(mockStore, mockRandom)

		newPlayer := entities.NewHostPlayer{
			ID:       "fbb75599-9f7a-4392-b523-fd433b3208ea",
			Nickname: "Majiy01",
		}

		newCreatedPlayer := entities.NewPlayer{
			ID:       newPlayer.ID,
			Nickname: "Majiy01",
			Avatar:   []byte(""),
		}

		ctx := context.Background()
		mockRandom.EXPECT().GetAvatar().Return(newCreatedPlayer.Avatar)
		mockStore.EXPECT().
			CreateRoom(ctx, newCreatedPlayer, entities.NewRoom{GameName: "fibbing_it"}).
			Return("ABC12", nil)
		room, err := service.Create(ctx, "fibbing_it", newPlayer)

		assert.NoError(t, err)
		assert.Len(t, room.Players, 1)
		assert.Equal(t, newPlayer.Nickname, room.Players[0].Nickname)
	})

	t.Run("Should throw error when fail to create room in DB", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)

		service := service.NewRoomService(mockStore, mockRandom)

		newPlayer := entities.NewHostPlayer{
			ID:       "fbb75599-9f7a-4392-b523-fd433b3208ea",
			Nickname: "Majiy01",
		}

		newCreatedPlayer := entities.NewPlayer{
			ID:       newPlayer.ID,
			Nickname: "Majiy01",
			Avatar:   []byte(""),
		}

		ctx := context.Background()
		mockRandom.EXPECT().GetAvatar().Return(newCreatedPlayer.Avatar)
		mockStore.EXPECT().
			CreateRoom(ctx, newCreatedPlayer, entities.NewRoom{GameName: "fibbing_it"}).
			Return("", fmt.Errorf("failed to create room"))
		_, err := service.Create(ctx, "fibbing_it", newPlayer)

		assert.Error(t, err)
	})
}

func TestRoomServiceJoin(t *testing.T) {
	t.Run("Should join room in DB successfully", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)

		service := service.NewRoomService(mockStore, mockRandom)

		newPlayer := entities.NewHostPlayer{
			ID: "fbb75599-9f7a-4392-b523-fd433b3208ea",
		}

		newCreatedPlayer := entities.NewPlayer{
			ID:       newPlayer.ID,
			Nickname: "Majiy00",
			Avatar:   []byte(""),
		}

		ctx := context.Background()
		mockRandom.EXPECT().GetNickname().Return(newCreatedPlayer.Nickname)
		mockRandom.EXPECT().GetAvatar().Return(newCreatedPlayer.Avatar)
		mockStore.EXPECT().
			AddPlayerToRoom(ctx, newCreatedPlayer, "ABC12").
			Return([]sqlc.GetAllPlayersInRoomRow{
				{
					ID:       "b75599-9f7a-4392-b523-fd433b3208ea",
					Nickname: "EmotionalTiger",
					Avatar:   []byte(""),
					RoomCode: "ABC12",
				},
				{
					ID:       newCreatedPlayer.ID,
					Nickname: newCreatedPlayer.Nickname,
					Avatar:   []byte(""),
					RoomCode: "ABC12",
				},
			}, nil)
		room, err := service.Join(ctx, "ABC12", newPlayer.ID, newPlayer.Nickname)

		assert.NoError(t, err)
		assert.Equal(t, "ABC12", room.Code, "room code should be the one the player tries to join")
		assert.Len(t, room.Players, 2, "should be two players in the room")
		assert.NotEmpty(
			t,
			room.Players[1].Nickname,
			"should get a nickname set if the user didn't provide one",
		)
	})

	t.Run("Should join room in DB successfully, when user sets own nickname", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)

		service := service.NewRoomService(mockStore, mockRandom)

		newPlayer := entities.NewHostPlayer{
			ID:       "fbb75599-9f7a-4392-b523-fd433b3208ea",
			Nickname: "Majiy00",
		}

		newCreatedPlayer := entities.NewPlayer{
			ID:       newPlayer.ID,
			Nickname: "Majiy00",
			Avatar:   []byte(""),
		}

		ctx := context.Background()
		mockRandom.EXPECT().GetAvatar().Return(newCreatedPlayer.Avatar)
		mockStore.EXPECT().
			AddPlayerToRoom(ctx, newCreatedPlayer, "ABC12").
			Return([]sqlc.GetAllPlayersInRoomRow{
				{
					ID:       "b75599-9f7a-4392-b523-fd433b3208ea",
					Nickname: "EmotionalTiger",
					Avatar:   []byte(""),
					RoomCode: "ABC12",
				},
				{
					ID:       newCreatedPlayer.ID,
					Nickname: newCreatedPlayer.Nickname,
					Avatar:   []byte(""),
					RoomCode: "ABC12",
				},
			}, nil)
		room, err := service.Join(ctx, "ABC12", newPlayer.ID, newPlayer.Nickname)

		assert.NoError(t, err)
		assert.Equal(t, "ABC12", room.Code, "room code should be the one the player tries to join")
		assert.Len(t, room.Players, 2, "should be two players in the room")
		assert.NotEmpty(
			t,
			room.Players[1].Nickname,
			"should get a nickname set if the user didn't provide one",
		)
	})

	t.Run("Should throw error when fail to join room in DB", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)

		service := service.NewRoomService(mockStore, mockRandom)
		newCreatedPlayer := entities.NewPlayer{
			ID:       "fbb75599-9f7a-4392-b523-fd433b3208ea",
			Nickname: "Majiy01",
			Avatar:   []byte(""),
		}

		ctx := context.Background()
		mockRandom.EXPECT().GetAvatar().Return(newCreatedPlayer.Avatar)
		mockStore.EXPECT().
			AddPlayerToRoom(ctx, newCreatedPlayer, "ABC12").
			Return([]sqlc.GetAllPlayersInRoomRow{}, fmt.Errorf("failed to join room"))
		_, err := service.Join(ctx, "ABC12", newCreatedPlayer.ID, newCreatedPlayer.Nickname)

		assert.Error(t, err)
	})
}
