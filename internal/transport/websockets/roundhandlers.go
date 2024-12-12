package websockets

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"gitlab.com/hmajid2301/banterbus/internal/config"
	"gitlab.com/hmajid2301/banterbus/internal/service"
)

type RoundServicer interface {
	SubmitAnswer(ctx context.Context, playerID uuid.UUID, answer string, submittedAt time.Time) error
	SubmitVote(
		ctx context.Context,
		playerID uuid.UUID,
		votedNickname string,
		submittedAt time.Time,
	) (service.VotingState, error)
	UpdateStateToVoting(ctx context.Context, gameStateID uuid.UUID, deadline time.Time) (service.VotingState, error)
	ToggleAnswerIsReady(ctx context.Context, playerID uuid.UUID, submittedAt time.Time) (bool, error)
	GetVotingState(ctx context.Context, playerID uuid.UUID) (service.VotingState, error)
}

func (s *SubmitAnswer) Handle(ctx context.Context, client *client, sub *Subscriber) error {
	err := sub.roundService.SubmitAnswer(ctx, client.playerID, s.Answer, time.Now().UTC())
	if err != nil {
		errStr := "failed to submit answer, try again"
		clientErr := sub.updateClientAboutErr(ctx, client.playerID, errStr)
		return errors.Join(clientErr, err)
	}

	t := Toast{Message: "Answer Submitted", Type: "success"}
	toastJSON, err := json.Marshal(t)
	if err != nil {
		return err
	}

	err = sub.websocket.Publish(ctx, client.playerID, toastJSON)
	return err
}

func (t *ToggleAnswerIsReady) Handle(ctx context.Context, client *client, sub *Subscriber) error {
	allReady, err := sub.roundService.ToggleAnswerIsReady(ctx, client.playerID, time.Now().UTC())
	if err != nil {
		errStr := "failed to toggle you are ready, try again"
		clientErr := sub.updateClientAboutErr(ctx, client.playerID, errStr)
		return errors.Join(clientErr, err)
	}

	questionState, err := sub.playerService.GetQuestionState(ctx, client.playerID)
	if err != nil {
		errStr := "failed to toggle you are ready, try again"
		clientErr := sub.updateClientAboutErr(ctx, client.playerID, errStr)
		return errors.Join(clientErr, err)
	}

	// INFO: Only need to update state of one player, so question state here should only contain a single player.
	showRole := false
	err = sub.updateClientsAboutQuestion(ctx, questionState, showRole)
	if err != nil {
		return err
	}

	if allReady {
		votingState := VotingState{
			deadline:   time.Now().UTC().Add(config.AllReadyToNextScreenFor),
			subscriber: *sub,
		}
		go StartStateMachine(ctx, &votingState)
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

	err = sub.updateClientsAboutVoting(ctx, votingState)
	if err != nil {
		sub.logger.Error("failed to update clients", slog.Any("error", err))
	}

	return nil
}
