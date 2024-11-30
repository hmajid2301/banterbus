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

const roomCode = "ABC12"
const roomID = "fbb75599-9f7a-4392-b523-fd433b3208ea"
const hostPlayerID = "33333-9f7a-4392-b523-fd433b3208ea"

func TestLobbyServiceCreate(t *testing.T) {
	defaultNewHostPlayer := service.NewHostPlayer{
		ID: "11111-9f7a-4392-b523-fd433b3208ea",
	}

	defaultNewPlayer := service.NewPlayer{
		ID:       defaultNewHostPlayer.ID,
		Nickname: "Majiy00",
		Avatar:   []byte(""),
	}

	t.Run("Should create room successfully", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom)

		ctx := context.Background()
		mockRandom.EXPECT().GetNickname().Return(defaultNewPlayer.Nickname)
		mockRandom.EXPECT().GetAvatar().Return(defaultNewPlayer.Avatar)
		mockRandom.EXPECT().GetRoomCode().Return(roomCode)
		mockStore.EXPECT().GetRoomByCode(ctx, roomCode).Return(sqlc.Room{}, sql.ErrNoRows)
		mockRandom.EXPECT().GetID().Return(roomID)

		createRoom := getCreateRoomParams(defaultNewPlayer, defaultNewHostPlayer)
		mockStore.EXPECT().CreateRoom(ctx, createRoom).Return(nil)
		lobby, err := srv.Create(ctx, "fibbing_it", defaultNewHostPlayer)

		expectedLobby := service.Lobby{
			Code: roomCode,
			Players: []service.LobbyPlayer{
				{
					ID:       defaultNewPlayer.ID,
					Nickname: defaultNewPlayer.Nickname,
					Avatar:   string(defaultNewPlayer.Avatar),
					IsHost:   true,
					IsReady:  false,
				},
			},
		}

		assert.NoError(t, err)
		assert.Equal(t, expectedLobby, lobby)
	})

	t.Run("Should create room when nickname is passed", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom)

		newHostPlayer := service.NewHostPlayer{
			Nickname: "MyNickname",
			ID:       "11111-9f7a-4392-b523-fd433b3208ea",
		}

		newPlayer := service.NewPlayer{
			ID:       newHostPlayer.ID,
			Nickname: newHostPlayer.Nickname,
			Avatar:   []byte(""),
		}

		ctx := context.Background()
		mockRandom.EXPECT().GetAvatar().Return(newPlayer.Avatar)
		mockRandom.EXPECT().GetRoomCode().Return(roomCode)
		mockStore.EXPECT().GetRoomByCode(ctx, roomCode).Return(sqlc.Room{}, sql.ErrNoRows)
		mockRandom.EXPECT().GetID().Return(roomID)

		createRoom := getCreateRoomParams(newPlayer, newHostPlayer)
		mockStore.EXPECT().CreateRoom(ctx, createRoom).Return(nil)
		lobby, err := srv.Create(ctx, "fibbing_it", newHostPlayer)

		expectedLobby := service.Lobby{
			Code: roomCode,
			Players: []service.LobbyPlayer{
				{
					ID:       newPlayer.ID,
					Nickname: newPlayer.Nickname,
					Avatar:   string(newPlayer.Avatar),
					IsHost:   true,
					IsReady:  false,
				},
			},
		}

		assert.NoError(t, err)
		assert.Equal(t, expectedLobby, lobby)
	})

	t.Run("Should create room successfully, code is used by room", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom)

		ctx := context.Background()
		mockRandom.EXPECT().GetNickname().Return(defaultNewPlayer.Nickname)
		mockRandom.EXPECT().GetAvatar().Return(defaultNewPlayer.Avatar)
		mockRandom.EXPECT().GetRoomCode().Return(roomCode)
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(sqlc.Room{RoomState: sqlc.ROOMSTATE_CREATED.String()}, nil).
			Times(1)
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(sqlc.Room{RoomState: sqlc.ROOMSTATE_ABANDONED.String()}, nil).
			Times(1)
		mockRandom.EXPECT().GetID().Return(roomID)

		createRoom := getCreateRoomParams(defaultNewPlayer, defaultNewHostPlayer)
		mockStore.EXPECT().CreateRoom(ctx, createRoom).Return(nil)
		lobby, err := srv.Create(ctx, "fibbing_it", defaultNewHostPlayer)

		expectedLobby := service.Lobby{
			Code: roomCode,
			Players: []service.LobbyPlayer{
				{
					ID:       defaultNewPlayer.ID,
					Nickname: defaultNewPlayer.Nickname,
					Avatar:   string(defaultNewPlayer.Avatar),
					IsHost:   true,
					IsReady:  false,
				},
			},
		}

		assert.NoError(t, err)
		assert.Equal(t, expectedLobby, lobby)
	})

	t.Run("Should throw error when we fail to get room code", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom)

		ctx := context.Background()
		mockRandom.EXPECT().GetNickname().Return(defaultNewPlayer.Nickname)
		mockRandom.EXPECT().GetAvatar().Return(defaultNewPlayer.Avatar)
		mockRandom.EXPECT().GetRoomCode().Return(roomCode)
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(sqlc.Room{}, fmt.Errorf("failed to get room code")).
			Times(1)
		_, err := srv.Create(ctx, "fibbing_it", defaultNewHostPlayer)
		assert.Error(t, err)
	})

	t.Run("Should throw error when we fail to create room in DB", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom)

		ctx := context.Background()
		mockRandom.EXPECT().GetNickname().Return(defaultNewPlayer.Nickname)
		mockRandom.EXPECT().GetAvatar().Return(defaultNewPlayer.Avatar)
		mockRandom.EXPECT().GetRoomCode().Return(roomCode)
		mockStore.EXPECT().GetRoomByCode(ctx, roomCode).Return(sqlc.Room{}, sql.ErrNoRows).Times(1)
		mockRandom.EXPECT().GetID().Return(roomID)

		createRoom := getCreateRoomParams(defaultNewPlayer, defaultNewHostPlayer)
		mockStore.EXPECT().CreateRoom(ctx, createRoom).Return(fmt.Errorf("failed to create room"))
		_, err := srv.Create(ctx, "fibbing_it", defaultNewHostPlayer)
		assert.Error(t, err)
	})
}

