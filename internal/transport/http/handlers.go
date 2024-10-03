package transport

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Server struct {
	logger    *slog.Logger
	websocket websocketer
	srv       *http.Server
}

type websocketer interface {
	Subscribe(ctx context.Context, r *http.Request, w http.ResponseWriter) (err error)
}

func NewServer(websocketer websocketer, logger *slog.Logger, staticFS http.FileSystem) *Server {
	s := &Server{
		websocket: websocketer,
		logger:    logger,
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(staticFS))

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
		Addr:         "0.0.0.0:8080",
		ReadTimeout:  time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      handler,
	}
	s.srv = srv

	return s
}

func (s *Server) Serve() error {
	s.logger.Info("starting server")
	err := s.srv.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down server")
	err := s.srv.Shutdown(ctx)
	return err
}

func (s *Server) subscribeHandler(w http.ResponseWriter, r *http.Request) {
	s.logger.Debug("subscribe handler called")
	err := s.websocket.Subscribe(r.Context(), r, w)
	if err != nil {
		s.logger.Error("subscribe failed", slog.Any("error", err))
		return
	}
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *Server) readiness(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
