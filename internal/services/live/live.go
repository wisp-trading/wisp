package live

import (
	"context"
	"fmt"
	"os"

	"github.com/wisp-trading/sdk/pkg/types/config"
	strategyTypes "github.com/wisp-trading/wisp/pkg/strategy"

	"github.com/wisp-trading/sdk/pkg/types/logging"
	"github.com/wisp-trading/wisp/pkg/live"
)

type LiveService interface {
	ExecuteStrategy(ctx context.Context, strategy *config.Strategy) error
}

// liveService orchestrates live trading by coordinating other services
type liveService struct {
	connectorService config.ConnectorService
	compile          strategyTypes.CompileService
	logger           logging.ApplicationLogger
	manager          live.InstanceManager
}

func NewLiveService(
	connectorService config.ConnectorService,
	compileSvc strategyTypes.CompileService,
	logger logging.ApplicationLogger,
	manager live.InstanceManager,
) LiveService {
	return &liveService{
		connectorService: connectorService,
		compile:          compileSvc,
		logger:           logger,
		manager:          manager,
	}
}

// ExecuteStrategy runs the selected strategy with all its configured exchanges
func (s *liveService) ExecuteStrategy(ctx context.Context, strategy *config.Strategy) error {
	if strategy == nil {
		return fmt.Errorf("no strategy selected")
	}
	if strategy.Error != "" {
		return fmt.Errorf("strategy '%s' has invalid config.yml: %s", strategy.Name, strategy.Error)
	}
	if len(strategy.Exchanges) == 0 {
		return fmt.Errorf(
			"strategy '%s' has no exchanges in config.yml — add at least one (e.g. hyperliquid)",
			strategy.Name,
		)
	}

	// 1. Pre-validate keys for strategy exchanges (hard fail with actionable message)
	connectorConfigs, err := s.connectorService.GetConnectorConfigsForStrategy(strategy.Exchanges)
	if err != nil {
		return fmt.Errorf(
			"cannot start strategy '%s': %w\n\nFix: wisp → Settings → add/enable keys for %v (~/.wisp/connectors.yml)\nInclude connectors.Module in the strategy fx graph",
			strategy.Name, err, strategy.Exchanges,
		)
	}

	s.logger.Info("Validated connector configs", "strategy", strategy.Name, "connectors", len(connectorConfigs))

	// 2. Compile standalone binary (main.go required)
	if err := s.compile.CompileStrategy(strategy.Path); err != nil {
		return fmt.Errorf("failed to compile strategy: %w", err)
	}

	// 3. Get current working directory as framework root
	frameworkRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	_, err = s.manager.Start(ctx, strategy, frameworkRoot)
	return err
}
