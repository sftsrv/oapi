package preview

import (
	"os"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/glamour/v2"
	lg "charm.land/lipgloss/v2"
	"github.com/sftsrv/oapi/internal/openapi"
	"github.com/sftsrv/oapi/internal/picker"
	"github.com/sftsrv/oapi/internal/theme"
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

	theme := "light"
	if hasDarkBG {
		theme = "dark"
	}

	renderer, _ := glamour.NewTermRenderer(glamour.WithWordWrap(width), glamour.WithStandardStyle(theme))

	return renderer
}

func New(spec *openapi.OpenAPISpec) Model {
	renderer := newRenderer(40)

	return Model{
		renderer: renderer,
		spec:     spec,
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

func (s Model) preview() tea.Cmd {
	return func() tea.Msg {
		md := ""

		if s.path == "" {
			md += "# " + s.spec.Info.Title
			md += "\n\n"
			md += "> version: " + s.spec.Info.Version
		} else {
			p, ok := s.spec.Paths[s.path]

			if !ok {
				return nil
			}

			md += "# " + s.path
			md += "\n\n"

			methods := tree.SortedKeys(p)
			methodIndex := picker.Clamp(s.tabIndex, 0, len(methods)-1)
			m := methods[methodIndex]
			method := p[m]

			for i, m := range methods {
				if i == methodIndex {
					md += "`" + strings.ToUpper(m) + "`"
				} else {
					md += "&nbsp;" + strings.ToUpper(m) + "&nbsp;"

				}
				md += "  "
			}

			md += "\n\n"
			md += "> " + method.Summary
			md += "\n\n"

			if len(method.Parameters) == 0 {
				md += "### No Parameters"
				md += "\n\n"
			} else {
				md += "### Parameters"
				md += "\n\n"
				for _, p := range method.Parameters {
					md += "#### " + p.Name

					t := p.Schema.Type

					if p.Schema.Ref != "" {
						t += "(" + p.Schema.Ref + ")"
					}

					md += " - " + t
					if p.Required {
						md += "(required)"
					}

					md += "\n\n"
					md += p.Description
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

		}

		rendered, err := s.renderer.Render(md)
		if err != nil {
			return PreviewResultMsg{s.path, true, err.Error()}
		}

		return PreviewResultMsg{s.path, false, rendered}
	}
}

func (m Model) View() string {
	return lg.JoinVertical(
		lg.Center,
		lg.NewStyle().
			Width(m.width).
			PaddingLeft(1).
			PaddingRight(1).
			Background(theme.ColorPrimary).
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
