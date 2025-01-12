package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/hmajid2301/banterbus/internal/banterbustest"
	transporthttp "gitlab.com/hmajid2301/banterbus/internal/transport/http"
)

func TestIntegrationAddQuestionHandler(t *testing.T) {
	srv, err := banterbustest.NewTestServer()
	require.NoError(t, err)
	defer srv.Close()

	t.Run("Should successfully add new question", func(t *testing.T) {
		question := transporthttp.NewQuestion{
			Text:      "Do you like cats",
			GroupName: "cat_group",
			RoundType: "free_form",
		}

		jsonData, err := json.Marshal(question)
		require.NoError(t, err)

		ctx := context.Background()
		endpoint := fmt.Sprintf("%s/question", srv.URL)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(jsonData))
		require.NoError(t, err)

		client := &http.Client{}
		resp, err := client.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("Should fail to add new question, invalid method", func(t *testing.T) {
		question := map[string]string{
			"text": "do",
		}

		jsonData, err := json.Marshal(question)
		require.NoError(t, err)

		ctx := context.Background()
		endpoint := fmt.Sprintf("%s/question", srv.URL)
		req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, bytes.NewReader(jsonData))
		require.NoError(t, err)

		client := &http.Client{}
		resp, err := client.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})

	t.Run("Should fail to add new question, missing required fields", func(t *testing.T) {
		question := map[string]string{
			"text": "do",
		}

		jsonData, err := json.Marshal(question)
		require.NoError(t, err)

		ctx := context.Background()
		endpoint := fmt.Sprintf("%s/question", srv.URL)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(jsonData))
		require.NoError(t, err)

		client := &http.Client{}
		resp, err := client.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

// func TestIntegrationAddQuestionTranslationHandler(t *testing.T) {
// 	srv, err := banterbustest.NewTestServer()
// 	require.NoError(t, err)
// 	defer srv.Close()
//
// 	t.Run("Should successfully add new question translation", func(t *testing.T) {
// 		question := transporthttp.NewQuestionTranslation{
// 			Text: "Do you like cattos",
// 		}
//
// 		jsonData, err := json.Marshal(question)
// 		require.NoError(t, err)
//
// 		ctx := context.Background()
// 		endpoint := fmt.Sprintf("%s/question/%s/locale/%s", srv.URL, "", "pt-PT")
// 		req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, bytes.NewReader(jsonData))
// 		require.NoError(t, err)
//
// 		client := &http.Client{}
// 		resp, err := client.Do(req)
// 		require.NoError(t, err)
// 		defer resp.Body.Close()
//
// 		require.Equal(t, http.StatusCreated, resp.StatusCode)
// 	})
//
// 	t.Run("Should fail to add new question, invalid method", func(t *testing.T) {
// 		question := map[string]string{
// 			"text": "do",
// 		}
//
// 		jsonData, err := json.Marshal(question)
// 		require.NoError(t, err)
//
// 		ctx := context.Background()
// 		endpoint := fmt.Sprintf("%s/question", srv.URL)
// 		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, bytes.NewReader(jsonData))
// 		require.NoError(t, err)
//
// 		client := &http.Client{}
// 		resp, err := client.Do(req)
// 		require.NoError(t, err)
// 		defer resp.Body.Close()
//
// 		require.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
// 	})
//
// 	t.Run("Should fail to add new question, missing required fields", func(t *testing.T) {
// 		question := map[string]string{
// 			"text": "do",
// 		}
//
// 		jsonData, err := json.Marshal(question)
// 		require.NoError(t, err)
//
// 		ctx := context.Background()
// 		endpoint := fmt.Sprintf("%s/question", srv.URL)
// 		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(jsonData))
// 		require.NoError(t, err)
//
// 		client := &http.Client{}
// 		resp, err := client.Do(req)
// 		require.NoError(t, err)
// 		defer resp.Body.Close()
//
// 		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
// 	})
// }

// func TestIntegrationGetGroupsHandler(t *testing.T) {
// 	srv, err := banterbustest.NewTestServer()
// 	require.NoError(t, err)
// 	defer srv.Close()
//
// 	t.Run("Should successfully get group", func(t *testing.T) {
// 		ctx := context.Background()
// 		endpoint := fmt.Sprintf("%s/question/group", srv.URL)
// 		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, http.NoBody)
// 		require.NoError(t, err)
//
// 		client := &http.Client{}
// 		resp, err := client.Do(req)
// 		assert.NoError(t, err)
// 		defer resp.Body.Close()
//
// 		assert.Equal(t, http.StatusOK, resp.StatusCode)
// 		var groups transporthttp.Group
// 		err = json.NewDecoder(resp.Body).Decode(&groups)
// 		assert.NoError(t, err)
//
// 		expectedGroups := transporthttp.Group{
// 			Groups: []service.Group{
// 				{Name: "programming_group"},
// 				{Name: "programming_group"},
// 				{Name: "horse_group"},
// 				{Name: "horse_group"},
// 				{Name: "colour_group"},
// 				{Name: "colour_group"},
// 				{Name: "cat_group"},
// 				{Name: "cat_group"},
// 				{Name: "bike_group"},
// 				{Name: "bike_group"},
// 				{Name: "animal_group"},
// 				{Name: "animal_group"},
// 				{Name: "all"},
// 				{Name: "all"},
// 			},
// 		}
// 		assert.Equal(t, expectedGroups, groups)
// 	})
// }

func TestIntegrationGetQuestionsHandler(t *testing.T) {
	srv, err := banterbustest.NewTestServer()
	require.NoError(t, err)
	defer srv.Close()

	// TODO: enable with pagination
	// t.Run("Should successfully get handler", func(t *testing.T) {
	// 	ctx := context.Background()
	// 	endpoint := fmt.Sprintf("%s/question", srv.URL)
	// 	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, http.NoBody)
	// 	require.NoError(t, err)
	//
	// 	client := &http.Client{}
	// 	resp, err := client.Do(req)
	// 	assert.NoError(t, err)
	// 	defer resp.Body.Close()
	//
	// 	assert.Equal(t, http.StatusOK, resp.StatusCode)
	// 	var questions []service.Question
	// 	err = json.NewDecoder(resp.Body).Decode(&questions)
	// 	assert.NoError(t, err)
	//
	// 	expectedQuestions := []service.Question{}
	// 	assert.ElementsMatch(t, expectedQuestions, questions)
	// })
}

//  TODO: add more tests for rest of handlers
