package main

import (
	// "bufio"
	"flag"
	"fmt"
	"os"

	"github.com/sftsrv/oapi/internal/openapi"
	// "strings"
	// "github.com/sftsrv/oapi/tree"
	// "github.com/sftsrv/oapi/ui"
)

var usage = `oapi

A TUI for previewing OpenAPI Specs

Run the application by giving it the path to an [OpenAPI Spec](https://learn.openapis.org/) in JSON or YAML format

'''
oapi path/to/spec.json
'''

`

func main() {
	help := flag.Bool("help", false, "show help menu")
	flat := flag.Bool("flat", false, "flatten direct paths")

	flag.Parse()

	if *help || len(os.Args) < 2 {
		fmt.Println(usage, flat)
		flag.Usage()
		return
	}

	path := os.Args[1]

	result, err := openapi.FromFile(path)
	if err != nil {
		panic(err)
	}

	fmt.Println(result)

	// ui.Run(t, *preview, *pattern, *flat)
}
