package monitor

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/wisp-trading/wisp/internal/ui"
)

// Health bar characters
const (
	HealthFilled = "●"
	HealthEmpty  = "○"
)

// Status icons
const (
	IconRunning = "🟢"
	IconWarning = "🟡"
	IconStopped = "🔴"
	IconUnknown = "⚪"
)

// Monitor-specific styles (extending base UI styles)
var (
	// Panel style for boxed content
	PanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ui.ColorPrimary).
			Padding(1, 2)

	// Tab styles
	TabStyle = lipgloss.NewStyle().
			Foreground(ui.ColorMuted).
			Padding(0, 2)

	TabActiveStyle = lipgloss.NewStyle().
			Foreground(ui.ColorPrimary).
			Bold(true).
			Padding(0, 2).
			Underline(true)

	// Table row styles
	TableHeaderStyle = lipgloss.NewStyle().
				Foreground(ui.ColorMuted).
				Bold(true)

	TableRowStyle = lipgloss.NewStyle().
			Padding(0, 1)

	TableRowSelectedStyle = lipgloss.NewStyle().
				Padding(0, 1).
				Background(ui.ColorBgSelected).
				Foreground(ui.ColorPrimary)

	// PnL styles
	PnLProfitStyle = lipgloss.NewStyle().
			Foreground(ui.ColorSuccess).
			Bold(true)

	PnLLossStyle = lipgloss.NewStyle().
			Foreground(ui.ColorDanger).
			Bold(true)

	PnLNeutralStyle = lipgloss.NewStyle().
			Foreground(ui.ColorMuted)
)

// FormatPnL formats a PnL value with appropriate styling
func FormatPnL(value float64) string {
	if value > 0 {
		return PnLProfitStyle.Render(fmt.Sprintf("+$%.2f", value))
	} else if value < 0 {
		return PnLLossStyle.Render(fmt.Sprintf("-$%.2f", -value))
	}
	return PnLNeutralStyle.Render("$0.00")
}

// FormatHealthBar renders a health bar (0-5)
func FormatHealthBar(level int) string {
	bar := ""
	for i := 0; i < 5; i++ {
		if i < level {
			bar += lipgloss.NewStyle().Foreground(ui.ColorSuccess).Render(HealthFilled)
		} else {
			bar += lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(HealthEmpty)
		}
	}
	return bar
}

// GetStatusIcon returns the appropriate icon for a status
func GetStatusIcon(status string) string {
	switch status {
	case "running":
		return IconRunning
	case "warning":
		return IconWarning
	case "stopped":
		return IconStopped
	default:
		return IconUnknown
	}
}

// GetStatusStyle returns the appropriate style for a status
func GetStatusStyle(status string) lipgloss.Style {
	switch status {
	case "running":
		return ui.StatusReadyStyle
	case "warning":
		return ui.StatusRunningStyle
	case "stopped":
		return ui.StatusDangerStyle
	default:
		return lipgloss.NewStyle().Foreground(ui.ColorMuted)
	}
}
