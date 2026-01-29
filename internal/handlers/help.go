package handlers

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/wisp-trading/wisp/internal/ui"
)

// helpModel represents the help screen TUI
type helpModel struct {
	scrollOffset   int
	viewportHeight int
	quitting       bool
}

func (m helpModel) Init() tea.Cmd {
	return nil
}

func (m helpModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewportHeight = msg.Height - 6
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "enter", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.scrollOffset > 0 {
				m.scrollOffset--
			}
		case "down", "j":
			m.scrollOffset++
		}
	}
	return m, nil
}

func (m helpModel) View() string {

	// Build content as lines for scrolling
	lines := []string{}
	lines = append(lines, ui.TitleCenteredStyle.Render("🚀 WISP CLI v0.1.0"))
	lines = append(lines, ui.MutedStyle.Render("Trading infrastructure platform"))
	lines = append(lines, "")
	lines = append(lines, ui.SectionHeaderStyle.Render("📋 COMMANDS"))
	lines = append(lines, "")

	commands := []struct{ cmd, desc string }{
		{"wisp", "Launch interactive menu"},
		{"wisp init <name>", "Create a new trading project"},
		{"wisp backtest", "Run backtests interactively"},
		{"wisp live", "Deploy strategies to live trading"},
		{"wisp analyze", "Analyze backtest results"},
		{"wisp version", "Show version information"},
	}

	for _, c := range commands {
		lines = append(lines, "  "+ui.CommandStyle.Render(c.cmd))
		lines = append(lines, "    "+ui.TextStyle.Render(c.desc))
		lines = append(lines, "")
	}

	// Handle scrolling
	start := m.scrollOffset
	end := len(lines)
	if m.viewportHeight > 0 && start+m.viewportHeight < end {
		end = start + m.viewportHeight
	}
	if start > len(lines) {
		start = len(lines)
	}
	if end > len(lines) {
		end = len(lines)
	}

	visibleLines := lines[start:end]
	var s string
	for _, line := range visibleLines {
		s += line + "\n"
	}

	// Scroll indicators
	if start > 0 {
		s = ui.MutedStyle.Render("↑ Scroll up for more") + "\n" + s
	}
	if end < len(lines) {
		s += ui.MutedStyle.Render("↓ Scroll down for more") + "\n"
	}

	s += "\n" + ui.MutedStyle.Render("↑↓/jk Scroll  q/esc/enter Exit")

	return "\n" + s + "\n"
}

func showHelp() error {
	m := helpModel{viewportHeight: 20}
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
