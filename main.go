package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sftsrv/oapi/internal/openapi"
	"github.com/sftsrv/oapi/internal/ui"
)

var usage = `oapi

A TUI for previewing OpenAPI Specs

Run the application by giving it the path to an [OpenAPI Spec](https://learn.openapis.org/) in JSON or YAML format

'''
oapi path/to/spec.json
'''

Uses the following key bindings:

- '/' - search
- '↓↑/jk' - navigate
- '→/l' - expand
- '←/h' - collapse
- ']/[' - expand/collapse all
- '}/{' - resize
- 'tab/shift+tab' - tabs
- 'b/p/r/s/a/n' - toggle sections (body/params/response/schemas/all/none)

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

	spec, err := openapi.FromFile(path)
	if err != nil {
		panic(err)
	}

	ui.Run(spec, *flat)
}
