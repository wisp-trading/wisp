package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wisp-trading/wisp/internal/ui"
)

type ThemeCommand struct {
	Cmd *cobra.Command
}

func NewThemeCommand() *ThemeCommand {
	cmd := &cobra.Command{
		Use:   "theme [name]",
		Short: "Manage UI themes",
		Long: `List available themes or set the active theme.

Examples:
  wisp theme               # List all available themes
  wisp theme default       # Set theme to default
  wisp theme orange
  wisp theme golden     `,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				listThemes()
			} else {
				setTheme(args[0])
			}
		},
	}

	return &ThemeCommand{Cmd: cmd}
}

func listThemes() {
	themes := ui.GetAvailableThemes()
	current := ui.GetCurrentTheme()
	currentName := ""
	if current != nil {
		currentName = current.Name
	}

	fmt.Println("\n" + ui.TitleStyle.Render("Available Themes:") + "\n")

	for _, name := range themes {
		theme := ui.GetTheme(name)
		if theme == nil {
			continue
		}

		marker := "  "
		if name == currentName {
			marker = ui.SelectedItemStyle.Render("▶ ")
		}

		themeName := ui.CommandStyle.Render(name)
		description := ui.DescriptionStyle.Render(theme.Description)

		fmt.Printf("%s%s - %s\n", marker, themeName, description)
	}

	fmt.Println()
}

func setTheme(name string) {
	if err := ui.SetTheme(name); err != nil {
		errorMsg := ui.ErrorBoxStyle.Render(fmt.Sprintf("❌ %v", err))
		fmt.Println("\n" + errorMsg + "\n")

		fmt.Println("Available themes:")
		for _, themeName := range ui.GetAvailableThemes() {
			fmt.Printf("  - %s\n", themeName)
		}
		fmt.Println()
		return
	}

	// Save theme preference
	prefs := &ui.Preferences{Theme: name}
	if err := ui.SavePreferences(prefs); err != nil {
		fmt.Println(ui.MutedStyle.Render(fmt.Sprintf("⚠️  Warning: Could not save preference: %v", err)))
	}

	successMsg := ui.StatusReadyStyle.Render(fmt.Sprintf("✓ Theme set to '%s'", name))
	fmt.Println("\n" + successMsg)

	theme := ui.GetTheme(name)
	if theme != nil {
		fmt.Println(ui.SubtitleStyle.Render(theme.Description))
	}

	fmt.Println(ui.MutedStyle.Render("Theme preference saved to ~/.wisp/preferences.yml"))
	fmt.Println()
}
