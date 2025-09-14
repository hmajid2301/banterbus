package http

import (
	"log/slog"
	"net/http"

	"gitlab.com/hmajid2301/banterbus/internal/views"
	"gitlab.com/hmajid2301/banterbus/internal/views/pages"

	"github.com/a-h/templ"
)

// TODO: Simplified handler struct?
func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	languages, err := views.ListLanguages()
	if err != nil {
		s.Logger.ErrorContext(r.Context(), "failed to list supported languages", slog.Any("error", err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	templ.Handler(pages.Index(languages, s.Config.Environment)).ServeHTTP(w, r)
}

func (s *Server) joinHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	roomCode := r.PathValue("room_code")

	languages, err := views.ListLanguages()
	if err != nil {
		s.Logger.ErrorContext(ctx, "failed to list supported languages", slog.Any("error", err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	templ.Handler(pages.Join(languages, s.Config.Environment, roomCode)).ServeHTTP(w, r)
}

func (s *Server) subscribeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := s.Websocket.Subscribe(r, w)
	if err != nil {
		s.Logger.ErrorContext(ctx, "subscribe failed", slog.Any("error", err))
		return
	}
}
