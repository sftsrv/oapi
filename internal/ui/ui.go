package ui

import (
	"fmt"
	"maps"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	lg "github.com/charmbracelet/lipgloss"
	"github.com/sftsrv/oapi/internal/openapi"
	"github.com/sftsrv/oapi/internal/picker"
	"github.com/sftsrv/oapi/internal/preview"
	"github.com/sftsrv/oapi/internal/theme"
	"github.com/sftsrv/oapi/internal/tree"
)

type window struct {
	width  int
	height int
}

type Path string

type Model struct {
	window window

	spec *openapi.OpenAPISpec
	tree *tree.Tree

	hovered    *tree.Item
	selected   *tree.Item
	pathPicker picker.Model[*tree.Item]

	preview preview.Model
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (w *window) updateWindowSize(width int, height int) {
	w.width = width
	w.height = height
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.window.updateWindowSize(msg.Width, msg.Height)

		pickerWidth := int(float32(msg.Width)*0.25) - 1
		previewWidth := int(float32(msg.Width)*0.75) - 1

		m.pathPicker = m.pathPicker.Height(msg.Height - 1).Width(pickerWidth)
		m.preview, cmd = m.preview.Height(msg.Height - 1).Width(previewWidth).Update(msg)
		return m, cmd

	case picker.SelectedMsg[*tree.Item]:
		m.selected = msg.Selected
		return m, tea.Quit

	case picker.HoverMsg[*tree.Item]:
		m.hovered = msg.Hovered

		preview, c := m.preview.SetPath(msg.Hovered.GetPath())
		m.preview = preview

		return m, c

	case preview.PreviewReadyMsg:
		hovered := m.hovered
		if hovered != nil {
			preview, c := m.preview.SetPath(hovered.GetPath())

			m.preview = preview
			return m, c
		}

	case preview.PreviewResultMsg:
		m.preview = m.preview.SetContent(msg)
		return m, nil

	case tea.KeyMsg:
		str := msg.String()

		switch str {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "left", "h":
			if m.hovered != nil && (str == "left" || !m.pathPicker.IsSearching()) {
				m.hovered.Collapse()
				m.pathPicker, cmd = m.pathPicker.Items(tree.ToItems(m.tree)).Update(msg)
				return m, cmd
			}

		case "right", "l":
			if m.hovered != nil && (str == "right" || !m.pathPicker.IsSearching()) {
				m.hovered.Expand()
				m.pathPicker, cmd = m.pathPicker.Items(tree.ToItems(m.tree)).Update(msg)
				return m, cmd
			}

		case "]":
			if !m.pathPicker.IsSearching() {
				m.tree.ExpandAll()
				m.pathPicker, cmd = m.pathPicker.Items(tree.ToItems(m.tree)).Update(msg)
				return m, cmd
			}

		case "[":
			if !m.pathPicker.IsSearching() {
				m.tree.CollapseAll()
				m.pathPicker, cmd = m.pathPicker.Items(tree.ToItems(m.tree)).Update(msg)
				return m, cmd
			}

		case "{":
			pathPicker, pathPickerCmd := m.pathPicker.Update(picker.ResizeMsg{Adjust: -1})
			preview, previewCmd := m.preview.Update(preview.ResizeMsg{Adjust: +1})

			m.pathPicker = pathPicker
			m.preview = preview

			return m, tea.Batch(pathPickerCmd, previewCmd)

		case "}":
			pathPicker, pathPickerCmd := m.pathPicker.Update(picker.ResizeMsg{Adjust: +1})
			preview, previewCmd := m.preview.Update(preview.ResizeMsg{Adjust: -1})

			m.pathPicker = pathPicker
			m.preview = preview

			return m, tea.Batch(pathPickerCmd, previewCmd)
		}

	case tea.MouseMsg:
		m.preview, cmd = m.preview.Update(msg)
		return m, cmd

	}

	m.pathPicker, cmd = m.pathPicker.Update(msg)
	return m, cmd
}

func helpView(m Model) string {
	item := func(icon string, title string) string {
		return lg.JoinHorizontal(
			lg.Top,
			theme.Secondary.MarginLeft(2).Render(icon),
			theme.Faded.MarginRight(2).Render(" ", title),
		)
	}

	help := ""
	if m.pathPicker.IsSearching() {
		help += item("esc", "close search")
		help += item("↓↑", "navigate")
		help += item("→", "expand")
		help += item("←", "collapse")
		help += item("ctrk+c", "quit")
	} else {
		help += item("/", "search")
		help += item("↓↑/jk", "navigate")
		help += item("→/l", "expand")
		help += item("←/h", "collapse")
		help += item("]/[", "expand/collapse all")
		help += item("}/{", "resize")
		help += item("ctrk+c/q", "quit")
	}

	return lg.NewStyle().Render(help)
}

func (m Model) View() string {
	return lg.JoinVertical(
		lg.Left,
		lg.JoinHorizontal(lg.Top, m.pathPicker.View(), m.preview.View()),
		helpView(m),
	)
}

func initialModel(spec *openapi.OpenAPISpec, f *tree.Tree) Model {
	items := tree.ToItems(f)

	return Model{
		spec:       spec,
		tree:       f,
		pathPicker: picker.New[*tree.Item]().Title("Items").Accent(theme.ColorPrimary).Items(items),
		preview:    preview.New(),
	}
}

func endpointTree(spec *openapi.OpenAPISpec) *tree.Tree {
	endpoints := []string{
		"/",
	}
	paths := maps.Keys(spec.Paths)

	for path := range paths {
		methods := maps.Keys(spec.Paths[path])
		for method := range methods {
			endpoint := fmt.Sprintf("%s/%s", path, strings.ToUpper(method))
			endpoints = append(endpoints, endpoint)
		}
	}

	return tree.PathsToTree(endpoints)
}

func Run(spec *openapi.OpenAPISpec, flat bool) {
	t := endpointTree(spec)

	if flat {
		t.Flatten()
	}

	m := initialModel(spec, t)
	p := tea.NewProgram(m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	result, err := p.Run()
	if err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}

	selected := result.(Model).selected

	if selected != nil {
		fmt.Println(result.(Model).selected.GetPath())
	}
}
