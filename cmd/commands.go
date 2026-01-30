package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// Commands aggregates all cobra commands
type Commands struct {
	Init     *cobra.Command
	Backtest *cobra.Command
	Analyze  *cobra.Command
	Version  *cobra.Command
	Theme    *cobra.Command
}

// CommandParams uses fx.In to inject named commands
type CommandParams struct {
	fx.In
	Init     *cobra.Command `name:"init"`
	Backtest *cobra.Command `name:"backtest"`
	Analyze  *cobra.Command `name:"analyze"`
	Version  *cobra.Command `name:"version"`
	Theme    *cobra.Command `name:"theme"`
}

// NewCommands assembles all commands (created by individual providers)
func NewCommands(params CommandParams) *Commands {
	return &Commands{
		Init:     params.Init,
		Backtest: params.Backtest,
		Analyze:  params.Analyze,
		Version:  params.Version,
		Theme:    params.Theme,
	}
}
