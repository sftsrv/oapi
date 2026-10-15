package picker

import (
	"fmt"
	"image/color"
	"strings"

	tea "charm.land/bubbletea/v2"
	lg "charm.land/lipgloss/v2"
	"github.com/sftsrv/oapi/internal/theme"
)

type Model[I Item] struct {
	accent    color.Color
	title     string
	search    string
	searching bool
	cursor    int
	count     int
	items     []I
	filtered  []I
	width     int
}

func New[I Item]() Model[I] {
	return Model[I]{
		accent: theme.ColorPrimary,
		count:  5,
	}
}

func (m Model[I]) Items(items []I) Model[I] {
	m.items = items
	return m.applyFilter()
}

// The height of the picker is header + count == 1 + count
func (m Model[I]) GetHeight() int {
	return 2 + m.count
}

// The count depends on how much space we have
func (m Model[I]) Height(height int) Model[I] {
	m.count = height - 2
	return m.applyFilter()
}

func (m Model[I]) Width(width int) Model[I] {
	m.width = width
	return m
}

func (m Model[I]) GetWidth() int {
	return m.width
}

// if the search is changed externally then filters need to be re-applied
func (m Model[I]) Search(search string) Model[I] {
	m.search = search
	return m.applyFilter()
}

func (m Model[I]) Searching(searching bool) Model[I] {
	m.searching = searching
	return m.applyFilter()
}

func (m Model[I]) IsSearching() bool {
	return m.searching
}

func (m Model[I]) Accent(accent color.Color) Model[I] {
	m.accent = accent
	return m
}

func (m Model[I]) Title(title string) Model[I] {
	m.title = title
	return m
}

func (_ Model[I]) Init() tea.Cmd {
	return nil
}

func truncate(str string, width int) string {
	return lg.NewStyle().
		MaxHeight(1).
		MaxWidth(width).
		Render(str)
}

func indicator(accent color.Color, selected bool, title string) string {
	if !selected {
		return lg.NewStyle().
			PaddingRight(1).
			Render("") + title
	}

	line := lg.JoinHorizontal(
		lg.Top,
		lg.NewStyle().
			PaddingRight(0).
			Foreground(accent).
			Render("→"),
		lg.NewStyle().
			Foreground(accent).
			Bold(true).
			Render(title),
	)

	return line
}

func (m Model[I]) View() string {
	count := fmt.Sprintf("(%d/%d)", m.cursor+1, len(m.filtered))

	fallback := "/ to search"
	if m.search != "" {
		fallback = m.search
	}

	header := lg.JoinVertical(lg.Left,
		theme.
			Heading.
			Width(m.width).
			Background(m.accent).
			Render(m.title+" "+count),
		theme.Faded.MarginLeft(1).Render(fallback),
	)

	if m.searching {
		header = lg.JoinVertical(
			lg.Left,
			theme.Heading.Width(m.width).
				Background(m.accent).
				Render("Search "+count),
			m.search+"_",
		)
	}

	cursor, items := m.cursorWindow()
	content := []string{}

	for i, item := range items {
		content = append(content, truncate(indicator(m.accent, i == cursor, item), m.width-1))
	}

	if len(items) < m.count {
		content = append(content, theme.Faded.Render("no more items"))
	}

	return lg.NewStyle().
		Width(m.width).
		Render(
			lg.JoinVertical(
				lg.Top,
				header,
				lg.NewStyle().
					BorderLeft(true).
					BorderForeground(m.accent).
					BorderStyle(lg.NormalBorder()).
					Render(lg.JoinVertical(lg.Top, content...)),
			))
}

// Gets the cursor position in a relative window with one item padding if possible.
// Prefers to keep cursor at the top
func (m Model[I]) cursorWindow() (int, []string) {
	itemCount := len(m.filtered)

	if m.cursor < 2 {
		items := m.filtered[0:min(m.count, itemCount)]
		return m.cursor, getTitles(items)
	}

	if m.cursor > itemCount-1 {
		items := m.filtered[max(0, itemCount-m.count-1):itemCount]
		lastItem := len(items) - 1
		return lastItem, getTitles(items)
	}

	first := m.cursor - 1
	last := min(m.cursor+m.count-1, itemCount)
	items := m.filtered[first:last]

	return 1, getTitles(items)

}

func (m Model[I]) applyFilter() Model[I] {
	if m.search == "" {
		m.filtered = m.items
		return m
	}

	m.filtered = []I{}
	for _, item := range m.items {
		if strings.Contains(item.Search(), m.search) {
			m.filtered = append(m.filtered, item)
		}
	}

	return m
}

func (m Model[I]) cursorUp() Model[I] {
	maxIndex := max(len(m.filtered)-1, 0)
	m.cursor = clamp(m.cursor-1, 0, maxIndex)

	return m
}

func (m Model[I]) cursorDown() Model[I] {
	maxIndex := max(len(m.filtered)-1, 0)
	m.cursor = clamp(m.cursor+1, 0, maxIndex)

	return m
}

type SelectedMsg[I Item] struct {
	Selected I
}

type HoverMsg[I Item] struct {
	Hovered I
}

type ResizeMsg struct {
	Adjust int
}

func (m Model[I]) selectedMsg() tea.Cmd {
	if m.cursor >= len(m.filtered) {
		return nil
	}

	return func() tea.Msg {
		return SelectedMsg[I]{
			m.filtered[m.cursor],
		}
	}
}

func (m Model[I]) hoverMsg() tea.Cmd {
	if m.cursor >= len(m.filtered) {
		return nil
	}

	return func() tea.Msg {
		return HoverMsg[I]{
			m.filtered[m.cursor],
		}
	}
}

func (m Model[I]) Update(msg tea.Msg) (Model[I], tea.Cmd) {
	switch msg := msg.(type) {
	case ResizeMsg:
		m.width += msg.Adjust

	case tea.KeyMsg:
		str := msg.String()
		if m.searching {
			switch str {

			case "up":
				m = m.cursorUp()

			case "down":
				m = m.cursorDown()

			case "esc":
				m.searching = false

			case "enter":
				return m, m.selectedMsg()

			case "backspace":
				if m.search != "" {
					m.search = m.search[0 : len(m.search)-1]
					m = m.applyFilter()
				}

			default:
				if len(str) == 1 {
					m.search += str
					m = m.applyFilter()
				}
			}
		} else {
			switch str {
			case "up", "k":
				m = m.cursorUp()

			case "down", "j":
				m = m.cursorDown()

			case " ", "enter":
				return m, m.selectedMsg()

			case "esc":
				m.searching = false

			case "/":
				m.searching = true
			}
		}
	}

	return m, m.hoverMsg()
}

func clamp(i int, min int, max int) int {
	if i > max {
		return max
	}

	if i < min {
		return min
	}

	return i
}
