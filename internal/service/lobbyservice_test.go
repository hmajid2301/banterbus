package service_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mdobak/go-xerrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/hmajid2301/banterbus/internal/service"
	mockService "gitlab.com/hmajid2301/banterbus/internal/service/mocks"
	"gitlab.com/hmajid2301/banterbus/internal/store/db"
)

var (
	defaultHostPlayerID        = uuid.Must(uuid.FromString("0193a62a-4dff-774c-850a-b1fe78e2a8d1"))
	defaultOtherPlayerID       = uuid.Must(uuid.FromString("0193a62a-4dff-774c-850a-b1fe78e2a8d2"))
	roomID                     = uuid.Must(uuid.FromString("0193a62a-4dff-774c-850a-b1fe78e2a8d3"))
	roundID                    = uuid.Must(uuid.FromString("0193a62a-4dff-774c-850a-b1fe78e2a8d4"))
	hostPlayerID               = uuid.Must(uuid.FromString("0193a62a-4dff-774c-850a-b1fe78e2a8d5"))
	playerID                   = uuid.Must(uuid.FromString("0193a625-dad1-7095-9abb-bebdad739381"))
	roomCode                   = "ABCD"
	gameName                   = "fibbing_it"
	defaultHostNickname        = "Host Player"
	defaultOtherPlayerNickname = "Other Player"
)

var (
	defaultNewPlayer = service.NewPlayer{
		ID:       defaultHostPlayerID,
		Nickname: "Host Player",
		Avatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Host Player",
	}

	defaultNewHostPlayer = service.NewHostPlayer{
		ID:       defaultHostPlayerID,
		Nickname: "Host Player",
	}
)

var hostPlayer = service.NewHostPlayer{
	ID:       defaultHostPlayerID,
	Nickname: "test",
}

var testHostPlayer = service.NewHostPlayer{
	ID:       defaultHostPlayerID,
	Nickname: defaultHostNickname,
}

func TestLobbyServiceCreate(t *testing.T) {
	t.Parallel()

	t.Run("Should create room successfully", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom, "en-GB")

		ctx, err := getI18nCtx(t.Context())
		require.NoError(t, err)

		mockRandom.EXPECT().GetAvatar(defaultNewPlayer.Nickname).Return(defaultNewPlayer.Avatar)
		mockRandom.EXPECT().GetRoomCode().Return(roomCode)
		mockStore.EXPECT().GetRoomByCode(ctx, roomCode).Return(db.Room{}, sql.ErrNoRows)
		mockRandom.EXPECT().GetID().Return(roomID, nil)

		createRoom := getCreateRoomParams(defaultNewPlayer, defaultNewHostPlayer)
		mockStore.EXPECT().CreateRoom(ctx, createRoom).Return(nil)
		result, err := srv.Create(ctx, gameName, defaultNewHostPlayer)

		expectedResult := service.LobbyCreationResult{
			Lobby: service.Lobby{
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
			},
			NewPlayerID: defaultNewPlayer.ID,
		}

		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
	})

	t.Run("Should create room when nickname is passed", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom, "en-GB")

		newHostPlayer := service.NewHostPlayer{
			Nickname: "MyNickname",
			ID:       uuid.Must(uuid.FromString("0193a626-2586-7784-9b5b-104d927d64ca")),
		}

		newPlayer := service.NewPlayer{
			ID:       newHostPlayer.ID,
			Nickname: newHostPlayer.Nickname,
			Avatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=MyNickname",
		}

		ctx, err := getI18nCtx(t.Context())
		require.NoError(t, err)

		mockRandom.EXPECT().GetAvatar(newHostPlayer.Nickname).Return(newPlayer.Avatar)
		mockRandom.EXPECT().GetRoomCode().Return(roomCode)
		mockStore.EXPECT().GetRoomByCode(ctx, roomCode).Return(db.Room{}, sql.ErrNoRows)
		mockRandom.EXPECT().GetID().Return(roomID, nil)

		createRoom := getCreateRoomParams(newPlayer, newHostPlayer)
		mockStore.EXPECT().CreateRoom(ctx, createRoom).Return(nil)
		result, err := srv.Create(ctx, gameName, newHostPlayer)

		expectedResult := service.LobbyCreationResult{
			Lobby: service.Lobby{
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
			},
			NewPlayerID: newPlayer.ID,
		}

		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
	})

	t.Run("Should create room successfully, code is used by room", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom, "en-GB")

		ctx, err := getI18nCtx(t.Context())
		require.NoError(t, err)

		mockRandom.EXPECT().GetAvatar(defaultNewPlayer.Nickname).Return(defaultNewPlayer.Avatar)
		mockRandom.EXPECT().GetRoomCode().Return(roomCode)
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(db.Room{RoomState: db.Created.String()}, nil).
			Times(1)
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(db.Room{RoomState: db.Abandoned.String()}, nil).
			Times(1)
		mockRandom.EXPECT().GetID().Return(roomID, nil)

		createRoom := getCreateRoomParams(defaultNewPlayer, defaultNewHostPlayer)
		mockStore.EXPECT().CreateRoom(ctx, createRoom).Return(nil)
		result, err := srv.Create(ctx, gameName, defaultNewHostPlayer)

		expectedResult := service.LobbyCreationResult{
			Lobby: service.Lobby{
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
			},
			NewPlayerID: defaultNewPlayer.ID,
		}

		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
	})

	t.Run("Should throw error when we fail to get room code", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom, "en-GB")

		ctx, err := getI18nCtx(t.Context())
		require.NoError(t, err)

		mockRandom.EXPECT().GetAvatar(defaultNewPlayer.Nickname).Return(defaultNewPlayer.Avatar)
		mockRandom.EXPECT().GetRoomCode().Return(roomCode)
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(db.Room{}, xerrors.New("failed to get room code")).
			Times(1)
		_, err = srv.Create(ctx, gameName, defaultNewHostPlayer)
		assert.Error(t, err)
	})
}

