package themes

import "github.com/wisp-trading/wisp/internal/ui"

func init() {
	ui.RegisterTheme(&ui.Theme{
		Name:        "golden",
		Description: "Golden theme - fantasy RTS interface",
		Colors: ui.ColorPalette{
			Primary:    "#FFD700", // Gold - selected units/primary
			Secondary:  "#8B4513", // Saddle brown - wood UI
			Success:    "#32CD32", // Lime green - HP/mana bars
			Warning:    "#FF8C00", // Dark orange - warnings
			Danger:     "#DC143C", // Crimson - damage/blood
			Muted:      "#8B7355", // Brown gray - disabled
			Text:       "#F5DEB3", // Wheat - readable text
			White:      "#FFFACD", // Lemon chiffon - highlights
			BgDark:     "#1A0F0A", // Very dark brown
			BgMedium:   "#2D1810", // Dark wood
			BgLight:    "#3F2415", // Medium wood
			BgSelected: "#4A2D1A", // Selected brown-gold
		},
	})
}
