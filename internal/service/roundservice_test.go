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

func TestRoundServiceSubmitAnswer(t *testing.T) {
	roomID := uuid.MustParse("0193a62a-4dff-774c-850a-b1fe78e2a8d2")
	roundID := uuid.MustParse("0193a62a-364e-751a-9088-cf3b9711153e")

	t.Run("Should successfully submit answer", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		now := time.Now()
		ctx := context.Background()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(db.Room{
			ID:        roomID,
			RoomState: db.ROOMSTATE_PLAYING.String(),
		}, nil)
		mockStore.EXPECT().GetLatestRoundByPlayerID(ctx, playerID).Return(
			db.GetLatestRoundByPlayerIDRow{
				ID:             roundID,
				SubmitDeadline: pgtype.Timestamp{Time: now.Add(1 * time.Hour)},
			}, nil)

		u := uuid.Must(uuid.NewV7())
		mockRandomizer.EXPECT().GetID().Return(u)
		mockStore.EXPECT().AddFibbingItAnswer(ctx, db.AddFibbingItAnswerParams{
			ID:       u,
			RoundID:  roundID,
			PlayerID: playerID,
			Answer:   "My answer",
		}).Return(db.FibbingItAnswer{}, nil)

		err := srv.SubmitAnswer(ctx, playerID, "My answer", now)
		assert.NoError(t, err)
	})

	t.Run("Should fail to submit answer because we fail to get room details", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		now := time.Now()
		ctx := context.Background()

		mockStore.EXPECT().
			GetRoomByPlayerID(ctx, playerID).
			Return(db.Room{}, fmt.Errorf("failed to get room details"))
		err := srv.SubmitAnswer(ctx, playerID, "My answer", now)
		assert.Error(t, err)
	})

	t.Run("Should fail to submit answer because room not in PLAYING state", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		now := time.Now()
		ctx := context.Background()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(db.Room{
			ID:        roomID,
			RoomState: db.ROOMSTATE_CREATED.String(),
		}, nil)

		err := srv.SubmitAnswer(ctx, playerID, "My answer", now)
		assert.ErrorContains(t, err, "room is not in PLAYING state")
	})

	t.Run("Should fail to submit answer because failed to get latest round", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		service := service.NewRoundService(mockStore, mockRandomizer)

		now := time.Now()
		ctx := context.Background()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(db.Room{
			ID:        roomID,
			RoomState: db.ROOMSTATE_PLAYING.String(),
		}, nil)
		mockStore.EXPECT().GetLatestRoundByPlayerID(ctx, playerID).Return(
			db.GetLatestRoundByPlayerIDRow{},
			fmt.Errorf("failed to get latest round"),
		)

		err := service.SubmitAnswer(ctx, playerID, "My answer", now)
		assert.Error(t, err)
	})

	t.Run("Should fail to submit answer because after submit time", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		now := time.Now()
		ctx := context.Background()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(db.Room{
			ID:        roomID,
			RoomState: db.ROOMSTATE_PLAYING.String(),
		}, nil)
		mockStore.EXPECT().GetLatestRoundByPlayerID(ctx, playerID).Return(
			db.GetLatestRoundByPlayerIDRow{
				ID: roundID,
			}, nil)

		err := srv.SubmitAnswer(ctx, playerID, "My answer", now)
		assert.ErrorContains(t, err, "answer submission deadline has passed")
	})

	t.Run("Should fail to submit answer because failed to add answer to DB", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		now := time.Now()
		ctx := context.Background()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(db.Room{
			ID:        roomID,
			RoomState: db.ROOMSTATE_PLAYING.String(),
		}, nil)
		mockStore.EXPECT().GetLatestRoundByPlayerID(ctx, playerID).Return(
			db.GetLatestRoundByPlayerIDRow{
				ID:             roundID,
				SubmitDeadline: pgtype.Timestamp{Time: now.Add(1 * time.Hour)},
			}, nil)

		mockRandomizer.EXPECT().GetID().Return(defaultHostPlayerID)
		mockStore.EXPECT().AddFibbingItAnswer(ctx, db.AddFibbingItAnswerParams{
			ID:       defaultHostPlayerID,
			RoundID:  roundID,
			PlayerID: playerID,
			Answer:   "My answer",
		}).Return(db.FibbingItAnswer{}, fmt.Errorf("failed to add answer to DB"))

		err := srv.SubmitAnswer(ctx, playerID, "My answer", now)
		assert.Error(t, err)
	})
}

