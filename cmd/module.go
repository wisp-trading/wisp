package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// Module provides all command-related dependencies
var Module = fx.Module("commands",
	fx.Provide(
		NewRootCommand,
		NewInitCommand,
		NewBacktestCommand,
		NewAnalyzeCommand,
		NewVersionCommand,
		NewThemeCommand,
		fx.Annotate(
			func(c *ThemeCommand) *cobra.Command { return c.Cmd },
			fx.ResultTags(`name:"theme"`),
		),
		NewCommands,
	),
	fx.Invoke(registerCommands),
)

type registerCommandsParams struct {
	fx.In

	Root  *RootCommand
	Cmds  *Commands
	Theme *ThemeCommand
}

// registerCommands wires up the command tree
func registerCommands(p registerCommandsParams) {
	p.Root.Cmd.AddCommand(p.Cmds.Init)
	p.Root.Cmd.AddCommand(p.Cmds.Backtest)
	p.Root.Cmd.AddCommand(p.Cmds.Analyze)
	p.Root.Cmd.AddCommand(p.Cmds.Version)
	p.Root.Cmd.AddCommand(p.Theme.Cmd)
}
