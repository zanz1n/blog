package main

import (
	"fmt"

	"github.com/zanz1n/blog/config"
)

func main() {
	fmt.Println("Running", config.Name, config.Version)
}
