package statemachine

import (
	"context"
	"log/slog"
	"time"

	"github.com/gofrs/uuid/v5"

	"gitlab.com/hmajid2301/banterbus/internal/service"
	"gitlab.com/hmajid2301/banterbus/internal/store/db"
)

type testRoundService struct{}

func (t *testRoundService) UpdateStateToQuestion(ctx context.Context, gameStateID uuid.UUID, deadline time.Time, nextRound bool) (service.QuestionState, error) {
	return service.QuestionState{}, nil
}

func (t *testRoundService) AreAllPlayersAnswerReady(ctx context.Context, gameStateID uuid.UUID) (bool, error) {
	return false, nil
}

func (t *testRoundService) UpdateStateToVoting(ctx context.Context, gameStateID uuid.UUID, deadline time.Time) (service.VotingState, error) {
	return service.VotingState{}, nil
}

func (t *testRoundService) AreAllPlayersVotingReady(ctx context.Context, gameStateID uuid.UUID) (bool, error) {
	return false, nil
}

func (t *testRoundService) UpdateStateToReveal(ctx context.Context, gameStateID uuid.UUID, deadline time.Time) (service.RevealRoleState, error) {
	return service.RevealRoleState{}, nil
}

func (t *testRoundService) UpdateStateToScore(ctx context.Context, gameStateID uuid.UUID, deadline time.Time, scoring service.Scoring) (service.ScoreState, error) {
	return service.ScoreState{}, nil
}

func (t *testRoundService) UpdateStateToWinner(ctx context.Context, gameStateID uuid.UUID, deadline time.Time) (service.WinnerState, error) {
	return service.WinnerState{}, nil
}

func (t *testRoundService) FinishGame(ctx context.Context, gameStateID uuid.UUID) error {
	return nil
}

func (t *testRoundService) GetGameState(ctx context.Context, gameStateID uuid.UUID) (db.FibbingItGameState, error) {
	return db.FibbingITQuestion, nil
}

type testClientUpdater struct{}

func (t *testClientUpdater) UpdateClientsAboutQuestion(ctx context.Context, questionState service.QuestionState, showModal bool) error {
	return nil
}

func (t *testClientUpdater) UpdateClientsAboutVoting(ctx context.Context, votingState service.VotingState) error {
	return nil
}

func (t *testClientUpdater) UpdateClientsAboutReveal(ctx context.Context, revealState service.RevealRoleState) error {
	return nil
}

func (t *testClientUpdater) UpdateClientsAboutScore(ctx context.Context, scoringState service.ScoreState) error {
	return nil
}

func (t *testClientUpdater) UpdateClientsAboutWinner(ctx context.Context, winnerState service.WinnerState) error {
	return nil
}

type testTransitioner struct{}

func (t *testTransitioner) StartStateMachine(ctx context.Context, gameStateID uuid.UUID, state State) {
}

func testLogger() *slog.Logger {
	return slog.Default()
}

func testScoring() service.Scoring {
	return service.Scoring{
		GuessedFibber:      100,
		FibberEvadeCapture: 150,
	}
}
