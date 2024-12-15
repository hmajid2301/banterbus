package transport

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/invopop/ctxi18n"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"gitlab.com/hmajid2301/banterbus/internal/views"
	"gitlab.com/hmajid2301/banterbus/internal/views/pages"
)

type Server struct {
	Logger    *slog.Logger
	Websocket websocketer
	Config    ServerConfig
	Server    *http.Server
}

type ServerConfig struct {
	Host          string
	Port          int
	DefaultLocale i18n.Code
}

type websocketer interface {
	Subscribe(r *http.Request, w http.ResponseWriter) (err error)
}

func NewServer(websocketer websocketer, logger *slog.Logger, staticFS http.FileSystem, config ServerConfig) *Server {
	s := &Server{
		Websocket: websocketer,
		Logger:    logger,
		Config:    config,
	}

	mux := http.NewServeMux()
	mux.Handle("/", s.LocaleMiddleware(http.HandlerFunc(s.indexHandler)))
	mux.Handle("/static/", http.StripPrefix("/static", http.FileServer(staticFS)))
	mux.Handle("/ws", s.LocaleMiddleware(http.HandlerFunc(s.subscribeHandler)))
	mux.Handle("/join/{room_code}", s.LocaleMiddleware(http.HandlerFunc(s.joinHandler)))
	mux.HandleFunc("/health", s.healthHandler)
	mux.HandleFunc("/readiness", s.readinessHandler)
	mux.Handle("/metrics", promhttp.Handler())

	handler := otelhttp.NewHandler(mux, "/")
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

func (s *Server) Serve() error {
	s.Logger.Info("starting server")
	err := s.Server.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.Logger.Info("shutting down server")
	err := s.Server.Shutdown(ctx)
	return err
}

func (s *Server) subscribeHandler(w http.ResponseWriter, r *http.Request) {
	s.Logger.Debug("subscribe handler called")
	err := s.Websocket.Subscribe(r, w)
	if err != nil {
		s.Logger.Error("subscribe failed", slog.Any("error", err))
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

func (s *Server) healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *Server) readinessHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *Server) joinHandler(w http.ResponseWriter, r *http.Request) {
	s.Logger.Debug("join handler called")
	roomCode := r.PathValue("room_code")

	languages, err := views.ListLanguages()
	if err != nil {
		s.Logger.ErrorContext(r.Context(), "failed to list supported languages", slog.Any("error", err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	templ.Handler(pages.Join(languages, roomCode)).ServeHTTP(w, r)
}

func (s *Server) LocaleMiddleware(next http.Handler) http.Handler {
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
				s.Logger.Error("error setting locale", slog.Any("error", err), slog.String("locale", locale))
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
