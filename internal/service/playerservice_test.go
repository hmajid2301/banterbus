package service_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/hmajid2301/banterbus/internal/service"
	mockService "gitlab.com/hmajid2301/banterbus/internal/service/mocks"
	sqlc "gitlab.com/hmajid2301/banterbus/internal/store/db"
)

const playerID = "33333-9f7a-4392-b523-fd433b3208ea"
const hostPlayer = "44444-9f7a-4392-b523-fd433b3208ea"

func TestPlayerServiceUpdateNickname(t *testing.T) {
	t.Run("Should successfully update nickname", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(sqlc.Room{
			ID:        roomID,
			RoomState: sqlc.ROOMSTATE_CREATED.String(),
		}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, playerID).Return([]sqlc.GetAllPlayersInRoomRow{
			{
				Nickname: "Old Nickname",
			},
		}, nil).Times(1)
		mockStore.EXPECT().UpdateNickname(ctx, sqlc.UpdateNicknameParams{
			Nickname: "New Nickname",
			ID:       playerID,
		}).Return(sqlc.Player{}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, playerID).Return([]sqlc.GetAllPlayersInRoomRow{
			{
				ID:         playerID,
				RoomCode:   roomCode,
				HostPlayer: hostPlayer,
				Nickname:   "New Nickname",
				IsReady:    sql.NullBool{Bool: false, Valid: true},
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
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := context.Background()
		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(
			sqlc.Room{}, fmt.Errorf("failed to get room details"),
		)

		_, err := srv.UpdateNickname(ctx, "New Nickname", playerID)
		assert.Error(t, err)
	})

	t.Run("Should fail to update nickname because lobby not in CREATED room state", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(sqlc.Room{
			ID:        roomID,
			RoomState: sqlc.ROOMSTATE_FINISHED.String(),
		}, nil)

		_, err := srv.UpdateNickname(ctx, "New Nickname", playerID)
		assert.Error(t, err)
	})

	t.Run("Should fail to update nickname because, nickname already exists in lobby", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(sqlc.Room{
			ID:        roomID,
			RoomState: sqlc.ROOMSTATE_CREATED.String(),
		}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, playerID).Return([]sqlc.GetAllPlayersInRoomRow{
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
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(sqlc.Room{
			ID:        roomID,
			RoomState: sqlc.ROOMSTATE_CREATED.String(),
		}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, playerID).Return([]sqlc.GetAllPlayersInRoomRow{
			{
				Nickname: "Old Nickname",
			},
		}, nil).Times(1)
		mockStore.EXPECT().UpdateNickname(ctx, sqlc.UpdateNicknameParams{
			Nickname: "New Nickname",
			ID:       playerID,
		}).Return(sqlc.Player{}, fmt.Errorf("failed to update nickname"))

		_, err := srv.UpdateNickname(ctx, "New Nickname", playerID)
		assert.Error(t, err)
	})

	t.Run("Should fail to update nickname because get all players failed from DB", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(sqlc.Room{
			ID:        roomID,
			RoomState: sqlc.ROOMSTATE_CREATED.String(),
		}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, playerID).Return([]sqlc.GetAllPlayersInRoomRow{
			{
				Nickname: "Old Nickname",
			},
		}, nil).Times(1)
		mockStore.EXPECT().UpdateNickname(ctx, sqlc.UpdateNicknameParams{
			Nickname: "New Nickname",
			ID:       playerID,
		}).Return(sqlc.Player{}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, playerID).Return(
			[]sqlc.GetAllPlayersInRoomRow{}, fmt.Errorf("failed to get all players in room"),
		).Times(1)

		_, err := srv.UpdateNickname(ctx, "New Nickname", playerID)
		assert.Error(t, err)
	})
}

func TestPlayerServiceGenerateAvatar(t *testing.T) {
	t.Run("Should successfully generate avatar", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(sqlc.Room{
			ID:        roomID,
			RoomState: sqlc.ROOMSTATE_CREATED.String(),
		}, nil)
		mockRandomizer.EXPECT().GetAvatar().Return([]byte("avatar1"))
		mockStore.EXPECT().UpdateAvatar(ctx, sqlc.UpdateAvatarParams{
			Avatar: []byte("avatar1"),
			ID:     playerID,
		}).Return(sqlc.Player{}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, playerID).Return([]sqlc.GetAllPlayersInRoomRow{
			{
				ID:         playerID,
				Avatar:     []byte("avatar1"),
				RoomCode:   roomCode,
				HostPlayer: hostPlayer,
				Nickname:   "nickname",
				IsReady:    sql.NullBool{Bool: false, Valid: true},
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
					Avatar:   "avatar1",
					IsReady:  false,
					IsHost:   false,
				},
			},
		}
		assert.Equal(t, expectedLobby, lobby)
	})

	t.Run("Should fail to update avatar because we fail to get room details from DB", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(
			sqlc.Room{}, fmt.Errorf("failed to get room details"),
		)

		_, err := srv.GenerateNewAvatar(ctx, playerID)
		assert.Error(t, err)
	})

	t.Run("Should fail to update avatar because lobby is not in CREATED state", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(sqlc.Room{
			ID:        roomID,
			RoomState: sqlc.ROOMSTATE_PLAYING.String(),
		}, nil)

		_, err := srv.GenerateNewAvatar(ctx, playerID)
		assert.ErrorContains(t, err, "room is not in CREATED state")
	})

	t.Run("Should fail to update avatar because we fail to update avatar in DB", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(sqlc.Room{
			ID:        roomID,
			RoomState: sqlc.ROOMSTATE_CREATED.String(),
		}, nil)
		mockRandomizer.EXPECT().GetAvatar().Return([]byte("avatar1"))
		mockStore.EXPECT().UpdateAvatar(ctx, sqlc.UpdateAvatarParams{
			Avatar: []byte("avatar1"),
			ID:     playerID,
		}).Return(sqlc.Player{}, fmt.Errorf("failed to update avatar"))

		_, err := srv.GenerateNewAvatar(ctx, playerID)
		assert.Error(t, err)
	})

	t.Run("Should fail to update avatar because we fail to get all players in room from DB", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(sqlc.Room{
			ID:        roomID,
			RoomState: sqlc.ROOMSTATE_CREATED.String(),
		}, nil)
		mockRandomizer.EXPECT().GetAvatar().Return([]byte("avatar1"))
		mockStore.EXPECT().UpdateAvatar(ctx, sqlc.UpdateAvatarParams{
			Avatar: []byte("avatar1"),
			ID:     playerID,
		}).Return(sqlc.Player{}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, playerID).Return(
			[]sqlc.GetAllPlayersInRoomRow{}, fmt.Errorf("failed to get all players in room"),
		).Times(1)

		_, err := srv.GenerateNewAvatar(ctx, playerID)
		assert.Error(t, err)
	})
}

