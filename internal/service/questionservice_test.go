package service_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"gitlab.com/hmajid2301/banterbus/internal/service"
	mockService "gitlab.com/hmajid2301/banterbus/internal/service/mocks"
	"gitlab.com/hmajid2301/banterbus/internal/store/db"
)

func TestQuestionAdd(t *testing.T) {
	t.Run("Should successfully add new question", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewQuestionService(mockStore, mockRandom, "en-GB")

		ctx := context.Background()

		roundType := "multiple_choice"
		groupName := "example"
		text := "an example question"

		mockStore.EXPECT().CreateQuestion(ctx, db.CreateQuestionArgs{
			GameName:  "fibbing_it",
			GroupName: groupName,
			RoundType: roundType,
			Text:      text,
			Locale:    "en-GB",
		}).Return(uuid.UUID{}, nil)

		question, err := srv.Add(ctx, text, groupName, roundType)
		assert.NoError(t, err)

		expectedQuestion := service.Question{
			ID:        question.ID,
			Text:      text,
			GroupName: groupName,
			Locale:    "en-GB",
			RoundType: roundType,
			Enabled:   true,
		}
		assert.Equal(t, expectedQuestion, question)
	})

	t.Run("Should fail to add new question, db fails", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewQuestionService(mockStore, mockRandom, "en-GB")

		ctx := context.Background()

		roundType := "multiple_choice"
		groupName := "example"
		text := "an example question"

		mockStore.EXPECT().CreateQuestion(ctx, db.CreateQuestionArgs{
			GameName:  "fibbing_it",
			GroupName: groupName,
			RoundType: roundType,
			Text:      text,
			Locale:    "en-GB",
		}).Return(uuid.UUID{}, fmt.Errorf("failed to create question"))

		_, err := srv.Add(ctx, text, groupName, roundType)
		assert.Error(t, err)
	})
}

func TestQuestionAddTranslation(t *testing.T) {
	t.Run("Should successfully add new question translation", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewQuestionService(mockStore, mockRandom, "en-GB")

		ctx := context.Background()

		text := "portguese words"
		locale := "pt-PT"

		u := uuid.Must(uuid.NewV7())
		questionID := uuid.Must(uuid.NewV7())
		mockRandom.EXPECT().GetID().Return(u)
		mockStore.EXPECT().AddQuestionTranslation(ctx, db.AddQuestionTranslationParams{
			ID:         u,
			Question:   text,
			Locale:     locale,
			QuestionID: questionID,
		}).Return(db.QuestionsI18n{}, nil)

		question, err := srv.AddTranslation(ctx, questionID, text, locale)
		assert.NoError(t, err)

		expectedQuestion := service.QuestionTranslation{
			Text:   text,
			Locale: locale,
		}
		assert.Equal(t, expectedQuestion, question)
	})

	t.Run("Should fail to add new question translation, db fails", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewQuestionService(mockStore, mockRandom, "en-GB")

		ctx := context.Background()

		text := "portguese words"
		locale := "pt-PT"

		u := uuid.Must(uuid.NewV7())
		questionID := uuid.Must(uuid.NewV7())
		mockRandom.EXPECT().GetID().Return(u)
		mockStore.EXPECT().AddQuestionTranslation(ctx, db.AddQuestionTranslationParams{
			ID:         u,
			Question:   text,
			Locale:     locale,
			QuestionID: questionID,
		}).Return(db.QuestionsI18n{}, fmt.Errorf("failed to add question translation"))

		_, err := srv.AddTranslation(ctx, questionID, text, locale)
		assert.Error(t, err)
	})
}

func TestQuestionGetGroups(t *testing.T) {
	t.Run("Should successfully get all group", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewQuestionService(mockStore, mockRandom, "en-GB")

		ctx := context.Background()

		u := uuid.Must(uuid.NewV7())
		mockStore.EXPECT().GetGroups(ctx).Return([]db.QuestionsGroup{
			{
				ID:        u,
				GroupName: "cat",
			},
			{
				ID:        u,
				GroupName: "programming",
			},
		}, nil)

		groups, err := srv.GetGroups(ctx)
		assert.NoError(t, err)
		expectedGroups := []service.Group{
			{
				ID:   u.String(),
				Name: "cat",
			},
			{
				ID:   u.String(),
				Name: "programming",
			},
		}
		assert.Equal(t, expectedGroups, groups)
	})

	t.Run("Should fail get all group, request to DB fails", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewQuestionService(mockStore, mockRandom, "en-GB")

		ctx := context.Background()

		mockStore.EXPECT().GetGroups(ctx).Return(nil, fmt.Errorf("failed to make request"))
		_, err := srv.GetGroups(ctx)
		assert.Error(t, err)
	})
}

