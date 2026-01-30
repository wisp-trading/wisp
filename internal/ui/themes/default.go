package themes

import "github.com/wisp-trading/wisp/internal/ui"

func init() {
	ui.RegisterTheme(&ui.Theme{
		Name:        "default",
		Description: "Default WISP theme - Modern cyan and purple",
		Colors: ui.ColorPalette{
			Primary:    "#00D9FF", // Cyan
			Secondary:  "#7C3AED", // Purple
			Success:    "#10B981", // Green
			Warning:    "#F59E0B", // Orange
			Danger:     "#EF4444", // Red
			Muted:      "#6B7280", // Gray
			Text:       "#D1D5DB", // Light gray
			White:      "#FFFFFF", // White
			BgDark:     "#1F2937", // Dark blue-gray
			BgMedium:   "#374151", // Medium blue-gray
			BgLight:    "#4B5563", // Light blue-gray
			BgSelected: "#1E293B", // Selected dark blue
		},
	})
}
