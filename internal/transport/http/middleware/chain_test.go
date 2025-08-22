package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChain(t *testing.T) {
	// Create a simple handler
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Create middleware that sets a header
	middleware1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Middleware-1", "true")
			next.ServeHTTP(w, r)
		})
	}

	middleware2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Middleware-2", "true")
			next.ServeHTTP(w, r)
		})
	}

	// Create chain and apply middleware
	chain := NewChain(middleware1, middleware2)
	handler := chain.Then(finalHandler)

	// Test the handler
	req := httptest.NewRequest("GET", "/test", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	// Check response
	if recorder.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", recorder.Code)
	}

	if recorder.Header().Get("X-Middleware-1") != "true" {
		t.Error("middleware1 did not run")
	}

	if recorder.Header().Get("X-Middleware-2") != "true" {
		t.Error("middleware2 did not run")
	}

	if recorder.Body.String() != "success" {
		t.Errorf("expected 'success', got '%s'", recorder.Body.String())
	}
}

func TestGroup(t *testing.T) {
	// Create base middleware
	baseMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Base", "true")
			next.ServeHTTP(w, r)
		})
	}

	// Create group middleware
	groupMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Group", "true")
			next.ServeHTTP(w, r)
		})
	}

	// Create handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create group
	group := NewGroup(baseMiddleware).With(groupMiddleware)
	wrappedHandler := group.Handler(handler)

	// Test the handler
	req := httptest.NewRequest("GET", "/test", nil)
	recorder := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(recorder, req)

	// Check that both middleware ran
	if recorder.Header().Get("X-Base") != "true" {
		t.Error("base middleware did not run")
	}

	if recorder.Header().Get("X-Group") != "true" {
		t.Error("group middleware did not run")
	}
}

func TestRouter(t *testing.T) {
	// Create base middleware
	baseMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Base", "true")
			next.ServeHTTP(w, r)
		})
	}

	// Create router with base middleware
	router := NewRouter(baseMiddleware)

	// Add a simple handler
	router.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	})

	// Test the handler
	req := httptest.NewRequest("GET", "/test", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	// Check response
	if recorder.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", recorder.Code)
	}

	if recorder.Header().Get("X-Base") != "true" {
		t.Error("base middleware did not run")
	}

	if recorder.Body.String() != "test" {
		t.Errorf("expected 'test', got '%s'", recorder.Body.String())
	}
}
