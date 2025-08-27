package service

import (
	"context"

	"github.com/gofrs/uuid/v5"

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
	const gameName = "fibbing_it" // TODO: make this configurable
	u, err := q.store.CreateQuestion(ctx, db.CreateQuestionArgs{
		GameName:  gameName,
		GroupName: group,
		RoundType: roundType,
		Text:      text,
		Locale:    q.defaultLocale,
	})

	return Question{
		ID:        u.String(),
		Text:      text,
		GroupName: group,
		Locale:    q.defaultLocale,
		RoundType: roundType,
		Enabled:   true,
	}, err
}

func (q QuestionService) AddTranslation(
	ctx context.Context,
	questionID uuid.UUID,
	text string,
	locale string,
) (QuestionTranslation, error) {
	u, err := q.randomizer.GetID()
	if err != nil {
		return QuestionTranslation{}, err
	}
	_, err = q.store.AddQuestionTranslation(ctx, db.AddQuestionTranslationParams{
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

func (q QuestionService) GetGroups(ctx context.Context) ([]Group, error) {
	groupsDB, err := q.store.GetGroups(ctx)
	if err != nil {
		return nil, err
	}

	groups := []Group{}
	for _, g := range groupsDB {
		group := Group{Name: g.GroupName, ID: g.ID.String()}
		groups = append(groups, group)
	}

	return groups, nil
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
			ID:        q.ID.String(),
			Text:      q.Question,
			GroupName: q.GroupName,
			Locale:    q.Locale,
			RoundType: q.RoundType,
			Enabled:   q.Enabled.Bool,
		}
		questions = append(questions, question)
	}

	return questions, err
}

func (q QuestionService) AddGroup(ctx context.Context, name string, groupType ...string) (Group, error) {
	u, err := q.randomizer.GetID()
	if err != nil {
		return Group{}, err
	}

	// Use provided groupType or default to "questions"
	gType := "questions"
	if len(groupType) > 0 {
		gType = groupType[0]
	}

	_, err = q.store.AddGroup(ctx, db.AddGroupParams{
		ID:        u,
		GroupName: name,
		GroupType: gType,
	})
	if err != nil {
		return Group{}, err
	}

	return Group{
		ID:   u.String(),
		Name: name,
		Type: gType,
	}, nil
}

func (q QuestionService) DisableQuestion(ctx context.Context, id uuid.UUID) error {
	_, err := q.store.DisableQuestion(ctx, id)
	return err
}

func (q QuestionService) EnableQuestion(ctx context.Context, id uuid.UUID) error {
	_, err := q.store.EnableQuestion(ctx, id)
	return err
}
