package transport

import (
	"context"
	"log/slog"
	"net/http"
)

type server struct {
	logger      *slog.Logger
	websocketer websocketer
	mux         http.ServeMux
}

type websocketer interface {
	Subscribe(ctx context.Context, r *http.Request, w http.ResponseWriter) (err error)
}

func NewServer(websocketer websocketer, logger *slog.Logger) *server {
	s := &server{
		websocketer: websocketer,
		logger:      logger,
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
	err := s.websocketer.Subscribe(r.Context(), r, w)
	if err != nil {
		s.logger.Error("subscribe failed", slog.Any("error", err))
		return
	}
}
