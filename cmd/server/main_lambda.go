//go:build lambda
// +build lambda

package main

import (
	"flag"
	"fmt"

	"github.com/zanz1n/blog/config"
)

func init() {
	flag.Parse()
}

func main() {
	fmt.Println("Running", config.Name, config.Version)
}
