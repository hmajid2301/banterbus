package middleware

import (
	"net/http"
	"slices"
)

// Chain represents a chain of middleware functions
type Chain []func(http.Handler) http.Handler

// NewChain creates a new middleware chain
func NewChain(middlewares ...func(http.Handler) http.Handler) Chain {
	return Chain(middlewares)
}

// Then applies the middleware chain to an http.Handler
func (c Chain) Then(h http.Handler) http.Handler {
	for _, mw := range slices.Backward(c) {
		h = mw(h)
	}
	return h
}

// ThenFunc applies the middleware chain to an http.HandlerFunc
func (c Chain) ThenFunc(fn http.HandlerFunc) http.Handler {
	return c.Then(fn)
}

// Append adds middleware to the chain and returns a new chain
func (c Chain) Append(middlewares ...func(http.Handler) http.Handler) Chain {
	return append(c, middlewares...)
}

// Group represents a group of routes with shared middleware
type Group struct {
	chain  Chain
	router *Router
}

// NewGroup creates a new route group with the given middleware
func NewGroup(middlewares ...func(http.Handler) http.Handler) *Group {
	return &Group{
		chain: NewChain(middlewares...),
	}
}

// With adds middleware to the group and returns a new group
func (g *Group) With(middlewares ...func(http.Handler) http.Handler) *Group {
	return &Group{
		chain:  g.chain.Append(middlewares...),
		router: g.router,
	}
}

// Handler applies the group's middleware to a handler
func (g *Group) Handler(h http.Handler) http.Handler {
	return g.chain.Then(h)
}

// HandlerFunc applies the group's middleware to a handler function
func (g *Group) HandlerFunc(fn http.HandlerFunc) http.Handler {
	return g.chain.ThenFunc(fn)
}

// Handle registers a handler with the group's middleware applied
func (g *Group) Handle(pattern string, handler http.Handler) {
	g.router.mux.Handle(pattern, g.chain.Then(handler))
}

// HandleFunc registers a handler function with the group's middleware applied
func (g *Group) HandleFunc(pattern string, handler http.HandlerFunc) {
	g.router.mux.Handle(pattern, g.chain.ThenFunc(handler))
}

// Router provides a convenient way to organize routes with middleware
type Router struct {
	base Chain
	mux  *http.ServeMux
}

// NewRouter creates a new router with base middleware
func NewRouter(baseMiddleware ...func(http.Handler) http.Handler) *Router {
	return &Router{
		base: NewChain(baseMiddleware...),
		mux:  http.NewServeMux(),
	}
}

// Group creates a route group with additional middleware
func (router *Router) Group(name string, middlewares ...func(http.Handler) http.Handler) *Group {
	return &Group{
		chain:  router.base.Append(middlewares...),
		router: router,
	}
}

// Handle registers a handler with the router
func (router *Router) Handle(pattern string, handler http.Handler) {
	router.mux.Handle(pattern, router.base.Then(handler))
}

// HandleFunc registers a handler function with the router
func (router *Router) HandleFunc(pattern string, handler http.HandlerFunc) {
	router.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		router.base.ThenFunc(handler).ServeHTTP(w, r)
	})
}

// ServeHTTP implements http.Handler
func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.mux.ServeHTTP(w, r)
}