func getCreateRoomParams(newCreatedPlayer service.NewPlayer, newPlayer service.NewHostPlayer) db.CreateRoomArgs {
	addPlayer := db.AddPlayerParams{
		ID:       newCreatedPlayer.ID,
		Avatar:   newCreatedPlayer.Avatar,
		Nickname: newCreatedPlayer.Nickname,
		Locale:   pgtype.Text{String: "en-GB"},
	}
	addRoom := db.AddRoomParams{
		ID:         roomID,
		GameName:   gameName,
		RoomCode:   roomCode,
		RoomState:  db.Created.String(),
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
	t.Parallel()

	t.Run("Should allow player to join room successfully", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom, "en-GB")

		ctx, err := getI18nCtx(t.Context())
		require.NoError(t, err)

		generatedPlayerID := uuid.Must(uuid.NewV4())
		generatedAvatar := "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=" + generatedPlayerID.String()
		playerNickname := "example"

		mockRandom.EXPECT().GetID().Return(generatedPlayerID, nil).Once()
		mockStore.EXPECT().
			GetRoomByPlayerID(ctx, generatedPlayerID).
			Return(db.Room{}, sql.ErrNoRows)
		mockRandom.EXPECT().GetAvatar(playerNickname).Return(generatedAvatar).Once()
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(db.Room{ID: roomID, RoomState: db.Created.String()}, nil)
		mockStore.EXPECT().GetAllPlayerByRoomCode(ctx, roomCode).Return([]db.GetAllPlayerByRoomCodeRow{
			{
				Nickname: "Hello",
			},
		}, nil)

		addPlayer := db.AddPlayerParams{
			ID:       generatedPlayerID,
			Avatar:   generatedAvatar,
			Nickname: playerNickname,
			Locale:   pgtype.Text{String: "en-GB"},
		}

		addRoomPlayer := db.AddRoomPlayerParams{
			RoomID:   roomID,
			PlayerID: generatedPlayerID,
		}

		addPlayerToRoom := db.AddPlayerToRoomArgs{
			Player:     addPlayer,
			RoomPlayer: addRoomPlayer,
		}
		mockStore.EXPECT().AddPlayerToRoom(ctx, addPlayerToRoom).Return(nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, generatedPlayerID).Return([]db.GetAllPlayersInRoomRow{
			{
				ID:         uuid.Must(uuid.FromString("0193a628-51bc-7a60-9204-a8667771f278")),
				Nickname:   "EmotionalTiger",
				Avatar:     "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=EmotionalTiger",
				IsReady:    pgtype.Bool{Bool: false, Valid: true},
				HostPlayer: uuid.Must(uuid.FromString("0193a628-8b7b-7ad9-aefe-031ec85289fa")),
			},
			{
				ID:         uuid.Must(uuid.FromString("0193a628-8b7b-7ad9-aefe-031ec85289fa")),
				Nickname:   "Hello",
				Avatar:     "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Hello",
				IsReady:    pgtype.Bool{Bool: false, Valid: true},
				HostPlayer: uuid.Must(uuid.FromString("0193a628-8b7b-7ad9-aefe-031ec85289fa")),
			},
			{
				ID:         generatedPlayerID,
				Nickname:   playerNickname,
				Avatar:     generatedAvatar,
				IsReady:    pgtype.Bool{Bool: false, Valid: true},
				HostPlayer: uuid.Must(uuid.FromString("0193a628-8b7b-7ad9-aefe-031ec85289fa")),
				RoomCode:   roomCode,
			},
		}, nil)
		result, err := srv.Join(ctx, roomCode, uuid.Nil, playerNickname)
		expectedResult := service.LobbyJoinResult{
			Lobby: service.Lobby{
				Code: roomCode,
				Players: []service.LobbyPlayer{
					{
						ID:       uuid.Must(uuid.FromString("0193a628-51bc-7a60-9204-a8667771f278")),
						Nickname: "EmotionalTiger",
						Avatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=EmotionalTiger",
						IsHost:   false,
						IsReady:  false,
					},
					{
						ID:       uuid.Must(uuid.FromString("0193a628-8b7b-7ad9-aefe-031ec85289fa")),
						Nickname: "Hello",
						Avatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Hello",
						IsReady:  false,
						IsHost:   true,
					},
					{
						ID:       generatedPlayerID,
						Nickname: playerNickname,
						Avatar:   generatedAvatar,
						IsHost:   false,
						IsReady:  false,
					},
				},
			},
			NewPlayerID: generatedPlayerID,
		}
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
	})

	t.Run("Should allow player to join room successfully, when they pass nickname", func(t *testing.T) {
		t.Parallel()
		nickname := "MyNickname"
		playerID := uuid.Must(uuid.NewV4())
		playerAvatar := "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=" + nickname

		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom, "en-GB")

		ctx, err := getI18nCtx(t.Context())
		require.NoError(t, err)

		mockStore.EXPECT().
			GetRoomByPlayerID(ctx, playerID).
			Return(db.Room{}, sql.ErrNoRows)
		mockRandom.EXPECT().
			GetAvatar(nickname).
			Return(playerAvatar).Once()
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(db.Room{ID: roomID, RoomState: db.Created.String()}, nil)
		mockStore.EXPECT().GetAllPlayerByRoomCode(ctx, roomCode).Return([]db.GetAllPlayerByRoomCodeRow{
			{
				Nickname: "Hello",
			},
		}, nil)

		addPlayer := db.AddPlayerParams{
			ID:       playerID,
			Avatar:   playerAvatar,
			Nickname: nickname,
			Locale:   pgtype.Text{String: "en-GB"},
		}

		addRoomPlayer := db.AddRoomPlayerParams{
			RoomID:   roomID,
			PlayerID: playerID,
		}

		addPlayerToRoom := db.AddPlayerToRoomArgs{
			Player:     addPlayer,
			RoomPlayer: addRoomPlayer,
		}
		mockStore.EXPECT().AddPlayerToRoom(ctx, addPlayerToRoom).Return(nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, playerID).Return([]db.GetAllPlayersInRoomRow{
			{
				ID:         uuid.Must(uuid.FromString("0193a628-51bc-7a60-9204-a8667771f278")),
				Nickname:   nickname,
				Avatar:     playerAvatar,
				IsReady:    pgtype.Bool{Bool: false, Valid: true},
				HostPlayer: uuid.Must(uuid.FromString("0193a628-8b7b-7ad9-aefe-031ec85289fa")),
			},
			{
				ID:         uuid.Must(uuid.FromString("0193a628-8b7b-7ad9-aefe-031ec85289fa")),
				Nickname:   "Hello",
				Avatar:     "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Hello",
				IsReady:    pgtype.Bool{Bool: false, Valid: true},
				HostPlayer: uuid.Must(uuid.FromString("0193a628-8b7b-7ad9-aefe-031ec85289fa")),
			},
		}, nil)
		result, err := srv.Join(ctx, roomCode, playerID, nickname)
		expectedResult := service.LobbyJoinResult{
			Lobby: service.Lobby{
				Code: roomCode,
				Players: []service.LobbyPlayer{
					{
						ID:       uuid.Must(uuid.FromString("0193a628-51bc-7a60-9204-a8667771f278")),
						Nickname: nickname,
						Avatar:   playerAvatar,
						IsHost:   false,
						IsReady:  false,
					},
					{
						ID:       uuid.Must(uuid.FromString("0193a628-8b7b-7ad9-aefe-031ec85289fa")),
						Nickname: "Hello",
						Avatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Hello",
						IsReady:  false,
						IsHost:   true,
					},
				},
			},
			NewPlayerID: playerID,
		}
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
	})

	t.Run("Should fail to join room, player id alredy in room", func(t *testing.T) {
		t.Parallel()
		playerID := uuid.Must(uuid.NewV4()) // Specific player ID for this test
		playerNickname := "some_nickname"

		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(db.Room{RoomCode: roomCode}, nil).Once()

		_, err := srv.Join(ctx, roomCode, playerID, playerNickname) // Pass specific playerID
		assert.ErrorIs(t, err, service.ErrPlayerAlreadyInRoom)
	})

	t.Run("Should fail to join room, not in CREATED room state", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		generatedPlayerID := uuid.Must(uuid.NewV4())
		playerNickname := ""
		generatedAvatar := "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=" + playerNickname

		mockRandom.EXPECT().GetID().Return(generatedPlayerID, nil).Once()
		mockStore.EXPECT().
			GetRoomByPlayerID(ctx, generatedPlayerID).
			Return(db.Room{}, sql.ErrNoRows)
		mockRandom.EXPECT().GetAvatar(playerNickname).Return(generatedAvatar).Once()
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(db.Room{ID: roomID, RoomState: db.Playing.String()}, nil)

		_, err := srv.Join(ctx, roomCode, uuid.Nil, playerNickname) // Pass uuid.Nil
		assert.ErrorContains(t, err, "room is not in CREATED state")
	})

	t.Run("Should fail to join room, nickname already exists", func(t *testing.T) {
		t.Parallel()
		nickname := "Hello"

		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		generatedPlayerID := uuid.Must(uuid.NewV4())
		generatedAvatar := "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=" + nickname

		mockRandom.EXPECT().GetID().Return(generatedPlayerID, nil).Once()
		mockStore.EXPECT().
			GetRoomByPlayerID(ctx, generatedPlayerID).
			Return(db.Room{}, sql.ErrNoRows)
		mockRandom.EXPECT().GetAvatar(nickname).Return(generatedAvatar).Once()
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(db.Room{ID: roomID, RoomState: db.Created.String()}, nil)
		mockStore.EXPECT().GetAllPlayerByRoomCode(ctx, roomCode).Return([]db.GetAllPlayerByRoomCodeRow{
			{
				Nickname: "Hello",
			},
		}, nil)

		_, err := srv.Join(ctx, roomCode, uuid.Nil, nickname) // Pass uuid.Nil
		assert.ErrorContains(t, err, "nickname already exists in room")
	})

	t.Run("Should fail to join room because we fail to add player to DB", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom, "en-GB")

		ctx, err := getI18nCtx(t.Context())
		require.NoError(t, err)

		generatedPlayerID := uuid.Must(uuid.NewV4())
		playerNickname := ""
		generatedAvatar := "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=" + playerNickname

		mockRandom.EXPECT().GetID().Return(generatedPlayerID, nil).Once()
		mockStore.EXPECT().
			GetRoomByPlayerID(ctx, generatedPlayerID).
			Return(db.Room{}, sql.ErrNoRows)
		mockRandom.EXPECT().GetAvatar(playerNickname).Return(generatedAvatar).Once()
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(db.Room{ID: roomID, RoomState: db.Created.String()}, nil)
		mockStore.EXPECT().GetAllPlayerByRoomCode(ctx, roomCode).Return([]db.GetAllPlayerByRoomCodeRow{
			{
				Nickname: "Hello",
			},
		}, nil)

		addPlayer := db.AddPlayerParams{
			ID:       generatedPlayerID,
			Avatar:   generatedAvatar,
			Nickname: playerNickname,
			Locale:   pgtype.Text{String: "en-GB"},
		}

		addRoomPlayer := db.AddRoomPlayerParams{
			RoomID:   roomID,
			PlayerID: generatedPlayerID,
		}

		addPlayerToRoom := db.AddPlayerToRoomArgs{
			Player:     addPlayer,
			RoomPlayer: addRoomPlayer,
		}
		mockStore.EXPECT().AddPlayerToRoom(ctx, addPlayerToRoom).Return(xerrors.New("failed to add player to room"))

		_, err = srv.Join(ctx, roomCode, uuid.Nil, playerNickname) // Pass uuid.Nil
		assert.Error(t, err)
	})

	t.Run("Should fail to join room because fail to get all players in room", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom, "en-GB")

		ctx, err := getI18nCtx(t.Context())
		require.NoError(t, err)

		generatedPlayerID := uuid.Must(uuid.NewV4())
		playerNickname := ""
		generatedAvatar := "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=" + playerNickname

		mockRandom.EXPECT().GetID().Return(generatedPlayerID, nil).Once()
		mockStore.EXPECT().
			GetRoomByPlayerID(ctx, generatedPlayerID).
			Return(db.Room{}, sql.ErrNoRows)
		mockRandom.EXPECT().GetAvatar(playerNickname).Return(generatedAvatar).Once()
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(db.Room{ID: roomID, RoomState: db.Created.String()}, nil)
		mockStore.EXPECT().GetAllPlayerByRoomCode(ctx, roomCode).Return([]db.GetAllPlayerByRoomCodeRow{
			{
				Nickname: "Hello",
			},
		}, nil)

		addPlayer := db.AddPlayerParams{
			ID:       generatedPlayerID,
			Avatar:   generatedAvatar,
			Nickname: playerNickname,
			Locale:   pgtype.Text{String: "en-GB"},
		}

		addRoomPlayer := db.AddRoomPlayerParams{
			RoomID:   roomID,
			PlayerID: generatedPlayerID,
		}

		addPlayerToRoom := db.AddPlayerToRoomArgs{
			Player:     addPlayer,
			RoomPlayer: addRoomPlayer,
		}
		mockStore.EXPECT().AddPlayerToRoom(ctx, addPlayerToRoom).Return(nil)
		mockStore.EXPECT().
			GetAllPlayersInRoom(ctx, generatedPlayerID).
			Return([]db.GetAllPlayersInRoomRow{}, xerrors.New("failed to get all players in room"))
		_, err = srv.Join(ctx, roomCode, uuid.Nil, playerNickname) // Pass uuid.Nil
		assert.Error(t, err)
	})
}