func TestRoundServiceToggleAnswerIsReady(t *testing.T) {
	t.Run("Should successfully toggle answer ready state", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, playerID).Return(db.GameState{
			State:          db.GAMESTATE_FIBBING_IT_QUESTION.String(),
			SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(1 * time.Hour)},
		}, nil)
		mockStore.EXPECT().ToggleAnswerIsReady(ctx, playerID).Return(db.FibbingItAnswer{}, nil)
		mockStore.EXPECT().GetAllPlayerAnswerIsReady(ctx, playerID).Return(false, nil)

		allReady, err := srv.ToggleAnswerIsReady(ctx, playerID, time.Now().UTC())
		assert.NoError(t, err)
		assert.False(t, allReady)
	})

	t.Run("Should successfully toggle answer ready state and return all players are ready", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, playerID).Return(db.GameState{
			State:          db.GAMESTATE_FIBBING_IT_QUESTION.String(),
			SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(1 * time.Hour)},
		}, nil)
		mockStore.EXPECT().ToggleAnswerIsReady(ctx, playerID).Return(db.FibbingItAnswer{}, nil)
		mockStore.EXPECT().GetAllPlayerAnswerIsReady(ctx, playerID).Return(true, nil)

		allReady, err := srv.ToggleAnswerIsReady(ctx, playerID, time.Now().UTC())
		assert.NoError(t, err)
		assert.True(t, allReady)
	})

	t.Run("Should fail to toggle answer ready state, because we fail to get game state", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, playerID).Return(
			db.GameState{}, fmt.Errorf("failed to get state"),
		)

		_, err := srv.ToggleAnswerIsReady(ctx, playerID, time.Now().UTC())
		assert.Error(t, err)
	})

	t.Run("Should fail to toggle answer ready state because game state not in show question", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, playerID).Return(db.GameState{
			State:          db.GAMESTATE_FIBBING_IT_VOTING.String(),
			SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(1 * time.Hour)},
		}, nil)

		_, err := srv.ToggleAnswerIsReady(ctx, playerID, time.Now().UTC())
		assert.ErrorContains(t, err, "room game state is not in FIBBING_IT_QUESTION state")
	})

	t.Run(
		"Should fail to toggle answer ready state, because we fail to toggle answer ready state in DB",
		func(t *testing.T) {
			mockStore := mockService.NewMockStorer(t)
			mockRandomizer := mockService.NewMockRandomizer(t)
			srv := service.NewRoundService(mockStore, mockRandomizer)

			ctx := context.Background()

			mockStore.EXPECT().GetGameStateByPlayerID(ctx, playerID).Return(db.GameState{
				State:          db.GAMESTATE_FIBBING_IT_QUESTION.String(),
				SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(1 * time.Hour)},
			}, nil)
			mockStore.EXPECT().ToggleAnswerIsReady(ctx, playerID).Return(
				db.FibbingItAnswer{}, fmt.Errorf("failed to toggle answer is ready"),
			)

			_, err := srv.ToggleAnswerIsReady(ctx, playerID, time.Now().UTC())
			assert.Error(t, err)
		},
	)

	t.Run(
		"Should fail to toggle toggle answer ready state because we fail to get all players ready status from DB",
		func(t *testing.T) {
			mockStore := mockService.NewMockStorer(t)
			mockRandomizer := mockService.NewMockRandomizer(t)
			srv := service.NewRoundService(mockStore, mockRandomizer)

			ctx := context.Background()

			mockStore.EXPECT().GetGameStateByPlayerID(ctx, playerID).Return(db.GameState{
				State:          db.GAMESTATE_FIBBING_IT_QUESTION.String(),
				SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(1 * time.Hour)},
			}, nil)
			mockStore.EXPECT().ToggleAnswerIsReady(ctx, playerID).Return(db.FibbingItAnswer{}, nil)
			mockStore.EXPECT().GetAllPlayerAnswerIsReady(ctx, playerID).Return(
				false, fmt.Errorf("failed to get player answer is ready status"),
			)

			_, err := srv.ToggleAnswerIsReady(ctx, playerID, time.Now().UTC())
			assert.Error(t, err)
		},
	)

	t.Run(
		"Should fail to toggle toggle answer ready state because after deadline",
		func(t *testing.T) {
			mockStore := mockService.NewMockStorer(t)
			mockRandomizer := mockService.NewMockRandomizer(t)
			srv := service.NewRoundService(mockStore, mockRandomizer)

			ctx := context.Background()

			mockStore.EXPECT().GetGameStateByPlayerID(ctx, playerID).Return(db.GameState{
				State:          db.GAMESTATE_FIBBING_IT_QUESTION.String(),
				SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(-1 * time.Second)},
			}, nil)

			_, err := srv.ToggleAnswerIsReady(ctx, playerID, time.Now().UTC())
			assert.Error(t, err)
		},
	)
}

func TestRoundServiceUpdateStateToVoting(t *testing.T) {
	gameStateID := uuid.MustParse("fbb75599-9f7a-4392-b523-fd433b3208ea")
	roundID := uuid.MustParse("0193a62a-364e-751a-9088-cf3b9711153e")

	t.Run("Should successfully update state to voting", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		ctx := context.Background()
		now := time.Now().Add(30 * time.Second)

		mockStore.EXPECT().GetGameState(ctx, gameStateID).Return(db.GameState{
			State: db.GAMESTATE_FIBBING_IT_QUESTION.String(),
		}, nil)
		mockStore.EXPECT().UpdateGameState(ctx, db.UpdateGameStateParams{
			ID:             gameStateID,
			SubmitDeadline: pgtype.Timestamp{Time: now, Valid: true},
			State:          db.GAMESTATE_FIBBING_IT_VOTING.String(),
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
				Avatar:         []byte("avatar1"),
				Question:       "My question",
				Votes:          0,
				SubmitDeadline: pgtype.Timestamp{Time: now},
			},
			{
				GameStateID:    gameStateID,
				PlayerID:       defaultOtherPlayerID,
				Nickname:       "Player 2",
				Avatar:         []byte("avatar2"),
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
					ID:       defaultHostPlayerID,
					Nickname: "Player 1",
					Avatar:   "avatar1",
					Votes:    0,
					IsReady:  false,
				},
				{
					ID:       defaultOtherPlayerID,
					Nickname: "Player 2",
					Avatar:   "avatar2",
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
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		ctx := context.Background()
		now := time.Now().Add(30 * time.Second)

		mockStore.EXPECT().GetGameState(ctx, gameStateID).Return(
			db.GameState{}, fmt.Errorf("failed to get game state"),
		)
		_, err := srv.UpdateStateToVoting(ctx, gameStateID, now)
		assert.Error(t, err)
	})

	t.Run("Should fail because state FIBBING_IT_VOTING", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		ctx := context.Background()
		now := time.Now().Add(30 * time.Second)

		mockStore.EXPECT().GetGameState(ctx, gameStateID).Return(db.GameState{
			State: db.GAMESTATE_FIBBING_IT_VOTING.String(),
		}, nil)

		_, err := srv.UpdateStateToVoting(ctx, gameStateID, now)
		assert.ErrorContains(t, err, "game state is not in FIBBING_IT_QUESTION state")
	})

	t.Run("Should fail because update game state", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		ctx := context.Background()
		now := time.Now().Add(30 * time.Second)

		mockStore.EXPECT().GetGameState(ctx, gameStateID).Return(db.GameState{
			State: db.GAMESTATE_FIBBING_IT_QUESTION.String(),
		}, nil)
		mockStore.EXPECT().UpdateGameState(ctx, db.UpdateGameStateParams{
			ID:             gameStateID,
			SubmitDeadline: pgtype.Timestamp{Time: now, Valid: true},
			State:          db.GAMESTATE_FIBBING_IT_VOTING.String(),
		}).Return(db.GameState{}, fmt.Errorf("failed to update game state"))

		_, err := srv.UpdateStateToVoting(ctx, gameStateID, now)
		assert.Error(t, err)
	})
}