func getCreateRoomParams(newCreatedPlayer service.NewPlayer, newPlayer service.NewHostPlayer) sqlc.CreateRoomParams {
	addPlayer := sqlc.AddPlayerParams{
		ID:       newCreatedPlayer.ID,
		Avatar:   newCreatedPlayer.Avatar,
		Nickname: newCreatedPlayer.Nickname,
	}
	addRoom := sqlc.AddRoomParams{
		ID:         roomID,
		GameName:   "fibbing_it",
		RoomCode:   roomCode,
		RoomState:  sqlc.ROOMSTATE_CREATED.String(),
		HostPlayer: newPlayer.ID,
	}

	addRoomPlayer := sqlc.AddRoomPlayerParams{
		RoomID:   addRoom.ID,
		PlayerID: newPlayer.ID,
	}

	createRoom := sqlc.CreateRoomParams{
		Room:       addRoom,
		Player:     addPlayer,
		RoomPlayer: addRoomPlayer,
	}
	return createRoom
}

func TestLobbyServiceJoin(t *testing.T) {
	defaultNewPlayer := service.NewPlayer{
		ID:       "11111-9f7a-4392-b523-fd433b3208ea",
		Nickname: "",
		Avatar:   []byte(""),
	}

	t.Run("Should allow player to join room successfully", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom)

		ctx := context.Background()
		mockRandom.EXPECT().GetNickname().Return(defaultNewPlayer.Nickname)
		mockRandom.EXPECT().GetAvatar().Return(defaultNewPlayer.Avatar)
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(sqlc.Room{ID: roomID, RoomState: sqlc.ROOMSTATE_CREATED.String()}, nil)
		mockStore.EXPECT().GetAllPlayerByRoomCode(ctx, roomCode).Return([]sqlc.GetAllPlayerByRoomCodeRow{
			{
				Nickname: "Hello",
			},
		}, nil)

		addPlayer := sqlc.AddPlayerParams{
			ID:       defaultNewPlayer.ID,
			Avatar:   defaultNewPlayer.Avatar,
			Nickname: defaultNewPlayer.Nickname,
		}

		addRoomPlayer := sqlc.AddRoomPlayerParams{
			RoomID:   roomID,
			PlayerID: defaultNewPlayer.ID,
		}

		addPlayerToRoom := sqlc.AddPlayerToRoomArgs{
			Player:     addPlayer,
			RoomPlayer: addRoomPlayer,
		}
		mockStore.EXPECT().AddPlayerToRoom(ctx, addPlayerToRoom).Return(nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, defaultNewPlayer.ID).Return([]sqlc.GetAllPlayersInRoomRow{
			{
				ID:         "b75599-9f7a-4392-b523-fd433b3208ea",
				Nickname:   "EmotionalTiger",
				Avatar:     []byte(""),
				IsReady:    sql.NullBool{Bool: false, Valid: true},
				HostPlayer: "22222-9f7a-4392-b523-fd433b3208ea",
			},
			{
				ID:         "22222-9f7a-4392-b523-fd433b3208ea",
				Nickname:   "Hello",
				Avatar:     []byte(""),
				IsReady:    sql.NullBool{Bool: false, Valid: true},
				HostPlayer: "22222-9f7a-4392-b523-fd433b3208ea",
			},
		}, nil)
		lobby, err := srv.Join(ctx, roomCode, defaultNewPlayer.ID, defaultNewPlayer.Nickname)
		expectedLobby := service.Lobby{
			Code: roomCode,
			Players: []service.LobbyPlayer{
				{
					ID:       "b75599-9f7a-4392-b523-fd433b3208ea",
					Nickname: "EmotionalTiger",
					Avatar:   "",
					IsReady:  false,
					IsHost:   false,
				},
				{
					ID:       "22222-9f7a-4392-b523-fd433b3208ea",
					Nickname: "Hello",
					Avatar:   "",
					IsReady:  false,
					IsHost:   true,
				},
			},
		}
		assert.NoError(t, err)
		assert.Equal(t, expectedLobby, lobby)
	})

	t.Run("Should allow player to join room successfully, when they pass nickname", func(t *testing.T) {
		nickname := "MyNickname"

		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom)

		ctx := context.Background()
		mockRandom.EXPECT().GetAvatar().Return(defaultNewPlayer.Avatar)
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(sqlc.Room{ID: roomID, RoomState: sqlc.ROOMSTATE_CREATED.String()}, nil)
		mockStore.EXPECT().GetAllPlayerByRoomCode(ctx, roomCode).Return([]sqlc.GetAllPlayerByRoomCodeRow{
			{
				Nickname: "Hello",
			},
		}, nil)

		addPlayer := sqlc.AddPlayerParams{
			ID:       defaultNewPlayer.ID,
			Avatar:   defaultNewPlayer.Avatar,
			Nickname: nickname,
		}

		addRoomPlayer := sqlc.AddRoomPlayerParams{
			RoomID:   roomID,
			PlayerID: defaultNewPlayer.ID,
		}

		addPlayerToRoom := sqlc.AddPlayerToRoomArgs{
			Player:     addPlayer,
			RoomPlayer: addRoomPlayer,
		}
		mockStore.EXPECT().AddPlayerToRoom(ctx, addPlayerToRoom).Return(nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, defaultNewPlayer.ID).Return([]sqlc.GetAllPlayersInRoomRow{
			{
				ID:         "b75599-9f7a-4392-b523-fd433b3208ea",
				Nickname:   nickname,
				Avatar:     []byte(""),
				IsReady:    sql.NullBool{Bool: false, Valid: true},
				HostPlayer: "22222-9f7a-4392-b523-fd433b3208ea",
			},
			{
				ID:         "22222-9f7a-4392-b523-fd433b3208ea",
				Nickname:   "Hello",
				Avatar:     []byte(""),
				IsReady:    sql.NullBool{Bool: false, Valid: true},
				HostPlayer: "22222-9f7a-4392-b523-fd433b3208ea",
			},
		}, nil)
		lobby, err := srv.Join(ctx, roomCode, defaultNewPlayer.ID, nickname)
		expectedLobby := service.Lobby{
			Code: roomCode,
			Players: []service.LobbyPlayer{
				{
					ID:       "b75599-9f7a-4392-b523-fd433b3208ea",
					Nickname: nickname,
					Avatar:   "",
					IsReady:  false,
					IsHost:   false,
				},
				{
					ID:       "22222-9f7a-4392-b523-fd433b3208ea",
					Nickname: "Hello",
					Avatar:   "",
					IsReady:  false,
					IsHost:   true,
				},
			},
		}
		assert.NoError(t, err)
		assert.Equal(t, expectedLobby, lobby)
	})

	t.Run("Should fail to join room, not in CREATED room state", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom)

		ctx := context.Background()
		mockRandom.EXPECT().GetNickname().Return(defaultNewPlayer.Nickname)
		mockRandom.EXPECT().GetAvatar().Return(defaultNewPlayer.Avatar)
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(sqlc.Room{ID: roomID, RoomState: sqlc.ROOMSTATE_PLAYING.String()}, nil)

		_, err := srv.Join(ctx, roomCode, defaultNewPlayer.ID, defaultNewPlayer.Nickname)
		assert.ErrorContains(t, err, "room is not in CREATED state")
	})

	t.Run("Should fail to join room, nickname already exists", func(t *testing.T) {
		nickname := "Hello"

		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom)

		ctx := context.Background()
		mockRandom.EXPECT().GetAvatar().Return(defaultNewPlayer.Avatar)
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(sqlc.Room{ID: roomID, RoomState: sqlc.ROOMSTATE_CREATED.String()}, nil)
		mockStore.EXPECT().GetAllPlayerByRoomCode(ctx, roomCode).Return([]sqlc.GetAllPlayerByRoomCodeRow{
			{
				Nickname: "Hello",
			},
		}, nil)

		_, err := srv.Join(ctx, roomCode, defaultNewPlayer.ID, nickname)
		assert.ErrorContains(t, err, "nickname already exists in room")
	})

	t.Run("Should fail to join room because we fail to add player to DB", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom)

		ctx := context.Background()
		mockRandom.EXPECT().GetNickname().Return(defaultNewPlayer.Nickname)
		mockRandom.EXPECT().GetAvatar().Return(defaultNewPlayer.Avatar)
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(sqlc.Room{ID: roomID, RoomState: sqlc.ROOMSTATE_CREATED.String()}, nil)
		mockStore.EXPECT().GetAllPlayerByRoomCode(ctx, roomCode).Return([]sqlc.GetAllPlayerByRoomCodeRow{
			{
				Nickname: "Hello",
			},
		}, nil)

		addPlayer := sqlc.AddPlayerParams{
			ID:       defaultNewPlayer.ID,
			Avatar:   defaultNewPlayer.Avatar,
			Nickname: defaultNewPlayer.Nickname,
		}

		addRoomPlayer := sqlc.AddRoomPlayerParams{
			RoomID:   roomID,
			PlayerID: defaultNewPlayer.ID,
		}

		addPlayerToRoom := sqlc.AddPlayerToRoomArgs{
			Player:     addPlayer,
			RoomPlayer: addRoomPlayer,
		}
		mockStore.EXPECT().AddPlayerToRoom(ctx, addPlayerToRoom).Return(fmt.Errorf("failed to add player to room"))

		_, err := srv.Join(ctx, roomCode, defaultNewPlayer.ID, defaultNewPlayer.Nickname)
		assert.Error(t, err)
	})

	t.Run("Should fail to join room because fail to get all players in room", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom)

		ctx := context.Background()
		mockRandom.EXPECT().GetNickname().Return(defaultNewPlayer.Nickname)
		mockRandom.EXPECT().GetAvatar().Return(defaultNewPlayer.Avatar)
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(sqlc.Room{ID: roomID, RoomState: sqlc.ROOMSTATE_CREATED.String()}, nil)
		mockStore.EXPECT().GetAllPlayerByRoomCode(ctx, roomCode).Return([]sqlc.GetAllPlayerByRoomCodeRow{
			{
				Nickname: "Hello",
			},
		}, nil)

		addPlayer := sqlc.AddPlayerParams{
			ID:       defaultNewPlayer.ID,
			Avatar:   defaultNewPlayer.Avatar,
			Nickname: defaultNewPlayer.Nickname,
		}

		addRoomPlayer := sqlc.AddRoomPlayerParams{
			RoomID:   roomID,
			PlayerID: defaultNewPlayer.ID,
		}

		addPlayerToRoom := sqlc.AddPlayerToRoomArgs{
			Player:     addPlayer,
			RoomPlayer: addRoomPlayer,
		}
		mockStore.EXPECT().AddPlayerToRoom(ctx, addPlayerToRoom).Return(nil)
		mockStore.EXPECT().
			GetAllPlayersInRoom(ctx, defaultNewPlayer.ID).
			Return([]sqlc.GetAllPlayersInRoomRow{}, fmt.Errorf("failed to get all players in room"))
		_, err := srv.Join(ctx, roomCode, defaultNewPlayer.ID, defaultNewPlayer.Nickname)
		assert.Error(t, err)
	})
}

