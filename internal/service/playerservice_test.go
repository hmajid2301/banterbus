package service_test

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"

	"gitlab.com/hmajid2301/banterbus/internal/service"
	mockService "gitlab.com/hmajid2301/banterbus/internal/service/mocks"
	"gitlab.com/hmajid2301/banterbus/internal/store/db"
)

// Import constants from lobbyservice_test.go to maintain consistency

var avatarURL = "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=randomSeed"

func TestPlayerServiceUpdateNickname(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully update nickname", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockPlayerStore(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := t.Context()

		mockStore.EXPECT().UpdateNicknameWithPlayers(ctx, db.UpdateNicknameArgs{
			PlayerID: playerID,
			Nickname: "New Nickname",
		}).Return(db.UpdateNicknameResult{
			Players: []db.GetAllPlayersInRoomRow{
				{
					ID:         playerID,
					RoomCode:   roomCode,
					HostPlayer: hostPlayerID,
					Nickname:   "New Nickname",
					IsReady:    pgtype.Bool{Bool: false, Valid: true},
				},
			},
		}, nil)

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
		mockStore := mockService.NewMockPlayerStore(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := t.Context()
		mockStore.EXPECT().UpdateNicknameWithPlayers(ctx, db.UpdateNicknameArgs{
			PlayerID: playerID,
			Nickname: "New Nickname",
		}).Return(db.UpdateNicknameResult{}, errors.New("failed to get room details"))

		_, err := srv.UpdateNickname(ctx, "New Nickname", playerID)
		assert.Error(t, err)
	})

	t.Run("Should fail to update nickname because lobby not in CREATED room state", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockPlayerStore(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := t.Context()

		mockStore.EXPECT().UpdateNicknameWithPlayers(ctx, db.UpdateNicknameArgs{
			PlayerID: playerID,
			Nickname: "New Nickname",
		}).Return(db.UpdateNicknameResult{}, errors.New("room is not in CREATED state"))

		_, err := srv.UpdateNickname(ctx, "New Nickname", playerID)
		assert.Error(t, err)
	})

	t.Run("Should fail to update nickname because, nickname already exists in lobby", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockPlayerStore(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := t.Context()

		mockStore.EXPECT().UpdateNicknameWithPlayers(ctx, db.UpdateNicknameArgs{
			PlayerID: playerID,
			Nickname: "New Nickname",
		}).Return(db.UpdateNicknameResult{}, errors.New("nickname already exists"))

		_, err := srv.UpdateNickname(ctx, "New Nickname", playerID)
		assert.ErrorContains(t, err, "nickname already exists")
	})

	t.Run("Should fail to update nickname because update nickname in DB failed", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockPlayerStore(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := t.Context()

		mockStore.EXPECT().UpdateNicknameWithPlayers(ctx, db.UpdateNicknameArgs{
			PlayerID: playerID,
			Nickname: "New Nickname",
		}).Return(db.UpdateNicknameResult{}, errors.New("failed to update nickname"))

		_, err := srv.UpdateNickname(ctx, "New Nickname", playerID)
		assert.Error(t, err)
	})

	t.Run("Should fail to update nickname because get all players failed from DB", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockPlayerStore(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := t.Context()

		mockStore.EXPECT().UpdateNicknameWithPlayers(ctx, db.UpdateNicknameArgs{
			PlayerID: playerID,
			Nickname: "New Nickname",
		}).Return(db.UpdateNicknameResult{}, errors.New("failed to get all players in room"))

		_, err := srv.UpdateNickname(ctx, "New Nickname", playerID)
		assert.Error(t, err)
	})
}

