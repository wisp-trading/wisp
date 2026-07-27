package cmd

import (
	"github.com/spf13/cobra"
	setup "github.com/wisp-trading/wisp/internal/setup/types"
	"go.uber.org/fx"
)

type InitCommandResult struct {
	fx.Out
	InitCommand *cobra.Command `name:"init"`
}

// NewInitCommand creates the init command
func NewInitCommand(handler setup.InitHandler) InitCommandResult {
	return InitCommandResult{
		InitCommand: &cobra.Command{
			Use:   "init <name>",
			Short: "Scaffold a project (strategies/<name> standalone starter)",
			Long: `Create a new project directory with strategies/starter/:
  main.go     StartStandalone + Wait
  config.yml  exchanges/assets only (no secrets)
  go.mod

Keys go in ~/.wisp/connectors.yml via: wisp → Settings

Example:
  wisp init my-bot
  cd my-bot/strategies/starter && go mod tidy && go run .`,
			RunE: handler.Handle,
		},
	}
}
