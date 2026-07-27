package browse

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/donderom/bubblon"
	"github.com/wisp-trading/sdk/pkg/types/config"
	"github.com/wisp-trading/wisp/internal/handlers/strategies/compile"
	"github.com/wisp-trading/wisp/internal/handlers/strategies/live"
	"github.com/wisp-trading/wisp/internal/ui"
)

type ActionType int

const (
	ActionCompile ActionType = iota
	ActionStartTrading
)

var actionNames = map[ActionType]string{
	ActionStartTrading: "Start Live",
	ActionCompile:      "Compile",
}

type StrategyDetailView interface {
	tea.Model
}

// strategyDetailView represents the strategy detail view with action options (STRATEGY screen)
type strategyDetailView struct {
	strategy       *config.Strategy
	actions        []ActionType
	cursor         int
	compileFactory compile.CompileViewFactory
	liveFactory    live.LiveViewFactory
}

// newStrategyDetailView is the private constructor called by the factory
func newStrategyDetailView(
	compileFactory compile.CompileViewFactory,
	liveFactory live.LiveViewFactory,
	s *config.Strategy,
) tea.Model {
	return &strategyDetailView{
		strategy:       s,
		actions:        []ActionType{ActionCompile, ActionStartTrading},
		cursor:         0,
		compileFactory: compileFactory,
		liveFactory:    liveFactory,
	}
}

func (m *strategyDetailView) Init() tea.Cmd {
	return nil
}

func (m *strategyDetailView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q", "esc", "backspace":
			return m, bubblon.Cmd(bubblon.Close())
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.actions)-1 {
				m.cursor++
			}
		case "enter", " ":
			if m.strategy != nil && m.strategy.Error != "" {
				return m, nil // fix config first
			}
			action := m.actions[m.cursor]
			switch action {
			case ActionCompile:
				compileView := m.compileFactory(m.strategy)
				return m, bubblon.Open(compileView)
			case ActionStartTrading:
				liveView := m.liveFactory(m.strategy)
				return m, bubblon.Open(liveView)
			}
		}
	}
	return m, nil
}

func (m *strategyDetailView) View() string {
	if m.strategy == nil {
		return ui.BoxStyle.Render("Strategy not found")
	}

	var content string
	content += ui.TitleStyle.Render(m.strategy.Name) + "\n"
	if m.strategy.Error != "" {
		content += ui.ErrorBoxStyle.Render("config.yml: "+m.strategy.Error) + "\n\n"
		content += ui.MutedStyle.Render("Fix strategies/"+m.strategy.Name+"/config.yml then r refresh.") + "\n"
		content += ui.MutedStyle.Render("q/Esc back")
		return ui.BoxStyle.Render(content)
	}
	if len(m.strategy.Exchanges) > 0 {
		content += ui.MutedStyle.Render("exchanges: ") +
			ui.ValueStyle.Render(fmt.Sprintf("%v", m.strategy.Exchanges)) + "\n"
	}
	content += ui.MutedStyle.Render("Keys: Settings → ~/.wisp/connectors.yml") + "\n"
	content += ui.MutedStyle.Render("Domain = connector MarketType (not a YAML field)") + "\n\n"
	content += ui.SubtitleStyle.Render("Select action:") + "\n\n"

	for i, action := range m.actions {
		actionName := actionNames[action]
		if i == m.cursor {
			content += ui.StrategyNameSelectedStyle.Render("▶ "+actionName) + "\n"
		} else {
			content += "  " + actionName + "\n"
		}
	}

	content += "\n" + ui.MutedStyle.Render("↑↓  ↵ select   q/Esc back")

	return ui.BoxStyle.Render(content)
}
