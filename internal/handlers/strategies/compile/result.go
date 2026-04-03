package compile

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/donderom/bubblon"
	"github.com/wisp-trading/sdk/pkg/types/config"
	"github.com/wisp-trading/wisp/internal/ui"
)

type resultModel struct {
	strategy *config.Strategy
	err      error
}

// NewResultModel creates a result model that shows compilation result
func NewResultModel(strat *config.Strategy, err error) tea.Model {
	return &resultModel{
		strategy: strat,
		err:      err,
	}
}

func (m *resultModel) Init() tea.Cmd {
	return nil
}

func (m *resultModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "enter":
			// Return to detail view (parent) by closing this result view
			return m, bubblon.Cmd(bubblon.Close())
		}
	}
	return m, nil
}

func (m *resultModel) View() string {
	// Title
	title := ui.TitleStyle.Render("📦 Compilation Result")
	strategyName := ui.StrategyNameStyle.Render(m.strategy.Name)

	var statusSection string
	if m.err == nil {
		// Success
		statusIcon := ui.StatusReadyStyle.Render("✅ SUCCESS")
		message := ui.SubtitleStyle.Render("Strategy has been compiled successfully")
		details := ui.MutedStyle.Italic(false).
			Render("• Plugin binary created\n• Ready for backtest or live trading")

		statusSection = lipgloss.JoinVertical(
			lipgloss.Left,
			statusIcon,
			"",
			message,
			"",
			details,
		)
	} else {
		// Failure
		statusIcon := ui.StatusErrorStyle.Render("❌ FAILED")
		message := ui.SubtitleStyle.Render("Compilation encountered errors")
		errorMsg := ui.StatusErrorStyle.Render(fmt.Sprintf("\nError:\n%v", m.err))

		statusSection = lipgloss.JoinVertical(
			lipgloss.Left,
			statusIcon,
			"",
			message,
			errorMsg,
		)
	}

	help := ui.HelpStyle.Render("Press Enter or q to return")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		strategyName,
		"",
		statusSection,
		"",
		help,
	)

	return ui.BoxStyle.Render(content)
}
