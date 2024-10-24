package websockets

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/google/uuid"
	slogctx "github.com/veqryn/slog-context"

	"gitlab.com/hmajid2301/banterbus/internal/logging"
	"gitlab.com/hmajid2301/banterbus/internal/telemetry"
)

type Subscriber struct {
	rooms         map[string]*room
	lobbyService  LobbyServicer
	playerService PlayerServicer
	logger        *slog.Logger
	handlers      map[string]WSHandler
}

type message struct {
	MessageType string `json:"message_type"`
}

type WSHandler interface {
	Handle(ctx context.Context, client *client, sub *Subscriber) error
	Validate() error
}

func NewSubscriber(lobbyService LobbyServicer, playerService PlayerServicer, logger *slog.Logger) *Subscriber {
	s := &Subscriber{
		lobbyService:  lobbyService,
		playerService: playerService,
		logger:        logger,
		rooms:         make(map[string]*room),
	}

	s.handlers = map[string]WSHandler{
		"create_room":            &CreateRoom{},
		"update_player_nickname": &UpdateNickname{},
		"generate_new_avatar":    &GenerateNewAvatar{},
		"join_lobby":             &JoinLobby{},
		"toggle_player_is_ready": &TogglePlayerIsReady{},
		"start_game":             &StartGame{},
	}

	return s
}

func (s *Subscriber) Subscribe(r *http.Request, w http.ResponseWriter) (err error) {
	ctx := context.Background()
	connection, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		return err
	}

	client := NewClient(connection)

	defer func() {
		err = connection.Close()
	}()

	quit := make(chan struct{})
	go s.handleMessages(ctx, quit, connection, client)

	for {
		select {
		case msg := <-client.messages:
			cleanedMessage := logging.StripSVGData(string(msg))
			s.logger.DebugContext(ctx, "sending message", slog.String("message", cleanedMessage))

			err := connection.SetWriteDeadline(time.Now().Add(time.Second * 10))
			if err != nil {
				s.logger.ErrorContext(ctx, "failed to set timeout", slog.Any("error", err))
			}
			err = wsutil.WriteServerText(connection, msg)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			quit <- struct{}{}
			s.logger.DebugContext(ctx, "context done")
			return ctx.Err()
		}
	}
}

func (s *Subscriber) handleMessages(ctx context.Context, quit <-chan struct{}, connection net.Conn, client *client) {
	for {
		select {
		case <-quit:
			return
		default:
			err := s.handleMessage(ctx, client, connection)
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

func (s *Subscriber) handleMessage(ctx context.Context, client *client, connection net.Conn) error {
	correlationID := uuid.NewString()
	ctx = slogctx.Append(ctx, "player_id", client.playerID)
	ctx = slogctx.Append(ctx, "correlation_id", correlationID)

	_, r, err := wsutil.NextReader(connection, ws.StateServerSide)
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

	err = handler.Handle(ctx, client, s)
	if err != nil {
		return fmt.Errorf("error in handler function: %w", err)
	}

	s.logger.DebugContext(ctx, "finished handling request")
	return nil
}
