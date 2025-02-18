package websockets

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"gitlab.com/hmajid2301/banterbus/internal/service"
	"gitlab.com/hmajid2301/banterbus/internal/store/db"
)

type RoundServicer interface {
	GetGameState(ctx context.Context, playerID uuid.UUID) (db.FibbingItGameState, error)
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
	ToggleVotingIsReady(ctx context.Context, playerID uuid.UUID, submittedAt time.Time) (bool, error)
	UpdateStateToReveal(ctx context.Context, gameStateID uuid.UUID, deadline time.Time) (service.RevealRoleState, error)
	GetRevealState(ctx context.Context, playerID uuid.UUID) (service.RevealRoleState, error)
	UpdateStateToScore(
		ctx context.Context,
		gameStateID uuid.UUID,
		deadline time.Time,
		scoring service.Scoring,
	) (service.ScoreState, error)
	GetScoreState(ctx context.Context, scoring service.Scoring, playerID uuid.UUID) (service.ScoreState, error)
	UpdateStateToQuestion(
		ctx context.Context,
		gameStateID uuid.UUID,
		deadline time.Time,
		nextRound bool,
	) (service.QuestionState, error)
	GetQuestionState(ctx context.Context, playerID uuid.UUID) (service.QuestionState, error)
	UpdateStateToWinner(ctx context.Context, gameStateID uuid.UUID, deadline time.Time) (service.WinnerState, error)
	GetWinnerState(ctx context.Context, playerID uuid.UUID) (service.WinnerState, error)
	FinishGame(ctx context.Context, gameStateID uuid.UUID) error
}

func (s *SubmitAnswer) Handle(ctx context.Context, client *Client, sub *Subscriber) error {
	err := sub.roundService.SubmitAnswer(ctx, client.playerID, s.Answer, time.Now().UTC())
	if err != nil {
		errStr := "Failed to submit answer"
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

func (t *ToggleAnswerIsReady) Handle(ctx context.Context, client *Client, sub *Subscriber) error {
	allReady, err := sub.roundService.ToggleAnswerIsReady(ctx, client.playerID, time.Now().UTC())
	if err != nil {
		errStr := "Failed to toggle you are ready."
		clientErr := sub.updateClientAboutErr(ctx, client.playerID, errStr)
		return errors.Join(clientErr, err)
	}

	questionState, err := sub.roundService.GetQuestionState(ctx, client.playerID)
	if err != nil {
		errStr := "Failed to toggle you are ready."
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
			GameStateID: questionState.GameStateID,
			Subscriber:  *sub,
		}
		go votingState.Start(ctx)
	}

	return nil
}

func (s *SubmitVote) Handle(ctx context.Context, client *Client, sub *Subscriber) error {
	votingState, err := sub.roundService.SubmitVote(ctx, client.playerID, s.VotedPlayerNickname, time.Now())
	if err != nil {
		errStr := "Failed to submit vote."
		clientErr := sub.updateClientAboutErr(ctx, client.playerID, errStr)
		return errors.Join(clientErr, err)
	}

	err = sub.updateClientsAboutVoting(ctx, votingState)
	if err != nil {
		sub.logger.ErrorContext(ctx, "failed to update clients", slog.Any("error", err))
	}

	t := Toast{Message: "Vote Submitted", Type: "success"}
	toastJSON, err := json.Marshal(t)
	if err != nil {
		return err
	}

	err = sub.websocket.Publish(ctx, client.playerID, toastJSON)
	return err
}

func (t *ToggleVotingIsReady) Handle(ctx context.Context, client *Client, sub *Subscriber) error {
	allReady, err := sub.roundService.ToggleVotingIsReady(ctx, client.playerID, time.Now().UTC())
	if err != nil {
		errStr := "Failed to toggle voting you are ready."
		if errors.Is(err, service.ErrMustSubmitVote) {
			errStr = "Must submit vote before readying up."
		}
		clientErr := sub.updateClientAboutErr(ctx, client.playerID, errStr)
		return errors.Join(clientErr, err)
	}

	votingState, err := sub.roundService.GetVotingState(ctx, client.playerID)
	if err != nil {
		errStr := "Failed to toggle voting you are ready."
		clientErr := sub.updateClientAboutErr(ctx, client.playerID, errStr)
		return errors.Join(clientErr, err)
	}

	err = sub.updateClientsAboutVoting(ctx, votingState)
	if err != nil {
		return err
	}

	if allReady {
		revealState := RevealState{
			GameStateID: votingState.GameStateID,
			Subscriber:  *sub,
		}
		go revealState.Start(ctx)
	}

	return nil
}
