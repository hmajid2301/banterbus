package http

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"math"
	"net/http"
	"strconv"

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
	GetQuestions(
		ctx context.Context,
		filters service.GetQuestionFilters,
		limit int32,
		pageNum int32,
	) ([]service.Question, error)
	DisableQuestion(ctx context.Context, id uuid.UUID) error
	EnableQuestion(ctx context.Context, id uuid.UUID) error
	AddGroup(ctx context.Context, name string) (service.Group, error)
	GetGroups(ctx context.Context) ([]service.Group, error)
}

type NewQuestion struct {
	Text      string `json:"text"       validate:"required"`
	GroupName string `json:"group_name" validate:"required"`
	RoundType string `json:"round_type" validate:"required"`
}

func (s *Server) addQuestionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

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

type Question struct {
	Questions []service.Question
}

func (s *Server) getQuestionsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// TODO: validate round_type and group_name
	roundType := r.URL.Query().Get("round_type")
	groupName := r.URL.Query().Get("group_name")
	limitQuery := r.URL.Query().Get("limit")
	pageNumQuery := r.URL.Query().Get("page_num")

	if limitQuery == "" {
		limitQuery = "100"
	}

	if pageNumQuery == "" {
		pageNumQuery = "1"
	}

	limit, err := strconv.Atoi(limitQuery)
	if err != nil {
		s.Logger.ErrorContext(ctx, "failed to parse limit", slog.Any("error", err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if limit < 0 || limit > math.MaxInt32 {
		s.Logger.ErrorContext(ctx, "limit out of range must be greater than 0", slog.Any("limit", limit))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	pageNum, err := strconv.Atoi(pageNumQuery)
	if err != nil {
		s.Logger.ErrorContext(ctx, "failed to parse pageNum", slog.Any("error", err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if pageNum < 1 || pageNum > math.MaxInt32 {
		s.Logger.ErrorContext(ctx, "pageNum out of range must be greater than 0", slog.Any("page_num", pageNum))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	filters := service.GetQuestionFilters{
		Locale:    "",
		RoundType: roundType,
		GroupName: groupName,
	}

	//nolint:gosec // disable G109
	// We check the max limit above
	questions, err := s.QuestionService.GetQuestions(ctx, filters, int32(limit), int32(pageNum))
	if err != nil {
		s.Logger.ErrorContext(ctx, "failed to get questions", slog.Any("error", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	respBody := Question{
		Questions: questions,
	}

	resp, err := json.Marshal(respBody)
	if err != nil {
		s.Logger.ErrorContext(ctx, "failed to encode questions", slog.Any("error", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, err = w.Write(resp)
	if err != nil {
		s.Logger.ErrorContext(ctx, "failed to write JSON", slog.Any("error", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) disableQuestionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := r.PathValue("id")

	questionID, err := uuid.Parse(id)
	if err != nil {
		s.Logger.ErrorContext(ctx, "failed to parse question UUID", slog.Any("error", err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	err = s.QuestionService.DisableQuestion(ctx, questionID)
	if err != nil {
		s.Logger.ErrorContext(ctx, "failed to disable question", slog.Any("error", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) enableQuestionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := r.PathValue("id")

	questionID, err := uuid.Parse(id)
	if err != nil {
		s.Logger.ErrorContext(ctx, "failed to parse question UUID", slog.Any("error", err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	err = s.QuestionService.EnableQuestion(ctx, questionID)
	if err != nil {
		s.Logger.ErrorContext(ctx, "failed to enable question", slog.Any("error", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type NewGroup struct {
	Name string `json:"group_name" validate:"required"`
}

func (s *Server) addGroupHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.Logger.ErrorContext(ctx, "failed to ready request body", slog.Any("error", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var newGroup NewGroup
	if err := json.Unmarshal(body, &newGroup); err != nil {
		s.Logger.ErrorContext(ctx, "failed to unmarshal json", slog.Any("error", err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	validate := validator.New()
	err = validate.Struct(newGroup)
	if err != nil {
		s.Logger.ErrorContext(ctx, "failed to validate json", slog.Any("error", err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	_, err = s.QuestionService.AddGroup(ctx, newGroup.Name)
	if err != nil {
		s.Logger.ErrorContext(ctx, "failed to add groups", slog.Any("error", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

type Group struct {
	Groups []service.Group
}

func (s *Server) getGroupsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	groups, err := s.QuestionService.GetGroups(ctx)
	if err != nil {
		s.Logger.ErrorContext(ctx, "failed to get groups", slog.Any("error", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	respBody := Group{
		Groups: groups,
	}

	resp, err := json.Marshal(respBody)
	if err != nil {
		s.Logger.ErrorContext(ctx, "failed to encode groups", slog.Any("error", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, err = w.Write(resp)
	if err != nil {
		s.Logger.ErrorContext(ctx, "failed to write JSON", slog.Any("error", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
