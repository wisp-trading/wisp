package cmd

import (
	"github.com/spf13/cobra"
	core "github.com/wisp-trading/wisp/internal/handlers"
)

// RootCommand wraps the root cobra command
type RootCommand struct {
	Cmd *cobra.Command
}

// NewRootCommand creates the root command
func NewRootCommand(handler core.RootHandler) *RootCommand {
	cmd := &cobra.Command{
		Use:   "wisp",
		Short: "Wisp - Trading infrastructure platform",
		Long: `Wisp CLI - Backtesting and live trading infrastructure

Use Wisp to:
  • Configure backtests via YAML
  • Run backtests locally with deterministic simulation
  • Deploy strategies live
  • Analyze results

Examples:
  wisp                         Launch interactive TUI menu (default)
  wisp --cli                   Show traditional CLI help
  wisp init my-project         Create a new project (CLI mode)
  wisp init                    Create a new project (TUI mode)
  wisp backtest --cli --config backtest.yaml    Run backtest via CLI
  wisp backtest                Run backtest via TUI
  wisp live --cli --strategy arbitrage --exchange binance    Run live via CLI
  wisp live                    Run live via TUI`,
		RunE: handler.Handle,
	}

	cmd.PersistentFlags().Bool("cli", false, "Use CLI mode instead of interactive TUI")

	return &RootCommand{Cmd: cmd}
}
