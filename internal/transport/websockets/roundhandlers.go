package websockets

import (
	"context"
	"fmt"
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
	) ([]service.VotingPlayer, error)
	UpdateStateToVoting(
		ctx context.Context,
		players []service.PlayerWithRole,
		gameStateID string,
		deadline time.Time,
	) ([]service.VotingPlayer, error)
}

func (s *SubmitAnswer) Handle(ctx context.Context, client *client, sub *Subscriber) error {
	err := sub.roundService.SubmitAnswer(ctx, client.playerID, s.Answer, time.Now())
	if err != nil {
		errStr := "failed to submit answer, try again"
		clientErr := sub.updateClientAboutErr(ctx, client.playerID, errStr)
		return fmt.Errorf("%w: %w", err, clientErr)
	}

	return nil
}

func MoveToVoting(ctx context.Context, sub *Subscriber, players []service.PlayerWithRole, gameStateID string) {
	deadline := time.Now().Add(VotingScreenDelay)
	votingPlayers, err := sub.roundService.UpdateStateToVoting(ctx, players, gameStateID, deadline)
	if err != nil {
		sub.logger.Error("failed to update game state", slog.Any("error", err))
	}

	err = sub.updateClientAboutVoting(ctx, votingPlayers)
	if err != nil {
		sub.logger.Error("failed to move to voting", slog.Any("error", err))
	}
}

func (s *SubmitVote) Handle(ctx context.Context, client *client, sub *Subscriber) error {
	votes, err := sub.roundService.SubmitVote(ctx, client.playerID, s.VotedPlayerNickname, time.Now())
	if err != nil {
		errStr := "failed to submit vote, try again"
		clientErr := sub.updateClientAboutErr(ctx, client.playerID, errStr)
		return fmt.Errorf("%w: %w", err, clientErr)
	}

	err = sub.updateClientAboutVoting(ctx, votes)
	if err != nil {
		sub.logger.Error("failed to update clients", slog.Any("error", err))
	}

	return nil
}
