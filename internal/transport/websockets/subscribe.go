package websockets

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/google/uuid"
	"github.com/invopop/ctxi18n"
	"github.com/redis/go-redis/v9"
	slogctx "github.com/veqryn/slog-context"

	"gitlab.com/hmajid2301/banterbus/internal/entities"
	"gitlab.com/hmajid2301/banterbus/internal/logging"
	"gitlab.com/hmajid2301/banterbus/internal/store"
	"gitlab.com/hmajid2301/banterbus/internal/telemetry"
	"gitlab.com/hmajid2301/banterbus/internal/views/sections"
)

type Subscriber struct {
	lobbyService  LobbyServicer
	playerService PlayerServicer
	roundService  RoundServicer
	logger        *slog.Logger
	handlers      map[string]WSHandler
	websocket     Websocketer
}

type Websocketer interface {
	Subscribe(ctx context.Context, id string) <-chan *redis.Message
	Publish(ctx context.Context, id string, msg []byte) error
	Close(id string) error
}

type message struct {
	MessageType string `json:"message_type"`
}

type WSHandler interface {
	Handle(ctx context.Context, client *client, sub *Subscriber) error
	Validate() error
}

func NewSubscriber(
	lobbyService LobbyServicer,
	playerService PlayerServicer,
	roundService RoundServicer,
	logger *slog.Logger,
	websocket Websocketer,
) *Subscriber {
	s := &Subscriber{
		lobbyService:  lobbyService,
		playerService: playerService,
		roundService:  roundService,
		logger:        logger,
		websocket:     websocket,
	}

	s.handlers = map[string]WSHandler{
		"create_room":            &CreateRoom{},
		"update_player_nickname": &UpdateNickname{},
		"generate_new_avatar":    &GenerateNewAvatar{},
		"join_lobby":             &JoinLobby{},
		"toggle_player_is_ready": &TogglePlayerIsReady{},
		"kick_player":            &KickPlayer{},
		"start_game":             &StartGame{},
		"submit_answer":          &SubmitAnswer{},
	}

	return s
}

func (s *Subscriber) Subscribe(r *http.Request, w http.ResponseWriter) (err error) {
	ctx := context.Background()

	connection, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		return err
	}

	cookie, err := r.Cookie("player_id")
	if err != nil {
		return fmt.Errorf("failed to fetch `player_id` from cookie: %w", err)
	}

	playerID := cookie.Value
	subscribeCh := s.websocket.Subscribe(ctx, playerID)
	client := newClient(connection, playerID, subscribeCh)

	err = s.Reconnect(ctx, client)
	if err != nil {
		s.logger.WarnContext(ctx, "failed to reconnect", slog.Any("error", err))
	}

	defer func() {
		err = connection.Close()
	}()

	quit := make(chan struct{})
	go s.handleMessages(ctx, quit, client)

	writeTimeout := 10
	err = connection.SetWriteDeadline(time.Now().Add(time.Second * time.Duration(writeTimeout)))
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to set timeout", slog.Any("error", err))
		return err
	}

	for {
		select {
		// INFO: Send message to client.
		case msg := <-client.messagesCh:
			// TODO: only run when debug logging
			cleanedMessage := logging.StripSVGData(msg.Payload)
			s.logger.DebugContext(ctx, "sending message", slog.String("message", cleanedMessage))
			err = wsutil.WriteServerText(connection, []byte(msg.Payload))
			if err != nil {
				return err
			}
		case <-ctx.Done():
			s.logger.DebugContext(ctx, "context done")
			quit <- struct{}{}
			return ctx.Err()
		}
	}
}