func TestRoundServiceSubmitVote(t *testing.T) {
	roundID := uuid.MustParse("0193a62a-7740-7bce-849d-0e462465ca0e")

	t.Run("Should successfully submit vote", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, defaultHostPlayerID).Return(db.GameState{
			State: db.GAMESTATE_FIBBING_IT_VOTING.String(),
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

		u := uuid.Must(uuid.NewV7())
		mockRandomizer.EXPECT().GetID().Return(u)
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
				Avatar:         []byte("avatar1"),
				Votes:          int64(0),
				Question:       "My question",
				Round:          1,
				SubmitDeadline: pgtype.Timestamp{Time: deadline},
				Answer:         pgtype.Text{String: "My answer"},
			},
			{
				PlayerID:       defaultOtherPlayerID,
				Nickname:       "Player 2",
				Avatar:         []byte("avatar2"),
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
					Avatar:   "avatar1",
					Votes:    0,
					Answer:   "My answer",
				},
				{
					ID:       defaultOtherPlayerID,
					Nickname: "Player 2",
					Avatar:   "avatar2",
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
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		ctx := context.Background()
		now := time.Now().Add(30 * time.Second)

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, defaultHostPlayerID).Return(
			db.GameState{}, fmt.Errorf("failed to get game state"),
		)

		_, err := srv.SubmitVote(ctx, defaultHostPlayerID, "Player 2", now)
		assert.Error(t, err)
	})

	t.Run("Should fail because game state not in voting state", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		ctx := context.Background()
		now := time.Now().Add(30 * time.Second)

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, defaultHostPlayerID).Return(db.GameState{
			State: db.GAMESTATE_FIBBING_IT_QUESTION.String(),
		}, nil)

		_, err := srv.SubmitVote(ctx, defaultHostPlayerID, "Player 2", now)
		assert.ErrorContains(t, err, "game state is not in FIBBING_IT_VOTING state")
	})

	t.Run("Should fail because we fail to get all players in room", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		ctx := context.Background()
		now := time.Now().Add(30 * time.Second)

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, defaultHostPlayerID).Return(db.GameState{
			State: db.GAMESTATE_FIBBING_IT_VOTING.String(),
		}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, defaultHostPlayerID).Return(
			[]db.GetAllPlayersInRoomRow{}, fmt.Errorf("failed to get all players in room"),
		)

		_, err := srv.SubmitVote(ctx, defaultHostPlayerID, "Player 2", now)
		assert.Error(t, err)
	})

	t.Run("Should fail because we voted for themselves", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		ctx := context.Background()
		now := time.Now().Add(30 * time.Second)

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, defaultHostPlayerID).Return(db.GameState{
			State: db.GAMESTATE_FIBBING_IT_VOTING.String(),
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
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		ctx := context.Background()
		now := time.Now().Add(30 * time.Second)

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, defaultHostPlayerID).Return(db.GameState{
			State: db.GAMESTATE_FIBBING_IT_VOTING.String(),
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
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		ctx := context.Background()
		now := time.Now().Add(30 * time.Second)

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, defaultHostPlayerID).Return(db.GameState{
			State: db.GAMESTATE_FIBBING_IT_VOTING.String(),
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
			db.GetLatestRoundByPlayerIDRow{}, fmt.Errorf("failed to get latest round"),
		)

		_, err := srv.SubmitVote(ctx, defaultHostPlayerID, "Player 2", now)
		assert.Error(t, err)
	})

	t.Run("Should fail because we are passed the submit deadline", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		ctx := context.Background()
		now := time.Now().Add(30 * time.Second)

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, defaultHostPlayerID).Return(db.GameState{
			State: db.GAMESTATE_FIBBING_IT_VOTING.String(),
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
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		ctx := context.Background()
		now := time.Now().Add(30 * time.Second)

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, defaultHostPlayerID).Return(db.GameState{
			State: db.GAMESTATE_FIBBING_IT_VOTING.String(),
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

		u := uuid.Must(uuid.NewV7())
		mockRandomizer.EXPECT().GetID().Return(u)
		mockStore.EXPECT().UpsertFibbingItVote(ctx, db.UpsertFibbingItVoteParams{
			ID:               u,
			RoundID:          roundID,
			PlayerID:         defaultHostPlayerID,
			VotedForPlayerID: defaultOtherPlayerID,
		}).Return(fmt.Errorf("failed to upsert fibbing it vote"))

		_, err := srv.SubmitVote(ctx, defaultHostPlayerID, "Player 2", now)
		assert.Error(t, err)
	})

	t.Run("Should fail because we fail to get vote count", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		ctx := context.Background()
		now := time.Now().Add(30 * time.Second)

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, defaultHostPlayerID).Return(db.GameState{
			State: db.GAMESTATE_FIBBING_IT_VOTING.String(),
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
		u := uuid.Must(uuid.NewV7())
		mockRandomizer.EXPECT().GetID().Return(u)
		mockStore.EXPECT().UpsertFibbingItVote(ctx, db.UpsertFibbingItVoteParams{
			ID:               u,
			RoundID:          roundID,
			PlayerID:         defaultHostPlayerID,
			VotedForPlayerID: defaultOtherPlayerID,
		}).Return(nil)
		mockStore.EXPECT().GetVotingState(ctx, roundID).Return(
			[]db.GetVotingStateRow{}, fmt.Errorf("failed to get vote count"),
		)

		_, err := srv.SubmitVote(ctx, defaultHostPlayerID, "Player 2", now)
		assert.Error(t, err)
	})
}

func TestPlayerServiceGetVotingState(t *testing.T) {
	roundID := uuid.MustParse("0193a629-e26c-7326-8df4-81ad3ec82214")
	gameStateID := uuid.MustParse("fbb75599-9f7a-4392-b523-fd433b3208ea")

	t.Run("Should successfully get voting state", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		ctx := context.Background()

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
					Votes:       1,
					Avatar:      []byte(""),
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
					Avatar:   "",
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
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetLatestRoundByPlayerID(ctx, playerID).Return(
			db.GetLatestRoundByPlayerIDRow{}, fmt.Errorf("failed to get round info"),
		)
		_, err := srv.GetVotingState(ctx, playerID)
		assert.Error(t, err)
	})

	t.Run("Should fail to get voting state because fail to get votes from DB", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetLatestRoundByPlayerID(ctx, playerID).Return(
			db.GetLatestRoundByPlayerIDRow{
				ID: roundID,
			}, nil)
		mockStore.EXPECT().GetVotingState(ctx, roundID).Return(
			[]db.GetVotingStateRow{}, fmt.Errorf("failed to get votes"),
		)

		_, err := srv.GetVotingState(ctx, playerID)

		assert.Error(t, err)
	})
}

