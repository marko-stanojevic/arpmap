package ui

import "github.com/charmbracelet/lipgloss"

// Style definitions for the sysinfo TUI.
// Separated from model.go for maintainability.

var (
	// Title style for main heading.
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(1)

	// Header style for section titles.
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39"))

	// Label style for metric names (left column).
	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			Width(16)

	// Value style for metric values (right column).
	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255"))

	// Bar filled portion of progress bars.
	barFilledStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205"))

	// Bar empty portion of progress bars.
	barEmptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("237"))

	// Footer style for status line.
	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			MarginTop(1)

	// Error style for error messages.
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	// Box style for section containers.
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1).
			MarginBottom(1)
)
