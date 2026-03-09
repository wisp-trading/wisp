package tabs

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/wisp-trading/wisp/internal/ui"
)

// Shared styles for tab views
var (
	tableHeaderStyle = lipgloss.NewStyle().
				Foreground(ui.ColorMuted).
				Bold(true)

	profitStyle  = ui.PnLProfitStyle
	lossStyle    = ui.PnLLossStyle
	neutralStyle = ui.PnLNeutralStyle

	// marketDataSubTabStyle is used for inactive sub-tabs within the Market Data tab
	marketDataSubTabStyle = lipgloss.NewStyle().
				Foreground(ui.ColorMuted).
				Padding(0, 1)

	// marketDataSubTabActiveStyle is used for the active sub-tab within the Market Data tab
	marketDataSubTabActiveStyle = lipgloss.NewStyle().
					Foreground(ui.ColorSecondary).
					Bold(true).
					Padding(0, 1).
					Underline(true)
)

// FormatPnL formats a PnL value with appropriate styling
// This is a convenience wrapper around ui.FormatPnL for backward compatibility
var FormatPnL = ui.FormatPnL
