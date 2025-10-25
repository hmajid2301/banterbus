package websockets

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/gofrs/uuid/v5"

	"gitlab.com/hmajid2301/banterbus/internal/service"
	"gitlab.com/hmajid2301/banterbus/internal/statemachine"
	"gitlab.com/hmajid2301/banterbus/internal/store/db"
	"gitlab.com/hmajid2301/banterbus/internal/telemetry"
)

type RoundServicer interface {
	GetGameState(ctx context.Context, playerID uuid.UUID) (db.FibbingItGameState, error)
	GetGameStateByID(ctx context.Context, gameStateID uuid.UUID) (db.FibbingItGameState, error)
	SubmitAnswer(ctx context.Context, playerID uuid.UUID, answer string, submittedAt time.Time) error
	SubmitVote(
		ctx context.Context,
		playerID uuid.UUID,
		votedNickname string,
		submittedAt time.Time,
	) (service.VotingState, error)
	UpdateStateToVoting(ctx context.Context, gameStateID uuid.UUID, deadline time.Time) (service.VotingState, error)
	ToggleAnswerIsReady(ctx context.Context, playerID uuid.UUID, submittedAt time.Time) (bool, error)
	AreAllPlayersAnswerReady(ctx context.Context, gameStateID uuid.UUID) (bool, error)
	GetVotingState(ctx context.Context, playerID uuid.UUID) (service.VotingState, error)
	ToggleVotingIsReady(ctx context.Context, playerID uuid.UUID, submittedAt time.Time) (bool, error)
	AreAllPlayersVotingReady(ctx context.Context, gameStateID uuid.UUID) (bool, error)
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
	PauseGame(ctx context.Context, playerID uuid.UUID) (service.PauseStatus, error)
	ResumeGame(ctx context.Context, playerID uuid.UUID) (service.PauseStatus, error)
	GetPauseStatus(ctx context.Context, gameStateID uuid.UUID) (service.PauseStatus, error)
	GetAllPlayersByGameStateID(ctx context.Context, gameStateID uuid.UUID) ([]db.GetAllPlayersByGameStateIDRow, error)
}