func TestLobbyServiceKickPlayer(t *testing.T) {
	defaultNewPlayer := service.NewPlayer{
		ID:       "11111-9f7a-4392-b523-fd433b3208ea",
		Nickname: "Hello",
		Avatar:   []byte(""),
	}

	t.Run("Should kick player from lobby successfully", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom)

		ctx := context.Background()
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(sqlc.Room{ID: roomID, HostPlayer: hostPlayerID, RoomState: sqlc.ROOMSTATE_CREATED.String()}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, hostPlayerID).Return([]sqlc.GetAllPlayersInRoomRow{
			{
				ID:         defaultNewPlayer.ID,
				Nickname:   "Hello",
				HostPlayer: hostPlayerID,
			},
			{
				ID:         hostPlayerID,
				Nickname:   "EmotionalTiger",
				HostPlayer: hostPlayerID,
			},
		}, nil)

		mockStore.EXPECT().RemovePlayerFromRoom(ctx, defaultNewPlayer.ID).Return(sqlc.RoomsPlayer{}, nil)
		lobby, playerKickedID, err := srv.KickPlayer(ctx, roomCode, hostPlayerID, defaultNewPlayer.Nickname)
		expectedLobby := service.Lobby{
			Code: roomCode,
			Players: []service.LobbyPlayer{
				{
					ID:       hostPlayerID,
					Nickname: "EmotionalTiger",
					Avatar:   "",
					IsReady:  false,
					IsHost:   true,
				},
			},
		}
		assert.NoError(t, err)
		assert.Equal(t, expectedLobby, lobby)
		assert.Equal(t, defaultNewPlayer.ID, playerKickedID)
	})

	t.Run("Should fail to kick player, because cannot find room by code", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom)

		ctx := context.Background()
		mockStore.EXPECT().GetRoomByCode(ctx, roomCode).Return(sqlc.Room{}, fmt.Errorf("failed to get room by code"))
		_, _, err := srv.KickPlayer(ctx, roomCode, hostPlayerID, defaultNewPlayer.Nickname)
		assert.Error(t, err)
	})

	t.Run("Should fail to kick player because player is not host", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom)

		ctx := context.Background()
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(sqlc.Room{ID: roomID, HostPlayer: hostPlayerID, RoomState: sqlc.ROOMSTATE_CREATED.String()}, nil)
		_, _, err := srv.KickPlayer(ctx, roomCode, defaultNewPlayer.ID, defaultNewPlayer.Nickname)
		assert.ErrorContains(t, err, "player is not the host of the room")
	})

	t.Run("Should fail to kick player because room is not in CREATED state", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom)

		ctx := context.Background()
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(sqlc.Room{ID: roomID, HostPlayer: hostPlayerID, RoomState: sqlc.ROOMSTATE_PLAYING.String()}, nil)
		_, _, err := srv.KickPlayer(ctx, roomCode, hostPlayerID, defaultNewPlayer.Nickname)
		assert.ErrorContains(t, err, "room is not in CREATED state")
	})

	t.Run("Should fail to kick player because we fail to get all players in room from DB", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom)

		ctx := context.Background()
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(sqlc.Room{ID: roomID, HostPlayer: hostPlayerID, RoomState: sqlc.ROOMSTATE_CREATED.String()}, nil)
		mockStore.EXPECT().
			GetAllPlayersInRoom(ctx, hostPlayerID).
			Return([]sqlc.GetAllPlayersInRoomRow{}, fmt.Errorf("failed to get all players in room"))

		_, _, err := srv.KickPlayer(ctx, roomCode, hostPlayerID, defaultNewPlayer.Nickname)
		assert.Error(t, err)
	})

	t.Run("Should fail to kick player because player with nickname is not in room", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom)

		ctx := context.Background()
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(sqlc.Room{ID: roomID, HostPlayer: hostPlayerID, RoomState: sqlc.ROOMSTATE_CREATED.String()}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, hostPlayerID).Return([]sqlc.GetAllPlayersInRoomRow{
			{
				ID:         hostPlayerID,
				Nickname:   "EmotionalTiger",
				HostPlayer: hostPlayerID,
			},
		}, nil)

		_, _, err := srv.KickPlayer(ctx, roomCode, hostPlayerID, defaultNewPlayer.Nickname)
		assert.ErrorContains(t, err, "player with nickname Hello not found to kick")
	})

	t.Run("Should fail to kick player because failed to remove player from room in DB", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom)

		ctx := context.Background()
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(sqlc.Room{ID: roomID, HostPlayer: hostPlayerID, RoomState: sqlc.ROOMSTATE_CREATED.String()}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, hostPlayerID).Return([]sqlc.GetAllPlayersInRoomRow{
			{
				ID:         defaultNewPlayer.ID,
				Nickname:   "Hello",
				HostPlayer: hostPlayerID,
			},
			{
				ID:         hostPlayerID,
				Nickname:   "EmotionalTiger",
				HostPlayer: hostPlayerID,
			},
		}, nil)

		mockStore.EXPECT().
			RemovePlayerFromRoom(ctx, defaultNewPlayer.ID).
			Return(sqlc.RoomsPlayer{}, fmt.Errorf("failed to remove player from room"))
		_, _, err := srv.KickPlayer(ctx, roomCode, hostPlayerID, defaultNewPlayer.Nickname)
		assert.Error(t, err)
	})
}

