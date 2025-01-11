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
		}).Return(nil)

		question, err := srv.Add(ctx, text, groupName, roundType)
		assert.NoError(t, err)

		expectedQuestion := service.Question{

			Text:      text,
			GroupName: groupName,
			Locale:    "en-GB",
			RoundType: roundType,
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
		}).Return(fmt.Errorf("failed to create question"))

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

func TestQuestionGetGroupNames(t *testing.T) {
	t.Run("Should successfully get all group names", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewQuestionService(mockStore, mockRandom, "en-GB")

		ctx := context.Background()

		expectedGroups := []string{"cat", "cricket", "programming"}
		mockStore.EXPECT().GetGroupNames(ctx).Return(expectedGroups, nil)

		groups, err := srv.GetGroupNames(ctx)
		assert.NoError(t, err)
		assert.Equal(t, expectedGroups, groups)
	})

	t.Run("Should fail get all group names, request to DB fails", func(t *testing.T) {
		mockStore := mockService.NewMockStorer(t)
		mockRandom := mockService.NewMockRandomizer(t)
		srv := service.NewQuestionService(mockStore, mockRandom, "en-GB")

		ctx := context.Background()

		mockStore.EXPECT().GetGroupNames(ctx).Return(nil, fmt.Errorf("failed to make request"))
		_, err := srv.GetGroupNames(ctx)
		assert.Error(t, err)
	})
}
