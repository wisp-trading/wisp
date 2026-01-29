package handlers

import (
	"github.com/wisp-trading/wisp/internal/handlers/settings"
	"github.com/wisp-trading/wisp/internal/handlers/strategies"
	"github.com/wisp-trading/wisp/internal/handlers/strategies/monitor"
	"go.uber.org/fx"
)

// Module provides the root handler
var Module = fx.Module("handlers",
	// Monitor view factory
	monitor.Module,
	settings.Module,

	fx.Provide(strategies.NewStrategyBrowser),
	fx.Provide(NewRootHandler),
)