func TestPlayerServiceTogglePlayerIsReady(t *testing.T) {
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
			mockStore := mockService.NewMockStorer(t)
			mockRandomizer := mockService.NewMockRandomizer(t)
			srv := service.NewPlayerService(mockStore, mockRandomizer)

			ctx := context.Background()

			mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(sqlc.Room{
				ID:        roomID,
				RoomState: sqlc.ROOMSTATE_CREATED.String(),
			}, nil)
			mockStore.EXPECT().GetPlayerByID(ctx, playerID).Return(sqlc.Player{
				IsReady: sql.NullBool{Bool: tt.initialIsReady, Valid: true},
			}, nil)
			mockStore.EXPECT().UpdateIsReady(ctx, sqlc.UpdateIsReadyParams{
				IsReady: sql.NullBool{Bool: tt.expectedIsReady, Valid: true},
				ID:      playerID,
			}).Return(sqlc.Player{}, nil)
			mockStore.EXPECT().GetAllPlayersInRoom(ctx, playerID).Return([]sqlc.GetAllPlayersInRoomRow{
				{
					ID:         playerID,
					Avatar:     []byte("avatar1"),
					RoomCode:   roomCode,
					HostPlayer: hostPlayer,
					Nickname:   "nickname",
					IsReady:    sql.NullBool{Bool: tt.expectedIsReady, Valid: true},
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
						Avatar:   "avatar1",
						IsReady:  tt.expectedIsReady,
						IsHost:   false,
					},
				},
			}
			assert.Equal(t, expectedLobby, lobby)
		})
	}

	t.Run("Should fail to toggle player because we fail to get room details from DB", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := context.Background()
		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(
			sqlc.Room{}, fmt.Errorf("failed to get room details"),
		)

		_, err := srv.TogglePlayerIsReady(ctx, playerID)
		assert.Error(t, err)
	})

	t.Run("Should fail to toggle player because lobby is not in CREATED state", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(sqlc.Room{
			ID:        roomID,
			RoomState: sqlc.ROOMSTATE_PLAYING.String(),
		}, nil)

		_, err := srv.TogglePlayerIsReady(ctx, playerID)
		assert.ErrorContains(t, err, "room is not in CREATED state")
	})

	t.Run("Should fail to toggle player because fail to get player from DB", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(sqlc.Room{
			ID:        roomID,
			RoomState: sqlc.ROOMSTATE_CREATED.String(),
		}, nil)
		mockStore.EXPECT().GetPlayerByID(ctx, playerID).Return(sqlc.Player{
			IsReady: sql.NullBool{Bool: false, Valid: true},
		}, fmt.Errorf("failed to get player"))

		_, err := srv.TogglePlayerIsReady(ctx, playerID)
		assert.Error(t, err)
	})

	t.Run("Should fail to toggle player because fail to update ready status in DB", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(sqlc.Room{
			ID:        roomID,
			RoomState: sqlc.ROOMSTATE_CREATED.String(),
		}, nil)
		mockStore.EXPECT().GetPlayerByID(ctx, playerID).Return(sqlc.Player{
			IsReady: sql.NullBool{Bool: false, Valid: true},
		}, nil)
		mockStore.EXPECT().UpdateIsReady(ctx, sqlc.UpdateIsReadyParams{
			IsReady: sql.NullBool{Bool: true, Valid: true},
			ID:      playerID,
		}).Return(sqlc.Player{}, fmt.Errorf("failed to update is ready"))

		_, err := srv.TogglePlayerIsReady(ctx, playerID)
		assert.Error(t, err)
	})

	t.Run("Should fail to toggle player because fail to get players in room from DB", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(sqlc.Room{
			ID:        roomID,
			RoomState: sqlc.ROOMSTATE_CREATED.String(),
		}, nil)
		mockStore.EXPECT().GetPlayerByID(ctx, playerID).Return(sqlc.Player{
			IsReady: sql.NullBool{Bool: false, Valid: true},
		}, nil)
		mockStore.EXPECT().UpdateIsReady(ctx, sqlc.UpdateIsReadyParams{
			IsReady: sql.NullBool{Bool: true, Valid: true},
			ID:      playerID,
		}).Return(sqlc.Player{}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, playerID).Return(
			[]sqlc.GetAllPlayersInRoomRow{}, fmt.Errorf("failed to get players in room"),
		)

		_, err := srv.TogglePlayerIsReady(ctx, playerID)
		assert.Error(t, err)
	})
}

