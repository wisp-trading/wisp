package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Theme represents a complete color palette for the TUI
type Theme struct {
	Name        string
	Description string
	Colors      ColorPalette
}

// ColorPalette defines all semantic colors used in the UI
type ColorPalette struct {
	// Primary colors
	Primary   string // Main brand/accent color
	Secondary string // Secondary accents
	Success   string // Success states (green)
	Warning   string // Warning states (orange)
	Danger    string // Error/danger states (red)
	Muted     string // De-emphasized text
	Text      string // Normal text
	White     string // High contrast text

	// Background colors
	BgDark     string // Darkest background
	BgMedium   string // Medium background
	BgLight    string // Light background
	BgSelected string // Selected item background
}

var (
	// currentTheme holds the active theme
	currentTheme *Theme

	// availableThemes maps theme names to Theme objects
	availableThemes = make(map[string]*Theme)
)

// RegisterTheme adds a theme to the available themes
func RegisterTheme(theme *Theme) {
	availableThemes[theme.Name] = theme
}

// GetAvailableThemes returns list of theme names
func GetAvailableThemes() []string {
	themes := make([]string, 0, len(availableThemes))
	for name := range availableThemes {
		themes = append(themes, name)
	}
	return themes
}

// GetTheme returns a theme by name
func GetTheme(name string) *Theme {
	return availableThemes[name]
}

// SetTheme switches to the specified theme and rebuilds all styles
func SetTheme(name string) error {
	theme, exists := availableThemes[name]
	if !exists {
		return fmt.Errorf("theme '%s' not found", name)
	}

	currentTheme = theme
	applyTheme()
	return nil
}

// GetCurrentTheme returns the active theme
func GetCurrentTheme() *Theme {
	if currentTheme == nil {
		// Default to "default" theme if none set
		if defaultTheme, exists := availableThemes["default"]; exists {
			currentTheme = defaultTheme
		}
	}
	return currentTheme
}

// applyTheme updates all color variables and rebuilds styles
func applyTheme() {
	palette := currentTheme.Colors

	// Update color variables
	ColorPrimary = lipgloss.Color(palette.Primary)
	ColorSecondary = lipgloss.Color(palette.Secondary)
	ColorSuccess = lipgloss.Color(palette.Success)
	ColorWarning = lipgloss.Color(palette.Warning)
	ColorDanger = lipgloss.Color(palette.Danger)
	ColorMuted = lipgloss.Color(palette.Muted)
	ColorText = lipgloss.Color(palette.Text)
	ColorWhite = lipgloss.Color(palette.White)
	ColorBgDark = lipgloss.Color(palette.BgDark)
	ColorBgMedium = lipgloss.Color(palette.BgMedium)
	ColorBgLight = lipgloss.Color(palette.BgLight)
	ColorBgSelected = lipgloss.Color(palette.BgSelected)

	// Rebuild all styles with new colors
	rebuildStyles()
}

// rebuildStyles recreates all style definitions with current color palette
func rebuildStyles() {
	// ===== Title Styles =====

	TitleStyle = lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Bold(true).
		Padding(1, 2).
		MarginBottom(1)

	TitleCenteredStyle = lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Bold(true).
		PaddingTop(1).
		PaddingBottom(1).
		Align(lipgloss.Center)

	SectionHeaderStyle = lipgloss.NewStyle().
		Foreground(ColorSecondary).
		Bold(true).
		PaddingTop(1)

	SubtitleStyle = lipgloss.NewStyle().
		Foreground(ColorMuted).
		Italic(true)

	// ===== Box Styles =====

	BoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Padding(1, 2).
		Width(70)

	MenuBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorSecondary).
		Padding(2, 4).
		Width(50)

	DetailBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Padding(1, 2).
		Width(68)

	ErrorBoxStyle = lipgloss.NewStyle().
		Foreground(ColorDanger).
		Bold(true).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorDanger).
		Padding(1, 2)

	ConfirmBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(ColorDanger).
		Padding(2, 4).
		Width(70)

	// ===== Item/List Styles =====

	ItemStyle = lipgloss.NewStyle().
		PaddingLeft(2)

	SelectedItemStyle = lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Bold(true).
		PaddingLeft(0)

	StrategyItemStyle = lipgloss.NewStyle().
		Padding(1, 2).
		MarginBottom(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBgLight).
		Width(70)

	StrategyItemSelectedStyle = lipgloss.NewStyle().
		Padding(1, 2).
		MarginBottom(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Background(ColorBgSelected).
		Width(70)

	// ===== Text Styles =====

	StrategyNameStyle = lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Bold(true)

	StrategyNameSelectedStyle = lipgloss.NewStyle().
		Foreground(ColorSuccess).
		Bold(true)

	StrategyDescStyle = lipgloss.NewStyle().
		Foreground(ColorMuted).
		Italic(true)

	StrategyMetaStyle = lipgloss.NewStyle().
		Foreground(ColorSecondary).
		MarginTop(1)

	LabelStyle = lipgloss.NewStyle().
		Foreground(ColorMuted).
		Width(15)

	ValueStyle = lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Bold(true)

	DescriptionStyle = lipgloss.NewStyle().
		Foreground(ColorMuted).
		PaddingLeft(4).
		Width(60)

	CommandStyle = lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Bold(true)

	TextStyle = lipgloss.NewStyle().
		Foreground(ColorText)

	MutedStyle = lipgloss.NewStyle().
		Foreground(ColorMuted).
		Italic(true)

	// ===== Status Styles =====

	StatusReadyStyle = lipgloss.NewStyle().
		Foreground(ColorSuccess).
		Bold(true)

	StatusRunningStyle = lipgloss.NewStyle().
		Foreground(ColorWarning).
		Bold(true)

	StatusDangerStyle = lipgloss.NewStyle().
		Foreground(ColorDanger).
		Bold(true)

	StatusDisabledStyle = lipgloss.NewStyle().
		Foreground(ColorMuted).
		Bold(true)

	StatusErrorStyle = lipgloss.NewStyle().
		Foreground(ColorDanger).
		MarginTop(1)

	// ===== Badge Styles =====

	NetworkBadgeStyle = lipgloss.NewStyle().
		Foreground(ColorSecondary)

	NetworkBadgeWarningStyle = lipgloss.NewStyle().
		Foreground(ColorWarning)

	// ===== Input Styles =====

	InputStyle = lipgloss.NewStyle().
		Foreground(ColorSuccess).
		Bold(true)

	// ===== Help/Navigation Styles =====

	HelpStyle = lipgloss.NewStyle().
		Foreground(ColorMuted).
		Padding(1, 2).
		MarginTop(1)

	KeyHintStyle = lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Bold(true)

	// ===== Confirmation Dialog Styles =====

	ConfirmTitleStyle = lipgloss.NewStyle().
		Foreground(ColorDanger).
		Bold(true).
		Align(lipgloss.Center)

	ConfirmFieldStyle = lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Bold(true)

	ConfirmValueStyle = lipgloss.NewStyle().
		Foreground(ColorWhite)

	ConfirmWarningStyle = lipgloss.NewStyle().
		Foreground(ColorWarning).
		Bold(true).
		Align(lipgloss.Center).
		MarginTop(1).
		MarginBottom(1)

	// ===== PnL Styles =====

	PnLProfitStyle = lipgloss.NewStyle().
		Foreground(ColorSuccess).
		Bold(true)

	PnLLossStyle = lipgloss.NewStyle().
		Foreground(ColorDanger).
		Bold(true)

	PnLNeutralStyle = lipgloss.NewStyle().
		Foreground(ColorMuted)
}
