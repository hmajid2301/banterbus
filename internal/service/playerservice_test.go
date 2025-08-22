package service_test

import (
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mdobak/go-xerrors"
	"github.com/stretchr/testify/assert"

	"gitlab.com/hmajid2301/banterbus/internal/service"
	mockService "gitlab.com/hmajid2301/banterbus/internal/service/mocks"
	"gitlab.com/hmajid2301/banterbus/internal/store/db"
)

var playerID = uuid.Must(uuid.FromString("0193a625-dad1-7095-9abb-bebdad739381"))
var avatarURL = "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=randomSeed"

func TestPlayerServiceUpdateNickname(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully update nickname", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := t.Context()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(db.Room{
			ID:        roomID,
			RoomState: db.Created.String(),
		}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, playerID).Return([]db.GetAllPlayersInRoomRow{
			{
				Nickname: "Old Nickname",
			},
		}, nil).Times(1)
		mockStore.EXPECT().UpdateNickname(ctx, db.UpdateNicknameParams{
			Nickname: "New Nickname",
			ID:       playerID,
		}).Return(db.Player{}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, playerID).Return([]db.GetAllPlayersInRoomRow{
			{
				ID:         playerID,
				RoomCode:   roomCode,
				HostPlayer: hostPlayerID,
				Nickname:   "New Nickname",
				IsReady:    pgtype.Bool{Bool: false, Valid: true},
			},
		}, nil).Times(1)

		lobby, err := srv.UpdateNickname(ctx, "New Nickname", playerID)

		assert.NoError(t, err)
		expectedLobby := service.Lobby{
			Code: roomCode,
			Players: []service.LobbyPlayer{
				{
					ID:       playerID,
					Nickname: "New Nickname",
					IsReady:  false,
					IsHost:   false,
				},
			},
		}
		assert.Equal(t, expectedLobby, lobby)
	})

	t.Run("Should fail to update nickname because we failed to get room details", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := t.Context()
		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(
			db.Room{}, xerrors.New("failed to get room details"),
		)

		_, err := srv.UpdateNickname(ctx, "New Nickname", playerID)
		assert.Error(t, err)
	})

	t.Run("Should fail to update nickname because lobby not in CREATED room state", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := t.Context()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(db.Room{
			ID:        roomID,
			RoomState: db.Finished.String(),
		}, nil)

		_, err := srv.UpdateNickname(ctx, "New Nickname", playerID)
		assert.Error(t, err)
	})

	t.Run("Should fail to update nickname because, nickname already exists in lobby", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := t.Context()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(db.Room{
			ID:        roomID,
			RoomState: db.Created.String(),
		}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, playerID).Return([]db.GetAllPlayersInRoomRow{
			{
				Nickname: "Old Nickname",
			},
			{
				Nickname: "New Nickname",
			},
		}, nil).Times(1)

		_, err := srv.UpdateNickname(ctx, "New Nickname", playerID)
		assert.ErrorContains(t, err, "nickname already exists")
	})

	t.Run("Should fail to update nickname because update nickname in DB failed", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := t.Context()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(db.Room{
			ID:        roomID,
			RoomState: db.Created.String(),
		}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, playerID).Return([]db.GetAllPlayersInRoomRow{
			{
				Nickname: "Old Nickname",
			},
		}, nil).Times(1)
		mockStore.EXPECT().UpdateNickname(ctx, db.UpdateNicknameParams{
			Nickname: "New Nickname",
			ID:       playerID,
		}).Return(db.Player{}, xerrors.New("failed to update nickname"))

		_, err := srv.UpdateNickname(ctx, "New Nickname", playerID)
		assert.Error(t, err)
	})

	t.Run("Should fail to update nickname because get all players failed from DB", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := t.Context()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(db.Room{
			ID:        roomID,
			RoomState: db.Created.String(),
		}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, playerID).Return([]db.GetAllPlayersInRoomRow{
			{
				Nickname: "Old Nickname",
			},
		}, nil).Times(1)
		mockStore.EXPECT().UpdateNickname(ctx, db.UpdateNicknameParams{
			Nickname: "New Nickname",
			ID:       playerID,
		}).Return(db.Player{}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, playerID).Return(
			[]db.GetAllPlayersInRoomRow{}, xerrors.New("failed to get all players in room"),
		).Times(1)

		_, err := srv.UpdateNickname(ctx, "New Nickname", playerID)
		assert.Error(t, err)
	})
}

