// Package ui implements the Bubble Tea TUI model, update loop, and view renderer.
package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/marko-stanojevic/hostr/internal/sysinfo"
)

// Styles are defined in styles.go

// ──────────────────────────────────────────────
// Messages
// ──────────────────────────────────────────────

type tickMsg time.Time
type infoMsg sysinfo.Info
type errMsg error

// ──────────────────────────────────────────────
// Model
// ──────────────────────────────────────────────

// Model is the root Bubble Tea model for the sysinfo TUI.
type Model struct {
	collector sysinfo.Collector
	info      sysinfo.Info
	spinner   spinner.Model
	loading   bool
	err       error
	width     int
	height    int
}

// NewModel constructs a Model with sensible defaults.
func NewModel() Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return Model{
		collector: sysinfo.NewCollector(),
		spinner:   s,
		loading:   true,
	}
}

// ──────────────────────────────────────────────
// Init
// ──────────────────────────────────────────────

// Init starts the spinner and kicks off the first data collection.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.collectCmd(),
	)
}

// ──────────────────────────────────────────────
// Update
// ──────────────────────────────────────────────

// Update handles all incoming messages and returns the updated model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			m.loading = true
			return m, tea.Batch(m.spinner.Tick, m.collectCmd())
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case infoMsg:
		m.info = sysinfo.Info(msg)
		m.loading = false
		m.err = nil // Clear previous errors on success
		return m, tickCmd()

	case errMsg:
		// Retain previous data and log error, don't blank the display
		m.err = msg
		m.loading = false

	case tickMsg:
		return m, m.collectCmd()

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

// ──────────────────────────────────────────────
// View
// ──────────────────────────────────────────────

// View renders the current model state as a string.
func (m Model) View() string {
	if m.loading && m.info.CollectedAt.IsZero() {
		return fmt.Sprintf("\n  %s Loading system info…\n", m.spinner.View())
	}
	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("\n  Error: %v\n\nPress q to quit.", m.err))
	}

	var b strings.Builder

	b.WriteString(titleStyle.Render("⬡  System Info") + "\n")

	// System section
	sysLines := []string{
		row("Hostname", m.info.Hostname),
		row("OS / Arch", fmt.Sprintf("%s / %s", m.info.OS, m.info.Arch)),
		row("Go Version", m.info.GoVersion),
		row("Uptime", m.info.Uptime),
	}
	b.WriteString(section("System", sysLines))

	// CPU section
	cpuLines := []string{
		row("Model", m.info.CPUModel),
		row("Cores", fmt.Sprintf("%d", m.info.CPUCores)),
		row("Usage", progressBar(m.info.CPUUsage, 30)),
	}
	b.WriteString(section("CPU", cpuLines))

	// Memory section
	memLines := []string{
		row("Total", m.info.MemTotal),
		row("Used", m.info.MemUsed),
		row("Usage", progressBar(m.info.MemPercent, 30)),
	}
	b.WriteString(section("Memory", memLines))

	// Disk section
	diskLines := []string{
		row("Total", m.info.DiskTotal),
		row("Used", m.info.DiskUsed),
		row("Usage", progressBar(m.info.DiskPercent, 30)),
	}
	b.WriteString(section("Disk", diskLines))

	ts := m.info.CollectedAt.Format("15:04:05")
	if m.loading {
		ts += "  " + m.spinner.View()
	}
	b.WriteString(footerStyle.Render(fmt.Sprintf("Updated %s   •   r refresh   •   q quit", ts)))

	return b.String()
}

// ──────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────

func section(title string, lines []string) string {
	content := headerStyle.Render(title) + "\n" + strings.Join(lines, "\n")
	return boxStyle.Render(content) + "\n"
}

func row(label, value string) string {
	return "  " + labelStyle.Render(label) + valueStyle.Render(value)
}

func progressBar(percent float64, width int) string {
	filled := int(percent / 100 * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled

	bar := barFilledStyle.Render(strings.Repeat("█", filled)) +
		barEmptyStyle.Render(strings.Repeat("░", empty))

	return fmt.Sprintf("%s %4.1f%%", bar, percent)
}

// ──────────────────────────────────────────────
// Commands
// ──────────────────────────────────────────────

// collectCmd runs metric collection off the main goroutine.
// It uses the model's configured collector (allowing testing/mocking).
func (m Model) collectCmd() tea.Cmd {
	return func() tea.Msg {
		info, err := m.collector.Collect(context.Background())
		if err != nil {
			return errMsg(err)
		}
		return infoMsg(info)
	}
}

// tickCmd schedules the next automatic refresh (every 3 seconds).
func tickCmd() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
