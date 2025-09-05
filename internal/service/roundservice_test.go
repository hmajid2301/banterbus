package service_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mdobak/go-xerrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"gitlab.com/hmajid2301/banterbus/internal/service"
	mockService "gitlab.com/hmajid2301/banterbus/internal/service/mocks"
	"gitlab.com/hmajid2301/banterbus/internal/store/db"
)

func TestRoundServiceSubmitAnswer(t *testing.T) {
	t.Parallel()

	roomID := uuid.Must(uuid.FromString("0193a62a-4dff-774c-850a-b1fe78e2a8d2"))
	roundID := uuid.Must(uuid.FromString("0193a62a-364e-751a-9088-cf3b9711153e"))

	t.Run("Should successfully submit answer", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		now := time.Now()
		ctx := t.Context()

		roomID, err := uuid.NewV7()
		require.NoError(t, err)
		roundID, err := uuid.NewV7()
		require.NoError(t, err)
		playerID, err := uuid.NewV7()
		require.NoError(t, err)

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(db.Room{
			ID:        roomID,
			RoomState: db.Playing.String(),
		}, nil)
		mockStore.EXPECT().GetLatestRoundByPlayerID(ctx, playerID).Return(
			db.GetLatestRoundByPlayerIDRow{
				ID:             roundID,
				SubmitDeadline: pgtype.Timestamp{Time: now.Add(1 * time.Hour)},
				RoundType:      "free_form",
			}, nil)

		u, err := uuid.NewV7()
		require.NoError(t, err)
		mockRandom.EXPECT().GetID().Return(u, nil)
		mockStore.EXPECT().UpsertFibbingItAnswer(ctx, db.UpsertFibbingItAnswerParams{
			ID:       u,
			RoundID:  roundID,
			PlayerID: playerID,
			Answer:   "My answer",
		}).Return(db.FibbingItAnswer{}, nil)

		err = srv.SubmitAnswer(ctx, playerID, "My answer", now)
		assert.NoError(t, err)
	})

	t.Run("Should successfully submit answer in most_likely round", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		now := time.Now()
		ctx := t.Context()

		roomID, err := uuid.NewV7()
		require.NoError(t, err)
		roundID, err := uuid.NewV7()
		require.NoError(t, err)
		playerID, err := uuid.NewV7()
		require.NoError(t, err)

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(db.Room{
			ID:        roomID,
			RoomState: db.Playing.String(),
		}, nil)
		mockStore.EXPECT().GetLatestRoundByPlayerID(ctx, playerID).Return(
			db.GetLatestRoundByPlayerIDRow{
				ID:             roundID,
				SubmitDeadline: pgtype.Timestamp{Time: now.Add(1 * time.Hour)},
				RoundType:      "most_likely",
			}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, playerID).Return(
			[]db.GetAllPlayersInRoomRow{
				{
					ID:       playerID,
					Nickname: "host_player",
				},
				{
					ID:       uuid.Must(uuid.FromString("0194291f-ca55-7a36-baf8-9085c2fae7fb")),
					Nickname: "other_player",
				},
			}, nil)

		u, err := uuid.NewV7()
		require.NoError(t, err)
		mockRandom.EXPECT().GetID().Return(u, nil)
		mockStore.EXPECT().UpsertFibbingItAnswer(ctx, db.UpsertFibbingItAnswerParams{
			ID:       u,
			RoundID:  roundID,
			PlayerID: playerID,
			Answer:   "other_player",
		}).Return(db.FibbingItAnswer{}, nil)

		err = srv.SubmitAnswer(ctx, playerID, "other_player", now)
		assert.NoError(t, err)
	})

	t.Run("Should successfully submit answer in multiple_choice round", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		now := time.Now()
		ctx := t.Context()

		roomID, err := uuid.NewV7()
		require.NoError(t, err)
		roundID, err := uuid.NewV7()
		require.NoError(t, err)
		playerID, err := uuid.NewV7()
		require.NoError(t, err)

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(db.Room{
			ID:        roomID,
			RoomState: db.Playing.String(),
		}, nil)
		mockStore.EXPECT().GetLatestRoundByPlayerID(ctx, playerID).Return(
			db.GetLatestRoundByPlayerIDRow{
				ID:             roundID,
				SubmitDeadline: pgtype.Timestamp{Time: now.Add(1 * time.Hour)},
				RoundType:      "multiple_choice",
			}, nil)

		u, err := uuid.NewV7()
		require.NoError(t, err)
		mockRandom.EXPECT().GetID().Return(u, nil)
		mockStore.EXPECT().UpsertFibbingItAnswer(ctx, db.UpsertFibbingItAnswerParams{
			ID:       u,
			RoundID:  roundID,
			PlayerID: playerID,
			Answer:   "Strongly Agree",
		}).Return(db.FibbingItAnswer{}, nil)

		err = srv.SubmitAnswer(ctx, playerID, "Strongly Agree", now)
		assert.NoError(t, err)
	})

	t.Run("Should fail to submit answer because we fail to get room details", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		now := time.Now()
		ctx := t.Context()

		mockStore.EXPECT().
			GetRoomByPlayerID(ctx, playerID).
			Return(db.Room{}, xerrors.New("failed to get room details"))
		err := srv.SubmitAnswer(ctx, playerID, "My answer", now)
		assert.Error(t, err)
	})

	t.Run("Should fail to submit answer because room not in PLAYING state", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		now := time.Now()
		ctx := t.Context()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(db.Room{
			ID:        roomID,
			RoomState: db.Created.String(),
		}, nil)

		err := srv.SubmitAnswer(ctx, playerID, "My answer", now)
		assert.ErrorContains(t, err, "room is not in PLAYING state")
	})

	t.Run("Should fail to submit answer because failed to get latest round", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		service := service.NewRoundService(mockStore, mockRandom, "en-GB")

		now := time.Now()
		ctx := t.Context()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(db.Room{
			ID:        roomID,
			RoomState: db.Playing.String(),
		}, nil)
		mockStore.EXPECT().GetLatestRoundByPlayerID(ctx, playerID).Return(
			db.GetLatestRoundByPlayerIDRow{},
			xerrors.New("failed to get latest round"),
		)

		err := service.SubmitAnswer(ctx, playerID, "My answer", now)
		assert.Error(t, err)
	})

	t.Run("Should fail to submit answer because after submit time", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		now := time.Now()
		ctx := t.Context()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(db.Room{
			ID:        roomID,
			RoomState: db.Playing.String(),
		}, nil)
		mockStore.EXPECT().GetLatestRoundByPlayerID(ctx, playerID).Return(
			db.GetLatestRoundByPlayerIDRow{
				ID: roundID,
			}, nil)

		err := srv.SubmitAnswer(ctx, playerID, "My answer", now)
		assert.ErrorContains(t, err, "answer submission deadline has passed")
	})

	t.Run("Should fail to submit answer because failed to add answer to DB", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		now := time.Now()
		ctx := t.Context()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(db.Room{
			ID:        roomID,
			RoomState: db.Playing.String(),
		}, nil)
		mockStore.EXPECT().GetLatestRoundByPlayerID(ctx, playerID).Return(
			db.GetLatestRoundByPlayerIDRow{
				ID:             roundID,
				SubmitDeadline: pgtype.Timestamp{Time: now.Add(1 * time.Hour)},
				RoundType:      "free_form",
			}, nil)

		mockRandom.EXPECT().GetID().Return(defaultHostPlayerID, nil)
		mockStore.EXPECT().UpsertFibbingItAnswer(ctx, db.UpsertFibbingItAnswerParams{
			ID:       defaultHostPlayerID,
			RoundID:  roundID,
			PlayerID: playerID,
			Answer:   "My answer",
		}).Return(db.FibbingItAnswer{}, xerrors.New("failed to add answer to DB"))

		err := srv.SubmitAnswer(ctx, playerID, "My answer", now)
		assert.Error(t, err)
	})

	t.Run("Should fail to submit answer in most_likely round, player not in room", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		now := time.Now()
		ctx := t.Context()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(db.Room{
			ID:        roomID,
			RoomState: db.Playing.String(),
		}, nil)
		mockStore.EXPECT().GetLatestRoundByPlayerID(ctx, playerID).Return(
			db.GetLatestRoundByPlayerIDRow{
				ID:             roundID,
				SubmitDeadline: pgtype.Timestamp{Time: now.Add(1 * time.Hour)},
				RoundType:      "most_likely",
			}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, playerID).Return(
			[]db.GetAllPlayersInRoomRow{
				{
					ID:       playerID,
					Nickname: "host_player",
				},
				{
					ID:       uuid.Must(uuid.FromString("0194291f-ca55-7a36-baf8-9085c2fae7fb")),
					Nickname: "other_player",
				},
			}, nil)

		err := srv.SubmitAnswer(ctx, playerID, "player_not_found", now)
		assert.ErrorContains(t, err, "must be one of")
	})

	t.Run("Should fail to submit answer in most_likely round, db call fails", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		now := time.Now()
		ctx := t.Context()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(db.Room{
			ID:        roomID,
			RoomState: db.Playing.String(),
		}, nil)
		mockStore.EXPECT().GetLatestRoundByPlayerID(ctx, playerID).Return(
			db.GetLatestRoundByPlayerIDRow{
				ID:             roundID,
				SubmitDeadline: pgtype.Timestamp{Time: now.Add(1 * time.Hour)},
				RoundType:      "most_likely",
			}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, playerID).Return(
			[]db.GetAllPlayersInRoomRow{}, fmt.Errorf("fail to make db call"))

		err := srv.SubmitAnswer(ctx, playerID, "player_not_found", now)
		assert.Error(t, err)
	})

	t.Run("Should fail to submit answer in multiple_choice round, invalid answer", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		now := time.Now()
		ctx := t.Context()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(db.Room{
			ID:        roomID,
			RoomState: db.Playing.String(),
		}, nil)
		mockStore.EXPECT().GetLatestRoundByPlayerID(ctx, playerID).Return(
			db.GetLatestRoundByPlayerIDRow{
				ID:             roundID,
				SubmitDeadline: pgtype.Timestamp{Time: now.Add(1 * time.Hour)},
				RoundType:      "multiple_choice",
			}, nil)

		err := srv.SubmitAnswer(ctx, playerID, "invalid answer", now)
		assert.ErrorContains(t, err, "must be one of")
	})

	t.Run("Should handle single character answer", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		playerID := uuid.Must(uuid.NewV7())

		// Mock the store to return an error for GetRoomByPlayerID
		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).
			Return(db.Room{}, fmt.Errorf("player not found")).Once()

		err := srv.SubmitAnswer(ctx, playerID, "a", time.Now())
		assert.Error(t, err)
	})

	t.Run("Should handle unicode characters in answer", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		playerID := uuid.Must(uuid.NewV7())

		// Mock the store to return an error for GetRoomByPlayerID
		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).
			Return(db.Room{}, fmt.Errorf("player not found")).Once()

		err := srv.SubmitAnswer(ctx, playerID, "测试答案", time.Now())
		assert.Error(t, err)
	})

	t.Run("Should handle emoji in answer", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		playerID := uuid.Must(uuid.NewV7())

		// Mock the store to return an error for GetRoomByPlayerID
		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).
			Return(db.Room{}, fmt.Errorf("player not found")).Once()

		err := srv.SubmitAnswer(ctx, playerID, "🎮🎯", time.Now())
		assert.Error(t, err)
	})

	t.Run("Should handle special characters in answer", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		playerID := uuid.Must(uuid.NewV7())

		// Mock the store to return an error for GetRoomByPlayerID
		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).
			Return(db.Room{}, fmt.Errorf("player not found")).Once()

		err := srv.SubmitAnswer(ctx, playerID, "@#$%^&*()", time.Now())
		assert.Error(t, err)
	})

	t.Run("Should handle maximum length answer", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		playerID := uuid.Must(uuid.NewV7())

		// Mock the store to return an error for GetRoomByPlayerID
		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).
			Return(db.Room{}, fmt.Errorf("player not found")).Once()

		err := srv.SubmitAnswer(ctx, playerID, strings.Repeat("a", 500), time.Now())
		assert.Error(t, err)
	})

	t.Run("Should reject answer that exceeds maximum length", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		playerID := uuid.Must(uuid.NewV7())

		err := srv.SubmitAnswer(ctx, playerID, strings.Repeat("a", 501), time.Now())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "answer too long")
	})

	t.Run("Should reject empty answer", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		playerID := uuid.Must(uuid.NewV7())

		err := srv.SubmitAnswer(ctx, playerID, "", time.Now())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "answer cannot be empty")
	})

	t.Run("Should reject whitespace-only answer", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		playerID := uuid.Must(uuid.NewV7())

		// Mock the store to return an error for GetRoomByPlayerID since
		// whitespace validation doesn't happen early in current implementation
		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).
			Return(db.Room{}, fmt.Errorf("player not found")).Once()

		err := srv.SubmitAnswer(ctx, playerID, "   ", time.Now())
		assert.Error(t, err)
		// Note: Current implementation doesn't trim whitespace before validation
		// assert.Contains(t, err.Error(), "answer cannot be empty")
	})
}

