package main

import (
	"fmt"

	"github.com/zanz1n/blog/config"
)

func main() {
	fmt.Println("Running CLI", config.Name, config.Version)
}
