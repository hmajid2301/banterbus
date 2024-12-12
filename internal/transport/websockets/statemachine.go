package websockets

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"gitlab.com/hmajid2301/banterbus/internal/config"
)

type State interface {
	Start(ctx context.Context)
	Update(ctx context.Context)
}

type QuestionState struct {
	GameStateID uuid.UUID
	Subscriber  Subscriber
}

func (q *QuestionState) Start(ctx context.Context) {
	q.Update(ctx)
}

func (q *QuestionState) Update(ctx context.Context) {
	deadline := time.Now().UTC().Add(config.ShowVotingScreenFor)
	v := &VotingState{GameStateID: q.GameStateID, Deadline: deadline, Subscriber: q.Subscriber}
	go v.Start(ctx)
}

type VotingState struct {
	Subscriber  Subscriber
	GameStateID uuid.UUID
	Deadline    time.Time
}

func (v *VotingState) Start(ctx context.Context) {
	time.Sleep(time.Until(v.Deadline))

	deadline := time.Now().UTC().Add(config.ShowVotingScreenFor)
	votingState, err := v.Subscriber.roundService.UpdateStateToVoting(ctx, v.GameStateID, deadline)
	if err != nil {
		v.Subscriber.logger.Error(
			"failed to update game state to voting",
			slog.Any("error", err),
			slog.String("game_state_id", v.GameStateID.String()),
		)
		return
	}

	err = v.Subscriber.updateClientsAboutVoting(ctx, votingState)
	if err != nil {
		v.Subscriber.logger.Error(
			"failed to update clients to voting screen",
			slog.Any("error", err),
			slog.String("game_state_id", v.GameStateID.String()),
		)
		return
	}

	v.Update(ctx)
}

func (v *VotingState) Update(ctx context.Context) {
	fmt.Println("VotingState.Update", ctx)
}

type RevealState struct {
	Deadline time.Time
}

type ScoringState struct {
	Deadline time.Time
}

func StartStateMachine(ctx context.Context, initialState State) {
	initialState.Start(ctx)
}