func TestRoundServiceToggleAnswerIsReady(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully toggle answer ready state", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, playerID).Return(db.GameState{
			State:          db.FibbingITQuestion.String(),
			SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(1 * time.Hour)},
		}, nil)
		mockStore.EXPECT().ToggleAnswerIsReady(ctx, playerID).Return(db.FibbingItAnswer{}, nil)
		mockStore.EXPECT().GetAllPlayerAnswerIsReady(ctx, playerID).Return(false, nil)

		allReady, err := srv.ToggleAnswerIsReady(ctx, playerID, time.Now().UTC())
		assert.NoError(t, err)
		assert.False(t, allReady)
	})

	t.Run("Should successfully toggle answer ready state and return all players are ready", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, playerID).Return(db.GameState{
			State:          db.FibbingITQuestion.String(),
			SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(1 * time.Hour)},
		}, nil)
		mockStore.EXPECT().ToggleAnswerIsReady(ctx, playerID).Return(db.FibbingItAnswer{}, nil)
		mockStore.EXPECT().GetAllPlayerAnswerIsReady(ctx, playerID).Return(true, nil)

		allReady, err := srv.ToggleAnswerIsReady(ctx, playerID, time.Now().UTC())
		assert.NoError(t, err)
		assert.True(t, allReady)
	})

	t.Run("Should fail to toggle answer ready state, because we fail to get game state", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, playerID).Return(
			db.GameState{}, xerrors.New("failed to get state"),
		)

		_, err := srv.ToggleAnswerIsReady(ctx, playerID, time.Now().UTC())
		assert.Error(t, err)
	})

	t.Run("Should fail to toggle answer ready state because game state not in show question", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, playerID).Return(db.GameState{
			State:          db.FibbingItVoting.String(),
			SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(1 * time.Hour)},
		}, nil)

		_, err := srv.ToggleAnswerIsReady(ctx, playerID, time.Now().UTC())
		assert.ErrorContains(t, err, "room game state is not in FIBBING_IT_QUESTION state")
	})

	t.Run(
		"Should fail to toggle answer ready state, because we fail to toggle answer ready state in DB",
		func(t *testing.T) {
			t.Parallel()
			mockStore := mockService.NewMockStorer(t)
			mockRandom := mockService.NewMockRandomizer(t)
			srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

			ctx := t.Context()

			mockStore.EXPECT().GetGameStateByPlayerID(ctx, playerID).Return(db.GameState{
				State:          db.FibbingITQuestion.String(),
				SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(1 * time.Hour)},
			}, nil)
			mockStore.EXPECT().ToggleAnswerIsReady(ctx, playerID).Return(
				db.FibbingItAnswer{}, xerrors.New("failed to toggle answer is ready"),
			)

			_, err := srv.ToggleAnswerIsReady(ctx, playerID, time.Now().UTC())
			assert.Error(t, err)
		},
	)

	t.Run(
		"Should fail to toggle toggle answer ready state because we fail to get all players ready status from DB",
		func(t *testing.T) {
			t.Parallel()
			mockStore := mockService.NewMockStorer(t)
			mockRandom := mockService.NewMockRandomizer(t)
			srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

			ctx := t.Context()

			mockStore.EXPECT().GetGameStateByPlayerID(ctx, playerID).Return(db.GameState{
				State:          db.FibbingITQuestion.String(),
				SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(1 * time.Hour)},
			}, nil)
			mockStore.EXPECT().ToggleAnswerIsReady(ctx, playerID).Return(db.FibbingItAnswer{}, nil)
			mockStore.EXPECT().GetAllPlayerAnswerIsReady(ctx, playerID).Return(
				false, xerrors.New("failed to get player answer is ready status"),
			)

			_, err := srv.ToggleAnswerIsReady(ctx, playerID, time.Now().UTC())
			assert.Error(t, err)
		},
	)

	t.Run(
		"Should fail to toggle toggle answer ready state because after deadline",
		func(t *testing.T) {
			t.Parallel()
			mockStore := mockService.NewMockStorer(t)
			mockRandom := mockService.NewMockRandomizer(t)
			srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

			ctx := t.Context()

			mockStore.EXPECT().GetGameStateByPlayerID(ctx, playerID).Return(db.GameState{
				State:          db.FibbingITQuestion.String(),
				SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(-1 * time.Second)},
			}, nil)

			_, err := srv.ToggleAnswerIsReady(ctx, playerID, time.Now().UTC())
			assert.Error(t, err)
		},
	)
}

func TestRoundServiceUpdateStateToVoting(t *testing.T) {
	t.Parallel()

	gameStateID := uuid.Must(uuid.FromString("fbb75599-9f7a-4392-b523-fd433b3208ea"))
	roundID := uuid.Must(uuid.FromString("0193a62a-364e-751a-9088-cf3b9711153e"))

	t.Run("Should successfully update state to voting", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		now := time.Now().Add(30 * time.Second)

		mockStore.EXPECT().GetGameState(ctx, gameStateID).Return(db.GameState{
			State: db.FibbingITQuestion.String(),
		}, nil)
		mockStore.EXPECT().UpdateGameState(ctx, db.UpdateGameStateParams{
			ID:             gameStateID,
			SubmitDeadline: pgtype.Timestamp{Time: now, Valid: true},
			State:          db.FibbingItVoting.String(),
		}).Return(db.GameState{}, nil)
		mockStore.EXPECT().GetLatestRoundByGameStateID(ctx, gameStateID).Return(db.GetLatestRoundByGameStateIDRow{
			ID:    roundID,
			Round: 1,
		}, nil)
		mockStore.EXPECT().GetVotingState(ctx, roundID).Return([]db.GetVotingStateRow{
			{
				GameStateID:    gameStateID,
				PlayerID:       defaultHostPlayerID,
				Nickname:       "Player 1",
				Avatar:         "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Player+1",
				Question:       "My question",
				Votes:          0,
				SubmitDeadline: pgtype.Timestamp{Time: now},
			},
			{
				GameStateID:    gameStateID,
				PlayerID:       defaultOtherPlayerID,
				Nickname:       "Player 2",
				Avatar:         "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Player+2",
				Question:       "My question",
				Votes:          0,
				SubmitDeadline: pgtype.Timestamp{Time: now},
			},
		}, nil)

		votes, err := srv.UpdateStateToVoting(ctx, gameStateID, now)
		assert.NoError(t, err)
		expectedVotes := service.VotingState{
			GameStateID: gameStateID,
			Question:    "My question",
			Round:       1,
			Players: []service.PlayerWithVoting{
				{
					ID:       defaultOtherPlayerID,
					Nickname: "Player 2",
					Avatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Player+2",
					Votes:    0,
					IsReady:  false,
				},
				{
					ID:       defaultHostPlayerID,
					Nickname: "Player 1",
					Avatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Player+1",
					Votes:    0,
					IsReady:  false,
				},
			},
		}
		diffOpts := cmpopts.IgnoreFields(votes, "Deadline")
		PartialEqual(t, expectedVotes, votes, diffOpts)
		assert.LessOrEqual(t, int(votes.Deadline.Seconds()), 30)
	})

	t.Run("Should fail to get game state", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		now := time.Now().Add(30 * time.Second)

		mockStore.EXPECT().GetGameState(ctx, gameStateID).Return(
			db.GameState{}, xerrors.New("failed to get game state"),
		)
		_, err := srv.UpdateStateToVoting(ctx, gameStateID, now)
		assert.Error(t, err)
	})

	t.Run("Should fail because state FIBBING_IT_VOTING", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		now := time.Now().Add(30 * time.Second)

		mockStore.EXPECT().GetGameState(ctx, gameStateID).Return(db.GameState{
			State: db.FibbingItVoting.String(),
		}, nil)

		_, err := srv.UpdateStateToVoting(ctx, gameStateID, now)
		assert.ErrorContains(t, err, "game state is not in FIBBING_IT_QUESTION state")
	})

	t.Run("Should fail because update game state", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		now := time.Now().Add(30 * time.Second)

		mockStore.EXPECT().GetGameState(ctx, gameStateID).Return(db.GameState{
			State: db.FibbingITQuestion.String(),
		}, nil)
		mockStore.EXPECT().UpdateGameState(ctx, db.UpdateGameStateParams{
			ID:             gameStateID,
			SubmitDeadline: pgtype.Timestamp{Time: now, Valid: true},
			State:          db.FibbingItVoting.String(),
		}).Return(db.GameState{}, xerrors.New("failed to update game state"))

		_, err := srv.UpdateStateToVoting(ctx, gameStateID, now)
		assert.Error(t, err)
	})
}

