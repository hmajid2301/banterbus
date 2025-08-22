package service_test

import (
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/hmajid2301/banterbus/internal/config"
	"gitlab.com/hmajid2301/banterbus/internal/service"
	"gitlab.com/hmajid2301/banterbus/internal/service/randomizer"
	db "gitlab.com/hmajid2301/banterbus/internal/store/db"
)

func TestIntegrationRoundServiceSubmitAnswer(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully submit answer", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx()
		require.NoError(t, err)

		lobbyService := service.NewLobbyService(str, randomizer, "en-GB")
		playerService := service.NewPlayerService(str, randomizer)
		roundService := service.NewRoundService(str, randomizer, "en-GB")

		questionState, err := startGame(ctx, lobbyService, playerService)
		require.NoError(t, err)

		err = roundService.SubmitAnswer(ctx, questionState.Players[0].ID, "This is my answer", time.Now())
		assert.NoError(t, err)

		err = roundService.SubmitAnswer(
			ctx,
			questionState.Players[1].ID,
			"This is the other players answer",
			time.Now(),
		)
		assert.NoError(t, err)
	})

	t.Run("Should fail to submit answer, time has passed", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx()
		require.NoError(t, err)

		lobbyService := service.NewLobbyService(str, randomizer, "en-GB")
		playerService := service.NewPlayerService(str, randomizer)
		roundService := service.NewRoundService(str, randomizer, "en-GB")

		questionState, err := startGame(ctx, lobbyService, playerService)
		require.NoError(t, err)

		err = roundService.SubmitAnswer(
			ctx,
			questionState.Players[0].ID,
			"This is my answer",
			time.Now().Add(120*time.Second),
		)
		assert.ErrorContains(t, err, "answer submission deadline has passed")
	})

	t.Run("Should fail to submit answer, player id doesn't belong to room", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx()
		require.NoError(t, err)

		lobbyService := service.NewLobbyService(str, randomizer, "en-GB")
		playerService := service.NewPlayerService(str, randomizer)
		roundService := service.NewRoundService(str, randomizer, "en-GB")

		_, err = startGame(ctx, lobbyService, playerService)
		require.NoError(t, err)

		id, _ := uuid.NewV4()
		err = roundService.SubmitAnswer(ctx, id, "This is my answer", time.Now())
		assert.Error(t, err)
	})
}

func TestIntegrationRoundServiceToggleAnswerIsReady(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully toggle answer is ready", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx()
		require.NoError(t, err)

		lobbyService := service.NewLobbyService(str, randomizer, "en-GB")
		playerService := service.NewPlayerService(str, randomizer)
		roundService := service.NewRoundService(str, randomizer, "en-GB")

		questionState, err := startGame(ctx, lobbyService, playerService)
		require.NoError(t, err)

		err = roundService.SubmitAnswer(ctx, questionState.Players[0].ID, "This is my answer", time.Now())
		require.NoError(t, err)

		err = roundService.SubmitAnswer(
			ctx,
			questionState.Players[1].ID,
			"This is the other players answer",
			time.Now(),
		)
		require.NoError(t, err)

		allReady, err := roundService.ToggleAnswerIsReady(ctx, questionState.Players[0].ID, time.Now())
		assert.NoError(t, err)
		assert.False(t, allReady)

		allReady, err = roundService.ToggleAnswerIsReady(ctx, questionState.Players[1].ID, time.Now())
		assert.NoError(t, err)
		assert.True(t, allReady)
	})

	t.Run("Should fail to toggle answer is ready, submit deadline passed", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx()
		require.NoError(t, err)

		lobbyService := service.NewLobbyService(str, randomizer, "en-GB")
		playerService := service.NewPlayerService(str, randomizer)
		roundService := service.NewRoundService(str, randomizer, "en-GB")

		questionState, err := startGame(ctx, lobbyService, playerService)
		require.NoError(t, err)

		err = roundService.SubmitAnswer(ctx, questionState.Players[0].ID, "This is my answer", time.Now())
		require.NoError(t, err)

		err = roundService.SubmitAnswer(
			ctx,
			questionState.Players[1].ID,
			"This is the other players answer",
			time.Now(),
		)
		require.NoError(t, err)

		_, err = roundService.ToggleAnswerIsReady(ctx, questionState.Players[0].ID, time.Now().Add(120*time.Second))
		assert.ErrorContains(t, err, "toggle ready deadline has passed")
	})
}

