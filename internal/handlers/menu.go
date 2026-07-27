package handlers

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/donderom/bubblon"
	"github.com/wisp-trading/wisp/internal/router"
	"github.com/wisp-trading/wisp/internal/setup/types"
	"github.com/wisp-trading/wisp/internal/ui"
)

// mainMenuModel represents the main menu TUI
type mainMenuModel struct {
	choices  []string
	cursor   int
	router   router.Router
	scaffold types.ScaffoldService
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
			switch m.choices[m.cursor] {
			case "Strategies":
				return m, m.router.NavigateTo(router.RouteStrategyList)
			case "Monitor":
				return m, m.router.NavigateTo(router.RouteMonitor)
			case "Settings":
				return m, m.router.NavigateTo(router.RouteSettingsList)
			case "Create New Project":
				return m, bubblon.Open(NewProjectCreateView(m.router, m.scaffold))
			case "Help":
				return m, bubblon.Open(NewHelpView(m.router))
			}
		}
	}
	return m, nil
}

func (m mainMenuModel) View() string {
	title := ui.TitleCenteredStyle.Render("WISP")

	var s string
	s += "\n" + title + "\n\n"
	s += ui.MutedStyle.Render("Project · Keys · Live strategies") + "\n\n"

	icons := []string{"📂", "📊", "⚙️", "🆕", "ℹ️"}

	for i, choice := range m.choices {
		cursor := "  "
		if m.cursor == i {
			cursor = "▶ "
			s += ui.SelectedItemStyle.Render(cursor+icons[i]+" "+choice) + "\n"
		} else {
			s += ui.ItemStyle.Render(cursor+icons[i]+" "+choice) + "\n"
		}
	}

	s += "\n" + ui.MutedStyle.Render("↑↓ Navigate  ↵ Select  q Quit")

	return ui.MenuBoxStyle.Render(s)
}
