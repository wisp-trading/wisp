package manager

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	sdkconfig "github.com/wisp-trading/sdk/pkg/types/config"
	"github.com/wisp-trading/sdk/pkg/types/logging"
	"github.com/wisp-trading/wisp/pkg/live"
)

type processSpawner struct {
	logger logging.ApplicationLogger
}

// NewProcessSpawner creates a new process spawner
func NewProcessSpawner(logger logging.ApplicationLogger) live.ProcessSpawner {
	return &processSpawner{
		logger: logger,
	}
}

// Spawn creates a strategy process from a standalone binary only.
func (ps *processSpawner) Spawn(ctx context.Context, strategy *sdkconfig.Strategy) (*exec.Cmd, error) {
	instanceLogDir := fmt.Sprintf(".wisp/instances/%s", strategy.Name)
	if err := os.MkdirAll(instanceLogDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create instance log directory: %w", err)
	}

	stdoutLog := fmt.Sprintf("%s/stdout.log", instanceLogDir)
	stderrLog := fmt.Sprintf("%s/stderr.log", instanceLogDir)

	stdoutFile, err := os.OpenFile(stdoutLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open stdout log: %w", err)
	}
	stderrFile, err := os.OpenFile(stderrLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		_ = stdoutFile.Close()
		return nil, fmt.Errorf("failed to open stderr log: %w", err)
	}

	cmd, err := ps.buildCommand(strategy)
	if err != nil {
		_ = stdoutFile.Close()
		_ = stderrFile.Close()
		return nil, err
	}

	// Detach from parent process group so the strategy survives TUI exit.
	_ = ctx
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	cmd.Stdout = stdoutFile
	cmd.Stderr = stderrFile

	ps.logger.Info("Spawning strategy process",
		"strategy", strategy.Name,
		"mode", "standalone",
		"path", cmd.Path,
		"args", cmd.Args,
		"stdout_log", stdoutLog,
		"stderr_log", stderrLog,
	)

	return cmd, nil
}

func (ps *processSpawner) buildCommand(strategy *sdkconfig.Strategy) (*exec.Cmd, error) {
	strategyPath := strategy.Path
	if strategyPath == "" {
		strategyPath = filepath.Join("strategies", strategy.Name)
	}

	bin := filepath.Join(strategyPath, strategy.Name)
	if abs, err := filepath.Abs(bin); err == nil {
		bin = abs
	}

	st, err := os.Stat(bin)
	if err != nil || st.IsDir() {
		return nil, fmt.Errorf(
			"standalone binary not found at %s — compile a main.go strategy (StartStandalone + Wait); plugin packaging has been removed",
			bin,
		)
	}

	configDir := strategyPath
	if abs, err := filepath.Abs(configDir); err == nil {
		configDir = abs
	}
	settings := sdkconfig.ResolveSettingsPath("")
	cmd := exec.Command(bin,
		"--config", configDir,
		"--wisp", settings,
	)
	return cmd, nil
}

// AttachMonitor starts monitoring process for crashes
func (ps *processSpawner) AttachMonitor(instance *live.Instance) error {
	if instance.Cmd == nil {
		return fmt.Errorf("command not set on instance")
	}
	ps.logger.Info("Attached monitor to instance", "strategy", instance.StrategyName, "pid", instance.PID)
	return nil
}
