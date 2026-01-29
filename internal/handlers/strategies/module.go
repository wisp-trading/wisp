package strategies

import (
	"github.com/wisp-trading/wisp/internal/handlers/strategies/browse"
	"github.com/wisp-trading/wisp/internal/handlers/strategies/compile"
	"go.uber.org/fx"
)

var Module = fx.Options(
	browse.Module,
	compile.Module,
)
