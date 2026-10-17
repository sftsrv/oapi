package preview

import (
	"encoding/json"
	"os"
	"slices"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/glamour/v2"
	"charm.land/glamour/v2/styles"
	lg "charm.land/lipgloss/v2"
	"github.com/sftsrv/oapi/internal/openapi"
	"github.com/sftsrv/oapi/internal/picker"
	"github.com/sftsrv/oapi/internal/tree"
)

type Model struct {
	spec     *openapi.OpenAPISpec
	path     string
	ready    bool
	width    int
	height   int
	renderer *glamour.TermRenderer
	viewport viewport.Model
	tabIndex int

	focus *Focus
}

type Focus struct {
	body       bool
	parameters bool
	codes      bool
	responses  bool
	schemas    bool
}

func (f Focus) hasAll() bool {
	return f.body && f.parameters && f.responses && f.schemas && f.codes
}

func (f Focus) hasAny() bool {
	return f.body || f.parameters || f.responses || f.schemas || f.codes
}

func (f Focus) showAll() bool {
	if !f.hasAny() {
		return true
	}

	return f.hasAll()
}

func (m Model) EnableAll() (Model, tea.Cmd) {
	f := m.focus

	f.body = true
	f.parameters = true
	f.responses = true
	f.codes = true
	f.schemas = true

	return m, m.preview()
}

func (m Model) DisableAll() (Model, tea.Cmd) {
	f := m.focus

	f.body = false
	f.parameters = false
	f.responses = false
	f.codes = false
	f.schemas = false

	return m, m.preview()
}

func (m Model) ToggleBody() (Model, tea.Cmd) {
	if m.focus.hasAll() {
		m.DisableAll()
	}
	m.focus.body = !m.focus.body
	return m, m.preview()
}

func (m Model) ToggleParameters() (Model, tea.Cmd) {
	if m.focus.hasAll() {
		m.DisableAll()
	}
	m.focus.parameters = !m.focus.parameters
	return m, m.preview()
}

func (m Model) ToggleResponses() (Model, tea.Cmd) {
	if m.focus.hasAll() {
		m.DisableAll()
	}
	m.focus.responses = !m.focus.responses
	return m, m.preview()
}

func (m Model) ToggleCodes() (Model, tea.Cmd) {
	if m.focus.hasAll() {
		m.DisableAll()
	}
	m.focus.codes = !m.focus.codes
	return m, m.preview()
}

func (m Model) ToggleSchemas() (Model, tea.Cmd) {
	if m.focus.hasAll() {
		m.DisableAll()
	}
	m.focus.schemas = !m.focus.schemas
	return m, m.preview()
}

type PreviewResultMsg struct {
	path    string
	isError bool
	content string
}

type PreviewReadyMsg struct{}

type UpdateRendererMsg struct{ renderer *glamour.TermRenderer }

type ResizeMsg struct {
	Adjust int
}

var hasDarkBG = lg.HasDarkBackground(os.Stdin, os.Stdout)

func newRenderer(width int) *glamour.TermRenderer {

	theme := styles.LightStyle
	if hasDarkBG {
		theme = styles.DraculaStyle
	}

	renderer, _ := glamour.NewTermRenderer(glamour.WithWordWrap(width), glamour.WithStandardStyle(theme))

	return renderer
}

func New(spec *openapi.OpenAPISpec) Model {
	renderer := newRenderer(40)

	return Model{
		renderer: renderer,
		spec:     spec,
		focus:    &Focus{},
	}
}

func (m Model) SetPath(path string) (Model, tea.Cmd) {
	m.path = path
	m.tabIndex = 0
	return m, m.preview()
}

func (m Model) RendererUpdated(msg UpdateRendererMsg) (Model, tea.Cmd) {
	m.renderer = msg.renderer
	return m, m.preview()
}

func (m Model) SetContent(preview PreviewResultMsg) Model {
	if m.path == preview.path {
		m.viewport.SetContent(preview.content)
	}

	return m
}

func (m Model) ClearPath() Model {
	m.path = ""
	m.viewport.SetContent("")
	return m
}

func (m Model) Width(width int) Model {
	m.width = width
	return m
}

func (m Model) GetWidth() int {
	return m.width
}

func (m Model) Height(height int) Model {
	m.height = height
	return m
}