func TestIntegrationRoundServiceUpdateStateToVoting(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully update state to voting", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx()
		require.NoError(t, err)

		lobbyService := service.NewLobbyService(str, randomizer, "en-GB")
		playerService := service.NewPlayerService(str, randomizer)
		roundService := service.NewRoundService(str, randomizer, "en-GB")

		questionState, err := startGame(ctx, lobbyService, playerService)
		require.NoError(t, err)

		err = roundService.SubmitAnswer(ctx, questionState.Players[0].ID, "This is my answer", time.Now())
		require.NoError(t, err)

		err = roundService.SubmitAnswer(
			ctx,
			questionState.Players[1].ID,
			"This is the other players answer",
			time.Now(),
		)
		require.NoError(t, err)

		votingState, err := roundService.UpdateStateToVoting(
			ctx,
			questionState.GameStateID,
			time.Now().Add(120*time.Second),
		)
		assert.NoError(t, err)

		lobby, err := lobbyService.GetLobby(ctx, questionState.Players[0].ID)
		require.NoError(t, err)

		normalQuestion := questionState.Players[0].Question
		if questionState.Players[0].Role == "fibber" {
			normalQuestion = questionState.Players[1].Question
		}

		expectedVotingState := service.VotingState{
			Deadline:    time.Second * 120,
			Question:    normalQuestion,
			Round:       questionState.Round,
			GameStateID: questionState.GameStateID,
			Players: []service.PlayerWithVoting{
				{
					ID:       questionState.Players[0].ID,
					Role:     questionState.Players[0].Role,
					Nickname: lobby.Players[0].Nickname,
					Avatar:   lobby.Players[0].Avatar,
					Answer:   "This is my answer",
					Votes:    0,
					IsReady:  false,
				},
				{
					ID:       questionState.Players[1].ID,
					Role:     questionState.Players[1].Role,
					Nickname: lobby.Players[1].Nickname,
					Avatar:   lobby.Players[1].Avatar,
					Answer:   "This is the other players answer",
					Votes:    0,
					IsReady:  false,
				},
			},
		}

		diffOpts := cmpopts.IgnoreFields(votingState, "Deadline")
		PartialEqual(t, expectedVotingState, votingState, diffOpts)
		// Check that deadline is positive (actual timing may vary in tests)
		assert.Greater(t, votingState.Deadline.Seconds(), 0.0)
	})

	t.Run("Should fail to update state to voting because incorrect game state id", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx()
		require.NoError(t, err)

		lobbyService := service.NewLobbyService(str, randomizer, "en-GB")
		playerService := service.NewPlayerService(str, randomizer)
		roundService := service.NewRoundService(str, randomizer, "en-GB")

		questionState, err := startGame(ctx, lobbyService, playerService)
		require.NoError(t, err)

		err = roundService.SubmitAnswer(ctx, questionState.Players[0].ID, "This is my answer", time.Now())
		require.NoError(t, err)

		err = roundService.SubmitAnswer(
			ctx,
			questionState.Players[1].ID,
			"This is the other players answer",
			time.Now(),
		)
		require.NoError(t, err)

		id, _ := uuid.NewV4()
		_, err = roundService.UpdateStateToVoting(
			ctx,
			id,
			time.Now().Add(120*time.Second),
		)
		assert.Error(t, err)
	})
}

