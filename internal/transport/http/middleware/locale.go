package middleware

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/invopop/ctxi18n"
)

func (m Middleware) Locale(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		locale := r.Header.Get("Accept-Language")
		pathSegments := strings.Split(r.URL.Path, "/")
		if len(pathSegments) > 1 {
			locale = pathSegments[1]
		}

		ctx, err := ctxi18n.WithLocale(r.Context(), locale)
		if err != nil {
			locale = m.DefaultLocale
			ctx, err = ctxi18n.WithLocale(r.Context(), locale)
			if err != nil {
				m.Logger.ErrorContext(
					ctx,
					"error setting locale",
					slog.Any("error", err),
					slog.String("locale", locale),
				)
				http.Error(w, "error setting locale", http.StatusBadRequest)
				return
			}
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "locale",
			Value:    locale,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			Expires:  time.Now().Add(time.Hour),
		})

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
