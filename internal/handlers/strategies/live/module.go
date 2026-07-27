package live

import (
	"github.com/wisp-trading/connectors/pkg/connectors"
	"github.com/wisp-trading/sdk/wisp"
	"github.com/wisp-trading/wisp/internal/services/live"
	"github.com/wisp-trading/wisp/internal/services/live/manager"
	"go.uber.org/fx"
)

// Module provides live trading: connectors registry + instance manager + TUI start.
// Strategies run as standalone binaries only (no plugin runtime in-process).
var Module = fx.Module("live",
	wisp.Module,
	connectors.Module,
	manager.Module,
	fx.Provide(live.NewLiveService),
	fx.Provide(NewLiveViewFactory),
)
