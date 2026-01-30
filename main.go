package main

import (
	"context"
	"log"

	"github.com/wisp-trading/wisp/cmd"
	"github.com/wisp-trading/wisp/internal/app"
	"github.com/wisp-trading/wisp/internal/ui"
	_ "github.com/wisp-trading/wisp/internal/ui/themes" // Register themes
	"go.uber.org/fx"
)

func main() {
	// Initialize theme system FIRST, before fx starts
	initializeTheme()

	fxApp := fx.New(
		app.Module,
		cmd.Module, // Use the command module
		fx.Invoke(runCLI),
		//fx.NopLogger, // Suppress fx startup logs
	)

	fxApp.Run()
}

func initializeTheme() {
	// Load user's theme preference from ~/.wisp/preferences.yml
	// This happens once at startup, before any handlers run
	if err := ui.LoadThemeFromPreferences(); err != nil {
		// Silently fall back to default theme if preferences can't be loaded
		_ = ui.SetTheme("default")
	}
}

func runCLI(lc fx.Lifecycle, shutdowner fx.Shutdowner, root *cmd.RootCommand) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Run the CLI in a goroutine so fx can manage lifecycle
			go func() {
				if err := root.Cmd.Execute(); err != nil {
					log.Printf("Error executing command: %v\n", err)
					log.Fatal(err)
				}
				// Shut down fx app after CLI command completes
				if err := shutdowner.Shutdown(); err != nil {
					log.Printf("Error shutting down: %v\n", err)
				}
			}()
			return nil
		},
	})
}
