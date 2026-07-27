package cmd

import (
	"github.com/spf13/cobra"
	backtesting "github.com/wisp-trading/wisp/internal/handlers/strategies/backtest/types"
	"go.uber.org/fx"
)

type AnalyzeCommandResult struct {
	fx.Out
	AnalyzeCommand *cobra.Command `name:"analyze"`
}

// NewAnalyzeCommand creates the analyze command
func NewAnalyzeCommand(handler backtesting.AnalyzeHandler) AnalyzeCommandResult {
	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Unavailable — no backtest results pipeline yet",
		Long: `Result analysis depends on backtest, which is not available in this build.

Use live monitoring instead: wisp → Monitor`,
		RunE: handler.Handle,
	}

	cmd.Flags().String("path", "./results", "Reserved (analyze not available)")

	return AnalyzeCommandResult{
		AnalyzeCommand: cmd,
	}
}
