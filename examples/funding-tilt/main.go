// funding-tilt: Hyperliquid perp funding + RSI fade strategy.
//
//	cd examples/funding-tilt
//	go test .          # pure decision logic
//	go run .           # needs ~/.wisp/connectors.yml (Settings → hyperliquid)
//	FUNDING_TILT_LIVE=1 go run .   # actually place orders (still dry_run param careful)
//
// Default is dry_run (paper): watches HL funding/RSI and logs signals only.
package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/wisp-trading/connectors/pkg/connectors"
	"github.com/wisp-trading/sdk/pkg/types/runtime"
	"github.com/wisp-trading/sdk/pkg/types/strategy"
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
		connectors.Module,
		wisp.Module,
		fx.Provide(NewFundingTilt),
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

	log.Println("funding-tilt running — dry_run unless FUNDING_TILT_LIVE=1; Ctrl+C or Monitor → Stop")
	if err := rt.Wait(); err != nil {
		log.Printf("Wait: %v", err)
		os.Exit(1)
	}
}
