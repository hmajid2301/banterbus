package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"

	"gitlab.com/hmajid2301/banterbus/internal/service"
	mockService "gitlab.com/hmajid2301/banterbus/internal/service/mocks"
	"gitlab.com/hmajid2301/banterbus/internal/store/db"
)

var playerID = uuid.MustParse("0193a625-dad1-7095-9abb-bebdad739381")

func TestPlayerServiceUpdateNickname(t *testing.T) {
	t.Run("Should successfully update nickname", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(db.Room{
			ID:        roomID,
			RoomState: db.ROOMSTATE_CREATED.String(),
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
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := context.Background()
		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(
			db.Room{}, fmt.Errorf("failed to get room details"),
		)

		_, err := srv.UpdateNickname(ctx, "New Nickname", playerID)
		assert.Error(t, err)
	})

	t.Run("Should fail to update nickname because lobby not in CREATED room state", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(db.Room{
			ID:        roomID,
			RoomState: db.ROOMSTATE_FINISHED.String(),
		}, nil)

		_, err := srv.UpdateNickname(ctx, "New Nickname", playerID)
		assert.Error(t, err)
	})

	t.Run("Should fail to update nickname because, nickname already exists in lobby", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(db.Room{
			ID:        roomID,
			RoomState: db.ROOMSTATE_CREATED.String(),
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
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(db.Room{
			ID:        roomID,
			RoomState: db.ROOMSTATE_CREATED.String(),
		}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, playerID).Return([]db.GetAllPlayersInRoomRow{
			{
				Nickname: "Old Nickname",
			},
		}, nil).Times(1)
		mockStore.EXPECT().UpdateNickname(ctx, db.UpdateNicknameParams{
			Nickname: "New Nickname",
			ID:       playerID,
		}).Return(db.Player{}, fmt.Errorf("failed to update nickname"))

		_, err := srv.UpdateNickname(ctx, "New Nickname", playerID)
		assert.Error(t, err)
	})

	t.Run("Should fail to update nickname because get all players failed from DB", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(db.Room{
			ID:        roomID,
			RoomState: db.ROOMSTATE_CREATED.String(),
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
			[]db.GetAllPlayersInRoomRow{}, fmt.Errorf("failed to get all players in room"),
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

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(db.Room{
			ID:        roomID,
			RoomState: db.ROOMSTATE_CREATED.String(),
		}, nil)
		mockRandomizer.EXPECT().GetAvatar().Return([]byte("avatar1"))
		mockStore.EXPECT().UpdateAvatar(ctx, db.UpdateAvatarParams{
			Avatar: []byte("avatar1"),
			ID:     playerID,
		}).Return(db.Player{}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, playerID).Return([]db.GetAllPlayersInRoomRow{
			{
				ID:         playerID,
				Avatar:     []byte("avatar1"),
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
			db.Room{}, fmt.Errorf("failed to get room details"),
		)

		_, err := srv.GenerateNewAvatar(ctx, playerID)
		assert.Error(t, err)
	})

	t.Run("Should fail to update avatar because lobby is not in CREATED state", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(db.Room{
			ID:        roomID,
			RoomState: db.ROOMSTATE_PLAYING.String(),
		}, nil)

		_, err := srv.GenerateNewAvatar(ctx, playerID)
		assert.ErrorContains(t, err, "room is not in CREATED state")
	})

	t.Run("Should fail to update avatar because we fail to update avatar in DB", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(db.Room{
			ID:        roomID,
			RoomState: db.ROOMSTATE_CREATED.String(),
		}, nil)
		mockRandomizer.EXPECT().GetAvatar().Return([]byte("avatar1"))
		mockStore.EXPECT().UpdateAvatar(ctx, db.UpdateAvatarParams{
			Avatar: []byte("avatar1"),
			ID:     playerID,
		}).Return(db.Player{}, fmt.Errorf("failed to update avatar"))

		_, err := srv.GenerateNewAvatar(ctx, playerID)
		assert.Error(t, err)
	})

	t.Run("Should fail to update avatar because we fail to get all players in room from DB", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(db.Room{
			ID:        roomID,
			RoomState: db.ROOMSTATE_CREATED.String(),
		}, nil)
		mockRandomizer.EXPECT().GetAvatar().Return([]byte("avatar1"))
		mockStore.EXPECT().UpdateAvatar(ctx, db.UpdateAvatarParams{
			Avatar: []byte("avatar1"),
			ID:     playerID,
		}).Return(db.Player{}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, playerID).Return(
			[]db.GetAllPlayersInRoomRow{}, fmt.Errorf("failed to get all players in room"),
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

			mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(db.Room{
				ID:        roomID,
				RoomState: db.ROOMSTATE_CREATED.String(),
			}, nil)
			mockStore.EXPECT().TogglePlayerIsReady(ctx, playerID).Return(db.Player{}, nil)
			mockStore.EXPECT().GetAllPlayersInRoom(ctx, playerID).Return([]db.GetAllPlayersInRoomRow{
				{
					ID:         playerID,
					Avatar:     []byte("avatar1"),
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
			db.Room{}, fmt.Errorf("failed to get room details"),
		)

		_, err := srv.TogglePlayerIsReady(ctx, playerID)
		assert.Error(t, err)
	})

	t.Run("Should fail to toggle player because lobby is not in CREATED state", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(db.Room{
			ID:        roomID,
			RoomState: db.ROOMSTATE_PLAYING.String(),
		}, nil)

		_, err := srv.TogglePlayerIsReady(ctx, playerID)
		assert.ErrorContains(t, err, "room is not in CREATED state")
	})

	t.Run("Should fail to toggle player because fail to update ready status in DB", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(db.Room{
			ID:        roomID,
			RoomState: db.ROOMSTATE_CREATED.String(),
		}, nil)
		mockStore.EXPECT().TogglePlayerIsReady(ctx, playerID).Return(
			db.Player{}, fmt.Errorf("failed to update is ready"),
		)

		_, err := srv.TogglePlayerIsReady(ctx, playerID)
		assert.Error(t, err)
	})

	t.Run("Should fail to toggle player because fail to get players in room from DB", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(db.Room{
			ID:        roomID,
			RoomState: db.ROOMSTATE_CREATED.String(),
		}, nil)
		mockStore.EXPECT().TogglePlayerIsReady(ctx, playerID).Return(db.Player{}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, playerID).Return(
			[]db.GetAllPlayersInRoomRow{}, fmt.Errorf("failed to get players in room"),
		)

		_, err := srv.TogglePlayerIsReady(ctx, playerID)
		assert.Error(t, err)
	})
}

func TestPlayerServiceGetRoomState(t *testing.T) {
	tests := []struct {
		name          string
		roomState     db.RoomState
		expectedState db.RoomState
	}{
		{
			name:          "Should successfully get room state CREATED",
			roomState:     db.ROOMSTATE_CREATED,
			expectedState: db.ROOMSTATE_CREATED,
		},
		{
			name:          "Should successfully get room state PLAYING",
			roomState:     db.ROOMSTATE_PLAYING,
			expectedState: db.ROOMSTATE_PLAYING,
		},
		{
			name:          "Should successfully get room state PAUSED",
			roomState:     db.ROOMSTATE_PAUSED,
			expectedState: db.ROOMSTATE_PAUSED,
		},
		{
			name:          "Should successfully get room state FINISHED",
			roomState:     db.ROOMSTATE_FINISHED,
			expectedState: db.ROOMSTATE_FINISHED,
		},
		{
			name:          "Should successfully get room state ABANDONED",
			roomState:     db.ROOMSTATE_ABANDONED,
			expectedState: db.ROOMSTATE_ABANDONED,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := mockService.NewMockStorer(t)
			mockRandomizer := mockService.NewMockRandomizer(t)
			srv := service.NewPlayerService(mockStore, mockRandomizer)

			ctx := context.Background()
			mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(db.Room{
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
			db.Room{}, fmt.Errorf("failed to get room details"),
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

		mockStore.EXPECT().GetAllPlayersInRoom(ctx, playerID).Return([]db.GetAllPlayersInRoomRow{
			{
				ID:         playerID,
				RoomCode:   roomCode,
				HostPlayer: hostPlayerID,
				Nickname:   "nickname",
				IsReady:    pgtype.Bool{Bool: false, Valid: true},
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
			[]db.GetAllPlayersInRoomRow{}, fmt.Errorf("failed to get players in room"),
		)

		_, err := srv.GetLobby(ctx, playerID)
		assert.Error(t, err)
	})
}

func TestPlayerServiceGetGameState(t *testing.T) {
	tests := []struct {
		name          string
		gameState     db.GameStateEnum
		expectedState db.GameStateEnum
	}{
		{
			name:          "Should successfully get game state QUESTION",
			gameState:     db.GAMESTATE_FIBBING_IT_QUESTION,
			expectedState: db.GAMESTATE_FIBBING_IT_QUESTION,
		},
		{
			name:          "Should successfully get game state VOTING",
			gameState:     db.GAMESTATE_FIBBING_IT_VOTING,
			expectedState: db.GAMESTATE_FIBBING_IT_VOTING,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := mockService.NewMockStorer(t)
			mockRandomizer := mockService.NewMockRandomizer(t)
			srv := service.NewPlayerService(mockStore, mockRandomizer)

			ctx := context.Background()
			mockStore.EXPECT().GetGameStateByPlayerID(ctx, playerID).Return(db.GameState{
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
			db.GameState{}, fmt.Errorf("failed to get game state details"),
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

		deadline := time.Now().Add(5 * time.Second)
		mockStore.EXPECT().GetCurrentQuestionByPlayerID(ctx, playerID).Return(db.GetCurrentQuestionByPlayerIDRow{
			PlayerID:       playerID,
			Avatar:         []byte(""),
			Nickname:       "nickname",
			Role:           pgtype.Text{String: "fibber"},
			Question:       "fibber question",
			Round:          1,
			RoundType:      "free_form",
			RoomCode:       "ABC12",
			SubmitDeadline: pgtype.Timestamp{Time: deadline},
		}, nil)

		questionState, err := srv.GetQuestionState(ctx, playerID)

		assert.NoError(t, err)
		expectedGameState := service.QuestionState{
			Players: []service.PlayerWithRole{
				{
					ID:       playerID,
					Role:     "fibber",
					Question: "fibber question",
				},
			},
			Round:     1,
			RoundType: "free_form",
		}

		diffOpts := cmpopts.IgnoreFields(questionState, "Deadline")
		PartialEqual(t, expectedGameState, questionState, diffOpts)
		assert.LessOrEqual(t, int(questionState.Deadline.Seconds()), 5)
	})

	t.Run("Should successfully get question state, as normal fibber", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := context.Background()

		deadline := time.Now().Add(5 * time.Second)
		mockStore.EXPECT().GetCurrentQuestionByPlayerID(ctx, playerID).Return(db.GetCurrentQuestionByPlayerIDRow{
			PlayerID:       playerID,
			Avatar:         []byte(""),
			Nickname:       "nickname",
			Role:           pgtype.Text{String: "normal"},
			Question:       "normal question",
			Round:          1,
			RoundType:      "free_form",
			RoomCode:       "ABC12",
			SubmitDeadline: pgtype.Timestamp{Time: deadline},
		}, nil)

		questionState, err := srv.GetQuestionState(ctx, playerID)

		assert.NoError(t, err)
		expectedGameState := service.QuestionState{
			Players: []service.PlayerWithRole{
				{
					ID:       playerID,
					Role:     "normal",
					Question: "normal question",
				},
			},
			Round:     1,
			RoundType: "free_form",
		}

		diffOpts := cmpopts.IgnoreFields(questionState, "Deadline")
		PartialEqual(t, expectedGameState, questionState, diffOpts)
		assert.LessOrEqual(t, int(questionState.Deadline.Seconds()), 5)
	})

	t.Run("Should fail to get question state because we cannot fetch from DB", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewPlayerService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetCurrentQuestionByPlayerID(ctx, playerID).Return(
			db.GetCurrentQuestionByPlayerIDRow{}, fmt.Errorf("failed to get questions"),
		)

		_, err := srv.GetQuestionState(ctx, playerID)
		assert.Error(t, err)
	})
}
