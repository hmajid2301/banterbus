package recovery

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/gofrs/uuid/v5"

	"gitlab.com/hmajid2301/banterbus/internal/service"
	"gitlab.com/hmajid2301/banterbus/internal/statemachine"
	"gitlab.com/hmajid2301/banterbus/internal/store/db"
	"gitlab.com/hmajid2301/banterbus/internal/telemetry"
)

const (
	maxRetries              = 3
	retryDelay              = 500 * time.Millisecond
	defaultRecoveryDeadline = 30 * time.Second
)

type RecoveryStore interface {
	GetActiveGames(ctx context.Context) ([]db.GetActiveGamesRow, error)
	TryAcquireGameLock(ctx context.Context, gameStateID string) (bool, error)
	ReleaseGameLock(ctx context.Context, gameStateID string) error
	GetAllPlayersInRoom(ctx context.Context, playerID uuid.UUID) ([]db.GetAllPlayersInRoomRow, error)
}

type StateTransitioner interface {
	StartStateMachine(ctx context.Context, gameStateID uuid.UUID, state statemachine.State)
	NewStateDependencies() (*statemachine.StateDependencies, error)
}

type MessagePublisher interface {
	Publish(ctx context.Context, playerID uuid.UUID, message []byte) error
}

type RoundServicer interface {
	UpdateStateToQuestion(ctx context.Context, gameStateID uuid.UUID, deadline time.Time, nextRound bool) (service.QuestionState, error)
	UpdateStateToVoting(ctx context.Context, gameStateID uuid.UUID, deadline time.Time) (service.VotingState, error)
	UpdateStateToReveal(ctx context.Context, gameStateID uuid.UUID, deadline time.Time) (service.RevealRoleState, error)
	UpdateStateToScore(ctx context.Context, gameStateID uuid.UUID, deadline time.Time, scoring service.Scoring) (service.ScoreState, error)
	UpdateStateToWinner(ctx context.Context, gameStateID uuid.UUID, deadline time.Time) (service.WinnerState, error)
}

type Manager struct {
	store              RecoveryStore
	transitioner       StateTransitioner
	publisher          MessagePublisher
	roundService       RoundServicer
	logger             *slog.Logger
	recoveryInProgress atomic.Bool
	gamesRecovered     atomic.Int64
	gamesFailed        atomic.Int64
}

func NewManager(
	store RecoveryStore,
	transitioner StateTransitioner,
	publisher MessagePublisher,
	roundService RoundServicer,
	logger *slog.Logger,
) *Manager {
	return &Manager{
		store:        store,
		transitioner: transitioner,
		publisher:    publisher,
		roundService: roundService,
		logger:       logger,
	}
}

func (m *Manager) RecoverActiveGames(ctx context.Context) error {
	if !m.recoveryInProgress.CompareAndSwap(false, true) {
		m.logger.WarnContext(ctx, "recovery already in progress, skipping")
		return fmt.Errorf("recovery already in progress")
	}
	defer m.recoveryInProgress.Store(false)

	m.logger.InfoContext(ctx, "starting game recovery process")

	activeGames, err := m.store.GetActiveGames(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active games: %w", err)
	}

	if len(activeGames) == 0 {
		m.logger.InfoContext(ctx, "no active games found to recover")
		return nil
	}

	m.logger.InfoContext(ctx, "found active games to recover",
		slog.Int("count", len(activeGames)))

	skippedCount := 0

	for _, game := range activeGames {
		acquired, err := m.store.TryAcquireGameLock(ctx, game.GameStateID.String())
		if err != nil {
			m.logger.WarnContext(ctx, "failed to acquire lock for game",
				slog.String("game_state_id", game.GameStateID.String()),
				slog.Any("error", err))
			skippedCount++
			continue
		}

		if !acquired {
			m.logger.DebugContext(ctx, "skipping game (another server owns it)",
				slog.String("game_state_id", game.GameStateID.String()))
			skippedCount++
			continue
		}

		m.logger.InfoContext(ctx, "acquired lock for game, starting recovery",
			slog.String("game_state_id", game.GameStateID.String()),
			slog.String("state", game.State),
			slog.String("room_code", game.RoomCode))

		go m.recoverGameWithRetry(ctx, game)
	}

	recovered := m.gamesRecovered.Load()
	failed := m.gamesFailed.Load()

	if err := telemetry.RecordGamesRecoveredCount(ctx, float64(recovered)); err != nil {
		m.logger.WarnContext(ctx, "failed to record games recovered metric", slog.Any("error", err))
	}

	if err := telemetry.RecordGamesRecoveryFailedCount(ctx, float64(failed)); err != nil {
		m.logger.WarnContext(ctx, "failed to record games recovery failed metric", slog.Any("error", err))
	}

	m.logger.InfoContext(ctx, "game recovery process completed",
		slog.Int64("recovered", recovered),
		slog.Int64("failed", failed),
		slog.Int("skipped", skippedCount),
		slog.Int("total", len(activeGames)))

	return nil
}