func TestLobbyServiceKickPlayer(t *testing.T) {
	t.Parallel()

	defaultNewPlayer := service.NewPlayer{
		ID:       uuid.Must(uuid.FromString("0193a626-2586-7784-9b5b-104d927d64ca")),
		Nickname: "Hello",
		Avatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Hello",
	}

	t.Run("Should kick player from lobby successfully", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(db.Room{ID: roomID, HostPlayer: hostPlayerID, RoomState: db.Created.String()}, nil)
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
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		mockStore.EXPECT().GetRoomByCode(ctx, roomCode).Return(db.Room{}, xerrors.New("failed to get room by code"))
		_, _, err := srv.KickPlayer(ctx, roomCode, hostPlayerID, defaultNewPlayer.Nickname)
		assert.Error(t, err)
	})

	t.Run("Should fail to kick player because player is not host", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(db.Room{ID: roomID, HostPlayer: hostPlayerID, RoomState: db.Created.String()}, nil)
		_, _, err := srv.KickPlayer(ctx, roomCode, defaultNewPlayer.ID, defaultNewPlayer.Nickname)
		assert.ErrorContains(t, err, "player is not the host of the room")
	})

	t.Run("Should fail to kick player because room is not in CREATED state", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(db.Room{ID: roomID, HostPlayer: hostPlayerID, RoomState: db.Playing.String()}, nil)
		_, _, err := srv.KickPlayer(ctx, roomCode, hostPlayerID, defaultNewPlayer.Nickname)
		assert.ErrorContains(t, err, "room is not in CREATED state")
	})

	t.Run("Should fail to kick player because we fail to get all players in room from DB", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(db.Room{ID: roomID, HostPlayer: hostPlayerID, RoomState: db.Created.String()}, nil)
		mockStore.EXPECT().
			GetAllPlayersInRoom(ctx, hostPlayerID).
			Return([]db.GetAllPlayersInRoomRow{}, xerrors.New("failed to get all players in room"))

		_, _, err := srv.KickPlayer(ctx, roomCode, hostPlayerID, defaultNewPlayer.Nickname)
		assert.Error(t, err)
	})

	t.Run("Should fail to kick player because player with nickname is not in room", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(db.Room{ID: roomID, HostPlayer: hostPlayerID, RoomState: db.Created.String()}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, hostPlayerID).Return([]db.GetAllPlayersInRoomRow{
			{
				ID:         hostPlayerID,
				Nickname:   "EmotionalTiger",
				HostPlayer: hostPlayerID,
			},
		}, nil)

		_, _, err := srv.KickPlayer(ctx, roomCode, hostPlayerID, defaultNewPlayer.Nickname)
		assert.ErrorContains(t, err, "player with nickname Hello not found")
	})

	t.Run("Should fail to kick player because failed to remove player from room in DB", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(db.Room{ID: roomID, HostPlayer: hostPlayerID, RoomState: db.Created.String()}, nil)
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
			Return(db.RoomsPlayer{}, xerrors.New("failed to remove player from room"))
		_, _, err := srv.KickPlayer(ctx, roomCode, hostPlayerID, defaultNewPlayer.Nickname)
		assert.Error(t, err)
	})
}

