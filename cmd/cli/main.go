package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/zanz1n/blog/internal/server"
)

func init() {
	flag.Parse()
}

func main() {
	arg := flag.Arg(0)

	if arg == "export-routes" {
		exportRoutes()
		return
	} else {
		invalidArg(arg)
		return
	}
}

func invalidArg(arg string) {
	if arg == "" {
		fmt.Printf("An argument must be provided\n")
	} else {
		fmt.Printf("Invalid arg: %s\n", arg)
	}
	os.Exit(1)
}

func exportRoutes() {
	router := &RoutesMockup{}
	server.New(nil, nil, nil).Wire(router)

	arr := make([]string, len(router.Inner))

	for i, r := range router.Inner {
		arr[i] = r.String()
	}

	datab, err := json.Marshal(arr)
	if err != nil {
		panic(err)
	}
	data := string(datab)

	output := struct {
		Data string `json:"data"`
	}{data}

	jb, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		panic(err)
	}

	os.Stdout.Write(jb)
}
