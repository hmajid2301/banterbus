package middleware

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

type MyClaims struct {
	groups []string
	jwt.RegisteredClaims
}

func (m Middleware) ValidateAdminJWT(next http.Handler) http.Handler {
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

		token, err := jwt.ParseWithClaims(bearerToken, &MyClaims{}, m.Keyfunc)
		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		claims, ok := token.Claims.(*MyClaims)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		isAdmin := false
		for _, g := range claims.groups {
			if g == m.AdminGroup {
				isAdmin = true
			}
		}

		if !isAdmin {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
