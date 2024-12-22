package websockets

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"gitlab.com/hmajid2301/banterbus/internal/service"
	"gitlab.com/hmajid2301/banterbus/internal/store/db"
)

type State interface {
	Start(ctx context.Context)
	Update(ctx context.Context)
}

type QuestionState struct {
	gameStateID uuid.UUID
	subscriber  Subscriber
}

func (q *QuestionState) Start(ctx context.Context) {
	deadline := time.Now().UTC().Add(q.subscriber.config.Timings.ShowQuestionScreenFor)

	time.Sleep(time.Until(deadline))
	v := &VotingState{gameStateID: q.gameStateID, subscriber: q.subscriber}
	go v.Start(ctx)
}

type VotingState struct {
	subscriber  Subscriber
	gameStateID uuid.UUID
}

func (v *VotingState) Start(ctx context.Context) {
	deadline := time.Now().UTC().Add(v.subscriber.config.Timings.ShowVotingScreenFor)
	votingState, err := v.subscriber.roundService.UpdateStateToVoting(ctx, v.gameStateID, deadline)
	if err != nil {
		v.subscriber.logger.ErrorContext(
			ctx,
			"failed to update game state to voting",
			slog.Any("error", err),
			slog.String("game_state_id", v.gameStateID.String()),
		)
		return
	}

	err = v.subscriber.updateClientsAboutVoting(ctx, votingState)
	if err != nil {
		v.subscriber.logger.ErrorContext(
			ctx,
			"failed to update clients to voting screen",
			slog.Any("error", err),
			slog.String("game_state_id", v.gameStateID.String()),
		)
		return
	}

	time.Sleep(time.Until(deadline))
	r := &RevealState{gameStateID: v.gameStateID, subscriber: v.subscriber}
	go r.Start(ctx)
}

type RevealState struct {
	subscriber  Subscriber
	gameStateID uuid.UUID
}

func (r *RevealState) Start(ctx context.Context) {
	deadline := time.Now().UTC().Add(r.subscriber.config.Timings.ShowRevealScreenFor)
	revealState, err := r.subscriber.roundService.UpdateStateToReveal(ctx, r.gameStateID, deadline)
	if err != nil {
		r.subscriber.logger.ErrorContext(
			ctx,
			"failed to update game state to reveal",
			slog.Any("error", err),
			slog.String("game_state_id", r.gameStateID.String()),
		)
		return
	}

	err = r.subscriber.updateClientsAboutReveal(ctx, revealState)
	if err != nil {
		r.subscriber.logger.ErrorContext(
			ctx,
			"failed to update clients to reveal screen",
			slog.Any("error", err),
			slog.String("game_state_id", r.gameStateID.String()),
		)
		return
	}

	maxRounds := 3
	finalRound := revealState.Round == maxRounds
	fibberFound := revealState.ShouldReveal && revealState.VotedForPlayerRole == "fibber"
	nextState := db.FibbingITQuestion

	if finalRound || fibberFound {
		nextState = db.FibbingItScoring
	}

	time.Sleep(time.Until(deadline))
	if nextState == db.FibbingItScoring {
		s := &ScoringState{gameStateID: r.gameStateID, subscriber: r.subscriber}
		go s.Start(ctx)
	} else {
		q := &QuestionState{gameStateID: r.gameStateID, subscriber: r.subscriber}
		go q.Start(ctx)
	}
}

type ScoringState struct {
	subscriber  Subscriber
	gameStateID uuid.UUID
}

func (r *ScoringState) Start(ctx context.Context) {
	deadline := time.Now().UTC().Add(r.subscriber.config.Timings.ShowScoreScreenFor)
	scoring := service.Scoring{
		GuessedFibber:      r.subscriber.config.Scoring.GuessFibber,
		FibberEvadeCapture: r.subscriber.config.Scoring.FibberEvadeCapture,
	}

	scoringState, err := r.subscriber.roundService.UpdateStateToScore(ctx, r.gameStateID, deadline, scoring)
	if err != nil {
		r.subscriber.logger.ErrorContext(
			ctx,
			"failed to update game state to scoring",
			slog.Any("error", err),
			slog.String("game_state_id", r.gameStateID.String()),
		)
		return
	}

	err = r.subscriber.updateClientsAboutScore(ctx, scoringState)
	if err != nil {
		r.subscriber.logger.ErrorContext(
			ctx,
			"failed to update clients to scoring screen",
			slog.Any("error", err),
			slog.String("game_state_id", r.gameStateID.String()),
		)
		return
	}

	if scoringState.RoundType == "most_likely" && scoringState.RoundNumber == 3 {
		time.Sleep(time.Until(deadline))
		q := &FinalState{gameStateID: r.gameStateID, subscriber: r.subscriber}
		go q.Start(ctx)
	} else {
		time.Sleep(time.Until(deadline))
		q := &QuestionState{gameStateID: r.gameStateID, subscriber: r.subscriber}
		go q.Start(ctx)
	}
}

type FinalState struct {
	subscriber  Subscriber
	gameStateID uuid.UUID
}

func (r *FinalState) Start(ctx context.Context) {
	fmt.Println("Game Over", ctx)
}
