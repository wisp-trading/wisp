// funding-tilt: Hyperliquid perp funding + RSI fade strategy.
//
//	cd examples/funding-tilt
//	go test .          # pure decision logic + params
//	go run .           # dry_run (needs ~/.wisp/connectors.yml)
//	FUNDING_TILT_LIVE=1 go run .   # live CLI (dedicated wallet only)
//
// Prefer starting via wisp-control registry (StartInstance) for dual gate:
// dry_run registry injects WISP_DRY_RUN=1 so live env alone cannot arm.
package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/wisp-trading/connectors/pkg/connectors/hyperliquid"
	"github.com/wisp-trading/sdk/pkg/types/runtime"
	"github.com/wisp-trading/sdk/pkg/types/strategy"
	wisptypes "github.com/wisp-trading/sdk/pkg/types/wisp"
	"github.com/wisp-trading/sdk/wisp"
	"go.uber.org/fx"
)

func main() {
	configDir := flag.String("config", ".", "strategy config directory")
	wispPath := flag.String("wisp", "", "connector settings path (default: ~/.wisp/connectors.yml)")
	flag.Parse()

	ctx := context.Background()
	var (
		rt    runtime.Runtime
		strat strategy.Strategy
	)

	app := fx.New(
		// Only the venues this strategy uses — not connectors.Module (all venues).
		hyperliquid.Module,
		wisp.Module,
		fx.Provide(func(k wisptypes.Wisp) strategy.Strategy {
			return NewFundingTilt(k, *configDir)
		}),
		fx.Populate(&rt, &strat),
		fx.NopLogger,
	)

	if err := app.Start(ctx); err != nil {
		log.Fatalf("fx start: %v", err)
	}
	defer func() { _ = app.Stop(ctx) }()

	if err := rt.StartStandalone(strat, *configDir, *wispPath); err != nil {
		log.Fatalf("StartStandalone: %v\n\nHint: wisp → Settings → add hyperliquid keys to ~/.wisp/connectors.yml", err)
	}

	mode := "dry_run"
	if !resolveDryRun() {
		mode = "LIVE"
	}
	log.Printf("funding-tilt running mode=%s — Ctrl+C or Monitor → Stop", mode)
	if err := rt.Wait(); err != nil {
		log.Printf("Wait: %v", err)
		os.Exit(1)
	}
}
