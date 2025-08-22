package websockets

import (
	"context"
	"encoding/json"
	"log/slog"
)

// This file demonstrates how to use the new handler pattern
// It shows different ways to organize and compose handlers with middleware

// ExampleHandlerSetup shows how to set up handlers using the new pattern
func ExampleHandlerSetup(logger *slog.Logger) *HandlerRegistry {
	// Create base middleware that applies to all handlers
	baseMiddleware := NewChain(
		RecoveryMiddleware(),
		LoggingMiddleware(),
	)

	// Create registry with base middleware
	registry := NewHandlerRegistry(baseMiddleware...)

	// Register simple handlers
	registry.Register("ping", PingHandler())

	// Register handlers with additional middleware
	registry.Register("admin_action", AdminActionHandler(), RequireAdminMiddleware())

	// Register handlers with multiple middleware layers
	registry.Register("rate_limited_action",
		RateLimitedActionHandler(),
		RateLimitMiddleware(10), // 10 requests per minute
		RequireAuthMiddleware(),
	)

	return registry
}

// Example custom handlers

// PingHandler responds with a pong message
func PingHandler() HandlerFunc {
	return func(ctx context.Context, client *Client, sub *Subscriber) error {
		response := map[string]string{"type": "pong", "message": "Hello from server"}
		data, err := json.Marshal(response)
		if err != nil {
			return err
		}
		return sub.websocket.Publish(ctx, client.playerID, data)
	}
}

// AdminActionHandler performs admin-only actions
func AdminActionHandler() HandlerFunc {
	return func(ctx context.Context, client *Client, sub *Subscriber) error {
		// Admin-specific logic here
		sub.logger.InfoContext(ctx, "Admin action performed", "player_id", client.playerID)
		return nil
	}
}

// RateLimitedActionHandler demonstrates a handler that might need rate limiting
func RateLimitedActionHandler() HandlerFunc {
	return func(ctx context.Context, client *Client, sub *Subscriber) error {
		// Rate-limited action logic here
		sub.logger.InfoContext(ctx, "Rate limited action performed", "player_id", client.playerID)
		return nil
	}
}

// Example custom middleware

// RequireAdminMiddleware checks if the user has admin privileges
func RequireAdminMiddleware() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, client *Client, sub *Subscriber) error {
			// Check if player is admin (implementation would depend on your auth system)
			isAdmin, err := checkIfPlayerIsAdmin(ctx, client.playerID, sub)
			if err != nil {
				return err
			}

			if !isAdmin {
				return ErrUnauthorized{PlayerID: client.playerID}
			}

			return next(ctx, client, sub)
		}
	}
}

// RequireAuthMiddleware ensures the player is authenticated
func RequireAuthMiddleware() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, client *Client, sub *Subscriber) error {
			// Check authentication (implementation would depend on your auth system)
			isAuthenticated, err := checkPlayerAuthentication(ctx, client.playerID, sub)
			if err != nil {
				return err
			}

			if !isAuthenticated {
				return ErrUnauthenticated{PlayerID: client.playerID}
			}

			return next(ctx, client, sub)
		}
	}
}

// RateLimitMiddleware implements rate limiting
func RateLimitMiddleware(requestsPerMinute int) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, client *Client, sub *Subscriber) error {
			// Rate limiting logic here
			// This would typically use Redis or in-memory storage
			allowed, err := checkRateLimit(ctx, client.playerID, requestsPerMinute)
			if err != nil {
				return err
			}

			if !allowed {
				return ErrRateLimited{PlayerID: client.playerID}
			}

			return next(ctx, client, sub)
		}
	}
}

// Helper functions (these would need proper implementations)

func checkIfPlayerIsAdmin(ctx context.Context, playerID, sub interface{}) (bool, error) {
	// Implementation would check player's admin status
	return false, nil
}

func checkPlayerAuthentication(ctx context.Context, playerID, sub interface{}) (bool, error) {
	// Implementation would verify player authentication
	return true, nil
}

func checkRateLimit(ctx context.Context, playerID interface{}, requestsPerMinute int) (bool, error) {
	// Implementation would check rate limits
	return true, nil
}

// Custom error types

type ErrUnauthorized struct {
	PlayerID interface{}
}

func (e ErrUnauthorized) Error() string {
	return "unauthorized: insufficient privileges"
}

type ErrUnauthenticated struct {
	PlayerID interface{}
}

func (e ErrUnauthenticated) Error() string {
	return "unauthenticated: please login"
}

type ErrRateLimited struct {
	PlayerID interface{}
}

func (e ErrRateLimited) Error() string {
	return "rate limited: too many requests"
}

// JSONHandlerWrapper creates a generic handler that unmarshals JSON and validates
func JSONHandlerWrapper[T any](
	handler func(ctx context.Context, client *Client, sub *Subscriber, data T) error,
	validator func(T) error,
) HandlerFunc {
	return func(ctx context.Context, client *Client, sub *Subscriber) error {
		rawData, ok := ctx.Value("raw_json").([]byte)
		if !ok {
			return ErrNoRawJSONData{}
		}

		var data T
		if err := json.Unmarshal(rawData, &data); err != nil {
			return ErrJSONUnmarshal{Err: err}
		}

		if validator != nil {
			if err := validator(data); err != nil {
				return ErrValidation{Err: err}
			}
		}

		return handler(ctx, client, sub, data)
	}
}

// WSHandlerAdapter converts an existing WSHandler to HandlerFunc
func WSHandlerAdapter[T WSHandler](createHandler func() T) HandlerFunc {
	return func(ctx context.Context, client *Client, sub *Subscriber) error {
		rawData, ok := ctx.Value("raw_json").([]byte)
		if !ok {
			return ErrNoRawJSONData{}
		}

		handler := createHandler()

		if err := json.Unmarshal(rawData, handler); err != nil {
			return ErrJSONUnmarshal{Err: err}
		}

		if err := handler.Validate(); err != nil {
			return ErrValidation{Err: err}
		}

		return handler.Handle(ctx, client, sub)
	}
}

// Example usage of JSONHandlerWrapper:
// registry.Register("custom_action", JSONHandlerWrapper(
//     func(ctx context.Context, client *Client, sub *Subscriber, data CustomActionRequest) error {
//         // Handle the custom action
//         return nil
//     },
//     func(data CustomActionRequest) error {
//         // Validate the data
//         return data.Validate()
//     },
// ))
