package websockets

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"regexp"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/google/uuid"
	slogctx "github.com/veqryn/slog-context"
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

func (s *Subscriber) Subscribe(ctx context.Context, r *http.Request, w http.ResponseWriter) (err error) {
	connection, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		return err
	}

	client := NewClient(connection)

	defer func() {
		err = connection.Close()
	}()

	quit := make(chan struct{})
	go s.handleMessage(ctx, quit, connection, client)

	for {
		select {
		case msg := <-client.messages:
			cleanedMessage := stripSVGData(string(msg))
			fmt.Println(ctx.Value("player_id"))
			s.logger.DebugContext(ctx, "sending message", slog.String("message", cleanedMessage))

			err := connection.SetWriteDeadline(time.Now().Add(time.Second * 10))
			if err != nil {
				s.logger.Error("failed to set timeout", slog.Any("error", err))
			}
			err = wsutil.WriteServerText(connection, msg)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			quit <- struct{}{}
			return ctx.Err()
		}
	}
}

func (s *Subscriber) handleMessage(ctx context.Context, quit <-chan struct{}, connection net.Conn, client *client) {
	for {
		select {
		case <-quit:
			return
		default:
			correlation_id := uuid.NewString()
			ctx = slogctx.Append(ctx, "player_id", client.playerID)
			ctx = slogctx.Append(ctx, "correlation_id", correlation_id)

			_, r, err := wsutil.NextReader(connection, ws.StateServerSide)
			if err != nil {
				s.logger.Error("failed to get next message", slog.Any("error", err))
				return
			}

			data, err := io.ReadAll(r)
			if err != nil {
				s.logger.Error("failed to read message", slog.Any("error", err))
				return
			}

			var message message
			err = json.Unmarshal(data, &message)
			s.logger.DebugContext(ctx, "received message", slog.Any("message", message))
			if err != nil {
				s.logger.Error("failed to unmarshal message", slog.Any("error", err))
				return
			}

			s.logger.DebugContext(ctx, fmt.Sprintf("handle `%s`", message.MessageType))
			handler, ok := s.handlers[message.MessageType]
			if !ok {
				s.logger.ErrorContext(ctx, "handler not found for message type", slog.Any("error", err))
				return
			}

			err = json.Unmarshal(data, &handler)
			s.logger.DebugContext(ctx, "trying to unmarshal handler message", slog.Any("message", message))
			if err != nil {
				s.logger.ErrorContext(ctx, "failed to unmarshal for handler", slog.Any("error", err))
				return
			}

			err = handler.Validate()
			if err != nil {
				s.logger.ErrorContext(ctx, "error validating handler message", slog.Any("error", err))
				return
			}

			err = handler.Handle(ctx, client, s)
			if err != nil {
				s.logger.ErrorContext(ctx, "error in handler function", slog.Any("error", err))
				return
			}
			s.logger.DebugContext(ctx, "finished handling request")
		}
	}
}

func stripSVGData(message string) string {
	re := regexp.MustCompile(`data:image/svg\+xml;base64,[A-Za-z0-9+/=]+`)
	return re.ReplaceAllString(message, "")
}
