package websockets

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/google/uuid"
	"github.com/invopop/ctxi18n"
	"github.com/redis/go-redis/v9"
	slogctx "github.com/veqryn/slog-context"

	"gitlab.com/hmajid2301/banterbus/internal/logging"
	"gitlab.com/hmajid2301/banterbus/internal/telemetry"
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
	Subscribe(ctx context.Context, id uuid.UUID) <-chan *redis.Message
	Publish(ctx context.Context, id uuid.UUID, msg []byte) error
	Close(id uuid.UUID) error
}

type message struct {
	MessageType string `json:"message_type"`
	Header      header `json:"HEADERS"`
}

type header struct {
	Locale string `json:"Accept-Language"`
}

type WSHandler interface {
	Handle(ctx context.Context, client *client, sub *Subscriber) error
	Validate() error
}

var errConnectionClosed = fmt.Errorf("connection closed")

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
		"toggle_answer_is_ready": &ToggleAnswerIsReady{},
	}

	return s
}

func (s *Subscriber) Subscribe(r *http.Request, w http.ResponseWriter) (err error) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	var playerID uuid.UUID
	var buf bytes.Buffer

	cookie, err := r.Cookie("player_id")
	if err != nil {
		cookie = getPlayerIDCookie()
		http.SetCookie(w, cookie)
	} else {
		playerID, err = uuid.Parse(cookie.Value)
		if err != nil {
			cancel()
			return err
		}

		buf, err = s.Reconnect(ctx, playerID)
		if err != nil {
			s.logger.WarnContext(ctx, "failed to reconnect", slog.Any("error", err))
			cookie = getPlayerIDCookie()
			http.SetCookie(w, cookie)
		}
	}

	playerID, err = uuid.Parse(cookie.Value)
	if err != nil {
		cancel()
		return err
	}

	h := ws.HTTPUpgrader{
		Header: w.Header(),
	}
	connection, _, _, err := h.Upgrade(r, w)
	if err != nil {
		cancel()
		return err
	}

	subscribeCh := s.websocket.Subscribe(ctx, playerID)
	client := newClient(connection, playerID, subscribeCh)

	// INFO: Send the reconnection message to the client.
	if buf.Len() > 0 {
		err = s.websocket.Publish(ctx, playerID, buf.Bytes())
	}

	defer func() {
		err = connection.Close()
	}()

	go s.handleMessages(ctx, cancel, client)

	// TODO: workout what to do with this?
	// writeTimeout := 10
	// err = connection.SetWriteDeadline(time.Now().Add(time.Second * time.Duration(writeTimeout)))
	// if err != nil {
	// 	s.logger.ErrorContext(ctx, "failed to set timeout", slog.Any("error", err))
	// 	return err
	// }

	for {
		select {
		// INFO: Send message to client.
		case msg := <-client.messagesCh:
			// TODO: only run when debug logging
			cleanedMessage := logging.StripSVGData(msg.Payload)
			s.logger.DebugContext(ctx, "sending message", slog.String("message", cleanedMessage))
			err = wsutil.WriteServerText(connection, []byte(msg.Payload))
			if err != nil {
				s.logger.ErrorContext(ctx, "failed to write message", slog.Any("error", err))
				// return err
			}
		case <-ctx.Done():
			s.logger.DebugContext(ctx, "context done")
			cancel()
			return ctx.Err()
		}
	}
}

func getPlayerIDCookie() *http.Cookie {
	playerID := uuid.Must(uuid.NewV7()).String()

	cookie := &http.Cookie{
		Name:     "player_id",
		Value:    playerID,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(time.Hour),
	}
	return cookie
}

func (s *Subscriber) handleMessages(ctx context.Context, cancel context.CancelFunc, client *client) {
	for {
		select {
		case <-ctx.Done():
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

				if errors.Is(err, errConnectionClosed) {
					cancel()
					return
				}
			}
		}
	}
}

func (s *Subscriber) handleMessage(ctx context.Context, client *client) error {
	correlationID := uuid.NewString()
	ctx = slogctx.Append(ctx, "player_id", client.playerID)
	ctx = slogctx.Append(ctx, "correlation_id", correlationID)

	hdr, r, err := wsutil.NextReader(client.connection, ws.StateServerSide)
	if err != nil {
		if err == io.EOF {
			return nil
		} else if opErr, ok := err.(*net.OpError); ok && opErr.Err.Error() == "use of closed network connection" {
			return nil
		}

		return fmt.Errorf("failed to get next message: %w", err)
	}

	if hdr.OpCode == ws.OpClose {
		return errConnectionClosed
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

	locale := message.Header.Locale
	if message.Header.Locale == "" {
		locale = "en-GB"
	}

	ctx, err = ctxi18n.WithLocale(ctx, locale)
	if err != nil {
		s.logger.ErrorContext(
			ctx,
			"failed to set locale",
			slog.String("locale", locale),
			slog.Any("error", err),
		)
	}

	err = handler.Handle(ctx, client, s)
	if err != nil {
		return fmt.Errorf("error in handler function: %w", err)
	}

	s.logger.DebugContext(ctx, "finished handling request")
	return nil
}
