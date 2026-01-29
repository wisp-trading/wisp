package backtest

import (
	"github.com/wisp-trading/wisp/internal/handlers/strategies/backtest/handlers"
	"github.com/wisp-trading/wisp/internal/handlers/strategies/backtest/services"
	"go.uber.org/fx"
)

// Module provides all backtesting dependencies
var Module = fx.Module("backtesting",
	// Services
	fx.Provide(services.NewBacktestService),
	fx.Provide(services.NewAnalyzeService),

	// Handlers
	fx.Provide(handlers.NewBacktestHandler),
	fx.Provide(handlers.NewAnalyzeHandler),
)