func TestPlayerServiceGenerateAvatar(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully generate avatar", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockPlayerStore(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := t.Context()

		mockRandomizer.EXPECT().GetAvatar("").Return(avatarURL)
		mockStore.EXPECT().GenerateNewAvatarWithPlayers(ctx, db.GenerateNewAvatarArgs{
			PlayerID: playerID,
			Avatar:   avatarURL,
		}).Return(db.GenerateNewAvatarResult{
			Players: []db.GetAllPlayersInRoomRow{
				{
					ID:         playerID,
					Avatar:     avatarURL,
					RoomCode:   roomCode,
					HostPlayer: hostPlayerID,
					Nickname:   "nickname",
					IsReady:    pgtype.Bool{Bool: false, Valid: true},
				},
			},
		}, nil)

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
		mockStore := mockService.NewMockPlayerStore(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := t.Context()

		mockRandomizer.EXPECT().GetAvatar("").Return(avatarURL)
		mockStore.EXPECT().GenerateNewAvatarWithPlayers(ctx, db.GenerateNewAvatarArgs{
			PlayerID: playerID,
			Avatar:   avatarURL,
		}).Return(db.GenerateNewAvatarResult{}, errors.New("failed to get room details"))

		_, err := srv.GenerateNewAvatar(ctx, playerID)
		assert.Error(t, err)
	})

	t.Run("Should fail to update avatar because lobby is not in CREATED state", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockPlayerStore(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := t.Context()

		mockRandomizer.EXPECT().GetAvatar("").Return(avatarURL)
		mockStore.EXPECT().GenerateNewAvatarWithPlayers(ctx, db.GenerateNewAvatarArgs{
			PlayerID: playerID,
			Avatar:   avatarURL,
		}).Return(db.GenerateNewAvatarResult{}, errors.New("room is not in CREATED state"))

		_, err := srv.GenerateNewAvatar(ctx, playerID)
		assert.ErrorContains(t, err, "room is not in CREATED state")
	})

	t.Run("Should fail to update avatar because we fail to update avatar in DB", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockPlayerStore(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := t.Context()

		mockRandomizer.EXPECT().GetAvatar("").Return(avatarURL)
		mockStore.EXPECT().GenerateNewAvatarWithPlayers(ctx, db.GenerateNewAvatarArgs{
			PlayerID: playerID,
			Avatar:   avatarURL,
		}).Return(db.GenerateNewAvatarResult{}, errors.New("failed to update avatar"))

		_, err := srv.GenerateNewAvatar(ctx, playerID)
		assert.Error(t, err)
	})

	t.Run("Should fail to update avatar because we fail to get all players in room from DB", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockPlayerStore(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := t.Context()

		mockRandomizer.EXPECT().GetAvatar("").Return(avatarURL)
		mockStore.EXPECT().GenerateNewAvatarWithPlayers(ctx, db.GenerateNewAvatarArgs{
			PlayerID: playerID,
			Avatar:   avatarURL,
		}).Return(db.GenerateNewAvatarResult{}, errors.New("failed to get all players in room"))

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
			mockStore := mockService.NewMockPlayerStore(t)
			mockRandomizer := mockService.NewMockRandomizer(t)
			srv := service.NewPlayerService(mockStore, mockRandomizer)

			ctx := t.Context()

			mockStore.EXPECT().TogglePlayerReadyWithPlayers(ctx, db.TogglePlayerIsReadyArgs{
				PlayerID: playerID,
			}).Return(db.TogglePlayerIsReadyResult{
				Players: []db.GetAllPlayersInRoomRow{
					{
						ID:         playerID,
						Avatar:     avatarURL,
						RoomCode:   roomCode,
						HostPlayer: hostPlayerID,
						Nickname:   "nickname",
						IsReady:    pgtype.Bool{Bool: tt.expectedIsReady, Valid: true},
					},
				},
			}, nil)

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
		mockStore := mockService.NewMockPlayerStore(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := t.Context()
		mockStore.EXPECT().TogglePlayerReadyWithPlayers(ctx, db.TogglePlayerIsReadyArgs{
			PlayerID: playerID,
		}).Return(db.TogglePlayerIsReadyResult{}, errors.New("failed to get room details"))

		_, err := srv.TogglePlayerIsReady(ctx, playerID)
		assert.Error(t, err)
	})

	t.Run("Should fail to toggle player because lobby is not in CREATED state", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockPlayerStore(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := t.Context()

		mockStore.EXPECT().TogglePlayerReadyWithPlayers(ctx, db.TogglePlayerIsReadyArgs{
			PlayerID: playerID,
		}).Return(db.TogglePlayerIsReadyResult{}, errors.New("room is not in CREATED state"))

		_, err := srv.TogglePlayerIsReady(ctx, playerID)
		assert.ErrorContains(t, err, "room is not in CREATED state")
	})

	t.Run("Should fail to toggle player because fail to update ready status in DB", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockPlayerStore(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := t.Context()

		mockStore.EXPECT().TogglePlayerReadyWithPlayers(ctx, db.TogglePlayerIsReadyArgs{
			PlayerID: playerID,
		}).Return(db.TogglePlayerIsReadyResult{}, errors.New("failed to update is ready"))

		_, err := srv.TogglePlayerIsReady(ctx, playerID)
		assert.Error(t, err)
	})

	t.Run("Should fail to toggle player because fail to get players in room from DB", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockPlayerStore(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := t.Context()

		mockStore.EXPECT().TogglePlayerReadyWithPlayers(ctx, db.TogglePlayerIsReadyArgs{
			PlayerID: playerID,
		}).Return(db.TogglePlayerIsReadyResult{}, errors.New("failed to get players in room"))

		_, err := srv.TogglePlayerIsReady(ctx, playerID)
		assert.Error(t, err)
	})
}

func TestPlayerServiceUpdateLocale(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully update locale", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockPlayerStore(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := t.Context()
		newLocale := "fr-FR"

		mockStore.EXPECT().UpdateLocale(ctx, db.UpdateLocaleParams{
			ID:     playerID,
			Locale: pgtype.Text{String: newLocale},
		}).Return(db.Player{}, nil)

		err := srv.UpdateLocale(ctx, playerID, newLocale)
		assert.NoError(t, err)
	})

	t.Run("Should handle player not found gracefully", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockPlayerStore(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := t.Context()
		newLocale := "fr-FR"

		mockStore.EXPECT().UpdateLocale(ctx, db.UpdateLocaleParams{
			ID:     playerID,
			Locale: pgtype.Text{String: newLocale},
		}).Return(db.Player{}, errors.New("no rows in result set"))

		err := srv.UpdateLocale(ctx, playerID, newLocale)
		assert.ErrorIs(t, err, service.ErrPlayerNotFound)
	})

	t.Run("Should return error for other database errors", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockPlayerStore(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := t.Context()
		newLocale := "fr-FR"

		mockStore.EXPECT().UpdateLocale(ctx, db.UpdateLocaleParams{
			ID:     playerID,
			Locale: pgtype.Text{String: newLocale},
		}).Return(db.Player{}, errors.New("database connection error"))

		err := srv.UpdateLocale(ctx, playerID, newLocale)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database connection error")
	})
}
