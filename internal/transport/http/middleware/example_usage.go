package middleware

import (
	"log/slog"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

// This file demonstrates how to use the new HTTP middleware pattern
// It shows different ways to organize and compose HTTP middleware with routes

// ExampleHTTPSetup shows how to set up HTTP routes using the new pattern
func ExampleHTTPSetup(logger *slog.Logger, keyfunc jwt.Keyfunc) http.Handler {
	// Create middleware instance
	m := Middleware{
		DefaultLocale: "en",
		Logger:        logger,
		Keyfunc:       keyfunc,
		DisableAuth:   false,
		AdminGroup:    "admin",
	}

	// Create router with no base middleware
	router := NewRouter()

	// Public routes (no middleware)
	publicGroup := router.Group("public")
	publicGroup.HandleFunc("/health", healthHandler)
	publicGroup.HandleFunc("/metrics", metricsHandler)

	// Web routes (with locale middleware)
	webGroup := router.Group("web", m.Locale)
	webGroup.HandleFunc("/", indexHandler)
	webGroup.HandleFunc("/about", aboutHandler)
	webGroup.HandleFunc("/contact", contactHandler)

	// API routes (with locale + auth middleware)
	apiGroup := router.Group("api", m.Locale, m.ValidateJWT)
	apiGroup.HandleFunc("/profile", profileHandler)
	apiGroup.HandleFunc("/settings", settingsHandler)
	apiGroup.Handle("/data", methodRestrictedHandler("GET", dataHandler))

	// Admin routes (with locale + admin auth middleware)
	adminGroup := router.Group("admin", m.Locale, m.ValidateAdminJWT)
	adminGroup.HandleFunc("/dashboard", adminDashboardHandler)
	adminGroup.HandleFunc("/users", adminUsersHandler)
	adminGroup.Handle("/config", methodRestrictedHandler("PUT", adminConfigHandler))

	// Apply global logging middleware
	return m.Logging(router)
}

// ExampleAdvancedSetup shows more advanced middleware composition
func ExampleAdvancedSetup(logger *slog.Logger) http.Handler {
	// Create custom middleware
	rateLimitMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Rate limiting logic here
			next.ServeHTTP(w, r)
		})
	}

	corsMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			next.ServeHTTP(w, r)
		})
	}

	// Create base middleware chain
	baseChain := NewChain(corsMiddleware)

	// Create router with base middleware
	router := NewRouter(baseChain...)

	// API v1 routes (with rate limiting)
	v1Group := router.Group("v1", rateLimitMiddleware)
	v1Group.HandleFunc("/api/v1/users", v1UsersHandler)
	v1Group.HandleFunc("/api/v1/posts", v1PostsHandler)

	// API v2 routes (with different rate limiting)
	v2Group := router.Group("v2", rateLimitMiddleware) // Could use different rate limit
	v2Group.HandleFunc("/api/v2/users", v2UsersHandler)
	v2Group.HandleFunc("/api/v2/posts", v2PostsHandler)

	return router
}

// ExampleConditionalMiddleware shows how to apply middleware conditionally
func ExampleConditionalMiddleware(logger *slog.Logger, isDevelopment bool) http.Handler {
	m := Middleware{
		DefaultLocale: "en",
		Logger:        logger,
		DisableAuth:   isDevelopment, // Disable auth in development
	}

	router := NewRouter()

	// Debug routes (only in development)
	if isDevelopment {
		debugGroup := router.Group("debug")
		debugGroup.HandleFunc("/debug/vars", debugVarsHandler)
		debugGroup.HandleFunc("/debug/requests", debugRequestsHandler)
	}

	// API routes with conditional auth
	var apiMiddleware []func(http.Handler) http.Handler
	apiMiddleware = append(apiMiddleware, m.Locale)

	if !isDevelopment {
		apiMiddleware = append(apiMiddleware, m.ValidateJWT)
	}

	apiGroup := router.Group("api", apiMiddleware...)
	apiGroup.HandleFunc("/api/data", apiDataHandler)

	return m.Logging(router)
}

// Example handler functions

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("metrics"))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Welcome"))
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("About Us"))
}

func contactHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Contact"))
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Profile"))
}

func settingsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Settings"))
}

func dataHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Data"))
}

func adminDashboardHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Admin Dashboard"))
}

func adminUsersHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Admin Users"))
}

func adminConfigHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Admin Config"))
}

func v1UsersHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("V1 Users"))
}

func v1PostsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("V1 Posts"))
}

func v2UsersHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("V2 Users"))
}

func v2PostsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("V2 Posts"))
}

func debugVarsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Debug Vars"))
}

func debugRequestsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Debug Requests"))
}

func apiDataHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("API Data"))
}

// Helper function for method-restricted handlers
func methodRestrictedHandler(method string, handler http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == method {
			handler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
