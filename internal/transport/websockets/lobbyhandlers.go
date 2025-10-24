package websockets

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/gofrs/uuid/v5"

	"gitlab.com/hmajid2301/banterbus/internal/service"
	"gitlab.com/hmajid2301/banterbus/internal/statemachine"
	"gitlab.com/hmajid2301/banterbus/internal/store/db"
	"gitlab.com/hmajid2301/banterbus/internal/telemetry"
)

type LobbyServicer interface {
	Create(ctx context.Context, gameName string, player service.NewHostPlayer) (service.LobbyCreationResult, error)
	Join(
		ctx context.Context,
		roomCode string,
		playerID uuid.UUID,
		playerNickname string,
	) (service.LobbyJoinResult, error)
	Start(ctx context.Context, roomCode string, playerID uuid.UUID, deadline time.Time) (service.QuestionState, error)
	KickPlayer(
		ctx context.Context,
		roomCode string,
		playerID uuid.UUID,
		playerNicknameToKick string,
	) (service.Lobby, uuid.UUID, error)
	HandlePlayerDisconnect(ctx context.Context, playerID uuid.UUID) error
	GetLobby(ctx context.Context, playerID uuid.UUID) (service.Lobby, error)
	GetRoomState(ctx context.Context, playerID uuid.UUID) (db.RoomState, error)
}

func (c *CreateRoom) Handle(ctx context.Context, client *Client, sub *Subscriber) error {
	telemetry.AddGameContextToSpan(ctx, telemetry.GameContext{
		PlayerID: &client.playerID,
	})

	telemetry.AddPlayerActionAttributes(ctx, client.playerID.String(), "create_room", true, false)

	newPlayer := service.NewHostPlayer{
		ID:       client.playerID,
		Nickname: c.PlayerNickname,
	}
	result, err := sub.lobbyService.Create(ctx, c.GameName, newPlayer)
	if err != nil {
		telemetry.RecordBusinessLogicError(ctx, "create_room", err.Error(), telemetry.GameContext{
			PlayerID: &client.playerID,
		})
		errStr := "Failed to create room"
		clientErr := sub.updateClientAboutErr(ctx, client.playerID, errStr)
		return errors.Join(clientErr, err)
	}

	client.playerID = result.NewPlayerID

	telemetry.AddGameContextToSpan(ctx, telemetry.GameContext{
		PlayerID: &result.NewPlayerID,
		RoomCode: result.Lobby.Code,
	})

	telemetry.AddRoomStateAttributes(ctx, "Created", result.Lobby.Code, len(result.Lobby.Players))
	telemetry.AddGameStateTransition(ctx, "", "lobby_created", "host_action", nil)

	readyPlayers := 0
	for _, player := range result.Lobby.Players {
		if player.IsReady {
			readyPlayers++
		}
	}
	telemetry.AddLobbyStatusAttributes(
		ctx,
		result.Lobby.Code,
		len(result.Lobby.Players),
		readyPlayers,
		readyPlayers == len(result.Lobby.Players),
		false,
	)

	err = sub.updateClientsAboutLobby(ctx, result.Lobby)
	return err
}

func (j *JoinLobby) Handle(ctx context.Context, client *Client, sub *Subscriber) error {
	telemetry.AddGameContextToSpan(ctx, telemetry.GameContext{
		PlayerID: &client.playerID,
		RoomCode: j.RoomCode,
	})

	telemetry.AddPlayerActionAttributes(ctx, client.playerID.String(), "join_lobby", false, false)

	var err error
	var component bytes.Buffer
	result, err := sub.lobbyService.Join(ctx, j.RoomCode, client.playerID, j.PlayerNickname)
	if err != nil {
		if errors.Is(err, service.ErrPlayerAlreadyInRoom) {
			telemetry.AddPlayerConnectionAttributes(ctx, client.playerID.String(), "websocket", true, "")
			component, err = sub.Reconnect(ctx, client.playerID)
			if err == nil {
				webSocketErr := sub.websocket.Publish(ctx, client.playerID, component.Bytes())
				err = errors.Join(err, webSocketErr)
			} else {
				// Show the specific reconnection error to the user
				clientErr := sub.updateClientAboutErr(ctx, client.playerID, err.Error())
				return errors.Join(clientErr, err)
			}
		} else {
			telemetry.RecordBusinessLogicError(ctx, "join_lobby", err.Error(), telemetry.GameContext{
				PlayerID: &client.playerID,
				RoomCode: j.RoomCode,
			})
			errStr := "Failed to join room"
			if err == service.ErrNicknameExists {
				errStr = err.Error()
			}
			clientErr := sub.updateClientAboutErr(ctx, client.playerID, errStr)
			return errors.Join(clientErr, err)
		}
	}

	client.playerID = result.NewPlayerID

	sub.logger.InfoContext(ctx, "player successfully joined lobby",
		slog.String("player_id", result.NewPlayerID.String()),
		slog.String("nickname", j.PlayerNickname),
		slog.String("room_code", result.Lobby.Code),
		slog.Int("total_players", len(result.Lobby.Players)),
	)

	telemetry.AddRoomStateAttributes(ctx, "Created", result.Lobby.Code, len(result.Lobby.Players))
	telemetry.AddGameStateTransition(ctx, "", "player_joined", "player_action", nil)

	clientErr := sub.updateClientsAboutLobby(ctx, result.Lobby)
	return errors.Join(clientErr, err)
}