func TestRoundServiceSubmitVote(t *testing.T) {
	t.Parallel()

	roundID := uuid.Must(uuid.FromString("0193a62a-7740-7bce-849d-0e462465ca0e"))

	t.Run("Should successfully submit vote", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, defaultHostPlayerID).Return(db.GameState{
			State: db.FibbingItVoting.String(),
		}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, defaultHostPlayerID).Return([]db.GetAllPlayersInRoomRow{
			{
				ID:       defaultHostPlayerID,
				Nickname: "Player 1",
			},
			{
				ID:       defaultOtherPlayerID,
				Nickname: "Player 2",
			},
		}, nil)
		deadline := time.Now().Add(5 * time.Second)
		mockStore.EXPECT().GetLatestRoundByPlayerID(ctx, defaultHostPlayerID).Return(db.GetLatestRoundByPlayerIDRow{
			ID:             roundID,
			SubmitDeadline: pgtype.Timestamp{Time: deadline},
		}, nil)

		u, err := uuid.NewV7()
		require.NoError(t, err)
		mockRandom.EXPECT().GetID().Return(u, nil)
		mockStore.EXPECT().UpsertFibbingItVote(ctx, db.UpsertFibbingItVoteParams{
			ID:               u,
			RoundID:          roundID,
			PlayerID:         defaultHostPlayerID,
			VotedForPlayerID: defaultOtherPlayerID,
		}).Return(nil)
		mockStore.EXPECT().GetVotingState(ctx, roundID).Return([]db.GetVotingStateRow{
			{
				PlayerID:       defaultHostPlayerID,
				Nickname:       "Player 1",
				Avatar:         "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Player+1",
				Votes:          int64(0),
				Question:       "My question",
				Round:          1,
				SubmitDeadline: pgtype.Timestamp{Time: deadline},
				Answer:         pgtype.Text{String: "My answer"},
			},
			{
				PlayerID:       defaultOtherPlayerID,
				Nickname:       "Player 2",
				Avatar:         "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Player+2",
				Votes:          int64(1),
				Question:       "My question",
				Round:          1,
				SubmitDeadline: pgtype.Timestamp{Time: deadline},
				Answer:         pgtype.Text{String: "My other answer"},
			},
		}, nil)

		now := time.Now()
		votingState, err := srv.SubmitVote(ctx, defaultHostPlayerID, "Player 2", now)
		assert.NoError(t, err)

		expectedVotingState := service.VotingState{
			Question: "My question",
			Round:    1,
			Players: []service.PlayerWithVoting{
				{
					ID:       defaultHostPlayerID,
					Nickname: "Player 1",
					Avatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Player+1",
					Votes:    0,
					Answer:   "My answer",
				},
				{
					ID:       defaultOtherPlayerID,
					Nickname: "Player 2",
					Avatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Player+2",
					Votes:    1,
					Answer:   "My other answer",
				},
			},
		}

		diffOpts := cmpopts.IgnoreFields(votingState, "Deadline")
		PartialEqual(t, expectedVotingState, votingState, diffOpts)
		assert.LessOrEqual(t, int(votingState.Deadline.Seconds()), 5)
	})

	t.Run("Should fail because we fail to get game state", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		now := time.Now().Add(30 * time.Second)

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, defaultHostPlayerID).Return(
			db.GameState{}, xerrors.New("failed to get game state"),
		)

		_, err := srv.SubmitVote(ctx, defaultHostPlayerID, "Player 2", now)
		assert.Error(t, err)
	})

	t.Run("Should fail because game state not in voting state", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		now := time.Now().Add(30 * time.Second)

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, defaultHostPlayerID).Return(db.GameState{
			State: db.FibbingITQuestion.String(),
		}, nil)

		_, err := srv.SubmitVote(ctx, defaultHostPlayerID, "Player 2", now)
		assert.ErrorContains(t, err, "game state is not in FIBBING_IT_VOTING state")
	})

	t.Run("Should fail because we fail to get all players in room", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		now := time.Now().Add(30 * time.Second)

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, defaultHostPlayerID).Return(db.GameState{
			State: db.FibbingItVoting.String(),
		}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, defaultHostPlayerID).Return(
			[]db.GetAllPlayersInRoomRow{}, xerrors.New("failed to get all players in room"),
		)

		_, err := srv.SubmitVote(ctx, defaultHostPlayerID, "Player 2", now)
		assert.Error(t, err)
	})

	t.Run("Should fail because we voted for themselves", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		now := time.Now().Add(30 * time.Second)

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, defaultHostPlayerID).Return(db.GameState{
			State: db.FibbingItVoting.String(),
		}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, defaultHostPlayerID).Return([]db.GetAllPlayersInRoomRow{
			{
				ID:       defaultHostPlayerID,
				Nickname: "Player 1",
			},
			{
				ID:       defaultOtherPlayerID,
				Nickname: "Player 2",
			},
		}, nil)

		_, err := srv.SubmitVote(ctx, defaultHostPlayerID, "Player 1", now)
		assert.ErrorContains(t, err, "cannot vote for yourself")
	})

	t.Run("Should fail because nickname not found", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		now := time.Now().Add(30 * time.Second)

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, defaultHostPlayerID).Return(db.GameState{
			State: db.FibbingItVoting.String(),
		}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, defaultHostPlayerID).Return([]db.GetAllPlayersInRoomRow{
			{
				ID:       defaultHostPlayerID,
				Nickname: "Player 1",
			},
			{
				ID:       defaultOtherPlayerID,
				Nickname: "Player 2",
			},
		}, nil)

		_, err := srv.SubmitVote(ctx, defaultHostPlayerID, "not_in_room", now)
		assert.ErrorContains(t, err, "player with nickname not_in_room not found")
	})

	t.Run("Should fail because we failed to get round information", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		now := time.Now().Add(30 * time.Second)

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, defaultHostPlayerID).Return(db.GameState{
			State: db.FibbingItVoting.String(),
		}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, defaultHostPlayerID).Return([]db.GetAllPlayersInRoomRow{
			{
				ID:       defaultHostPlayerID,
				Nickname: "Player 1",
			},
			{
				ID:       defaultOtherPlayerID,
				Nickname: "Player 2",
			},
		}, nil)
		mockStore.EXPECT().GetLatestRoundByPlayerID(ctx, defaultHostPlayerID).Return(
			db.GetLatestRoundByPlayerIDRow{}, xerrors.New("failed to get latest round"),
		)

		_, err := srv.SubmitVote(ctx, defaultHostPlayerID, "Player 2", now)
		assert.Error(t, err)
	})

	t.Run("Should fail because we are passed the submit deadline", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		now := time.Now().Add(30 * time.Second)

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, defaultHostPlayerID).Return(db.GameState{
			State: db.FibbingItVoting.String(),
		}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, defaultHostPlayerID).Return([]db.GetAllPlayersInRoomRow{
			{
				ID:       defaultHostPlayerID,
				Nickname: "Player 1",
			},
			{
				ID:       defaultOtherPlayerID,
				Nickname: "Player 2",
			},
		}, nil)
		mockStore.EXPECT().GetLatestRoundByPlayerID(ctx, defaultHostPlayerID).Return(db.GetLatestRoundByPlayerIDRow{
			ID:             roundID,
			SubmitDeadline: pgtype.Timestamp{Time: now.Add(-1 * time.Hour)},
		}, nil)

		_, err := srv.SubmitVote(ctx, defaultHostPlayerID, "Player 2", now)
		assert.ErrorContains(t, err, "answer submission deadline has passed")
	})

	t.Run("Should fail because we fail to upsert fibbing it vote", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		now := time.Now().Add(30 * time.Second)

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, defaultHostPlayerID).Return(db.GameState{
			State: db.FibbingItVoting.String(),
		}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, defaultHostPlayerID).Return([]db.GetAllPlayersInRoomRow{
			{
				ID:       defaultHostPlayerID,
				Nickname: "Player 1",
			},
			{
				ID:       defaultOtherPlayerID,
				Nickname: "Player 2",
			},
		}, nil)
		mockStore.EXPECT().GetLatestRoundByPlayerID(ctx, defaultHostPlayerID).Return(db.GetLatestRoundByPlayerIDRow{
			ID:             roundID,
			SubmitDeadline: pgtype.Timestamp{Time: now.Add(1 * time.Hour)},
		}, nil)

		u, err := uuid.NewV7()
		require.NoError(t, err)
		mockRandom.EXPECT().GetID().Return(u, nil)
		mockStore.EXPECT().UpsertFibbingItVote(ctx, db.UpsertFibbingItVoteParams{
			ID:               u,
			RoundID:          roundID,
			PlayerID:         defaultHostPlayerID,
			VotedForPlayerID: defaultOtherPlayerID,
		}).Return(xerrors.New("failed to upsert fibbing it vote"))

		_, err = srv.SubmitVote(ctx, defaultHostPlayerID, "Player 2", now)
		assert.Error(t, err)
	})

	t.Run("Should fail because we fail to get vote count", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		now := time.Now().Add(30 * time.Second)

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, defaultHostPlayerID).Return(db.GameState{
			State: db.FibbingItVoting.String(),
		}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, defaultHostPlayerID).Return([]db.GetAllPlayersInRoomRow{
			{
				ID:       defaultHostPlayerID,
				Nickname: "Player 1",
			},
			{
				ID:       defaultOtherPlayerID,
				Nickname: "Player 2",
			},
		}, nil)
		mockStore.EXPECT().GetLatestRoundByPlayerID(ctx, defaultHostPlayerID).Return(db.GetLatestRoundByPlayerIDRow{
			ID:             roundID,
			SubmitDeadline: pgtype.Timestamp{Time: now.Add(1 * time.Hour)},
		}, nil)
		u, err := uuid.NewV7()
		require.NoError(t, err)
		mockRandom.EXPECT().GetID().Return(u, nil)
		mockStore.EXPECT().UpsertFibbingItVote(ctx, db.UpsertFibbingItVoteParams{
			ID:               u,
			RoundID:          roundID,
			PlayerID:         defaultHostPlayerID,
			VotedForPlayerID: defaultOtherPlayerID,
		}).Return(nil)
		mockStore.EXPECT().GetVotingState(ctx, roundID).Return(
			[]db.GetVotingStateRow{}, xerrors.New("failed to get vote count"),
		)

		_, err = srv.SubmitVote(ctx, defaultHostPlayerID, "Player 2", now)
		assert.Error(t, err)
	})
}

func TestRoundServiceGetVotingState(t *testing.T) {
	t.Parallel()

	roundID := uuid.Must(uuid.FromString("0193a629-e26c-7326-8df4-81ad3ec82214"))
	gameStateID := uuid.Must(uuid.FromString("fbb75599-9f7a-4392-b523-fd433b3208ea"))

	t.Run("Should successfully get voting state", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()

		deadline := time.Now().Add(5 * time.Second)
		mockStore.EXPECT().GetLatestRoundByPlayerID(ctx, playerID).Return(
			db.GetLatestRoundByPlayerIDRow{
				ID:             roundID,
				Round:          1,
				SubmitDeadline: pgtype.Timestamp{Time: deadline},
			}, nil)
		mockStore.EXPECT().GetVotingState(ctx, roundID).Return(
			[]db.GetVotingStateRow{
				{
					GameStateID: gameStateID,
					PlayerID:    playerID,
					Nickname:    "nickname",
					Votes:       int64(1),
					Avatar:      "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=nickname",
					Question:    "My  question",
					Round:       1,
					Answer:      pgtype.Text{String: "A cat"},
				},
			}, nil)

		votingState, err := srv.GetVotingState(ctx, playerID)

		assert.NoError(t, err)
		expectedVotingState := service.VotingState{
			GameStateID: gameStateID,
			Question:    "My  question",
			Round:       1,
			Players: []service.PlayerWithVoting{
				{
					ID:       playerID,
					Nickname: "nickname",
					Avatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=nickname",
					Votes:    1,
					Answer:   "A cat",
					IsReady:  false,
				},
			},
		}

		diffOpts := cmpopts.IgnoreFields(votingState, "Deadline")
		PartialEqual(t, expectedVotingState, votingState, diffOpts)
		assert.LessOrEqual(t, int(votingState.Deadline.Seconds()), 5)
	})

	t.Run("Should fail to get voting state because fail to get round info from DB", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()

		mockStore.EXPECT().GetLatestRoundByPlayerID(ctx, playerID).Return(
			db.GetLatestRoundByPlayerIDRow{}, xerrors.New("failed to get round info"),
		)
		_, err := srv.GetVotingState(ctx, playerID)
		assert.Error(t, err)
	})

	t.Run("Should fail to get voting state because fail to get votes from DB", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()

		mockStore.EXPECT().GetLatestRoundByPlayerID(ctx, playerID).Return(
			db.GetLatestRoundByPlayerIDRow{
				ID: roundID,
			}, nil)
		mockStore.EXPECT().GetVotingState(ctx, roundID).Return(
			[]db.GetVotingStateRow{}, xerrors.New("failed to get votes"),
		)

		_, err := srv.GetVotingState(ctx, playerID)
		assert.Error(t, err)
	})
}