func (m *Manager) recoverGameWithRetry(ctx context.Context, game db.GetActiveGamesRow) {
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			m.logger.InfoContext(ctx, "retrying game recovery",
				slog.String("game_state_id", game.GameStateID.String()),
				slog.Int("attempt", attempt+1),
				slog.Int("max_retries", maxRetries))
			time.Sleep(retryDelay * time.Duration(attempt))
		}

		err := m.recoverGame(ctx, game)
		if err == nil {
			m.gamesRecovered.Add(1)
			return
		}

		lastErr = err
		m.logger.WarnContext(ctx, "game recovery attempt failed",
			slog.String("game_state_id", game.GameStateID.String()),
			slog.Int("attempt", attempt+1),
			slog.Any("error", err))
	}

	m.gamesFailed.Add(1)
	m.logger.ErrorContext(ctx, "game recovery failed after all retries",
		slog.String("game_state_id", game.GameStateID.String()),
		slog.String("room_code", game.RoomCode),
		slog.String("state", game.State),
		slog.Any("error", lastErr))
}

func (m *Manager) recoverGame(ctx context.Context, game db.GetActiveGamesRow) error {
	defer func() {
		err := m.store.ReleaseGameLock(ctx, game.GameStateID.String())
		if err != nil {
			m.logger.WarnContext(ctx, "failed to release game lock",
				slog.String("game_state_id", game.GameStateID.String()),
				slog.Any("error", err))
		}
	}()

	deadline := game.SubmitDeadline.Time

	m.logger.InfoContext(ctx, "recovering game",
		slog.String("game_state_id", game.GameStateID.String()),
		slog.String("state", game.State),
		slog.Time("deadline", deadline))

	now := time.Now()
	timeRemaining := deadline.Sub(now)

	deps, err := m.transitioner.NewStateDependencies()
	if err != nil {
		return fmt.Errorf("failed to create state dependencies: %w", err)
	}

	var state statemachine.State

	switch game.State {
	case db.FibbingITQuestion.String():
		if timeRemaining <= 0 {
			m.logger.InfoContext(ctx, "question deadline passed, transitioning to voting",
				slog.String("game_state_id", game.GameStateID.String()))
			if err := m.transitionToVoting(ctx, game.GameStateID, deps); err != nil {
				return fmt.Errorf("failed to transition to voting: %w", err)
			}
			return nil
		}
		state, err = statemachine.NewQuestionState(game.GameStateID, false, deps)

	case db.FibbingItVoting.String():
		if timeRemaining <= 0 {
			m.logger.InfoContext(ctx, "voting deadline passed, transitioning to reveal",
				slog.String("game_state_id", game.GameStateID.String()))
			if err := m.transitionToReveal(ctx, game.GameStateID, deps); err != nil {
				return fmt.Errorf("failed to transition to reveal: %w", err)
			}
			return nil
		}
		state, err = statemachine.NewVotingState(game.GameStateID, deps)

	case db.FibbingItReveal.String():
		if timeRemaining <= 0 {
			m.logger.InfoContext(ctx, "reveal deadline passed, transitioning to scoring",
				slog.String("game_state_id", game.GameStateID.String()))
			if err := m.transitionToScore(ctx, game.GameStateID, deps); err != nil {
				return fmt.Errorf("failed to transition to score: %w", err)
			}
			return nil
		}
		state, err = statemachine.NewRevealState(game.GameStateID, deps)

	case db.FibbingItScoring.String():
		if timeRemaining <= 0 {
			m.logger.InfoContext(ctx, "scoring deadline passed, transitioning to next state",
				slog.String("game_state_id", game.GameStateID.String()))
			if err := m.transitionAfterScoring(ctx, game.GameStateID, deps); err != nil {
				return fmt.Errorf("failed to transition after scoring: %w", err)
			}
			return nil
		}
		state, err = statemachine.NewScoringState(game.GameStateID, deps)

	case db.FibbingItWinner.String():
		state, err = statemachine.NewWinnerState(game.GameStateID, deps)

	default:
		return fmt.Errorf("unknown game state: %s", game.State)
	}

	if err != nil {
		return fmt.Errorf("failed to create state machine for state %s: %w", game.State, err)
	}

	err = m.notifyPlayersOfRecovery(ctx, game.RoomID)
	if err != nil {
		m.logger.WarnContext(ctx, "failed to notify players of recovery (non-fatal)",
			slog.String("game_state_id", game.GameStateID.String()),
			slog.Any("error", err))
	}

	m.logger.InfoContext(ctx, "starting recovered state machine",
		slog.String("game_state_id", game.GameStateID.String()),
		slog.String("state", game.State),
		slog.Duration("time_remaining", timeRemaining))

	m.transitioner.StartStateMachine(ctx, game.GameStateID, state)
	return nil
}