func (s Subscriber) Reconnect(ctx context.Context, client *client) error {
	s.logger.DebugContext(ctx, "attempting to reconnect player", slog.String("player_id", client.playerID))
	roomState, err := s.playerService.GetRoomState(ctx, client.playerID)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	var component templ.Component
	switch roomState {
	case store.CREATED:
		lobby, err := s.playerService.GetLobby(ctx, client.playerID)
		if err != nil {
			errStr := "failed to reconnect to game"
			clientErr := s.updateClientAboutErr(ctx, client.playerID, errStr)
			return fmt.Errorf("%w: %w", err, clientErr)
		}

		var mePlayer entities.LobbyPlayer
		for _, player := range lobby.Players {
			if player.ID == client.playerID {
				mePlayer = player
			}
		}

		component = sections.Lobby(lobby.Code, lobby.Players, mePlayer)
	case store.PLAYING:
		gameState, err := s.playerService.GetGameState(ctx, client.playerID)
		if err != nil {
			errStr := "failed to reconnect to game"
			clientErr := s.updateClientAboutErr(ctx, client.playerID, errStr)
			return fmt.Errorf("%w: %w", err, clientErr)
		}

		component = sections.Game(gameState, gameState.Players[0])
	case store.PAUSED:
		return fmt.Errorf("cannot reconnect game to paused game, as this is not implemented")
	case store.ABANDONED:
		return fmt.Errorf("cannot reconnect game is abandoned")
	case store.FINISHED:
		return fmt.Errorf("cannot reconnect game is finished")
	default:
		return fmt.Errorf("unknown room state: %s", roomState)
	}

	err = component.Render(ctx, &buf)
	if err != nil {
		return err
	}

	err = s.websocket.Publish(ctx, client.playerID, buf.Bytes())
	return err
}

func (s *Subscriber) handleMessages(ctx context.Context, quit <-chan struct{}, client *client) {
	for {
		select {
		case <-quit:
			return
		default:
			err := s.handleMessage(ctx, client)
			if err != nil {
				s.logger.ErrorContext(ctx, "failed to handle message", slog.Any("error", err))
				err := telemetry.IncrementMessageReceivedError(ctx)
				if err != nil {
					s.logger.WarnContext(
						ctx,
						"failed to increment message received error metric",
						slog.Any("error", err),
					)
				}
			}
		}
	}
}

func (s *Subscriber) handleMessage(ctx context.Context, client *client) error {
	correlationID := uuid.NewString()
	ctx = slogctx.Append(ctx, "player_id", client.playerID)
	ctx = slogctx.Append(ctx, "correlation_id", correlationID)

	_, r, err := wsutil.NextReader(client.connection, ws.StateServerSide)
	if err != nil {
		if err == io.EOF {
			return nil
		}
		return fmt.Errorf("failed to get next message: %w", err)
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("failed to read message: %w", err)
	}

	var message message
	err = json.Unmarshal(data, &message)
	s.logger.DebugContext(ctx, "received message", slog.Any("message", message))
	if err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	err = telemetry.IncrementMessageReceived(ctx, message.MessageType)
	if err != nil {
		s.logger.Warn("failed to increment message type metric", slog.Any("error", err))
	}

	s.logger.DebugContext(ctx, fmt.Sprintf("handle `%s`", message.MessageType))
	handler, ok := s.handlers[message.MessageType]
	if !ok {
		return fmt.Errorf("handler not found for message type: %s", message.MessageType)
	}

	err = json.Unmarshal(data, &handler)
	s.logger.DebugContext(ctx, "trying to unmarshal handler message", slog.Any("message", message))
	if err != nil {
		return fmt.Errorf("failed to unmarshal for handler: %w", err)
	}

	err = handler.Validate()
	if err != nil {
		return fmt.Errorf("error validating handler message: %w", err)
	}

	// TODO: do we want to use accept-language header
	// TODO: default say en-US to en-GB?
	if client.locale != "" {
		ctx, err = ctxi18n.WithLocale(ctx, client.locale)
		if err != nil {
			s.logger.ErrorContext(
				ctx,
				"failed to set locale",
				slog.String("locale", client.locale),
				slog.Any("error", err),
			)
		}
	}

	err = handler.Handle(ctx, client, s)
	if err != nil {
		return fmt.Errorf("error in handler function: %w", err)
	}

	s.logger.DebugContext(ctx, "finished handling request")
	return nil
}
