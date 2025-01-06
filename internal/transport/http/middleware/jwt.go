package middleware

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

func (m Middleware) ValidateJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.DisableAuth {
			next.ServeHTTP(w, r)
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		bearerToken := authHeader[len("Bearer "):]
		token, err := jwt.Parse(bearerToken, m.Keyfunc)
		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
