package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"strings"
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
		AdminGroup:    "admin",
	}

	// Create router with no base middleware
	router := middleware.NewRouter()

	// Public routes (no middleware)
	publicGroup := router.Group("public")
	publicGroup.HandleFunc("/health", s.healthHandler)
	publicGroup.HandleFunc("/readiness", s.readinessHandler)
	publicGroup.Handle("/static/", http.StripPrefix("/static", http.FileServer(staticFS)))

	// Game routes (with locale middleware)
	gameGroup := router.Group("game", m.Locale)
	gameGroup.HandleFunc("/", s.indexHandler)
	gameGroup.HandleFunc("/ws", s.subscribeHandler)
	gameGroup.HandleFunc("/join/{room_code}", s.joinHandler)

	// API routes (with locale + auth middleware)
	apiGroup := router.Group("api", m.Locale, m.ValidateJWT)
	apiGroup.Handle("/question", s.questionHandler())
	apiGroup.HandleFunc("/question/{id}/locale/{locale}", s.addQuestionTranslationHandler)
	apiGroup.Handle("/question/group", s.questionGroupHandler())

	// Admin routes (with locale + admin auth middleware)
	adminGroup := router.Group("admin", m.Locale, m.ValidateAdminJWT)
	adminGroup.Handle("/question/{id}/enable", s.methodHandler("PUT", s.enableQuestionHandler))
	adminGroup.Handle("/question/{id}/disable", s.methodHandler("PUT", s.disableQuestionHandler))

	// Debug routes (local environment only)
	if config.Environment == "local" {
		debugGroup := router.Group("debug")
		debugGroup.HandleFunc("/debug/pprof/", pprof.Index)
		debugGroup.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		debugGroup.HandleFunc("/debug/pprof/profile", pprof.Profile)
		debugGroup.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		debugGroup.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	// Create final handler with logging and OTEL
	httpSpanName := func(_ string, r *http.Request) string {
		return fmt.Sprintf("HTTP %s %s", r.Method, r.URL.Path)
	}

	otelFilters := func(r *http.Request) bool {
		return r.URL.Path != "/health" && r.URL.Path != "/readiness" && !strings.HasPrefix(r.URL.Path, "/static")
	}

	// Apply logging middleware to the entire router
	routes := m.Logging(router)

	handler := otelhttp.NewHandler(
		routes,
		"/",
		otelhttp.WithFilter(otelFilters),
		otelhttp.WithSpanNameFormatter(httpSpanName),
	)
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

// Helper methods for route organization

// questionHandler handles both GET and POST requests for /question
func (s *Server) questionHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			s.addQuestionHandler(w, r)
		case http.MethodGet:
			s.getQuestionsHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

// questionGroupHandler handles both GET and POST requests for /question/group
func (s *Server) questionGroupHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			s.addGroupHandler(w, r)
		case http.MethodGet:
			s.getGroupsHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

// methodHandler restricts a handler to a specific HTTP method
func (s *Server) methodHandler(method string, handler http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == method {
			handler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
