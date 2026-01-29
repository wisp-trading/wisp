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
		Short: "Deploy strategies to live trading",
		RunE:  handler.Handle,
	}

	cmd.Flags().String("strategy", "", "Strategy name for non-interactive mode")
	cmd.Flags().String("exchange", "", "Exchange for non-interactive mode")

	return LiveCommandResult{
		LiveCommand: cmd,
	}
}
