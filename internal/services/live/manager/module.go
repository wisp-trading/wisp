package manager

import (
	"context"

	"github.com/wisp-trading/sdk/pkg/types/logging"
	"github.com/wisp-trading/sdk/pkg/types/monitoring"
	"github.com/wisp-trading/wisp/pkg/live"
	"go.uber.org/fx"
)

// Module provides the manager components via Fx
var Module = fx.Module(
	"live/manager",
	fx.Provide(
		NewFileStateStore,
		NewProcessSpawner,
		provideInstanceManager,
	),
	fx.Invoke(initializeInstanceManager),
)

type instanceManagerParams struct {
	fx.In
	StateStore live.StateStore
	Spawner    live.ProcessSpawner
	Logger     logging.ApplicationLogger
	Querier    monitoring.ViewQuerier `optional:"true"`
}

func provideInstanceManager(params instanceManagerParams) live.InstanceManager {
	return NewInstanceManager(params.StateStore, params.Spawner, params.Logger, params.Querier)
}

// initializeInstanceManager loads running instances from state file on startup
func initializeInstanceManager(lc fx.Lifecycle, manager live.InstanceManager) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Best-effort restore. Missing state on first run is expected.
			_ = manager.LoadRunning(ctx)
			return nil
		},
	})
}
