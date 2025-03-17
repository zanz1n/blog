//go:build lambda
// +build lambda

package main

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"os"

	"github.com/akrylysov/algnhsa"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/go-chi/chi/v5"
	"github.com/zanz1n/blog/web/templates/assets"
)

func init() {
	staicAssets := os.Getenv("STATIC_ASSETS")
	if staicAssets == "" {
		fatal(errors.New("environment variable `STATIC_ASSETS` not provided"))
	}

	_, err := url.Parse(staicAssets)
	if err != nil {
		fatal(err)
	}

	assets.SetStaticCDN(staicAssets)
}

func wireStaticRoutes(chi.Router) error {
	return nil
}

func listen(ctx context.Context, h http.Handler) error {
	lambdah := algnhsa.New(h, nil)
	lambda.StartWithOptions(lambdah, lambda.WithContext(ctx))
	return nil
}
