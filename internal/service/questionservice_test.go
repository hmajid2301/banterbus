package service_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/hmajid2301/banterbus/internal/service"
	mockService "gitlab.com/hmajid2301/banterbus/internal/service/mocks"
	"gitlab.com/hmajid2301/banterbus/internal/store/db"
)

func TestQuestionServiceAdd(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully add new question", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewQuestionService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()

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
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewQuestionService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()

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

func TestQuestionServiceAddTranslation(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully add new question translation", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewQuestionService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()

		text := "portguese words"
		locale := "pt-PT"

		var err error
		u, err := uuid.NewV7()
		require.NoError(t, err)
		questionID, err := uuid.NewV7()
		require.NoError(t, err)
		mockRandom.EXPECT().GetID().Return(u, nil)
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
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewQuestionService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()

		text := "portguese words"
		locale := "pt-PT"

		var err error
		u, err := uuid.NewV7()
		require.NoError(t, err)
		questionID, err := uuid.NewV7()
		require.NoError(t, err)
		mockRandom.EXPECT().GetID().Return(u, nil)
		mockStore.EXPECT().AddQuestionTranslation(ctx, db.AddQuestionTranslationParams{
			ID:         u,
			Question:   text,
			Locale:     locale,
			QuestionID: questionID,
		}).Return(db.QuestionsI18n{}, fmt.Errorf("failed to add question translation"))

		_, err = srv.AddTranslation(ctx, questionID, text, locale)
		assert.Error(t, err)
	})
}

func TestQuestionServiceGetGroups(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully get all group", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewQuestionService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()

		var err error
		u, err := uuid.NewV7()
		require.NoError(t, err)
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
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewQuestionService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()

		mockStore.EXPECT().GetGroups(ctx).Return(nil, fmt.Errorf("failed to make request"))
		_, err := srv.GetGroups(ctx)
		assert.Error(t, err)
	})
}

func TestQuestionServiceGetQuestions(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully get questions", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewQuestionService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()

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
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewQuestionService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()

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

func TestQuestionServiceDisableQuestion(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully disable question", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewQuestionService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()

		u, err := uuid.NewV7()
		require.NoError(t, err)
		mockStore.EXPECT().DisableQuestion(ctx, u).Return(db.Question{}, nil)

		err = srv.DisableQuestion(ctx, u)
		assert.NoError(t, err)
	})

	t.Run("Should fail to disable question, db fails", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewQuestionService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()

		u, err := uuid.NewV7()
		require.NoError(t, err)
		mockStore.EXPECT().DisableQuestion(ctx, u).Return(db.Question{}, errors.New("DB failed"))

		err = srv.DisableQuestion(ctx, u)
		assert.Error(t, err)
	})
}

func TestQuestionServiceEnableQuestion(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully enable question", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewQuestionService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()

		u, err := uuid.NewV7()
		require.NoError(t, err)
		mockStore.EXPECT().EnableQuestion(ctx, u).Return(db.Question{}, nil)

		err = srv.EnableQuestion(ctx, u)
		assert.NoError(t, err)
	})

	t.Run("Should fail to enable question, db fails", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewQuestionService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()

		u, err := uuid.NewV7()
		require.NoError(t, err)
		mockStore.EXPECT().EnableQuestion(ctx, u).Return(db.Question{}, fmt.Errorf("failed to enable question"))

		err = srv.EnableQuestion(ctx, u)
		assert.Error(t, err)
	})

}

func TestQuestionServiceAddGroup(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully add new group", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewQuestionService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()

		var err error
		u, err := uuid.NewV7()
		require.NoError(t, err)
		mockRandom.EXPECT().GetID().Return(u, nil)
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
			Type: "questions",
		}
		assert.Equal(t, expectedGroup, group)
	})

	t.Run("Should fail to add new group, db fails", func(t *testing.T) {
		t.Parallel()
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewQuestionService(mockStore, mockRandom, "en-GB")

		ctx := t.Context()

		var err error
		u, err := uuid.NewV7()
		require.NoError(t, err)
		mockRandom.EXPECT().GetID().Return(u, nil)
		mockStore.EXPECT().AddGroup(ctx, db.AddGroupParams{
			ID:        u,
			GroupName: "cat",
			GroupType: "questions",
		}).Return(db.QuestionsGroup{}, fmt.Errorf("failed to add group to db"))

		_, err = srv.AddGroup(ctx, "cat")
		assert.Error(t, err)
	})
}