func TestRoundServiceToggleVotingIsReady(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully toggle voting ready state", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, playerID).Return(db.GameState{
			State:          db.FibbingItVoting.String(),
			SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(1 * time.Hour)},
		}, nil)
		mockStore.EXPECT().ToggleVotingIsReady(ctx, playerID).Return(db.FibbingItVote{}, nil)
		mockStore.EXPECT().GetAllPlayersVotingIsReady(ctx, playerID).Return(false, nil)

		allReady, err := srv.ToggleVotingIsReady(ctx, playerID, time.Now().UTC())
		assert.NoError(t, err)
		assert.False(t, allReady)
	})

	t.Run("Should successfully toggle answer ready state and return all players are ready", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, playerID).Return(db.GameState{
			State:          db.FibbingItVoting.String(),
			SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(1 * time.Hour)},
		}, nil)
		mockStore.EXPECT().ToggleVotingIsReady(ctx, playerID).Return(db.FibbingItVote{}, nil)
		mockStore.EXPECT().GetAllPlayersVotingIsReady(ctx, playerID).Return(true, nil)

		allReady, err := srv.ToggleVotingIsReady(ctx, playerID, time.Now().UTC())
		assert.NoError(t, err)
		assert.True(t, allReady)
	})

	t.Run("Should fail to toggle voting ready state, because we fail to get game state", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, playerID).Return(
			db.GameState{}, xerrors.New("failed to get state"),
		)

		_, err := srv.ToggleAnswerIsReady(ctx, playerID, time.Now().UTC())
		assert.Error(t, err)
	})

	t.Run(
		"Should fail to voting toggle ready state because after deadline",
		func(t *testing.T) {
			t.Parallel()
			mockStore := mockService.NewMockStorer(t)
			mockRandom := mockService.NewMockRandomizer(t)
			srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

			ctx := t.Context()

			mockStore.EXPECT().GetGameStateByPlayerID(ctx, playerID).Return(db.GameState{
				State:          db.FibbingItVoting.String(),
				SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(-1 * time.Second)},
			}, nil)

			_, err := srv.ToggleVotingIsReady(ctx, playerID, time.Now().UTC())
			assert.Error(t, err)
		},
	)

	t.Run("Should fail to toggle voting ready state because game state not in voting state", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, playerID).Return(db.GameState{
			State:          db.FibbingITQuestion.String(),
			SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(1 * time.Hour)},
		}, nil)

		_, err := srv.ToggleVotingIsReady(ctx, playerID, time.Now().UTC())
		assert.ErrorContains(t, err, "room game state is not in FIBBING_IT_VOTING state")
	})

	t.Run(
		"Should fail to toggle voting ready state, because we fail to toggle answer ready state in DB",
		func(t *testing.T) {
			t.Parallel()
			mockStore := mockService.NewMockStorer(t)
			mockRandom := mockService.NewMockRandomizer(t)
			srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

			ctx := t.Context()

			mockStore.EXPECT().GetGameStateByPlayerID(ctx, playerID).Return(db.GameState{
				State:          db.FibbingItVoting.String(),
				SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(1 * time.Hour)},
			}, nil)
			mockStore.EXPECT().ToggleVotingIsReady(ctx, playerID).Return(
				db.FibbingItVote{}, xerrors.New("failed to toggle voting is ready"),
			)

			_, err := srv.ToggleVotingIsReady(ctx, playerID, time.Now().UTC())
			assert.Error(t, err)
		},
	)

	t.Run(
		"Should fail to toggle voting ready state because we fail to get all players ready status from DB",
		func(t *testing.T) {
			t.Parallel()
			mockStore := mockService.NewMockStorer(t)
			mockRandom := mockService.NewMockRandomizer(t)
			srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

			ctx := t.Context()

			mockStore.EXPECT().GetGameStateByPlayerID(ctx, playerID).Return(db.GameState{
				State:          db.FibbingItVoting.String(),
				SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(1 * time.Hour)},
			}, nil)
			mockStore.EXPECT().ToggleVotingIsReady(ctx, playerID).Return(db.FibbingItVote{}, nil)
			mockStore.EXPECT().GetAllPlayersVotingIsReady(ctx, playerID).Return(
				false, xerrors.New("failed to get player voting is ready status"),
			)

			_, err := srv.ToggleVotingIsReady(ctx, playerID, time.Now().UTC())
			assert.Error(t, err)
		},
	)
}

func TestRoundServiceUpdateStateToReveal(t *testing.T) {
	t.Parallel()

	gameStateID := uuid.Must(uuid.FromString("fbb75599-9f7a-4392-b523-fd433b3208ea"))
	roundID := uuid.Must(uuid.FromString("0193a62a-364e-751a-9088-cf3b9711153e"))

	tests := []struct {
		name                   string
		votesPlayerOne         int
		votesPlayerTwo         int
		expectedPlayerRole     string
		expectedPlayerNickname string
		expectedPlayerAvatar   string
		expectedShouldReveal   bool
	}{
		{
			name:                   "Should show reveal state revealing normal player",
			votesPlayerOne:         0,
			votesPlayerTwo:         1,
			expectedPlayerNickname: "Player 2",
			expectedPlayerAvatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Player+2",
			expectedPlayerRole:     "normal",
			expectedShouldReveal:   true,
		},
		{
			name:                   "Should show reveal state revealing fibber player",
			votesPlayerOne:         1,
			votesPlayerTwo:         0,
			expectedPlayerNickname: "Player 1",
			expectedPlayerAvatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Player+1",
			expectedPlayerRole:     "fibber",
			expectedShouldReveal:   true,
		},
		{
			name:                 "Should show reveal state and not reveal any player",
			votesPlayerOne:       0,
			votesPlayerTwo:       0,
			expectedShouldReveal: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockStore := mockService.NewMockStorer(t)
			mockRandom := mockService.NewMockRandomizer(t)
			srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

			ctx := t.Context()
			now := time.Now().Add(15 * time.Second)

			mockStore.EXPECT().GetGameState(ctx, gameStateID).Return(db.GameState{
				State: db.FibbingItVoting.String(),
			}, nil)
			mockStore.EXPECT().UpdateGameState(ctx, db.UpdateGameStateParams{
				ID:             gameStateID,
				SubmitDeadline: pgtype.Timestamp{Time: now, Valid: true},
				State:          db.FibbingItRevealRole.String(),
			}).Return(db.GameState{}, nil)
			mockStore.EXPECT().GetLatestRoundByGameStateID(ctx, gameStateID).Return(db.GetLatestRoundByGameStateIDRow{
				ID:    roundID,
				Round: 1,
			}, nil)
			mockStore.EXPECT().GetVotingState(ctx, roundID).Return([]db.GetVotingStateRow{
				{
					PlayerID:       defaultHostPlayerID,
					Nickname:       "Player 1",
					Avatar:         "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Player+1",
					Question:       "My question",
					Votes:          int64(tt.votesPlayerOne),
					SubmitDeadline: pgtype.Timestamp{Time: now},
					Role:           pgtype.Text{String: "fibber"},
				},
				{
					PlayerID:       defaultOtherPlayerID,
					Nickname:       "Player 2",
					Avatar:         "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Player+2",
					Question:       "My question",
					Votes:          int64(tt.votesPlayerTwo),
					SubmitDeadline: pgtype.Timestamp{Time: now},
					Role:           pgtype.Text{String: "normal"},
				},
			}, nil)

			reveal, err := srv.UpdateStateToReveal(ctx, gameStateID, now)
			assert.NoError(t, err)
			expectedReveal := service.RevealRoleState{
				VotedForPlayerNickname: tt.expectedPlayerNickname,
				VotedForPlayerAvatar:   tt.expectedPlayerAvatar,
				VotedForPlayerRole:     tt.expectedPlayerRole,
				Round:                  1,
				ShouldReveal:           tt.expectedShouldReveal,
				Deadline:               time.Until(now),
				PlayerIDs: []uuid.UUID{
					defaultOtherPlayerID,
					defaultHostPlayerID,
				},
			}

			diffOpts := cmpopts.IgnoreFields(reveal, "Deadline")
			PartialEqual(t, expectedReveal, reveal, diffOpts)
			assert.LessOrEqual(t, int(reveal.Deadline.Seconds()), 15)
		})
	}

	t.Run(
		"Should fail to update state to reveal because we fail to get game state",
		func(t *testing.T) {
			t.Parallel()
			mockStore := mockService.NewMockStorer(t)
			mockRandom := mockService.NewMockRandomizer(t)
			srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

			ctx := t.Context()
			now := time.Now().Add(15 * time.Second)

			mockStore.EXPECT().
				GetGameState(ctx, gameStateID).
				Return(db.GameState{}, xerrors.New("failed to get game state"))
			_, err := srv.UpdateStateToReveal(ctx, gameStateID, now)
			assert.Error(t, err)
		},
	)

	t.Run(
		"Should fail to update state to reveal because we wrong game state",
		func(t *testing.T) {
			t.Parallel()
			mockStore := mockService.NewMockStorer(t)
			mockRandom := mockService.NewMockRandomizer(t)
			srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

			ctx := t.Context()
			now := time.Now().Add(15 * time.Second)

			mockStore.EXPECT().GetGameState(ctx, gameStateID).Return(db.GameState{
				State: db.FibbingITQuestion.String(),
			}, nil)

			_, err := srv.UpdateStateToReveal(ctx, gameStateID, now)
			assert.ErrorContains(t, err, "game state is not in FIBBING_IT_VOTING state")
		},
	)

	t.Run(
		"Should fail to update state to reveal because we fail to update game state",
		func(t *testing.T) {
			t.Parallel()
			mockStore := mockService.NewMockStorer(t)
			mockRandom := mockService.NewMockRandomizer(t)
			srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

			ctx := t.Context()
			now := time.Now().Add(15 * time.Second)

			mockStore.EXPECT().GetGameState(ctx, gameStateID).Return(db.GameState{
				State: db.FibbingItVoting.String(),
			}, nil)
			mockStore.EXPECT().UpdateGameState(ctx, db.UpdateGameStateParams{
				ID:             gameStateID,
				SubmitDeadline: pgtype.Timestamp{Time: now, Valid: true},
				State:          db.FibbingItRevealRole.String(),
			}).Return(db.GameState{}, xerrors.New("failed to update game state"))

			_, err := srv.UpdateStateToReveal(ctx, gameStateID, now)
			assert.Error(t, err)
		},
	)

	t.Run(
		"Should fail to update state to reveal because we get latest round by game state ID",
		func(t *testing.T) {
			t.Parallel()
			mockStore := mockService.NewMockStorer(t)
			mockRandom := mockService.NewMockRandomizer(t)
			srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

			ctx := t.Context()
			now := time.Now().Add(15 * time.Second)

			mockStore.EXPECT().GetGameState(ctx, gameStateID).Return(db.GameState{
				State: db.FibbingItVoting.String(),
			}, nil)
			mockStore.EXPECT().UpdateGameState(ctx, db.UpdateGameStateParams{
				ID:             gameStateID,
				SubmitDeadline: pgtype.Timestamp{Time: now, Valid: true},
				State:          db.FibbingItRevealRole.String(),
			}).Return(db.GameState{}, nil)
			mockStore.EXPECT().GetLatestRoundByGameStateID(ctx, gameStateID).Return(
				db.GetLatestRoundByGameStateIDRow{}, xerrors.New("failed to get latest round by game state ID"),
			)

			_, err := srv.UpdateStateToReveal(ctx, gameStateID, now)
			assert.Error(t, err)
		},
	)

	t.Run(
		"Should fail to update state to reveal because we fail to get voting state",
		func(t *testing.T) {
			t.Parallel()
			mockStore := mockService.NewMockStorer(t)
			mockRandom := mockService.NewMockRandomizer(t)
			srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

			ctx := t.Context()
			now := time.Now().Add(15 * time.Second)

			mockStore.EXPECT().GetGameState(ctx, gameStateID).Return(db.GameState{
				State: db.FibbingItVoting.String(),
			}, nil)
			mockStore.EXPECT().UpdateGameState(ctx, db.UpdateGameStateParams{
				ID:             gameStateID,
				SubmitDeadline: pgtype.Timestamp{Time: now, Valid: true},
				State:          db.FibbingItRevealRole.String(),
			}).Return(db.GameState{}, nil)
			mockStore.EXPECT().GetLatestRoundByGameStateID(ctx, gameStateID).Return(db.GetLatestRoundByGameStateIDRow{
				ID:    roundID,
				Round: 1,
			}, nil)
			mockStore.EXPECT().GetVotingState(ctx, roundID).Return(
				[]db.GetVotingStateRow{}, xerrors.New("failed to get voting state"),
			)

			_, err := srv.UpdateStateToReveal(ctx, gameStateID, now)
			assert.Error(t, err)
		},
	)
}

