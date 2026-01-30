package themes

import "github.com/wisp-trading/wisp/internal/ui"

func init() {
	ui.RegisterTheme(&ui.Theme{
		Name:        "orange",
		Description: "Orange theme - medieval MMORPG UI",
		Colors: ui.ColorPalette{
			Primary:    "#FF9900", // Orange - signature RS color
			Secondary:  "#FFCC33", // Gold yellow - hover/selected
			Success:    "#00FF00", // Bright green - XP drops
			Warning:    "#FFFF00", // Yellow - warnings
			Danger:     "#FF0000", // Bright red - danger/death
			Muted:      "#666666", // Gray - disabled items
			Text:       "#FFFFE0", // Light yellow - readable text
			White:      "#FFFFFF", // White - important text
			BgDark:     "#2B1E0F", // Dark brown - UI background
			BgMedium:   "#3D2A18", // Medium brown
			BgLight:    "#4F3621", // Light brown
			BgSelected: "#5C4A2E", // Selected orange-brown
		},
	})
}
