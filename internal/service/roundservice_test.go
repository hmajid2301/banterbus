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
	sqlc "gitlab.com/hmajid2301/banterbus/internal/store/db"
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

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(sqlc.Room{
			ID:        roomID,
			RoomState: sqlc.ROOMSTATE_PLAYING.String(),
		}, nil)
		mockStore.EXPECT().GetLatestRoundByPlayerID(ctx, playerID).Return(
			sqlc.GetLatestRoundByPlayerIDRow{
				ID:             roundID,
				SubmitDeadline: pgtype.Timestamp{Time: now.Add(1 * time.Hour)},
			}, nil)

		u := uuid.Must(uuid.NewV7())
		mockRandomizer.EXPECT().GetID().Return(u)
		mockStore.EXPECT().AddFibbingItAnswer(ctx, sqlc.AddFibbingItAnswerParams{
			ID:       u,
			RoundID:  roundID,
			PlayerID: playerID,
			Answer:   "My answer",
		}).Return(sqlc.FibbingItAnswer{}, nil)

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
			Return(sqlc.Room{}, fmt.Errorf("failed to get room details"))
		err := srv.SubmitAnswer(ctx, playerID, "My answer", now)
		assert.Error(t, err)
	})

	t.Run("Should fail to submit answer because room not in PLAYING state", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		now := time.Now()
		ctx := context.Background()

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(sqlc.Room{
			ID:        roomID,
			RoomState: sqlc.ROOMSTATE_CREATED.String(),
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

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(sqlc.Room{
			ID:        roomID,
			RoomState: sqlc.ROOMSTATE_PLAYING.String(),
		}, nil)
		mockStore.EXPECT().GetLatestRoundByPlayerID(ctx, playerID).Return(
			sqlc.GetLatestRoundByPlayerIDRow{},
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

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(sqlc.Room{
			ID:        roomID,
			RoomState: sqlc.ROOMSTATE_PLAYING.String(),
		}, nil)
		mockStore.EXPECT().GetLatestRoundByPlayerID(ctx, playerID).Return(
			sqlc.GetLatestRoundByPlayerIDRow{
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

		mockStore.EXPECT().GetRoomByPlayerID(ctx, playerID).Return(sqlc.Room{
			ID:        roomID,
			RoomState: sqlc.ROOMSTATE_PLAYING.String(),
		}, nil)
		mockStore.EXPECT().GetLatestRoundByPlayerID(ctx, playerID).Return(
			sqlc.GetLatestRoundByPlayerIDRow{
				ID:             roundID,
				SubmitDeadline: pgtype.Timestamp{Time: now.Add(1 * time.Hour)},
			}, nil)

		mockRandomizer.EXPECT().GetID().Return(defaultHostPlayerID)
		mockStore.EXPECT().AddFibbingItAnswer(ctx, sqlc.AddFibbingItAnswerParams{
			ID:       defaultHostPlayerID,
			RoundID:  roundID,
			PlayerID: playerID,
			Answer:   "My answer",
		}).Return(sqlc.FibbingItAnswer{}, fmt.Errorf("failed to add answer to DB"))

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

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, playerID).Return(sqlc.GameState{
			State:          sqlc.GAMESTATE_FIBBING_IT_SHOW_QUESTION.String(),
			SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(1 * time.Hour)},
		}, nil)
		mockStore.EXPECT().ToggleAnswerIsReady(ctx, playerID).Return(sqlc.FibbingItAnswer{}, nil)
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

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, playerID).Return(sqlc.GameState{
			State:          sqlc.GAMESTATE_FIBBING_IT_SHOW_QUESTION.String(),
			SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(1 * time.Hour)},
		}, nil)
		mockStore.EXPECT().ToggleAnswerIsReady(ctx, playerID).Return(sqlc.FibbingItAnswer{}, nil)
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
			sqlc.GameState{}, fmt.Errorf("failed to get state"),
		)

		_, err := srv.ToggleAnswerIsReady(ctx, playerID, time.Now().UTC())
		assert.Error(t, err)
	})

	t.Run("Should fail to toggle answer ready state because game state not in show question", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		ctx := context.Background()

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, playerID).Return(sqlc.GameState{
			State:          sqlc.GAMESTATE_FIBBING_IT_VOTING.String(),
			SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(1 * time.Hour)},
		}, nil)

		_, err := srv.ToggleAnswerIsReady(ctx, playerID, time.Now().UTC())
		assert.ErrorContains(t, err, "room game state is not in FIBBING_IT_SHOW_QUESTION state")
	})

	t.Run(
		"Should fail to toggle answer ready state, because we fail to toggle answer ready state in DB",
		func(t *testing.T) {
			mockStore := mockService.NewMockStorer(t)
			mockRandomizer := mockService.NewMockRandomizer(t)
			srv := service.NewRoundService(mockStore, mockRandomizer)

			ctx := context.Background()

			mockStore.EXPECT().GetGameStateByPlayerID(ctx, playerID).Return(sqlc.GameState{
				State:          sqlc.GAMESTATE_FIBBING_IT_SHOW_QUESTION.String(),
				SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(1 * time.Hour)},
			}, nil)
			mockStore.EXPECT().ToggleAnswerIsReady(ctx, playerID).Return(
				sqlc.FibbingItAnswer{}, fmt.Errorf("failed to toggle answer is ready"),
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

			mockStore.EXPECT().GetGameStateByPlayerID(ctx, playerID).Return(sqlc.GameState{
				State:          sqlc.GAMESTATE_FIBBING_IT_SHOW_QUESTION.String(),
				SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(1 * time.Hour)},
			}, nil)
			mockStore.EXPECT().ToggleAnswerIsReady(ctx, playerID).Return(sqlc.FibbingItAnswer{}, nil)
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

			mockStore.EXPECT().GetGameStateByPlayerID(ctx, playerID).Return(sqlc.GameState{
				State:          sqlc.GAMESTATE_FIBBING_IT_SHOW_QUESTION.String(),
				SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(-1 * time.Second)},
			}, nil)

			_, err := srv.ToggleAnswerIsReady(ctx, playerID, time.Now().UTC())
			assert.Error(t, err)
		},
	)
}

