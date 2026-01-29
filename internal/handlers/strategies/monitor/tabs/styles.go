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
)

// FormatPnL formats a PnL value with appropriate styling
// This is a convenience wrapper around ui.FormatPnL for backward compatibility
var FormatPnL = ui.FormatPnL