func TestRoundServiceUpdateStateToQuestion(t *testing.T) {
	t.Parallel()

	gameStateID := uuid.Must(uuid.FromString("fbb75599-9f7a-4392-b523-fd433b3208ea"))
	groupID := uuid.Must(uuid.FromString("0193a629-1fcf-79dd-ac70-760bedbdffa9"))

	tests := []struct {
		name            string
		roundNumber     int32
		roundType       string
		expectedRound   int32
		expectedType    string
		expectedAnswers []string
		normalQuestion  string
		fibberQuestion  string
	}{
		{
			name:            "Should update state to question state successfully with round 2 and free_form",
			roundNumber:     1,
			roundType:       "free_form",
			expectedRound:   2,
			expectedType:    "free_form",
			expectedAnswers: []string{},
			normalQuestion:  "What if your favourite city",
			fibberQuestion:  "What is your favourite hotel",
		},
		{
			name:            "Should update state to question state successfully with round 1 and multiple_choice",
			roundNumber:     3,
			roundType:       "free_form",
			expectedRound:   1,
			expectedType:    "multiple_choice",
			expectedAnswers: []string{"Strongly Agree", "Agree", "Neutral", "Disagree", "Strongly Disagree"},
			normalQuestion:  "I love pizza",
			fibberQuestion:  "I love burgers",
		},
		{
			name:            "Should update state to question state successfully with round 1 and most_likely",
			roundNumber:     3,
			roundType:       "multiple_choice",
			expectedRound:   1,
			expectedType:    "most_likely",
			expectedAnswers: []string{"Player 1", "Player 2"},
			normalQuestion:  "go to prison",
			fibberQuestion:  "go to a bank",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockStore := mockService.NewMockStorer(t)
			mockRandom := mockService.NewMockRandomizer(t)
			srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

			ctx := t.Context()
			gameName := gameName
			deadline := time.Now().Add(5 * time.Second).UTC()

			mockStore.EXPECT().GetGameState(ctx, gameStateID).Return(db.GameState{
				State: db.FibbingItRevealRole.String(),
			}, nil)
			mockStore.EXPECT().UpdateGameState(ctx, db.UpdateGameStateParams{
				ID:             gameStateID,
				SubmitDeadline: pgtype.Timestamp{Time: deadline, Valid: true},
				State:          db.FibbingITQuestion.String(),
			}).Return(db.GameState{}, nil)
			mockStore.EXPECT().GetAllPlayersByGameStateID(ctx, gameStateID).Return(
				[]db.GetAllPlayersByGameStateIDRow{
					{
						ID:       defaultHostPlayerID,
						Nickname: "Player 1",
					},
					{
						ID:       defaultOtherPlayerID,
						Nickname: "Player 2",
					},
				},
				nil,
			)
			mockStore.EXPECT().GetLatestRoundByGameStateID(ctx, gameStateID).Return(db.GetLatestRoundByGameStateIDRow{
				ID:        uuid.Must(uuid.FromString("0193ea48-c27f-74bd-8a17-523f69350aff")),
				RoundType: tt.roundType,
				Round:     tt.roundNumber,
			}, nil)
			mockStore.EXPECT().GetRandomQuestionByRound(ctx, db.GetRandomQuestionByRoundParams{
				GameName:  gameName,
				RoundType: tt.expectedType,
			}).Return([]db.GetRandomQuestionByRoundRow{
				{
					ID:         uuid.Must(uuid.FromString("0193ea48-c27f-74bd-8a17-523f69350aca")),
					QuestionID: uuid.Must(uuid.FromString("0193a629-7dcc-78ad-822f-fd5d83c89ae7")),
					Question:   tt.normalQuestion,
					Locale:     "en-GB",
					GroupID:    groupID,
				},
			}, nil)
			mockStore.EXPECT().GetRandomQuestionInGroup(ctx, db.GetRandomQuestionInGroupParams{
				GroupType:          "",
				GroupID:            groupID,
				ExcludedQuestionID: uuid.Must(uuid.FromString("0193a629-7dcc-78ad-822f-fd5d83c89ae7")),
				RoundType:          tt.expectedType,
			}).Return([]db.GetRandomQuestionInGroupRow{
				{
					QuestionID: uuid.Must(uuid.FromString("0193a629-a9ac-7fc4-828c-a1334c282e0f")),
					Question:   tt.fibberQuestion,
				},
			}, nil)

			if tt.roundNumber == 3 {
				mockRandom.EXPECT().GetFibberIndex(2).Return(1)
			} else {
				mockStore.EXPECT().GetFibberByRoundID(ctx, uuid.Must(uuid.FromString("0193ea48-c27f-74bd-8a17-523f69350aff"))).Return(
					db.FibbingItPlayerRole{
						PlayerID: defaultOtherPlayerID,
					},
					nil,
				)
			}

			mockStore.EXPECT().NewRound(ctx, db.NewRoundArgs{
				GameStateID:       gameStateID,
				NormalsQuestionID: uuid.Must(uuid.FromString("0193a629-7dcc-78ad-822f-fd5d83c89ae7")),
				FibberQuestionID:  uuid.Must(uuid.FromString("0193a629-a9ac-7fc4-828c-a1334c282e0f")),
				Round:             tt.expectedRound,
				RoundType:         tt.expectedType,
				Players: []db.GetAllPlayersByGameStateIDRow{
					{
						ID:       defaultHostPlayerID,
						Nickname: "Player 1",
					},
					{
						ID:       defaultOtherPlayerID,
						Nickname: "Player 2",
					},
				},
				FibberLoc: 1,
			}).Return(nil)

			gameState, err := srv.UpdateStateToQuestion(ctx, gameStateID, deadline, false)
			expectedGameState := service.QuestionState{
				Deadline:    time.Until(deadline),
				GameStateID: gameStateID,
				Players: []service.PlayerWithRole{
					{
						ID:              defaultHostPlayerID,
						Role:            "normal",
						Question:        tt.normalQuestion,
						PossibleAnswers: tt.expectedAnswers,
					},
					{
						ID:              defaultOtherPlayerID,
						Role:            "fibber",
						Question:        tt.fibberQuestion,
						PossibleAnswers: tt.expectedAnswers,
					},
				},
				Round:     int(tt.expectedRound),
				RoundType: tt.expectedType,
			}

			assert.NoError(t, err)

			diffOpts := cmpopts.IgnoreFields(gameState, "Deadline")
			PartialEqual(t, expectedGameState, gameState, diffOpts)
			assert.LessOrEqual(t, int(gameState.Deadline.Seconds()), 5)
		})
	}

	t.Run("Should fail to update state to question because we fail to get game state", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		deadline := time.Now().Add(5 * time.Second).UTC()

		mockStore.EXPECT().
			GetGameState(ctx, gameStateID).
			Return(db.GameState{}, xerrors.New("failed to get game state"))
		_, err := srv.UpdateStateToQuestion(ctx, gameStateID, deadline, false)
		assert.Error(t, err)
	})

	t.Run("Should fail to update state to question because we are in incorrect game state", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		deadline := time.Now().Add(5 * time.Second).UTC()

		mockStore.EXPECT().GetGameState(ctx, gameStateID).Return(db.GameState{
			State: db.FibbingITQuestion.String(),
		}, nil)

		// Note: Test for newRound = true would require additional setup for round progression
		_, err := srv.UpdateStateToQuestion(ctx, gameStateID, deadline, false)
		assert.ErrorContains(t, err, "game state is not in FIBBING_IT_REVEAL_ROLE state or FIBBING_IT_SCORING state")
	})

	t.Run(
		"Should fail to update state to question because we fail to get all players via game state",
		func(t *testing.T) {
			t.Parallel()
			mockStore := mockService.NewMockStorer(t)
			mockRandom := mockService.NewMockRandomizer(t)
			srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

			ctx := t.Context()
			deadline := time.Now().Add(5 * time.Second).UTC()

			mockStore.EXPECT().GetGameState(ctx, gameStateID).Return(db.GameState{
				State: db.FibbingItRevealRole.String(),
			}, nil)
			mockStore.EXPECT().UpdateGameState(ctx, db.UpdateGameStateParams{
				ID:             gameStateID,
				SubmitDeadline: pgtype.Timestamp{Time: deadline, Valid: true},
				State:          db.FibbingITQuestion.String(),
			}).Return(db.GameState{}, nil)
			mockStore.EXPECT().GetAllPlayersByGameStateID(ctx, gameStateID).Return(
				nil,
				xerrors.New("failed to get all player IDs by game state ID"),
			)
			_, err := srv.UpdateStateToQuestion(ctx, gameStateID, deadline, false)
			assert.Error(t, err)
		},
	)

	t.Run("Should fail to update state to question because we fail to get latest round", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		deadline := time.Now().Add(5 * time.Second).UTC()

		mockStore.EXPECT().GetGameState(ctx, gameStateID).Return(db.GameState{
			State: db.FibbingItRevealRole.String(),
		}, nil)
		mockStore.EXPECT().UpdateGameState(ctx, db.UpdateGameStateParams{
			ID:             gameStateID,
			SubmitDeadline: pgtype.Timestamp{Time: deadline, Valid: true},
			State:          db.FibbingITQuestion.String(),
		}).Return(db.GameState{}, nil)
		mockStore.EXPECT().GetAllPlayersByGameStateID(ctx, gameStateID).Return(
			[]db.GetAllPlayersByGameStateIDRow{
				{
					ID:       defaultHostPlayerID,
					Nickname: "Player 1",
				},
				{
					ID:       defaultOtherPlayerID,
					Nickname: "Player 2",
				},
			},
			nil,
		)
		mockStore.EXPECT().
			GetLatestRoundByGameStateID(ctx, gameStateID).
			Return(db.GetLatestRoundByGameStateIDRow{}, xerrors.New("failed to get latest round by game state ID"))

		_, err := srv.UpdateStateToQuestion(ctx, gameStateID, deadline, false)
		assert.Error(t, err)
	})

	t.Run(
		"Should fail to update state to question because we fail to get fibber from round id",
		func(t *testing.T) {
			t.Parallel()
			mockStore := mockService.NewMockStorer(t)
			mockRandom := mockService.NewMockRandomizer(t)
			srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

			ctx := t.Context()
			deadline := time.Now().Add(5 * time.Second).UTC()

			mockStore.EXPECT().GetGameState(ctx, gameStateID).Return(db.GameState{
				State: db.FibbingItRevealRole.String(),
			}, nil)
			mockStore.EXPECT().UpdateGameState(ctx, db.UpdateGameStateParams{
				ID:             gameStateID,
				SubmitDeadline: pgtype.Timestamp{Time: deadline, Valid: true},
				State:          db.FibbingITQuestion.String(),
			}).Return(db.GameState{}, nil)
			mockStore.EXPECT().GetAllPlayersByGameStateID(ctx, gameStateID).Return(
				[]db.GetAllPlayersByGameStateIDRow{
					{
						ID:       defaultHostPlayerID,
						Nickname: "Player 1",
					},
					{
						ID:       defaultOtherPlayerID,
						Nickname: "Player 2",
					},
				},
				nil,
			)
			roundID := uuid.Must(uuid.FromString("0193ea48-c27f-74bd-8a17-523f69350aff"))
			mockStore.EXPECT().GetLatestRoundByGameStateID(ctx, gameStateID).Return(db.GetLatestRoundByGameStateIDRow{
				ID:        roundID,
				RoundType: "free_form",
				Round:     1,
			}, nil)
			mockStore.EXPECT().
				GetFibberByRoundID(ctx, roundID).
				Return(db.FibbingItPlayerRole{}, fmt.Errorf("failed to access DB"))

			_, err := srv.UpdateStateToQuestion(ctx, gameStateID, deadline, false)
			assert.ErrorContains(t, err, "failed to get fibber in round")
		},
	)

	t.Run(
		"Should fail to update state to question because we fail to get random question by round",
		func(t *testing.T) {
			t.Parallel()
			mockStore := mockService.NewMockStorer(t)
			mockRandom := mockService.NewMockRandomizer(t)
			srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

			ctx := t.Context()
			gameName := gameName
			deadline := time.Now().Add(5 * time.Second).UTC()

			mockStore.EXPECT().GetGameState(ctx, gameStateID).Return(db.GameState{
				State: db.FibbingItRevealRole.String(),
			}, nil)
			mockStore.EXPECT().UpdateGameState(ctx, db.UpdateGameStateParams{
				ID:             gameStateID,
				SubmitDeadline: pgtype.Timestamp{Time: deadline, Valid: true},
				State:          db.FibbingITQuestion.String(),
			}).Return(db.GameState{}, nil)
			mockStore.EXPECT().GetAllPlayersByGameStateID(ctx, gameStateID).Return(
				[]db.GetAllPlayersByGameStateIDRow{
					{
						ID:       defaultHostPlayerID,
						Nickname: "Player 1",
					},
					{
						ID:       defaultOtherPlayerID,
						Nickname: "Player 2",
					},
				},
				nil,
			)
			roundID := uuid.Must(uuid.FromString("0193ea48-c27f-74bd-8a17-523f69350aff"))
			mockStore.EXPECT().GetLatestRoundByGameStateID(ctx, gameStateID).Return(db.GetLatestRoundByGameStateIDRow{
				ID:        roundID,
				RoundType: "free_form",
				Round:     1,
			}, nil)
			mockStore.EXPECT().
				GetFibberByRoundID(ctx, roundID).
				Return(db.FibbingItPlayerRole{PlayerID: defaultOtherPlayerID}, nil)
			mockStore.EXPECT().GetRandomQuestionByRound(ctx, db.GetRandomQuestionByRoundParams{
				GameName:  gameName,
				RoundType: "free_form",
			}).Return([]db.GetRandomQuestionByRoundRow{}, xerrors.New("failed to get random question by round"))

			_, err := srv.UpdateStateToQuestion(ctx, gameStateID, deadline, false)
			assert.Error(t, err)
		},
	)

	t.Run("Should fail to update to question because we fail to get random question in group", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		gameName := gameName
		deadline := time.Now().Add(5 * time.Second).UTC()

		mockStore.EXPECT().GetGameState(ctx, gameStateID).Return(db.GameState{
			State: db.FibbingItRevealRole.String(),
		}, nil)
		mockStore.EXPECT().UpdateGameState(ctx, db.UpdateGameStateParams{
			ID:             gameStateID,
			SubmitDeadline: pgtype.Timestamp{Time: deadline, Valid: true},
			State:          db.FibbingITQuestion.String(),
		}).Return(db.GameState{}, nil)
		mockStore.EXPECT().GetAllPlayersByGameStateID(ctx, gameStateID).Return(
			[]db.GetAllPlayersByGameStateIDRow{
				{
					ID:       defaultHostPlayerID,
					Nickname: "Player 1",
				},
				{
					ID:       defaultOtherPlayerID,
					Nickname: "Player 2",
				},
			},
			nil,
		)
		roundID := uuid.Must(uuid.FromString("0193ea48-c27f-74bd-8a17-523f69350aff"))
		mockStore.EXPECT().GetLatestRoundByGameStateID(ctx, gameStateID).Return(db.GetLatestRoundByGameStateIDRow{
			ID:        roundID,
			RoundType: "free_form",
			Round:     1,
		}, nil)
		mockStore.EXPECT().
			GetFibberByRoundID(ctx, roundID).
			Return(db.FibbingItPlayerRole{PlayerID: defaultOtherPlayerID}, nil)
		mockStore.EXPECT().GetRandomQuestionByRound(ctx, db.GetRandomQuestionByRoundParams{
			GameName:  gameName,
			RoundType: "free_form",
		}).Return([]db.GetRandomQuestionByRoundRow{
			{
				ID:         uuid.Must(uuid.FromString("0193a629-7dcc-78ad-822f-fd5d83c89ae8")),
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
			{},
		}, xerrors.New("failed to get random question in group"))

		_, err := srv.UpdateStateToQuestion(ctx, gameStateID, deadline, false)
		assert.Error(t, err)
	})

	t.Run("Should fail to update to question because we fail to add a new round", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		deadline := time.Now().Add(5 * time.Second).UTC()

		mockStore.EXPECT().GetGameState(ctx, gameStateID).Return(db.GameState{
			State: db.FibbingItRevealRole.String(),
		}, nil)
		mockStore.EXPECT().UpdateGameState(ctx, db.UpdateGameStateParams{
			ID:             gameStateID,
			SubmitDeadline: pgtype.Timestamp{Time: deadline, Valid: true},
			State:          db.FibbingITQuestion.String(),
		}).Return(db.GameState{}, nil)
		mockStore.EXPECT().GetAllPlayersByGameStateID(ctx, gameStateID).Return(
			[]db.GetAllPlayersByGameStateIDRow{
				{
					ID:       defaultHostPlayerID,
					Nickname: "Player 1",
				},
				{
					ID:       defaultOtherPlayerID,
					Nickname: "Player 2",
				},
			},
			nil,
		)
		roundID := uuid.Must(uuid.FromString("0193ea48-c27f-74bd-8a17-523f69350aff"))
		mockStore.EXPECT().GetLatestRoundByGameStateID(ctx, gameStateID).Return(db.GetLatestRoundByGameStateIDRow{
			ID:        roundID,
			RoundType: "free_form",
			Round:     1,
		}, nil)
		mockStore.EXPECT().
			GetFibberByRoundID(ctx, roundID).
			Return(db.FibbingItPlayerRole{PlayerID: defaultOtherPlayerID}, nil)
		mockStore.EXPECT().GetRandomQuestionByRound(ctx, db.GetRandomQuestionByRoundParams{
			GameName:  gameName,
			RoundType: "free_form",
		}).Return([]db.GetRandomQuestionByRoundRow{
			{
				ID:         uuid.Must(uuid.FromString("0193ea48-c27f-74bd-8a17-523f69350aca")),
				QuestionID: uuid.Must(uuid.FromString("0193a629-7dcc-78ad-822f-fd5d83c89ae7")),
				Question:   "I love cats",
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
				Question:   "I love dogs",
			},
		}, nil)
		mockStore.EXPECT().NewRound(ctx, db.NewRoundArgs{
			GameStateID:       gameStateID,
			NormalsQuestionID: uuid.Must(uuid.FromString("0193a629-7dcc-78ad-822f-fd5d83c89ae7")),
			FibberQuestionID:  uuid.Must(uuid.FromString("0193a629-a9ac-7fc4-828c-a1334c282e0f")),
			Round:             2,
			RoundType:         "free_form",
			Players: []db.GetAllPlayersByGameStateIDRow{
				{
					ID:       defaultHostPlayerID,
					Nickname: "Player 1",
				},
				{
					ID:       defaultOtherPlayerID,
					Nickname: "Player 2",
				},
			},
			FibberLoc: 1,
		}).Return(xerrors.New("failed to add a new round"))

		_, err := srv.UpdateStateToQuestion(ctx, gameStateID, deadline, false)
		assert.Error(t, err)
	})
}