func (s *StartGame) Handle(ctx context.Context, client *Client, sub *Subscriber) error {
	telemetry.AddGameContextToSpan(ctx, telemetry.GameContext{
		PlayerID: &client.playerID,
		RoomCode: s.RoomCode,
	})

	telemetry.AddPlayerActionAttributes(ctx, client.playerID.String(), "start_game", true, true)

	deadline := time.Now().UTC().Add(sub.config.Timings.ShowQuestionScreenFor)
	telemetry.AddTimingAttributes(ctx, "question_deadline",
		sub.config.Timings.ShowQuestionScreenFor.String(),
		sub.config.Timings.ShowQuestionScreenFor.String(), false)

	questionState, err := sub.lobbyService.Start(ctx, s.RoomCode, client.playerID, deadline)
	if err != nil {
		telemetry.RecordBusinessLogicError(ctx, "start_game", err.Error(), telemetry.GameContext{
			PlayerID: &client.playerID,
			RoomCode: s.RoomCode,
		})
		errStr := "Failed to start game"
		if errors.Is(err, service.ErrNotEnoughPlayers) {
			errStr = fmt.Sprintf("At least %d players are needed to start the game", service.MinimumPlayers)
		}
		clientErr := sub.updateClientAboutErr(ctx, client.playerID, errStr)
		return errors.Join(clientErr, err)
	}

	telemetry.AddGameStateTransition(ctx, "Created", "Playing", "host_start", &questionState.GameStateID)
	telemetry.AddQuestionAttributes(ctx, questionState.GameStateID.String(), "",
		questionState.RoundType, questionState.Round, deadline.Format(time.RFC3339))

	showRole := true
	err = sub.UpdateClientsAboutQuestion(ctx, questionState, showRole)
	if err != nil {
		return err
	}

	stateCtx := telemetry.PropagateContext(ctx)

	deps, err := sub.NewStateDependencies()
	if err != nil {
		sub.logger.ErrorContext(ctx, "failed to build state dependencies",
			slog.Any("error", err),
			slog.String("game_state_id", questionState.GameStateID.String()))
		clientErr := sub.updateClientAboutErr(ctx, client.playerID, "Failed to start game state machine")
		return errors.Join(clientErr, err)
	}

	q, err := statemachine.NewQuestionState(questionState.GameStateID, false, deps)
	if err != nil {
		sub.logger.ErrorContext(ctx, "failed to create question state",
			slog.Any("error", err),
			slog.String("game_state_id", questionState.GameStateID.String()))
		clientErr := sub.updateClientAboutErr(ctx, client.playerID, "Failed to create question state")
		return errors.Join(clientErr, err)
	}
	sub.StartStateMachine(stateCtx, questionState.GameStateID, q)
	return nil
}

func (k *KickPlayer) Handle(ctx context.Context, client *Client, sub *Subscriber) error {
	updatedRoom, playerToKickID, err := sub.lobbyService.KickPlayer(
		ctx,
		k.RoomCode,
		client.playerID,
		k.PlayerNicknameToKick,
	)
	if err != nil {
		errStr := "Faled to kick player"
		clientErr := sub.updateClientAboutErr(ctx, client.playerID, errStr)
		return errors.Join(clientErr, err)
	}

	err = sub.updateClientsAboutLobby(ctx, updatedRoom)
	if err != nil {
		return fmt.Errorf("failed to send kick error message to player: %w", err)
	}

	// TODO: take user back to home page instead of just an error
	err = sub.updateClientAboutErr(ctx, playerToKickID, "you have been kicked from the room")
	return err
}
