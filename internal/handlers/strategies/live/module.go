package live

import (
	"github.com/wisp-trading/connectors/pkg/connectors"
	"github.com/wisp-trading/sdk/wisp"
	"github.com/wisp-trading/wisp/internal/services/live"
	"github.com/wisp-trading/wisp/internal/services/live/manager"
	"github.com/wisp-trading/wisp/internal/services/live/runtime"
	"go.uber.org/fx"
)

// Module provides all live trading dependencies including connectors registry and runtime
var Module = fx.Module("live",
	// Core SDK dependencies
	wisp.Module,

	// Live connectors
	connectors.Module,

	// Instance manager for multi-instance tracking and spawning
	manager.Module,

	// Runtime for strategy execution
	runtime.Module,

	// Services
	fx.Provide(live.NewLiveService),

	fx.Provide(
		NewLiveViewFactory,
	),
)