func (m *Manager) transitionToVoting(
	ctx context.Context,
	gameStateID uuid.UUID,
	deps *statemachine.StateDependencies,
) error {
	deadline := time.Now().Add(deps.Timings.ShowVotingScreenFor)
	_, err := m.roundService.UpdateStateToVoting(ctx, gameStateID, deadline)
	if err != nil {
		m.logger.ErrorContext(ctx, "failed to transition to voting",
			slog.String("game_state_id", gameStateID.String()),
			slog.Any("error", err))
		return err
	}

	votingState, err := statemachine.NewVotingState(gameStateID, deps)
	if err != nil {
		m.logger.ErrorContext(ctx, "failed to create voting state",
			slog.String("game_state_id", gameStateID.String()),
			slog.Any("error", err))
		return err
	}

	m.transitioner.StartStateMachine(ctx, gameStateID, votingState)
	return nil
}

func (m *Manager) transitionToReveal(
	ctx context.Context,
	gameStateID uuid.UUID,
	deps *statemachine.StateDependencies,
) error {
	deadline := time.Now().Add(deps.Timings.ShowRevealScreenFor)
	_, err := m.roundService.UpdateStateToReveal(ctx, gameStateID, deadline)
	if err != nil {
		m.logger.ErrorContext(ctx, "failed to transition to reveal",
			slog.String("game_state_id", gameStateID.String()),
			slog.Any("error", err))
		return err
	}

	revealState, err := statemachine.NewRevealState(gameStateID, deps)
	if err != nil {
		m.logger.ErrorContext(ctx, "failed to create reveal state",
			slog.String("game_state_id", gameStateID.String()),
			slog.Any("error", err))
		return err
	}

	m.transitioner.StartStateMachine(ctx, gameStateID, revealState)
	return nil
}

func (m *Manager) transitionToScore(
	ctx context.Context,
	gameStateID uuid.UUID,
	deps *statemachine.StateDependencies,
) error {
	deadline := time.Now().Add(deps.Timings.ShowScoreScreenFor)
	_, err := m.roundService.UpdateStateToScore(ctx, gameStateID, deadline, deps.Scoring)
	if err != nil {
		m.logger.ErrorContext(ctx, "failed to transition to score",
			slog.String("game_state_id", gameStateID.String()),
			slog.Any("error", err))
		return err
	}

	scoringState, err := statemachine.NewScoringState(gameStateID, deps)
	if err != nil {
		m.logger.ErrorContext(ctx, "failed to create scoring state",
			slog.String("game_state_id", gameStateID.String()),
			slog.Any("error", err))
		return err
	}

	m.transitioner.StartStateMachine(ctx, gameStateID, scoringState)
	return nil
}