func (m Model) NextTab() (Model, tea.Cmd) {
	p := m.spec.Paths[m.path]
	m.tabIndex++

	if m.tabIndex >= len(p) {
		m.tabIndex = 0
	}
	m.tabIndex = picker.Clamp(m.tabIndex, 0, len(p)-1)

	return m, m.preview()
}

func (m Model) PreviousTab() (Model, tea.Cmd) {
	p := m.spec.Paths[m.path]

	m.tabIndex--

	if m.tabIndex < 0 {
		m.tabIndex = len(p) - 1
	}

	m.tabIndex = picker.Clamp(m.tabIndex, 0, len(p)-1)

	return m, m.preview()
}

func schemaBlock(schema openapi.Schema) string {
	json, _ := json.MarshalIndent(schema, "", "  ")
	return "```json\n\n" + string(json) + "\n\n```"
}

func printContent(titlePrefix string, contents map[string]openapi.Content) string {
	md := ""
	keys := tree.SortedKeys(contents)

	for _, key := range keys {
		md += titlePrefix + " " + key
		md += "\n\n"

		md += schemaBlock(contents[key].Schema)
		md += "\n\n"
	}

	return md

}

func (s Model) preview() tea.Cmd {
	return func() tea.Msg {
		focus := s.focus
		md := ""

		p, isPath := s.spec.Paths[s.path]
		schema, isSchema := s.spec.Components.Schemas[getSchemaLocalNameFromRef(s.path)]

		if !isPath && !isSchema {
			if s.path != "" {
				md += "No endpoint at path `" + s.path + "`"
				md += "\n\n"
			}

			md += " " + s.spec.Info.Title
			md += "\n\n"
			md += "> version: " + s.spec.Info.Version

		} else if isPath {
			if focus.parameters {
				md += "`PARAMS`"
			} else {
				md += "PARAMS"
			}
			md += " "

			if focus.body {
				md += "`BODY`"
			} else {
				md += "BODY"
			}
			md += " "

			if focus.responses {
				md += "`RESPONSES`"
			} else {
				md += "RESPONSES"
			}
			md += " "

			if focus.codes {
				md += "`CODES`"
			} else {
				md += "CODES"
			}
			md += " "

			if focus.schemas {
				md += "`SCHEMAS`"
			} else {
				md += "SCHEMAS"
			}
			md += " "

			md += "\n\n"
			md += "-------"
			md += "\n\n"

			md += "\n\n"

			methods := tree.SortedKeys(p)
			methodIndex := picker.Clamp(s.tabIndex, 0, len(methods)-1)
			m := methods[methodIndex]
			method := p[m]

			for i, m := range methods {
				if i == methodIndex {
					md += "`" + strings.ToUpper(m) + "`"
				} else {
					md += strings.ToUpper(m)

				}
				md += "  "
			}

			md += "\n\n"
			md += "# " + method.Summary
			md += "\n\n"
			md += "> " + method.Description
			md += "\n\n"

			if (focus.showAll() || focus.parameters) && len(method.Parameters) > 0 {
				md += "## Parameters"
				md += "\n\n"

				slices.SortFunc(method.Parameters, func(a, b openapi.Parameter) int {
					return strings.Compare(a.Name, b.Name)
				})

				for _, p := range method.Parameters {
					md += "- " + p.Name

					t := p.Schema.Type

					if p.Schema.Ref != "" {
						t += "(" + p.Schema.Ref + ")"
					}

					if p.Required {
						t += "(required)"
					}
					t = "`" + t + "`"
					md += " " + t

					if p.Description != "" {
						md += " - " + p.Description
					}
					md += "\n\n"

					if len(p.Schema.Required) > 0 {
						md += "- required: "
						md += "\n"

						for _, r := range p.Schema.Required {
							md += "  - " + r
							md += "\n"
						}
					}
				}
			}

			if (focus.showAll() || focus.body) && len(method.RequestBody.Content) > 0 {
				md += "## Body"
				md += "\n\n"

				content := method.RequestBody.Content

				md += printContent("###", content)
			}

			if (focus.showAll() || focus.codes) && len(method.Responses) > 0 {
				md += "## Codes"
				md += "\n\n"

				for _, code := range tree.SortedKeys(method.Responses) {
					res := method.Responses[code]
					md += "### " + code
					md += "\n\n"

					md += "> " + res.Description
					md += "\n\n"
				}
			}

			if (focus.showAll() || focus.responses) && len(method.Responses) > 0 {
				md += "## Responses"
				md += "\n\n"

				for _, code := range tree.SortedKeys(method.Responses) {
					res := method.Responses[code]
					md += "### " + code
					md += "\n\n"

					md += "> " + res.Description
					md += "\n\n"

					md += printContent("####", res.Content)
					md += "\n\n"
				}
			}

			relatedSchemas := getRelatedSchemas(s.spec, method)

			if (focus.showAll() || focus.schemas) && len(relatedSchemas) > 0 {
				md += "## Related Schemas"
				md += "\n\n"

				relatedSchemaKeys := tree.SortedKeys(relatedSchemas)
				for _, s := range relatedSchemaKeys {
					md += "## " + s
					md += "\n\n"
					md += schemaBlock(relatedSchemas[s])
					md += "\n\n"
				}
			}

		} else if isSchema {
			md += "# " + getSchemaLocalNameFromRef(s.path)
			md += "\n\n"
			md += schemaBlock(schema)
			md += "\n\n"
		}

		rendered, err := s.renderer.Render(md)
		if err != nil {
			return PreviewResultMsg{s.path, true, err.Error()}
		}

		return PreviewResultMsg{s.path, false, rendered}
	}
}