func TestPlayerServiceGetRoomState(t *testing.T) {
	tests := []struct {
		name          string
		roomState     sqlc.RoomState
		expectedState sqlc.RoomState
	}{
		{
			name:          "Should successfully get room state CREATED",
			roomState:     sqlc.ROOMSTATE_CREATED,
			expectedState: sqlc.ROOMSTATE_CREATED,
		},
		{
			name:          "Should successfully get room state PLAYING",
			roomState:     sqlc.ROOMSTATE_PLAYING,
			expectedState: sqlc.ROOMSTATE_PLAYING,
		},
		{
			name:          "Should successfully get room state PAUSED",
			roomState:     sqlc.ROOMSTATE_PAUSED,
			expectedState: sqlc.ROOMSTATE_PAUSED,
		},
		{
			name:          "Should successfully get room state FINISHED",
			roomState:     sqlc.ROOMSTATE_FINISHED,
			expectedState: sqlc.ROOMSTATE_FINISHED,
		},
		{
			name:          "Should successfully get room state ABANDONED",
			roomState:     sqlc.ROOMSTATE_ABANDONED,
			expectedState: sqlc.ROOMSTATE_ABANDONED,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := mockService.NewMockStorer(t)
			mockRandomizer := mockService.NewMockRandomizer(t)
			srv := service.NewPlayerService(mockStore, mockRandomizer)

			ctx := context.Background()
			mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(sqlc.Room{
				RoomState: tt.roomState.String(),
			}, nil)

			roomState, err := srv.GetRoomState(ctx, playerID)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedState, roomState)
		})
	}

	t.Run("Should fail to get room state because we fail to get room details DB", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := context.Background()
		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(
			sqlc.Room{}, fmt.Errorf("failed to get room details"),
		)

		_, err := srv.GetRoomState(ctx, playerID)
		assert.Error(t, err)
	})
}

