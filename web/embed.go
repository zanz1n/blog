//go:build !debug && !lambda
// +build !debug,!lambda

package web

import "embed"

//go:embed dist/*
var EmbedAssets embed.FS
