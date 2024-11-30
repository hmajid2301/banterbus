package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gitlab.com/hmajid2301/banterbus/internal/service"
	mockService "gitlab.com/hmajid2301/banterbus/internal/service/mocks"
	sqlc "gitlab.com/hmajid2301/banterbus/internal/store/db"
)

func TestRoundServiceSubmitAnswer(t *testing.T) {
	roomID := "fbb75599-9f7a-4392-b523-fd433b3208ea"
	roundID := "222222-fbb75599-9f7a-4392-b523-fd433b3208ea"
	playerID := "33333-9f7a-4392-b523-fd433b3208ea"

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
				SubmitDeadline: now.Add(1 * time.Hour),
			}, nil)

		mockRandomizer.EXPECT().GetID().Return("12345")
		mockStore.EXPECT().AddFibbingItAnswer(ctx, sqlc.AddFibbingItAnswerParams{
			ID:       "12345",
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
				SubmitDeadline: now.Add(1 * time.Hour),
			}, nil)

		mockRandomizer.EXPECT().GetID().Return("12345")
		mockStore.EXPECT().AddFibbingItAnswer(ctx, sqlc.AddFibbingItAnswerParams{
			ID:       "12345",
			RoundID:  roundID,
			PlayerID: playerID,
			Answer:   "My answer",
		}).Return(sqlc.FibbingItAnswer{}, fmt.Errorf("failed to add answer to DB"))

		err := srv.SubmitAnswer(ctx, playerID, "My answer", now)
		assert.Error(t, err)
	})
}
func TestRoundServiceMoveToVoting(t *testing.T) {
	gameStateID := "fbb75599-9f7a-4392-b523-fd433b3208ea"

	t.Run("Should successfully update state to voting", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		players := []service.PlayerWithRole{
			{
				ID:       "12345",
				Nickname: "Player 1",
				Avatar:   []byte("avatar1"),
			},
			{
				ID:       "54678",
				Nickname: "Player 2",
				Avatar:   []byte("avatar2"),
			},
		}

		ctx := context.Background()
		now := time.Now().Add(30 * time.Second)

		mockStore.EXPECT().UpdateGameState(ctx, sqlc.UpdateGameStateParams{
			ID:             gameStateID,
			SubmitDeadline: now,
			State:          sqlc.GAMESTATE_FIBBING_IT_VOTING.String(),
		}).Return(sqlc.GameState{}, nil)

		votes, err := srv.UpdateStateToVoting(ctx, players, gameStateID, now)
		assert.NoError(t, err)
		expectedVotes := []service.VotingPlayer{
			{
				ID:       "12345",
				Nickname: "Player 1",
				Avatar:   "avatar1",
				Votes:    0,
			},
			{
				ID:       "54678",
				Nickname: "Player 2",
				Avatar:   "avatar2",
				Votes:    0,
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
				ID:       "12345",
				Nickname: "Player 1",
				Avatar:   []byte("avatar1"),
			},
			{
				ID:       "54678",
				Nickname: "Player 2",
				Avatar:   []byte("avatar2"),
			},
		}

		ctx := context.Background()
		now := time.Now().Add(30 * time.Second)

		mockStore.EXPECT().UpdateGameState(ctx, sqlc.UpdateGameStateParams{
			ID:             gameStateID,
			SubmitDeadline: now,
			State:          sqlc.GAMESTATE_FIBBING_IT_VOTING.String(),
		}).Return(
			sqlc.GameState{}, fmt.Errorf("failed to update game state"),
		)

		_, err := srv.UpdateStateToVoting(ctx, players, gameStateID, now)
		assert.Error(t, err)
	})
}

