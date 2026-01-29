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
		Short: "Run backtests",
		RunE:  handler.Handle,
	}

	cmd.Flags().String("config", "", "Path to backtest config file (for CLI mode)")

	return BacktestCommandResult{
		BacktestCommand: cmd,
	}
}
