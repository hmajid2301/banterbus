package statemachine

import (
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewVotingState_ReturnsErrorOnNilDeps(t *testing.T) {
	t.Parallel()

	gameStateID := uuid.Must(uuid.NewV4())

	state, err := NewVotingState(gameStateID, nil)
	require.Error(t, err)
	assert.Nil(t, state)
	assert.Contains(t, err.Error(), "dependencies cannot be nil")
}

func TestNewRevealState_ReturnsErrorOnNilDeps(t *testing.T) {
	t.Parallel()

	gameStateID := uuid.Must(uuid.NewV4())

	state, err := NewRevealState(gameStateID, nil)
	require.Error(t, err)
	assert.Nil(t, state)
	assert.Contains(t, err.Error(), "dependencies cannot be nil")
}

func TestNewScoringState_ReturnsErrorOnNilDeps(t *testing.T) {
	t.Parallel()

	gameStateID := uuid.Must(uuid.NewV4())

	state, err := NewScoringState(gameStateID, nil)
	require.Error(t, err)
	assert.Nil(t, state)
	assert.Contains(t, err.Error(), "dependencies cannot be nil")
}

func TestNewWinnerState_ReturnsErrorOnNilDeps(t *testing.T) {
	t.Parallel()

	gameStateID := uuid.Must(uuid.NewV4())

	state, err := NewWinnerState(gameStateID, nil)
	require.Error(t, err)
	assert.Nil(t, state)
	assert.Contains(t, err.Error(), "dependencies cannot be nil")
}

func TestNewQuestionState_ReturnsErrorOnNilDeps(t *testing.T) {
	t.Parallel()

	gameStateID := uuid.Must(uuid.NewV4())

	state, err := NewQuestionState(gameStateID, false, nil)
	require.Error(t, err)
	assert.Nil(t, state)
	assert.Contains(t, err.Error(), "dependencies cannot be nil")
}
