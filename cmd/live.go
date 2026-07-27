package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wisp-trading/wisp/internal/handlers/strategies/live"
	"go.uber.org/fx"
)

type LiveCommandResult struct {
	fx.Out
	LiveCommand *cobra.Command `name:"live"`
}

// NewLiveCommand creates the live command
func NewLiveCommand(handler live.LiveHandler) LiveCommandResult {
	cmd := &cobra.Command{
		Use:   "live",
		Short: "Start a strategy live (compile + spawn standalone binary)",
		Long: `Compile and spawn a strategy as a background process.

Prefer the TUI: wisp → Strategies → Start Live · Monitor → Stop.

CLI mode needs --cli and a strategy name under ./strategies.

Keys: ~/.wisp/connectors.yml (Settings). Strategy config.yml has no secrets.`,
		RunE: handler.Handle,
	}

	cmd.Flags().String("strategy", "", "Strategy folder name under ./strategies (CLI mode)")
	cmd.Flags().String("exchange", "", "Unused legacy flag (exchanges come from strategy config.yml)")

	return LiveCommandResult{
		LiveCommand: cmd,
	}
}
