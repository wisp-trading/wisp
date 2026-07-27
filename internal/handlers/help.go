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
   Creates strategies/starter/ with main.go (StartStandalone + Wait).
   No API keys in the project.

2. Exchange keys
   Menu → Settings → pick an exchange → fill fields → Save
   Saved to ~/.wisp/connectors.yml (shared by every strategy).

3. Live trading
   Menu → Strategies → open a strategy → Start Live
   Compiles the standalone binary and spawns it.
   Menu → Monitor → select instance → S Stop
   (graceful HTTP shutdown, force only if needed).

Tips
  • Strategy config.yml lists exchanges/assets only — never secrets.
  • Prefer Hyperliquid perps for production; other venues are beta.
  • Strategies: ←→ pages · click row · r refresh · q back
  • Settings form: ↑↓/Tab fields · Enter next/save · Ctrl+X/Esc cancel
  • Main menu: 1–5 jump · q quit
`)
	help := ui.MutedStyle.Render("↵ / q / Esc  Back")
	return ui.MenuBoxStyle.Width(72).Render("\n" + title + "\n" + body + "\n" + help + "\n")
}
