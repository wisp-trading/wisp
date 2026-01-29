package app

import (
	"github.com/wisp-trading/wisp/internal/handlers"
	"github.com/wisp-trading/wisp/internal/handlers/strategies"
	"github.com/wisp-trading/wisp/internal/handlers/strategies/backtest"
	"github.com/wisp-trading/wisp/internal/handlers/strategies/live"
	"github.com/wisp-trading/wisp/internal/router"
	"github.com/wisp-trading/wisp/internal/services/compile"
	"github.com/wisp-trading/wisp/internal/setup"
	"go.uber.org/fx"
)

// Module provides all application dependencies by composing domain modules
var Module = fx.Options(
	backtest.Module,
	setup.Module,
	handlers.Module,
	router.Module,
	live.Module,
	strategies.Module,
	compile.Module,
)
