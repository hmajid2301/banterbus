package transport

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/golang-jwt/jwt/v5"
	"github.com/invopop/ctxi18n"
	"github.com/invopop/ctxi18n/i18n"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"gitlab.com/hmajid2301/banterbus/internal/views"
	"gitlab.com/hmajid2301/banterbus/internal/views/pages"
)

type Server struct {
	Logger    *slog.Logger
	Websocket websocketer
	Config    ServerConfig
	Server    *http.Server
	Keyfunc   jwt.Keyfunc
}

type ServerConfig struct {
	Host          string
	Port          int
	Environment   string
	DefaultLocale i18n.Code
}

type websocketer interface {
	Subscribe(r *http.Request, w http.ResponseWriter) (err error)
}

func NewServer(
	websocketer websocketer,
	logger *slog.Logger,
	staticFS http.FileSystem,
	keyfunc jwt.Keyfunc,
	config ServerConfig,
) *Server {
	s := &Server{
		Websocket: websocketer,
		Logger:    logger,
		Config:    config,
		Keyfunc:   keyfunc,
	}

	mux := http.NewServeMux()
	mux.Handle("/", s.localeMiddleware(http.HandlerFunc(s.indexHandler)))
	mux.Handle("/static/", http.StripPrefix("/static", http.FileServer(staticFS)))
	mux.Handle("/ws", s.localeMiddleware(http.HandlerFunc(s.subscribeHandler)))
	mux.Handle("/join/{room_code}", s.localeMiddleware(http.HandlerFunc(s.joinHandler)))
	mux.HandleFunc("/health", s.healthHandler)
	mux.HandleFunc("/readiness", s.readinessHandler)
	mux.Handle("/question", s.jwtMiddleware(http.HandlerFunc(s.addQuestionHandler)))

	if config.Environment == "local" {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	handler := otelhttp.NewHandler(mux, "/", otelhttp.WithFilter(func(r *http.Request) bool {
		return r.URL.Path != "/health" && r.URL.Path != "readiness"
	}))
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

func (s *Server) subscribeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	s.Logger.DebugContext(ctx, "subscribe handler called")
	err := s.Websocket.Subscribe(r, w)
	if err != nil {
		s.Logger.ErrorContext(ctx, "subscribe failed", slog.Any("error", err))
		return
	}
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	languages, err := views.ListLanguages()
	if err != nil {
		s.Logger.ErrorContext(r.Context(), "failed to list supported languages", slog.Any("error", err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	templ.Handler(pages.Index(languages)).ServeHTTP(w, r)
}

func (s *Server) addQuestionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
}

func (s *Server) jwtMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		bearerToken := authHeader[len("Bearer "):]
		token, err := jwt.Parse(bearerToken, s.Keyfunc)
		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *Server) readinessHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *Server) joinHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	s.Logger.DebugContext(ctx, "join handler called")
	roomCode := r.PathValue("room_code")

	languages, err := views.ListLanguages()
	if err != nil {
		s.Logger.ErrorContext(ctx, "failed to list supported languages", slog.Any("error", err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	templ.Handler(pages.Join(languages, roomCode)).ServeHTTP(w, r)
}

func (s *Server) localeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		locale := r.Header.Get("Accept-Language")
		pathSegments := strings.Split(r.URL.Path, "/")
		if len(pathSegments) > 1 {
			locale = pathSegments[1]
		}

		ctx, err := ctxi18n.WithLocale(r.Context(), locale)
		if err != nil {
			locale = s.Config.DefaultLocale.String()
			ctx, err = ctxi18n.WithLocale(r.Context(), locale)
			if err != nil {
				s.Logger.ErrorContext(
					ctx,
					"error setting locale",
					slog.Any("error", err),
					slog.String("locale", locale),
				)
				http.Error(w, "error setting locale", http.StatusBadRequest)
				return
			}
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "locale",
			Value:    locale,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			Expires:  time.Now().Add(time.Hour),
		})

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
