package ws

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"

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
	mux            http.ServeMux
}

func NewHTTPServer(roomServicer RoomServicer, roomRandomizer RoomRandomizer) *server {
	s := &server{
		rooms:          make(map[string]*room),
		roomServicer:   roomServicer,
		roomRandomizer: roomRandomizer,
	}

	s.eventHandlers = map[string]func(context.Context, *client, message) ([]byte, error){
		"room_created": s.handleRoomCreatedEvent,
	}
	s.mux.Handle("/", http.FileServer(http.Dir("./static")))
	s.mux.HandleFunc("/ws", s.subscribeHandler)

	return s
}

func (s *server) Serve() error {
	log.Println("Starting server")
	err := http.ListenAndServe(":8080", &s.mux)
	if err != nil {
		return err
	}

	return nil
}

func (s *server) subscribeHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("subscribeHandler")
	err := s.subscribe(r.Context(), r, w)
	if err != nil {
		log.Println("subscribeHandler failed", err)
		return
	}
}

func (s *server) subscribe(ctx context.Context, r *http.Request, w http.ResponseWriter) (err error) {
	connection, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		return err
	}
	defer connection.Close()

	client := NewClient(connection)

	defer func() {
		if err := connection.Close(); err != nil {
			log.Printf("failed to close connection: %v", err)
		}
	}()

	ticker := time.NewTicker(time.Millisecond * 250)
	defer ticker.Stop()

	for {
		select {
		case msg := <-client.messages:
			log.Println("sending message")
			err := connection.SetWriteDeadline(time.Now().Add(time.Second * 10))
			if err != nil {
				log.Println("failed to set timeout ", err)
			}
			err = wsutil.WriteServerText(connection, msg)
			if err != nil {
				log.Println("failed to write message", err)
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			msg, _, err := wsutil.ReadClientData(connection)
			if err != nil {
				if !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
					log.Println("failed to read message", err)
					continue
				}
				return err
			}
			if msg != nil {
				var message message
				err := json.Unmarshal(msg, &message)
				if err != nil {
					log.Println("failed to unmarshal message", err)
					continue
				}

				handlerFunc, ok := s.eventHandlers[message.EventName]
				if !ok {
					log.Println("unknown event name", message.EventName)
					continue
				}

				msg, err := handlerFunc(ctx, client, message)
				if err != nil {
					log.Println("failed to handle message", err)
					continue
				}
				client.messages <- msg
			}
		}
	}
}
