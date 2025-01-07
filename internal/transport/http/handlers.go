package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/invopop/ctxi18n/i18n"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"gitlab.com/hmajid2301/banterbus/internal/transport/http/middleware"
)

type Server struct {
	Logger          *slog.Logger
	Websocket       websocketer
	Config          ServerConfig
	Server          *http.Server
	QuestionService QuestionServicer
}

type ServerConfig struct {
	Host          string
	Port          int
	Environment   string
	DefaultLocale i18n.Code
	AuthDisabled  bool
}

type websocketer interface {
	Subscribe(r *http.Request, w http.ResponseWriter) (err error)
}

func NewServer(
	websocketer websocketer,
	logger *slog.Logger,
	staticFS http.FileSystem,
	keyfunc jwt.Keyfunc,
	questionService QuestionServicer,
	config ServerConfig,

) *Server {
	s := &Server{
		Websocket:       websocketer,
		Logger:          logger,
		Config:          config,
		QuestionService: questionService,
	}

	handler := s.setupHTTPRoutes(config, keyfunc, staticFS)
	writeTimeout := 10
	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.Config.Host, s.Config.Port),
		ReadTimeout:  time.Second,
		WriteTimeout: time.Duration(writeTimeout) * time.Second,
		Handler:      handler,
	}
	s.Server = httpServer

	return s
}

func (s *Server) setupHTTPRoutes(config ServerConfig, keyfunc jwt.Keyfunc, staticFS http.FileSystem) http.Handler {
	m := middleware.Middleware{
		DefaultLocale: config.DefaultLocale.String(),
		Logger:        s.Logger,
		Keyfunc:       keyfunc,
		DisableAuth:   config.AuthDisabled,
	}

	mux := http.NewServeMux()
	mux.Handle("/", m.Locale(http.HandlerFunc(s.indexHandler)))
	mux.Handle("/static/", http.StripPrefix("/static", http.FileServer(staticFS)))
	mux.Handle("/ws", m.Locale(http.HandlerFunc(s.subscribeHandler)))
	mux.Handle("/join/{room_code}", m.Locale(http.HandlerFunc(s.joinHandler)))

	mux.HandleFunc("/health", s.healthHandler)
	mux.HandleFunc("/readiness", s.readinessHandler)

	mux.Handle("/question", m.ValidateJWT(http.HandlerFunc(s.addQuestionHandler)))
	mux.Handle("/question/{id}/locale/{locale}", m.ValidateJWT(http.HandlerFunc(s.addQuestionTranslationHandler)))

	if config.Environment == "local" {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	handler := otelhttp.NewHandler(mux, "/", otelhttp.WithFilter(func(r *http.Request) bool {
		return r.URL.Path != "/health" && r.URL.Path != "/readiness"
	}))

	return handler
}

func (s *Server) Serve(ctx context.Context) error {
	s.Logger.InfoContext(ctx, "starting server")
	err := s.Server.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.Logger.InfoContext(ctx, "shutting down server")
	err := s.Server.Shutdown(ctx)
	return err
}
