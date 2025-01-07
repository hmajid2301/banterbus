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