func TestLobbyServiceStart(t *testing.T) {
	gameName := "fibbing_it"
	groupID := "12345-9f7a-4392-b523-fd433b3208ea"
	gameStateID := "77777-9f7a-4392-b523-fd433b3208ea"

	defaultNewPlayer := service.NewPlayer{
		ID:       "11111-9f7a-4392-b523-fd433b3208ea",
		Nickname: "Hello",
		Avatar:   []byte(""),
	}

	t.Run("Should start game successfully", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom)

		ctx := context.Background()
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(
				sqlc.Room{
					ID:         roomID,
					GameName:   gameName,
					HostPlayer: hostPlayerID,
					RoomState:  sqlc.ROOMSTATE_CREATED.String(),
				}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, hostPlayerID).Return([]sqlc.GetAllPlayersInRoomRow{
			{
				ID:         defaultNewPlayer.ID,
				Nickname:   "Hello",
				HostPlayer: hostPlayerID,
				IsReady:    sql.NullBool{Bool: true, Valid: true},
				RoomCode:   roomCode,
			},
			{
				ID:         hostPlayerID,
				Nickname:   "EmotionalTiger",
				HostPlayer: hostPlayerID,
				IsReady:    sql.NullBool{Bool: true, Valid: true},
				RoomCode:   roomCode,
			},
		}, nil)

		mockStore.EXPECT().GetRandomQuestionByRound(ctx, sqlc.GetRandomQuestionByRoundParams{
			GameName:     gameName,
			LanguageCode: "en-GB",
			Round:        "free_form",
		}).Return(sqlc.Question{
			ID:       "555555-9f7a-4392-b523-fd433b3208ea",
			Question: "What is the capital of France?",
			GroupID:  groupID,
		}, nil)
		mockStore.EXPECT().GetRandomQuestionInGroup(ctx, sqlc.GetRandomQuestionInGroupParams{
			GroupID: groupID,
			ID:      "555555-9f7a-4392-b523-fd433b3208ea",
		}).Return(sqlc.GetRandomQuestionInGroupRow{
			ID:       "666666-9f7a-4392-b523-fd433b3208ea",
			Question: "What is the capital of Germany?",
		}, nil)
		mockRandom.EXPECT().GetFibberIndex(2).Return(1)
		mockRandom.EXPECT().GetID().Return(gameStateID)
		mockStore.EXPECT().StartGame(ctx, sqlc.StartGameArgs{
			GameStateID:       gameStateID,
			RoomID:            roomID,
			NormalsQuestionID: "555555-9f7a-4392-b523-fd433b3208ea",
			FibberQuestionID:  "666666-9f7a-4392-b523-fd433b3208ea",
			Players: []sqlc.GetAllPlayersInRoomRow{
				{
					ID:         defaultNewPlayer.ID,
					Nickname:   "Hello",
					IsReady:    sql.NullBool{Bool: true, Valid: true},
					RoomCode:   roomCode,
					HostPlayer: hostPlayerID,
				},
				{
					ID:         hostPlayerID,
					Nickname:   "EmotionalTiger",
					IsReady:    sql.NullBool{Bool: true, Valid: true},
					RoomCode:   roomCode,
					HostPlayer: hostPlayerID,
				},
			},
			FibberLoc: 1,
		}).Return(nil)

		gameState, err := srv.Start(ctx, roomCode, hostPlayerID)
		expectedGameState := service.QuestionState{
			GameStateID: gameStateID,
			Players: []service.PlayerWithRole{
				{
					ID:       defaultNewPlayer.ID,
					Nickname: "Hello",
					Role:     "normal",
					Question: "What is the capital of France?",
				},
				{
					ID:       hostPlayerID,
					Nickname: "EmotionalTiger",
					Role:     "fibber",
					Question: "What is the capital of Germany?",
				},
			},
			Round:     1,
			RoundType: "free_form",
			RoomCode:  roomCode,
		}

		assert.NoError(t, err)
		assert.Equal(t, expectedGameState, gameState)
	})

	t.Run("Should fail to start game because host did not start game", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom)

		ctx := context.Background()
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(
				sqlc.Room{ID: roomID,
					GameName:   gameName,
					HostPlayer: hostPlayerID,
					RoomState:  sqlc.ROOMSTATE_CREATED.String(),
				}, nil)

		_, err := srv.Start(ctx, roomCode, defaultNewPlayer.ID)
		assert.ErrorContains(t, err, "")
	})

	t.Run("Should fail to start game because room state not in CREATED state", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom)

		ctx := context.Background()
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(
				sqlc.Room{ID: roomID,
					GameName:   gameName,
					HostPlayer: hostPlayerID,
					RoomState:  sqlc.ROOMSTATE_PLAYING.String(),
				}, nil)
		_, err := srv.Start(ctx, roomCode, hostPlayerID)
		assert.ErrorContains(t, err, "room is not in CREATED state")
	})

	t.Run("Should fail to start game because we failed to get all players in room", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom)

		ctx := context.Background()
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(
				sqlc.Room{ID: roomID,
					GameName:   gameName,
					HostPlayer: hostPlayerID,
					RoomState:  sqlc.ROOMSTATE_CREATED.String(),
				}, nil)
		mockStore.EXPECT().
			GetAllPlayersInRoom(ctx, hostPlayerID).
			Return([]sqlc.GetAllPlayersInRoomRow{}, fmt.Errorf("failed to get all players in room"))

		_, err := srv.Start(ctx, roomCode, hostPlayerID)
		assert.Error(t, err)
	})

	t.Run("Should fail to start game because too few players in room", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom)

		ctx := context.Background()
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(
				sqlc.Room{ID: roomID,
					GameName:   gameName,
					HostPlayer: hostPlayerID,
					RoomState:  sqlc.ROOMSTATE_CREATED.String(),
				}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, hostPlayerID).Return([]sqlc.GetAllPlayersInRoomRow{
			{
				ID:         defaultNewPlayer.ID,
				Nickname:   "Hello",
				HostPlayer: hostPlayerID,
				IsReady:    sql.NullBool{Bool: true, Valid: true},
				RoomCode:   roomCode,
			},
		}, nil)

		_, err := srv.Start(ctx, roomCode, hostPlayerID)
		assert.ErrorContains(t, err, "not enough players to start the game")
	})

	t.Run("Should fail to start game because not every player is ready", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom)

		ctx := context.Background()
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(
				sqlc.Room{
					ID:         roomID,
					GameName:   gameName,
					HostPlayer: hostPlayerID,
					RoomState:  sqlc.ROOMSTATE_CREATED.String(),
				},
				nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, hostPlayerID).Return([]sqlc.GetAllPlayersInRoomRow{
			{
				ID:         defaultNewPlayer.ID,
				Nickname:   "Hello",
				HostPlayer: hostPlayerID,
				IsReady:    sql.NullBool{Bool: true, Valid: true},
				RoomCode:   roomCode,
			},
			{
				ID:         hostPlayerID,
				Nickname:   "EmotionalTiger",
				HostPlayer: hostPlayerID,
				IsReady:    sql.NullBool{Bool: false, Valid: true},
				RoomCode:   roomCode,
			},
		}, nil)

		_, err := srv.Start(ctx, roomCode, hostPlayerID)
		assert.ErrorContains(t, err, "not all players are ready")
	})

	t.Run("Should fail to start game because we fail to get random question for normals", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom)

		ctx := context.Background()
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(
				sqlc.Room{
					ID:         roomID,
					GameName:   gameName,
					HostPlayer: hostPlayerID,
					RoomState:  sqlc.ROOMSTATE_CREATED.String(),
				}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, hostPlayerID).Return([]sqlc.GetAllPlayersInRoomRow{
			{
				ID:         defaultNewPlayer.ID,
				Nickname:   "Hello",
				HostPlayer: hostPlayerID,
				IsReady:    sql.NullBool{Bool: true, Valid: true},
				RoomCode:   roomCode,
			},
			{
				ID:         hostPlayerID,
				Nickname:   "EmotionalTiger",
				HostPlayer: hostPlayerID,
				IsReady:    sql.NullBool{Bool: true, Valid: true},
				RoomCode:   roomCode,
			},
		}, nil)

		mockStore.EXPECT().GetRandomQuestionByRound(ctx, sqlc.GetRandomQuestionByRoundParams{
			GameName:     gameName,
			LanguageCode: "en-GB",
			Round:        "free_form",
		}).Return(sqlc.Question{}, fmt.Errorf("failed to get random question for normals"))
		_, err := srv.Start(ctx, roomCode, hostPlayerID)
		assert.Error(t, err)
	})

	t.Run("Should fail to start game because we fail to get random question for fibber", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom)

		ctx := context.Background()
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(
				sqlc.Room{ID: roomID,
					GameName:   gameName,
					HostPlayer: hostPlayerID,
					RoomState:  sqlc.ROOMSTATE_CREATED.String(),
				},
				nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, hostPlayerID).Return([]sqlc.GetAllPlayersInRoomRow{
			{
				ID:         defaultNewPlayer.ID,
				Nickname:   "Hello",
				HostPlayer: hostPlayerID,
				IsReady:    sql.NullBool{Bool: true, Valid: true},
				RoomCode:   roomCode,
			},
			{
				ID:         hostPlayerID,
				Nickname:   "EmotionalTiger",
				HostPlayer: hostPlayerID,
				IsReady:    sql.NullBool{Bool: true, Valid: true},
				RoomCode:   roomCode,
			},
		}, nil)

		mockStore.EXPECT().GetRandomQuestionByRound(ctx, sqlc.GetRandomQuestionByRoundParams{
			GameName:     gameName,
			LanguageCode: "en-GB",
			Round:        "free_form",
		}).Return(sqlc.Question{
			ID:       "555555-9f7a-4392-b523-fd433b3208ea",
			Question: "What is the capital of France?",
			GroupID:  groupID,
		}, nil)
		mockStore.EXPECT().GetRandomQuestionInGroup(ctx, sqlc.GetRandomQuestionInGroupParams{
			GroupID: groupID,
			ID:      "555555-9f7a-4392-b523-fd433b3208ea",
		}).Return(sqlc.GetRandomQuestionInGroupRow{}, fmt.Errorf("failed to get random question for fibber"))

		_, err := srv.Start(ctx, roomCode, hostPlayerID)
		assert.Error(t, err)
	})

	t.Run("Should fail to start game because we fail to start game in DB", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom)

		ctx := context.Background()
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(
				sqlc.Room{
					ID:         roomID,
					GameName:   gameName,
					HostPlayer: hostPlayerID,
					RoomState:  sqlc.ROOMSTATE_CREATED.String(),
				},
				nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, hostPlayerID).Return([]sqlc.GetAllPlayersInRoomRow{
			{
				ID:         defaultNewPlayer.ID,
				Nickname:   "Hello",
				HostPlayer: hostPlayerID,
				IsReady:    sql.NullBool{Bool: true, Valid: true},
				RoomCode:   roomCode,
			},
			{
				ID:         hostPlayerID,
				Nickname:   "EmotionalTiger",
				HostPlayer: hostPlayerID,
				IsReady:    sql.NullBool{Bool: true, Valid: true},
				RoomCode:   roomCode,
			},
		}, nil)

		mockStore.EXPECT().GetRandomQuestionByRound(ctx, sqlc.GetRandomQuestionByRoundParams{
			GameName:     gameName,
			LanguageCode: "en-GB",
			Round:        "free_form",
		}).Return(sqlc.Question{
			ID:       "555555-9f7a-4392-b523-fd433b3208ea",
			Question: "What is the capital of France?",
			GroupID:  groupID,
		}, nil)
		mockStore.EXPECT().GetRandomQuestionInGroup(ctx, sqlc.GetRandomQuestionInGroupParams{
			GroupID: groupID,
			ID:      "555555-9f7a-4392-b523-fd433b3208ea",
		}).Return(sqlc.GetRandomQuestionInGroupRow{
			ID:       "666666-9f7a-4392-b523-fd433b3208ea",
			Question: "What is the capital of Germany?",
		}, nil)
		mockRandom.EXPECT().GetFibberIndex(2).Return(1)
		mockRandom.EXPECT().GetID().Return(gameStateID)
		mockStore.EXPECT().StartGame(ctx, sqlc.StartGameArgs{
			GameStateID:       gameStateID,
			RoomID:            roomID,
			NormalsQuestionID: "555555-9f7a-4392-b523-fd433b3208ea",
			FibberQuestionID:  "666666-9f7a-4392-b523-fd433b3208ea",
			Players: []sqlc.GetAllPlayersInRoomRow{
				{
					ID:         defaultNewPlayer.ID,
					Nickname:   "Hello",
					IsReady:    sql.NullBool{Bool: true, Valid: true},
					RoomCode:   roomCode,
					HostPlayer: hostPlayerID,
				},
				{
					ID:         hostPlayerID,
					Nickname:   "EmotionalTiger",
					IsReady:    sql.NullBool{Bool: true, Valid: true},
					RoomCode:   roomCode,
					HostPlayer: hostPlayerID,
				},
			},
			FibberLoc: 1,
		}).Return(fmt.Errorf("failed to start game"))

		_, err := srv.Start(ctx, roomCode, hostPlayerID)
		assert.Error(t, err)
	})
}