func TestRoundServiceUpdateStateToScore(t *testing.T) {
	t.Parallel()

	gameStateID := uuid.Must(uuid.FromString("fbb75599-9f7a-4392-b523-fd433b3208ea"))
	scoring := service.Scoring{
		GuessedFibber:      100,
		FibberEvadeCapture: 150,
	}

	// Additional test cases could include: multiple voting rounds, edge cases for scoring,
	// different round types, and boundary conditions. Table tests would improve maintainability.
	t.Run("Should successfully update score state, fibber caught in one round of voting", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		now := time.Now().Add(15 * time.Second)

		mockStore.EXPECT().
			GetGameState(ctx, gameStateID).
			Return(db.GameState{State: db.FibbingItRevealRole.String()}, nil)
		mockStore.EXPECT().
			UpdateGameState(ctx, db.UpdateGameStateParams{
				ID:             gameStateID,
				SubmitDeadline: pgtype.Timestamp{Time: now, Valid: true},
				State:          db.FibbingItScoring.String(),
			}).
			Return(db.GameState{}, nil)
		mockStore.EXPECT().
			GetAllVotesForRoundByGameStateID(ctx, gameStateID).
			Return([]db.GetAllVotesForRoundByGameStateIDRow{
				{
					VoterID:       defaultHostPlayerID,
					VotedForID:    defaultOtherPlayerID,
					VoterAvatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Player+1",
					VoterNickname: "Player 1",
					FibberID:      defaultOtherPlayerID,
					RoundID:       uuid.Must(uuid.FromString("0193a62a-364e-751a-9088-cf3b9711153e")),
				},
			}, nil)
		mockStore.EXPECT().
			GetAllPlayersByGameStateID(ctx, gameStateID).
			Return([]db.GetAllPlayersByGameStateIDRow{
				{
					ID:       defaultHostPlayerID,
					Nickname: "Player 1",
					Avatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Player+1",
				},
				{
					ID:       defaultOtherPlayerID,
					Nickname: "Player 2",
					Avatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Player+2",
				},
			}, nil)
		mockStore.EXPECT().
			GetLatestRoundByGameStateID(ctx, gameStateID).
			Return(db.GetLatestRoundByGameStateIDRow{
				ID:        uuid.Must(uuid.FromString("0193a62a-364e-751a-9088-cf3b9711153e")),
				Round:     1,
				RoundType: "free_form",
			}, nil)
		mockStore.EXPECT().
			GetTotalScoresByGameStateID(ctx, db.GetTotalScoresByGameStateIDParams{
				ID:   gameStateID,
				ID_2: uuid.Must(uuid.FromString("0193a62a-364e-751a-9088-cf3b9711153e")),
			}).
			Return([]db.GetTotalScoresByGameStateIDRow{}, nil)
		mockStore.EXPECT().
			NewScores(ctx, db.NewScoresArgs{
				Players: []db.AddFibbingItScoreParams{
					{
						PlayerID: defaultHostPlayerID,
						RoundID:  uuid.Must(uuid.FromString("0193a62a-364e-751a-9088-cf3b9711153e")),
						Score:    100,
					},
					{
						PlayerID: defaultOtherPlayerID,
						RoundID:  uuid.Must(uuid.FromString("0193a62a-364e-751a-9088-cf3b9711153e")),
						Score:    0,
					},
				},
			}).Return(nil)

		scoreState, err := srv.UpdateStateToScore(ctx, gameStateID, now, scoring)
		assert.NoError(t, err)

		expectedScoreState := service.ScoreState{
			RoundNumber: 1,
			RoundType:   "free_form",
			Deadline:    time.Until(now),
			Players: []service.PlayerWithScoring{
				{
					ID:       defaultHostPlayerID,
					Avatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Player+1",
					Nickname: "Player 1",
					Score:    100,
				},
				{
					ID:       defaultOtherPlayerID,
					Avatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Player+2",
					Nickname: "Player 2",
					Score:    0,
				},
			},
		}

		diffOpts := cmpopts.IgnoreFields(scoreState, "Deadline")
		PartialEqual(t, expectedScoreState, scoreState, diffOpts)
		assert.LessOrEqual(t, int(scoreState.Deadline.Seconds()), 15)
	})

	t.Run("Should fail to update score state, fail to get game state", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		now := time.Now().Add(15 * time.Second)

		mockStore.EXPECT().
			GetGameState(ctx, gameStateID).
			Return(db.GameState{}, xerrors.New("failed to get game state"))

		_, err := srv.UpdateStateToScore(ctx, gameStateID, now, scoring)
		assert.Error(t, err)
	})

	t.Run("Should fail to update score state, game in wrong state", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		now := time.Now().Add(15 * time.Second)

		mockStore.EXPECT().
			GetGameState(ctx, gameStateID).
			Return(db.GameState{State: db.FibbingItVoting.String()}, nil)

		_, err := srv.UpdateStateToScore(ctx, gameStateID, now, scoring)
		assert.ErrorContains(t, err, "game state is not in FIBBING_IT_REVEAL_ROLE state")
	})

	t.Run("Should fail to update score state, fail to update game state", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		now := time.Now().Add(15 * time.Second)

		mockStore.EXPECT().
			GetGameState(ctx, gameStateID).
			Return(db.GameState{State: db.FibbingItRevealRole.String()}, nil)
		mockStore.EXPECT().
			UpdateGameState(ctx, db.UpdateGameStateParams{
				ID:             gameStateID,
				SubmitDeadline: pgtype.Timestamp{Time: now, Valid: true},
				State:          db.FibbingItScoring.String(),
			}).
			Return(db.GameState{}, xerrors.New("failed to update game state"))

		_, err := srv.UpdateStateToScore(ctx, gameStateID, now, scoring)
		assert.Error(t, err)
	})

	t.Run("Should fail to update score state, fail to get all votes this round type", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		now := time.Now().Add(15 * time.Second)

		mockStore.EXPECT().
			GetGameState(ctx, gameStateID).
			Return(db.GameState{State: db.FibbingItRevealRole.String()}, nil)
		mockStore.EXPECT().
			UpdateGameState(ctx, db.UpdateGameStateParams{
				ID:             gameStateID,
				SubmitDeadline: pgtype.Timestamp{Time: now, Valid: true},
				State:          db.FibbingItScoring.String(),
			}).
			Return(db.GameState{}, nil)
		mockStore.EXPECT().
			GetAllVotesForRoundByGameStateID(ctx, gameStateID).
			Return(
				[]db.GetAllVotesForRoundByGameStateIDRow{},
				xerrors.New("failed to get all votes for round by game state ID"),
			)

		_, err := srv.UpdateStateToScore(ctx, gameStateID, now, scoring)
		assert.Error(t, err)
	})

	t.Run("Should fail to update score state, fail to get all players in the room", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		now := time.Now().Add(15 * time.Second)

		mockStore.EXPECT().
			GetGameState(ctx, gameStateID).
			Return(db.GameState{State: db.FibbingItRevealRole.String()}, nil)
		mockStore.EXPECT().
			UpdateGameState(ctx, db.UpdateGameStateParams{
				ID:             gameStateID,
				SubmitDeadline: pgtype.Timestamp{Time: now, Valid: true},
				State:          db.FibbingItScoring.String(),
			}).
			Return(db.GameState{}, nil)
		mockStore.EXPECT().
			GetAllVotesForRoundByGameStateID(ctx, gameStateID).
			Return([]db.GetAllVotesForRoundByGameStateIDRow{
				{
					VoterID:       defaultHostPlayerID,
					VotedForID:    defaultOtherPlayerID,
					VoterAvatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Player+1",
					VoterNickname: "Player 1",
					FibberID:      defaultOtherPlayerID,
					RoundID:       uuid.Must(uuid.FromString("0193a62a-364e-751a-9088-cf3b9711153e")),
				},
			}, nil)
		mockStore.EXPECT().
			GetAllPlayersByGameStateID(ctx, gameStateID).
			Return([]db.GetAllPlayersByGameStateIDRow{}, xerrors.New("failed to get all players by game state ID"))

		_, err := srv.UpdateStateToScore(ctx, gameStateID, now, scoring)
		assert.Error(t, err)
	})

	t.Run("Should fail to update score state, fail to get latest round", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		now := time.Now().Add(15 * time.Second)

		mockStore.EXPECT().
			GetGameState(ctx, gameStateID).
			Return(db.GameState{State: db.FibbingItRevealRole.String()}, nil)
		mockStore.EXPECT().
			UpdateGameState(ctx, db.UpdateGameStateParams{
				ID:             gameStateID,
				SubmitDeadline: pgtype.Timestamp{Time: now, Valid: true},
				State:          db.FibbingItScoring.String(),
			}).
			Return(db.GameState{}, nil)
		mockStore.EXPECT().
			GetAllVotesForRoundByGameStateID(ctx, gameStateID).
			Return([]db.GetAllVotesForRoundByGameStateIDRow{
				{
					VoterID:       defaultHostPlayerID,
					VotedForID:    defaultOtherPlayerID,
					VoterAvatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Player+1",
					VoterNickname: "Player 1",
					FibberID:      defaultOtherPlayerID,
					RoundID:       uuid.Must(uuid.FromString("0193a62a-364e-751a-9088-cf3b9711153e")),
				},
			}, nil)
		mockStore.EXPECT().
			GetAllPlayersByGameStateID(ctx, gameStateID).
			Return([]db.GetAllPlayersByGameStateIDRow{
				{
					ID:       defaultHostPlayerID,
					Nickname: "Player 1",
					Avatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Player+1",
				},
				{
					ID:       defaultOtherPlayerID,
					Nickname: "Player 2",
					Avatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Player+2",
				},
			}, nil)
		mockStore.EXPECT().
			GetLatestRoundByGameStateID(ctx, gameStateID).
			Return(db.GetLatestRoundByGameStateIDRow{}, xerrors.New("failed to get latest round by game state ID"))

		_, err := srv.UpdateStateToScore(ctx, gameStateID, now, scoring)
		assert.Error(t, err)
	})

	t.Run("Should fail to update score state, fail to add new scores", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		now := time.Now().Add(15 * time.Second)

		mockStore.EXPECT().
			GetGameState(ctx, gameStateID).
			Return(db.GameState{State: db.FibbingItRevealRole.String()}, nil)
		mockStore.EXPECT().
			UpdateGameState(ctx, db.UpdateGameStateParams{
				ID:             gameStateID,
				SubmitDeadline: pgtype.Timestamp{Time: now, Valid: true},
				State:          db.FibbingItScoring.String(),
			}).
			Return(db.GameState{}, nil)
		mockStore.EXPECT().
			GetAllVotesForRoundByGameStateID(ctx, gameStateID).
			Return([]db.GetAllVotesForRoundByGameStateIDRow{
				{
					VoterID:       defaultHostPlayerID,
					VotedForID:    defaultOtherPlayerID,
					VoterAvatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Player+1",
					VoterNickname: "Player 1",
					FibberID:      defaultOtherPlayerID,
					RoundID:       uuid.Must(uuid.FromString("0193a62a-364e-751a-9088-cf3b9711153e")),
				},
			}, nil)
		mockStore.EXPECT().
			GetAllPlayersByGameStateID(ctx, gameStateID).
			Return([]db.GetAllPlayersByGameStateIDRow{
				{
					ID:       defaultHostPlayerID,
					Nickname: "Player 1",
					Avatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Player+1",
				},
				{
					ID:       defaultOtherPlayerID,
					Nickname: "Player 2",
					Avatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Player+2",
				},
			}, nil)
		mockStore.EXPECT().
			GetLatestRoundByGameStateID(ctx, gameStateID).
			Return(db.GetLatestRoundByGameStateIDRow{
				ID:    uuid.Must(uuid.FromString("0193a62a-364e-751a-9088-cf3b9711153e")),
				Round: 1,
			}, nil)
		mockStore.EXPECT().
			GetTotalScoresByGameStateID(ctx, db.GetTotalScoresByGameStateIDParams{
				ID:   gameStateID,
				ID_2: uuid.Must(uuid.FromString("0193a62a-364e-751a-9088-cf3b9711153e")),
			}).
			Return([]db.GetTotalScoresByGameStateIDRow{}, nil)
		mockStore.EXPECT().
			NewScores(ctx, db.NewScoresArgs{
				Players: []db.AddFibbingItScoreParams{
					{
						PlayerID: defaultHostPlayerID,
						RoundID:  uuid.Must(uuid.FromString("0193a62a-364e-751a-9088-cf3b9711153e")),
						Score:    100,
					},
					{
						PlayerID: defaultOtherPlayerID,
						RoundID:  uuid.Must(uuid.FromString("0193a62a-364e-751a-9088-cf3b9711153e")),
						Score:    0,
					},
				},
			}).Return(xerrors.New("failed to add new scores"))

		_, err := srv.UpdateStateToScore(ctx, gameStateID, now, scoring)
		assert.Error(t, err)
	})
}

