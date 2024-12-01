package websockets

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"gitlab.com/hmajid2301/banterbus/internal/service"
)

// TODO: Move this to a config file
const VotingScreenDelay = 60 * time.Second

type RoundServicer interface {
	SubmitAnswer(ctx context.Context, playerID, answer string, submittedAt time.Time) error
	SubmitVote(
		ctx context.Context,
		playerID string,
		votedNickname string,
		submittedAt time.Time,
	) (service.VotingState, error)
	UpdateStateToVoting(ctx context.Context, updateVoting service.UpdateVotingState) (service.VotingState, error)
}

func (s *SubmitAnswer) Handle(ctx context.Context, client *client, sub *Subscriber) error {
	err := sub.roundService.SubmitAnswer(ctx, client.playerID, s.Answer, time.Now())
	if err != nil {
		errStr := "failed to submit answer, try again"
		clientErr := sub.updateClientAboutErr(ctx, client.playerID, errStr)
		return errors.Join(clientErr, err)
	}

	return nil
}

func (s *SubmitVote) Handle(ctx context.Context, client *client, sub *Subscriber) error {
	votingState, err := sub.roundService.SubmitVote(ctx, client.playerID, s.VotedPlayerNickname, time.Now())
	if err != nil {
		errStr := "failed to submit vote, try again"
		clientErr := sub.updateClientAboutErr(ctx, client.playerID, errStr)
		return errors.Join(clientErr, err)
	}

	err = sub.updateClientAboutVoting(ctx, votingState)
	if err != nil {
		sub.logger.Error("failed to update clients", slog.Any("error", err))
	}

	return nil
}

func MoveToVoting(
	ctx context.Context,
	sub *Subscriber,
	players []service.PlayerWithRole,
	gameStateID string,
	round int,
) {
	deadline := time.Now().Add(VotingScreenDelay)
	updateState := service.UpdateVotingState{
		GameStateID: gameStateID,
		Players:     players,
		Deadline:    deadline,
		Round:       round,
	}

	votingState, err := sub.roundService.UpdateStateToVoting(ctx, updateState)
	if err != nil {
		sub.logger.Error("failed to update game state", slog.Any("error", err))
	}

	err = sub.updateClientAboutVoting(ctx, votingState)
	if err != nil {
		sub.logger.Error("failed to move to voting", slog.Any("error", err))
	}
}
