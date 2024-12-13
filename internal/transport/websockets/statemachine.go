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
	gameStateID uuid.UUID
	subscriber  Subscriber
}

func (q *QuestionState) Start(ctx context.Context) {
	q.Update(ctx)
}

func (q *QuestionState) Update(ctx context.Context) {
	deadline := time.Now().UTC().Add(config.ShowVotingScreenFor)
	v := &VotingState{gameStateID: q.gameStateID, deadline: deadline, subscriber: q.subscriber}
	go v.Start(ctx)
}

type VotingState struct {
	subscriber  Subscriber
	gameStateID uuid.UUID
	deadline    time.Time
}

func (v *VotingState) Start(ctx context.Context) {
	time.Sleep(time.Until(v.deadline))

	deadline := time.Now().UTC().Add(config.ShowVotingScreenFor)
	votingState, err := v.subscriber.roundService.UpdateStateToVoting(ctx, v.gameStateID, deadline)
	if err != nil {
		v.subscriber.logger.Error(
			"failed to update game state to voting",
			slog.Any("error", err),
			slog.String("game_state_id", v.gameStateID.String()),
		)
		return
	}

	err = v.subscriber.updateClientsAboutVoting(ctx, votingState)
	if err != nil {
		v.subscriber.logger.Error(
			"failed to update clients to voting screen",
			slog.Any("error", err),
			slog.String("game_state_id", v.gameStateID.String()),
		)
		return
	}

	v.Update(ctx)
}

func (v *VotingState) Update(ctx context.Context) {
	fmt.Println("VotingState.Update", ctx)
}

type RevealState struct {
	subscriber Subscriber
	roundID    uuid.UUID
	deadline   time.Time
}

func (v *RevealState) Start(ctx context.Context) {
	fmt.Println("VotingState.Start", ctx)

	v.Update(ctx)
}

func (v *RevealState) Update(ctx context.Context) {
	fmt.Println("VotingState.Update", ctx)
}

type ScoringState struct {
	Deadline time.Time
}

func StartStateMachine(ctx context.Context, initialState State) {
	initialState.Start(ctx)
}
