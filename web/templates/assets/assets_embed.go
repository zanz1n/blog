//go:build embed || debug
// +build embed debug

package assets

func init() {
	SetStaticCDN("/static")
}
