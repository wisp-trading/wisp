package settings

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wisp-trading/sdk/pkg/types/config"
	"github.com/wisp-trading/wisp/internal/router"
	"github.com/wisp-trading/wisp/internal/ui"
)

// DeleteConfirmModel — simple Stay/Leave-style confirm (no huh).
type DeleteConfirmModel struct {
	connectorName string
	config        config.Configuration
	router        router.Router
	cursor        int // 0 cancel, 1 delete
	err           error
}

func NewDeleteConfirmView(
	cfg config.Configuration,
	r router.Router,
	connectorName string,
) tea.Model {
	return &DeleteConfirmModel{
		config:        cfg,
		router:        r,
		connectorName: connectorName,
		cursor:        0,
	}
}

func (m *DeleteConfirmModel) Init() tea.Cmd { return nil }

func (m *DeleteConfirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.err != nil {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "q" || msg.String() == "esc" || msg.String() == "enter" {
				return m, m.router.Back()
			}
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h", "up", "k":
			m.cursor = 0
		case "right", "l", "down", "j", "tab":
			m.cursor = 1
		case "esc", "q", "n", "N", "ctrl+x":
			return m, m.router.Back()
		case "y", "Y":
			return m, m.doDelete()
		case "enter", " ":
			if m.cursor == 1 {
				return m, m.doDelete()
			}
			return m, m.router.Back()
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *DeleteConfirmModel) doDelete() tea.Cmd {
	if err := m.config.RemoveConnector(m.connectorName); err != nil {
		m.err = err
		return nil
	}
	return m.router.Back()
}

func (m *DeleteConfirmModel) View() string {
	if m.err != nil {
		return ui.ErrorBoxStyle.Render("❌ "+m.err.Error()) + "\n\n" +
			ui.MutedStyle.Render("q / Esc back")
	}
	var cancel, del string
	if m.cursor == 0 {
		cancel = ui.SelectedItemStyle.Render("[ Cancel ]")
		del = ui.MutedStyle.Render("  Delete  ")
	} else {
		cancel = ui.MutedStyle.Render("  Cancel  ")
		del = ui.SelectedItemStyle.Render("[ Delete ]")
	}
	body := fmt.Sprintf(
		"Delete connector %s?\n\nThis cannot be undone.\n\n%s   %s\n\n%s",
		m.connectorName,
		cancel,
		del,
		ui.MutedStyle.Render("←→ choose  ↵ confirm  y delete  n/Esc cancel"),
	)
	return ui.ErrorBoxStyle.Width(52).Render(body)
}
