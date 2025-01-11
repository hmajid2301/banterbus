package http

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"gitlab.com/hmajid2301/banterbus/internal/service"
)

type QuestionServicer interface {
	Add(ctx context.Context, text string, group string, roundType string) (service.Question, error)
	AddTranslation(
		ctx context.Context,
		questionID uuid.UUID,
		text string,
		locale string,
	) (service.QuestionTranslation, error)
	GetGroupNames(ctx context.Context) ([]string, error)
}

type NewQuestion struct {
	Text      string `json:"text"       validate:"required"`
	GroupName string `json:"group_name" validate:"required"`
	RoundType string `json:"round_type" validate:"required"`
}

func (s *Server) addQuestionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	s.Logger.DebugContext(ctx, "addQuestionHandler called")

	if r.Method != http.MethodPost {
		s.Logger.WarnContext(ctx, "invalid HTTP method, expected POST", slog.Any("method", r.Method))
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.Logger.ErrorContext(ctx, "failed to ready request body", slog.Any("error", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var newQuestion NewQuestion
	if err := json.Unmarshal(body, &newQuestion); err != nil {
		s.Logger.ErrorContext(ctx, "failed to unmarshal json", slog.Any("error", err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	validate := validator.New()
	err = validate.Struct(newQuestion)
	if err != nil {
		s.Logger.ErrorContext(ctx, "failed to validate json", slog.Any("error", err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	_, err = s.QuestionService.Add(ctx, newQuestion.Text, newQuestion.GroupName, newQuestion.RoundType)
	if err != nil {
		s.Logger.ErrorContext(ctx, "failed to add question", slog.Any("error", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

type NewQuestionTranslation struct {
	Text string `json:"text" validate:"required"`
}

func (s *Server) addQuestionTranslationHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	s.Logger.DebugContext(ctx, "addQuestionTranslationHandler called")

	if r.Method != http.MethodPut {
		s.Logger.WarnContext(ctx, "invalid HTTP method, expected PUT", slog.Any("method", r.Method))
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.Logger.ErrorContext(ctx, "failed to ready request body", slog.Any("error", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var newQuestion NewQuestionTranslation
	if err := json.Unmarshal(body, &newQuestion); err != nil {
		s.Logger.ErrorContext(ctx, "failed to unmarshal json", slog.Any("error", err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	validate := validator.New()
	err = validate.Struct(newQuestion)
	if err != nil {
		s.Logger.ErrorContext(ctx, "failed to validate json", slog.Any("error", err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	id := r.PathValue("id")
	locale := r.PathValue("locale")

	questionID, err := uuid.Parse(id)
	if err != nil {
		s.Logger.ErrorContext(ctx, "failed to parse question UUID", slog.Any("error", err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	_, err = s.QuestionService.AddTranslation(ctx, questionID, newQuestion.Text, locale)
	if err != nil {
		s.Logger.ErrorContext(ctx, "failed to add question translation", slog.Any("error", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) getGroupNamesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	s.Logger.DebugContext(ctx, "GetGroupNamesHandler called")

	if r.Method != http.MethodGet {
		s.Logger.WarnContext(ctx, "invalid HTTP method, expected GET", slog.Any("method", r.Method))
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	groupNames, err := s.QuestionService.GetGroupNames(ctx)
	if err != nil {
		s.Logger.ErrorContext(ctx, "failed to get group names", slog.Any("error", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(groupNames); err != nil {
		s.Logger.ErrorContext(ctx, "failed to encode group names", slog.Any("error", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
