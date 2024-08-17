package ws

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

type subscriber struct {
	rooms          map[string]*room
	roomServicer   RoomServicer
	playerServicer PlayerServicer
	logger         *slog.Logger
	eventHandlers  map[string]func(context.Context, *client, message) error
}

func NewSubscriber(roomServicer RoomServicer, playerServicer PlayerServicer, logger *slog.Logger) *subscriber {
	s := &subscriber{
		roomServicer:   roomServicer,
		playerServicer: playerServicer,
		logger:         logger,
	}

	s.eventHandlers = map[string]func(context.Context, *client, message) error{
		"create_room":            s.handleCreateRoomEvent,
		"update_player_nickname": s.handleUpdateNicknameEvent,
		"generate_new_avatar":    s.handleGenerateNewAvatarEvent,
		"join_room":              s.handleJoinRoomEvent,
	}

	return s
}

func (s *subscriber) Subscribe(ctx context.Context, r *http.Request, w http.ResponseWriter) (err error) {
	connection, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		return err
	}

	client := NewClient(connection)

	defer func() {
		err = connection.Close()
	}()

	tracer := otel.Tracer("subscribe")
	incomingMessages := make(chan message)
	quit := make(chan struct{})

	go s.handleMessage(quit, connection, incomingMessages)

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
		case message := <-incomingMessages:
			ctx, span := tracer.Start(ctx, message.EventName)
			s.logger.DebugContext(ctx, fmt.Sprintf("handle `%s` event", message.EventName))
			handlerFunc, ok := s.eventHandlers[message.EventName]
			if !ok {
				return fmt.Errorf("no handler for event %s", message.EventName)
			}

			err := handlerFunc(ctx, client, message)
			if err != nil {
				return err
			}
			span.End()
		}
	}
}

func (s *subscriber) handleMessage(quit chan struct{}, connection net.Conn, incomingMessages chan message) {
	// TODO: how to handle error?
	for {
		select {
		case <-quit:
			return
		default:
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
			incomingMessages <- message
		}
	}
}
