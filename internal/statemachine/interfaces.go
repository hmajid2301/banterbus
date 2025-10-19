package statemachine

import (
	"context"
	"log/slog"
	"time"

	"github.com/gofrs/uuid/v5"

	"gitlab.com/hmajid2301/banterbus/internal/service"
	"gitlab.com/hmajid2301/banterbus/internal/store/db"
)

type RoundService interface {
	UpdateStateToQuestion(ctx context.Context, gameStateID uuid.UUID, deadline time.Time, nextRound bool) (service.QuestionState, error)
	AreAllPlayersAnswerReady(ctx context.Context, gameStateID uuid.UUID) (bool, error)
	UpdateStateToVoting(ctx context.Context, gameStateID uuid.UUID, deadline time.Time) (service.VotingState, error)
	AreAllPlayersVotingReady(ctx context.Context, gameStateID uuid.UUID) (bool, error)
	UpdateStateToReveal(ctx context.Context, gameStateID uuid.UUID, deadline time.Time) (service.RevealRoleState, error)
	UpdateStateToScore(ctx context.Context, gameStateID uuid.UUID, deadline time.Time, scoring service.Scoring) (service.ScoreState, error)
	UpdateStateToWinner(ctx context.Context, gameStateID uuid.UUID, deadline time.Time) (service.WinnerState, error)
	FinishGame(ctx context.Context, gameStateID uuid.UUID) error
	GetGameStateByID(ctx context.Context, gameStateID uuid.UUID) (db.FibbingItGameState, error)
}

type ClientUpdater interface {
	UpdateClientsAboutQuestion(ctx context.Context, questionState service.QuestionState, showModal bool) error
	UpdateClientsAboutVoting(ctx context.Context, votingState service.VotingState) error
	UpdateClientsAboutReveal(ctx context.Context, revealState service.RevealRoleState) error
	UpdateClientsAboutScore(ctx context.Context, scoringState service.ScoreState) error
	UpdateClientsAboutWinner(ctx context.Context, winnerState service.WinnerState) error
}

type StateTransitioner interface {
	StartStateMachine(ctx context.Context, gameStateID uuid.UUID, state State)
}

type Timings struct {
	ShowQuestionScreenFor time.Duration
	ShowVotingScreenFor   time.Duration
	ShowRevealScreenFor   time.Duration
	ShowScoreScreenFor    time.Duration
	ShowWinnerScreenFor   time.Duration
}

type StateDependencies struct {
	RoundService  RoundService
	ClientUpdater ClientUpdater
	Transitioner  StateTransitioner
	Logger        *slog.Logger
	Timings       Timings
	Scoring       service.Scoring
}
