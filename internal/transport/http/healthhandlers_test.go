package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServer_HealthHandlers(t *testing.T) {
	t.Parallel()

	server := &Server{}

	t.Run("Should return OK for health check", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		server.healthHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Body.String())
	})

	t.Run("Should return OK for readiness check", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/ready", nil)
		w := httptest.NewRecorder()

		server.readinessHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Body.String())
	})

	t.Run("Should handle context cancellation gracefully", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(t.Context())
		req := httptest.NewRequest(http.MethodGet, "/health", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		cancel()

		server.healthHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Should handle POST method on health endpoint", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/health", nil)
		w := httptest.NewRecorder()

		server.healthHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Should handle HEAD method on ready endpoint", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodHead, "/ready", nil)
		w := httptest.NewRecorder()

		server.readinessHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Body.String())
	})
}