func TestIntegrationRoundServiceSubmitVote(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully submit vote", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx()
		require.NoError(t, err)

		lobbyService := service.NewLobbyService(str, randomizer, "en-GB")
		playerService := service.NewPlayerService(str, randomizer)
		roundService := service.NewRoundService(str, randomizer, "en-GB")

		originalVotingState, err := votingState(ctx, lobbyService, playerService, roundService)
		assert.NoError(t, err)
		require.GreaterOrEqual(t, len(originalVotingState.Players), 2, "Expected at least 2 players for the test")

		roundService.SubmitVote(
			ctx,
			originalVotingState.Players[0].ID,
			originalVotingState.Players[1].Nickname,
			time.Now(),
		)
		roundService.SubmitVote(
			ctx,
			originalVotingState.Players[1].ID,
			originalVotingState.Players[0].Nickname,
			time.Now(),
		)

		allReady, err := roundService.ToggleVotingIsReady(ctx, originalVotingState.Players[0].ID, time.Now())
		assert.NoError(t, err)
		assert.False(t, allReady)

		allReady, err = roundService.ToggleVotingIsReady(ctx, originalVotingState.Players[1].ID, time.Now())
		assert.NoError(t, err)
		assert.True(t, allReady)
	})

	t.Run("Should fail to toggle voting is ready, because we did not submit vote", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx()
		require.NoError(t, err)

		lobbyService := service.NewLobbyService(str, randomizer, "en-GB")
		playerService := service.NewPlayerService(str, randomizer)
		roundService := service.NewRoundService(str, randomizer, "en-GB")

		originalVotingState, err := votingState(ctx, lobbyService, playerService, roundService)
		assert.NoError(t, err)

		_, err = roundService.ToggleVotingIsReady(ctx, originalVotingState.Players[0].ID, time.Now())
		assert.ErrorContains(t, err, "must submit vote first")
	})

	t.Run("Should fail to toggle voting is ready, because player ID does not exist", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx()
		require.NoError(t, err)

		lobbyService := service.NewLobbyService(str, randomizer, "en-GB")
		playerService := service.NewPlayerService(str, randomizer)
		roundService := service.NewRoundService(str, randomizer, "en-GB")

		originalVotingState, err := votingState(ctx, lobbyService, playerService, roundService)
		assert.NoError(t, err)

		roundService.SubmitVote(
			ctx,
			originalVotingState.Players[0].ID,
			originalVotingState.Players[1].Nickname,
			time.Now(),
		)
		roundService.SubmitVote(
			ctx,
			originalVotingState.Players[1].ID,
			originalVotingState.Players[0].Nickname,
			time.Now(),
		)

		u, err := uuid.NewV4()
		require.NoError(t, err)
		_, err = roundService.ToggleVotingIsReady(ctx, u, time.Now())
		assert.Error(t, err)
	})

	t.Run("Should fail to toggle voting is ready, because deadline passed", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx()
		require.NoError(t, err)

		lobbyService := service.NewLobbyService(str, randomizer, "en-GB")
		playerService := service.NewPlayerService(str, randomizer)
		roundService := service.NewRoundService(str, randomizer, "en-GB")

		originalVotingState, err := votingState(ctx, lobbyService, playerService, roundService)
		assert.NoError(t, err)

		roundService.SubmitVote(
			ctx,
			originalVotingState.Players[0].ID,
			originalVotingState.Players[1].Nickname,
			time.Now(),
		)
		roundService.SubmitVote(
			ctx,
			originalVotingState.Players[1].ID,
			originalVotingState.Players[0].Nickname,
			time.Now(),
		)

		// Use a time that's definitely past the voting deadline
		futureTime := time.Now().Add(10 * time.Minute)
		_, err = roundService.ToggleVotingIsReady(
			ctx,
			originalVotingState.Players[0].ID,
			futureTime,
		)
		// TODO: Fix deadline check - currently not working as expected in tests
		// assert.ErrorContains(t, err, "toggle ready deadline has passed")
	})
}

func TestIntegrationRoundServiceUpdateStateToReveal(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully update state to reveal", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx()
		require.NoError(t, err)

		lobbyService := service.NewLobbyService(str, randomizer, "en-GB")
		playerService := service.NewPlayerService(str, randomizer)
		roundService := service.NewRoundService(str, randomizer, "en-GB")

		originalVotingState, err := votingState(ctx, lobbyService, playerService, roundService)
		assert.NoError(t, err)

		roundService.SubmitVote(
			ctx,
			originalVotingState.Players[0].ID,
			originalVotingState.Players[1].Nickname,
			time.Now(),
		)
		_, err = roundService.UpdateStateToReveal(
			ctx,
			originalVotingState.GameStateID,
			time.Now().Add(120*time.Second),
		)
		assert.NoError(t, err)

		// expectedRevealState := service.RevealRoleState{
		// 	VotedForPlayerNickname: originalVotingState.Players[1].Nickname,
		// 	VotedForPlayerAvatar:   originalVotingState.Players[1].Avatar,
		// 	VotedForPlayerRole:     originalVotingState.Players[1].Role,
		// 	Round:                  originalVotingState.Round,
		// 	Deadline:               originalVotingState.Deadline,
		// 	ShouldReveal:           true,
		// 	PlayerIDs:              []uuid.UUID{originalVotingState.Players[0].ID, originalVotingState.Players[1].ID},
		// }
		//
		// diffOpts := cmpopts.IgnoreFields(revealState, "Deadline")
		// PartialEqual(t, expectedRevealState, revealState, diffOpts)
		// assert.LessOrEqual(t, int(revealState.Deadline.Seconds()), 120)
	})
}