func TestRoundServiceSubmitVote(t *testing.T) {
	roundID := "222222-fbb75599-9f7a-4392-b523-fd433b3208ea"

	t.Run("Should successfully submit vote", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		ctx := context.Background()
		now := time.Now().Add(30 * time.Second)

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, "12345").Return(sqlc.GameState{
			State: sqlc.GAMESTATE_FIBBING_IT_VOTING.String(),
		}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, "12345").Return([]sqlc.GetAllPlayersInRoomRow{
			{
				ID:       "12345",
				Nickname: "Player 1",
			},
			{
				ID:       "54678",
				Nickname: "Player 2",
			},
		}, nil)
		mockStore.EXPECT().GetLatestRoundByPlayerID(ctx, "12345").Return(sqlc.GetLatestRoundByPlayerIDRow{
			ID:             roundID,
			SubmitDeadline: now.Add(1 * time.Hour),
		}, nil)
		mockStore.EXPECT().CountVotesByRoundID(ctx, roundID).Return([]sqlc.CountVotesByRoundIDRow{
			{
				VotedForPlayerID: "12345",
				Nickname:         "Player 1",
				Avatar:           []byte("avatar1"),
				VoteCount:        0,
			},
			{
				VotedForPlayerID: "54678",
				Nickname:         "Player 2",
				Avatar:           []byte("avatar2"),
				VoteCount:        1,
			},
		}, nil)

		votes, err := srv.SubmitVote(ctx, "12345", "Player 2", now)
		assert.NoError(t, err)
		expectedVotes := []service.VotingPlayer{
			{
				ID:       "12345",
				Nickname: "Player 1",
				Avatar:   "avatar1",
				Votes:    0,
			},
			{
				ID:       "54678",
				Nickname: "Player 2",
				Avatar:   "avatar2",
				Votes:    1,
			},
		}
		assert.Equal(t, expectedVotes, votes)
	})

	t.Run("Should fail because we fail to get game state", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		ctx := context.Background()
		now := time.Now().Add(30 * time.Second)

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, "12345").Return(
			sqlc.GameState{}, fmt.Errorf("failed to get game state"),
		)

		_, err := srv.SubmitVote(ctx, "12345", "Player 2", now)
		assert.Error(t, err)
	})

	t.Run("Should fail because game state not in voting state", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		ctx := context.Background()
		now := time.Now().Add(30 * time.Second)

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, "12345").Return(sqlc.GameState{
			State: sqlc.GAMESTATE_FIBBING_IT_SHOW_QUESTION.String(),
		}, nil)

		_, err := srv.SubmitVote(ctx, "12345", "Player 2", now)
		assert.ErrorContains(t, err, "game state is not in FIBBING_IT_VOTING state")
	})

	t.Run("Should fail because we fail to get all players in room", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		ctx := context.Background()
		now := time.Now().Add(30 * time.Second)

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, "12345").Return(sqlc.GameState{
			State: sqlc.GAMESTATE_FIBBING_IT_VOTING.String(),
		}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, "12345").Return(
			[]sqlc.GetAllPlayersInRoomRow{}, fmt.Errorf("failed to get all players in room"),
		)

		_, err := srv.SubmitVote(ctx, "12345", "Player 2", now)
		assert.Error(t, err)
	})

	t.Run("Should fail because we voted for themselves", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		ctx := context.Background()
		now := time.Now().Add(30 * time.Second)

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, "12345").Return(sqlc.GameState{
			State: sqlc.GAMESTATE_FIBBING_IT_VOTING.String(),
		}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, "12345").Return([]sqlc.GetAllPlayersInRoomRow{
			{
				ID:       "12345",
				Nickname: "Player 1",
			},
			{
				ID:       "54678",
				Nickname: "Player 2",
			},
		}, nil)

		_, err := srv.SubmitVote(ctx, "12345", "Player 1", now)
		assert.ErrorContains(t, err, "cannot vote for yourself")
	})

	t.Run("Should fail because nickname not found", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		ctx := context.Background()
		now := time.Now().Add(30 * time.Second)

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, "12345").Return(sqlc.GameState{
			State: sqlc.GAMESTATE_FIBBING_IT_VOTING.String(),
		}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, "12345").Return([]sqlc.GetAllPlayersInRoomRow{
			{
				ID:       "12345",
				Nickname: "Player 1",
			},
			{
				ID:       "54678",
				Nickname: "Player 2",
			},
		}, nil)

		_, err := srv.SubmitVote(ctx, "12345", "not_in_room", now)
		assert.ErrorContains(t, err, "player with nickname not_in_room not found")
	})

	t.Run("Should fail because we failed to get round information", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		ctx := context.Background()
		now := time.Now().Add(30 * time.Second)

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, "12345").Return(sqlc.GameState{
			State: sqlc.GAMESTATE_FIBBING_IT_VOTING.String(),
		}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, "12345").Return([]sqlc.GetAllPlayersInRoomRow{
			{
				ID:       "12345",
				Nickname: "Player 1",
			},
			{
				ID:       "54678",
				Nickname: "Player 2",
			},
		}, nil)
		mockStore.EXPECT().GetLatestRoundByPlayerID(ctx, "12345").Return(
			sqlc.GetLatestRoundByPlayerIDRow{}, fmt.Errorf("failed to get latest round"),
		)

		_, err := srv.SubmitVote(ctx, "12345", "Player 2", now)
		assert.Error(t, err)
	})

	t.Run("Should fail because we are passed the submit deadline", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		ctx := context.Background()
		now := time.Now().Add(30 * time.Second)

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, "12345").Return(sqlc.GameState{
			State: sqlc.GAMESTATE_FIBBING_IT_VOTING.String(),
		}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, "12345").Return([]sqlc.GetAllPlayersInRoomRow{
			{
				ID:       "12345",
				Nickname: "Player 1",
			},
			{
				ID:       "54678",
				Nickname: "Player 2",
			},
		}, nil)
		mockStore.EXPECT().GetLatestRoundByPlayerID(ctx, "12345").Return(sqlc.GetLatestRoundByPlayerIDRow{
			ID:             roundID,
			SubmitDeadline: now.Add(-1 * time.Hour),
		}, nil)

		_, err := srv.SubmitVote(ctx, "12345", "Player 2", now)
		assert.ErrorContains(t, err, "answer submission deadline has passed")
	})

	t.Run("Should fail because we fail to get vote count", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandomizer := mockService.NewMockRandomizer(t)
		srv := service.NewRoundService(mockStore, mockRandomizer)

		ctx := context.Background()
		now := time.Now().Add(30 * time.Second)

		mockStore.EXPECT().GetGameStateByPlayerID(ctx, "12345").Return(sqlc.GameState{
			State: sqlc.GAMESTATE_FIBBING_IT_VOTING.String(),
		}, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, "12345").Return([]sqlc.GetAllPlayersInRoomRow{
			{
				ID:       "12345",
				Nickname: "Player 1",
			},
			{
				ID:       "54678",
				Nickname: "Player 2",
			},
		}, nil)
		mockStore.EXPECT().GetLatestRoundByPlayerID(ctx, "12345").Return(sqlc.GetLatestRoundByPlayerIDRow{
			ID:             roundID,
			SubmitDeadline: now.Add(1 * time.Hour),
		}, nil)
		mockStore.EXPECT().CountVotesByRoundID(ctx, roundID).Return(
			[]sqlc.CountVotesByRoundIDRow{}, fmt.Errorf("failed to get vote count"),
		)

		_, err := srv.SubmitVote(ctx, "12345", "Player 2", now)
		assert.Error(t, err)
	})
}
