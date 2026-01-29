package setup

import (
	"github.com/wisp-trading/wisp/internal/setup/handlers"
	"github.com/wisp-trading/wisp/internal/setup/services"
	"go.uber.org/fx"
)

// Module provides all setup/scaffolding dependencies
var Module = fx.Module("setup",
	// Services
	fx.Provide(services.NewScaffoldService),

	// Handlers
	fx.Provide(handlers.NewInitHandler),
)
