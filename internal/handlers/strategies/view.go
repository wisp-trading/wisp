package strategies

import (
	"github.com/wisp-trading/sdk/pkg/types/config"
	"github.com/wisp-trading/wisp/internal/handlers/strategies/browse"
	"github.com/wisp-trading/wisp/internal/router"
	strategyTypes "github.com/wisp-trading/wisp/pkg/strategy"
)

// StrategyBrowser handles browsing strategies and selecting actions
type StrategyBrowser interface {
}

type strategyBrowser struct {
	strategyService config.StrategyConfig
	compileService  strategyTypes.CompileService
	listFactory     browse.StrategyListViewFactory
	router          router.Router
}

func NewStrategyBrowser(
	strategyService config.StrategyConfig,
	compileService strategyTypes.CompileService,
	listFactory browse.StrategyListViewFactory,
	r router.Router,
) StrategyBrowser {
	return &strategyBrowser{
		strategyService: strategyService,
		compileService:  compileService,
		listFactory:     listFactory,
		router:          r,
	}
}
