package handlers

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/wisp-trading/wisp/internal/router"
	"github.com/wisp-trading/wisp/internal/ui"
)

// helpModel is a short in-app guide for the three core flows.
type helpModel struct {
	router router.Router
}

func NewHelpView(r router.Router) tea.Model {
	return &helpModel{router: r}
}

func (m *helpModel) Init() tea.Cmd { return nil }

func (m *helpModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "enter", "ctrl+c":
			return m, m.router.Back()
		}
	}
	return m, nil
}

func (m *helpModel) View() string {
	title := ui.TitleStyle.Render("ℹ️  How to use Wisp")
	body := ui.ItemStyle.Render(`
Three things this CLI is for:

1. New project
   Menu → Create New Project
   Or:  wisp init my-bot
   Creates strategies/<name>/ with main.go (StartStandalone + Wait).
   No API keys in the project.

2. Exchange keys
   Menu → Settings
   Add Hyperliquid / Polymarket / … credentials.
   Saved to ~/.wisp/connectors.yml (shared by every strategy).

3. Live trading
   Menu → Strategies → open a strategy → Start Live
   Compiles the binary and runs it.
   Menu → Monitor → select instance → Stop (HTTP shutdown, then force if needed).

Tips
  • Strategy config.yml only lists exchanges/assets — never secrets.
  • Prefer Hyperliquid perps for production; other venues are beta/experimental.
  • q goes back; Ctrl+C quits the app from the main menu.
`)
	help := ui.MutedStyle.Render("↵ / q  Back")
	return ui.MenuBoxStyle.Width(72).Render("\n" + title + "\n" + body + "\n" + help + "\n")
}