func (m *Manager) transitionAfterScoring(
	ctx context.Context,
	gameStateID uuid.UUID,
	deps *statemachine.StateDependencies,
) error {
	deadline := time.Now().Add(deps.Timings.ShowScoreScreenFor)
	scoreState, err := m.roundService.UpdateStateToScore(ctx, gameStateID, deadline, deps.Scoring)
	if err != nil {
		m.logger.ErrorContext(ctx, "failed to get score state for recovery decision",
			slog.String("game_state_id", gameStateID.String()),
			slog.Any("error", err))
		return err
	}

	shouldEndGame := scoreState.TotalRounds >= 3 || scoreState.RoundType == service.RoundTypeMostLikely

	m.logger.InfoContext(ctx, "scoring recovery transition decision",
		slog.Int("round_number", scoreState.RoundNumber),
		slog.Int("total_rounds", scoreState.TotalRounds),
		slog.String("round_type", scoreState.RoundType),
		slog.Bool("should_end_game", shouldEndGame),
		slog.String("game_state_id", gameStateID.String()))

	if shouldEndGame {
		winnerDeadline := time.Now().Add(deps.Timings.ShowWinnerScreenFor)
		_, err := m.roundService.UpdateStateToWinner(ctx, gameStateID, winnerDeadline)
		if err != nil {
			m.logger.ErrorContext(ctx, "failed to transition to winner",
				slog.String("game_state_id", gameStateID.String()),
				slog.Any("error", err))
			return err
		}

		winnerState, err := statemachine.NewWinnerState(gameStateID, deps)
		if err != nil {
			m.logger.ErrorContext(ctx, "failed to create winner state",
				slog.String("game_state_id", gameStateID.String()),
				slog.Any("error", err))
			return err
		}

		m.transitioner.StartStateMachine(ctx, gameStateID, winnerState)
	} else {
		questionDeadline := time.Now().Add(deps.Timings.ShowQuestionScreenFor)
		_, err := m.roundService.UpdateStateToQuestion(ctx, gameStateID, questionDeadline, true)
		if err != nil {
			m.logger.ErrorContext(ctx, "failed to transition to question",
				slog.String("game_state_id", gameStateID.String()),
				slog.Any("error", err))
			return err
		}

		questionState, err := statemachine.NewQuestionState(gameStateID, true, deps)
		if err != nil {
			m.logger.ErrorContext(ctx, "failed to create question state",
				slog.String("game_state_id", gameStateID.String()),
				slog.Any("error", err))
			return err
		}

		m.transitioner.StartStateMachine(ctx, gameStateID, questionState)
	}

	return nil
}

func (m *Manager) IsRecoveryInProgress() bool {
	return m.recoveryInProgress.Load()
}

func (m *Manager) GetRecoveryStats() (recovered int64, failed int64) {
	return m.gamesRecovered.Load(), m.gamesFailed.Load()
}

func (m *Manager) notifyPlayersOfRecovery(ctx context.Context, roomID uuid.UUID) error {
	players, err := m.store.GetAllPlayersInRoom(ctx, roomID)
	if err != nil {
		return fmt.Errorf("failed to get players in room: %w", err)
	}

	m.logger.DebugContext(ctx, "notifying players of game recovery",
		slog.String("room_id", roomID.String()),
		slog.Int("player_count", len(players)))

	notification := map[string]interface{}{
		"type":    "game_recovered",
		"message": "Game recovered after server restart. Continuing...",
	}

	data, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	for _, player := range players {
		if err := m.publisher.Publish(ctx, player.ID, data); err != nil {
			m.logger.WarnContext(ctx, "failed to publish recovery notification to player",
				slog.String("player_id", player.ID.String()),
				slog.Any("error", err))
		}
	}

	return nil
}
