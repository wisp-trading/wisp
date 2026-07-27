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
		Short: "Wisp — project · exchange keys · live strategies",
		Long: `Wisp CLI — local tooling for standalone trading strategies.

Core loop (TUI is the default):
  1. Create a project     wisp init my-bot
  2. Add exchange keys    wisp → Settings  (~/.wisp/connectors.yml)
  3. Start / stop live    Strategies → Start Live · Monitor → Stop

Strategies are standalone binaries (main.go + StartStandalone + Wait).
Project config.yml lists exchanges/assets only — never secrets.

Examples:
  wisp                      Interactive menu
  wisp init my-bot          Scaffold project with strategies/starter
  wisp version
  wisp --cli                Cobra help (subcommand mode)
  wisp live --cli --strategy starter`,
		RunE: handler.Handle,
	}

	cmd.PersistentFlags().Bool("cli", false, "Use CLI mode instead of interactive TUI")

	return &RootCommand{Cmd: cmd}
}