func TestRoundServiceToggleVotingIsReady(t *testing.T) {
	t.Run("Should successfully toggle voting ready state", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, playerID).Return(db.GameState{
			State:          db.GAMESTATE_FIBBING_IT_VOTING.String(),
			SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(1 * time.Hour)},
		}, nil)
		mockStore.EXPECT().ToggleVotingIsReady(ctx, playerID).Return(db.FibbingItVote{}, nil)
		mockStore.EXPECT().GetAllPlayersVotingIsReady(ctx, playerID).Return(false, nil)

		allReady, err := srv.ToggleVotingIsReady(ctx, playerID, time.Now().UTC())
		assert.NoError(t, err)
		assert.False(t, allReady)
	})

	t.Run("Should successfully toggle answer ready state and return all players are ready", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, playerID).Return(db.GameState{
			State:          db.GAMESTATE_FIBBING_IT_VOTING.String(),
			SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(1 * time.Hour)},
		}, nil)
		mockStore.EXPECT().ToggleVotingIsReady(ctx, playerID).Return(db.FibbingItVote{}, nil)
		mockStore.EXPECT().GetAllPlayersVotingIsReady(ctx, playerID).Return(true, nil)

		allReady, err := srv.ToggleVotingIsReady(ctx, playerID, time.Now().UTC())
		assert.NoError(t, err)
		assert.True(t, allReady)
	})

	t.Run("Should fail to toggle voting ready state, because we fail to get game state", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, playerID).Return(
			db.GameState{}, fmt.Errorf("failed to get state"),
		)

		_, err := srv.ToggleAnswerIsReady(ctx, playerID, time.Now().UTC())
		assert.Error(t, err)
	})

	t.Run(
		"Should fail to voting toggle ready state because after deadline",
		func(t *testing.T) {
			mockStore := mockService.NewMockStorer(t)
			mockRandomizer := mockService.NewMockRandomizer(t)
			srv := service.NewRoundService(mockStore, mockRandomizer)

			ctx := context.Background()

			mockStore.EXPECT().GetGameStateByPlayerID(ctx, playerID).Return(db.GameState{
				State:          db.GAMESTATE_FIBBING_IT_VOTING.String(),
				SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(-1 * time.Second)},
			}, nil)

			_, err := srv.ToggleVotingIsReady(ctx, playerID, time.Now().UTC())
			assert.Error(t, err)
		},
	)

	t.Run("Should fail to toggle voting ready state because game state not in voting state", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, playerID).Return(db.GameState{
			State:          db.GAMESTATE_FIBBING_IT_QUESTION.String(),
			SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(1 * time.Hour)},
		}, nil)

		_, err := srv.ToggleVotingIsReady(ctx, playerID, time.Now().UTC())
		assert.ErrorContains(t, err, "room game state is not in FIBBING_IT_VOTING state")
	})

	t.Run(
		"Should fail to toggle voting ready state, because we fail to toggle answer ready state in DB",
		func(t *testing.T) {
			mockStore := mockService.NewMockStorer(t)
			mockRandomizer := mockService.NewMockRandomizer(t)
			srv := service.NewRoundService(mockStore, mockRandomizer)

			ctx := context.Background()

			mockStore.EXPECT().GetGameStateByPlayerID(ctx, playerID).Return(db.GameState{
				State:          db.GAMESTATE_FIBBING_IT_VOTING.String(),
				SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(1 * time.Hour)},
			}, nil)
			mockStore.EXPECT().ToggleVotingIsReady(ctx, playerID).Return(
				db.FibbingItVote{}, fmt.Errorf("failed to toggle voting is ready"),
			)

			_, err := srv.ToggleVotingIsReady(ctx, playerID, time.Now().UTC())
			assert.Error(t, err)
		},
	)

	t.Run(
		"Should fail to toggle voting ready state because we fail to get all players ready status from DB",
		func(t *testing.T) {
			mockStore := mockService.NewMockStorer(t)
			mockRandomizer := mockService.NewMockRandomizer(t)
			srv := service.NewRoundService(mockStore, mockRandomizer)

			ctx := context.Background()

			mockStore.EXPECT().GetGameStateByPlayerID(ctx, playerID).Return(db.GameState{
				State:          db.GAMESTATE_FIBBING_IT_VOTING.String(),
				SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(1 * time.Hour)},
			}, nil)
			mockStore.EXPECT().ToggleVotingIsReady(ctx, playerID).Return(db.FibbingItVote{}, nil)
			mockStore.EXPECT().GetAllPlayersVotingIsReady(ctx, playerID).Return(
				false, fmt.Errorf("failed to get player voting is ready status"),
			)

			_, err := srv.ToggleVotingIsReady(ctx, playerID, time.Now().UTC())
			assert.Error(t, err)
		},
	)
}

