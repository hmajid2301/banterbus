package transport

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"gitlab.com/hmajid2301/banterbus/internal/views/pages"
)

type Server struct {
	logger    *slog.Logger
	websocket websocketer
	Server    *http.Server
}

type websocketer interface {
	Subscribe(r *http.Request, w http.ResponseWriter) (err error)
}

func NewServer(websocketer websocketer, logger *slog.Logger, staticFS http.FileSystem) *Server {
	s := &Server{
		websocket: websocketer,
		logger:    logger,
	}

	mux := http.NewServeMux()
	mux.Handle("/", templ.Handler(pages.Index()))
	mux.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(staticFS)))
	mux.Handle("/metrics", promhttp.Handler())

	handleFunc := func(pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
		handler := otelhttp.WithRouteTag(pattern, http.HandlerFunc(handlerFunc))
		mux.Handle(pattern, handler)
	}

	handleFunc("/ws", s.subscribeHandler)
	handleFunc("/health", s.health)
	handleFunc("/readiness", s.readiness)

	handler := otelhttp.NewHandler(mux, "/")
	writeTimeout := 10
	httpServer := &http.Server{
		Addr:         "0.0.0.0:8080",
		ReadTimeout:  time.Second,
		WriteTimeout: time.Duration(writeTimeout) * time.Second,
		Handler:      handler,
	}
	s.Server = httpServer

	return s
}

func (s *Server) Serve() error {
	s.logger.Info("starting server")
	err := s.Server.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down server")
	err := s.Server.Shutdown(ctx)
	return err
}

func (s *Server) subscribeHandler(w http.ResponseWriter, r *http.Request) {
	s.logger.Debug("subscribe handler called")
	err := s.websocket.Subscribe(r, w)
	if err != nil {
		s.logger.Error("subscribe failed", slog.Any("error", err))
		return
	}
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *Server) readiness(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
