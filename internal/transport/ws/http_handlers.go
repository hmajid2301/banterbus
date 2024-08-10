package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"go.opentelemetry.io/otel"

	"gitlab.com/hmajid2301/banterbus/internal/entities"
)

type RoomServicer interface {
	CreateRoom(ctx context.Context, roomCode string) (entities.Room, error)
}

type RoomRandomizer interface {
	GetRoomCode() string
}

type server struct {
	rooms          map[string]*room
	roomRandomizer RoomRandomizer
	eventHandlers  map[string]func(context.Context, *client, message) ([]byte, error)
	roomServicer   RoomServicer
	logger         *slog.Logger
	mux            http.ServeMux
}

func NewHTTPServer(roomServicer RoomServicer, roomRandomizer RoomRandomizer, logger *slog.Logger) *server {
	s := &server{
		rooms:          make(map[string]*room),
		roomServicer:   roomServicer,
		roomRandomizer: roomRandomizer,
		logger:         logger,
	}

	s.eventHandlers = map[string]func(context.Context, *client, message) ([]byte, error){
		"room_created": s.handleRoomCreatedEvent,
	}
	s.mux.Handle("/", http.FileServer(http.Dir("./static")))
	s.mux.HandleFunc("/ws", s.subscribeHandler)

	return s
}

func (s *server) Serve() error {
	s.logger.Info("starting server")
	err := http.ListenAndServe(":8080", &s.mux)
	if err != nil {
		return err
	}

	return nil
}

func (s *server) subscribeHandler(w http.ResponseWriter, r *http.Request) {
	s.logger.Debug("subscribe handler called")
	err := s.subscribe(r.Context(), r, w)
	if err != nil {
		s.logger.Error("subscribe failed", slog.Any("error", err))
		return
	}
}

func (s *server) subscribe(ctx context.Context, r *http.Request, w http.ResponseWriter) (err error) {
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

	// TODO: how to handle error?
	go func() {
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
	}()

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

			msg, err := handlerFunc(ctx, client, message)
			if err != nil {
				return err
			}
			client.messages <- msg
			span.End()
		}
	}
}
