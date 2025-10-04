package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/hmajid2301/banterbus/internal/transport/http/middleware"
)

func TestValidateJWT(t *testing.T) {
	t.Parallel()

	t.Run("Should return unauthorized when no authorization header", func(t *testing.T) {
		t.Parallel()

		// Create middleware instance
		m := middleware.Middleware{}

		// Create a test handler
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		})

		// Wrap with JWT middleware
		wrappedHandler := m.ValidateJWT(testHandler)

		// Create test request without authorization header
		req := httptest.NewRequest("GET", "/test", nil)
		recorder := httptest.NewRecorder()

		// Execute request
		wrappedHandler.ServeHTTP(recorder, req)

		// Verify response
		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	})

	t.Run("Should return unauthorized with invalid bearer token format", func(t *testing.T) {
		t.Parallel()

		// Create middleware instance
		m := middleware.Middleware{}

		// Create a test handler
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		})

		// Wrap with JWT middleware
		wrappedHandler := m.ValidateJWT(testHandler)

		// Create test request with invalid authorization header
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "InvalidFormat")
		recorder := httptest.NewRecorder()

		// Execute request
		wrappedHandler.ServeHTTP(recorder, req)

		// Verify response
		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	})

	t.Run("Should return unauthorized with malformed JWT", func(t *testing.T) {
		t.Parallel()

		// Create middleware instance
		m := middleware.Middleware{}

		// Create a test handler
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		})

		// Wrap with JWT middleware
		wrappedHandler := m.ValidateJWT(testHandler)

		// Create test request with malformed JWT
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid.jwt.token")
		recorder := httptest.NewRecorder()

		// Execute request
		wrappedHandler.ServeHTTP(recorder, req)

		// Verify response
		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	})
}

func TestValidateAdminJWT(t *testing.T) {
	t.Parallel()

	t.Run("Should return unauthorized when no authorization header", func(t *testing.T) {
		t.Parallel()

		// Create middleware instance
		m := middleware.Middleware{}

		// Create a test handler
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		})

		// Wrap with admin JWT middleware
		wrappedHandler := m.ValidateAdminJWT(testHandler)

		// Create test request without authorization header
		req := httptest.NewRequest("GET", "/admin/test", nil)
		recorder := httptest.NewRecorder()

		// Execute request
		wrappedHandler.ServeHTTP(recorder, req)

		// Verify response
		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	})
}
