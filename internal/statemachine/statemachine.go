package statemachine

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofrs/uuid/v5"
	"go.opentelemetry.io/otel/baggage"

	"gitlab.com/hmajid2301/banterbus/internal/telemetry"
)

type State interface {
	Start(ctx context.Context) error
}

type stateMachineEntry struct {
	cancel     context.CancelFunc
	generation int64
}

type Manager struct {
	active      sync.Map
	mu          sync.Mutex
	wg          sync.WaitGroup
	generation  atomic.Int64
	count       atomic.Int64
	shutdownCtx context.Context
	logger      *slog.Logger
}

func NewManager(shutdownCtx context.Context, logger *slog.Logger) *Manager {
	return &Manager{
		shutdownCtx: shutdownCtx,
		logger:      logger,
	}
}

func (m *Manager) Start(ctx context.Context, gameStateID uuid.UUID, state State) {
	stateMachineCtx, cancel := context.WithCancel(m.shutdownCtx)

	if bag := baggage.FromContext(ctx); bag.Len() > 0 {
		stateMachineCtx = baggage.ContextWithBaggage(stateMachineCtx, bag)
	}

	m.mu.Lock()
	gen := m.generation.Add(1)

	// INFO: This handles race conditions where multiple sources try to start the same state concurrently.
	// For example:
	// - Question state timer expires → tries to start voting state
	// - Player clicks ready at the same time → also tries to start voting state
	// The manager cancels the old state machine and starts the new one, ensuring only one runs.
	// This is safe because the service layer enforces state transitions, so duplicate transitions
	// will fail gracefully (e.g., ErrNotInQuestionState if already in voting).
	if existingInterface, loaded := m.active.Load(gameStateID); loaded {
		if existing, ok := existingInterface.(*stateMachineEntry); ok {
			m.logger.DebugContext(ctx, "canceling existing state machine before starting new one",
				slog.String("game_state_id", gameStateID.String()),
				slog.Int64("old_generation", existing.generation),
				slog.Int64("new_generation", gen))
			existing.cancel()
		}
	}

	entry := &stateMachineEntry{
		cancel:     cancel,
		generation: gen,
	}
	m.active.Store(gameStateID, entry)
	m.wg.Add(1)
	count := m.count.Add(1)
	m.mu.Unlock()

	telemetry.UpdateActiveStateMachineCount(count)

	go func() {
		defer func() {
			if currentInterface, loaded := m.active.Load(gameStateID); loaded {
				if current, ok := currentInterface.(*stateMachineEntry); ok && current.generation == gen {
					m.active.Delete(gameStateID)
					count := m.count.Add(-1)
					telemetry.UpdateActiveStateMachineCount(count)
				}
			}
			cancel()
			m.wg.Done()
		}()

		if err := state.Start(stateMachineCtx); err != nil {
			m.logger.ErrorContext(stateMachineCtx, "state machine failed",
				slog.Any("error", err),
				slog.String("game_state_id", gameStateID.String()))
		}
	}()
}

func (m *Manager) Stop(ctx context.Context, gameStateID uuid.UUID) {
	if entryInterface, loaded := m.active.LoadAndDelete(gameStateID); loaded {
		if entry, ok := entryInterface.(*stateMachineEntry); ok {
			m.logger.DebugContext(ctx, "stopping state machine",
				slog.String("game_state_id", gameStateID.String()))
			entry.cancel()
		}
	}
}

func (m *Manager) CancelAll(ctx context.Context) {
	m.logger.InfoContext(ctx, "canceling all active state machines for graceful shutdown")

	m.mu.Lock()
	defer m.mu.Unlock()

	var toDelete []uuid.UUID
	count := 0

	m.active.Range(func(key, value any) bool {
		if gameStateID, ok := key.(uuid.UUID); ok {
			toDelete = append(toDelete, gameStateID)
			if entry, ok := value.(*stateMachineEntry); ok {
				m.logger.DebugContext(ctx, "canceling state machine",
					slog.String("game_state_id", gameStateID.String()))
				entry.cancel()
				count++
			}
		}
		return true
	})

	for _, id := range toDelete {
		m.active.Delete(id)
	}

	currentCount := m.count.Add(-int64(count))

	m.logger.InfoContext(ctx, "canceled all active state machines",
		slog.Int("count", count))

	telemetry.UpdateActiveStateMachineCount(currentCount)
}

func (m *Manager) Wait(ctx context.Context, timeout time.Duration) bool {
	m.logger.InfoContext(ctx, "waiting for state machines to complete database writes",
		slog.Duration("timeout", timeout))

	done := make(chan struct{})
	go func() {
		defer close(done)
		m.wg.Wait()
	}()

	select {
	case <-done:
		m.logger.InfoContext(ctx, "all state machines completed gracefully")
		return true
	case <-time.After(timeout):
		m.logger.WarnContext(ctx, "timeout waiting for state machines, forcing shutdown")
		return false
	}
}
