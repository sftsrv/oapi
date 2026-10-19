# oapi

A TUI for previewing OpenAPI Specs

![oapi in action](./images/oapi.gif)

## Installation

Install using `go get` as follows:

```sh
go get github.com/sftsrv/oapi@latest
```

## Usage

```sh
oapi [...flags] path_to_spec
```


Run the application by giving it the path to an [OpenAPI Spec](https://learn.openapis.org/) in JSON or YAML format

```sh
oapi openapi/example.json
```

Flags can also be provided to the initial state. Flags must always be passed before the path, for example:

```sh
oapi --json --body path/to/spec.json
```


Uses the following key bindings:

- `/` - search
- `↓↑/jk` - navigate
- `→/l` - expand
- `←/h` - collapse
- `]/[` - expand/collapse all
- `}/{` - resize
- `tab/shift+tab` - tabs
- `b/p/c/r/s/J/X/T/a` - toggle sections (body/parameters/codes/responses/schemas/JSON/XML/text/all)

