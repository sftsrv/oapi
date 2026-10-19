package main

import (
	"flag"
	"fmt"

	"github.com/sftsrv/oapi/internal/openapi"
	"github.com/sftsrv/oapi/internal/preview"
	"github.com/sftsrv/oapi/internal/ui"
)

var usage = `oapi

'''
oapi [...flags] path_to_spec
'''

A TUI for previewing OpenAPI Specs

Run the application by giving it the path to an [OpenAPI Spec](https://learn.openapis.org/) in JSON or YAML format

'''
oapi path/to/spec.json
'''

Flags can also be provided to the initial state. Flags must always be passed before the path, for example:

'''
oapi --json --body path/to/spec.json
'''

Uses the following key bindings:

- '/' - search
- '↓↑/jk' - navigate
- '→/l' - expand
- '←/h' - collapse
- ']/[' - expand/collapse all
- '}/{' - resize
- 'tab/shift+tab' - tabs
- 'b/p/r/s/J/X/T/a' - toggle sections (body/parameters/response/schemas/all)

`

var help = flag.Bool("help", false, "show help menu")
var flat = flag.Bool("flat", false, "flatten direct paths")

var body = flag.Bool("b", false, "initially enable body filter")
var parameters = flag.Bool("p", false, "initially enable parameters filter")
var codes = flag.Bool("c", false, "initially enable codes filter")
var responses = flag.Bool("r", false, "initially enable responses filter")
var schemas = flag.Bool("s", false, "initially enable schemas filter")

var json = flag.Bool("J", false, "initially enable json filter")
var xml = flag.Bool("X", false, "initially enable xml filter")
var text = flag.Bool("T", false, "initially enable text filter")

func main() {
	flag.Parse()

	args := flag.Args()

	if *help || len(args) < 1 {
		fmt.Println(usage)
		flag.Usage()
		return
	}

	path := args[0]

	spec, err := openapi.FromFile(path)
	if err != nil {
		panic(err)
	}

	ui.Run(spec, *flat, preview.Focus{
		Body:       *body,
		Parameters: *parameters,
		Codes:      *codes,
		Responses:  *responses,
		Schemas:    *schemas,
	}, preview.MimeTypes{
		Json: *json,
		Xml:  *xml,
		Text: *text,
	})
}
