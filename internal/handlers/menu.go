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

func (m mainMenuModel) activate() (tea.Model, tea.Cmd) {
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
	return m, nil
}

func (m mainMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			// Menu items start ~4 lines into the box (title + blank + prompt + blank).
			// Approximate: row relative to content; each item is one line.
			// With alt-screen mouse coords are absolute — map loosely to choice index.
			// First choice roughly at Y offset after header (~5–6).
			idx := msg.Y - 6
			if idx >= 0 && idx < len(m.choices) {
				m.cursor = idx
				return m.activate()
			}
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter", " ":
			return m.activate()
		case "1", "2", "3", "4", "5":
			i := int(msg.String()[0] - '1')
			if i >= 0 && i < len(m.choices) {
				m.cursor = i
				return m.activate()
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

	s += "\n" + ui.MutedStyle.Render("↑↓ / 1-5  Navigate   ↵ / click  Select   q/Esc  Quit")

	return ui.MenuBoxStyle.Render(s)
}
