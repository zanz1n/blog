//go:build debug && !embed && !lambda
// +build debug,!embed,!lambda

package main

import (
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
)

func wireStaticRoutes(r chi.Router) error {
	subfs := os.DirFS("./web/dist")

	fileServer := http.FileServerFS(subfs)
	r.Mount("/static/", http.StripPrefix("/static", fileServer))

	return nil
}