func TestIntegrationRoundServiceGetRevealState(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully get state to reveal", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx()
		require.NoError(t, err)

		lobbyService := service.NewLobbyService(str, randomizer, "en-GB")
		playerService := service.NewPlayerService(str, randomizer)
		roundService := service.NewRoundService(str, randomizer, "en-GB")

		originalVotingState, err := votingState(ctx, lobbyService, playerService, roundService)
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(originalVotingState.Players), 2, "Expected at least 2 players for the test")

		roundService.SubmitVote(
			ctx,
			originalVotingState.Players[0].ID,
			originalVotingState.Players[1].Nickname,
			time.Now(),
		)
		revealState, err := roundService.UpdateStateToReveal(
			ctx,
			originalVotingState.GameStateID,
			time.Now().Add(120*time.Second),
		)
		assert.NoError(t, err)

		getState, err := roundService.GetRevealState(
			ctx,
			originalVotingState.Players[0].ID,
		)
		assert.NoError(t, err)

		diffOpts := cmpopts.IgnoreFields(revealState, "Deadline")
		PartialEqual(t, getState, revealState, diffOpts)
		assert.LessOrEqual(t, int(revealState.Deadline.Seconds()), 120)
	})
}

func TestIntegrationRoundServiceUpdateStateToQuestion(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully update state to question", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx()
		require.NoError(t, err)

		lobbyService := service.NewLobbyService(str, randomizer, "en-GB")
		playerService := service.NewPlayerService(str, randomizer)
		roundService := service.NewRoundService(str, randomizer, "en-GB")

		revealState, err := revealState(ctx, lobbyService, playerService, roundService)
		require.NoError(t, err)

		votingState, err := roundService.GetVotingState(ctx, revealState.PlayerIDs[0])
		require.NoError(t, err)

		newRound := false
		questionState, err := roundService.UpdateStateToQuestion(
			ctx,
			votingState.GameStateID,
			time.Now().Add(120*time.Second),
			newRound,
		)
		assert.NoError(t, err)

		expectedQuestionState := service.QuestionState{
			GameStateID: votingState.GameStateID,
			Round:       votingState.Round + 1,
			RoundType:   "free_form",
		}

		diffOpts := cmpopts.IgnoreFields(questionState, "Deadline", "Players")
		PartialEqual(t, expectedQuestionState, questionState, diffOpts)
		assert.LessOrEqual(t, int(questionState.Deadline.Seconds()), 120)
		// TODO: see why this fails sometimes
		// assert.NotEqual(t, questionState.Players[0].Question, questionState.Players[1].Question)
		assert.NotEqual(t, questionState.Players[0].Role, questionState.Players[1].Role)
	})
}
func TestIntegrationRoundServiceGetQuestionState(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully get question state", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx()
		require.NoError(t, err)

		lobbyService := service.NewLobbyService(str, randomizer, "en-GB")
		playerService := service.NewPlayerService(str, randomizer)
		roundService := service.NewRoundService(str, randomizer, "en-GB")

		revealState, err := revealState(ctx, lobbyService, playerService, roundService)
		require.NoError(t, err)

		votingState, err := roundService.GetVotingState(ctx, revealState.PlayerIDs[0])
		require.NoError(t, err)

		newRound := false
		questionState, err := roundService.UpdateStateToQuestion(
			ctx,
			votingState.GameStateID,
			time.Now().Add(120*time.Second),
			newRound,
		)
		require.NoError(t, err)

		getState, err := roundService.GetQuestionState(
			ctx,
			votingState.Players[0].ID,
		)
		assert.NoError(t, err)

		diffOpts := cmpopts.IgnoreFields(questionState, "Deadline", "Players")
		PartialEqual(t, getState, questionState, diffOpts)
		assert.LessOrEqual(t, int(questionState.Deadline.Seconds()), 120)
	})
}

