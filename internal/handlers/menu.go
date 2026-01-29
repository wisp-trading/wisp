package handlers

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/wisp-trading/wisp/internal/router"
	handlers2 "github.com/wisp-trading/wisp/internal/setup/handlers"
	"github.com/wisp-trading/wisp/internal/ui"
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
	title := ui.TitleCenteredStyle.Render("WISP CLI v0.1.0")

	var s string
	s += "\n" + title + "\n\n"
	s += ui.MutedStyle.Render("What would you like to do?") + "\n\n"

	icons := []string{"📂", "📊", "⚙️", "ℹ️", "🆕"}

	for i, choice := range m.choices {
		cursor := "  "
		if m.cursor == i {
			cursor = "▶ "
			s += ui.SelectedItemStyle.Render(cursor+icons[i]+" "+choice) + "\n"
		} else {
			s += ui.ItemStyle.Render(cursor+icons[i]+" "+choice) + "\n"
		}
	}

	s += "\n" + ui.MutedStyle.Render("↑↓/jk Navigate  ↵ Select  q Quit")

	return ui.MenuBoxStyle.Render(s)
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
