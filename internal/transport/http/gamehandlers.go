package http

import (
	"log/slog"
	"net/http"

	"github.com/a-h/templ"

	"gitlab.com/hmajid2301/banterbus/internal/views"
	"gitlab.com/hmajid2301/banterbus/internal/views/pages"
)

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	languages, err := views.ListLanguages()
	if err != nil {
		s.Logger.ErrorContext(r.Context(), "failed to list supported languages", slog.Any("error", err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	templ.Handler(pages.Index(languages)).ServeHTTP(w, r)
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

func (s *Server) subscribeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	s.Logger.DebugContext(ctx, "subscribe handler called")
	err := s.Websocket.Subscribe(r, w)
	if err != nil {
		s.Logger.ErrorContext(ctx, "subscribe failed", slog.Any("error", err))
		return
	}
}
