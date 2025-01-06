package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/hmajid2301/banterbus/internal/service"
	"gitlab.com/hmajid2301/banterbus/internal/service/randomizer"
	db "gitlab.com/hmajid2301/banterbus/internal/store/db"
)

func TestIntegrationQuestionServiceAdd(t *testing.T) {
	t.Run("Should successfully add question", func(t *testing.T) {
		pool, teardown := setupSubtest(t)
		defer teardown()

		str, err := db.NewDB(pool)
		assert.NoError(t, err)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx()
		require.NoError(t, err)

		questionService := service.NewQuestionService(str, randomizer, "en-GB")

		text := "what do you think of cats"
		group := "cat_group"
		roundType := "free_form"

		q, err := questionService.Add(ctx, text, group, roundType)
		assert.NoError(t, err)
		expectedQuestion := service.Question{
			Text:      text,
			GroupName: group,
			Locale:    "en-GB",
			RoundType: roundType,
		}

		assert.Equal(t, expectedQuestion, q)
	})
}

// func TestIntegrationQuestionServiceAddTranslation(t *testing.T) {
// 	t.Run("Should successfully add question translation", func(t *testing.T) {
// 		pool, teardown := setupSubtest(t)
// 		defer teardown()
//
// 		str, err := db.NewDB(pool)
// 		assert.NoError(t, err)
// 		randomizer := randomizer.NewUserRandomizer()
//
// 		ctx, err := getI18nCtx()
// 		require.NoError(t, err)
//
// 		questionService := service.NewQuestionService(str, randomizer, "en-GB")
//
// 		text := "what do you think of cats"
// 		group := "cat_group"
// 		roundType := "free_form"
//
// 		q, err := questionService.Add(ctx, text, group, roundType)
//
// 		textTranslation := "what do you think of cats"
// 		q, err := questionService.AddTranslation(ctx, q.questionID, textTranslation, "pt-PT")
// 		assert.NoError(t, err)
//
// 		expectedQuestion := service.QuestionTranslation{
// 			Text:   text,
// 			Locale: "pt-PT",
// 		}
// 		assert.Equal(t, expectedQuestion, q)
// 	})
// }
