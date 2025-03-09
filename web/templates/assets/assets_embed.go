//go:build !lambda
// +build !lambda

package assets

func init() {
	SetStaticCDN("/static")
}
