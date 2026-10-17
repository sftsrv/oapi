package openapi

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
)

const SCHEMA_PREFIX = "#/components/schemas/"

type OpenAPISpec struct {
	Info       Info             `json:"info"`
	Paths      map[string]Paths `json:"paths,omitempty"`
	Components Components       `json:"components"`
}

type Info struct {
	Version string `json:"version"`
	Title   string `json:"title"`
}

type Components struct {
	Schemas Schemas `json:"schemas"`
}

type Schemas = map[string]Schema

type Paths = map[string]Path

type Path struct {
	Summary     string              `json:"summary,omitempty"`
	Description string              `json:"description,omitempty"`
	OperationId string              `json:"operationId,omitempty"`
	Tags        []string            `json:"tags,omitempty"`
	Parameters  []Parameter         `json:"parameters,omitempty"`
	RequestBody RequestBody         `json:"requestBody"`
	Responses   map[string]Response `json:"responses,omitempty"`
}

type RequestBody struct {
	Description string             `json:"description,omitempty"`
	Content     map[string]Content `json:"content,omitempty"`
}

type Parameter struct {
	Name        string `json:"name,omitempty"`
	Required    bool   `json:"required,omitempty"`
	Description string `json:"description,omitempty"`
	Schema      Schema `json:"schema"`
}

type Response struct {
	Description string             `json:"description,omitempty"`
	Content     map[string]Content `json:"content,omitempty"`
}

type Content struct {
	Schema Schema `json:"schema"`
}

type Schema struct {
	Type       string   `json:"type,omitempty"`
	Format     string   `json:"format,omitempty"`
	Ref        string   `json:"$ref,omitempty"`
	Required   []string `json:"required,omitempty"`
	MaxItems   int      `json:"maxItems,omitempty"`
	Properties Schemas  `json:"properties,omitempty"`
	Items      *Schema  `json:"items,omitempty"`
}

func FromFile(path string) (*OpenAPISpec, error) {
	var result OpenAPISpec

	bytes, err := os.ReadFile(path)
	if err != nil {
		return &result, err
	}

	if strings.HasSuffix(path, ".json") {
		err = json.Unmarshal(bytes, &result)
		return &result, err
	}

	err = yaml.Unmarshal(bytes, &result)
	return &result, err
}