func TestRoundServiceUpdateStateToReveal(t *testing.T) {
	gameStateID := uuid.MustParse("fbb75599-9f7a-4392-b523-fd433b3208ea")
	roundID := uuid.MustParse("0193a62a-364e-751a-9088-cf3b9711153e")

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
			expectedPlayerAvatar:   "avatar2",
			expectedPlayerRole:     "normal",
			expectedShouldReveal:   true,
		},
		{
			name:                   "Should show reveal state revealing fibber player",
			votesPlayerOne:         1,
			votesPlayerTwo:         0,
			expectedPlayerNickname: "Player 1",
			expectedPlayerAvatar:   "avatar1",
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
			mockStore := mockService.NewMockStorer(t)
			mockRandomizer := mockService.NewMockRandomizer(t)
			srv := service.NewRoundService(mockStore, mockRandomizer)

			ctx := context.Background()
			now := time.Now().Add(15 * time.Second)

			mockStore.EXPECT().GetGameState(ctx, gameStateID).Return(db.GameState{
				State: db.GAMESTATE_FIBBING_IT_VOTING.String(),
			}, nil)
			mockStore.EXPECT().UpdateGameState(ctx, db.UpdateGameStateParams{
				ID:             gameStateID,
				SubmitDeadline: pgtype.Timestamp{Time: now, Valid: true},
				State:          db.GAMESTATE_FIBBING_IT_REVEAL_ROLE.String(),
			}).Return(db.GameState{}, nil)
			mockStore.EXPECT().GetLatestRoundByGameStateID(ctx, gameStateID).Return(db.GetLatestRoundByGameStateIDRow{
				ID:    roundID,
				Round: 1,
			}, nil)
			mockStore.EXPECT().GetVotingState(ctx, roundID).Return([]db.GetVotingStateRow{
				{
					PlayerID:       defaultHostPlayerID,
					Nickname:       "Player 1",
					Avatar:         []byte("avatar1"),
					Question:       "My question",
					Votes:          tt.votesPlayerOne,
					SubmitDeadline: pgtype.Timestamp{Time: now},
					Role:           pgtype.Text{String: "fibber"},
				},
				{
					PlayerID:       defaultOtherPlayerID,
					Nickname:       "Player 2",
					Avatar:         []byte("avatar2"),
					Question:       "My question",
					Votes:          tt.votesPlayerTwo,
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
					defaultHostPlayerID,
					defaultOtherPlayerID,
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
			mockStore := mockService.NewMockStorer(t)
			mockRandomizer := mockService.NewMockRandomizer(t)
			srv := service.NewRoundService(mockStore, mockRandomizer)

			ctx := context.Background()
			now := time.Now().Add(15 * time.Second)

			mockStore.EXPECT().
				GetGameState(ctx, gameStateID).
				Return(db.GameState{}, fmt.Errorf("failed to get game state"))
			_, err := srv.UpdateStateToReveal(ctx, gameStateID, now)
			assert.Error(t, err)
		},
	)

	t.Run(
		"Should fail to update state to reveal because we wrong game state",
		func(t *testing.T) {
			mockStore := mockService.NewMockStorer(t)
			mockRandomizer := mockService.NewMockRandomizer(t)
			srv := service.NewRoundService(mockStore, mockRandomizer)

			ctx := context.Background()
			now := time.Now().Add(15 * time.Second)

			mockStore.EXPECT().GetGameState(ctx, gameStateID).Return(db.GameState{
				State: db.GAMESTATE_FIBBING_IT_QUESTION.String(),
			}, nil)

			_, err := srv.UpdateStateToReveal(ctx, gameStateID, now)
			assert.ErrorContains(t, err, "game state is not in FIBBING_IT_VOTING state")
		},
	)

	t.Run(
		"Should fail to update state to reveal because we fail to update game state",
		func(t *testing.T) {
			mockStore := mockService.NewMockStorer(t)
			mockRandomizer := mockService.NewMockRandomizer(t)
			srv := service.NewRoundService(mockStore, mockRandomizer)

			ctx := context.Background()
			now := time.Now().Add(15 * time.Second)

			mockStore.EXPECT().GetGameState(ctx, gameStateID).Return(db.GameState{
				State: db.GAMESTATE_FIBBING_IT_VOTING.String(),
			}, nil)
			mockStore.EXPECT().UpdateGameState(ctx, db.UpdateGameStateParams{
				ID:             gameStateID,
				SubmitDeadline: pgtype.Timestamp{Time: now, Valid: true},
				State:          db.GAMESTATE_FIBBING_IT_REVEAL_ROLE.String(),
			}).Return(db.GameState{}, fmt.Errorf("failed to update game state"))

			_, err := srv.UpdateStateToReveal(ctx, gameStateID, now)
			assert.Error(t, err)
		},
	)

	t.Run(
		"Should fail to update state to reveal because we get latest round by game state ID",
		func(t *testing.T) {
			mockStore := mockService.NewMockStorer(t)
			mockRandomizer := mockService.NewMockRandomizer(t)
			srv := service.NewRoundService(mockStore, mockRandomizer)

			ctx := context.Background()
			now := time.Now().Add(15 * time.Second)

			mockStore.EXPECT().GetGameState(ctx, gameStateID).Return(db.GameState{
				State: db.GAMESTATE_FIBBING_IT_VOTING.String(),
			}, nil)
			mockStore.EXPECT().UpdateGameState(ctx, db.UpdateGameStateParams{
				ID:             gameStateID,
				SubmitDeadline: pgtype.Timestamp{Time: now, Valid: true},
				State:          db.GAMESTATE_FIBBING_IT_REVEAL_ROLE.String(),
			}).Return(db.GameState{}, nil)
			mockStore.EXPECT().GetLatestRoundByGameStateID(ctx, gameStateID).Return(
				db.GetLatestRoundByGameStateIDRow{}, fmt.Errorf("failed to get latest round by game state ID"),
			)

			_, err := srv.UpdateStateToReveal(ctx, gameStateID, now)
			assert.Error(t, err)
		},
	)

	t.Run(
		"Should fail to update state to reveal because we fail to get voting state",
		func(t *testing.T) {
			mockStore := mockService.NewMockStorer(t)
			mockRandomizer := mockService.NewMockRandomizer(t)
			srv := service.NewRoundService(mockStore, mockRandomizer)

			ctx := context.Background()
			now := time.Now().Add(15 * time.Second)

			mockStore.EXPECT().GetGameState(ctx, gameStateID).Return(db.GameState{
				State: db.GAMESTATE_FIBBING_IT_VOTING.String(),
			}, nil)
			mockStore.EXPECT().UpdateGameState(ctx, db.UpdateGameStateParams{
				ID:             gameStateID,
				SubmitDeadline: pgtype.Timestamp{Time: now, Valid: true},
				State:          db.GAMESTATE_FIBBING_IT_REVEAL_ROLE.String(),
			}).Return(db.GameState{}, nil)
			mockStore.EXPECT().GetLatestRoundByGameStateID(ctx, gameStateID).Return(db.GetLatestRoundByGameStateIDRow{
				ID:    roundID,
				Round: 1,
			}, nil)
			mockStore.EXPECT().GetVotingState(ctx, roundID).Return(
				[]db.GetVotingStateRow{}, fmt.Errorf("failed to get voting state"),
			)

			_, err := srv.UpdateStateToReveal(ctx, gameStateID, now)
			assert.Error(t, err)
		},
	)
}

func TestRoundServiceUpdateStateToQuestion(t *testing.T) {
	gameStateID := uuid.MustParse("fbb75599-9f7a-4392-b523-fd433b3208ea")
	groupID := uuid.MustParse("0193a629-1fcf-79dd-ac70-760bedbdffa9")

	tests := []struct {
		name                    string
		roundNumber             int32
		roundType               string
		expectedRound           int32
		expectedType            string
		expectedAnswerPlayerOne []string
		expectedAnswerPlayerTwo []string
		normalQuestion          string
		fibberQuestion          string
	}{
		{
			name:                    "Should update state to question state successfully with round 2 and free_form",
			roundNumber:             1,
			roundType:               "free_form",
			expectedRound:           2,
			expectedType:            "free_form",
			expectedAnswerPlayerOne: []string{},
			expectedAnswerPlayerTwo: []string{},
			normalQuestion:          "What if your favourite city",
			fibberQuestion:          "What is your favourite hotel",
		},
		{
			name:                    "Should update state to question state successfully with round 1 and multiple_choice",
			roundNumber:             3,
			roundType:               "free_form",
			expectedRound:           1,
			expectedType:            "multiple_choice",
			expectedAnswerPlayerOne: []string{"Strongly Agree", "Agree", "Neutral", "Disagree", "Strongly Disagree"},
			expectedAnswerPlayerTwo: []string{"Strongly Agree", "Agree", "Neutral", "Disagree", "Strongly Disagree"},
			normalQuestion:          "I love pizza",
			fibberQuestion:          "I love burgers",
		},
		{
			name:                    "Should update state to question state successfully with round 1 and most_likely",
			roundNumber:             3,
			roundType:               "multiple_choice",
			expectedRound:           1,
			expectedType:            "most_likely",
			expectedAnswerPlayerOne: []string{"Player 2"},
			expectedAnswerPlayerTwo: []string{"Player 1"},
			normalQuestion:          "go to prison",
			fibberQuestion:          "go to a bank",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := mockService.NewMockStorer(t)
			mockRandom := mockService.NewMockRandomizer(t)
			srv := service.NewRoundService(mockStore, mockRandom)

			ctx := context.Background()
			gameName := gameName
			deadline := time.Now().Add(5 * time.Second).UTC()

			mockStore.EXPECT().GetGameState(ctx, gameStateID).Return(db.GameState{
				State: db.GAMESTATE_FIBBING_IT_REVEAL_ROLE.String(),
			}, nil)
			mockStore.EXPECT().UpdateGameState(ctx, db.UpdateGameStateParams{
				ID:             gameStateID,
				SubmitDeadline: pgtype.Timestamp{Time: deadline, Valid: true},
				State:          db.GAMESTATE_FIBBING_IT_QUESTION.String(),
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
				RoundType: tt.roundType,
				Round:     tt.roundNumber,
			}, nil)
			mockStore.EXPECT().GetRandomQuestionByRound(ctx, db.GetRandomQuestionByRoundParams{
				GameName:     gameName,
				LanguageCode: "en-GB",
				Round:        tt.expectedType,
			}).Return(db.Question{
				ID:       uuid.MustParse("0193a629-7dcc-78ad-822f-fd5d83c89ae7"),
				Question: tt.normalQuestion,
				GroupID:  groupID,
			}, nil)
			mockStore.EXPECT().GetRandomQuestionInGroup(ctx, db.GetRandomQuestionInGroupParams{
				GroupID: groupID,
				ID:      uuid.MustParse("0193a629-7dcc-78ad-822f-fd5d83c89ae7"),
			}).Return(db.GetRandomQuestionInGroupRow{
				ID:       uuid.MustParse("0193a629-a9ac-7fc4-828c-a1334c282e0f"),
				Question: tt.fibberQuestion,
			}, nil)

			mockRandom.EXPECT().GetFibberIndex(2).Return(1)
			mockStore.EXPECT().NewRound(ctx, db.NewRoundArgs{
				GameStateID:       gameStateID,
				NormalsQuestionID: uuid.MustParse("0193a629-7dcc-78ad-822f-fd5d83c89ae7"),
				FibberQuestionID:  uuid.MustParse("0193a629-a9ac-7fc4-828c-a1334c282e0f"),
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

			gameState, err := srv.UpdateStateToQuestion(ctx, gameStateID, deadline)
			expectedGameState := service.QuestionState{
				Deadline:    time.Until(deadline),
				GameStateID: gameStateID,
				Players: []service.PlayerWithRole{
					{
						ID:              defaultHostPlayerID,
						Role:            "normal",
						Question:        tt.normalQuestion,
						PossibleAnswers: tt.expectedAnswerPlayerOne,
					},
					{
						ID:              defaultOtherPlayerID,
						Role:            "fibber",
						Question:        tt.fibberQuestion,
						PossibleAnswers: tt.expectedAnswerPlayerTwo,
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
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom)

		ctx := context.Background()
		deadline := time.Now().Add(5 * time.Second).UTC()

		mockStore.EXPECT().GetGameState(ctx, gameStateID).Return(db.GameState{}, fmt.Errorf("failed to get game state"))
		_, err := srv.UpdateStateToQuestion(ctx, gameStateID, deadline)
		assert.Error(t, err)
	})

	t.Run("Should fail to update state to question because we are in incorrect game state", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom)

		ctx := context.Background()
		deadline := time.Now().Add(5 * time.Second).UTC()

		mockStore.EXPECT().GetGameState(ctx, gameStateID).Return(db.GameState{
			State: db.GAMESTATE_FIBBING_IT_QUESTION.String(),
		}, nil)

		_, err := srv.UpdateStateToQuestion(ctx, gameStateID, deadline)
		assert.ErrorContains(t, err, "game state is not in FIBBING_IT_REVEAL_ROLE state or FIBBING_IT_SCORING state")
	})

	t.Run(
		"Should fail to update state to question because we fail to get all players via game state",
		func(t *testing.T) {
			mockStore := mockService.NewMockStorer(t)
			mockRandom := mockService.NewMockRandomizer(t)
			srv := service.NewRoundService(mockStore, mockRandom)

			ctx := context.Background()
			deadline := time.Now().Add(5 * time.Second).UTC()

			mockStore.EXPECT().GetGameState(ctx, gameStateID).Return(db.GameState{
				State: db.GAMESTATE_FIBBING_IT_REVEAL_ROLE.String(),
			}, nil)
			mockStore.EXPECT().UpdateGameState(ctx, db.UpdateGameStateParams{
				ID:             gameStateID,
				SubmitDeadline: pgtype.Timestamp{Time: deadline, Valid: true},
				State:          db.GAMESTATE_FIBBING_IT_QUESTION.String(),
			}).Return(db.GameState{}, nil)
			mockStore.EXPECT().GetAllPlayersByGameStateID(ctx, gameStateID).Return(
				nil,
				fmt.Errorf("failed to get all player IDs by game state ID"),
			)
			_, err := srv.UpdateStateToQuestion(ctx, gameStateID, deadline)
			assert.Error(t, err)
		},
	)

	t.Run("Should fail to update state to question because we fail to get latest round", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom)

		ctx := context.Background()
		deadline := time.Now().Add(5 * time.Second).UTC()

		mockStore.EXPECT().GetGameState(ctx, gameStateID).Return(db.GameState{
			State: db.GAMESTATE_FIBBING_IT_REVEAL_ROLE.String(),
		}, nil)
		mockStore.EXPECT().UpdateGameState(ctx, db.UpdateGameStateParams{
			ID:             gameStateID,
			SubmitDeadline: pgtype.Timestamp{Time: deadline, Valid: true},
			State:          db.GAMESTATE_FIBBING_IT_QUESTION.String(),
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
			Return(db.GetLatestRoundByGameStateIDRow{}, fmt.Errorf("failed to get latest round by game state ID"))

		_, err := srv.UpdateStateToQuestion(ctx, gameStateID, deadline)
		assert.Error(t, err)
	})

	t.Run(
		"Should fail to update state to question because we fail to get random question by round",
		func(t *testing.T) {
			mockStore := mockService.NewMockStorer(t)
			mockRandom := mockService.NewMockRandomizer(t)
			srv := service.NewRoundService(mockStore, mockRandom)

			ctx := context.Background()
			gameName := gameName
			deadline := time.Now().Add(5 * time.Second).UTC()

			mockStore.EXPECT().GetGameState(ctx, gameStateID).Return(db.GameState{
				State: db.GAMESTATE_FIBBING_IT_REVEAL_ROLE.String(),
			}, nil)
			mockStore.EXPECT().UpdateGameState(ctx, db.UpdateGameStateParams{
				ID:             gameStateID,
				SubmitDeadline: pgtype.Timestamp{Time: deadline, Valid: true},
				State:          db.GAMESTATE_FIBBING_IT_QUESTION.String(),
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
				RoundType: "free_form",
				Round:     1,
			}, nil)
			mockStore.EXPECT().GetRandomQuestionByRound(ctx, db.GetRandomQuestionByRoundParams{
				GameName:     gameName,
				LanguageCode: "en-GB",
				Round:        "free_form",
			}).Return(db.Question{}, fmt.Errorf("failed to get random question by round"))

			_, err := srv.UpdateStateToQuestion(ctx, gameStateID, deadline)
			assert.Error(t, err)
		},
	)

	t.Run("Should fail to update to question because we fail to get random question in group", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom)

		ctx := context.Background()
		gameName := gameName
		deadline := time.Now().Add(5 * time.Second).UTC()

		mockStore.EXPECT().GetGameState(ctx, gameStateID).Return(db.GameState{
			State: db.GAMESTATE_FIBBING_IT_REVEAL_ROLE.String(),
		}, nil)
		mockStore.EXPECT().UpdateGameState(ctx, db.UpdateGameStateParams{
			ID:             gameStateID,
			SubmitDeadline: pgtype.Timestamp{Time: deadline, Valid: true},
			State:          db.GAMESTATE_FIBBING_IT_QUESTION.String(),
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
			RoundType: "free_form",
			Round:     1,
		}, nil)
		mockStore.EXPECT().GetRandomQuestionByRound(ctx, db.GetRandomQuestionByRoundParams{
			GameName:     gameName,
			LanguageCode: "en-GB",
			Round:        "free_form",
		}).Return(db.Question{
			ID:       uuid.MustParse("0193a629-7dcc-78ad-822f-fd5d83c89ae7"),
			Question: "I love cats",
			GroupID:  groupID,
		}, nil)
		mockStore.EXPECT().GetRandomQuestionInGroup(ctx, db.GetRandomQuestionInGroupParams{
			GroupID: groupID,
			ID:      uuid.MustParse("0193a629-7dcc-78ad-822f-fd5d83c89ae7"),
		}).Return(db.GetRandomQuestionInGroupRow{}, fmt.Errorf("failed to get random question in group"))

		_, err := srv.UpdateStateToQuestion(ctx, gameStateID, deadline)
		assert.Error(t, err)
	})

	t.Run("Should fail to update to question because we fail to add a new round", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandom)

		ctx := context.Background()
		deadline := time.Now().Add(5 * time.Second).UTC()

		mockStore.EXPECT().GetGameState(ctx, gameStateID).Return(db.GameState{
			State: db.GAMESTATE_FIBBING_IT_REVEAL_ROLE.String(),
		}, nil)
		mockStore.EXPECT().UpdateGameState(ctx, db.UpdateGameStateParams{
			ID:             gameStateID,
			SubmitDeadline: pgtype.Timestamp{Time: deadline, Valid: true},
			State:          db.GAMESTATE_FIBBING_IT_QUESTION.String(),
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
			RoundType: "free_form",
			Round:     1,
		}, nil)
		mockStore.EXPECT().GetRandomQuestionByRound(ctx, db.GetRandomQuestionByRoundParams{
			GameName:     gameName,
			LanguageCode: "en-GB",
			Round:        "free_form",
		}).Return(db.Question{
			ID:       uuid.MustParse("0193a629-7dcc-78ad-822f-fd5d83c89ae7"),
			Question: "I love cats",
			GroupID:  groupID,
		}, nil)
		mockStore.EXPECT().GetRandomQuestionInGroup(ctx, db.GetRandomQuestionInGroupParams{
			GroupID: groupID,
			ID:      uuid.MustParse("0193a629-7dcc-78ad-822f-fd5d83c89ae7"),
		}).Return(db.GetRandomQuestionInGroupRow{
			ID:       uuid.MustParse("0193a629-a9ac-7fc4-828c-a1334c282e0f"),
			Question: "I love dogs",
		}, nil)
		mockRandom.EXPECT().GetFibberIndex(2).Return(1)
		mockStore.EXPECT().NewRound(ctx, db.NewRoundArgs{
			GameStateID:       gameStateID,
			NormalsQuestionID: uuid.MustParse("0193a629-7dcc-78ad-822f-fd5d83c89ae7"),
			FibberQuestionID:  uuid.MustParse("0193a629-a9ac-7fc4-828c-a1334c282e0f"),
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
		}).Return(fmt.Errorf("failed to add a new round"))

		_, err := srv.UpdateStateToQuestion(ctx, gameStateID, deadline)
		assert.Error(t, err)
	})
}
