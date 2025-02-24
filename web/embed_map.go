//go:build !debug
// +build !debug

package web

import _ "embed"

//go:embed .source-map.json
var EmbedSourceMap []byte
