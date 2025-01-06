package service

import (
	"context"

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
	})

	return Question{
		Text:      text,
		GroupName: group,
		Locale:    q.defaultLocale,
		RoundType: roundType,
	}, err
}
