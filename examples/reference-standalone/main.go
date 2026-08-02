// Reference process host for Wisp strategies.
//
// Pattern (blessed):
//
//	fx app Start → runtime.StartStandalone → runtime.Wait
//
// Wait unifies OS signals and monitoring HTTP /shutdown so the process always exits.
package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/wisp-trading/connectors/pkg/connectors/hyperliquid"
	"github.com/wisp-trading/sdk/pkg/types/runtime"
	"github.com/wisp-trading/sdk/pkg/types/strategy"
	"github.com/wisp-trading/sdk/wisp"
	"go.uber.org/fx"
)

func main() {
	configDir := flag.String("config", ".", "strategy config directory")
	// Empty --wisp uses ~/.wisp/connectors.yml (or WISP_SETTINGS / project-local migration).
	wispYml := flag.String("wisp", "", "connector settings path (default: ~/.wisp/connectors.yml)")
	flag.Parse()

	ctx := context.Background()

	var (
		rt    runtime.Runtime
		strat strategy.Strategy
	)

	app := fx.New(
		// Compose only venues listed in config.yml (here: hyperliquid).
		hyperliquid.Module,
		wisp.Module,
		fx.Provide(NewReferenceStrategy),
		fx.Populate(&rt, &strat),
		fx.NopLogger,
	)

	if err := app.Start(ctx); err != nil {
		log.Fatalf("fx start: %v", err)
	}
	defer func() {
		if err := app.Stop(ctx); err != nil {
			log.Printf("fx stop: %v", err)
		}
	}()

	if err := rt.StartStandalone(strat, *configDir, *wispYml); err != nil {
		log.Fatalf("StartStandalone: %v", err)
	}

	log.Println("strategy started — waiting for signal or remote /shutdown")
	if err := rt.Wait(); err != nil {
		log.Printf("Wait: %v", err)
		os.Exit(1)
	}
	log.Println("shutdown complete")
}
