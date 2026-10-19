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
	mimes *MimeTypes
}

type Focus struct {
	Body       bool
	Parameters bool
	Codes      bool
	Responses  bool
	Schemas    bool
}

type MimeTypes struct {
	Json bool
	Xml  bool
	Text bool
}

func (f Focus) hasAll() bool {
	return f.Body && f.Parameters && f.Responses && f.Schemas && f.Codes
}

func (f Focus) hasAny() bool {
	return f.Body || f.Parameters || f.Responses || f.Schemas || f.Codes
}

func (f Focus) showAll() bool {
	if !f.hasAny() {
		return true
	}

	return f.hasAll()
}

func (f MimeTypes) hasAll() bool {
	return f.Json && f.Xml && f.Text
}

func (f MimeTypes) hasAny() bool {
	return f.Json || f.Xml || f.Text
}

func (f MimeTypes) showAll() bool {
	if !f.hasAny() {
		return true
	}

	return f.hasAll()
}

func (m Model) DisableAll() (Model, tea.Cmd) {

	m.focus.Body = false
	m.focus.Parameters = false
	m.focus.Responses = false
	m.focus.Codes = false
	m.focus.Schemas = false

	m.mimes.Json = false
	m.mimes.Xml = false
	m.mimes.Text = false

	return m, m.preview()
}

func (m Model) ToggleBody() (Model, tea.Cmd) {
	m.focus.Body = !m.focus.Body
	return m, m.preview()
}

func (m Model) ToggleParameters() (Model, tea.Cmd) {
	m.focus.Parameters = !m.focus.Parameters
	return m, m.preview()
}

func (m Model) ToggleResponses() (Model, tea.Cmd) {
	m.focus.Responses = !m.focus.Responses
	return m, m.preview()
}

func (m Model) ToggleCodes() (Model, tea.Cmd) {
	m.focus.Codes = !m.focus.Codes
	return m, m.preview()
}

func (m Model) ToggleSchemas() (Model, tea.Cmd) {
	m.focus.Schemas = !m.focus.Schemas
	return m, m.preview()
}

func (m Model) ToggleJson() (Model, tea.Cmd) {
	m.mimes.Json = !m.mimes.Json
	return m, m.preview()
}

func (m Model) ToggleXml() (Model, tea.Cmd) {
	m.mimes.Xml = !m.mimes.Xml
	return m, m.preview()
}

func (m Model) ToggleText() (Model, tea.Cmd) {
	m.mimes.Text = !m.mimes.Text
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

func tab(str string, selected bool) string {
	str = strings.ToUpper(str)
	// Light automatically adds spacing around code blocks when selected
	light := !hasDarkBG
	if light && !selected {
		return "&nbsp;" + str + "&nbsp;"
	}

	if selected {
		return "`" + str + "`"

	}

	return str
}

func newRenderer(width int) *glamour.TermRenderer {
	theme := styles.LightStyle
	if hasDarkBG {
		theme = styles.TokyoNightStyle
	}

	renderer, _ := glamour.NewTermRenderer(glamour.WithWordWrap(width), glamour.WithStandardStyle(theme))

	return renderer
}

func New(spec *openapi.OpenAPISpec, focus *Focus, mime *MimeTypes) Model {
	renderer := newRenderer(40)

	return Model{
		renderer: renderer,
		spec:     spec,
		focus:    focus,
		mimes:    mime,
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

func (c MimeTypes) printContent(titlePrefix string, contents map[string]openapi.Content) string {
	md := ""

	if c.showAll() {
		keys := tree.SortedKeys(contents)

		for _, key := range keys {
			md += titlePrefix + " " + key
			md += "\n\n"

			md += schemaBlock(contents[key].Schema)
			md += "\n\n"
		}

		return md
	}

	printMime := func(t string) {
		data, hasT := contents[t]
		if !hasT {
			return
		}

		md += titlePrefix + " " + t
		md += "\n\n"

		md += schemaBlock(data.Schema)
		md += "\n\n"

	}

	if c.Json {
		printMime("application/json")
	}
	if c.Xml {
		printMime("application/xml")
	}
	if c.Text {
		printMime("application/text")
	}

	return md
}

func (s Model) preview() tea.Cmd {
	return func() tea.Msg {
		focus := s.focus
		mimes := s.mimes
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
			md += tab("params", focus.Parameters)
			md += " "

			md += tab("body", focus.Body)
			md += " "

			md += tab("responses", focus.Responses)
			md += " "

			md += tab("codes", focus.Codes)
			md += " "

			md += tab("schemas", focus.Schemas)
			md += " "

			md += tab("json", mimes.Json)
			md += " "

			md += tab("xml", mimes.Xml)
			md += " "

			md += tab("text", mimes.Text)
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
				md += tab(m, i == methodIndex)
				md += "  "
			}
			md += "\n\n"

			if len(method.Tags) > 0 {
				md += "**Tags: #" + strings.Join(method.Tags, " #") + "**"
				md += "\n\n"
			}

			md += "# " + method.Summary
			md += "\n\n"
			md += "> " + method.Description
			md += "\n\n"

			if (focus.showAll() || focus.Parameters) && len(method.Parameters) > 0 {
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

			if (focus.showAll() || focus.Body) && len(method.RequestBody.Content) > 0 {
				md += "## Body"
				md += "\n\n"

				content := method.RequestBody.Content

				md += mimes.printContent("###", content)
			}

			if (focus.showAll() || focus.Codes) && len(method.Responses) > 0 {
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

			if (focus.showAll() || focus.Responses) && len(method.Responses) > 0 {
				md += "## Responses"
				md += "\n\n"

				for _, code := range tree.SortedKeys(method.Responses) {
					res := method.Responses[code]
					md += "### " + code
					md += "\n\n"

					md += "> " + res.Description
					md += "\n\n"

					md += mimes.printContent("####", res.Content)
					md += "\n\n"
				}
			}

			relatedSchemas := getRelatedSchemas(s.spec, method)

			if (focus.showAll() || focus.Schemas) && len(relatedSchemas) > 0 {
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
