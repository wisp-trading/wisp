package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Color palette - centralized color definitions for the entire TUI.
// These colors define the visual identity of the WISP CLI.
var (
	// Primary colors
	ColorPrimary   = lipgloss.Color("#00D9FF") // Cyan - primary brand color
	ColorSecondary = lipgloss.Color("#7C3AED") // Purple - secondary accents
	ColorSuccess   = lipgloss.Color("#10B981") // Green - success states
	ColorWarning   = lipgloss.Color("#F59E0B") // Orange - warning states
	ColorDanger    = lipgloss.Color("#EF4444") // Red - danger/error states
	ColorMuted     = lipgloss.Color("#6B7280") // Gray - muted text
	ColorText      = lipgloss.Color("#D1D5DB") // Light gray - normal text
	ColorWhite     = lipgloss.Color("#FFFFFF") // White - high contrast text

	// Background colors
	ColorBgDark     = lipgloss.Color("#1F2937")
	ColorBgMedium   = lipgloss.Color("#374151")
	ColorBgLight    = lipgloss.Color("#4B5563")
	ColorBgSelected = lipgloss.Color("#1E293B")
)

// Style definitions - reusable lipgloss styles for consistent UI components.
// All handlers should use these styles instead of creating inline styles.
var (
	// ===== Title Styles =====

	// TitleStyle is used for main view titles (e.g., "STRATEGIES", "SETTINGS")
	TitleStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true).
			Padding(1, 2).
			MarginBottom(1)

	// TitleCenteredStyle is used for centered titles in dialogs
	TitleCenteredStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true).
				PaddingTop(1).
				PaddingBottom(1).
				Align(lipgloss.Center)

	// SectionHeaderStyle is used for section headers within a view
	SectionHeaderStyle = lipgloss.NewStyle().
				Foreground(ColorSecondary).
				Bold(true).
				PaddingTop(1)

	// SubtitleStyle is used for secondary text below titles
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Italic(true)

	// ===== Box Styles =====

	// BoxStyle is the standard container box with primary border
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(1, 2).
			Width(70)

	// MenuBoxStyle is used for main menu and selection dialogs
	MenuBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSecondary).
			Padding(2, 4).
			Width(50)

	// DetailBoxStyle is used for detailed information displays
	DetailBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(1, 2).
			Width(68)

	// ErrorBoxStyle is used for error messages
	ErrorBoxStyle = lipgloss.NewStyle().
			Foreground(ColorDanger).
			Bold(true).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorDanger).
			Padding(1, 2)

	// ConfirmBoxStyle is used for dangerous confirmation dialogs
	ConfirmBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(ColorDanger).
			Padding(2, 4).
			Width(70)

	// ===== Item/List Styles =====

	// ItemStyle is used for unselected list items
	ItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	// SelectedItemStyle is used for selected list items with cursor
	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true).
				PaddingLeft(0)

	// StrategyItemStyle is used for strategy list items (legacy, consider using ItemStyle)
	StrategyItemStyle = lipgloss.NewStyle().
				Padding(1, 2).
				MarginBottom(1).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorBgLight).
				Width(70)

	// StrategyItemSelectedStyle is used for selected strategy items
	StrategyItemSelectedStyle = lipgloss.NewStyle().
					Padding(1, 2).
					MarginBottom(1).
					Border(lipgloss.RoundedBorder()).
					BorderForeground(ColorPrimary).
					Background(ColorBgSelected).
					Width(70)

	// ===== Text Styles =====

	// StrategyNameStyle is used for strategy names
	StrategyNameStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true)

	// StrategyNameSelectedStyle is used for selected strategy names
	StrategyNameSelectedStyle = lipgloss.NewStyle().
					Foreground(ColorSuccess).
					Bold(true)

	// StrategyDescStyle is used for strategy descriptions
	StrategyDescStyle = lipgloss.NewStyle().
				Foreground(ColorMuted).
				Italic(true)

	// StrategyMetaStyle is used for strategy metadata
	StrategyMetaStyle = lipgloss.NewStyle().
				Foreground(ColorSecondary).
				MarginTop(1)

	// LabelStyle is used for field labels in forms and details
	LabelStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Width(15)

	// ValueStyle is used for field values
	ValueStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	// DescriptionStyle is used for descriptive text
	DescriptionStyle = lipgloss.NewStyle().
				Foreground(ColorMuted).
				PaddingLeft(4).
				Width(60)

	// CommandStyle is used for command names in help
	CommandStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	// TextStyle is used for regular text content
	TextStyle = lipgloss.NewStyle().
			Foreground(ColorText)

	// MutedStyle is used for de-emphasized text
	MutedStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Italic(true)

	// ===== Status Styles =====

	// StatusReadyStyle indicates a ready/enabled/success state
	StatusReadyStyle = lipgloss.NewStyle().
				Foreground(ColorSuccess).
				Bold(true)

	// StatusRunningStyle indicates a running/active/warning state
	StatusRunningStyle = lipgloss.NewStyle().
				Foreground(ColorWarning).
				Bold(true)

	// StatusDangerStyle indicates a danger/critical state
	StatusDangerStyle = lipgloss.NewStyle().
				Foreground(ColorDanger).
				Bold(true)

	// StatusDisabledStyle indicates a disabled/inactive state
	StatusDisabledStyle = lipgloss.NewStyle().
				Foreground(ColorMuted).
				Bold(true)

	// StatusErrorStyle is used for error messages
	StatusErrorStyle = lipgloss.NewStyle().
				Foreground(ColorDanger).
				MarginTop(1)

	// ===== Badge Styles =====

	// NetworkBadgeStyle is used for network indicators (mainnet/testnet)
	NetworkBadgeStyle = lipgloss.NewStyle().
				Foreground(ColorSecondary)

	// NetworkBadgeWarningStyle is used for testnet badges
	NetworkBadgeWarningStyle = lipgloss.NewStyle().
					Foreground(ColorWarning)

	// ===== Input Styles =====

	// InputStyle is used for input fields and user text entry
	InputStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	// ===== Help/Navigation Styles =====

	// HelpStyle is used for help text at bottom of views
	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Padding(1, 2).
			MarginTop(1)

	// KeyHintStyle is used for keyboard shortcut hints
	KeyHintStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	// ===== Confirmation Dialog Styles =====

	// ConfirmTitleStyle is used for confirmation dialog titles
	ConfirmTitleStyle = lipgloss.NewStyle().
				Foreground(ColorDanger).
				Bold(true).
				Align(lipgloss.Center)

	// ConfirmFieldStyle is used for field names in confirmations
	ConfirmFieldStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true)

	// ConfirmValueStyle is used for values in confirmations
	ConfirmValueStyle = lipgloss.NewStyle().
				Foreground(ColorWhite)

	// ConfirmWarningStyle is used for warning text in confirmations
	ConfirmWarningStyle = lipgloss.NewStyle().
				Foreground(ColorWarning).
				Bold(true).
				Align(lipgloss.Center).
				MarginTop(1).
				MarginBottom(1)

	// ===== PnL Styles =====

	// PnLProfitStyle is used for positive profit values
	PnLProfitStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	// PnLLossStyle is used for negative loss values
	PnLLossStyle = lipgloss.NewStyle().
			Foreground(ColorDanger).
			Bold(true)

	// PnLNeutralStyle is used for zero PnL values
	PnLNeutralStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)
)

// FormatPnL formats a profit/loss value with appropriate styling.
// Positive values are green, negative are red, zero is muted.
func FormatPnL(value float64) string {
	if value > 0 {
		return PnLProfitStyle.Render(fmt.Sprintf("+$%.2f", value))
	} else if value < 0 {
		return PnLLossStyle.Render(fmt.Sprintf("-$%.2f", -value))
	}
	return PnLNeutralStyle.Render("$0.00")
}

// RenderProgressBar creates a styled progress bar with the given percentage and width.
// Used for compilation progress, loading indicators, etc.
func RenderProgressBar(percent float64, width int) string {
	filled := int(percent * float64(width))
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("█", filled)
	empty := strings.Repeat("░", width-filled)

	percentStr := fmt.Sprintf("%.0f%%", percent*100)

	progressStyle := lipgloss.NewStyle().
		Foreground(ColorSuccess).
		Bold(true)

	emptyStyle := lipgloss.NewStyle().
		Foreground(ColorMuted)

	return progressStyle.Render(bar) + emptyStyle.Render(empty) + " " + percentStr
}