func getSchemaLocalNameFromRef(ref string) string {
	return strings.TrimPrefix(ref, openapi.SCHEMA_PREFIX)
}

func getSchemaLocalName(schema openapi.Schema) string {
	return getSchemaLocalNameFromRef(schema.Ref)
}

func getSchemaNames(schema openapi.Schema) []string {
	schemas := []string{}

	if schema.Ref != "" {
		schemas = append(schemas, getSchemaLocalName(schema))
	}

	if schema.Items != nil && schema.Items.Ref != "" {
		schemas = append(schemas, getSchemaLocalName(*schema.Items))
	}

	childKeys := tree.SortedKeys(schema.Properties)
	for _, key := range childKeys {
		child := schema.Properties[key]
		schemas = append(schemas, getSchemaNames(child)...)
	}

	return schemas
}

func getRelatedSchemas(openAPISpec *openapi.OpenAPISpec, method openapi.Path) map[string]openapi.Schema {
	keys := []string{}

	responses := tree.SortedKeys(method.Responses)
	for _, m := range responses {
		content := method.Responses[m].Content
		contents := tree.SortedKeys(content)
		for _, c := range contents {
			keys = append(keys, getSchemaNames(content[c].Schema)...)
		}
	}

	requests := tree.SortedKeys(method.RequestBody.Content)
	for _, m := range requests {
		content := method.Responses[m].Content
		contents := tree.SortedKeys(content)
		for _, c := range contents {
			keys = append(keys, getSchemaNames(content[c].Schema)...)
		}
	}

	schemas := map[string]openapi.Schema{}

	slices.Sort(keys)
	for _, key := range keys {
		schema, ok := openAPISpec.Components.Schemas[key]
		if ok {
			schemas[key] = schema
		}
	}

	return schemas
}

func (m Model) View() string {
	return lg.JoinVertical(
		lg.Center,
		lg.NewStyle().
			Width(m.width).
			PaddingLeft(1).
			PaddingRight(1).
			Bold(true).
			Render(m.path),
		m.viewport.View(),
	)
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case ResizeMsg:
		m.width += msg.Adjust
		m.viewport.SetWidth(m.width)
		cmds = append(cmds, func() tea.Msg {
			return UpdateRendererMsg{
				renderer: newRenderer(m.width),
			}
		})

	case tea.WindowSizeMsg:
		if !m.ready {
			// Since this program is using the full size of the viewport we
			// need to wait until we've received the window dimensions before
			// we can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.
			m.viewport = viewport.New(viewport.WithHeight(m.height-1), viewport.WithWidth(m.width))
			m.viewport.MouseWheelEnabled = true

			m.viewport.YPosition = 0
			m.viewport.SetContent("")
			m.ready = true
			readyCmd := func() tea.Msg {
				return PreviewReadyMsg{}
			}

			cmds = append(cmds, readyCmd)
			cmds = append(cmds, func() tea.Msg {
				return UpdateRendererMsg{
					renderer: newRenderer(m.width),
				}
			})

		} else {
			m.viewport.SetWidth(m.width)
			m.viewport.SetHeight(m.height)
			cmds = append(cmds, func() tea.Msg {
				return UpdateRendererMsg{
					renderer: newRenderer(m.width),
				}
			})
		}
	}

	// Handle keyboard and mouse events in the viewport
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}
