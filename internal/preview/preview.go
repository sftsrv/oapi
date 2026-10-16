package preview

import (
	"os"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/glamour/v2"
	lg "charm.land/lipgloss/v2"
	"github.com/sftsrv/oapi/internal/openapi"
	"github.com/sftsrv/oapi/internal/theme"
)

type Model struct {
	spec     *openapi.OpenAPISpec
	path     string
	ready    bool
	width    int
	height   int
	renderer *glamour.TermRenderer
	viewport viewport.Model
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

func newRenderer(width int) *glamour.TermRenderer {
	hasDarkBG := lg.HasDarkBackground(os.Stdin, os.Stdout)

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

func (s Model) preview() tea.Cmd {
	return func() tea.Msg {
		md := ""

		if s.path == "" {
			md += "# " + s.spec.Info.Title
			md += "\n\n"
			md += "> version: " + s.spec.Info.Version
		} else {
			parts := strings.Split(s.path, "/")
			pathParts := parts[:len(parts)-1]
			method := parts[len(parts)-1]

			path := strings.Join(pathParts, "/")
			p, ok := s.spec.Paths[path]

			if !ok {
				return nil
			}

			ep, ok := p[method]
			if !ok {
				return nil
			}

			md += "# " + strings.ToUpper(method) + ":" + path
			md += "\n\n"
			md += "> " + ep.Summary
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
			Background(theme.ColorSecondary).
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
