package runtime

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

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

func (r *liveRuntime) Run(strategyDir string) error {
	wispPath := "wisp.yml"
	cfg, err := r.configLoader.LoadForStrategy(strategyDir, wispPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	r.logger.Info("Config loaded", "strategy", cfg.Strategy.Name)

	err = r.runtime.Start(strategyDir, wispPath)
	if err != nil {
		return fmt.Errorf("failed to start: %w", err)
	}

	r.logger.Info("SDK startup complete")
	r.logger.Info("Strategy running, keeping process alive...")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	sig := <-sigChan
	r.logger.Info("Received shutdown signal", "signal", sig)

	r.logger.Info("Stopping strategy...")
	if err := r.runtime.Stop(); err != nil {
		r.logger.Error("Failed to stop strategy", "error", err)
	}

	r.logger.Info("Shutdown complete")
	return nil
}
