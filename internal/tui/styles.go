package tui

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/ystsbry/bako/internal/model"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4"))

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	cursorRowStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#3C3C5A"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8A8A8A"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8A8A8A"))

	errStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F87")).
			Bold(true)

	okStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#5FD787"))

	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 2)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#8A8A8A")).
				Padding(0, 2)

	fieldLabelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#B4A7F5"))
)

// statusStyle returns the badge style for a PBI status.
func statusStyle(s model.Status) lipgloss.Style {
	base := lipgloss.NewStyle().Bold(true).Padding(0, 1)
	switch s {
	case model.StatusTodo:
		return base.Foreground(lipgloss.Color("#1C1C1C")).Background(lipgloss.Color("#D7AF5F"))
	case model.StatusProgress:
		return base.Foreground(lipgloss.Color("#1C1C1C")).Background(lipgloss.Color("#5FAFFF"))
	case model.StatusDone:
		return base.Foreground(lipgloss.Color("#1C1C1C")).Background(lipgloss.Color("#5FD787"))
	default:
		return base
	}
}

// statusBadge renders a fixed-width status badge for list rows.
func statusBadge(s model.Status) string {
	return statusStyle(s).Render(padRight(s.Label(), 8))
}

// padRight pads s with spaces to at least n runes.
func padRight(s string, n int) string {
	for len([]rune(s)) < n {
		s += " "
	}
	return s
}
