package service

import (
	"context"

	"github.com/google/uuid"

	"gitlab.com/hmajid2301/banterbus/internal/store/db"
)

type QuestionService struct {
	store         Storer
	randomizer    Randomizer
	defaultLocale string
}

func NewQuestionService(store Storer, randomizer Randomizer, defaultLocale string) *QuestionService {
	return &QuestionService{store: store, randomizer: randomizer, defaultLocale: defaultLocale}
}

func (q QuestionService) Add(
	ctx context.Context,
	text string,
	group string,
	roundType string,
) (Question, error) {
	// TODO: do not hardcode game name here
	err := q.store.CreateQuestion(ctx, db.CreateQuestionArgs{
		GameName:  "fibbing_it",
		GroupName: group,
		RoundType: roundType,
		Text:      text,
		Locale:    q.defaultLocale,
	})

	return Question{
		Text:      text,
		GroupName: group,
		Locale:    q.defaultLocale,
		RoundType: roundType,
	}, err
}

func (q QuestionService) AddTranslation(
	ctx context.Context,
	questionID uuid.UUID,
	text string,
	locale string,
) (QuestionTranslation, error) {
	u := q.randomizer.GetID()
	_, err := q.store.AddQuestionTranslation(ctx, db.AddQuestionTranslationParams{
		ID:         u,
		Question:   text,
		Locale:     locale,
		QuestionID: questionID,
	})

	return QuestionTranslation{
		Text:   text,
		Locale: locale,
	}, err
}

func (q QuestionService) GetGroupNames(ctx context.Context) ([]string, error) {
	return q.store.GetGroupNames(ctx)
}

type GetQuestionFilters struct {
	Locale    string
	RoundType string
	GroupName string
	Enabled   string
}

// TODO: fix pagination
func (q QuestionService) GetQuestions(
	ctx context.Context,
	filters GetQuestionFilters,
	limit int32,
	pageNum int32,
) ([]Question, error) {
	offset := (pageNum - 1) * limit
	// TODO: refactor query to include actual names i.e. Locale instead of column1
	qq, err := q.store.GetQuestions(ctx, db.GetQuestionsParams{
		Column1: filters.Locale,
		Column2: filters.RoundType,
		Column3: filters.GroupName,
		Column4: true,
		Limit:   limit,
		Offset:  offset,
	})

	questions := []Question{}
	for _, q := range qq {
		question := Question{
			Text:      q.Question,
			GroupName: q.GroupName,
			Locale:    q.Locale,
			RoundType: q.RoundType,
		}
		questions = append(questions, question)
	}

	return questions, err
}
