//go:build dev

package http

import (
	"net/http/pprof"

	"gitlab.com/hmajid2301/banterbus/internal/transport/http/middleware"
)

func (s *Server) registerDebugRoutes(router *middleware.Router) {
	debugGroup := router.Group("debug")
	debugGroup.HandleFunc("/debug/pprof/", pprof.Index)
	debugGroup.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	debugGroup.HandleFunc("/debug/pprof/profile", pprof.Profile)
	debugGroup.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	debugGroup.HandleFunc("/debug/pprof/trace", pprof.Trace)
}
