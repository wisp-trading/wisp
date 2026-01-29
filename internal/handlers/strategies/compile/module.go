package compile

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/wisp-trading/sdk/pkg/types/config"
	strategyTypes "github.com/wisp-trading/wisp/pkg/strategy"
	"go.uber.org/fx"
)

// CompileViewFactory creates compile views with transient strategy data
type CompileViewFactory func(*config.Strategy) tea.Model

// Module provides compile view constructor in DI
var Module = fx.Module("compile",
	fx.Provide(
		NewCompileViewFactory,
	),
)

// NewCompileViewFactory creates the factory function for compile views
func NewCompileViewFactory(
	compileService strategyTypes.CompileService,
) CompileViewFactory {
	return func(s *config.Strategy) tea.Model {
		model := NewCompileModel(compileService)
		model.SetStrategy(s)
		return model
	}
}
