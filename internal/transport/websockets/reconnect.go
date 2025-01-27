package websockets

import (
	"bytes"
	"context"
	"errors"

	"github.com/a-h/templ"
	"github.com/google/uuid"
	"github.com/mdobak/go-xerrors"
	slogctx "github.com/veqryn/slog-context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"gitlab.com/hmajid2301/banterbus/internal/service"
	"gitlab.com/hmajid2301/banterbus/internal/store/db"
	"gitlab.com/hmajid2301/banterbus/internal/views/sections"
)

const errStr = "failed to reconnect to game"

func (s Subscriber) Reconnect(ctx context.Context, playerID uuid.UUID) (bytes.Buffer, error) {
	tracer := otel.Tracer("")
	// TODO: share ctx attributes? use logs from otel
	ctx = slogctx.Append(ctx, "player_id", playerID)
	ctx, span := tracer.Start(
		ctx,
		"reconnect",
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(attribute.String("player_id", playerID.String())),
	)
	s.logger.DebugContext(ctx, "attempting to reconnect player")

	var buf bytes.Buffer
	roomState, err := s.lobbyService.GetRoomState(ctx, playerID)
	if err != nil {
		return buf, err
	}
	span.AddEvent("room_state", trace.WithAttributes(attribute.String("room_state", roomState.String())))

	buf, err = s.reconnectOnRoomState(ctx, roomState, playerID)
	if err != nil {
		return buf, err
	}

	span.AddEvent("trying_to_reconnect", trace.WithAttributes(attribute.Bool("reconnected", err == nil)))
	return buf, err
}

func (s Subscriber) reconnectOnRoomState(
	ctx context.Context,
	roomState db.RoomState,
	playerID uuid.UUID,
) (bytes.Buffer, error) {
	var buf bytes.Buffer
	var component templ.Component
	var err error

	switch roomState {
	case db.Created:
		lobby, err := s.lobbyService.GetLobby(ctx, playerID)
		if err != nil {
			clientErr := s.updateClientAboutErr(ctx, playerID, errStr)
			return buf, errors.Join(clientErr, err)
		}

		var mePlayer service.LobbyPlayer
		for _, player := range lobby.Players {
			if player.ID == playerID {
				mePlayer = player
			}
		}

		component = sections.Lobby(lobby.Code, lobby.Players, mePlayer, s.rules)
	case db.Playing:
		component, err = s.reconnectToPlayingGame(ctx, playerID)
		if err != nil {
			return buf, err
		}
	case db.Paused:
		return buf, xerrors.New("cannot reconnect game to paused game, as this is not implemented")
	case db.Abandoned:
		return buf, xerrors.New("cannot reconnect game is abandoned")
	case db.Finished:
		return buf, xerrors.New("cannot reconnect game is finished")
	default:
		return buf, xerrors.New("unknown room state: %s", roomState)
	}

	err = component.Render(ctx, &buf)
	if err != nil {
		return buf, err
	}
	return buf, nil
}

func (s Subscriber) reconnectToPlayingGame(ctx context.Context, playerID uuid.UUID) (templ.Component, error) {
	var component templ.Component
	gameState, err := s.roundService.GetGameState(ctx, playerID)
	if err != nil {
		clientErr := s.updateClientAboutErr(ctx, playerID, errStr)
		return component, errors.Join(clientErr, err)
	}

	switch gameState {
	case db.FibbingITQuestion:
		question, err := s.roundService.GetQuestionState(ctx, playerID)
		if err != nil {
			clientErr := s.updateClientAboutErr(ctx, playerID, errStr)
			return component, errors.Join(clientErr, err)
		}

		showRole := false
		component = sections.Question(question, question.Players[0], showRole)
	case db.FibbingItVoting:
		voting, err := s.roundService.GetVotingState(ctx, playerID)
		if err != nil {
			clientErr := s.updateClientAboutErr(ctx, playerID, errStr)
			return component, errors.Join(clientErr, err)
		}
		component = sections.Voting(voting, voting.Players[0])
	case db.FibbingItRevealRole:
		reveal, err := s.roundService.GetRevealState(ctx, playerID)
		if err != nil {
			clientErr := s.updateClientAboutErr(ctx, playerID, errStr)
			return component, errors.Join(clientErr, err)
		}
		component = sections.Reveal(reveal)
	case db.FibbingItScoring:
		scoring := service.Scoring{
			GuessedFibber:      s.config.Scoring.GuessFibber,
			FibberEvadeCapture: s.config.Scoring.FibberEvadeCapture,
		}

		score, err := s.roundService.GetScoreState(ctx, scoring, playerID)
		if err != nil {
			clientErr := s.updateClientAboutErr(ctx, playerID, errStr)
			return component, errors.Join(clientErr, err)
		}

		maxScore := 0
		for _, player := range score.Players {
			if player.Score > maxScore {
				maxScore = player.Score
			}
		}

		component = sections.Score(score, score.Players[0], maxScore)
	case db.FibbingItWinner:
		state, err := s.roundService.GetWinnerState(ctx, playerID)
		if err != nil {
			clientErr := s.updateClientAboutErr(ctx, playerID, errStr)
			return component, errors.Join(clientErr, err)
		}

		maxScore := 0
		for _, player := range state.Players {
			if player.Score > maxScore {
				maxScore = player.Score
			}
		}
		component = sections.Winner(state, maxScore)
	default:
		return component, xerrors.New("unknown game state: %s", gameState)
	}

	return component, nil
}
