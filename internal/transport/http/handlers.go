package transport

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type server struct {
	logger    *slog.Logger
	websocket websocketer
	srv       *http.Server
}

type websocketer interface {
	Subscribe(ctx context.Context, r *http.Request, w http.ResponseWriter) (err error)
}

func NewServer(websocketer websocketer, logger *slog.Logger) *server {
	s := &server{
		websocket: websocketer,
		logger:    logger,
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("./static")))

	handleFunc := func(pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
		// Configure the "http.route" for the HTTP instrumentation.
		handler := otelhttp.WithRouteTag(pattern, http.HandlerFunc(handlerFunc))
		mux.Handle(pattern, handler)
	}

	handler := otelhttp.NewHandler(mux, "/")
	handleFunc("/ws", s.subscribeHandler)
	handleFunc("/health", s.health)
	handleFunc("/readiness", s.readiness)

	srv := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      handler,
	}
	s.srv = srv

	return s
}

func (s *server) Serve() error {
	s.logger.Info("starting server")
	err := s.srv.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}

func (s *server) subscribeHandler(w http.ResponseWriter, r *http.Request) {
	s.logger.Debug("subscribe handler called")
	err := s.websocket.Subscribe(r.Context(), r, w)
	if err != nil {
		s.logger.Error("subscribe failed", slog.Any("error", err))
		return
	}
}

func (s *server) health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *server) readiness(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
