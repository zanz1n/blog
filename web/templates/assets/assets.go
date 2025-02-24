package assets

import "strings"

var staticCdn string

func SetStaticCDN(url string) {
	staticCdn = strings.TrimSuffix(url, "/")
}
