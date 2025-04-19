package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

var _ chi.Router = &RoutesMockup{}
var _ fmt.Stringer = &Route{}

type Route struct {
	Method  string
	Pattern string
}

// String implements fmt.Stringer.
func (r *Route) String() string {
	return fmt.Sprintf("%s %s", r.Method, r.Pattern)
}

type RoutesMockup struct {
	Inner []Route
}

func (r *RoutesMockup) add(method, pattern string) {
	r.Inner = append(r.Inner, Route{Method: method, Pattern: pattern})
}

// Delete implements chi.Router.
func (r *RoutesMockup) Delete(pattern string, h http.HandlerFunc) {
	r.add(http.MethodDelete, pattern)
}

// Get implements chi.Router.
func (r *RoutesMockup) Get(pattern string, h http.HandlerFunc) {
	r.add(http.MethodGet, pattern)
}

// Head implements chi.Router.
func (r *RoutesMockup) Head(pattern string, h http.HandlerFunc) {
	r.add(http.MethodHead, pattern)
}

// Options implements chi.Router.
func (r *RoutesMockup) Options(pattern string, h http.HandlerFunc) {
	r.add(http.MethodOptions, pattern)
}

// Patch implements chi.Router.
func (r *RoutesMockup) Patch(pattern string, h http.HandlerFunc) {
	r.add(http.MethodPatch, pattern)
}

// Post implements chi.Router.
func (r *RoutesMockup) Post(pattern string, h http.HandlerFunc) {
	r.add(http.MethodPost, pattern)
}

// Put implements chi.Router.
func (r *RoutesMockup) Put(pattern string, h http.HandlerFunc) {
	r.add(http.MethodPut, pattern)
}

// Method implements chi.Router.
func (r *RoutesMockup) Method(method string, pattern string, h http.Handler) {
	r.add(method, pattern)
}

// MethodFunc implements chi.Router.
func (r *RoutesMockup) MethodFunc(method string, pattern string, h http.HandlerFunc) {
	r.add(method, pattern)
}

func (r *RoutesMockup) Connect(pattern string, h http.HandlerFunc)                     {}
func (r *RoutesMockup) Find(rctx *chi.Context, method string, path string) string      { return "" }
func (r *RoutesMockup) Group(fn func(r chi.Router)) chi.Router                         { return r }
func (r *RoutesMockup) Handle(pattern string, h http.Handler)                          {}
func (r *RoutesMockup) HandleFunc(pattern string, h http.HandlerFunc)                  {}
func (r *RoutesMockup) Match(rctx *chi.Context, method string, path string) bool       { return false }
func (r *RoutesMockup) MethodNotAllowed(h http.HandlerFunc)                            {}
func (r *RoutesMockup) Middlewares() chi.Middlewares                                   { return nil }
func (r *RoutesMockup) Mount(pattern string, h http.Handler)                           {}
func (r *RoutesMockup) NotFound(h http.HandlerFunc)                                    {}
func (r *RoutesMockup) Route(pattern string, fn func(r chi.Router)) chi.Router         { return r }
func (r *RoutesMockup) Routes() []chi.Route                                            { return nil }
func (r *RoutesMockup) ServeHTTP(http.ResponseWriter, *http.Request)                   {}
func (r *RoutesMockup) Trace(pattern string, h http.HandlerFunc)                       {}
func (r *RoutesMockup) Use(middlewares ...func(http.Handler) http.Handler)             {}
func (r *RoutesMockup) With(middlewares ...func(http.Handler) http.Handler) chi.Router { return r }