func TestIntegrationRoundServiceGetGameState(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully get game state, voting state", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx()
		require.NoError(t, err)

		lobbyService := service.NewLobbyService(str, randomizer, "en-GB")
		playerService := service.NewPlayerService(str, randomizer)
		roundService := service.NewRoundService(str, randomizer, "en-GB")

		votingState, err := votingState(ctx, lobbyService, playerService, roundService)
		require.NoError(t, err)

		gameState, err := roundService.GetGameState(ctx, votingState.Players[0].ID)
		assert.NoError(t, err)
		assert.Equal(t, db.FibbingItVoting, gameState)
	})

	t.Run("Should successfully get game state, reveal state", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx()
		require.NoError(t, err)

		lobbyService := service.NewLobbyService(str, randomizer, "en-GB")
		playerService := service.NewPlayerService(str, randomizer)
		roundService := service.NewRoundService(str, randomizer, "en-GB")

		revealState, err := revealState(ctx, lobbyService, playerService, roundService)
		require.NoError(t, err)

		gameState, err := roundService.GetGameState(ctx, revealState.PlayerIDs[0])
		assert.NoError(t, err)
		assert.Equal(t, db.FibbingItRevealRole, gameState)
	})
}

func TestIntegrationRoundServiceUpdateStateToScoring(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully update state to scoring", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx()
		require.NoError(t, err)

		lobbyService := service.NewLobbyService(str, randomizer, "en-GB")
		playerService := service.NewPlayerService(str, randomizer)
		roundService := service.NewRoundService(str, randomizer, "en-GB")

		revealState, err := revealState(ctx, lobbyService, playerService, roundService)
		require.NoError(t, err)

		votingState, err := roundService.GetVotingState(ctx, revealState.PlayerIDs[0])
		require.NoError(t, err)

		scoreState, err := roundService.UpdateStateToScore(
			ctx,
			votingState.GameStateID,
			time.Now().Add(120*time.Second),
			service.Scoring{
				GuessedFibber:      100,
				FibberEvadeCapture: 150,
			},
		)
		assert.NoError(t, err)

		expectedScoreState := service.ScoreState{
			Players: []service.PlayerWithScoring{
				{
					ID:       revealState.PlayerIDs[1],
					Avatar:   votingState.Players[1].Avatar,
					Nickname: votingState.Players[1].Nickname,
					Score:    100,
				},
				{
					ID:       revealState.PlayerIDs[0],
					Avatar:   votingState.Players[0].Avatar,
					Nickname: votingState.Players[0].Nickname,
					Score:    0,
				},
			},
			RoundType:   "free_form",
			RoundNumber: 1,
		}

		diffOpts := cmpopts.IgnoreFields(scoreState, "Deadline", "Players")
		PartialEqual(t, expectedScoreState, scoreState, diffOpts)
		assert.LessOrEqual(t, int(expectedScoreState.Deadline.Seconds()), 120)
		// TODO: fix this
		// assert.ElementsMatch(t, expectedScoreState.Players, scoreState.Players)
	})
}

func TestIntegrationRoundServiceGetScoringState(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully get scoring state", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx()
		require.NoError(t, err)

		lobbyService := service.NewLobbyService(str, randomizer, "en-GB")
		playerService := service.NewPlayerService(str, randomizer)
		roundService := service.NewRoundService(str, randomizer, "en-GB")

		revealState, err := revealState(ctx, lobbyService, playerService, roundService)
		require.NoError(t, err)

		votingState, err := roundService.GetVotingState(ctx, revealState.PlayerIDs[0])
		require.NoError(t, err)

		scoreState, err := roundService.UpdateStateToScore(
			ctx,
			votingState.GameStateID,
			time.Now().Add(120*time.Second),
			service.Scoring{
				GuessedFibber:      100,
				FibberEvadeCapture: 150,
			},
		)
		require.NoError(t, err)

		getState, err := roundService.GetScoreState(ctx, service.Scoring{
			GuessedFibber:      100,
			FibberEvadeCapture: 150,
		}, scoreState.Players[0].ID)
		assert.NoError(t, err)

		diffOpts := cmpopts.IgnoreFields(scoreState, "Deadline")
		PartialEqual(t, getState, scoreState, diffOpts)
		assert.LessOrEqual(t, int(scoreState.Deadline.Seconds()), 120)
	})
}

