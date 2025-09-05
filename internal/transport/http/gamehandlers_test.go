package http_test

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/invopop/ctxi18n"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/hmajid2301/banterbus/internal/service"
	httpTransport "gitlab.com/hmajid2301/banterbus/internal/transport/http"
	"gitlab.com/hmajid2301/banterbus/internal/views"
)

type mockWebsocketer struct {
	subscribeErr error
}

func (m *mockWebsocketer) Subscribe(r *http.Request, w http.ResponseWriter) error {
	return m.subscribeErr
}

type mockQuestionServicer struct {
	questions []service.Question
	groups    []service.Group
	addErr    error
	getErr    error
}

func (m *mockQuestionServicer) Add(
	ctx context.Context,
	text string,
	group string,
	roundType string,
) (service.Question, error) {
	if m.addErr != nil {
		return service.Question{}, m.addErr
	}
	q := service.Question{
		ID:        uuid.Must(uuid.NewV4()).String(),
		Text:      text,
		GroupName: group,
		RoundType: roundType,
		Enabled:   true,
	}
	m.questions = append(m.questions, q)
	return q, nil
}

func (m *mockQuestionServicer) AddTranslation(
	ctx context.Context,
	questionID uuid.UUID,
	text string,
	locale string,
) (service.QuestionTranslation, error) {
	if m.addErr != nil {
		return service.QuestionTranslation{}, m.addErr
	}
	return service.QuestionTranslation{
		Text:   text,
		Locale: locale,
	}, nil
}

func (m *mockQuestionServicer) GetQuestions(
	ctx context.Context,
	filters service.GetQuestionFilters,
	limit int32,
	pageNum int32,
) ([]service.Question, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.questions, nil
}

func (m *mockQuestionServicer) DisableQuestion(ctx context.Context, id uuid.UUID) error {
	return m.addErr
}

func (m *mockQuestionServicer) EnableQuestion(ctx context.Context, id uuid.UUID) error {
	return m.addErr
}

func (m *mockQuestionServicer) AddGroup(ctx context.Context, name string, groupType ...string) (service.Group, error) {
	if m.addErr != nil {
		return service.Group{}, m.addErr
	}
	g := service.Group{
		ID:   uuid.Must(uuid.NewV4()).String(),
		Name: name,
	}
	if len(groupType) > 0 {
		g.Type = groupType[0]
	}
	m.groups = append(m.groups, g)
	return g, nil
}

func (m *mockQuestionServicer) GetGroups(ctx context.Context) ([]service.Group, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.groups, nil
}

func setupGameHandlersTest(t *testing.T) (*httptest.Server, *httpTransport.Server) {
	loc := i18n.Code("en-GB")
	err := ctxi18n.LoadWithDefault(views.Locales, loc)
	require.NoError(t, err)

	mockWS := &mockWebsocketer{}
	mockQS := &mockQuestionServicer{}

	// Create a test logger to avoid nil pointer issues
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	server := httpTransport.NewServer(
		mockWS,
		logger,
		http.Dir("../../../static"),
		nil,
		mockQS,
		httpTransport.ServerConfig{
			Host:          "localhost",
			Port:          8080,
			Environment:   "test",
			DefaultLocale: loc,
			AuthDisabled:  true,
		},
	)

	testServer := httptest.NewServer(server.Server.Handler)
	t.Cleanup(testServer.Close)

	return testServer, server
}

func TestGameHandlerIndex(t *testing.T) {
	t.Parallel()

	t.Run("Should return HTML page successfully", func(t *testing.T) {
		t.Parallel()
		testServer, _ := setupGameHandlersTest(t)

		resp, err := http.Get(testServer.URL + "/")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("Content-Type"), "text/html")
	})
}

func TestGameHandlerJoin(t *testing.T) {
	t.Parallel()

	t.Run("Should return HTML page for valid room code", func(t *testing.T) {
		t.Parallel()
		testServer, _ := setupGameHandlersTest(t)

		resp, err := http.Get(testServer.URL + "/join/ABC12")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("Content-Type"), "text/html")
	})

	t.Run("Should return HTML page for another valid room code", func(t *testing.T) {
		t.Parallel()
		testServer, _ := setupGameHandlersTest(t)

		resp, err := http.Get(testServer.URL + "/join/XYZ99")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("Content-Type"), "text/html")
	})

	t.Run("Should return HTML page for empty room code", func(t *testing.T) {
		t.Parallel()
		testServer, _ := setupGameHandlersTest(t)

		resp, err := http.Get(testServer.URL + "/join/")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("Content-Type"), "text/html")
	})
}

func TestGameHandlerSubscribe(t *testing.T) {
	t.Parallel()

	t.Run("Should handle successful WebSocket subscription", func(t *testing.T) {
		t.Parallel()
		loc := i18n.Code("en-GB")
		err := ctxi18n.LoadWithDefault(views.Locales, loc)
		require.NoError(t, err)

		mockWS := &mockWebsocketer{subscribeErr: nil}
		mockQS := &mockQuestionServicer{}
		logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

		server := httpTransport.NewServer(
			mockWS,
			logger,
			http.Dir("../../../static"),
			nil,
			mockQS,
			httpTransport.ServerConfig{
				Host:          "localhost",
				Port:          8080,
				Environment:   "test",
				DefaultLocale: loc,
				AuthDisabled:  true,
			},
		)

		req := httptest.NewRequest("GET", "/ws", nil)
		req.Header.Set("Connection", "upgrade")
		req.Header.Set("Upgrade", "websocket")
		req.Header.Set("Sec-WebSocket-Version", "13")
		req.Header.Set("Sec-WebSocket-Key", "test-key")

		w := httptest.NewRecorder()
		server.Server.Handler.ServeHTTP(w, req)

		// Note: httptest.NewRecorder cannot properly test WebSocket upgrades.
		// Since the mock websocketer doesn't perform an actual upgrade,
		// we get a regular HTTP response (200) instead of switching protocols (101).
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Should handle WebSocket subscription error", func(t *testing.T) {
		t.Parallel()
		loc := i18n.Code("en-GB")
		err := ctxi18n.LoadWithDefault(views.Locales, loc)
		require.NoError(t, err)

		mockWS := &mockWebsocketer{subscribeErr: assert.AnError}
		mockQS := &mockQuestionServicer{}
		logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

		server := httpTransport.NewServer(
			mockWS,
			logger,
			http.Dir("../../../static"),
			nil,
			mockQS,
			httpTransport.ServerConfig{
				Host:          "localhost",
				Port:          8080,
				Environment:   "test",
				DefaultLocale: loc,
				AuthDisabled:  true,
			},
		)

		req := httptest.NewRequest("GET", "/ws", nil)
		req.Header.Set("Connection", "upgrade")
		req.Header.Set("Upgrade", "websocket")
		req.Header.Set("Sec-WebSocket-Version", "13")
		req.Header.Set("Sec-WebSocket-Key", "test-key")

		w := httptest.NewRecorder()
		server.Server.Handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