func TestPlayerServiceGenerateAvatar(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully generate avatar", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := t.Context()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(db.Room{
			ID:        roomID,
			RoomState: db.Created.String(),
		}, nil)
		mockRandomizer.EXPECT().GetAvatar("").Return(avatarURL)
		mockStore.EXPECT().UpdateAvatar(ctx, db.UpdateAvatarParams{
			Avatar: avatarURL,
			ID:     playerID,
		}).Return(db.Player{}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, playerID).Return([]db.GetAllPlayersInRoomRow{
			{
				ID:         playerID,
				Avatar:     avatarURL,
				RoomCode:   roomCode,
				HostPlayer: hostPlayerID,
				Nickname:   "nickname",
				IsReady:    pgtype.Bool{Bool: false, Valid: true},
			},
		}, nil).Times(1)

		lobby, err := srv.GenerateNewAvatar(ctx, playerID)

		assert.NoError(t, err)
		expectedLobby := service.Lobby{
			Code: roomCode,
			Players: []service.LobbyPlayer{
				{
					ID:       playerID,
					Nickname: "nickname",
					Avatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=randomSeed",
					IsReady:  false,
					IsHost:   false,
				},
			},
		}
		assert.Equal(t, expectedLobby, lobby)
	})

	t.Run("Should fail to update avatar because we fail to get room details from DB", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := t.Context()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(
			db.Room{}, xerrors.New("failed to get room details"),
		)

		_, err := srv.GenerateNewAvatar(ctx, playerID)
		assert.Error(t, err)
	})

	t.Run("Should fail to update avatar because lobby is not in CREATED state", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := t.Context()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(db.Room{
			ID:        roomID,
			RoomState: db.Playing.String(),
		}, nil)

		_, err := srv.GenerateNewAvatar(ctx, playerID)
		assert.ErrorContains(t, err, "room is not in CREATED state")
	})

	t.Run("Should fail to update avatar because we fail to update avatar in DB", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := t.Context()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(db.Room{
			ID:        roomID,
			RoomState: db.Created.String(),
		}, nil)
		mockRandomizer.EXPECT().GetAvatar("").Return(avatarURL)
		mockStore.EXPECT().UpdateAvatar(ctx, db.UpdateAvatarParams{
			Avatar: avatarURL,
			ID:     playerID,
		}).Return(db.Player{}, xerrors.New("failed to update avatar"))

		_, err := srv.GenerateNewAvatar(ctx, playerID)
		assert.Error(t, err)
	})

	t.Run("Should fail to update avatar because we fail to get all players in room from DB", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := t.Context()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(db.Room{
			ID:        roomID,
			RoomState: db.Created.String(),
		}, nil)
		mockRandomizer.EXPECT().GetAvatar("").Return(avatarURL)
		mockStore.EXPECT().UpdateAvatar(ctx, db.UpdateAvatarParams{
			Avatar: avatarURL,
			ID:     playerID,
		}).Return(db.Player{}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, playerID).Return(
			[]db.GetAllPlayersInRoomRow{}, xerrors.New("failed to get all players in room"),
		).Times(1)

		_, err := srv.GenerateNewAvatar(ctx, playerID)
		assert.Error(t, err)
	})
}

func TestPlayerServiceTogglePlayerIsReady(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		initialIsReady  bool
		expectedIsReady bool
	}{
		{
			name:            "Should successfully toggle player is ready, from not ready -> ready",
			initialIsReady:  false,
			expectedIsReady: true,
		},
		{
			name:            "Should successfully toggle player is ready, from ready -> not ready",
			initialIsReady:  true,
			expectedIsReady: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockStore := mockService.NewMockStorer(t)
			mockRandomizer := mockService.NewMockRandomizer(t)
			srv := service.NewPlayerService(mockStore, mockRandomizer)

			ctx := t.Context()

			mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(db.Room{
				ID:        roomID,
				RoomState: db.Created.String(),
			}, nil)
			mockStore.EXPECT().TogglePlayerIsReady(ctx, playerID).Return(db.Player{}, nil)
			mockStore.EXPECT().GetAllPlayersInRoom(ctx, playerID).Return([]db.GetAllPlayersInRoomRow{
				{
					ID:         playerID,
					Avatar:     avatarURL,
					RoomCode:   roomCode,
					HostPlayer: hostPlayerID,
					Nickname:   "nickname",
					IsReady:    pgtype.Bool{Bool: tt.expectedIsReady, Valid: true},
				},
			}, nil).Times(1)

			lobby, err := srv.TogglePlayerIsReady(ctx, playerID)

			assert.NoError(t, err)
			expectedLobby := service.Lobby{
				Code: roomCode,
				Players: []service.LobbyPlayer{
					{
						ID:       playerID,
						Nickname: "nickname",
						Avatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=randomSeed",
						IsReady:  tt.expectedIsReady,
						IsHost:   false,
					},
				},
			}
			assert.Equal(t, expectedLobby, lobby)
		})
	}

	t.Run("Should fail to toggle player because we fail to get room details from DB", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := t.Context()
		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(
			db.Room{}, xerrors.New("failed to get room details"),
		)

		_, err := srv.TogglePlayerIsReady(ctx, playerID)
		assert.Error(t, err)
	})

	t.Run("Should fail to toggle player because lobby is not in CREATED state", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := t.Context()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(db.Room{
			ID:        roomID,
			RoomState: db.Playing.String(),
		}, nil)

		_, err := srv.TogglePlayerIsReady(ctx, playerID)
		assert.ErrorContains(t, err, "room is not in CREATED state")
	})

	t.Run("Should fail to toggle player because fail to update ready status in DB", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := t.Context()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(db.Room{
			ID:        roomID,
			RoomState: db.Created.String(),
		}, nil)
		mockStore.EXPECT().TogglePlayerIsReady(ctx, playerID).Return(
			db.Player{}, xerrors.New("failed to update is ready"),
		)

		_, err := srv.TogglePlayerIsReady(ctx, playerID)
		assert.Error(t, err)
	})

	t.Run("Should fail to toggle player because fail to get players in room from DB", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := t.Context()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(db.Room{
			ID:        roomID,
			RoomState: db.Created.String(),
		}, nil)
		mockStore.EXPECT().TogglePlayerIsReady(ctx, playerID).Return(db.Player{}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, playerID).Return(
			[]db.GetAllPlayersInRoomRow{}, xerrors.New("failed to get players in room"),
		)

		_, err := srv.TogglePlayerIsReady(ctx, playerID)
		assert.Error(t, err)
	})
}