func TestRoundServiceUpdateStateToVoting(t *testing.T) {
	gameStateID := uuid.MustParse("fbb75599-9f7a-4392-b523-fd433b3208ea")

	t.Run("Should successfully update state to voting", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		players := []service.PlayerWithRole{
			{
				ID:       defaultHostPlayerID,
				Nickname: "Player 1",
				Avatar:   []byte("avatar1"),
				Role:     "fibber",
				Question: "My other question",
			},
			{
				ID:       defaultOtherPlayerID,
				Nickname: "Player 2",
				Avatar:   []byte("avatar2"),
				Role:     "normal",
				Question: "My question",
			},
		}

		ctx := context.Background()
		now := time.Now().Add(30 * time.Second)

		mockStore.EXPECT().UpdateGameState(ctx, sqlc.UpdateGameStateParams{
			ID:             gameStateID,
			SubmitDeadline: pgtype.Timestamp{Time: now},
			State:          sqlc.GAMESTATE_FIBBING_IT_VOTING.String(),
		}).Return(sqlc.GameState{}, nil)

		updateState := service.UpdateVotingState{
			Players:     players,
			GameStateID: gameStateID,
			Deadline:    now,
			Round:       1,
		}

		votes, err := srv.UpdateStateToVoting(ctx, updateState)
		assert.NoError(t, err)
		expectedVotes := service.VotingState{
			Question: "My question",
			Round:    1,
			Players: []service.PlayerWithVoting{
				{
					ID:       defaultHostPlayerID,
					Nickname: "Player 1",
					Avatar:   "avatar1",
					Votes:    0,
				},
				{
					ID:       defaultOtherPlayerID,
					Nickname: "Player 2",
					Avatar:   "avatar2",
					Votes:    0,
				},
			},
		}
		assert.Equal(t, expectedVotes, votes)
	})

	t.Run("Should fail to move state because fail to update DB", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		players := []service.PlayerWithRole{
			{
				ID:       defaultHostPlayerID,
				Nickname: "Player 1",
				Avatar:   []byte("avatar1"),
			},
			{
				ID:       defaultOtherPlayerID,
				Nickname: "Player 2",
				Avatar:   []byte("avatar2"),
			},
		}

		ctx := context.Background()
		now := time.Now().Add(30 * time.Second)

		mockStore.EXPECT().UpdateGameState(ctx, sqlc.UpdateGameStateParams{
			ID:             gameStateID,
			SubmitDeadline: pgtype.Timestamp{Time: now},
			State:          sqlc.GAMESTATE_FIBBING_IT_VOTING.String(),
		}).Return(
			sqlc.GameState{}, fmt.Errorf("failed to update game state"),
		)

		updateState := service.UpdateVotingState{
			Players:     players,
			GameStateID: gameStateID,
			Deadline:    now,
			Round:       1,
		}

		_, err := srv.UpdateStateToVoting(ctx, updateState)
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

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, defaultHostPlayerID).Return(sqlc.GameState{
			State: sqlc.GAMESTATE_FIBBING_IT_VOTING.String(),
		}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, defaultHostPlayerID).Return([]sqlc.GetAllPlayersInRoomRow{
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
		mockStore.EXPECT().GetLatestRoundByPlayerID(ctx, defaultHostPlayerID).Return(sqlc.GetLatestRoundByPlayerIDRow{
			ID:             roundID,
			SubmitDeadline: pgtype.Timestamp{Time: deadline},
		}, nil)
		mockStore.EXPECT().GetVotingState(ctx, roundID).Return([]sqlc.GetVotingStateRow{
			{
				VotedForPlayerID: defaultHostPlayerID,
				Nickname:         "Player 1",
				Avatar:           []byte("avatar1"),
				VoteCount:        0,
				Question:         "My question",
				Round:            1,
				SubmitDeadline:   pgtype.Timestamp{Time: deadline},
			},
			{
				VotedForPlayerID: defaultOtherPlayerID,
				Nickname:         "Player 2",
				Avatar:           []byte("avatar2"),
				VoteCount:        1,
				Question:         "My question",
				Round:            1,
				SubmitDeadline:   pgtype.Timestamp{Time: deadline},
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
				},
				{
					ID:       defaultOtherPlayerID,
					Nickname: "Player 2",
					Avatar:   "avatar2",
					Votes:    1,
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
			sqlc.GameState{}, fmt.Errorf("failed to get game state"),
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

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, defaultHostPlayerID).Return(sqlc.GameState{
			State: sqlc.GAMESTATE_FIBBING_IT_SHOW_QUESTION.String(),
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

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, defaultHostPlayerID).Return(sqlc.GameState{
			State: sqlc.GAMESTATE_FIBBING_IT_VOTING.String(),
		}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, defaultHostPlayerID).Return(
			[]sqlc.GetAllPlayersInRoomRow{}, fmt.Errorf("failed to get all players in room"),
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

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, defaultHostPlayerID).Return(sqlc.GameState{
			State: sqlc.GAMESTATE_FIBBING_IT_VOTING.String(),
		}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, defaultHostPlayerID).Return([]sqlc.GetAllPlayersInRoomRow{
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

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, defaultHostPlayerID).Return(sqlc.GameState{
			State: sqlc.GAMESTATE_FIBBING_IT_VOTING.String(),
		}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, defaultHostPlayerID).Return([]sqlc.GetAllPlayersInRoomRow{
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

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, defaultHostPlayerID).Return(sqlc.GameState{
			State: sqlc.GAMESTATE_FIBBING_IT_VOTING.String(),
		}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, defaultHostPlayerID).Return([]sqlc.GetAllPlayersInRoomRow{
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
			sqlc.GetLatestRoundByPlayerIDRow{}, fmt.Errorf("failed to get latest round"),
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

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, defaultHostPlayerID).Return(sqlc.GameState{
			State: sqlc.GAMESTATE_FIBBING_IT_VOTING.String(),
		}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, defaultHostPlayerID).Return([]sqlc.GetAllPlayersInRoomRow{
			{
				ID:       defaultHostPlayerID,
				Nickname: "Player 1",
			},
			{
				ID:       defaultOtherPlayerID,
				Nickname: "Player 2",
			},
		}, nil)
		mockStore.EXPECT().GetLatestRoundByPlayerID(ctx, defaultHostPlayerID).Return(sqlc.GetLatestRoundByPlayerIDRow{
			ID:             roundID,
			SubmitDeadline: pgtype.Timestamp{Time: now.Add(-1 * time.Hour)},
		}, nil)

		_, err := srv.SubmitVote(ctx, defaultHostPlayerID, "Player 2", now)
		assert.ErrorContains(t, err, "answer submission deadline has passed")
	})

	t.Run("Should fail because we fail to get vote count", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		ctx := context.Background()
		now := time.Now().Add(30 * time.Second)

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, defaultHostPlayerID).Return(sqlc.GameState{
			State: sqlc.GAMESTATE_FIBBING_IT_VOTING.String(),
		}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, defaultHostPlayerID).Return([]sqlc.GetAllPlayersInRoomRow{
			{
				ID:       defaultHostPlayerID,
				Nickname: "Player 1",
			},
			{
				ID:       defaultOtherPlayerID,
				Nickname: "Player 2",
			},
		}, nil)
		mockStore.EXPECT().GetLatestRoundByPlayerID(ctx, defaultHostPlayerID).Return(sqlc.GetLatestRoundByPlayerIDRow{
			ID:             roundID,
			SubmitDeadline: pgtype.Timestamp{Time: now.Add(1 * time.Hour)},
		}, nil)
		mockStore.EXPECT().GetVotingState(ctx, roundID).Return(
			[]sqlc.GetVotingStateRow{}, fmt.Errorf("failed to get vote count"),
		)

		_, err := srv.SubmitVote(ctx, defaultHostPlayerID, "Player 2", now)
		assert.Error(t, err)
	})
}
