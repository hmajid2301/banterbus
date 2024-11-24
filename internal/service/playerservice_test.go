package service_test

// func TestUpdateNickname(t *testing.T) {
// 	t.Run("Should update nickname successfully", func(t *testing.T) {
// 		mockStore := mockService.NewMockStorer(t)
// 		mockRandom := mockService.NewMockRandomizer(t)
// 		service := service.NewPlayerService(mockStore, mockRandom)
//
// 		ctx := context.Background()
// 		mockStore.EXPECT().
// 			UpdateNickname(ctx, "new_nickname", "fbb75599-9f7a-4392-b523-fd433b3208ea").
// 			Return([]sqlc.GetAllPlayersInRoomRow{
// 				{
// 					ID:       "b75599-9f7a-4392-b523-fd433b3208ea",
// 					Nickname: "new_nickname",
// 					Avatar:   []byte(""),
// 					RoomCode: "ABC12",
// 				},
// 			}, nil)
//
// 		room, err := service.UpdateNickname(
// 			ctx,
// 			"new_nickname",
// 			"fbb75599-9f7a-4392-b523-fd433b3208ea",
// 		)
//
// 		assert.NoError(t, err)
// 		assert.Equal(t, "ABC12", room.Code)
// 		assert.Len(t, room.Players, 1)
// 		assert.NotEmpty(t, room.Players[0].Nickname)
// 	})
//
// 	t.Run("Should throw error when fail to update nickname in DB", func(t *testing.T) {
// 		mockStore := mockService.NewMockStorer(t)
// 		mockRandom := mockService.NewMockRandomizer(t)
// 		service := service.NewPlayerService(mockStore, mockRandom)
//
// 		ctx := context.Background()
// 		mockStore.EXPECT().
// 			UpdateNickname(ctx, "new_nickname", "fbb75599-9f7a-4392-b523-fd433b3208ea").
// 			Return([]sqlc.GetAllPlayersInRoomRow{}, fmt.Errorf("failed to update nickname"))
// 		_, err := service.UpdateNickname(
// 			ctx,
// 			"new_nickname",
// 			"fbb75599-9f7a-4392-b523-fd433b3208ea",
// 		)
// 		assert.Error(t, err)
// 	})
// }
//
// func TestGenerateAvatar(t *testing.T) {
// 	t.Run("Should generate new avatar successfully", func(t *testing.T) {
// 		mockStore := mockService.NewMockStorer(t)
// 		mockRandom := mockService.NewMockRandomizer(t)
// 		service := service.NewPlayerService(mockStore, mockRandom)
//
// 		ctx := context.Background()
// 		mockRandom.EXPECT().GetAvatar().Return([]byte("123"))
// 		mockStore.EXPECT().
// 			UpdateAvatar(ctx, []byte("123"), "fbb75599-9f7a-4392-b523-fd433b3208ea").
// 			Return([]sqlc.GetAllPlayersInRoomRow{
// 				{
// 					ID:       "fbb75599-9f7a-4392-b523-fd433b3208ea",
// 					Nickname: "Majiy00",
// 					Avatar:   []byte("123"),
// 					RoomCode: "ABC12",
// 				},
// 			}, nil)
//
// 		room, err := service.GenerateNewAvatar(ctx, "fbb75599-9f7a-4392-b523-fd433b3208ea")
//
// 		assert.NoError(t, err)
// 		assert.Equal(t, "ABC12", room.Code)
// 		assert.Len(t, room.Players, 1)
// 		assert.NotEmpty(t, room.Players[0].Avatar)
// 	})
//
// 	t.Run("Should throw error when fail to update nickname in DB", func(t *testing.T) {
// 		mockStore := mockService.NewMockStorer(t)
// 		mockRandom := mockService.NewMockRandomizer(t)
// 		service := service.NewPlayerService(mockStore, mockRandom)
//
// 		ctx := context.Background()
// 		mockRandom.EXPECT().GetAvatar().Return([]byte("123"))
// 		mockStore.EXPECT().
// 			UpdateAvatar(ctx, []byte("123"), "fbb75599-9f7a-4392-b523-fd433b3208ea").
// 			Return([]sqlc.GetAllPlayersInRoomRow{}, fmt.Errorf("failed to generate new avatar"))
// 		_, err := service.GenerateNewAvatar(ctx, "fbb75599-9f7a-4392-b523-fd433b3208ea")
// 		assert.Error(t, err)
// 	})
// }
//
// func TestToggleIsReady(t *testing.T) {
// 	t.Run("Should toggle player ready status successfully", func(t *testing.T) {
// 		mockStore := mockService.NewMockStorer(t)
// 		mockRandom := mockService.NewMockRandomizer(t)
// 		service := service.NewPlayerService(mockStore, mockRandom)
//
// 		ctx := context.Background()
// 		mockStore.EXPECT().
// 			ToggleIsReady(ctx, "fbb75599-9f7a-4392-b523-fd433b3208ea").
// 			Return([]sqlc.GetAllPlayersInRoomRow{
// 				{
// 					ID:       "fbb75599-9f7a-4392-b523-fd433b3208ea",
// 					Nickname: "Majiy00",
// 					Avatar:   []byte("123"),
// 					RoomCode: "ABC12",
// 					IsReady:  sql.NullBool{Bool: true, Valid: true},
// 				},
// 			}, nil)
//
// 		room, err := service.TogglePlayerIsReady(ctx, "fbb75599-9f7a-4392-b523-fd433b3208ea")
//
// 		assert.NoError(t, err)
// 		assert.Len(t, room.Players, 1)
// 		assert.True(t, room.Players[0].IsReady)
// 	})
//
// 	t.Run("Should throw error when fail to toggle player ready status in DB", func(t *testing.T) {
// 		mockStore := mockService.NewMockStorer(t)
// 		mockRandom := mockService.NewMockRandomizer(t)
// 		service := service.NewPlayerService(mockStore, mockRandom)
//
// 		ctx := context.Background()
// 		mockStore.EXPECT().
// 			ToggleIsReady(ctx, "fbb75599-9f7a-4392-b523-fd433b3208ea").
// 			Return([]sqlc.GetAllPlayersInRoomRow{}, fmt.Errorf("failed to update is ready status"))
// 		_, err := service.TogglePlayerIsReady(ctx, "fbb75599-9f7a-4392-b523-fd433b3208ea")
// 		assert.Error(t, err)
// 	})
// }
//
// func TestGetLobby(t *testing.T) {
// 	t.Run("Should lobby state using player id successfully", func(t *testing.T) {
// 		mockStore := mockService.NewMockStorer(t)
// 		mockRandom := mockService.NewMockRandomizer(t)
// 		service := service.NewPlayerService(mockStore, mockRandom)
//
// 		ctx := context.Background()
// 		mockStore.EXPECT().
// 			GetLobbyByPlayerID(ctx, "fbb75599-9f7a-4392-b523-fd433b3208ea").
// 			Return([]sqlc.GetAllPlayersInRoomRow{
// 				{
// 					ID:       "fbb75599-9f7a-4392-b523-fd433b3208ea",
// 					Nickname: "Majiy00",
// 					Avatar:   []byte("123"),
// 					RoomCode: "ABC12",
// 					IsReady:  sql.NullBool{Bool: true, Valid: true},
// 				},
// 			}, nil)
//
// 		lobby, err := service.GetLobby(ctx, "fbb75599-9f7a-4392-b523-fd433b3208ea")
//
// 		assert.NoError(t, err)
// 		assert.Len(t, lobby.Players, 1)
// 	})
//
// 	t.Run("Should throw error when store returns an error", func(t *testing.T) {
// 		mockStore := mockService.NewMockStorer(t)
// 		mockRandom := mockService.NewMockRandomizer(t)
// 		service := service.NewPlayerService(mockStore, mockRandom)
//
// 		ctx := context.Background()
// 		mockStore.EXPECT().
// 			GetLobbyByPlayerID(ctx, "fbb75599-9f7a-4392-b523-fd433b3208ea").
// 			Return([]sqlc.GetAllPlayersInRoomRow{}, fmt.Errorf("failed to get lobby status"))
// 		_, err := service.GetLobby(ctx, "fbb75599-9f7a-4392-b523-fd433b3208ea")
// 		assert.Error(t, err)
// 	})
// }
//
// func TestGetGameState(t *testing.T) {
// 	t.Run("Should successfully get game state using player id", func(t *testing.T) {
// 		mockStore := mockService.NewMockStorer(t)
// 		mockRandom := mockService.NewMockRandomizer(t)
// 		service := service.NewPlayerService(mockStore, mockRandom)
//
// 		ctx := context.Background()
// 		mockStore.EXPECT().
// 			GetGameStateByPlayerID(ctx, "fbb75599-9f7a-4392-b523-fd433b3208ea").
// 			Return(GameState{
// 				Players: []PlayerWithRole{
// 					{
// 						ID:       "fbb75599-9f7a-4392-b523-fd433b3208ea",
// 						Nickname: "host_player",
// 						Role:     "fibber",
// 						Question: "Am I a unicorn",
// 					},
// 				},
// 				Round:     1,
// 				RoundType: "free_form",
// 				RoomCode:  "ABC12",
// 			}, nil)
//
// 		gameState, err := service.GetGameState(ctx, "fbb75599-9f7a-4392-b523-fd433b3208ea")
//
// 		assert.NoError(t, err)
// 		assert.Len(t, gameState.Players, 1)
// 	})
//
// 	t.Run("Should throw error when store returns an error", func(t *testing.T) {
// 		mockStore := mockService.NewMockStorer(t)
// 		mockRandom := mockService.NewMockRandomizer(t)
// 		service := service.NewPlayerService(mockStore, mockRandom)
//
// 		ctx := context.Background()
// 		mockStore.EXPECT().
// 			GetGameStateByPlayerID(ctx, "fbb75599-9f7a-4392-b523-fd433b3208ea").
// 			Return(GameState{}, fmt.Errorf("failed to get game state status"))
// 		_, err := service.GetGameState(ctx, "fbb75599-9f7a-4392-b523-fd433b3208ea")
// 		assert.Error(t, err)
// 	})
// }
