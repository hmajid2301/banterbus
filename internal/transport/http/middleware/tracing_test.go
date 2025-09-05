package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestTracingMiddleware_TestNameCorrelation(t *testing.T) {
	// Create a test handler that checks for baggage
	var receivedTestName string
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		bag := baggage.FromContext(ctx)
		testNameMember := bag.Member("test_name")
		receivedTestName = testNameMember.Value()
		w.WriteHeader(http.StatusOK)
	})

	// Create middleware instance
	middleware := Middleware{}
	handler := middleware.Tracing(testHandler)

	// Create test request with X-Test-Name header
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Test-Name", "TestCorrelationWorking")

	// Create response recorder
	w := httptest.NewRecorder()

	// Set up minimal tracing context to avoid panics
	_, span := noop.NewTracerProvider().Tracer("test").Start(context.Background(), "test")
	ctx := trace.ContextWithSpan(req.Context(), span)
	req = req.WithContext(ctx)

	// Execute the handler
	handler.ServeHTTP(w, req)

	// Verify the test name was preserved in baggage
	if receivedTestName != "TestCorrelationWorking" {
		t.Errorf("Expected test_name 'TestCorrelationWorking' in baggage, got '%s'", receivedTestName)
	}

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestTracingMiddleware_QueryParameterCorrelation(t *testing.T) {
	// Create a test handler that checks for baggage
	var receivedTestName string
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		bag := baggage.FromContext(ctx)
		testNameMember := bag.Member("test_name")
		receivedTestName = testNameMember.Value()
		w.WriteHeader(http.StatusOK)
	})

	// Create middleware instance
	middleware := Middleware{}
	handler := middleware.Tracing(testHandler)

	// Create test request with test_name query parameter
	req := httptest.NewRequest("GET", "/ws?test_name=TestQueryCorrelation", nil)

	// Create response recorder
	w := httptest.NewRecorder()

	// Set up minimal tracing context to avoid panics
	_, span := noop.NewTracerProvider().Tracer("test").Start(context.Background(), "test")
	ctx := trace.ContextWithSpan(req.Context(), span)
	req = req.WithContext(ctx)

	// Execute the handler
	handler.ServeHTTP(w, req)

	// Verify the test name was preserved in baggage
	if receivedTestName != "TestQueryCorrelation" {
		t.Errorf("Expected test_name 'TestQueryCorrelation' in baggage, got '%s'", receivedTestName)
	}

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestTracingMiddleware_BaggageFirstPriority(t *testing.T) {
	// Test that baggage takes priority over headers and query params
	member, _ := baggage.NewMember("test_name", "BaggageTestName")
	bag, _ := baggage.New(member)

	// Create a test handler that checks for baggage
	var receivedTestName string
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		bag := baggage.FromContext(ctx)
		testNameMember := bag.Member("test_name")
		receivedTestName = testNameMember.Value()
		w.WriteHeader(http.StatusOK)
	})

	// Create test request with conflicting test names
	req := httptest.NewRequest("GET", "/ws?test_name=QueryTestName", nil)
	req.Header.Set("X-Test-Name", "HeaderTestName")

	// Add existing baggage to the request context
	ctx := baggage.ContextWithBaggage(req.Context(), bag)
	req = req.WithContext(ctx)

	// Create response recorder
	w := httptest.NewRecorder()

	// Set up minimal tracing context to avoid panics
	_, span := noop.NewTracerProvider().Tracer("test").Start(ctx, "test")
	ctx = trace.ContextWithSpan(ctx, span)
	req = req.WithContext(ctx)

	// Create middleware instance and execute
	middleware := Middleware{}
	handler := middleware.Tracing(testHandler)
	handler.ServeHTTP(w, req)

	// Should prioritize baggage over headers and query params
	if receivedTestName != "BaggageTestName" {
		t.Errorf("Expected baggage test_name 'BaggageTestName' to take priority, got '%s'", receivedTestName)
	}

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