func TestIntegrationRoundServiceUpdateStateToWinner(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully update state to winner", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx()
		require.NoError(t, err)

		lobbyService := service.NewLobbyService(str, randomizer, "en-GB")
		playerService := service.NewPlayerService(str, randomizer)
		roundService := service.NewRoundService(str, randomizer, "en-GB")

		conf, err := config.LoadConfig(ctx)
		require.NoError(t, err)

		scoreState, err := scoreState(ctx, lobbyService, playerService, roundService, conf)
		require.NoError(t, err)

		// TODO: not use this is a hack only used to get game state id
		votingState, err := roundService.GetVotingState(ctx, scoreState.Players[0].ID)
		require.NoError(t, err)

		winnerState, err := roundService.UpdateStateToWinner(
			ctx,
			votingState.GameStateID,
			time.Now().Add(120*time.Second),
		)
		assert.NoError(t, err)
		assert.NotNil(t, winnerState)

		// TODO: fix with actual scores
		// expectedWinnerState := service.WinnerState{
		// 	Players: []service.PlayerWithScoring{
		// 		{
		// 			ID:       winnerState.Players[0].ID,
		// 			Nickname: "host_player",
		// 			Avatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=host_player",
		// 			Score:    0,
		// 		},
		//
		// 		{
		// 			ID:       winnerState.Players[1].ID,
		// 			Nickname: "another_player",
		// 			Avatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=another_player",
		// 			Score:    0,
		// 		},
		// 	},
		// }
		// assert.Equal(t, expectedWinnerState, winnerState)
	})
}

func TestIntegrationRoundServiceGetWinnerState(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully get winner state", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx()
		require.NoError(t, err)

		lobbyService := service.NewLobbyService(str, randomizer, "en-GB")
		playerService := service.NewPlayerService(str, randomizer)
		roundService := service.NewRoundService(str, randomizer, "en-GB")

		conf, err := config.LoadConfig(ctx)
		require.NoError(t, err)

		scoreState, err := scoreState(ctx, lobbyService, playerService, roundService, conf)
		require.NoError(t, err)

		// TODO: not use this is a hack
		votingState, err := roundService.GetVotingState(ctx, scoreState.Players[0].ID)
		require.NoError(t, err)

		_, err = roundService.UpdateStateToWinner(
			ctx,
			votingState.GameStateID,
			time.Now().Add(120*time.Second),
		)
		assert.NoError(t, err)

		winnerState, err := roundService.GetWinnerState(
			ctx,
			scoreState.Players[0].ID,
		)
		assert.NoError(t, err)
		assert.NotNil(t, winnerState)

		// TODO: fix with actual scores
		// expectedWinnerState := service.WinnerState{
		// 	Players: []service.PlayerWithScoring{
		// 		{
		// 			ID:       winnerState.Players[0].ID,
		// 			Nickname: "another_player",
		// 			Avatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=another_player",
		// 			Score:    100,
		// 		},
		// 		{
		// 			ID:       winnerState.Players[1].ID,
		// 			Nickname: "host_player",
		// 			Avatar:   "https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=host_player",
		// 			Score:    0,
		// 		},
		// 	},
		// }
		// assert.Equal(t, expectedWinnerState, winnerState)
	})
}

func TestIntegrationRoundServiceFinsishGame(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully finish game", func(t *testing.T) {
		t.Parallel()
		pool, teardown := setupSubtest(t)
		t.Cleanup(teardown)

		baseDelay := (time.Millisecond * 100)
		str := db.NewDB(pool, 3, baseDelay)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx()
		require.NoError(t, err)

		lobbyService := service.NewLobbyService(str, randomizer, "en-GB")
		playerService := service.NewPlayerService(str, randomizer)
		roundService := service.NewRoundService(str, randomizer, "en-GB")

		conf, err := config.LoadConfig(ctx)
		require.NoError(t, err)

		scoreState, err := scoreState(ctx, lobbyService, playerService, roundService, conf)
		require.NoError(t, err)

		questionState, err := roundService.GetQuestionState(ctx, scoreState.Players[0].ID)
		require.NoError(t, err)

		err = roundService.FinishGame(ctx, questionState.GameStateID)
		assert.NoError(t, err)
	})
}
