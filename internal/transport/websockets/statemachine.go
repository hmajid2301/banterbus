package websockets

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"gitlab.com/hmajid2301/banterbus/internal/service"
	"gitlab.com/hmajid2301/banterbus/internal/store/db"
)

type State interface {
	Start(ctx context.Context)
}

type QuestionState struct {
	GameStateID uuid.UUID
	Subscriber  Subscriber
	NextRound   bool
}

func (q *QuestionState) Start(ctx context.Context) {
	deadline := time.Now().UTC().Add(q.Subscriber.config.Timings.ShowQuestionScreenFor)
	questionState, err := q.Subscriber.roundService.UpdateStateToQuestion(ctx, q.GameStateID, deadline, q.NextRound)
	if err != nil {
		q.Subscriber.logger.ErrorContext(
			ctx,
			"failed to update game state to question",
			slog.Any("error", err),
			slog.String("game_state_id", q.GameStateID.String()),
		)
		return
	}

	showModal := true
	err = q.Subscriber.updateClientsAboutQuestion(ctx, questionState, showModal)
	if err != nil {
		q.Subscriber.logger.ErrorContext(
			ctx,
			"failed to update clients to question screen",
			slog.Any("error", err),
			slog.String("game_state_id", q.GameStateID.String()),
		)
		return
	}

	time.Sleep(time.Until(deadline))
	v := &VotingState{GameStateID: q.GameStateID, Subscriber: q.Subscriber}
	go v.Start(ctx)
}

type VotingState struct {
	Subscriber  Subscriber
	GameStateID uuid.UUID
}

func (v *VotingState) Start(ctx context.Context) {
	deadline := time.Now().UTC().Add(v.Subscriber.config.Timings.ShowVotingScreenFor)
	votingState, err := v.Subscriber.roundService.UpdateStateToVoting(ctx, v.GameStateID, deadline)
	if err != nil {
		v.Subscriber.logger.ErrorContext(
			ctx,
			"failed to update game state to voting",
			slog.Any("error", err),
			slog.String("game_state_id", v.GameStateID.String()),
		)
		return
	}

	err = v.Subscriber.updateClientsAboutVoting(ctx, votingState)
	if err != nil {
		v.Subscriber.logger.ErrorContext(
			ctx,
			"failed to update clients to voting screen",
			slog.Any("error", err),
			slog.String("game_state_id", v.GameStateID.String()),
		)
		return
	}

	time.Sleep(time.Until(deadline))
	r := &RevealState{GameStateID: v.GameStateID, Subscriber: v.Subscriber}
	go r.Start(ctx)
}

type RevealState struct {
	Subscriber  Subscriber
	GameStateID uuid.UUID
}

func (r *RevealState) Start(ctx context.Context) {
	deadline := time.Now().UTC().Add(r.Subscriber.config.Timings.ShowRevealScreenFor)
	revealState, err := r.Subscriber.roundService.UpdateStateToReveal(ctx, r.GameStateID, deadline)
	if err != nil {
		r.Subscriber.logger.ErrorContext(
			ctx,
			"failed to update game state to reveal",
			slog.Any("error", err),
			slog.String("game_state_id", r.GameStateID.String()),
		)
		return
	}

	err = r.Subscriber.updateClientsAboutReveal(ctx, revealState)
	if err != nil {
		r.Subscriber.logger.ErrorContext(
			ctx,
			"failed to update clients to reveal screen",
			slog.Any("error", err),
			slog.String("game_state_id", r.GameStateID.String()),
		)
		return
	}

	maxRounds := 3
	finalRound := revealState.Round == maxRounds
	fibberFound := revealState.ShouldReveal && revealState.VotedForPlayerRole == "fibber"
	nextState := db.FibbingITQuestion

	if finalRound || fibberFound {
		nextState = db.FibbingItScoring
		// TODO: turn this string into say ROUND_3
		if revealState.RoundType == "most_likely" {
			nextState = db.FibbingItWinner
		}
	}

	time.Sleep(time.Until(deadline))
	switch nextState {
	case db.FibbingItWinner:
		w := &WinnerState{GameStateID: r.GameStateID, Subscriber: r.Subscriber}
		go w.Start(ctx)
	case db.FibbingItScoring:
		s := &ScoringState{GameStateID: r.GameStateID, Subscriber: r.Subscriber}
		go s.Start(ctx)
	default:
		q := &QuestionState{GameStateID: r.GameStateID, Subscriber: r.Subscriber, NextRound: false}
		go q.Start(ctx)
	}
}

type ScoringState struct {
	Subscriber  Subscriber
	GameStateID uuid.UUID
}

func (r *ScoringState) Start(ctx context.Context) {
	deadline := time.Now().UTC().Add(r.Subscriber.config.Timings.ShowScoreScreenFor)
	scoring := service.Scoring{
		GuessedFibber:      r.Subscriber.config.Scoring.GuessFibber,
		FibberEvadeCapture: r.Subscriber.config.Scoring.FibberEvadeCapture,
	}

	scoringState, err := r.Subscriber.roundService.UpdateStateToScore(ctx, r.GameStateID, deadline, scoring)
	if err != nil {
		r.Subscriber.logger.ErrorContext(
			ctx,
			"failed to update game state to scoring",
			slog.Any("error", err),
			slog.String("game_state_id", r.GameStateID.String()),
		)
		return
	}

	err = r.Subscriber.updateClientsAboutScore(ctx, scoringState)
	if err != nil {
		r.Subscriber.logger.ErrorContext(
			ctx,
			"failed to update clients to scoring screen",
			slog.Any("error", err),
			slog.String("game_state_id", r.GameStateID.String()),
		)
		return
	}

	time.Sleep(time.Until(deadline))
	q := &QuestionState{GameStateID: r.GameStateID, Subscriber: r.Subscriber, NextRound: true}
	go q.Start(ctx)
}

type WinnerState struct {
	Subscriber  Subscriber
	GameStateID uuid.UUID
}

func (r *WinnerState) Start(ctx context.Context) {
	deadline := time.Now().UTC().Add(r.Subscriber.config.Timings.ShowWinnerScreenFor)
	winnerState, err := r.Subscriber.roundService.UpdateStateToWinner(ctx, r.GameStateID, deadline)
	if err != nil {
		r.Subscriber.logger.ErrorContext(
			ctx,
			"failed to update game state to winner",
			slog.Any("error", err),
			slog.String("game_state_id", r.GameStateID.String()),
		)
		return
	}

	err = r.Subscriber.updateClientsAboutWinner(ctx, winnerState)
	if err != nil {
		r.Subscriber.logger.ErrorContext(
			ctx,
			"failed to update clients to winner screen",
			slog.Any("error", err),
			slog.String("game_state_id", r.GameStateID.String()),
		)
		return
	}

	time.Sleep(time.Until(deadline))
	err = r.Subscriber.roundService.FinishGame(ctx, r.GameStateID)
	if err != nil {
		r.Subscriber.logger.ErrorContext(
			ctx,
			"failed to finish",
			slog.Any("error", err),
			slog.String("game_state_id", r.GameStateID.String()),
		)
		return
	}
}
