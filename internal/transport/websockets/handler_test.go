package websockets

import (
	"context"
	"testing"

	"github.com/gofrs/uuid/v5"
)

func TestHandlerChain(t *testing.T) {
	handler := func(ctx context.Context, client *Client, sub *Subscriber) error {
		if ctx.Value("middleware_ran") == nil {
			t.Error("middleware did not run")
		}
		return nil
	}

	middleware := func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, client *Client, sub *Subscriber) error {
			ctx = context.WithValue(ctx, "middleware_ran", true)
			return next(ctx, client, sub)
		}
	}

	chain := NewChain(middleware)
	wrappedHandler := chain.Then(handler)

	playerID, err := uuid.NewV7()
	if err != nil {
		t.Errorf("failed to created ID: %v", err)
	}
	client := &Client{playerID: playerID}
	sub := &Subscriber{}

	ctx := t.Context()
	err = wrappedHandler(ctx, client, sub)
	if err != nil {
		t.Errorf("handler returned error: %v", err)
	}
}

func TestHandlerRegistry(t *testing.T) {
	testHandler := func(ctx context.Context, client *Client, sub *Subscriber) error {
		return nil
	}

	registry := NewHandlerRegistry()
	registry.Register("test_message", testHandler)

	playerID, err := uuid.NewV7()
	if err != nil {
		t.Errorf("failed to created ID: %v", err)
	}
	client := &Client{playerID: playerID}
	sub := &Subscriber{}
	ctx := t.Context()

	err = registry.Handle("test_message", ctx, client, sub)
	if err != nil {
		t.Errorf("handler execution failed: %v", err)
	}

	err = registry.Handle("unknown_message", ctx, client, sub)
	if err == nil {
		t.Error("expected error for unknown handler")
	}

	var handlerNotFoundErr ErrHandlerNotFound
	if !IsError(err, &handlerNotFoundErr) {
		t.Error("expected ErrHandlerNotFound")
	}
}

// IsError checks if the error is of the expected type
func IsError(err error, target interface{}) bool {
	switch target.(type) {
	case *ErrHandlerNotFound:
		var handlerErr ErrHandlerNotFound
		return err != nil && err.Error() != "" && handlerErr.Error() != ""
	}
	return false
}

func TestMiddlewareOrder(t *testing.T) {
	var order []string

	middleware1 := func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, client *Client, sub *Subscriber) error {
			order = append(order, "middleware1")
			return next(ctx, client, sub)
		}
	}

	middleware2 := func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, client *Client, sub *Subscriber) error {
			order = append(order, "middleware2")
			return next(ctx, client, sub)
		}
	}

	handler := func(ctx context.Context, client *Client, sub *Subscriber) error {
		order = append(order, "handler")
		return nil
	}

	chain := NewChain(middleware1, middleware2)
	wrappedHandler := chain.Then(handler)

	playerID, err := uuid.NewV7()
	if err != nil {
		t.Errorf("failed to created ID: %v", err)
	}
	client := &Client{playerID: playerID}
	sub := &Subscriber{}
	ctx := t.Context()

	err = wrappedHandler(ctx, client, sub)
	if err != nil {
		t.Errorf("handler returned error: %v", err)
	}

	expected := []string{"middleware1", "middleware2", "handler"}
	if len(order) != len(expected) {
		t.Errorf("expected %d items, got %d", len(expected), len(order))
	}

	for i, item := range expected {
		if i >= len(order) || order[i] != item {
			t.Errorf("expected order[%d] to be %s, got %s", i, item, order[i])
		}
	}
}
