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
	"go.opentelemetry.io/otel"
)

type Subscriber struct {
	rooms          map[string]*room
	roomServicer   RoomServicer
	playerServicer PlayerServicer
	logger         *slog.Logger
	handlers       map[string]WSHandler
}

type message struct {
	MessageType string `json:"message_type"`
}

type WSHandler interface {
	Handle(ctx context.Context, client *client, sub *Subscriber) error
}

func NewSubscriber(
	roomServicer RoomServicer,
	playerServicer PlayerServicer,
	logger *slog.Logger,
) *Subscriber {
	s := &Subscriber{
		roomServicer:   roomServicer,
		playerServicer: playerServicer,
		logger:         logger,
		rooms:          make(map[string]*room),
	}

	s.handlers = map[string]WSHandler{
		"create_room":            &CreateRoom{},
		"update_player_nickname": &UpdateNickname{},
		"generate_new_avatar":    &GenerateNewAvatar{},
		"join_room":              &JoinRoom{},
	}

	return s
}

func (s *Subscriber) Subscribe(
	ctx context.Context,
	r *http.Request,
	w http.ResponseWriter,
) (err error) {
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
			s.logger.Debug("sending message", slog.String("message", string(msg)))
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

func (s *Subscriber) handleMessage(
	ctx context.Context,
	quit <-chan struct{},
	connection net.Conn,
	client *client,
) {
	// TODO: how to handle error?
	for {
		select {
		case <-quit:
			return
		default:
			tracer := otel.Tracer("subscribe")
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
			s.logger.Debug("received message", slog.Any("message", message))
			if err != nil {
				s.logger.Error("failed to unmarshal message", slog.Any("error", err))
				return
			}
			ctx, span := tracer.Start(ctx, message.MessageType)
			s.logger.DebugContext(ctx, fmt.Sprintf("handle `%s`", message.MessageType))
			handler, ok := s.handlers[message.MessageType]
			if !ok {
				s.logger.Error("handler not found for message type", slog.Any("error", err))
				return
			}

			err = json.Unmarshal(data, &handler)
			s.logger.Debug("trying to unmarshal handler message", slog.Any("message", message))
			if err != nil {
				s.logger.Error("failed to unmarshal for handler", slog.Any("error", err))
				return
			}

			err = handler.Handle(ctx, client, s)
			if err != nil {
				s.logger.Error("error in handler function", slog.Any("error", err))
				return
			}
			span.End()
		}
	}
}
