package runtime

import (
	"fmt"

	"github.com/wisp-trading/sdk/pkg/types/config"
	"github.com/wisp-trading/sdk/pkg/types/logging"
	"github.com/wisp-trading/sdk/pkg/types/runtime"
	"github.com/wisp-trading/wisp/pkg/live"
)

type liveRuntime struct {
	logger       logging.ApplicationLogger
	runtime      runtime.Runtime
	configLoader config.StartupConfigLoader
}

func NewRuntime(
	logger logging.ApplicationLogger,
	runtime runtime.Runtime,
	configLoader config.StartupConfigLoader,
) live.Runtime {
	return &liveRuntime{
		logger:       logger,
		runtime:      runtime,
		configLoader: configLoader,
	}
}

// Run starts a strategy (plugin path — legacy) and blocks on the shared
// shutdown contract: OS signals and remote HTTP /shutdown both exit cleanly.
func (r *liveRuntime) Run(strategyDir string) error {
	wispPath := "wisp.yml"
	cfg, err := r.configLoader.LoadForStrategy(strategyDir, wispPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	r.logger.Info("Config loaded", "strategy", cfg.Strategy.Name)
	r.logger.Warn("Plugin packaging path is legacy; prefer standalone binaries with StartStandalone + Wait")

	err = r.runtime.Start(strategyDir, wispPath)
	if err != nil {
		return fmt.Errorf("failed to start: %w", err)
	}

	r.logger.Info("SDK startup complete")
	r.logger.Info("Strategy running; waiting for signal or remote /shutdown...")

	// Shared contract: Wait selects on OS signals + ShutdownRequested, then Stop.
	if err := r.runtime.Wait(); err != nil {
		r.logger.Error("Failed during shutdown wait", "error", err)
		return err
	}

	r.logger.Info("Shutdown complete")
	return nil
}