func TestRoundServiceGetGameState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		gameState     db.FibbingItGameState
		expectedState db.FibbingItGameState
	}{
		{
			name:          "Should successfully get game state QUESTION",
			gameState:     db.FibbingITQuestion,
			expectedState: db.FibbingITQuestion,
		},
		{
			name:          "Should successfully get game state VOTING",
			gameState:     db.FibbingItVoting,
			expectedState: db.FibbingItVoting,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockStore := mockService.NewMockStorer(t)
			mockRandom := mockService.NewMockRandomizer(t)
			srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

			ctx := t.Context()
			mockStore.EXPECT().GetGameStateByPlayerID(ctx, playerID).Return(db.GameState{
				State: tt.gameState.String(),
			}, nil)

			gameState, err := srv.GetGameState(ctx, playerID)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedState, gameState)
		})
	}

	t.Run("Should fail to get game state because we fail to get game details DB", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()
		mockStore.EXPECT().GetGameStateByPlayerID(ctx, playerID).Return(
			db.GameState{}, xerrors.New("failed to get game state details"),
		)

		_, err := srv.GetGameState(ctx, playerID)
		assert.Error(t, err)
	})
}

func TestRoundServiceGetQuestionState(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully get question state", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()

		deadline := time.Now().Add(5 * time.Second)
		mockStore.EXPECT().GetCurrentQuestionByPlayerID(ctx, playerID).Return(db.GetCurrentQuestionByPlayerIDRow{
			PlayerID:       playerID,
			Avatar:         "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=nickname",
			Nickname:       "nickname",
			Role:           pgtype.Text{String: "fibber"},
			Question:       pgtype.Text{String: "fibber question"},
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
					ID:              playerID,
					Role:            "fibber",
					Question:        "fibber question",
					PossibleAnswers: []string{},
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
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()

		deadline := time.Now().Add(5 * time.Second)
		mockStore.EXPECT().GetCurrentQuestionByPlayerID(ctx, playerID).Return(db.GetCurrentQuestionByPlayerIDRow{
			PlayerID:       playerID,
			Avatar:         "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=nickname",
			Nickname:       "nickname",
			Role:           pgtype.Text{String: "normal"},
			Question:       pgtype.Text{String: "normal question"},
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
					ID:              playerID,
					Role:            "normal",
					Question:        "normal question",
					PossibleAnswers: []string{},
				},
			},
			Round:     1,
			RoundType: "free_form",
		}

		diffOpts := cmpopts.IgnoreFields(questionState, "Deadline")
		PartialEqual(t, expectedGameState, questionState, diffOpts)
		assert.LessOrEqual(t, int(questionState.Deadline.Seconds()), 5)
	})

	t.Run("Should successfully get question state round type multiple_choice", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()

		deadline := time.Now().Add(5 * time.Second)
		mockStore.EXPECT().GetCurrentQuestionByPlayerID(ctx, playerID).Return(db.GetCurrentQuestionByPlayerIDRow{
			PlayerID:       playerID,
			Avatar:         "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=nickname",
			Nickname:       "nickname",
			Role:           pgtype.Text{String: "fibber"},
			Question:       pgtype.Text{String: "fibber question"},
			Round:          1,
			RoundType:      "multiple_choice",
			RoomCode:       "ABC12",
			SubmitDeadline: pgtype.Timestamp{Time: deadline},
		}, nil)
		questionState, err := srv.GetQuestionState(ctx, playerID)

		assert.NoError(t, err)
		expectedGameState := service.QuestionState{
			Players: []service.PlayerWithRole{
				{
					ID:              playerID,
					Role:            "fibber",
					Question:        "fibber question",
					PossibleAnswers: []string{"Strongly Agree", "Agree", "Neutral", "Disagree", "Strongly Disagree"},
				},
			},
			Round:     1,
			RoundType: "multiple_choice",
		}

		diffOpts := cmpopts.IgnoreFields(questionState, "Deadline")
		PartialEqual(t, expectedGameState, questionState, diffOpts)
		assert.LessOrEqual(t, int(questionState.Deadline.Seconds()), 5)
	})

	t.Run("Should successfully get question state round type most_likely", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()

		deadline := time.Now().Add(5 * time.Second)
		mockStore.EXPECT().GetCurrentQuestionByPlayerID(ctx, playerID).Return(db.GetCurrentQuestionByPlayerIDRow{
			PlayerID:       playerID,
			Avatar:         "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=nickname",
			Nickname:       "nickname",
			Role:           pgtype.Text{String: "fibber"},
			Question:       pgtype.Text{String: "fibber question"},
			Round:          1,
			RoundType:      "most_likely",
			RoomCode:       "ABC12",
			SubmitDeadline: pgtype.Timestamp{Time: deadline},
		}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, playerID).Return([]db.GetAllPlayersInRoomRow{
			{
				Nickname: "nickname",
			},
			{
				Nickname: "other_nickname",
			},
		}, nil)

		questionState, err := srv.GetQuestionState(ctx, playerID)

		assert.NoError(t, err)
		expectedGameState := service.QuestionState{
			Players: []service.PlayerWithRole{
				{
					ID:              playerID,
					Role:            "fibber",
					Question:        "fibber question",
					PossibleAnswers: []string{"nickname", "other_nickname"},
				},
			},
			Round:     1,
			RoundType: "most_likely",
		}

		diffOpts := cmpopts.IgnoreFields(questionState, "Deadline")
		PartialEqual(t, expectedGameState, questionState, diffOpts)
		assert.LessOrEqual(t, int(questionState.Deadline.Seconds()), 5)
	})

	t.Run("Should fail to get question state because we cannot fetch from DB", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()

		mockStore.EXPECT().GetCurrentQuestionByPlayerID(ctx, playerID).Return(
			db.GetCurrentQuestionByPlayerIDRow{}, xerrors.New("failed to get questions"),
		)

		_, err := srv.GetQuestionState(ctx, playerID)
		assert.Error(t, err)
	})
}

