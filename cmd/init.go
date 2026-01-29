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
			Short: "Create a new Wisp project",
			RunE:  handler.Handle,
		},
	}
}
