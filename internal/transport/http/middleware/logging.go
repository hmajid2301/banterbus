package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

// responseWriter is a minimal wrapper for http.ResponseWriter that allows the
// written HTTP status code to be captured for logging.
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}

	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true
}

func (m Middleware) Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/ws" || strings.HasPrefix(path, "/static") || path == "/readiness" || path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				m.Logger.ErrorContext(r.Context(),
					"Request failed",
					slog.Any("trace", debug.Stack()),
				)
			}
		}()

		start := time.Now()
		wrapped := wrapResponseWriter(w)
		next.ServeHTTP(wrapped, r)
		m.Logger.InfoContext(
			r.Context(),
			"HTTP Request",
			slog.Int("status", wrapped.status),
			slog.String("method", r.Method),
			slog.String("path", r.URL.EscapedPath()),
			slog.Duration("duration", time.Since(start)),
		)
	})
}
