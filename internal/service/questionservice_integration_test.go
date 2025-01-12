package service_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/hmajid2301/banterbus/internal/service"
	"gitlab.com/hmajid2301/banterbus/internal/service/randomizer"
	db "gitlab.com/hmajid2301/banterbus/internal/store/db"
)

var defaultText = "what do you think of cats"
var defaultGroup = "cat_group"
var defaultRoundType = "free_form"

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

		q, err := questionService.Add(ctx, defaultText, defaultGroup, defaultRoundType)
		assert.NoError(t, err)
		expectedQuestion := service.Question{
			ID:        q.ID,
			Text:      defaultText,
			GroupName: defaultGroup,
			Locale:    "en-GB",
			RoundType: defaultRoundType,
			Enabled:   true,
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

		groups, err := questionService.GetGroups(ctx)
		assert.NoError(t, err)

		// TODO: fix this with questions vs answers
		expectedGroups := []service.Group{
			{ID: "01945c66-891a-7894-ae92-c18087c73a23", Name: "programming_group"},
			{ID: "01945c66-891b-797e-820b-17c4cc98a566", Name: "programming_group"},
			{ID: "01945c66-891c-7942-9a2a-339a62a74800", Name: "horse_group"},
			{ID: "01945c66-891c-7794-ab47-8f70974b0037", Name: "horse_group"},
			{ID: "01945c66-891c-7aa2-b6ca-088679706a5b", Name: "colour_group"},
			{ID: "01945c66-891c-7be5-9382-7b6399a6b09b", Name: "colour_group"},
			{ID: "01945c66-891b-7d3e-804c-f2e170b0b0ce", Name: "cat_group"},
			{ID: "01945c66-891b-7bb1-875f-a2be41d430ab", Name: "cat_group"},
			{ID: "01945c66-891c-74d5-9870-7a8777e37588", Name: "bike_group"},
			{ID: "01945c66-891c-717f-a177-3240919b638f", Name: "bike_group"},
			{ID: "01945c66-891c-7d8a-b404-be384c9515a6", Name: "animal_group"},
			{ID: "01945c66-891c-7ee0-88fb-5703de2a1b59", Name: "animal_group"},
			{ID: "01945c66-891d-723b-8fc9-b373d799e95b", Name: "all"},
			{ID: "01945c66-891d-70fc-945a-e30fa7b09bfb", Name: "all"},
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
				ID:        "4b1355bb-82de-40c8-8eda-0c634091cc3c",
				Text:      "to get arrested",
				GroupName: "all",
				Locale:    "en-GB",
				RoundType: "most_likely",
				Enabled:   true,
			},
			{
				ID:        "a91af98c-f989-4e00-aa14-7a34e732519e",
				Text:      "to eat ice-cream from the tub",
				GroupName: "all",
				Locale:    "en-GB",
				RoundType: "most_likely",
				Enabled:   true,
			},
			{
				ID:        "fac6a98f-e3b5-4328-999c-b39fd86657ba",
				Text:      "to fight a police person",
				GroupName: "all",
				Locale:    "en-GB",
				RoundType: "most_likely",
				Enabled:   true,
			},
			{
				ID:        "6b60f097-b714-4f9e-b8cb-de75a7890381",
				Text:      "to fight a horse",
				GroupName: "all",
				Locale:    "en-GB",
				RoundType: "most_likely",
				Enabled:   true,
			},
			{
				ID:        "93dd56a8-c8a3-4c63-93dc-9d890c4d2b74",
				Text:      "What do you think about programmers?",
				GroupName: "programming_group",
				Locale:    "en-GB",
				RoundType: "free_form",
				Enabled:   true,
			},
			{
				ID:        "066e7a8a-b0b7-44d4-b882-582a64151c15",
				Text:      "What don't you like about programmers?",
				GroupName: "programming_group",
				Locale:    "en-GB",
				RoundType: "free_form",
				Enabled:   true,
			},
			{
				ID:        "654327b9-36a2-4d75-b4bf-d68d19fcfe7c",
				Text:      "what don't you think about programmers?",
				GroupName: "programming_group",
				Locale:    "en-GB",
				RoundType: "free_form",
				Enabled:   true,
			},
			{
				ID:        "281bc3c7-f55d-4a8a-88cf-4e0d67d2825e",
				Text:      "what dont you think about cats",
				GroupName: "cat_group",
				Locale:    "en-GB",
				RoundType: "free_form",
				Enabled:   true,
			},
			{
				ID:        "fc1a3c9f-3d98-452e-b77e-c6c7f353176d",
				Text:      "what don't you like about cats?",
				GroupName: "cat_group",
				Locale:    "en-GB",
				RoundType: "free_form",
				Enabled:   true,
			},
			{
				ID:        "393dae17-84fe-449d-ba0f-8c9d320a46e6",
				Text:      "what do you like about cats?",
				GroupName: "cat_group",
				Locale:    "en-GB",
				RoundType: "free_form",
				Enabled:   true,
			},
			{
				ID:        "393dae17-84fe-449d-ba0f-8c9d320a46e7",
				Text:      "what do you think about cats",
				GroupName: "cat_group",
				Locale:    "en-GB",
				RoundType: "free_form",
				Enabled:   true,
			},
			{
				ID:        "9347fb72-77f9-44c4-9a7c-27109e29dd97",
				Text:      "A funny question?",
				GroupName: "bike_group",
				Locale:    "en-GB",
				RoundType: "free_form",
				Enabled:   true,
			},
			{
				ID:        "8aa9f87f-31d9-4421-aae5-2024ca730348",
				Text:      "Favourite bike colour?",
				GroupName: "bike_group",
				Locale:    "en-GB",
				RoundType: "free_form",
				Enabled:   true,
			},
			{
				ID:        "89b20c84-12ae-444d-ad9c-26f72d3f28ab",
				Text:      "What do you think about camels?",
				GroupName: "horse_group",
				Locale:    "en-GB",
				RoundType: "multiple_choice",
				Enabled:   true,
			},
			{
				ID:        "68ed9133-dc58-41bb-b642-c48470998127",
				Text:      "What do you think about horses?",
				GroupName: "horse_group",
				Locale:    "en-GB",
				RoundType: "multiple_choice",
				Enabled:   true,
			},
			{
				ID:        "56dae1b2-06e9-4339-a2b0-892a18444e15",
				Text:      "What is your favourite colour?",
				GroupName: "colour_group",
				Locale:    "en-GB",
				RoundType: "multiple_choice",
				Enabled:   true,
			},
			{
				ID:        "8a76fbb3-c9ad-47b2-a195-9d8623ab8da0",
				Text:      "What is your least favourite colour?",
				GroupName: "colour_group",
				Locale:    "en-GB",
				RoundType: "multiple_choice",
				Enabled:   true,
			},
			{
				ID:        "cb9b99be-e66f-4f1b-9a62-64415a824b31",
				Text:      "Strongly Agree",
				GroupName: "animal_group",
				Locale:    "en-GB",
				RoundType: "multiple_choice",
				Enabled:   true,
			},
			{
				ID:        "e72b7503-88ea-440e-a42c-bc6f2f444b08",
				Text:      "Agree",
				GroupName: "animal_group",
				Locale:    "en-GB",
				RoundType: "multiple_choice",
				Enabled:   true,
			},
			{
				ID:        "fa459292-bca0-47ec-81f1-5fb48036f3ea",
				Text:      "Disagree",
				GroupName: "animal_group",
				Locale:    "en-GB",
				RoundType: "multiple_choice",
				Enabled:   true,
			},
			{
				ID:        "e90d613d-2e6c-4331-9204-9b685c0795b7",
				Text:      "Are cats cute?",
				GroupName: "animal_group",
				Locale:    "en-GB",
				RoundType: "multiple_choice",
				Enabled:   true,
			},
			{
				ID:        "89deb03f-66be-4265-91e6-dedd9227718a",
				Text:      "Dogs are cuter than cats?",
				GroupName: "animal_group",
				Locale:    "en-GB",
				RoundType: "multiple_choice",
				Enabled:   true,
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
				ID:        "cb9b99be-e66f-4f1b-9a62-64415a824b31",
				Text:      "Strongly Agree",
				GroupName: "animal_group",
				Locale:    "en-GB",
				RoundType: "multiple_choice",
				Enabled:   true,
			},
			{
				ID:        "e72b7503-88ea-440e-a42c-bc6f2f444b08",
				Text:      "Agree",
				GroupName: "animal_group",
				Locale:    "en-GB",
				RoundType: "multiple_choice",
				Enabled:   true,
			},
			{
				ID:        "fa459292-bca0-47ec-81f1-5fb48036f3ea",
				Text:      "Disagree",
				GroupName: "animal_group",
				Locale:    "en-GB",
				RoundType: "multiple_choice",
				Enabled:   true,
			},
			{
				ID:        "e90d613d-2e6c-4331-9204-9b685c0795b7",
				Text:      "Are cats cute?",
				GroupName: "animal_group",
				Locale:    "en-GB",
				RoundType: "multiple_choice",
				Enabled:   true,
			},
			{
				ID:        "89deb03f-66be-4265-91e6-dedd9227718a",
				Text:      "Dogs are cuter than cats?",
				GroupName: "animal_group",
				Locale:    "en-GB",
				RoundType: "multiple_choice",
				Enabled:   true,
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
				ID:        "93dd56a8-c8a3-4c63-93dc-9d890c4d2b74",
				Text:      "What do you think about programmers?",
				GroupName: "programming_group",
				Locale:    "en-GB",
				RoundType: "free_form",
				Enabled:   true,
			},
			{
				ID:        "066e7a8a-b0b7-44d4-b882-582a64151c15",
				Text:      "What don't you like about programmers?",
				GroupName: "programming_group",
				Locale:    "en-GB",
				RoundType: "free_form",
				Enabled:   true,
			},
			{
				ID:        "654327b9-36a2-4d75-b4bf-d68d19fcfe7c",
				Text:      "what don't you think about programmers?",
				GroupName: "programming_group",
				Locale:    "en-GB",
				RoundType: "free_form",
				Enabled:   true,
			},
			{
				ID:        "281bc3c7-f55d-4a8a-88cf-4e0d67d2825e",
				Text:      "what dont you think about cats",
				GroupName: "cat_group",
				Locale:    "en-GB",
				RoundType: "free_form",
				Enabled:   true,
			},
			{
				ID:        "fc1a3c9f-3d98-452e-b77e-c6c7f353176d",
				Text:      "what don't you like about cats?",
				GroupName: "cat_group",
				Locale:    "en-GB",
				RoundType: "free_form",
				Enabled:   true,
			},
			{
				ID:        "393dae17-84fe-449d-ba0f-8c9d320a46e6",
				Text:      "what do you like about cats?",
				GroupName: "cat_group",
				Locale:    "en-GB",
				RoundType: "free_form",
				Enabled:   true,
			},
			{
				ID:        "393dae17-84fe-449d-ba0f-8c9d320a46e7",
				Text:      "what do you think about cats",
				GroupName: "cat_group",
				Locale:    "en-GB",
				RoundType: "free_form",
				Enabled:   true,
			},
			{
				ID:        "9347fb72-77f9-44c4-9a7c-27109e29dd97",
				Text:      "A funny question?",
				GroupName: "bike_group",
				Locale:    "en-GB",
				RoundType: "free_form",
				Enabled:   true,
			},
			{
				ID:        "8aa9f87f-31d9-4421-aae5-2024ca730348",
				Text:      "Favourite bike colour?",
				GroupName: "bike_group",
				Locale:    "en-GB",
				RoundType: "free_form",
				Enabled:   true,
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
				ID:        "281bc3c7-f55d-4a8a-88cf-4e0d67d2825e",
				Text:      "what dont you think about cats",
				GroupName: "cat_group",
				Locale:    "en-GB",
				RoundType: "free_form",
				Enabled:   true,
			},
			{
				ID:        "fc1a3c9f-3d98-452e-b77e-c6c7f353176d",
				Text:      "what don't you like about cats?",
				GroupName: "cat_group",
				Locale:    "en-GB",
				RoundType: "free_form",
				Enabled:   true,
			},
			{
				ID:        "393dae17-84fe-449d-ba0f-8c9d320a46e6",
				Text:      "what do you like about cats?",
				GroupName: "cat_group",
				Locale:    "en-GB",
				RoundType: "free_form",
				Enabled:   true,
			},
			{
				ID:        "393dae17-84fe-449d-ba0f-8c9d320a46e7",
				Text:      "what do you think about cats",
				GroupName: "cat_group",
				Locale:    "en-GB",
				RoundType: "free_form",
				Enabled:   true,
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
				ID:        questions[0].ID,
				Text:      "to get arrested",
				GroupName: "all",
				Locale:    "en-GB",
				RoundType: "most_likely",
				Enabled:   true,
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

func TestIntegrationQuestionAddGroup(t *testing.T) {
	t.Run("Should successfully add group", func(t *testing.T) {
		pool, teardown := setupSubtest(t)
		defer teardown()

		str, err := db.NewDB(pool)
		assert.NoError(t, err)
		randomizer := randomizer.NewUserRandomizer()

		ctx, err := getI18nCtx()
		require.NoError(t, err)

		questionService := service.NewQuestionService(str, randomizer, "en-GB")

		group, err := questionService.AddGroup(ctx, "cat")
		assert.NoError(t, err)
		expectedGroup := service.Group{
			ID:   group.ID,
			Name: "cat",
		}
		assert.Equal(t, expectedGroup, group)
	})
}

func TestIntegrationQuestionDisableQuestion(t *testing.T) {
	t.Run("Should successfully disable question", func(t *testing.T) {
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
		require.NoError(t, err)

		err = questionService.DisableQuestion(ctx, uuid.MustParse(q.ID))
		assert.NoError(t, err)
	})
}

func TestIntegrationQuestionEnableQuestion(t *testing.T) {
	t.Run("Should successfully enable question", func(t *testing.T) {
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
		require.NoError(t, err)

		err = questionService.DisableQuestion(ctx, uuid.MustParse(q.ID))
		require.NoError(t, err)

		err = questionService.EnableQuestion(ctx, uuid.MustParse(q.ID))
		assert.NoError(t, err)
	})
}
