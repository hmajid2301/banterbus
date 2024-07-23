package transport

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"nhooyr.io/websocket"
)

type server struct {
	subscribersMessageBuffer int
	mux                      http.ServeMux
	subscriberMutex          sync.Mutex
	subscribers              map[subscriber]struct{}
}

type subscriber struct {
	msgs chan []byte
}

func NewHTTPServer() *server {
	s := &server{
		subscribersMessageBuffer: 10,
		subscribers:              make(map[subscriber]struct{}),
	}
	s.mux.Handle("/", http.FileServer(http.Dir("./static")))
	s.mux.HandleFunc("/ws", s.subscribeHandler)

	return s
}

func (s *server) Start() error {
	fmt.Println("Starting server")
	go func() {
		for {
			timeStamp := time.Now().Format("2006-01-02 15:04:05")
			msg := []byte(`
      <div hx-swap-oob="innerHTML:#timestamp">
        <p><i style="color: green" class="fa fa-circle"></i> ` + timeStamp + `</p>
      </div>`)
			s.Broadcast(msg)
			time.Sleep(3 * time.Second)
		}

	}()
	err := http.ListenAndServe(":8080", &s.mux)
	if err != nil {
		return err
	}

	return nil
}

func (s *server) subscribeHandler(w http.ResponseWriter, r *http.Request) {
	err := s.subscribe(r.Context(), r, w)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (s *server) subscribe(ctx context.Context, r *http.Request, w http.ResponseWriter) (err error) {
	var c *websocket.Conn
	sub := subscriber{
		msgs: make(chan []byte, s.subscribersMessageBuffer),
	}
	s.addSubscriber(sub)
	c, err = websocket.Accept(w, r, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := c.CloseNow(); err != nil {
			log.Printf("Failed to close connection: %v", err)
		}
	}()

	ctx = c.CloseRead(ctx)
	for {
		select {
		case msg := <-sub.msgs:
			ctx, cancel := context.WithTimeout(ctx, time.Second)
			defer cancel()
			err := c.Write(ctx, websocket.MessageText, msg)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s *server) Broadcast(msg []byte) {
	s.subscriberMutex.Lock()
	defer s.subscriberMutex.Unlock()
	for sub := range s.subscribers {
		sub.msgs <- msg
	}
}

func (s *server) addSubscriber(sub subscriber) {
	s.subscriberMutex.Lock()
	s.subscribers[sub] = struct{}{}
	s.subscriberMutex.Unlock()
	fmt.Println("Subscriber added", sub)
}
