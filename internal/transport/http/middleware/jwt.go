package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func (m Middleware) ValidateJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.DisableAuth {
			next.ServeHTTP(w, r)
			return
		}

		bearerToken, err := getBearerToken(r.Header.Get("authorization"))
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(bearerToken, m.Keyfunc)
		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getBearerToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", fmt.Errorf("no authorization header")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", fmt.Errorf("no bearer prefix in authorization header")
	}

	bearerToken := strings.TrimPrefix(authHeader, "Bearer ")
	if bearerToken == "" {
		return "", fmt.Errorf("no jwt in authorization header")
	}
	return bearerToken, nil
}
