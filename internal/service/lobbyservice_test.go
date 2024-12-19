package service_test

import (
	"context"
	"database/sql"
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

const roomCode = "ABC12"
const gameName = "fibbing_it"

// TODO: move to generic
var roomID = uuid.MustParse("0193a627-db9b-7af3-88d8-6b3164d4b969")
var hostPlayerID = uuid.MustParse("0193a623-8423-74b3-b991-896d7c6df52a")

func TestLobbyServiceCreate(t *testing.T) {
	defaultNewHostPlayer := service.NewHostPlayer{
		ID: uuid.MustParse("0193a626-2586-7784-9b5b-104d927d64ca"),
	}

	defaultNewPlayer := service.NewPlayer{
		ID:       defaultNewHostPlayer.ID,
		Nickname: "Majiy00",
		Avatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Majiy00",
	}

	t.Run("Should create room successfully", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom)

		ctx := context.Background()
		mockRandom.EXPECT().GetNickname().Return(defaultNewPlayer.Nickname)
		mockRandom.EXPECT().GetAvatar(defaultNewPlayer.Nickname).Return(defaultNewPlayer.Avatar)
		mockRandom.EXPECT().GetRoomCode().Return(roomCode)
		mockStore.EXPECT().GetRoomByCode(ctx, roomCode).Return(db.Room{}, sql.ErrNoRows)
		mockRandom.EXPECT().GetID().Return(roomID)

		createRoom := getCreateRoomParams(defaultNewPlayer, defaultNewHostPlayer)
		mockStore.EXPECT().CreateRoom(ctx, createRoom).Return(nil)
		lobby, err := srv.Create(ctx, gameName, defaultNewHostPlayer)

		expectedLobby := service.Lobby{
			Code: roomCode,
			Players: []service.LobbyPlayer{
				{
					ID:       defaultNewPlayer.ID,
					Nickname: defaultNewPlayer.Nickname,
					Avatar:   defaultNewPlayer.Avatar,
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
			ID:       uuid.MustParse("0193a626-2586-7784-9b5b-104d927d64ca"),
		}

		newPlayer := service.NewPlayer{
			ID:       newHostPlayer.ID,
			Nickname: newHostPlayer.Nickname,
			Avatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=MyNickname",
		}

		ctx := context.Background()
		mockRandom.EXPECT().GetAvatar(newHostPlayer.Nickname).Return(newPlayer.Avatar)
		mockRandom.EXPECT().GetRoomCode().Return(roomCode)
		mockStore.EXPECT().GetRoomByCode(ctx, roomCode).Return(db.Room{}, sql.ErrNoRows)
		mockRandom.EXPECT().GetID().Return(roomID)

		createRoom := getCreateRoomParams(newPlayer, newHostPlayer)
		mockStore.EXPECT().CreateRoom(ctx, createRoom).Return(nil)
		lobby, err := srv.Create(ctx, gameName, newHostPlayer)

		expectedLobby := service.Lobby{
			Code: roomCode,
			Players: []service.LobbyPlayer{
				{
					ID:       newPlayer.ID,
					Nickname: newPlayer.Nickname,
					Avatar:   newPlayer.Avatar,
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
		mockRandom.EXPECT().GetAvatar(defaultNewPlayer.Nickname).Return(defaultNewPlayer.Avatar)
		mockRandom.EXPECT().GetRoomCode().Return(roomCode)
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(db.Room{RoomState: db.ROOMSTATE_CREATED.String()}, nil).
			Times(1)
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(db.Room{RoomState: db.ROOMSTATE_ABANDONED.String()}, nil).
			Times(1)
		mockRandom.EXPECT().GetID().Return(roomID)

		createRoom := getCreateRoomParams(defaultNewPlayer, defaultNewHostPlayer)
		mockStore.EXPECT().CreateRoom(ctx, createRoom).Return(nil)
		lobby, err := srv.Create(ctx, gameName, defaultNewHostPlayer)

		expectedLobby := service.Lobby{
			Code: roomCode,
			Players: []service.LobbyPlayer{
				{
					ID:       defaultNewPlayer.ID,
					Nickname: defaultNewPlayer.Nickname,
					Avatar:   defaultNewPlayer.Avatar,
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
		mockRandom.EXPECT().GetAvatar(defaultNewPlayer.Nickname).Return(defaultNewPlayer.Avatar)
		mockRandom.EXPECT().GetRoomCode().Return(roomCode)
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(db.Room{}, fmt.Errorf("failed to get room code")).
			Times(1)
		_, err := srv.Create(ctx, gameName, defaultNewHostPlayer)
		assert.Error(t, err)
	})

	t.Run("Should throw error when we fail to create room in DB", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom)

		ctx := context.Background()
		mockRandom.EXPECT().GetNickname().Return(defaultNewPlayer.Nickname)
		mockRandom.EXPECT().GetAvatar(defaultNewPlayer.Nickname).Return(defaultNewPlayer.Avatar)
		mockRandom.EXPECT().GetRoomCode().Return(roomCode)
		mockStore.EXPECT().GetRoomByCode(ctx, roomCode).Return(db.Room{}, sql.ErrNoRows).Times(1)
		mockRandom.EXPECT().GetID().Return(roomID)

		createRoom := getCreateRoomParams(defaultNewPlayer, defaultNewHostPlayer)
		mockStore.EXPECT().CreateRoom(ctx, createRoom).Return(fmt.Errorf("failed to create room"))
		_, err := srv.Create(ctx, gameName, defaultNewHostPlayer)
		assert.Error(t, err)
	})
}

func getCreateRoomParams(newCreatedPlayer service.NewPlayer, newPlayer service.NewHostPlayer) db.CreateRoomArgs {
	addPlayer := db.AddPlayerParams{
		ID:       newCreatedPlayer.ID,
		Avatar:   newCreatedPlayer.Avatar,
		Nickname: newCreatedPlayer.Nickname,
	}
	addRoom := db.AddRoomParams{
		ID:         roomID,
		GameName:   gameName,
		RoomCode:   roomCode,
		RoomState:  db.ROOMSTATE_CREATED.String(),
		HostPlayer: newPlayer.ID,
	}

	addRoomPlayer := db.AddRoomPlayerParams{
		RoomID:   addRoom.ID,
		PlayerID: newPlayer.ID,
	}

	createRoom := db.CreateRoomArgs{
		Room:       addRoom,
		Player:     addPlayer,
		RoomPlayer: addRoomPlayer,
	}
	return createRoom
}

func TestLobbyServiceJoin(t *testing.T) {
	defaultNewPlayer := service.NewPlayer{
		ID:       uuid.MustParse("0193a626-2586-7784-9b5b-104d927d64ca"),
		Nickname: "",
		Avatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=",
	}

	t.Run("Should allow player to join room successfully", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom)

		ctx := context.Background()
		mockRandom.EXPECT().GetNickname().Return(defaultNewPlayer.Nickname)
		mockRandom.EXPECT().GetAvatar(defaultNewPlayer.Nickname).Return(defaultNewPlayer.Avatar)
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(db.Room{ID: roomID, RoomState: db.ROOMSTATE_CREATED.String()}, nil)
		mockStore.EXPECT().GetAllPlayerByRoomCode(ctx, roomCode).Return([]db.GetAllPlayerByRoomCodeRow{
			{
				Nickname: "Hello",
			},
		}, nil)

		addPlayer := db.AddPlayerParams{
			ID:       defaultNewPlayer.ID,
			Avatar:   defaultNewPlayer.Avatar,
			Nickname: defaultNewPlayer.Nickname,
		}

		addRoomPlayer := db.AddRoomPlayerParams{
			RoomID:   roomID,
			PlayerID: defaultNewPlayer.ID,
		}

		addPlayerToRoom := db.AddPlayerToRoomArgs{
			Player:     addPlayer,
			RoomPlayer: addRoomPlayer,
		}
		mockStore.EXPECT().AddPlayerToRoom(ctx, addPlayerToRoom).Return(nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, defaultNewPlayer.ID).Return([]db.GetAllPlayersInRoomRow{
			{
				ID:         uuid.MustParse("0193a628-51bc-7a60-9204-a8667771f278"),
				Nickname:   "EmotionalTiger",
				Avatar:     "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=EmotionalTiger",
				IsReady:    pgtype.Bool{Bool: false, Valid: true},
				HostPlayer: uuid.MustParse("0193a628-8b7b-7ad9-aefe-031ec85289fa"),
			},
			{
				ID:         uuid.MustParse("0193a628-8b7b-7ad9-aefe-031ec85289fa"),
				Nickname:   "Hello",
				Avatar:     "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Hello",
				IsReady:    pgtype.Bool{Bool: false, Valid: true},
				HostPlayer: uuid.MustParse("0193a628-8b7b-7ad9-aefe-031ec85289fa"),
			},
		}, nil)
		lobby, err := srv.Join(ctx, roomCode, defaultNewPlayer.ID, defaultNewPlayer.Nickname)
		expectedLobby := service.Lobby{
			Code: roomCode,
			Players: []service.LobbyPlayer{
				{
					ID:       uuid.MustParse("0193a628-51bc-7a60-9204-a8667771f278"),
					Nickname: "EmotionalTiger",
					Avatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=EmotionalTiger",
					IsReady:  false,
					IsHost:   false,
				},
				{
					ID:       uuid.MustParse("0193a628-8b7b-7ad9-aefe-031ec85289fa"),
					Nickname: "Hello",
					Avatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Hello",
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
		mockRandom.EXPECT().
			GetAvatar(nickname).
			Return("https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=MyNickname")
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(db.Room{ID: roomID, RoomState: db.ROOMSTATE_CREATED.String()}, nil)
		mockStore.EXPECT().GetAllPlayerByRoomCode(ctx, roomCode).Return([]db.GetAllPlayerByRoomCodeRow{
			{
				Nickname: "Hello",
			},
		}, nil)

		addPlayer := db.AddPlayerParams{
			ID:       defaultNewPlayer.ID,
			Avatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=MyNickname",
			Nickname: nickname,
		}

		addRoomPlayer := db.AddRoomPlayerParams{
			RoomID:   roomID,
			PlayerID: defaultNewPlayer.ID,
		}

		addPlayerToRoom := db.AddPlayerToRoomArgs{
			Player:     addPlayer,
			RoomPlayer: addRoomPlayer,
		}
		mockStore.EXPECT().AddPlayerToRoom(ctx, addPlayerToRoom).Return(nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, defaultNewPlayer.ID).Return([]db.GetAllPlayersInRoomRow{
			{
				ID:         uuid.MustParse("0193a628-51bc-7a60-9204-a8667771f278"),
				Nickname:   nickname,
				Avatar:     "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=MyNickname",
				IsReady:    pgtype.Bool{Bool: false, Valid: true},
				HostPlayer: uuid.MustParse("0193a628-8b7b-7ad9-aefe-031ec85289fa"),
			},
			{
				ID:         uuid.MustParse("0193a628-8b7b-7ad9-aefe-031ec85289fa"),
				Nickname:   "Hello",
				Avatar:     "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Hello",
				IsReady:    pgtype.Bool{Bool: false, Valid: true},
				HostPlayer: uuid.MustParse("0193a628-8b7b-7ad9-aefe-031ec85289fa"),
			},
		}, nil)
		lobby, err := srv.Join(ctx, roomCode, defaultNewPlayer.ID, nickname)
		expectedLobby := service.Lobby{
			Code: roomCode,
			Players: []service.LobbyPlayer{
				{
					ID:       uuid.MustParse("0193a628-51bc-7a60-9204-a8667771f278"),
					Nickname: nickname,
					Avatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=MyNickname",
					IsReady:  false,
					IsHost:   false,
				},
				{
					ID:       uuid.MustParse("0193a628-8b7b-7ad9-aefe-031ec85289fa"),
					Nickname: "Hello",
					Avatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Hello",
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
		mockRandom.EXPECT().GetAvatar(defaultNewPlayer.Nickname).Return(defaultNewPlayer.Avatar)
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(db.Room{ID: roomID, RoomState: db.ROOMSTATE_PLAYING.String()}, nil)

		_, err := srv.Join(ctx, roomCode, defaultNewPlayer.ID, defaultNewPlayer.Nickname)
		assert.ErrorContains(t, err, "room is not in CREATED state")
	})

	t.Run("Should fail to join room, nickname already exists", func(t *testing.T) {
		nickname := "Hello"

		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom)

		ctx := context.Background()
		mockRandom.EXPECT().GetAvatar(nickname).Return(defaultNewPlayer.Avatar)
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(db.Room{ID: roomID, RoomState: db.ROOMSTATE_CREATED.String()}, nil)
		mockStore.EXPECT().GetAllPlayerByRoomCode(ctx, roomCode).Return([]db.GetAllPlayerByRoomCodeRow{
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
		mockRandom.EXPECT().GetAvatar(defaultNewPlayer.Nickname).Return(defaultNewPlayer.Avatar)
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(db.Room{ID: roomID, RoomState: db.ROOMSTATE_CREATED.String()}, nil)
		mockStore.EXPECT().GetAllPlayerByRoomCode(ctx, roomCode).Return([]db.GetAllPlayerByRoomCodeRow{
			{
				Nickname: "Hello",
			},
		}, nil)

		addPlayer := db.AddPlayerParams{
			ID:       defaultNewPlayer.ID,
			Avatar:   defaultNewPlayer.Avatar,
			Nickname: defaultNewPlayer.Nickname,
		}

		addRoomPlayer := db.AddRoomPlayerParams{
			RoomID:   roomID,
			PlayerID: defaultNewPlayer.ID,
		}

		addPlayerToRoom := db.AddPlayerToRoomArgs{
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
		mockRandom.EXPECT().GetAvatar(defaultNewPlayer.Nickname).Return(defaultNewPlayer.Avatar)
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(db.Room{ID: roomID, RoomState: db.ROOMSTATE_CREATED.String()}, nil)
		mockStore.EXPECT().GetAllPlayerByRoomCode(ctx, roomCode).Return([]db.GetAllPlayerByRoomCodeRow{
			{
				Nickname: "Hello",
			},
		}, nil)

		addPlayer := db.AddPlayerParams{
			ID:       defaultNewPlayer.ID,
			Avatar:   defaultNewPlayer.Avatar,
			Nickname: defaultNewPlayer.Nickname,
		}

		addRoomPlayer := db.AddRoomPlayerParams{
			RoomID:   roomID,
			PlayerID: defaultNewPlayer.ID,
		}

		addPlayerToRoom := db.AddPlayerToRoomArgs{
			Player:     addPlayer,
			RoomPlayer: addRoomPlayer,
		}
		mockStore.EXPECT().AddPlayerToRoom(ctx, addPlayerToRoom).Return(nil)
		mockStore.EXPECT().
			GetAllPlayersInRoom(ctx, defaultNewPlayer.ID).
			Return([]db.GetAllPlayersInRoomRow{}, fmt.Errorf("failed to get all players in room"))
		_, err := srv.Join(ctx, roomCode, defaultNewPlayer.ID, defaultNewPlayer.Nickname)
		assert.Error(t, err)
	})
}

func TestLobbyServiceKickPlayer(t *testing.T) {
	defaultNewPlayer := service.NewPlayer{
		ID:       uuid.MustParse("0193a626-2586-7784-9b5b-104d927d64ca"),
		Nickname: "Hello",
		Avatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Hello",
	}

	t.Run("Should kick player from lobby successfully", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom)

		ctx := context.Background()
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(db.Room{ID: roomID, HostPlayer: hostPlayerID, RoomState: db.ROOMSTATE_CREATED.String()}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, hostPlayerID).Return([]db.GetAllPlayersInRoomRow{
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

		mockStore.EXPECT().RemovePlayerFromRoom(ctx, defaultNewPlayer.ID).Return(db.RoomsPlayer{}, nil)
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
		mockStore.EXPECT().GetRoomByCode(ctx, roomCode).Return(db.Room{}, fmt.Errorf("failed to get room by code"))
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
			Return(db.Room{ID: roomID, HostPlayer: hostPlayerID, RoomState: db.ROOMSTATE_CREATED.String()}, nil)
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
			Return(db.Room{ID: roomID, HostPlayer: hostPlayerID, RoomState: db.ROOMSTATE_PLAYING.String()}, nil)
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
			Return(db.Room{ID: roomID, HostPlayer: hostPlayerID, RoomState: db.ROOMSTATE_CREATED.String()}, nil)
		mockStore.EXPECT().
			GetAllPlayersInRoom(ctx, hostPlayerID).
			Return([]db.GetAllPlayersInRoomRow{}, fmt.Errorf("failed to get all players in room"))

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
			Return(db.Room{ID: roomID, HostPlayer: hostPlayerID, RoomState: db.ROOMSTATE_CREATED.String()}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, hostPlayerID).Return([]db.GetAllPlayersInRoomRow{
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
			Return(db.Room{ID: roomID, HostPlayer: hostPlayerID, RoomState: db.ROOMSTATE_CREATED.String()}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, hostPlayerID).Return([]db.GetAllPlayersInRoomRow{
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
			Return(db.RoomsPlayer{}, fmt.Errorf("failed to remove player from room"))
		_, _, err := srv.KickPlayer(ctx, roomCode, hostPlayerID, defaultNewPlayer.Nickname)
		assert.Error(t, err)
	})
}

func TestLobbyServiceStart(t *testing.T) {
	groupID := uuid.MustParse("0193a629-1fcf-79dd-ac70-760bedbdffa9")
	gameStateID := uuid.MustParse("0193a629-373b-7a3e-b6c2-0e7d2f95ce43")

	defaultNewPlayer := service.NewPlayer{
		ID:       uuid.MustParse("0193a626-2586-7784-9b5b-104d927d64ca"),
		Nickname: "Hello",
		Avatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Hello",
	}

	t.Run("Should start game successfully", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom)

		ctx := context.Background()
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(
				db.Room{
					ID:         roomID,
					GameName:   gameName,
					HostPlayer: hostPlayerID,
					RoomState:  db.ROOMSTATE_CREATED.String(),
				}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, hostPlayerID).Return([]db.GetAllPlayersInRoomRow{
			{
				ID:         defaultNewPlayer.ID,
				Nickname:   "Hello",
				HostPlayer: hostPlayerID,
				IsReady:    pgtype.Bool{Bool: true, Valid: true},
				RoomCode:   roomCode,
			},
			{
				ID:         hostPlayerID,
				Nickname:   "EmotionalTiger",
				HostPlayer: hostPlayerID,
				IsReady:    pgtype.Bool{Bool: true, Valid: true},
				RoomCode:   roomCode,
			},
		}, nil)

		mockStore.EXPECT().GetRandomQuestionByRound(ctx, db.GetRandomQuestionByRoundParams{
			GameName:     gameName,
			LanguageCode: "en-GB",
			Round:        "free_form",
		}).Return(db.Question{
			ID:       uuid.MustParse("0193a629-7dcc-78ad-822f-fd5d83c89ae7"),
			Question: "What is the capital of France?",
			GroupID:  groupID,
		}, nil)
		mockStore.EXPECT().GetRandomQuestionInGroup(ctx, db.GetRandomQuestionInGroupParams{
			GroupID: groupID,
			ID:      uuid.MustParse("0193a629-7dcc-78ad-822f-fd5d83c89ae7"),
		}).Return(db.GetRandomQuestionInGroupRow{
			ID:       uuid.MustParse("0193a629-a9ac-7fc4-828c-a1334c282e0f"),
			Question: "What is the capital of Germany?",
		}, nil)
		mockRandom.EXPECT().GetFibberIndex(2).Return(1)
		mockRandom.EXPECT().GetID().Return(gameStateID)
		deadline := time.Now().Add(5 * time.Second)
		mockStore.EXPECT().StartGame(ctx, db.StartGameArgs{
			GameStateID:       gameStateID,
			RoomID:            roomID,
			NormalsQuestionID: uuid.MustParse("0193a629-7dcc-78ad-822f-fd5d83c89ae7"),
			FibberQuestionID:  uuid.MustParse("0193a629-a9ac-7fc4-828c-a1334c282e0f"),
			Players: []db.GetAllPlayersInRoomRow{
				{
					ID:         defaultNewPlayer.ID,
					Nickname:   "Hello",
					IsReady:    pgtype.Bool{Bool: true, Valid: true},
					RoomCode:   roomCode,
					HostPlayer: hostPlayerID,
				},
				{
					ID:         hostPlayerID,
					Nickname:   "EmotionalTiger",
					IsReady:    pgtype.Bool{Bool: true, Valid: true},
					RoomCode:   roomCode,
					HostPlayer: hostPlayerID,
				},
			},
			FibberLoc: 1,
			Deadline:  deadline,
		}).Return(nil)

		gameState, err := srv.Start(ctx, roomCode, hostPlayerID, deadline)
		expectedGameState := service.QuestionState{
			GameStateID: gameStateID,
			Players: []service.PlayerWithRole{
				{
					ID:       defaultNewPlayer.ID,
					Role:     "normal",
					Question: "What is the capital of France?",
				},
				{
					ID:       hostPlayerID,
					Role:     "fibber",
					Question: "What is the capital of Germany?",
				},
			},
			Round:     1,
			RoundType: "free_form",
		}

		assert.NoError(t, err)

		diffOpts := cmpopts.IgnoreFields(gameState, "Deadline")
		PartialEqual(t, expectedGameState, gameState, diffOpts)
		assert.LessOrEqual(t, int(gameState.Deadline.Seconds()), 5)
	})

	t.Run("Should fail to start game because host did not start game", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom)

		ctx := context.Background()
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(
				db.Room{ID: roomID,
					GameName:   gameName,
					HostPlayer: hostPlayerID,
					RoomState:  db.ROOMSTATE_CREATED.String(),
				}, nil)

		deadline := time.Now().Add(5 * time.Second)
		_, err := srv.Start(ctx, roomCode, defaultNewPlayer.ID, deadline)
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
				db.Room{ID: roomID,
					GameName:   gameName,
					HostPlayer: hostPlayerID,
					RoomState:  db.ROOMSTATE_PLAYING.String(),
				}, nil)

		deadline := time.Now().Add(5 * time.Second)
		_, err := srv.Start(ctx, roomCode, hostPlayerID, deadline)
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
				db.Room{ID: roomID,
					GameName:   gameName,
					HostPlayer: hostPlayerID,
					RoomState:  db.ROOMSTATE_CREATED.String(),
				}, nil)
		mockStore.EXPECT().
			GetAllPlayersInRoom(ctx, hostPlayerID).
			Return([]db.GetAllPlayersInRoomRow{}, fmt.Errorf("failed to get all players in room"))

		deadline := time.Now().Add(5 * time.Second)
		_, err := srv.Start(ctx, roomCode, hostPlayerID, deadline)
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
				db.Room{ID: roomID,
					GameName:   gameName,
					HostPlayer: hostPlayerID,
					RoomState:  db.ROOMSTATE_CREATED.String(),
				}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, hostPlayerID).Return([]db.GetAllPlayersInRoomRow{
			{
				ID:         defaultNewPlayer.ID,
				Nickname:   "Hello",
				HostPlayer: hostPlayerID,
				IsReady:    pgtype.Bool{Bool: true, Valid: true},
				RoomCode:   roomCode,
			},
		}, nil)

		deadline := time.Now().Add(5 * time.Second)
		_, err := srv.Start(ctx, roomCode, hostPlayerID, deadline)
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
				db.Room{
					ID:         roomID,
					GameName:   gameName,
					HostPlayer: hostPlayerID,
					RoomState:  db.ROOMSTATE_CREATED.String(),
				},
				nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, hostPlayerID).Return([]db.GetAllPlayersInRoomRow{
			{
				ID:         defaultNewPlayer.ID,
				Nickname:   "Hello",
				HostPlayer: hostPlayerID,
				IsReady:    pgtype.Bool{Bool: true, Valid: true},
				RoomCode:   roomCode,
			},
			{
				ID:         hostPlayerID,
				Nickname:   "EmotionalTiger",
				HostPlayer: hostPlayerID,
				IsReady:    pgtype.Bool{Bool: false, Valid: true},
				RoomCode:   roomCode,
			},
		}, nil)

		deadline := time.Now().Add(5 * time.Second)
		_, err := srv.Start(ctx, roomCode, hostPlayerID, deadline)
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
				db.Room{
					ID:         roomID,
					GameName:   gameName,
					HostPlayer: hostPlayerID,
					RoomState:  db.ROOMSTATE_CREATED.String(),
				}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, hostPlayerID).Return([]db.GetAllPlayersInRoomRow{
			{
				ID:         defaultNewPlayer.ID,
				Nickname:   "Hello",
				HostPlayer: hostPlayerID,
				IsReady:    pgtype.Bool{Bool: true, Valid: true},
				RoomCode:   roomCode,
			},
			{
				ID:         hostPlayerID,
				Nickname:   "EmotionalTiger",
				HostPlayer: hostPlayerID,
				IsReady:    pgtype.Bool{Bool: true, Valid: true},
				RoomCode:   roomCode,
			},
		}, nil)

		mockStore.EXPECT().GetRandomQuestionByRound(ctx, db.GetRandomQuestionByRoundParams{
			GameName:     gameName,
			LanguageCode: "en-GB",
			Round:        "free_form",
		}).Return(db.Question{}, fmt.Errorf("failed to get random question for normals"))

		deadline := time.Now().Add(5 * time.Second)
		_, err := srv.Start(ctx, roomCode, hostPlayerID, deadline)
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
				db.Room{ID: roomID,
					GameName:   gameName,
					HostPlayer: hostPlayerID,
					RoomState:  db.ROOMSTATE_CREATED.String(),
				},
				nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, hostPlayerID).Return([]db.GetAllPlayersInRoomRow{
			{
				ID:         defaultNewPlayer.ID,
				Nickname:   "Hello",
				HostPlayer: hostPlayerID,
				IsReady:    pgtype.Bool{Bool: true, Valid: true},
				RoomCode:   roomCode,
			},
			{
				ID:         hostPlayerID,
				Nickname:   "EmotionalTiger",
				HostPlayer: hostPlayerID,
				IsReady:    pgtype.Bool{Bool: true, Valid: true},
				RoomCode:   roomCode,
			},
		}, nil)

		mockStore.EXPECT().GetRandomQuestionByRound(ctx, db.GetRandomQuestionByRoundParams{
			GameName:     gameName,
			LanguageCode: "en-GB",
			Round:        "free_form",
		}).Return(db.Question{
			ID:       uuid.MustParse("0193a629-7dcc-78ad-822f-fd5d83c89ae7"),
			Question: "What is the capital of France?",
			GroupID:  groupID,
		}, nil)
		mockStore.EXPECT().GetRandomQuestionInGroup(ctx, db.GetRandomQuestionInGroupParams{
			GroupID: groupID,
			ID:      uuid.MustParse("0193a629-7dcc-78ad-822f-fd5d83c89ae7"),
		}).Return(db.GetRandomQuestionInGroupRow{}, fmt.Errorf("failed to get random question for fibber"))

		deadline := time.Now().Add(5 * time.Second)
		_, err := srv.Start(ctx, roomCode, hostPlayerID, deadline)
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
				db.Room{
					ID:         roomID,
					GameName:   gameName,
					HostPlayer: hostPlayerID,
					RoomState:  db.ROOMSTATE_CREATED.String(),
				},
				nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, hostPlayerID).Return([]db.GetAllPlayersInRoomRow{
			{
				ID:         defaultNewPlayer.ID,
				Nickname:   "Hello",
				HostPlayer: hostPlayerID,
				IsReady:    pgtype.Bool{Bool: true, Valid: true},
				RoomCode:   roomCode,
			},
			{
				ID:         hostPlayerID,
				Nickname:   "EmotionalTiger",
				HostPlayer: hostPlayerID,
				IsReady:    pgtype.Bool{Bool: true, Valid: true},
				RoomCode:   roomCode,
			},
		}, nil)

		mockStore.EXPECT().GetRandomQuestionByRound(ctx, db.GetRandomQuestionByRoundParams{
			GameName:     gameName,
			LanguageCode: "en-GB",
			Round:        "free_form",
		}).Return(db.Question{
			ID:       uuid.MustParse("0193a629-7dcc-78ad-822f-fd5d83c89ae7"),
			Question: "What is the capital of France?",
			GroupID:  groupID,
		}, nil)
		mockStore.EXPECT().GetRandomQuestionInGroup(ctx, db.GetRandomQuestionInGroupParams{
			GroupID: groupID,
			ID:      uuid.MustParse("0193a629-7dcc-78ad-822f-fd5d83c89ae7"),
		}).Return(db.GetRandomQuestionInGroupRow{
			ID:       uuid.MustParse("0193a629-a9ac-7fc4-828c-a1334c282e0f"),
			Question: "What is the capital of Germany?",
		}, nil)
		mockRandom.EXPECT().GetFibberIndex(2).Return(1)
		mockRandom.EXPECT().GetID().Return(gameStateID)
		deadline := time.Now().Add(5 * time.Second)
		mockStore.EXPECT().StartGame(ctx, db.StartGameArgs{
			GameStateID:       gameStateID,
			RoomID:            roomID,
			NormalsQuestionID: uuid.MustParse("0193a629-7dcc-78ad-822f-fd5d83c89ae7"),
			FibberQuestionID:  uuid.MustParse("0193a629-a9ac-7fc4-828c-a1334c282e0f"),
			Players: []db.GetAllPlayersInRoomRow{
				{
					ID:         defaultNewPlayer.ID,
					Nickname:   "Hello",
					IsReady:    pgtype.Bool{Bool: true, Valid: true},
					RoomCode:   roomCode,
					HostPlayer: hostPlayerID,
				},
				{
					ID:         hostPlayerID,
					Nickname:   "EmotionalTiger",
					IsReady:    pgtype.Bool{Bool: true, Valid: true},
					RoomCode:   roomCode,
					HostPlayer: hostPlayerID,
				},
			},
			FibberLoc: 1,
			Deadline:  deadline,
		}).Return(fmt.Errorf("failed to start game"))

		_, err := srv.Start(ctx, roomCode, hostPlayerID, deadline)
		assert.Error(t, err)
	})
}

func TestLobbyServiceGetRoomState(t *testing.T) {
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
			lobbyService := service.NewLobbyService(mockStore, mockRandomizer)

			ctx := context.Background()
			mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(db.Room{
				RoomState: tt.roomState.String(),
			}, nil)

			roomState, err := lobbyService.GetRoomState(ctx, playerID)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedState, roomState)
		})
	}

	t.Run("Should fail to get room state because we fail to get room details DB", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandomizer)

		ctx := context.Background()
		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(
			db.Room{}, fmt.Errorf("failed to get room details"),
		)

		_, err := srv.GetRoomState(ctx, playerID)
		assert.Error(t, err)
	})
}

func TestLobbyerviceGetLobby(t *testing.T) {
	t.Run("Should successfully get lobby", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandomizer)

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
		srv := service.NewLobbyService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetAllPlayersInRoom(ctx, playerID).Return(
			[]db.GetAllPlayersInRoomRow{}, fmt.Errorf("failed to get players in room"),
		)

		_, err := srv.GetLobby(ctx, playerID)
		assert.Error(t, err)
	})
}
