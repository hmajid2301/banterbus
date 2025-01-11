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

func TestIntegrationQuestionGetGroupNames(t *testing.T) {
	t.Run("Should successfully get group names", func(t *testing.T) {
		pool, teardown := setupSubtest(t)
		defer teardown()

		str, err := db.NewDB(pool)
		assert.NoError(t, err)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx()
		require.NoError(t, err)

		questionService := service.NewQuestionService(str, randomizer, "en-GB")

		groups, err := questionService.GetGroupNames(ctx)
		assert.NoError(t, err)
		expectedGroups := []string{
			"programming_group",
			"horse_group",
			"colour_group",
			"cat_group",
			"bike_group",
			"animal_group",
			"all",
		}
		assert.Equal(t, expectedGroups, groups)
	})
}

func TestIntegrationQuestionGetQuestions(t *testing.T) {
	t.Run("Should successfully get questions with no filters", func(t *testing.T) {
		pool, teardown := setupSubtest(t)
		defer teardown()

		str, err := db.NewDB(pool)
		assert.NoError(t, err)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx()
		require.NoError(t, err)

		questionService := service.NewQuestionService(str, randomizer, "en-GB")

		filters := service.GetQuestionFilters{}
		questions, err := questionService.GetQuestions(ctx, filters, 100, 1)
		assert.NoError(t, err)

		expectedQuestions := []service.Question{
			{
				Text:      "to get arrested",
				GroupName: "all",
				Locale:    "en-GB",
				RoundType: "most_likely",
			},
			{
				Text:      "to eat ice-cream from the tub",
				GroupName: "all",
				Locale:    "en-GB",
				RoundType: "most_likely",
			},
			{
				Text:      "to fight a police person",
				GroupName: "all",
				Locale:    "en-GB",
				RoundType: "most_likely",
			},
			{
				Text:      "to fight a horse",
				GroupName: "all",
				Locale:    "en-GB",
				RoundType: "most_likely",
			},
			{
				Text:      "What do you think about programmers?",
				GroupName: "programming_group",
				Locale:    "en-GB",
				RoundType: "free_form",
			},
			{
				Text:      "What don't you like about programmers?",
				GroupName: "programming_group",
				Locale:    "en-GB",
				RoundType: "free_form",
			},
			{
				Text:      "what don't you think about programmers?",
				GroupName: "programming_group",
				Locale:    "en-GB",
				RoundType: "free_form",
			},
			{
				Text:      "what dont you think about cats",
				GroupName: "cat_group",
				Locale:    "en-GB",
				RoundType: "free_form",
			},
			{
				Text:      "what don't you like about cats?",
				GroupName: "cat_group",
				Locale:    "en-GB",
				RoundType: "free_form",
			},
			{
				Text:      "what do you like about cats?",
				GroupName: "cat_group",
				Locale:    "en-GB",
				RoundType: "free_form",
			},
			{
				Text:      "what do you think about cats",
				GroupName: "cat_group",
				Locale:    "en-GB",
				RoundType: "free_form",
			},
			{
				Text:      "A funny question?",
				GroupName: "bike_group",
				Locale:    "en-GB",
				RoundType: "free_form",
			},
			{
				Text:      "Favourite bike colour?",
				GroupName: "bike_group",
				Locale:    "en-GB",
				RoundType: "free_form",
			},
			{
				Text:      "What do you think about camels?",
				GroupName: "horse_group",
				Locale:    "en-GB",
				RoundType: "multiple_choice",
			},
			{
				Text:      "What do you think about horses?",
				GroupName: "horse_group",
				Locale:    "en-GB",
				RoundType: "multiple_choice",
			},
			{
				Text:      "What is your favourite colour?",
				GroupName: "colour_group",
				Locale:    "en-GB",
				RoundType: "multiple_choice",
			},
			{
				Text:      "What is your least favourite colour?",
				GroupName: "colour_group",
				Locale:    "en-GB",
				RoundType: "multiple_choice",
			},
			{
				Text:      "Strongly Agree",
				GroupName: "animal_group",
				Locale:    "en-GB",
				RoundType: "multiple_choice",
			},
			{
				Text:      "Agree",
				GroupName: "animal_group",
				Locale:    "en-GB",
				RoundType: "multiple_choice",
			},
			{
				Text:      "Disagree",
				GroupName: "animal_group",
				Locale:    "en-GB",
				RoundType: "multiple_choice",
			},
			{
				Text:      "Are cats cute?",
				GroupName: "animal_group",
				Locale:    "en-GB",
				RoundType: "multiple_choice",
			},
			{
				Text:      "Dogs are cuter than cats?",
				GroupName: "animal_group",
				Locale:    "en-GB",
				RoundType: "multiple_choice",
			},
		}
		assert.Equal(t, expectedQuestions, questions)
	})

	t.Run("Should successfully get questions with group filter", func(t *testing.T) {
		pool, teardown := setupSubtest(t)
		defer teardown()

		str, err := db.NewDB(pool)
		assert.NoError(t, err)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx()
		require.NoError(t, err)

		questionService := service.NewQuestionService(str, randomizer, "en-GB")

		filters := service.GetQuestionFilters{
			GroupName: "animal_group",
		}
		questions, err := questionService.GetQuestions(ctx, filters, 100, 1)
		assert.NoError(t, err)

		expectedQuestions := []service.Question{
			{
				Text:      "Strongly Agree",
				GroupName: "animal_group",
				Locale:    "en-GB",
				RoundType: "multiple_choice",
			},
			{
				Text:      "Agree",
				GroupName: "animal_group",
				Locale:    "en-GB",
				RoundType: "multiple_choice",
			},
			{
				Text:      "Disagree",
				GroupName: "animal_group",
				Locale:    "en-GB",
				RoundType: "multiple_choice",
			},
			{
				Text:      "Are cats cute?",
				GroupName: "animal_group",
				Locale:    "en-GB",
				RoundType: "multiple_choice",
			},
			{
				Text:      "Dogs are cuter than cats?",
				GroupName: "animal_group",
				Locale:    "en-GB",
				RoundType: "multiple_choice",
			},
		}
		assert.Equal(t, expectedQuestions, questions)
	})

	t.Run("Should successfully get questions with round type filter", func(t *testing.T) {
		pool, teardown := setupSubtest(t)
		defer teardown()

		str, err := db.NewDB(pool)
		assert.NoError(t, err)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx()
		require.NoError(t, err)

		questionService := service.NewQuestionService(str, randomizer, "en-GB")

		filters := service.GetQuestionFilters{
			RoundType: "free_form",
		}
		questions, err := questionService.GetQuestions(ctx, filters, 100, 1)
		assert.NoError(t, err)

		expectedQuestions := []service.Question{
			{
				Text:      "What do you think about programmers?",
				GroupName: "programming_group",
				Locale:    "en-GB",
				RoundType: "free_form",
			},
			{
				Text:      "What don't you like about programmers?",
				GroupName: "programming_group",
				Locale:    "en-GB",
				RoundType: "free_form",
			},
			{
				Text:      "what don't you think about programmers?",
				GroupName: "programming_group",
				Locale:    "en-GB",
				RoundType: "free_form",
			},
			{
				Text:      "what dont you think about cats",
				GroupName: "cat_group",
				Locale:    "en-GB",
				RoundType: "free_form",
			},
			{
				Text:      "what don't you like about cats?",
				GroupName: "cat_group",
				Locale:    "en-GB",
				RoundType: "free_form",
			},
			{
				Text:      "what do you like about cats?",
				GroupName: "cat_group",
				Locale:    "en-GB",
				RoundType: "free_form",
			},
			{
				Text:      "what do you think about cats",
				GroupName: "cat_group",
				Locale:    "en-GB",
				RoundType: "free_form",
			},
			{
				Text:      "A funny question?",
				GroupName: "bike_group",
				Locale:    "en-GB",
				RoundType: "free_form",
			},
			{
				Text:      "Favourite bike colour?",
				GroupName: "bike_group",
				Locale:    "en-GB",
				RoundType: "free_form",
			},
		}
		assert.Equal(t, expectedQuestions, questions)
	})

	t.Run("Should successfully get questions with all the filters", func(t *testing.T) {
		pool, teardown := setupSubtest(t)
		defer teardown()

		str, err := db.NewDB(pool)
		assert.NoError(t, err)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx()
		require.NoError(t, err)

		questionService := service.NewQuestionService(str, randomizer, "en-GB")

		filters := service.GetQuestionFilters{
			GroupName: "cat_group",
			RoundType: "free_form",
		}
		questions, err := questionService.GetQuestions(ctx, filters, 100, 1)
		assert.NoError(t, err)

		expectedQuestions := []service.Question{
			{
				Text:      "what dont you think about cats",
				GroupName: "cat_group",
				Locale:    "en-GB",
				RoundType: "free_form",
			},
			{
				Text:      "what don't you like about cats?",
				GroupName: "cat_group",
				Locale:    "en-GB",
				RoundType: "free_form",
			},
			{
				Text:      "what do you like about cats?",
				GroupName: "cat_group",
				Locale:    "en-GB",
				RoundType: "free_form",
			},
			{
				Text:      "what do you think about cats",
				GroupName: "cat_group",
				Locale:    "en-GB",
				RoundType: "free_form",
			},
		}
		assert.Equal(t, expectedQuestions, questions)
	})

	t.Run("Should successfully get questions with limit", func(t *testing.T) {
		pool, teardown := setupSubtest(t)
		defer teardown()

		str, err := db.NewDB(pool)
		assert.NoError(t, err)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx()
		require.NoError(t, err)

		questionService := service.NewQuestionService(str, randomizer, "en-GB")

		filters := service.GetQuestionFilters{}
		questions, err := questionService.GetQuestions(ctx, filters, 1, 1)
		assert.NoError(t, err)

		expectedQuestions := []service.Question{

			{
				Text:      "to get arrested",
				GroupName: "all",
				Locale:    "en-GB",
				RoundType: "most_likely",
			},
		}
		assert.Equal(t, expectedQuestions, questions)
	})

	// t.Run("Should successfully get questions with limit and offset", func(t *testing.T) {
	// 	pool, teardown := setupSubtest(t)
	// 	defer teardown()
	//
	// 	str, err := db.NewDB(pool)
	// 	assert.NoError(t, err)
	// 	randomizer := randomizer.NewUserRandomizer()
	//
	// 	ctx, err := getI18nCtx()
	// 	require.NoError(t, err)
	//
	// 	questionService := service.NewQuestionService(str, randomizer, "en-GB")
	//
	// 	filters := service.GetQuestionFilters{}
	// 	questions, err := questionService.GetQuestions(ctx, filters, 1, 1)
	// 	assert.NoError(t, err)
	//
	// 	expectedQuestions := []service.Question{
	// 		{
	// 			Text:      "to eat ice-cream from the tub",
	// 			GroupName: "all",
	// 			Locale:    "en-GB",
	// 			RoundType: "most_likely",
	// 		},
	// 	}
	// 	assert.Equal(t, expectedQuestions, questions)
	// })
}
