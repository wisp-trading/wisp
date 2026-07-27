package cmd

import (
	"github.com/spf13/cobra"
	backtesting "github.com/wisp-trading/wisp/internal/handlers/strategies/backtest/types"
	"go.uber.org/fx"
)

type BacktestCommandResult struct {
	fx.Out
	BacktestCommand *cobra.Command `name:"backtest"`
}

// NewBacktestCommand creates the backtest command
func NewBacktestCommand(handler backtesting.BacktestHandler) BacktestCommandResult {
	cmd := &cobra.Command{
		Use:   "backtest",
		Short: "Unavailable — use TUI Start Live for strategies",
		Long: `Backtest simulation is not shipped in this CLI build.

Use the supported loop instead:
  wisp init my-bot
  wisp            # Settings → keys · Strategies → Start Live · Monitor`,
		RunE: handler.Handle,
	}

	cmd.Flags().String("config", "", "Reserved (backtest not available)")

	return BacktestCommandResult{
		BacktestCommand: cmd,
	}
}