func TestPlayerServiceGetLobby(t *testing.T) {
	t.Run("Should successfully get lobby", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetAllPlayersInRoom(ctx, playerID).Return([]sqlc.GetAllPlayersInRoomRow{
			{
				ID:         playerID,
				RoomCode:   roomCode,
				HostPlayer: hostPlayer,
				Nickname:   "nickname",
				IsReady:    sql.NullBool{Bool: false, Valid: true},
			},
		}, nil)
		lobby, err := srv.GetLobby(ctx, playerID)

		assert.NoError(t, err)
		expectedLobby := service.Lobby{
			Code: roomCode,
			Players: []service.LobbyPlayer{
				{
					ID:       playerID,
					Nickname: "nickname",
					IsReady:  false,
					IsHost:   false,
				},
			},
		}
		assert.Equal(t, expectedLobby, lobby)
	})

	t.Run("Should fail to get lobby because cannot get details from DB", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetAllPlayersInRoom(ctx, playerID).Return(
			[]sqlc.GetAllPlayersInRoomRow{}, fmt.Errorf("failed to get players in room"),
		)

		_, err := srv.GetLobby(ctx, playerID)
		assert.Error(t, err)
	})
}

func TestPlayerServiceGetGameState(t *testing.T) {
	tests := []struct {
		name          string
		gameState     sqlc.GameStateEnum
		expectedState sqlc.GameStateEnum
	}{
		{
			name:          "Should successfully get game state QUESTION",
			gameState:     sqlc.GAMESTATE_FIBBING_IT_SHOW_QUESTION,
			expectedState: sqlc.GAMESTATE_FIBBING_IT_SHOW_QUESTION,
		},
		{
			name:          "Should successfully get game state VOTING",
			gameState:     sqlc.GAMESTATE_FIBBING_IT_VOTING,
			expectedState: sqlc.GAMESTATE_FIBBING_IT_VOTING,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := mockService.NewMockStorer(t)
			mockRandomizer := mockService.NewMockRandomizer(t)
			srv := service.NewPlayerService(mockStore, mockRandomizer)

			ctx := context.Background()
			mockStore.EXPECT().GetGameStateByPlayerID(ctx, playerID).Return(sqlc.GameState{
				State: tt.gameState.String(),
			}, nil)

			gameState, err := srv.GetGameState(ctx, playerID)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedState, gameState)
		})
	}

	t.Run("Should fail to get game state because we fail to get game details DB", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := context.Background()
		mockStore.EXPECT().GetGameStateByPlayerID(ctx, playerID).Return(
			sqlc.GameState{}, fmt.Errorf("failed to get game state details"),
		)

		_, err := srv.GetGameState(ctx, playerID)
		assert.Error(t, err)
	})
}

