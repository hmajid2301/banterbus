package ws

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type server struct {
	mux http.ServeMux
}

type Message struct {
	Data interface{} `json:"data"`
	Type string      `json:"type"`
}

func NewHTTPServer() *server {
	s := &server{}
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
	connection, err := websocket.Accept(w, r, nil)
	if err != nil {
		return err
	}
	defer connection.Close()

	client := NewClient(connection)

	defer func() {
		if err := connection.CloseNow(); err != nil {
			log.Printf("Failed to close connection: %v", err)
		}
	}()

	ticker := time.NewTicker(time.Millisecond * 250)
	defer ticker.Stop()

	for {
		select {
		case msg := <-client.messages:
			err := connection.SetWriteDeadline(time.Now().Add(time.Second))
			if err != nil {
				log.Println("failed to set timeout ", err)
			}
			_, err = connection.Write(msg)
			if err != nil {
				log.Println("failed to write message", err)
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			msg, op, err := wsutil.ReadClientData(connection)
			if err != nil {
				if !errors.Is(err, io.ErrUnexpectedEOF) || !errors.Is(err, io.EOF) {
					log.Println("failed to read message", err)
					continue
				}
				// return err?
			}
			if msg != nil {
				log.Println("msg", string(msg), op)
			}
		}
	}
}
