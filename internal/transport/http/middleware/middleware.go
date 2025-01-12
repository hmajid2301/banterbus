package middleware

import (
	"log/slog"

	"github.com/golang-jwt/jwt/v5"
)

type Middleware struct {
	DefaultLocale string
	Logger        *slog.Logger
	Keyfunc       jwt.Keyfunc
	DisableAuth   bool
	AdminGroup    string
}
