package service_test

import (
	"context"
	"fmt"
	"testing"

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
		}).Return(fmt.Errorf("failed to create question"))

		_, err := srv.Add(ctx, text, groupName, roundType)
		assert.Error(t, err)
	})
}