func (s *SubmitAnswer) Handle(ctx context.Context, client *Client, sub *Subscriber) error {
	telemetry.AddPlayerActionAttributes(ctx, client.playerID.String(), "submit_answer", false, false)
	telemetry.AddAnswerAttributes(
		ctx,
		client.playerID.String(),
		s.Answer,
		true,
		[]string{},
		time.Now().Format(time.RFC3339),
	)

	err := sub.roundService.SubmitAnswer(ctx, client.playerID, s.Answer, time.Now().UTC())
	if err != nil {
		telemetry.RecordBusinessLogicError(ctx, "submit_answer", err.Error(), telemetry.GameContext{
			PlayerID: &client.playerID,
		})
		errStr := "Failed to submit answer"
		clientErr := sub.updateClientAboutErr(ctx, client.playerID, errStr)
		return errors.Join(clientErr, err)
	}

	currentGameState, err := sub.roundService.GetGameState(ctx, client.playerID)
	if err != nil {
		sub.logger.ErrorContext(ctx, "failed to get game state after answer submission", slog.Any("error", err))
	} else if currentGameState != db.FibbingITQuestion {
		return nil
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

	currentGameState, err := sub.roundService.GetGameState(ctx, client.playerID)
	if err != nil {
		sub.logger.ErrorContext(ctx, "failed to get game state after toggle ready", slog.Any("error", err))
	} else if currentGameState == db.FibbingITQuestion {
		questionState, err := sub.roundService.GetQuestionState(ctx, client.playerID)
		if err != nil {
			errStr := "Failed to get updated question state."
			clientErr := sub.updateClientAboutErr(ctx, client.playerID, errStr)
			return errors.Join(clientErr, err)
		}

		// INFO: Only updating individual player's ready status, not full state refresh
		showRole := false
		err = sub.UpdateClientsAboutQuestion(ctx, questionState, showRole)
		if err != nil {
			return err
		}
	}

	if allReady {
		questionState, err := sub.roundService.GetQuestionState(ctx, client.playerID)
		if err != nil {
			return err
		}

		deps, err := sub.NewStateDependencies()
		if err != nil {
			sub.logger.ErrorContext(ctx, "failed to build state dependencies",
				slog.Any("error", err),
				slog.String("game_state_id", questionState.GameStateID.String()))
			clientErr := sub.updateClientAboutErr(ctx, client.playerID, "Failed to start voting state machine")
			return errors.Join(clientErr, err)
		}

		votingState, err := statemachine.NewVotingState(questionState.GameStateID, deps)
		if err != nil {
			sub.logger.ErrorContext(ctx, "failed to create voting state",
				slog.Any("error", err),
				slog.String("game_state_id", questionState.GameStateID.String()))
			clientErr := sub.updateClientAboutErr(ctx, client.playerID, "Failed to create voting state")
			return errors.Join(clientErr, err)
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

	err = sub.UpdateClientsAboutVoting(ctx, votingState)
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

	err = sub.UpdateClientsAboutVoting(ctx, votingState)
	if err != nil {
		return err
	}

	if allReady {
		deps, err := sub.NewStateDependencies()
		if err != nil {
			sub.logger.ErrorContext(ctx, "failed to build state dependencies",
				slog.Any("error", err),
				slog.String("game_state_id", votingState.GameStateID.String()))
			clientErr := sub.updateClientAboutErr(ctx, client.playerID, "Failed to start reveal state machine")
			return errors.Join(clientErr, err)
		}

		revealState, err := statemachine.NewRevealState(votingState.GameStateID, deps)
		if err != nil {
			sub.logger.ErrorContext(ctx, "failed to create reveal state",
				slog.Any("error", err),
				slog.String("game_state_id", votingState.GameStateID.String()))
			clientErr := sub.updateClientAboutErr(ctx, client.playerID, "Failed to create reveal state")
			return errors.Join(clientErr, err)
		}
		go revealState.Start(ctx)
	}

	return nil
}

func (p *PauseGame) Handle(ctx context.Context, client *Client, sub *Subscriber) error {
	telemetry.AddPlayerActionAttributes(ctx, client.playerID.String(), "pause_game", true, false)

	pauseStatus, err := sub.roundService.PauseGame(ctx, client.playerID)
	if err != nil {
		errStr := "Failed to pause game"
		if errors.Is(err, service.ErrNotHost) {
			errStr = "Only the host can pause the game"
		} else if errors.Is(err, service.ErrNoPauseTimeRemaining) {
			errStr = "No pause time remaining (5 minute limit reached)"
		} else if errors.Is(err, service.ErrGameAlreadyPaused) {
			errStr = "Game is already paused"
		} else if errors.Is(err, service.ErrGameNotStarted) {
			errStr = "Cannot pause game that has not started"
		}
		clientErr := sub.updateClientAboutErr(ctx, client.playerID, errStr)
		return errors.Join(clientErr, err)
	}

	questionState, err := sub.roundService.GetQuestionState(ctx, client.playerID)
	if err == nil {
		sub.stopStateMachine(ctx, questionState.GameStateID)
		sub.updateClientsAboutPause(ctx, pauseStatus, questionState.GameStateID)
		return nil
	}

	votingState, err := sub.roundService.GetVotingState(ctx, client.playerID)
	if err == nil {
		sub.stopStateMachine(ctx, votingState.GameStateID)
		sub.updateClientsAboutPause(ctx, pauseStatus, votingState.GameStateID)
		return nil
	}

	return sub.updateClientAboutErr(ctx, client.playerID, "Failed to pause game - unknown game state")
}

func (r *ResumeGame) Handle(ctx context.Context, client *Client, sub *Subscriber) error {
	telemetry.AddPlayerActionAttributes(ctx, client.playerID.String(), "resume_game", true, false)

	pauseStatus, err := sub.roundService.ResumeGame(ctx, client.playerID)
	if err != nil {
		errStr := "Failed to resume game"
		if errors.Is(err, service.ErrNotHost) {
			errStr = "Only the host can resume the game"
		} else if errors.Is(err, service.ErrGameNotPaused) {
			errStr = "Game is not paused"
		} else if errors.Is(err, service.ErrGameNotStarted) {
			errStr = "Cannot resume game that has not started"
		}
		clientErr := sub.updateClientAboutErr(ctx, client.playerID, errStr)
		return errors.Join(clientErr, err)
	}

	questionState, err := sub.roundService.GetQuestionState(ctx, client.playerID)
	if err == nil {
		sub.updateClientsAboutResume(ctx, pauseStatus, questionState.GameStateID)
		sub.restartStateMachineAfterResume(ctx, questionState.GameStateID, pauseStatus.State, pauseStatus.SubmitDeadline)
		return nil
	}

	votingState, err := sub.roundService.GetVotingState(ctx, client.playerID)
	if err == nil {
		sub.updateClientsAboutResume(ctx, pauseStatus, votingState.GameStateID)
		sub.restartStateMachineAfterResume(ctx, votingState.GameStateID, pauseStatus.State, pauseStatus.SubmitDeadline)
		return nil
	}

	return sub.updateClientAboutErr(ctx, client.playerID, "Failed to resume game - unknown game state")
}
