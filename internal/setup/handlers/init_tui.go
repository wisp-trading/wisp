package handlers

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wisp-trading/wisp/internal/setup/services"
	"github.com/wisp-trading/wisp/internal/ui"
)

// StrategyTemplate is the init picker row (maps to strategies/<SDKExample>/).
type StrategyTemplate struct {
	Name        string
	DisplayName string
	Description string
	Icon        string
	SDKExample  string
}

// InitTUIModel — project name only (single starter template; no fake multi-template fetch).
type InitTUIModel struct {
	selectedStrategy StrategyTemplate
	projectName      string
	projectNameInput string
	err              error
	done             bool
	cancelled        bool
}

func NewInitTUIModel(strategy StrategyTemplate) InitTUIModel {
	return InitTUIModel{selectedStrategy: strategy}
}

func (m InitTUIModel) Init() tea.Cmd { return nil }

func (m InitTUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.cancelled = true
			return m, tea.Quit
		case "enter":
			if strings.TrimSpace(m.projectNameInput) == "" {
				m.err = fmt.Errorf("project name cannot be empty")
				return m, nil
			}
			m.projectName = strings.ReplaceAll(strings.TrimSpace(m.projectNameInput), " ", "_")
			m.done = true
			m.err = nil
			return m, tea.Quit
		case "backspace":
			if len(m.projectNameInput) > 0 {
				m.projectNameInput = m.projectNameInput[:len(m.projectNameInput)-1]
			}
			m.err = nil
		default:
			if len(msg.String()) == 1 {
				char := msg.String()[0]
				if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
					(char >= '0' && char <= '9') || char == '_' || char == '-' || char == ' ' {
					m.projectNameInput += msg.String()
					m.err = nil
				}
			}
		}
	}
	return m, nil
}

func (m InitTUIModel) View() string {
	title := ui.TitleStyle.Render("🆕  Create New Project")
	var s string
	s += "\n" + title + "\n\n"
	s += ui.MutedStyle.Render("Template: ") + ui.ValueStyle.Render(m.selectedStrategy.DisplayName) + "\n"
	s += ui.MutedStyle.Render(m.selectedStrategy.Description) + "\n\n"
	s += ui.MutedStyle.Render("Keys stay in Settings (~/.wisp/connectors.yml) — not in the project.") + "\n\n"
	s += ui.LabelStyle.Width(0).Render("Project name: ") + ui.InputStyle.Render(m.projectNameInput+"_") + "\n\n"
	if m.err != nil {
		s += ui.ErrorBoxStyle.Width(0).Render("✗ "+m.err.Error()) + "\n\n"
	}
	s += ui.MutedStyle.Render("↵ Create   q/Esc Cancel")
	return ui.MenuBoxStyle.Width(72).Render(s)
}

// RunInitTUI runs init and returns (strategyFolder, projectName).
func RunInitTUI() (strategy string, projectName string, err error) {
	tpl := starterTemplate()
	m := NewInitTUIModel(tpl)
	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return "", "", err
	}

	result := finalModel.(InitTUIModel)
	if result.cancelled || !result.done || result.projectName == "" {
		return "", "", fmt.Errorf("initialization cancelled")
	}
	return result.selectedStrategy.SDKExample, result.projectName, nil
}

// LoadStrategies returns the supported starter template (tests / callers).
func LoadStrategies() ([]StrategyTemplate, error) {
	return []StrategyTemplate{starterTemplate()}, nil
}

func starterTemplate() StrategyTemplate {
	meta := services.StarterTemplate()
	return StrategyTemplate{
		Name:        meta.Name,
		DisplayName: meta.DisplayName,
		Description: meta.Description,
		Icon:        meta.Icon,
		SDKExample:  meta.SDKExample,
	}
}
