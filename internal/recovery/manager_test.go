package recovery_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"gitlab.com/hmajid2301/banterbus/internal/recovery"
	mockRecovery "gitlab.com/hmajid2301/banterbus/internal/recovery/mocks"
	"gitlab.com/hmajid2301/banterbus/internal/service"
	"gitlab.com/hmajid2301/banterbus/internal/statemachine"
	"gitlab.com/hmajid2301/banterbus/internal/store/db"
)

func TestNewManager(t *testing.T) {
	t.Parallel()

	mockStore := mockRecovery.NewMockRecoveryStore(t)
	mockTransitioner := mockRecovery.NewMockStateTransitioner(t)
	mockPublisher := mockRecovery.NewMockMessagePublisher(t)
	mockRoundService := mockRecovery.NewMockRoundServicer(t)
	logger := slog.Default()

	manager := recovery.NewManager(mockStore, mockTransitioner, mockPublisher, mockRoundService, logger)

	assert.NotNil(t, manager)
}

func TestRecoverActiveGames(t *testing.T) {
	t.Parallel()

	t.Run("Should return successfully when no active games found", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		mockStore := mockRecovery.NewMockRecoveryStore(t)
		mockTransitioner := mockRecovery.NewMockStateTransitioner(t)
		mockPublisher := mockRecovery.NewMockMessagePublisher(t)
		mockRoundService := mockRecovery.NewMockRoundServicer(t)
		logger := slog.Default()

		manager := recovery.NewManager(mockStore, mockTransitioner, mockPublisher, mockRoundService, logger)

		mockStore.EXPECT().GetActiveGames(ctx).Return([]db.GetActiveGamesRow{}, nil)

		err := manager.RecoverActiveGames(ctx)
		assert.NoError(t, err)

		recovered, failed := manager.GetRecoveryStats()
		assert.Equal(t, int64(0), recovered)
		assert.Equal(t, int64(0), failed)
	})

	t.Run("Should return error when GetActiveGames fails", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		mockStore := mockRecovery.NewMockRecoveryStore(t)
		mockTransitioner := mockRecovery.NewMockStateTransitioner(t)
		mockPublisher := mockRecovery.NewMockMessagePublisher(t)
		mockRoundService := mockRecovery.NewMockRoundServicer(t)
		logger := slog.Default()

		manager := recovery.NewManager(mockStore, mockTransitioner, mockPublisher, mockRoundService, logger)

		expectedErr := errors.New("database error")
		mockStore.EXPECT().GetActiveGames(ctx).Return(nil, expectedErr)

		err := manager.RecoverActiveGames(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get active games")
	})

	t.Run("Should return error when recovery already in progress", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		mockStore := mockRecovery.NewMockRecoveryStore(t)
		mockTransitioner := mockRecovery.NewMockStateTransitioner(t)
		mockPublisher := mockRecovery.NewMockMessagePublisher(t)
		mockRoundService := mockRecovery.NewMockRoundServicer(t)
		logger := slog.Default()

		manager := recovery.NewManager(mockStore, mockTransitioner, mockPublisher, mockRoundService, logger)

		gameStateID := uuid.Must(uuid.NewV7())
		roomID := uuid.Must(uuid.NewV7())

		games := []db.GetActiveGamesRow{
			{
				GameStateID:    gameStateID,
				RoomID:         roomID,
				RoomCode:       "ABCD",
				State:          db.FibbingITQuestion.String(),
				SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(10 * time.Minute)},
			},
		}

		deps := &statemachine.StateDependencies{
			Timings: statemachine.Timings{
				ShowQuestionScreenFor: 60 * time.Second,
				ShowVotingScreenFor:   30 * time.Second,
				ShowRevealScreenFor:   15 * time.Second,
				ShowScoreScreenFor:    20 * time.Second,
				ShowWinnerScreenFor:   30 * time.Second,
			},
		}

		started := make(chan struct{})
		mockStore.EXPECT().GetActiveGames(ctx).RunAndReturn(func(ctx context.Context) ([]db.GetActiveGamesRow, error) {
			close(started)
			time.Sleep(100 * time.Millisecond)
			return games, nil
		}).Once()
		mockStore.EXPECT().TryAcquireGameLock(ctx, gameStateID.String()).Return(true, nil)
		mockTransitioner.EXPECT().NewStateDependencies().Return(deps, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, roomID).Return([]db.GetAllPlayersInRoomRow{}, nil)
		mockTransitioner.EXPECT().StartStateMachine(ctx, gameStateID, mock.AnythingOfType("*statemachine.QuestionState"))
		mockStore.EXPECT().ReleaseGameLock(ctx, gameStateID.String()).Return(nil)

		go func() {
			_ = manager.RecoverActiveGames(ctx)
		}()

		<-started

		err := manager.RecoverActiveGames(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "recovery already in progress")

		time.Sleep(200 * time.Millisecond)
	})

	t.Run("Should skip game when lock acquisition fails", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		mockStore := mockRecovery.NewMockRecoveryStore(t)
		mockTransitioner := mockRecovery.NewMockStateTransitioner(t)
		mockPublisher := mockRecovery.NewMockMessagePublisher(t)
		mockRoundService := mockRecovery.NewMockRoundServicer(t)
		logger := slog.Default()

		manager := recovery.NewManager(mockStore, mockTransitioner, mockPublisher, mockRoundService, logger)

		gameStateID := uuid.Must(uuid.NewV7())
		roomID := uuid.Must(uuid.NewV7())

		games := []db.GetActiveGamesRow{
			{
				GameStateID:    gameStateID,
				RoomID:         roomID,
				RoomCode:       "ABCD",
				State:          db.FibbingITQuestion.String(),
				SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(10 * time.Minute)},
			},
		}

		mockStore.EXPECT().GetActiveGames(ctx).Return(games, nil)
		mockStore.EXPECT().TryAcquireGameLock(ctx, gameStateID.String()).Return(false, errors.New("lock error"))

		err := manager.RecoverActiveGames(ctx)
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		recovered, failed := manager.GetRecoveryStats()
		assert.Equal(t, int64(0), recovered)
		assert.Equal(t, int64(0), failed)
	})

	t.Run("Should skip game when lock not acquired", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		mockStore := mockRecovery.NewMockRecoveryStore(t)
		mockTransitioner := mockRecovery.NewMockStateTransitioner(t)
		mockPublisher := mockRecovery.NewMockMessagePublisher(t)
		mockRoundService := mockRecovery.NewMockRoundServicer(t)
		logger := slog.Default()

		manager := recovery.NewManager(mockStore, mockTransitioner, mockPublisher, mockRoundService, logger)

		gameStateID := uuid.Must(uuid.NewV7())
		roomID := uuid.Must(uuid.NewV7())

		games := []db.GetActiveGamesRow{
			{
				GameStateID:    gameStateID,
				RoomID:         roomID,
				RoomCode:       "ABCD",
				State:          db.FibbingITQuestion.String(),
				SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(10 * time.Minute)},
			},
		}

		mockStore.EXPECT().GetActiveGames(ctx).Return(games, nil)
		mockStore.EXPECT().TryAcquireGameLock(ctx, gameStateID.String()).Return(false, nil)

		err := manager.RecoverActiveGames(ctx)
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		recovered, failed := manager.GetRecoveryStats()
		assert.Equal(t, int64(0), recovered)
		assert.Equal(t, int64(0), failed)
	})

	t.Run("Should recover question state with time remaining", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		mockStore := mockRecovery.NewMockRecoveryStore(t)
		mockTransitioner := mockRecovery.NewMockStateTransitioner(t)
		mockPublisher := mockRecovery.NewMockMessagePublisher(t)
		mockRoundService := mockRecovery.NewMockRoundServicer(t)
		logger := slog.Default()

		manager := recovery.NewManager(mockStore, mockTransitioner, mockPublisher, mockRoundService, logger)

		gameStateID := uuid.Must(uuid.NewV7())
		roomID := uuid.Must(uuid.NewV7())

		games := []db.GetActiveGamesRow{
			{
				GameStateID:    gameStateID,
				RoomID:         roomID,
				RoomCode:       "ABCD",
				State:          db.FibbingITQuestion.String(),
				SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(10 * time.Minute)},
			},
		}

		deps := &statemachine.StateDependencies{
			Timings: statemachine.Timings{
				ShowQuestionScreenFor: 60 * time.Second,
				ShowVotingScreenFor:   30 * time.Second,
				ShowRevealScreenFor:   15 * time.Second,
				ShowScoreScreenFor:    20 * time.Second,
				ShowWinnerScreenFor:   30 * time.Second,
			},
		}

		mockStore.EXPECT().GetActiveGames(ctx).Return(games, nil)
		mockStore.EXPECT().TryAcquireGameLock(ctx, gameStateID.String()).Return(true, nil)
		mockTransitioner.EXPECT().NewStateDependencies().Return(deps, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, roomID).Return([]db.GetAllPlayersInRoomRow{}, nil)
		mockTransitioner.EXPECT().StartStateMachine(ctx, gameStateID, mock.AnythingOfType("*statemachine.QuestionState"))
		mockStore.EXPECT().ReleaseGameLock(ctx, gameStateID.String()).Return(nil)

		err := manager.RecoverActiveGames(ctx)
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		recovered, failed := manager.GetRecoveryStats()
		assert.Equal(t, int64(1), recovered)
		assert.Equal(t, int64(0), failed)
	})

	t.Run("Should transition to voting when question deadline passed", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		mockStore := mockRecovery.NewMockRecoveryStore(t)
		mockTransitioner := mockRecovery.NewMockStateTransitioner(t)
		mockPublisher := mockRecovery.NewMockMessagePublisher(t)
		mockRoundService := mockRecovery.NewMockRoundServicer(t)
		logger := slog.Default()

		manager := recovery.NewManager(mockStore, mockTransitioner, mockPublisher, mockRoundService, logger)

		gameStateID := uuid.Must(uuid.NewV7())
		roomID := uuid.Must(uuid.NewV7())

		games := []db.GetActiveGamesRow{
			{
				GameStateID:    gameStateID,
				RoomID:         roomID,
				RoomCode:       "ABCD",
				State:          db.FibbingITQuestion.String(),
				SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(-10 * time.Minute)},
			},
		}

		deps := &statemachine.StateDependencies{
			Timings: statemachine.Timings{
				ShowQuestionScreenFor: 60 * time.Second,
				ShowVotingScreenFor:   30 * time.Second,
				ShowRevealScreenFor:   15 * time.Second,
				ShowScoreScreenFor:    20 * time.Second,
				ShowWinnerScreenFor:   30 * time.Second,
			},
		}

		mockStore.EXPECT().GetActiveGames(ctx).Return(games, nil)
		mockStore.EXPECT().TryAcquireGameLock(ctx, gameStateID.String()).Return(true, nil)
		mockTransitioner.EXPECT().NewStateDependencies().Return(deps, nil)
		mockRoundService.EXPECT().UpdateStateToVoting(ctx, gameStateID, mock.AnythingOfType("time.Time")).
			Return(service.VotingState{}, nil)
		mockTransitioner.EXPECT().StartStateMachine(ctx, gameStateID, mock.AnythingOfType("*statemachine.VotingState"))
		mockStore.EXPECT().ReleaseGameLock(ctx, gameStateID.String()).Return(nil)

		err := manager.RecoverActiveGames(ctx)
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		recovered, failed := manager.GetRecoveryStats()
		assert.Equal(t, int64(1), recovered)
		assert.Equal(t, int64(0), failed)
	})

	t.Run("Should recover voting state with time remaining", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		mockStore := mockRecovery.NewMockRecoveryStore(t)
		mockTransitioner := mockRecovery.NewMockStateTransitioner(t)
		mockPublisher := mockRecovery.NewMockMessagePublisher(t)
		mockRoundService := mockRecovery.NewMockRoundServicer(t)
		logger := slog.Default()

		manager := recovery.NewManager(mockStore, mockTransitioner, mockPublisher, mockRoundService, logger)

		gameStateID := uuid.Must(uuid.NewV7())
		roomID := uuid.Must(uuid.NewV7())

		games := []db.GetActiveGamesRow{
			{
				GameStateID:    gameStateID,
				RoomID:         roomID,
				RoomCode:       "ABCD",
				State:          db.FibbingItVoting.String(),
				SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(10 * time.Minute)},
			},
		}

		deps := &statemachine.StateDependencies{
			Timings: statemachine.Timings{
				ShowQuestionScreenFor: 60 * time.Second,
				ShowVotingScreenFor:   30 * time.Second,
				ShowRevealScreenFor:   15 * time.Second,
				ShowScoreScreenFor:    20 * time.Second,
				ShowWinnerScreenFor:   30 * time.Second,
			},
		}

		mockStore.EXPECT().GetActiveGames(ctx).Return(games, nil)
		mockStore.EXPECT().TryAcquireGameLock(ctx, gameStateID.String()).Return(true, nil)
		mockTransitioner.EXPECT().NewStateDependencies().Return(deps, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, roomID).Return([]db.GetAllPlayersInRoomRow{}, nil)
		mockTransitioner.EXPECT().StartStateMachine(ctx, gameStateID, mock.AnythingOfType("*statemachine.VotingState"))
		mockStore.EXPECT().ReleaseGameLock(ctx, gameStateID.String()).Return(nil)

		err := manager.RecoverActiveGames(ctx)
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		recovered, failed := manager.GetRecoveryStats()
		assert.Equal(t, int64(1), recovered)
		assert.Equal(t, int64(0), failed)
	})

	t.Run("Should transition to reveal when voting deadline passed", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		mockStore := mockRecovery.NewMockRecoveryStore(t)
		mockTransitioner := mockRecovery.NewMockStateTransitioner(t)
		mockPublisher := mockRecovery.NewMockMessagePublisher(t)
		mockRoundService := mockRecovery.NewMockRoundServicer(t)
		logger := slog.Default()

		manager := recovery.NewManager(mockStore, mockTransitioner, mockPublisher, mockRoundService, logger)

		gameStateID := uuid.Must(uuid.NewV7())
		roomID := uuid.Must(uuid.NewV7())

		games := []db.GetActiveGamesRow{
			{
				GameStateID:    gameStateID,
				RoomID:         roomID,
				RoomCode:       "ABCD",
				State:          db.FibbingItVoting.String(),
				SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(-10 * time.Minute)},
			},
		}

		deps := &statemachine.StateDependencies{
			Timings: statemachine.Timings{
				ShowQuestionScreenFor: 60 * time.Second,
				ShowVotingScreenFor:   30 * time.Second,
				ShowRevealScreenFor:   15 * time.Second,
				ShowScoreScreenFor:    20 * time.Second,
				ShowWinnerScreenFor:   30 * time.Second,
			},
		}

		mockStore.EXPECT().GetActiveGames(ctx).Return(games, nil)
		mockStore.EXPECT().TryAcquireGameLock(ctx, gameStateID.String()).Return(true, nil)
		mockTransitioner.EXPECT().NewStateDependencies().Return(deps, nil)
		mockRoundService.EXPECT().UpdateStateToReveal(ctx, gameStateID, mock.AnythingOfType("time.Time")).
			Return(service.RevealRoleState{}, nil)
		mockTransitioner.EXPECT().StartStateMachine(ctx, gameStateID, mock.AnythingOfType("*statemachine.RevealState"))
		mockStore.EXPECT().ReleaseGameLock(ctx, gameStateID.String()).Return(nil)

		err := manager.RecoverActiveGames(ctx)
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		recovered, failed := manager.GetRecoveryStats()
		assert.Equal(t, int64(1), recovered)
		assert.Equal(t, int64(0), failed)
	})

	t.Run("Should recover reveal state with time remaining", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		mockStore := mockRecovery.NewMockRecoveryStore(t)
		mockTransitioner := mockRecovery.NewMockStateTransitioner(t)
		mockPublisher := mockRecovery.NewMockMessagePublisher(t)
		mockRoundService := mockRecovery.NewMockRoundServicer(t)
		logger := slog.Default()

		manager := recovery.NewManager(mockStore, mockTransitioner, mockPublisher, mockRoundService, logger)

		gameStateID := uuid.Must(uuid.NewV7())
		roomID := uuid.Must(uuid.NewV7())

		games := []db.GetActiveGamesRow{
			{
				GameStateID:    gameStateID,
				RoomID:         roomID,
				RoomCode:       "ABCD",
				State:          db.FibbingItReveal.String(),
				SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(10 * time.Minute)},
			},
		}

		deps := &statemachine.StateDependencies{
			Timings: statemachine.Timings{
				ShowQuestionScreenFor: 60 * time.Second,
				ShowVotingScreenFor:   30 * time.Second,
				ShowRevealScreenFor:   15 * time.Second,
				ShowScoreScreenFor:    20 * time.Second,
				ShowWinnerScreenFor:   30 * time.Second,
			},
		}

		mockStore.EXPECT().GetActiveGames(ctx).Return(games, nil)
		mockStore.EXPECT().TryAcquireGameLock(ctx, gameStateID.String()).Return(true, nil)
		mockTransitioner.EXPECT().NewStateDependencies().Return(deps, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, roomID).Return([]db.GetAllPlayersInRoomRow{}, nil)
		mockTransitioner.EXPECT().StartStateMachine(ctx, gameStateID, mock.AnythingOfType("*statemachine.RevealState"))
		mockStore.EXPECT().ReleaseGameLock(ctx, gameStateID.String()).Return(nil)

		err := manager.RecoverActiveGames(ctx)
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		recovered, failed := manager.GetRecoveryStats()
		assert.Equal(t, int64(1), recovered)
		assert.Equal(t, int64(0), failed)
	})

	t.Run("Should transition to score when reveal deadline passed", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		mockStore := mockRecovery.NewMockRecoveryStore(t)
		mockTransitioner := mockRecovery.NewMockStateTransitioner(t)
		mockPublisher := mockRecovery.NewMockMessagePublisher(t)
		mockRoundService := mockRecovery.NewMockRoundServicer(t)
		logger := slog.Default()

		manager := recovery.NewManager(mockStore, mockTransitioner, mockPublisher, mockRoundService, logger)

		gameStateID := uuid.Must(uuid.NewV7())
		roomID := uuid.Must(uuid.NewV7())

		games := []db.GetActiveGamesRow{
			{
				GameStateID:    gameStateID,
				RoomID:         roomID,
				RoomCode:       "ABCD",
				State:          db.FibbingItReveal.String(),
				SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(-10 * time.Minute)},
			},
		}

		deps := &statemachine.StateDependencies{
			Timings: statemachine.Timings{
				ShowQuestionScreenFor: 60 * time.Second,
				ShowVotingScreenFor:   30 * time.Second,
				ShowRevealScreenFor:   15 * time.Second,
				ShowScoreScreenFor:    20 * time.Second,
				ShowWinnerScreenFor:   30 * time.Second,
			},
			Scoring: service.Scoring{},
		}

		mockStore.EXPECT().GetActiveGames(ctx).Return(games, nil)
		mockStore.EXPECT().TryAcquireGameLock(ctx, gameStateID.String()).Return(true, nil)
		mockTransitioner.EXPECT().NewStateDependencies().Return(deps, nil)
		mockRoundService.EXPECT().UpdateStateToScore(ctx, gameStateID, mock.AnythingOfType("time.Time"), deps.Scoring).
			Return(service.ScoreState{}, nil)
		mockTransitioner.EXPECT().StartStateMachine(ctx, gameStateID, mock.AnythingOfType("*statemachine.ScoringState"))
		mockStore.EXPECT().ReleaseGameLock(ctx, gameStateID.String()).Return(nil)

		err := manager.RecoverActiveGames(ctx)
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		recovered, failed := manager.GetRecoveryStats()
		assert.Equal(t, int64(1), recovered)
		assert.Equal(t, int64(0), failed)
	})

	t.Run("Should recover scoring state with time remaining", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		mockStore := mockRecovery.NewMockRecoveryStore(t)
		mockTransitioner := mockRecovery.NewMockStateTransitioner(t)
		mockPublisher := mockRecovery.NewMockMessagePublisher(t)
		mockRoundService := mockRecovery.NewMockRoundServicer(t)
		logger := slog.Default()

		manager := recovery.NewManager(mockStore, mockTransitioner, mockPublisher, mockRoundService, logger)

		gameStateID := uuid.Must(uuid.NewV7())
		roomID := uuid.Must(uuid.NewV7())

		games := []db.GetActiveGamesRow{
			{
				GameStateID:    gameStateID,
				RoomID:         roomID,
				RoomCode:       "ABCD",
				State:          db.FibbingItScoring.String(),
				SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(10 * time.Minute)},
			},
		}

		deps := &statemachine.StateDependencies{
			Timings: statemachine.Timings{
				ShowQuestionScreenFor: 60 * time.Second,
				ShowVotingScreenFor:   30 * time.Second,
				ShowRevealScreenFor:   15 * time.Second,
				ShowScoreScreenFor:    20 * time.Second,
				ShowWinnerScreenFor:   30 * time.Second,
			},
		}

		mockStore.EXPECT().GetActiveGames(ctx).Return(games, nil)
		mockStore.EXPECT().TryAcquireGameLock(ctx, gameStateID.String()).Return(true, nil)
		mockTransitioner.EXPECT().NewStateDependencies().Return(deps, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, roomID).Return([]db.GetAllPlayersInRoomRow{}, nil)
		mockTransitioner.EXPECT().StartStateMachine(ctx, gameStateID, mock.AnythingOfType("*statemachine.ScoringState"))
		mockStore.EXPECT().ReleaseGameLock(ctx, gameStateID.String()).Return(nil)

		err := manager.RecoverActiveGames(ctx)
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		recovered, failed := manager.GetRecoveryStats()
		assert.Equal(t, int64(1), recovered)
		assert.Equal(t, int64(0), failed)
	})

	t.Run("Should transition to winner when scoring deadline passed and game should end", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		mockStore := mockRecovery.NewMockRecoveryStore(t)
		mockTransitioner := mockRecovery.NewMockStateTransitioner(t)
		mockPublisher := mockRecovery.NewMockMessagePublisher(t)
		mockRoundService := mockRecovery.NewMockRoundServicer(t)
		logger := slog.Default()

		manager := recovery.NewManager(mockStore, mockTransitioner, mockPublisher, mockRoundService, logger)

		gameStateID := uuid.Must(uuid.NewV7())
		roomID := uuid.Must(uuid.NewV7())

		games := []db.GetActiveGamesRow{
			{
				GameStateID:    gameStateID,
				RoomID:         roomID,
				RoomCode:       "ABCD",
				State:          db.FibbingItScoring.String(),
				SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(-10 * time.Minute)},
			},
		}

		deps := &statemachine.StateDependencies{
			Timings: statemachine.Timings{
				ShowQuestionScreenFor: 60 * time.Second,
				ShowVotingScreenFor:   30 * time.Second,
				ShowRevealScreenFor:   15 * time.Second,
				ShowScoreScreenFor:    20 * time.Second,
				ShowWinnerScreenFor:   30 * time.Second,
			},
			Scoring: service.Scoring{},
		}

		mockStore.EXPECT().GetActiveGames(ctx).Return(games, nil)
		mockStore.EXPECT().TryAcquireGameLock(ctx, gameStateID.String()).Return(true, nil)
		mockTransitioner.EXPECT().NewStateDependencies().Return(deps, nil)
		mockRoundService.EXPECT().UpdateStateToScore(ctx, gameStateID, mock.AnythingOfType("time.Time"), deps.Scoring).
			Return(service.ScoreState{
				TotalRounds: 3,
				RoundNumber: 3,
				RoundType:   "free_form",
			}, nil)
		mockRoundService.EXPECT().UpdateStateToWinner(ctx, gameStateID, mock.AnythingOfType("time.Time")).
			Return(service.WinnerState{}, nil)
		mockTransitioner.EXPECT().StartStateMachine(ctx, gameStateID, mock.AnythingOfType("*statemachine.WinnerState"))
		mockStore.EXPECT().ReleaseGameLock(ctx, gameStateID.String()).Return(nil)

		err := manager.RecoverActiveGames(ctx)
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		recovered, failed := manager.GetRecoveryStats()
		assert.Equal(t, int64(1), recovered)
		assert.Equal(t, int64(0), failed)
	})

	t.Run("Should transition to question when scoring deadline passed and game should continue", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		mockStore := mockRecovery.NewMockRecoveryStore(t)
		mockTransitioner := mockRecovery.NewMockStateTransitioner(t)
		mockPublisher := mockRecovery.NewMockMessagePublisher(t)
		mockRoundService := mockRecovery.NewMockRoundServicer(t)
		logger := slog.Default()

		manager := recovery.NewManager(mockStore, mockTransitioner, mockPublisher, mockRoundService, logger)

		gameStateID := uuid.Must(uuid.NewV7())
		roomID := uuid.Must(uuid.NewV7())

		games := []db.GetActiveGamesRow{
			{
				GameStateID:    gameStateID,
				RoomID:         roomID,
				RoomCode:       "ABCD",
				State:          db.FibbingItScoring.String(),
				SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(-10 * time.Minute)},
			},
		}

		deps := &statemachine.StateDependencies{
			Timings: statemachine.Timings{
				ShowQuestionScreenFor: 60 * time.Second,
				ShowVotingScreenFor:   30 * time.Second,
				ShowRevealScreenFor:   15 * time.Second,
				ShowScoreScreenFor:    20 * time.Second,
				ShowWinnerScreenFor:   30 * time.Second,
			},
			Scoring: service.Scoring{},
		}

		mockStore.EXPECT().GetActiveGames(ctx).Return(games, nil)
		mockStore.EXPECT().TryAcquireGameLock(ctx, gameStateID.String()).Return(true, nil)
		mockTransitioner.EXPECT().NewStateDependencies().Return(deps, nil)
		mockRoundService.EXPECT().UpdateStateToScore(ctx, gameStateID, mock.AnythingOfType("time.Time"), deps.Scoring).
			Return(service.ScoreState{
				TotalRounds: 2,
				RoundNumber: 2,
				RoundType:   "free_form",
			}, nil)
		mockRoundService.EXPECT().UpdateStateToQuestion(ctx, gameStateID, mock.AnythingOfType("time.Time"), true).
			Return(service.QuestionState{}, nil)
		mockTransitioner.EXPECT().StartStateMachine(ctx, gameStateID, mock.AnythingOfType("*statemachine.QuestionState"))
		mockStore.EXPECT().ReleaseGameLock(ctx, gameStateID.String()).Return(nil)

		err := manager.RecoverActiveGames(ctx)
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		recovered, failed := manager.GetRecoveryStats()
		assert.Equal(t, int64(1), recovered)
		assert.Equal(t, int64(0), failed)
	})

	t.Run("Should recover winner state", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		mockStore := mockRecovery.NewMockRecoveryStore(t)
		mockTransitioner := mockRecovery.NewMockStateTransitioner(t)
		mockPublisher := mockRecovery.NewMockMessagePublisher(t)
		mockRoundService := mockRecovery.NewMockRoundServicer(t)
		logger := slog.Default()

		manager := recovery.NewManager(mockStore, mockTransitioner, mockPublisher, mockRoundService, logger)

		gameStateID := uuid.Must(uuid.NewV7())
		roomID := uuid.Must(uuid.NewV7())

		games := []db.GetActiveGamesRow{
			{
				GameStateID:    gameStateID,
				RoomID:         roomID,
				RoomCode:       "ABCD",
				State:          db.FibbingItWinner.String(),
				SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(10 * time.Minute)},
			},
		}

		deps := &statemachine.StateDependencies{
			Timings: statemachine.Timings{
				ShowQuestionScreenFor: 60 * time.Second,
				ShowVotingScreenFor:   30 * time.Second,
				ShowRevealScreenFor:   15 * time.Second,
				ShowScoreScreenFor:    20 * time.Second,
				ShowWinnerScreenFor:   30 * time.Second,
			},
		}

		mockStore.EXPECT().GetActiveGames(ctx).Return(games, nil)
		mockStore.EXPECT().TryAcquireGameLock(ctx, gameStateID.String()).Return(true, nil)
		mockTransitioner.EXPECT().NewStateDependencies().Return(deps, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, roomID).Return([]db.GetAllPlayersInRoomRow{}, nil)
		mockTransitioner.EXPECT().StartStateMachine(ctx, gameStateID, mock.AnythingOfType("*statemachine.WinnerState"))
		mockStore.EXPECT().ReleaseGameLock(ctx, gameStateID.String()).Return(nil)

		err := manager.RecoverActiveGames(ctx)
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		recovered, failed := manager.GetRecoveryStats()
		assert.Equal(t, int64(1), recovered)
		assert.Equal(t, int64(0), failed)
	})

	t.Run("Should fail recovery for unknown state", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		mockStore := mockRecovery.NewMockRecoveryStore(t)
		mockTransitioner := mockRecovery.NewMockStateTransitioner(t)
		mockPublisher := mockRecovery.NewMockMessagePublisher(t)
		mockRoundService := mockRecovery.NewMockRoundServicer(t)
		logger := slog.Default()

		manager := recovery.NewManager(mockStore, mockTransitioner, mockPublisher, mockRoundService, logger)

		gameStateID := uuid.Must(uuid.NewV7())
		roomID := uuid.Must(uuid.NewV7())

		games := []db.GetActiveGamesRow{
			{
				GameStateID:    gameStateID,
				RoomID:         roomID,
				RoomCode:       "ABCD",
				State:          "unknown_state",
				SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(10 * time.Minute)},
			},
		}

		deps := &statemachine.StateDependencies{
			Timings: statemachine.Timings{
				ShowQuestionScreenFor: 60 * time.Second,
				ShowVotingScreenFor:   30 * time.Second,
				ShowRevealScreenFor:   15 * time.Second,
				ShowScoreScreenFor:    20 * time.Second,
				ShowWinnerScreenFor:   30 * time.Second,
			},
		}

		mockStore.EXPECT().GetActiveGames(ctx).Return(games, nil)
		mockStore.EXPECT().TryAcquireGameLock(ctx, gameStateID.String()).Return(true, nil)
		mockTransitioner.EXPECT().NewStateDependencies().Return(deps, nil)
		mockStore.EXPECT().ReleaseGameLock(ctx, gameStateID.String()).Return(nil)

		err := manager.RecoverActiveGames(ctx)
		assert.NoError(t, err)

		time.Sleep(3 * time.Second)

		recovered, failed := manager.GetRecoveryStats()
		assert.Equal(t, int64(0), recovered)
		assert.Equal(t, int64(1), failed)
	})

	t.Run("Should continue recovery when notify players fails", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		mockStore := mockRecovery.NewMockRecoveryStore(t)
		mockTransitioner := mockRecovery.NewMockStateTransitioner(t)
		mockPublisher := mockRecovery.NewMockMessagePublisher(t)
		mockRoundService := mockRecovery.NewMockRoundServicer(t)
		logger := slog.Default()

		manager := recovery.NewManager(mockStore, mockTransitioner, mockPublisher, mockRoundService, logger)

		gameStateID := uuid.Must(uuid.NewV7())
		roomID := uuid.Must(uuid.NewV7())

		games := []db.GetActiveGamesRow{
			{
				GameStateID:    gameStateID,
				RoomID:         roomID,
				RoomCode:       "ABCD",
				State:          db.FibbingITQuestion.String(),
				SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(10 * time.Minute)},
			},
		}

		deps := &statemachine.StateDependencies{
			Timings: statemachine.Timings{
				ShowQuestionScreenFor: 60 * time.Second,
				ShowVotingScreenFor:   30 * time.Second,
				ShowRevealScreenFor:   15 * time.Second,
				ShowScoreScreenFor:    20 * time.Second,
				ShowWinnerScreenFor:   30 * time.Second,
			},
		}

		mockStore.EXPECT().GetActiveGames(ctx).Return(games, nil)
		mockStore.EXPECT().TryAcquireGameLock(ctx, gameStateID.String()).Return(true, nil)
		mockTransitioner.EXPECT().NewStateDependencies().Return(deps, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, roomID).Return(nil, errors.New("player fetch error"))
		mockTransitioner.EXPECT().StartStateMachine(ctx, gameStateID, mock.AnythingOfType("*statemachine.QuestionState"))
		mockStore.EXPECT().ReleaseGameLock(ctx, gameStateID.String()).Return(nil)

		err := manager.RecoverActiveGames(ctx)
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		recovered, failed := manager.GetRecoveryStats()
		assert.Equal(t, int64(1), recovered)
		assert.Equal(t, int64(0), failed)
	})

	t.Run("Should notify players successfully", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		mockStore := mockRecovery.NewMockRecoveryStore(t)
		mockTransitioner := mockRecovery.NewMockStateTransitioner(t)
		mockPublisher := mockRecovery.NewMockMessagePublisher(t)
		mockRoundService := mockRecovery.NewMockRoundServicer(t)
		logger := slog.Default()

		manager := recovery.NewManager(mockStore, mockTransitioner, mockPublisher, mockRoundService, logger)

		gameStateID := uuid.Must(uuid.NewV7())
		roomID := uuid.Must(uuid.NewV7())
		playerID1 := uuid.Must(uuid.NewV7())
		playerID2 := uuid.Must(uuid.NewV7())

		games := []db.GetActiveGamesRow{
			{
				GameStateID:    gameStateID,
				RoomID:         roomID,
				RoomCode:       "ABCD",
				State:          db.FibbingITQuestion.String(),
				SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(10 * time.Minute)},
			},
		}

		players := []db.GetAllPlayersInRoomRow{
			{ID: playerID1, Nickname: "Player1"},
			{ID: playerID2, Nickname: "Player2"},
		}

		deps := &statemachine.StateDependencies{
			Timings: statemachine.Timings{
				ShowQuestionScreenFor: 60 * time.Second,
				ShowVotingScreenFor:   30 * time.Second,
				ShowRevealScreenFor:   15 * time.Second,
				ShowScoreScreenFor:    20 * time.Second,
				ShowWinnerScreenFor:   30 * time.Second,
			},
		}

		mockStore.EXPECT().GetActiveGames(ctx).Return(games, nil)
		mockStore.EXPECT().TryAcquireGameLock(ctx, gameStateID.String()).Return(true, nil)
		mockTransitioner.EXPECT().NewStateDependencies().Return(deps, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, roomID).Return(players, nil)
		mockPublisher.EXPECT().Publish(ctx, playerID1, mock.AnythingOfType("[]uint8")).Return(nil)
		mockPublisher.EXPECT().Publish(ctx, playerID2, mock.AnythingOfType("[]uint8")).Return(nil)
		mockTransitioner.EXPECT().StartStateMachine(ctx, gameStateID, mock.AnythingOfType("*statemachine.QuestionState"))
		mockStore.EXPECT().ReleaseGameLock(ctx, gameStateID.String()).Return(nil)

		err := manager.RecoverActiveGames(ctx)
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		recovered, failed := manager.GetRecoveryStats()
		assert.Equal(t, int64(1), recovered)
		assert.Equal(t, int64(0), failed)
	})

	t.Run("Should continue recovery when release lock fails", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		mockStore := mockRecovery.NewMockRecoveryStore(t)
		mockTransitioner := mockRecovery.NewMockStateTransitioner(t)
		mockPublisher := mockRecovery.NewMockMessagePublisher(t)
		mockRoundService := mockRecovery.NewMockRoundServicer(t)
		logger := slog.Default()

		manager := recovery.NewManager(mockStore, mockTransitioner, mockPublisher, mockRoundService, logger)

		gameStateID := uuid.Must(uuid.NewV7())
		roomID := uuid.Must(uuid.NewV7())

		games := []db.GetActiveGamesRow{
			{
				GameStateID:    gameStateID,
				RoomID:         roomID,
				RoomCode:       "ABCD",
				State:          db.FibbingITQuestion.String(),
				SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(10 * time.Minute)},
			},
		}

		deps := &statemachine.StateDependencies{
			Timings: statemachine.Timings{
				ShowQuestionScreenFor: 60 * time.Second,
				ShowVotingScreenFor:   30 * time.Second,
				ShowRevealScreenFor:   15 * time.Second,
				ShowScoreScreenFor:    20 * time.Second,
				ShowWinnerScreenFor:   30 * time.Second,
			},
		}

		mockStore.EXPECT().GetActiveGames(ctx).Return(games, nil)
		mockStore.EXPECT().TryAcquireGameLock(ctx, gameStateID.String()).Return(true, nil)
		mockTransitioner.EXPECT().NewStateDependencies().Return(deps, nil)
		mockStore.EXPECT().GetAllPlayersInRoom(ctx, roomID).Return([]db.GetAllPlayersInRoomRow{}, nil)
		mockTransitioner.EXPECT().StartStateMachine(ctx, gameStateID, mock.AnythingOfType("*statemachine.QuestionState"))
		mockStore.EXPECT().ReleaseGameLock(ctx, gameStateID.String()).Return(errors.New("lock release error"))

		err := manager.RecoverActiveGames(ctx)
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		recovered, failed := manager.GetRecoveryStats()
		assert.Equal(t, int64(1), recovered)
		assert.Equal(t, int64(0), failed)
	})

	t.Run("Should fail recovery when state dependencies creation fails", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		mockStore := mockRecovery.NewMockRecoveryStore(t)
		mockTransitioner := mockRecovery.NewMockStateTransitioner(t)
		mockPublisher := mockRecovery.NewMockMessagePublisher(t)
		mockRoundService := mockRecovery.NewMockRoundServicer(t)
		logger := slog.Default()

		manager := recovery.NewManager(mockStore, mockTransitioner, mockPublisher, mockRoundService, logger)

		gameStateID := uuid.Must(uuid.NewV7())
		roomID := uuid.Must(uuid.NewV7())

		games := []db.GetActiveGamesRow{
			{
				GameStateID:    gameStateID,
				RoomID:         roomID,
				RoomCode:       "ABCD",
				State:          db.FibbingITQuestion.String(),
				SubmitDeadline: pgtype.Timestamp{Time: time.Now().Add(10 * time.Minute)},
			},
		}

		mockStore.EXPECT().GetActiveGames(ctx).Return(games, nil)
		mockStore.EXPECT().TryAcquireGameLock(ctx, gameStateID.String()).Return(true, nil)
		mockTransitioner.EXPECT().NewStateDependencies().Return(nil, errors.New("deps error"))
		mockStore.EXPECT().ReleaseGameLock(ctx, gameStateID.String()).Return(nil)

		err := manager.RecoverActiveGames(ctx)
		require.NoError(t, err)

		time.Sleep(3 * time.Second)

		recovered, failed := manager.GetRecoveryStats()
		assert.Equal(t, int64(0), recovered)
		assert.Equal(t, int64(1), failed)
	})
}

func TestIsRecoveryInProgress(t *testing.T) {
	t.Parallel()

	mockStore := mockRecovery.NewMockRecoveryStore(t)
	mockTransitioner := mockRecovery.NewMockStateTransitioner(t)
	mockPublisher := mockRecovery.NewMockMessagePublisher(t)
	mockRoundService := mockRecovery.NewMockRoundServicer(t)
	logger := slog.Default()

	manager := recovery.NewManager(mockStore, mockTransitioner, mockPublisher, mockRoundService, logger)

	assert.False(t, manager.IsRecoveryInProgress())
}

func TestGetRecoveryStats(t *testing.T) {
	t.Parallel()

	mockStore := mockRecovery.NewMockRecoveryStore(t)
	mockTransitioner := mockRecovery.NewMockStateTransitioner(t)
	mockPublisher := mockRecovery.NewMockMessagePublisher(t)
	mockRoundService := mockRecovery.NewMockRoundServicer(t)
	logger := slog.Default()

	manager := recovery.NewManager(mockStore, mockTransitioner, mockPublisher, mockRoundService, logger)

	recovered, failed := manager.GetRecoveryStats()
	assert.Equal(t, int64(0), recovered)
	assert.Equal(t, int64(0), failed)
}
