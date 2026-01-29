package handlers

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/wisp-trading/wisp/internal/router"
	handlers2 "github.com/wisp-trading/wisp/internal/setup/handlers"
)

// mainMenuModel represents the main menu TUI
type mainMenuModel struct {
	choices  []string
	cursor   int
	selected string
	router   router.Router
}

func (m mainMenuModel) Init() tea.Cmd {
	return nil
}

func (m mainMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter":
			// Navigate using the router instead of quitting
			switch m.choices[m.cursor] {
			case "Strategies":
				return m, func() tea.Msg {
					return router.NavigateMsg{Route: router.RouteStrategyList}
				}
			case "Monitor":
				return m, func() tea.Msg {
					return router.NavigateMsg{Route: router.RouteMonitor}
				}
			case "Settings":
				return m, func() tea.Msg {
					return router.NavigateMsg{Route: router.RouteSettingsList}
				}
			case "Help", "Create New Project":
				// TODO: Register these routes when implemented
				return m, nil
			}
		}
	}
	return m, nil
}

func (m mainMenuModel) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00D9FF")).
		PaddingTop(1).
		PaddingBottom(1).
		Align(lipgloss.Center)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7C3AED")).
		Padding(2, 4).
		Width(50)

	itemStyle := lipgloss.NewStyle().
		PaddingLeft(2)

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00D9FF")).
		Bold(true).
		PaddingLeft(0)

	mutedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Italic(true)

	title := titleStyle.Render("WISP CLI v0.1.0")

	var s string
	s += "\n" + title + "\n\n"
	s += mutedStyle.Render("What would you like to do?") + "\n\n"

	icons := []string{"📂", "📊", "⚙️", "ℹ️", "🆕"}

	for i, choice := range m.choices {
		cursor := "  "
		if m.cursor == i {
			cursor = "▶ "
			s += selectedStyle.Render(cursor+icons[i]+" "+choice) + "\n"
		} else {
			s += itemStyle.Render(cursor+icons[i]+" "+choice) + "\n"
		}
	}

	s += "\n" + mutedStyle.Render("↑↓/jk Navigate  ↵ Select  q Quit")

	return boxStyle.Render(s)
}

func (h *rootHandler) handleCreateProject(cmd *cobra.Command) error {
	// Run the init TUI flow
	strategyExample, projectName, err := handlers2.RunInitTUI()
	if err != nil {
		return err
	}

	// Create the project with the selected strategy
	return h.initHandler.HandleWithStrategy(strategyExample, projectName)
}
