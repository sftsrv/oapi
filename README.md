# oapi

A TUI for previewing OpenAPI Specs

![oapi in action](./images/oapi.gif)

## Installation

Install using `go get` as follows:

```sh
go get github.com/sftsrv/oapi@latest
```

## Usage

Run the application by giving it the path to an [OpenAPI Spec](https://learn.openapis.org/) in JSON or YAML format

```sh
oapi openapi/example.json
```

Uses the following key bindings:

- `/` - search
- `↓↑/jk` - navigate
- `→/l` - expand
- `←/h` - collapse
- `]/[` - expand/collapse all
- `}/{` - resize
- `tab/shift+tab` - tabs
- `b/p/r/s/a/n` - toggle sections (body/params/response/schemas/all/none)

