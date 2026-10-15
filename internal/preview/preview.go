package preview

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	lg "github.com/charmbracelet/lipgloss"
	"github.com/sftsrv/oapi/internal/openapi"
	"github.com/sftsrv/oapi/internal/theme"
)

type Model struct {
	spec     *openapi.OpenAPISpec
	path     string
	ready    bool
	width    int
	height   int
	viewport viewport.Model
}

type PreviewResultMsg struct {
	path    string
	isError bool
	content string
}

type PreviewReadyMsg struct{}

type ResizeMsg struct {
	Adjust int
}

func New() Model {
	return Model{}
}

func (m Model) SetPath(path string) (Model, tea.Cmd) {
	m.path = path

	cmd := m.preview()

	return m, cmd
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
		if s.path == "" {
			return PreviewResultMsg{s.path, false, "top-level-preview"}
		}

		parts := strings.Split(s.path, "/")
		path := parts[:len(parts)-1]
		method := parts[len(parts)-1]

		content := strings.Join(path, "/") + "\n" + method

		return PreviewResultMsg{s.path, false, content}
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
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {

	case ResizeMsg:
		m.width += msg.Adjust
		m.viewport.Width = m.width

	case tea.WindowSizeMsg:
		if !m.ready {
			// Since this program is using the full size of the viewport we
			// need to wait until we've received the window dimensions before
			// we can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.
			m.viewport = viewport.New(m.width, m.height-1)
			m.viewport.YPosition = 0
			m.viewport.SetContent("")
			m.ready = true
			readyCmd := func() tea.Msg {
				return PreviewReadyMsg{}
			}

			cmds = append(cmds, readyCmd)
		} else {
			m.viewport.Width = m.width
			m.viewport.Height = m.height
		}
	}

	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}