func TestQuestionGetQuestions(t *testing.T) {
	t.Run("Should successfully get questions", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewQuestionService(mockStore, mockRandom, "en-GB")

		ctx := context.Background()

		questionsDB := []db.GetQuestionsRow{
			{
				Question:  "Why are cats cool",
				GroupName: "cat",
				Locale:    "en-GB",
				RoundType: "free_form",
			},
			{
				Question:  "What is your favourite cat",
				GroupName: "cat",
				Locale:    "en-GB",
				RoundType: "free_form",
			},
		}
		mockStore.EXPECT().GetQuestions(ctx, db.GetQuestionsParams{
			Column1: "en-GB",
			Column2: "free_form",
			Column3: "cat",
			Column4: true,
			Limit:   100,
			Offset:  0,
		}).Return(questionsDB, nil)

		filters := service.GetQuestionFilters{
			Locale:    "en-GB",
			RoundType: "free_form",
			GroupName: "cat",
		}
		questions, err := srv.GetQuestions(ctx, filters, 100, 1)
		assert.NoError(t, err)
		expectedQuestions := []service.Question{
			{
				ID:        questions[0].ID,
				Text:      "Why are cats cool",
				GroupName: "cat",
				Locale:    "en-GB",
				RoundType: "free_form",
				Enabled:   false,
			},
			{
				ID:        questions[1].ID,
				Text:      "What is your favourite cat",
				GroupName: "cat",
				Locale:    "en-GB",
				RoundType: "free_form",
				Enabled:   false,
			},
		}
		assert.Equal(t, expectedQuestions, questions)
	})

	t.Run("Should fail to get questions, DB fails", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewQuestionService(mockStore, mockRandom, "en-GB")

		ctx := context.Background()
		mockStore.EXPECT().GetQuestions(ctx, db.GetQuestionsParams{
			Column1: "en-GB",
			Column2: "free_form",
			Column3: "cat",
			Column4: true,
			Limit:   100,
			Offset:  0,
		}).Return(nil, fmt.Errorf("fail to get questions"))

		filters := service.GetQuestionFilters{
			Locale:    "en-GB",
			RoundType: "free_form",
			GroupName: "cat",
		}
		_, err := srv.GetQuestions(ctx, filters, 100, 1)
		assert.Error(t, err)
	})
}

func TestQuestionDisable(t *testing.T) {
	t.Run("Should successfully disable question", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewQuestionService(mockStore, mockRandom, "en-GB")

		ctx := context.Background()

		u := uuid.Must(uuid.NewV7())
		mockStore.EXPECT().DisableQuestion(ctx, u).Return(db.Question{}, nil)

		err := srv.DisableQuestion(ctx, u)
		assert.NoError(t, err)
	})

	t.Run("Should fail to disable question, db fails", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewQuestionService(mockStore, mockRandom, "en-GB")

		ctx := context.Background()

		u := uuid.Must(uuid.NewV7())
		mockStore.EXPECT().DisableQuestion(ctx, u).Return(db.Question{}, fmt.Errorf("failed to disable question"))

		err := srv.DisableQuestion(ctx, u)
		assert.Error(t, err)
	})
}

func TestQuestionEnable(t *testing.T) {
	t.Run("Should successfully enable question", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewQuestionService(mockStore, mockRandom, "en-GB")

		ctx := context.Background()

		u := uuid.Must(uuid.NewV7())
		mockStore.EXPECT().EnableQuestion(ctx, u).Return(db.Question{}, nil)

		err := srv.EnableQuestion(ctx, u)
		assert.NoError(t, err)
	})

	t.Run("Should fail to enable question, db fails", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewQuestionService(mockStore, mockRandom, "en-GB")

		ctx := context.Background()

		u := uuid.Must(uuid.NewV7())
		mockStore.EXPECT().EnableQuestion(ctx, u).Return(db.Question{}, fmt.Errorf("failed to enable question"))

		err := srv.EnableQuestion(ctx, u)
		assert.Error(t, err)
	})
}

func TestQuestionAddGroup(t *testing.T) {
	t.Run("Should successfully add new group", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewQuestionService(mockStore, mockRandom, "en-GB")

		ctx := context.Background()

		u := uuid.Must(uuid.NewV7())
		mockRandom.EXPECT().GetID().Return(u)
		mockStore.EXPECT().AddGroup(ctx, db.AddGroupParams{
			ID:        u,
			GroupName: "cat",
			GroupType: "questions",
		}).Return(db.QuestionsGroup{}, nil)

		group, err := srv.AddGroup(ctx, "cat")
		assert.NoError(t, err)
		expectedGroup := service.Group{
			ID:   u.String(),
			Name: "cat",
		}
		assert.Equal(t, expectedGroup, group)
	})

	t.Run("Should fail to add new group, db fails", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewQuestionService(mockStore, mockRandom, "en-GB")

		ctx := context.Background()

		u := uuid.Must(uuid.NewV7())
		mockRandom.EXPECT().GetID().Return(u)
		mockStore.EXPECT().AddGroup(ctx, db.AddGroupParams{
			ID:        u,
			GroupName: "cat",
			GroupType: "questions",
		}).Return(db.QuestionsGroup{}, fmt.Errorf("failed to add group to db"))

		_, err := srv.AddGroup(ctx, "cat")
		assert.Error(t, err)
	})
}
