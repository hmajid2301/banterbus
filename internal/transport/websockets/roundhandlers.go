package websockets

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"gitlab.com/hmajid2301/banterbus/internal/config"
	"gitlab.com/hmajid2301/banterbus/internal/service"
	sqlc "gitlab.com/hmajid2301/banterbus/internal/store/db"
)

type RoundServicer interface {
	SubmitAnswer(ctx context.Context, playerID, answer string, submittedAt time.Time) error
	SubmitVote(
		ctx context.Context,
		playerID string,
		votedNickname string,
		submittedAt time.Time,
	) (service.VotingState, error)
	UpdateStateToVoting(ctx context.Context, updateVoting service.UpdateVotingState) (service.VotingState, error)
	ToggleAnswerIsReady(ctx context.Context, playerID string) (bool, error)
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
	allReady, err := sub.roundService.ToggleAnswerIsReady(ctx, client.playerID)
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
	err = sub.updateClientsAboutQuestion(ctx, questionState)
	if err != nil {
		return err
	}

	if allReady {
		time.Sleep(config.AllReadyToNextScreenFor)
		go MoveToVoting(ctx, sub, questionState.Players, questionState.GameStateID, questionState.Round)
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
	// TODO: Maybe mixing businsess logic with transport logic here look at fixing this
	gameState, err := sub.playerService.GetGameState(ctx, gameStateID)
	if err != nil {
		sub.logger.Error("failed to get game state", slog.Any("error", err))
		return
	}

	if gameState != sqlc.GAMESTATE_FIBBING_IT_SHOW_QUESTION {
		sub.logger.WarnContext(ctx, "game state is not in FIBBING_IT_SHOW_QUESTION state")
		return
	}

	deadline := time.Now().Add(config.ShowVotingScreenFor)
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
