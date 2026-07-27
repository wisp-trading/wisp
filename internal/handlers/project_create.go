package handlers

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wisp-trading/wisp/internal/router"
	"github.com/wisp-trading/wisp/internal/setup/types"
	"github.com/wisp-trading/wisp/internal/ui"
)

// projectCreateModel is a small in-router flow: name → scaffold starter strategy.
type projectCreateModel struct {
	router   router.Router
	scaffold types.ScaffoldService
	name     string
	err      error
	done     bool
	doneMsg  string
}

func NewProjectCreateView(r router.Router, scaffold types.ScaffoldService) tea.Model {
	return &projectCreateModel{router: r, scaffold: scaffold}
}

func (m *projectCreateModel) Init() tea.Cmd { return nil }

func (m *projectCreateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.done {
			switch msg.String() {
			case "q", "esc", "enter", "ctrl+c":
				return m, m.router.Back()
			}
			return m, nil
		}
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc", "q":
			return m, m.router.Back()
		case "enter":
			name := strings.TrimSpace(strings.ReplaceAll(m.name, " ", "_"))
			if name == "" {
				m.err = fmt.Errorf("project name cannot be empty")
				return m, nil
			}
			if err := m.scaffold.CreateProject(name); err != nil {
				m.err = err
				return m, nil
			}
			m.done = true
			m.err = nil
			m.doneMsg = fmt.Sprintf(
				"Created ./%s\n\nNext:\n  1. Settings → add exchange keys\n  2. Strategies → starter → Start Live\n  or: cd %s/strategies/starter && go mod tidy && go run .",
				name, name,
			)
			return m, nil
		case "backspace":
			if len(m.name) > 0 {
				m.name = m.name[:len(m.name)-1]
			}
			m.err = nil
		default:
			if len(msg.String()) == 1 {
				c := msg.String()[0]
				if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
					(c >= '0' && c <= '9') || c == '_' || c == '-' || c == ' ' {
					m.name += msg.String()
					m.err = nil
				}
			}
		}
	}
	return m, nil
}

func (m *projectCreateModel) View() string {
	title := ui.TitleStyle.Render("🆕  Create New Project")
	if m.done {
		body := ui.StatusReadyStyle.Render("✓ "+m.doneMsg) + "\n\n" +
			ui.MutedStyle.Render("↵ / q  Back to menu")
		return ui.MenuBoxStyle.Width(72).Render("\n" + title + "\n\n" + body + "\n")
	}

	var s string
	s += "\n" + title + "\n\n"
	s += ui.MutedStyle.Render("Creates strategies/starter/ with main.go (standalone).") + "\n"
	s += ui.MutedStyle.Render("Keys stay in Settings (~/.wisp/connectors.yml) — not in the project.") + "\n\n"
	s += ui.LabelStyle.Width(0).Render("Project name: ") + ui.InputStyle.Render(m.name+"_") + "\n\n"
	if m.err != nil {
		s += ui.ErrorBoxStyle.Width(0).Render("✗ "+m.err.Error()) + "\n\n"
	}
	s += ui.MutedStyle.Render("↵ Create   q Back")
	return ui.MenuBoxStyle.Width(72).Render(s)
}
