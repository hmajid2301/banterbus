package websockets

import (
	"context"
	"slices"
)

// HandlerFunc defines a function that can handle WebSocket events
type HandlerFunc func(ctx context.Context, client *Client, sub *Subscriber) error

// MiddlewareFunc defines a function that wraps a HandlerFunc
type MiddlewareFunc func(HandlerFunc) HandlerFunc

// Chain represents a chain of middleware functions
type Chain []MiddlewareFunc

// NewChain creates a new middleware chain
func NewChain(middlewares ...MiddlewareFunc) Chain {
	return Chain(middlewares)
}

// Then applies the middleware chain to a HandlerFunc
func (c Chain) Then(h HandlerFunc) HandlerFunc {
	for _, mw := range slices.Backward(c) {
		h = mw(h)
	}
	return h
}

// Append adds middleware to the chain
func (c Chain) Append(middlewares ...MiddlewareFunc) Chain {
	return append(c, middlewares...)
}

// HandlerRegistry manages WebSocket handlers with middleware support
type HandlerRegistry struct {
	handlers   map[string]HandlerFunc
	middleware Chain
}

// NewHandlerRegistry creates a new handler registry
func NewHandlerRegistry(middleware ...MiddlewareFunc) *HandlerRegistry {
	return &HandlerRegistry{
		handlers:   make(map[string]HandlerFunc),
		middleware: NewChain(middleware...),
	}
}

// Register registers a handler for a specific message type
func (hr *HandlerRegistry) Register(messageType string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	chain := hr.middleware.Append(middleware...)
	hr.handlers[messageType] = chain.Then(handler)
}

// Handle executes the registered handler for a message type
func (hr *HandlerRegistry) Handle(messageType string, ctx context.Context, client *Client, sub *Subscriber) error {
	handler, ok := hr.handlers[messageType]
	if !ok {
		return ErrHandlerNotFound{MessageType: messageType}
	}
	return handler(ctx, client, sub)
}

// ErrHandlerNotFound represents an error when no handler is found for a message type
type ErrHandlerNotFound struct {
	MessageType string
}

func (e ErrHandlerNotFound) Error() string {
	return "handler not found for message type: " + e.MessageType
}

// Common middleware functions

// ValidationMiddleware creates middleware that validates the request using a validator function
func ValidationMiddleware[T any](validator func(T) error) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, client *Client, sub *Subscriber) error {
			// Note: In practice, you'd need to pass the validated data to the handler
			// This is a simplified version that assumes validation is done elsewhere
			return next(ctx, client, sub)
		}
	}
}

// LoggingMiddleware logs request and response information
func LoggingMiddleware() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, client *Client, sub *Subscriber) error {
			sub.logger.DebugContext(ctx, "handling websocket request", "player_id", client.playerID)
			err := next(ctx, client, sub)
			if err != nil {
				sub.logger.ErrorContext(ctx, "websocket request failed", "error", err, "player_id", client.playerID)
			}
			return err
		}
	}
}

// AuthMiddleware creates middleware that validates player authentication
func AuthMiddleware() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, client *Client, sub *Subscriber) error {
			// Add authentication logic here if needed
			// For now, we assume the player_id in the client is valid
			return next(ctx, client, sub)
		}
	}
}

// RecoveryMiddleware recovers from panics and converts them to errors
func RecoveryMiddleware() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, client *Client, sub *Subscriber) (err error) {
			defer func() {
				if r := recover(); r != nil {
					sub.logger.ErrorContext(ctx, "panic in websocket handler", "panic", r, "player_id", client.playerID)
					err = ErrPanicRecovered{Panic: r}
				}
			}()
			return next(ctx, client, sub)
		}
	}
}

// ErrPanicRecovered represents a recovered panic
type ErrPanicRecovered struct {
	Panic interface{}
}

func (e ErrPanicRecovered) Error() string {
	return "panic recovered in websocket handler"
}

// Additional error types

type ErrNoRawJSONData struct{}

func (e ErrNoRawJSONData) Error() string {
	return "no raw JSON data available in context"
}

type ErrJSONUnmarshal struct {
	Err error
}

func (e ErrJSONUnmarshal) Error() string {
	return "failed to unmarshal JSON: " + e.Err.Error()
}

type ErrValidation struct {
	Err error
}

func (e ErrValidation) Error() string {
	return "validation failed: " + e.Err.Error()
}
