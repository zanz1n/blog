//go:build embed && !lambda
// +build embed,!lambda

package main

import (
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/zanz1n/blog/web"
)

func wireStaticRoutes(r chi.Router) error {
	subfs, err := fs.Sub(web.EmbedAssets, "dist")
	if err != nil {
		return err
	}

	fileServer := http.FileServerFS(subfs)
	r.Mount("/static/", http.StripPrefix("/static", fileServer))

	return nil
}
