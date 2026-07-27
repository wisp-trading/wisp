package cmd

import (
	"fmt"
	"runtime"
	"runtime/debug"

	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// Version is set at link time for release builds:
//
//	go build -ldflags "-X github.com/wisp-trading/wisp/cmd.Version=v0.2.0"
//
// Falls back to module build info, then "dev".
var Version = "dev"

type VersionCommandResult struct {
	fx.Out
	VersionCommand *cobra.Command `name:"version"`
}

// NewVersionCommand creates the version command
func NewVersionCommand() VersionCommandResult {
	return VersionCommandResult{
		VersionCommand: &cobra.Command{
			Use:   "version",
			Short: "Show version information",
			Run: func(c *cobra.Command, _ []string) {
				c.Println(formatVersion())
			},
		},
	}
}

func formatVersion() string {
	v := Version
	if v == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
			v = info.Main.Version
		}
	}
	return fmt.Sprintf("wisp %s (%s/%s)", v, runtime.GOOS, runtime.GOARCH)
}
