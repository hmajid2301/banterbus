package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/hmajid2301/banterbus/internal/transport/http/middleware"
)

func TestMetricsMiddleware(t *testing.T) {
	t.Parallel()

	t.Run("Should record HTTP metrics", func(t *testing.T) {
		t.Parallel()

		// Create middleware instance
		m := middleware.Middleware{}

		// Create a test handler
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("test response"))
		})

		// Wrap with metrics middleware
		wrappedHandler := m.Metrics(testHandler)

		// Create test request
		req := httptest.NewRequest("GET", "/test", nil)
		recorder := httptest.NewRecorder()

		// Execute request
		wrappedHandler.ServeHTTP(recorder, req)

		// Verify response
		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "test response", recorder.Body.String())
	})

	t.Run("Should handle error responses", func(t *testing.T) {
		t.Parallel()

		// Create middleware instance
		m := middleware.Middleware{}

		// Create a test handler that returns an error
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("error"))
		})

		// Wrap with metrics middleware
		wrappedHandler := m.Metrics(testHandler)

		// Create test request
		req := httptest.NewRequest("POST", "/error", nil)
		recorder := httptest.NewRecorder()

		// Execute request
		wrappedHandler.ServeHTTP(recorder, req)

		// Verify response
		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		assert.Equal(t, "error", recorder.Body.String())
	})
}