func TestLobbyServiceStart(t *testing.T) {
	t.Parallel()

	groupID := uuid.Must(uuid.FromString("0193a629-1fcf-79dd-ac70-760bedbdffa9"))
	gameStateID := uuid.Must(uuid.FromString("0193a629-373b-7a3e-b6c2-0e7d2f95ce43"))

	defaultNewPlayer := service.NewPlayer{
		ID:       uuid.Must(uuid.FromString("0193a626-2586-7784-9b5b-104d927d64ca")),
		Nickname: "Hello",
		Avatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Hello",
	}

	t.Run("Should start game successfully", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(
				db.Room{
					ID:         roomID,
					GameName:   gameName,
					HostPlayer: hostPlayerID,
					RoomState:  db.Created.String(),
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
			GameName:  gameName,
			RoundType: "free_form",
		}).Return([]db.GetRandomQuestionByRoundRow{
			{
				QuestionID: uuid.Must(uuid.FromString("0193a629-7dcc-78ad-822f-fd5d83c89ae7")),
				Question:   "What is the capital of France?",
				Locale:     "en-GB",
				GroupID:    groupID,
			},
		}, nil)
		mockStore.EXPECT().GetRandomQuestionInGroup(ctx, db.GetRandomQuestionInGroupParams{
			GroupType:          "",
			GroupID:            groupID,
			ExcludedQuestionID: uuid.Must(uuid.FromString("0193a629-7dcc-78ad-822f-fd5d83c89ae7")),
			RoundType:          "free_form",
		}).Return([]db.GetRandomQuestionInGroupRow{
			{
				QuestionID: uuid.Must(uuid.FromString("0193a629-a9ac-7fc4-828c-a1334c282e0f")),
				Question:   "What is the capital of Germany?",
			},
		}, nil)
		mockRandom.EXPECT().GetFibberIndex(2).Return(1)
		mockRandom.EXPECT().GetID().Return(gameStateID, nil)
		deadline := time.Now().Add(5 * time.Second)
		mockStore.EXPECT().StartGame(ctx, db.StartGameArgs{
			GameStateID:       gameStateID,
			RoomID:            roomID,
			NormalsQuestionID: uuid.Must(uuid.FromString("0193a629-7dcc-78ad-822f-fd5d83c89ae7")),
			FibberQuestionID:  uuid.Must(uuid.FromString("0193a629-a9ac-7fc4-828c-a1334c282e0f")),
			Players: []db.GetAllPlayersInRoomRow{
				{
					ID:         uuid.Must(uuid.FromString("0193a626-2586-7784-9b5b-104d927d64ca")),
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
			},
			FibberLoc: 1,
			Deadline:  deadline,
		}).Return(nil)

		gameState, err := srv.Start(ctx, roomCode, hostPlayerID, deadline)
		expectedGameState := service.QuestionState{
			GameStateID: gameStateID,
			Players: []service.PlayerWithRole{
				{
					ID:       uuid.Must(uuid.FromString("0193a626-2586-7784-9b5b-104d927d64ca")),
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
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(
				db.Room{ID: roomID,
					GameName:   gameName,
					HostPlayer: hostPlayerID,
					RoomState:  db.Created.String(),
				}, nil)

		deadline := time.Now().Add(5 * time.Second)
		_, err := srv.Start(ctx, roomCode, defaultNewPlayer.ID, deadline)
		assert.ErrorContains(t, err, "")
	})

	t.Run("Should fail to start game because room state not in CREATED state", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(
				db.Room{ID: roomID,
					GameName:   gameName,
					HostPlayer: hostPlayerID,
					RoomState:  db.Playing.String(),
				}, nil)

		deadline := time.Now().Add(5 * time.Second)
		_, err := srv.Start(ctx, roomCode, hostPlayerID, deadline)
		assert.ErrorContains(t, err, "room is not in CREATED state")
	})

	t.Run("Should fail to start game because we failed to get all players in room", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(
				db.Room{ID: roomID,
					GameName:   gameName,
					HostPlayer: hostPlayerID,
					RoomState:  db.Created.String(),
				}, nil)
		mockStore.EXPECT().
			GetAllPlayersInRoom(ctx, hostPlayerID).
			Return([]db.GetAllPlayersInRoomRow{}, xerrors.New("failed to get all players in room"))

		deadline := time.Now().Add(5 * time.Second)
		_, err := srv.Start(ctx, roomCode, hostPlayerID, deadline)
		assert.Error(t, err)
	})

	t.Run("Should fail to start game because too few players in room", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(
				db.Room{ID: roomID,
					GameName:   gameName,
					HostPlayer: hostPlayerID,
					RoomState:  db.Created.String(),
				}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, hostPlayerID).Return([]db.GetAllPlayersInRoomRow{
			{
				ID:         uuid.Must(uuid.FromString("0193a626-2586-7784-9b5b-104d927d64ca")),
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
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(
				db.Room{ID: roomID,
					GameName:   gameName,
					HostPlayer: hostPlayerID,
					RoomState:  db.Created.String(),
				},
				nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, hostPlayerID).Return([]db.GetAllPlayersInRoomRow{
			{
				ID:         uuid.Must(uuid.FromString("0193a626-2586-7784-9b5b-104d927d64ca")),
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
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(
				db.Room{ID: roomID,
					GameName:   gameName,
					HostPlayer: hostPlayerID,
					RoomState:  db.Created.String(),
				}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, hostPlayerID).Return([]db.GetAllPlayersInRoomRow{
			{
				ID:         uuid.Must(uuid.FromString("0193a626-2586-7784-9b5b-104d927d64ca")),
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
			GameName:  gameName,
			RoundType: "free_form",
		}).Return([]db.GetRandomQuestionByRoundRow{
			{
				QuestionID: uuid.Must(uuid.FromString("0193a629-7dcc-78ad-822f-fd5d83c89ae7")),
				Question:   "What is the capital of France?",
				Locale:     "en-GB",
				GroupID:    groupID,
			},
		}, xerrors.New("failed to get random question for normals"))

		deadline := time.Now().Add(5 * time.Second)
		_, err := srv.Start(ctx, roomCode, hostPlayerID, deadline)
		assert.Error(t, err)
	})

	t.Run("Should fail to start game because we fail to get random question for fibber", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(
				db.Room{ID: roomID,
					GameName:   gameName,
					HostPlayer: hostPlayerID,
					RoomState:  db.Created.String(),
				},
				nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, hostPlayerID).Return([]db.GetAllPlayersInRoomRow{
			{
				ID:         uuid.Must(uuid.FromString("0193a626-2586-7784-9b5b-104d927d64ca")),
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
			GameName:  gameName,
			RoundType: "free_form",
		}).Return([]db.GetRandomQuestionByRoundRow{
			{
				QuestionID: uuid.Must(uuid.FromString("0193a629-7dcc-78ad-822f-fd5d83c89ae7")),
				Question:   "What is the capital of France?",
				Locale:     "en-GB",
				GroupID:    groupID,
			},
		}, nil)
		mockStore.EXPECT().GetRandomQuestionInGroup(ctx, db.GetRandomQuestionInGroupParams{
			GroupType:          "",
			GroupID:            groupID,
			ExcludedQuestionID: uuid.Must(uuid.FromString("0193a629-7dcc-78ad-822f-fd5d83c89ae7")),
			RoundType:          "free_form",
		}).Return([]db.GetRandomQuestionInGroupRow{}, xerrors.New("failed to get random question for fibber"))

		deadline := time.Now().Add(5 * time.Second)
		_, err := srv.Start(ctx, roomCode, hostPlayerID, deadline)
		assert.Error(t, err)
	})

	t.Run("Should fail to start game because we fail to start game in DB", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		mockStore.EXPECT().
			GetRoomByCode(ctx, roomCode).
			Return(
				db.Room{ID: roomID,
					GameName:   gameName,
					HostPlayer: hostPlayerID,
					RoomState:  db.Created.String(),
				},
				nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, hostPlayerID).Return([]db.GetAllPlayersInRoomRow{
			{
				ID:         uuid.Must(uuid.FromString("0193a626-2586-7784-9b5b-104d927d64ca")),
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
			GameName:  gameName,
			RoundType: "free_form",
		}).Return([]db.GetRandomQuestionByRoundRow{
			{
				QuestionID: uuid.Must(uuid.FromString("0193a629-7dcc-78ad-822f-fd5d83c89ae7")),
				Question:   "What is the capital of France?",
				Locale:     "en-GB",
				GroupID:    groupID,
			},
		}, nil)
		mockStore.EXPECT().GetRandomQuestionInGroup(ctx, db.GetRandomQuestionInGroupParams{
			GroupType:          "",
			GroupID:            groupID,
			ExcludedQuestionID: uuid.Must(uuid.FromString("0193a629-7dcc-78ad-822f-fd5d83c89ae7")),
			RoundType:          "free_form",
		}).Return([]db.GetRandomQuestionInGroupRow{
			{
				QuestionID: uuid.Must(uuid.FromString("0193a629-a9ac-7fc4-828c-a1334c282e0f")),
				Question:   "What is the capital of Germany?",
			},
		}, nil)
		mockRandom.EXPECT().GetFibberIndex(2).Return(1)
		mockRandom.EXPECT().GetID().Return(gameStateID, nil)
		deadline := time.Now().Add(5 * time.Second)
		mockStore.EXPECT().StartGame(ctx, db.StartGameArgs{
			GameStateID:       gameStateID,
			RoomID:            roomID,
			NormalsQuestionID: uuid.Must(uuid.FromString("0193a629-7dcc-78ad-822f-fd5d83c89ae7")),
			FibberQuestionID:  uuid.Must(uuid.FromString("0193a629-a9ac-7fc4-828c-a1334c282e0f")),
			Players: []db.GetAllPlayersInRoomRow{
				{
					ID:         uuid.Must(uuid.FromString("0193a626-2586-7784-9b5b-104d927d64ca")),
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
			},
			FibberLoc: 1,
			Deadline:  deadline,
		}).Return(xerrors.New("failed to start game"))

		_, err := srv.Start(ctx, roomCode, hostPlayerID, deadline)
		assert.Error(t, err)
	})
}

func TestLobbyServiceGetRoomState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		roomState     db.RoomState
		expectedState db.RoomState
	}{
		{
			name:          "Should successfully get room state CREATED",
			roomState:     db.Created,
			expectedState: db.Created,
		},
		{
			name:          "Should successfully get room state PLAYING",
			roomState:     db.Playing,
			expectedState: db.Playing,
		},
		{
			name:          "Should successfully get room state PAUSED",
			roomState:     db.Paused,
			expectedState: db.Paused,
		},
		{
			name:          "Should successfully get room state FINISHED",
			roomState:     db.Finished,
			expectedState: db.Finished,
		},
		{
			name:          "Should successfully get room state ABANDONED",
			roomState:     db.Abandoned,
			expectedState: db.Abandoned,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockStore := mockService.NewMockStorer(t)
			mockRandom := mockService.NewMockRandomizer(t)
			lobbyService := service.NewLobbyService(mockStore, mockRandom, "en-GB")

			ctx := t.Context()
			mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(db.Room{
				RoomState: tt.roomState.String(),
			}, nil)

			roomState, err := lobbyService.GetRoomState(ctx, playerID)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedState, roomState)
		})
	}

	t.Run("Should fail to get room state because we fail to get room details DB", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(
			db.Room{}, xerrors.New("failed to get room details"),
		)

		_, err := srv.GetRoomState(ctx, playerID)
		assert.Error(t, err)
	})
}

func TestLobbyServiceGetLobby(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully get lobby", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()

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
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewLobbyService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()

		mockStore.EXPECT().GetAllPlayersInRoom(ctx, playerID).Return(
			[]db.GetAllPlayersInRoomRow{}, xerrors.New("failed to get players in room"),
		)

		_, err := srv.GetLobby(ctx, playerID)
		assert.Error(t, err)
	})
}
