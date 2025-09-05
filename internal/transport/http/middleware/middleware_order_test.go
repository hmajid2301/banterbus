package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestWebSocketTracingWithMiddlewareChain(t *testing.T) {
	// Create a test handler that checks for baggage (simulates WebSocket handler)
	var receivedTestName string
	websocketHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		bag := baggage.FromContext(ctx)
		testNameMember := bag.Member("test_name")
		receivedTestName = testNameMember.Value()
		w.WriteHeader(http.StatusOK)
	})

	// Create middleware instance
	m := Middleware{}

	// Simulate the middleware chain order from handlers.go
	// BEFORE fix: routes = m.Logging(router); routes = m.Tracing(routes)
	// AFTER fix: routes = m.Tracing(router); routes = m.Logging(routes)

	// Apply middleware in the NEW order (tracing first, then logging)
	routes := m.Tracing(websocketHandler)
	routes = m.Logging(routes)

	// Create test request to /ws with X-Test-Name header
	req := httptest.NewRequest("GET", "/ws?test_name=TestWebSocketCorrelation", nil)
	req.Header.Set("X-Test-Name", "TestWebSocketCorrelation")

	// Create response recorder
	w := httptest.NewRecorder()

	// Set up minimal tracing context to avoid panics
	_, span := noop.NewTracerProvider().Tracer("test").Start(req.Context(), "test")
	ctx := trace.ContextWithSpan(req.Context(), span)
	req = req.WithContext(ctx)

	// Execute the handler chain
	routes.ServeHTTP(w, req)

	// Verify the test name was preserved in baggage through the middleware chain
	if receivedTestName != "TestWebSocketCorrelation" {
		t.Errorf("Expected test_name 'TestWebSocketCorrelation' in baggage, got '%s'", receivedTestName)
	}

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestOldMiddlewareOrder(t *testing.T) {
	// Test that the OLD middleware order still works but doesn't provide optimal logging
	var receivedTestName string
	websocketHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		bag := baggage.FromContext(ctx)
		testNameMember := bag.Member("test_name")
		receivedTestName = testNameMember.Value()
		w.WriteHeader(http.StatusOK)
	})

	// Create middleware instance
	m := Middleware{}

	// Apply middleware in the OLD order (logging first, then tracing)
	// This still works for /ws requests since logging middleware calls next handler
	routes := m.Logging(websocketHandler)
	routes = m.Tracing(routes)

	// Create test request to /ws with X-Test-Name header
	req := httptest.NewRequest("GET", "/ws?test_name=TestOldOrder", nil)
	req.Header.Set("X-Test-Name", "TestOldOrder")

	// Create response recorder
	w := httptest.NewRecorder()

	// Set up minimal tracing context to avoid panics
	_, span := noop.NewTracerProvider().Tracer("test").Start(req.Context(), "test")
	ctx := trace.ContextWithSpan(req.Context(), span)
	req = req.WithContext(ctx)

	// Execute the handler chain
	routes.ServeHTTP(w, req)

	// With the old order, the test name should still be present because
	// logging middleware skips processing but calls the next handler (tracing)
	if receivedTestName != "TestOldOrder" {
		t.Errorf("With old middleware order, expected test_name 'TestOldOrder' but got '%s'", receivedTestName)
	}

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
