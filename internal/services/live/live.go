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
func (s *liveService) ExecuteStrategy(ctx context.Context, strat *config.Strategy) error {
	// 1. Pre-validate that we have connectors for this strategy's exchanges
	connectorConfigs, err := s.connectorService.GetConnectorConfigsForStrategy(strat.Exchanges)
	if err != nil {
		return fmt.Errorf("cannot start strategy '%s': %w\n\nPlease check:\n- exchanges.yml has entries for: %v\n- Required exchanges are enabled\n- Exchange connectors are available in the SDK",
			strat.Name, err, strat.Exchanges)
	}

	s.logger.Info("Validated connector configs", "strategy", strat.Name, "connectors", len(connectorConfigs))

	// 2. Compile strategy if needed
	if err := s.compile.CompileStrategy(strat.Path); err != nil {
		return fmt.Errorf("failed to compile strategy: %w", err)
	}

	// 3. Get current working directory as framework root
	frameworkRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	_, err = s.manager.Start(ctx, strat, frameworkRoot)
	return err
}