func TestPlayerServiceGetQuestionState(t *testing.T) {
	t.Run("Should successfully get question state", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetCurrentQuestionByPlayerID(ctx, playerID).Return(sqlc.GetCurrentQuestionByPlayerIDRow{
			ID:             playerID,
			Avatar:         []byte(""),
			Nickname:       "nickname",
			PlayerRole:     sql.NullString{String: "fibber"},
			FibberQuestion: sql.NullString{String: "fibber question"},
			NormalQuestion: sql.NullString{String: ""},
			Round:          1,
			RoundType:      "free_form",
			RoomCode:       "ABC12",
		}, nil)

		questionState, err := srv.GetQuestionState(ctx, playerID)

		assert.NoError(t, err)
		expectedGameState := service.QuestionState{
			Players: []service.PlayerWithRole{
				{
					ID:       playerID,
					Nickname: "nickname",
					Role:     "fibber",
					Avatar:   []byte(""),
					Question: "",
				},
			},
			Round:     1,
			RoundType: "free_form",
			RoomCode:  "ABC12",
		}
		assert.Equal(t, expectedGameState, questionState)
	})

	t.Run("Should successfully get question state, as normal fibber", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetCurrentQuestionByPlayerID(ctx, playerID).Return(sqlc.GetCurrentQuestionByPlayerIDRow{
			ID:             playerID,
			Avatar:         []byte(""),
			Nickname:       "nickname",
			PlayerRole:     sql.NullString{String: "normal"},
			FibberQuestion: sql.NullString{String: ""},
			NormalQuestion: sql.NullString{String: "normal question"},
			Round:          1,
			RoundType:      "free_form",
			RoomCode:       "ABC12",
		}, nil)

		questionState, err := srv.GetQuestionState(ctx, playerID)

		assert.NoError(t, err)
		expectedGameState := service.QuestionState{
			Players: []service.PlayerWithRole{
				{
					ID:       playerID,
					Nickname: "nickname",
					Role:     "normal",
					Avatar:   []byte(""),
					Question: "normal question",
				},
			},
			Round:     1,
			RoundType: "free_form",
			RoomCode:  "ABC12",
		}
		assert.Equal(t, expectedGameState, questionState)
	})

	t.Run("Should fail to get question state because we cannot fetch from DB", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetCurrentQuestionByPlayerID(ctx, playerID).Return(
			sqlc.GetCurrentQuestionByPlayerIDRow{}, fmt.Errorf("failed to get questions"),
		)

		_, err := srv.GetQuestionState(ctx, playerID)
		assert.Error(t, err)
	})
}

func TestPlayerServiceGetVotingState(t *testing.T) {
	roundID := "77777-fbb75599-9f7a-4392-b523-fd433b3208ea"

	t.Run("Should successfully get voting state", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetLatestRoundByPlayerID(ctx, playerID).Return(
			sqlc.GetLatestRoundByPlayerIDRow{
				ID: roundID,
			}, nil)
		mockStore.EXPECT().CountVotesByRoundID(ctx, roundID).Return(
			[]sqlc.CountVotesByRoundIDRow{
				{
					VotedForPlayerID: playerID,
					Nickname:         "nickname",
					VoteCount:        1,
					Avatar:           []byte(""),
				},
			}, nil)

		votingState, err := srv.GetVotingState(ctx, playerID)

		assert.NoError(t, err)
		expectedVotingState := []service.VotingPlayer{
			{
				ID:       playerID,
				Nickname: "nickname",
				Avatar:   "",
				Votes:    1,
			},
		}
		assert.Equal(t, expectedVotingState, votingState)
	})

	t.Run("Should fail to get voting state because fail to get round info from DB", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetLatestRoundByPlayerID(ctx, playerID).Return(
			sqlc.GetLatestRoundByPlayerIDRow{}, fmt.Errorf("failed to get round info"),
		)
		_, err := srv.GetVotingState(ctx, playerID)
		assert.Error(t, err)
	})

	t.Run("Should fail to get voting state because fail to get votes from DB", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetLatestRoundByPlayerID(ctx, playerID).Return(
			sqlc.GetLatestRoundByPlayerIDRow{
				ID: roundID,
			}, nil)
		mockStore.EXPECT().CountVotesByRoundID(ctx, roundID).Return(
			[]sqlc.CountVotesByRoundIDRow{}, fmt.Errorf("failed to get votes"),
		)

		_, err := srv.GetVotingState(ctx, playerID)

		assert.Error(t, err)
	})
}