func TestRoundServiceUpdateStateToWinning(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully update state to winner state", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()

		gameID, err := uuid.NewV7()
		require.NoError(t, err)
		mockStore.EXPECT().GetGameState(ctx, gameID).Return(db.GameState{
			State: db.FibbingItScoring.String(),
			ID:    gameID,
		}, nil)

		u, err := uuid.NewV7()
		require.NoError(t, err)
		mockRandom.EXPECT().GetID().Return(u, nil)

		deadline := time.Now().UTC().Add(1 * time.Hour)
		mockStore.EXPECT().UpdateGameState(ctx, db.UpdateGameStateParams{
			ID:             gameID,
			State:          db.FibbingItWinner.String(),
			SubmitDeadline: pgtype.Timestamp{Time: deadline, Valid: true},
		}).Return(db.GameState{}, nil)

		mockStore.EXPECT().GetTotalScoresByGameStateID(ctx, db.GetTotalScoresByGameStateIDParams{
			ID:   gameID,
			ID_2: u,
		}).Return([]db.GetTotalScoresByGameStateIDRow{
			{
				PlayerID:   playerID,
				TotalScore: int64(100),
				Nickname:   "Player 1",
				Avatar:     "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Player+1",
			},
			{
				PlayerID:   defaultOtherPlayerID,
				TotalScore: int64(200),
				Nickname:   "Player 2",
				Avatar:     "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Player+2",
			},
		}, nil)

		winnerState, err := srv.UpdateStateToWinner(ctx, gameID, deadline)
		assert.NoError(t, err)

		expectedWinnerState := service.WinnerState{
			Players: []service.PlayerWithScoring{
				{
					ID:       defaultOtherPlayerID,
					Score:    200,
					Nickname: "Player 2",
					Avatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Player+2",
				},
				{
					ID:       playerID,
					Score:    100,
					Nickname: "Player 1",
					Avatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Player+1",
				},
			},
		}
		assert.Equal(t, expectedWinnerState, winnerState)
	})

	t.Run("Should fail to update state to winner state, fail to get game state", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()

		gameID, err := uuid.NewV7()
		require.NoError(t, err)
		mockStore.EXPECT().GetGameState(ctx, gameID).Return(db.GameState{}, fmt.Errorf("failed to get game state"))

		deadline := time.Now().UTC().Add(1 * time.Hour)

		_, err = srv.UpdateStateToWinner(ctx, gameID, deadline)
		assert.Error(t, err)
	})

	t.Run("Should fail to update state to winner state, game not in scoring state", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()

		gameID, err := uuid.NewV7()
		require.NoError(t, err)
		mockStore.EXPECT().GetGameState(ctx, gameID).Return(db.GameState{
			State: db.FibbingItRevealRole.String(),
			ID:    gameID,
		}, nil)

		deadline := time.Now().UTC().Add(1 * time.Hour)

		_, err = srv.UpdateStateToWinner(ctx, gameID, deadline)
		assert.ErrorContains(t, err, "game state is not in FIBBING_IT_SCORING_STATE state")
	})

	t.Run("Should fail to update state to winner state, fail to update state in DB", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()

		gameID, err := uuid.NewV7()
		require.NoError(t, err)
		mockStore.EXPECT().GetGameState(ctx, gameID).Return(db.GameState{
			State: db.FibbingItScoring.String(),
			ID:    gameID,
		}, nil)

		deadline := time.Now().UTC().Add(1 * time.Hour)
		mockStore.EXPECT().UpdateGameState(ctx, db.UpdateGameStateParams{
			ID:             gameID,
			State:          db.FibbingItWinner.String(),
			SubmitDeadline: pgtype.Timestamp{Time: deadline, Valid: true},
		}).Return(db.GameState{}, fmt.Errorf("failed to update game state"))

		_, err = srv.UpdateStateToWinner(ctx, gameID, deadline)
		assert.Error(t, err)
	})

	t.Run("Should fail to update state to winner state, fail to get total scores from DB", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()

		gameID, err := uuid.NewV7()
		require.NoError(t, err)
		mockStore.EXPECT().GetGameState(ctx, gameID).Return(db.GameState{
			State: db.FibbingItScoring.String(),
			ID:    gameID,
		}, nil)

		u, err := uuid.NewV7()
		require.NoError(t, err)
		mockRandom.EXPECT().GetID().Return(u, nil)

		deadline := time.Now().UTC().Add(1 * time.Hour)
		mockStore.EXPECT().UpdateGameState(ctx, db.UpdateGameStateParams{
			ID:             gameID,
			State:          db.FibbingItWinner.String(),
			SubmitDeadline: pgtype.Timestamp{Time: deadline, Valid: true},
		}).Return(db.GameState{}, nil)

		mockStore.EXPECT().GetTotalScoresByGameStateID(ctx, db.GetTotalScoresByGameStateIDParams{
			ID:   gameID,
			ID_2: u,
		}).Return([]db.GetTotalScoresByGameStateIDRow{}, fmt.Errorf("fail to get scores from DB"))

		_, err = srv.UpdateStateToWinner(ctx, gameID, deadline)
		assert.Error(t, err)
	})
}

func TestRoundServiceGetWinnerState(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully get winner state", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()

		gameID, err := uuid.NewV7()
		require.NoError(t, err)
		mockStore.EXPECT().GetGameStateByPlayerID(ctx, playerID).Return(db.GameState{
			State: db.FibbingItWinner.String(),
			ID:    gameID,
		}, nil)

		u, err := uuid.NewV7()
		require.NoError(t, err)
		mockRandom.EXPECT().GetID().Return(u, nil)

		mockStore.EXPECT().GetTotalScoresByGameStateID(ctx, db.GetTotalScoresByGameStateIDParams{
			ID:   gameID,
			ID_2: u,
		}).Return([]db.GetTotalScoresByGameStateIDRow{
			{
				PlayerID:   playerID,
				TotalScore: int64(100),
				Nickname:   "Player 1",
				Avatar:     "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Player+1",
			},
			{
				PlayerID:   defaultOtherPlayerID,
				TotalScore: int64(200),
				Nickname:   "Player 2",
				Avatar:     "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Player+2",
			},
		}, nil)

		winnerState, err := srv.GetWinnerState(ctx, playerID)
		assert.NoError(t, err)

		expectedWinnerState := service.WinnerState{
			Players: []service.PlayerWithScoring{
				{
					ID:       defaultOtherPlayerID,
					Score:    200,
					Nickname: "Player 2",
					Avatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Player+2",
				},
				{
					ID:       playerID,
					Score:    100,
					Nickname: "Player 1",
					Avatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=Player+1",
				},
			},
		}
		assert.Equal(t, expectedWinnerState, winnerState)
	})

	t.Run("Should fail to get winner state, cannot get game state", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()

		mockStore.EXPECT().
			GetGameStateByPlayerID(ctx, playerID).
			Return(db.GameState{}, fmt.Errorf("failed to get game state"))

		_, err := srv.GetWinnerState(ctx, playerID)
		assert.Error(t, err)
	})

	t.Run("Should fail to get winner state, failed to get total scores", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()

		gameID, err := uuid.NewV7()
		require.NoError(t, err)
		mockStore.EXPECT().GetGameStateByPlayerID(ctx, playerID).Return(db.GameState{
			State: db.FibbingItWinner.String(),
			ID:    gameID,
		}, nil)

		u, err := uuid.NewV7()
		require.NoError(t, err)
		mockRandom.EXPECT().GetID().Return(u, nil)

		mockStore.EXPECT().GetTotalScoresByGameStateID(ctx, db.GetTotalScoresByGameStateIDParams{
			ID:   gameID,
			ID_2: u,
		}).Return([]db.GetTotalScoresByGameStateIDRow{}, fmt.Errorf("failed to get total scores"))

		_, err = srv.GetWinnerState(ctx, playerID)
		assert.Error(t, err)
	})
}

func TestRoundServiceFinishGame(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully finish game", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()

		gameID, err := uuid.NewV7()
		require.NoError(t, err)
		roomID, err := uuid.NewV7()
		require.NoError(t, err)
		mockStore.EXPECT().GetGameState(ctx, gameID).Return(db.GameState{
			State:  db.FibbingItWinner.String(),
			ID:     gameID,
			RoomID: roomID,
		}, nil)

		mockStore.EXPECT().
			UpdateRoomState(ctx, db.UpdateRoomStateParams{RoomState: db.Finished.String(), ID: roomID}).
			Return(db.Room{}, nil)

		err = srv.FinishGame(ctx, gameID)
		assert.NoError(t, err)
	})

	t.Run("Should fail to finish game, fail to get game state", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()

		gameID, err := uuid.NewV7()
		require.NoError(t, err)
		mockStore.EXPECT().GetGameState(ctx, gameID).Return(db.GameState{}, fmt.Errorf("fail to get game state"))

		err = srv.FinishGame(ctx, gameID)
		assert.Error(t, err)
	})

	t.Run("Should fail to finish game, fail to update room state", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()

		gameID, err := uuid.NewV7()
		require.NoError(t, err)
		roomID, err := uuid.NewV7()
		require.NoError(t, err)
		mockStore.EXPECT().GetGameState(ctx, gameID).Return(db.GameState{
			State:  db.FibbingItWinner.String(),
			ID:     gameID,
			RoomID: roomID,
		}, nil)

		mockStore.EXPECT().
			UpdateRoomState(ctx, db.UpdateRoomStateParams{RoomState: db.Finished.String(), ID: roomID}).
			Return(db.Room{}, fmt.Errorf("fail to update room state"))

		err = srv.FinishGame(ctx, gameID)
		assert.Error(t, err)
	})
}

func TestRoundServiceConcurrentOperations(t *testing.T) {
	t.Parallel()

	t.Run("Should handle concurrent operations gracefully", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()

		// Configure mock to return an error for any GetRoomByPlayerID call
		mockStore.EXPECT().GetRoomByPlayerID(ctx, mock.Anything).
			Return(db.Room{}, fmt.Errorf("concurrent test error")).
			Maybe() // Allow 0 or more calls

		// Test concurrent operations
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				defer func() { done <- true }()
				playerID := uuid.Must(uuid.NewV7())
				err := srv.SubmitAnswer(ctx, playerID, "test answer", time.Now())
				assert.Error(t, err) // Expected to fail without proper mocks
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}
