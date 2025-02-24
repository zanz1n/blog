//go:build embed && !lambda
// +build embed,!lambda

package web

import "embed"

//go:embed dist/*
var EmbedAssets embed.FS
