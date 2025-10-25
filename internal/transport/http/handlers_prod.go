//go:build !dev

package http

import (
	"gitlab.com/hmajid2301/banterbus/internal/transport/http/middleware"
)

func (s *Server) registerDebugRoutes(router *middleware.Router) {
}
