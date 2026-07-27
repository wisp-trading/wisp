package manager

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/wisp-trading/sdk/pkg/types/config"
	"github.com/wisp-trading/sdk/pkg/types/logging"
	"github.com/wisp-trading/sdk/pkg/types/monitoring"
	"github.com/wisp-trading/wisp/pkg/live"
)

// DefaultStopTimeout is how long Stop waits after a graceful request before SIGKILL.
const DefaultStopTimeout = 10 * time.Second

type instanceManager struct {
	mu         sync.RWMutex
	instances  map[string]*live.Instance
	stateStore live.StateStore
	spawner    live.ProcessSpawner
	logger     logging.ApplicationLogger
	querier    monitoring.ViewQuerier // optional: HTTP /shutdown when available
}

// NewInstanceManager creates a new instance manager.
// querier may be nil; when set, Stop prefers HTTP /shutdown before OS signals.
func NewInstanceManager(
	stateStore live.StateStore,
	spawner live.ProcessSpawner,
	logger logging.ApplicationLogger,
	querier monitoring.ViewQuerier,
) live.InstanceManager {
	return &instanceManager{
		instances:  make(map[string]*live.Instance),
		stateStore: stateStore,
		spawner:    spawner,
		logger:     logger,
		querier:    querier,
	}
}

// Start spawns a new strategy instance.
func (im *instanceManager) Start(ctx context.Context, strategy *config.Strategy, frameworkRoot string) (*live.Instance, error) {
	im.mu.Lock()
	defer im.mu.Unlock()

	for id, inst := range im.instances {
		if inst.StrategyName == strategy.Name && inst.Status == live.StatusRunning {
			if processAlive(inst.PID) {
				return nil, fmt.Errorf("strategy '%s' already running", strategy.Name)
			}
			inst.Status = live.StatusStopped
			delete(im.instances, id)
		}
	}

	cmd, err := im.spawner.Spawn(ctx, strategy)
	if err != nil {
		return nil, fmt.Errorf("failed to spawn process: %w", err)
	}

	instCtx, cancel := context.WithCancel(ctx)
	instance := &live.Instance{
		ID:              uuid.New().String(),
		StrategyName:    strategy.Name,
		StrategyPath:    strategy.Path,
		FrameworkRoot:   frameworkRoot,
		Status:          live.StatusRunning,
		StartedAt:       time.Now(),
		LastStatusCheck: time.Now(),
		Context:         instCtx,
		Cancel:          cancel,
		Cmd:             cmd,
	}

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start process: %w", err)
	}

	instance.PID = cmd.Process.Pid
	im.instances[instance.ID] = instance
	go im.monitorProcess(instance)
	_ = im.saveStateLocked()

	return instance, nil
}

// Stop gracefully terminates an instance.
// Order: HTTP /shutdown (if available) → wait for exit → SIGKILL last resort.
// Falls back to SIGINT when HTTP is unavailable, then same wait/kill.
func (im *instanceManager) Stop(instanceID string) error {
	im.mu.Lock()
	instance, exists := im.instances[instanceID]
	if !exists {
		im.mu.Unlock()
		return fmt.Errorf("instance not found: %s", instanceID)
	}
	strategyName := instance.StrategyName
	pid := instance.PID
	cmd := instance.Cmd
	if instance.Cancel != nil {
		instance.Cancel()
	}
	im.mu.Unlock()

	httpOK := false
	if im.querier != nil {
		if err := im.querier.Shutdown(strategyName); err != nil {
			im.logger.Warn("HTTP shutdown request failed, will signal process",
				"strategy", strategyName, "error", err)
		} else {
			httpOK = true
			im.logger.Info("Sent HTTP /shutdown", "strategy", strategyName, "id", instanceID)
		}
	}

	if !httpOK {
		if err := signalProcess(cmd, pid, os.Interrupt); err != nil {
			im.logger.Warn("Failed to signal process", "pid", pid, "error", err)
		}
	}

	if err := waitForExit(cmd, pid, DefaultStopTimeout); err != nil {
		im.logger.Warn("Graceful stop timeout, force killing", "instance", instanceID, "pid", pid)
		if killErr := killProcess(cmd, pid); killErr != nil {
			return fmt.Errorf("failed to kill process after timeout: %w", killErr)
		}
		_ = waitForExit(cmd, pid, 2*time.Second)
	}

	im.markStopped(instanceID)
	return nil
}

// StopByStrategyName gracefully terminates an instance by strategy name.
func (im *instanceManager) StopByStrategyName(strategyName string) error {
	im.mu.RLock()
	var instanceID string

	im.logger.Info("Searching for instance to stop",
		"strategy", strategyName,
		"total_instances", len(im.instances))

	for id, inst := range im.instances {
		if inst.StrategyName == strategyName && inst.Status == live.StatusRunning {
			instanceID = id
			break
		}
	}
	im.mu.RUnlock()

	if instanceID == "" {
		return fmt.Errorf("no running instance found for strategy: %s (instances in memory: %d) - try reloading instances from state",
			strategyName, len(im.instances))
	}

	return im.Stop(instanceID)
}

// Kill forcefully terminates an instance (PID-based when Cmd is nil / reattached).
func (im *instanceManager) Kill(instanceID string) error {
	im.mu.Lock()
	instance, exists := im.instances[instanceID]
	if !exists {
		im.mu.Unlock()
		return fmt.Errorf("instance not found: %s", instanceID)
	}
	cmd := instance.Cmd
	pid := instance.PID
	strategyName := instance.StrategyName
	if instance.Cancel != nil {
		instance.Cancel()
	}
	im.mu.Unlock()

	if err := killProcess(cmd, pid); err != nil {
		return fmt.Errorf("failed to kill process: %w", err)
	}

	im.markStopped(instanceID)
	im.logger.Info("Killed instance", "strategy", strategyName, "id", instanceID)
	return nil
}

// Get retrieves a specific instance.
func (im *instanceManager) Get(instanceID string) (*live.Instance, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	instance, exists := im.instances[instanceID]
	if !exists {
		return nil, fmt.Errorf("instance not found: %s", instanceID)
	}
	return instance, nil
}

// List returns all instances (filtered by status).
func (im *instanceManager) List(status live.InstanceStatus) ([]*live.Instance, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	var result []*live.Instance
	for _, instance := range im.instances {
		if status == "" || instance.Status == status {
			result = append(result, instance)
		}
	}
	return result, nil
}

// LoadRunning loads instances from state file (after restart).
func (im *instanceManager) LoadRunning(ctx context.Context) error {
	instances, err := im.stateStore.Load()
	if err != nil {
		return err
	}

	im.mu.Lock()
	defer im.mu.Unlock()

	for _, instance := range instances {
		if instance.Status != live.StatusRunning {
			continue
		}

		// FindProcess alone is not enough on Unix — use Signal(0)
		if !processAlive(instance.PID) {
			instance.Status = live.StatusCrashed
			instance.Error = "Process not found after restart"
			im.instances[instance.ID] = instance
			continue
		}

		instCtx, cancel := context.WithCancel(ctx)
		instance.Context = instCtx
		instance.Cancel = cancel
		// Cmd stays nil for reattached instances
		im.instances[instance.ID] = instance
		go im.monitorProcess(instance)
	}

	return im.saveStateLocked()
}

// SaveState persists current state to disk.
func (im *instanceManager) SaveState() error {
	im.mu.Lock()
	defer im.mu.Unlock()
	return im.saveStateLocked()
}

func (im *instanceManager) saveStateLocked() error {
	instances := make([]*live.Instance, 0, len(im.instances))
	for _, inst := range im.instances {
		instances = append(instances, inst)
	}
	return im.stateStore.Save(instances)
}

// Shutdown gracefully terminates all running instances.
func (im *instanceManager) Shutdown(ctx context.Context, timeout time.Duration) error {
	im.mu.RLock()
	instanceIDs := make([]string, 0, len(im.instances))
	for id, inst := range im.instances {
		if inst.Status == live.StatusRunning {
			instanceIDs = append(instanceIDs, id)
		}
	}
	im.mu.RUnlock()

	shutdownCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	done := make(chan error, len(instanceIDs))
	for _, id := range instanceIDs {
		go func(instID string) {
			done <- im.Stop(instID)
		}(id)
	}

	var errs []error
	for i := 0; i < len(instanceIDs); i++ {
		select {
		case err := <-done:
			if err != nil {
				errs = append(errs, err)
			}
		case <-shutdownCtx.Done():
			return shutdownCtx.Err()
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}
	return nil
}

func (im *instanceManager) markStopped(instanceID string) {
	im.mu.Lock()
	defer im.mu.Unlock()
	if inst, ok := im.instances[instanceID]; ok {
		inst.Status = live.StatusStopped
		inst.PID = 0
	}
	_ = im.saveStateLocked()
}

// monitorProcess monitors a running process for crashes using Signal(0).
func (im *instanceManager) monitorProcess(instance *live.Instance) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-instance.Context.Done():
			return
		case <-ticker.C:
			im.mu.RLock()
			pid := instance.PID
			status := instance.Status
			im.mu.RUnlock()

			if status != live.StatusRunning {
				return
			}

			if !processAlive(pid) {
				im.mu.Lock()
				if instance.Status == live.StatusRunning {
					instance.Status = live.StatusCrashed
					instance.Error = "Process exited unexpectedly"
					instance.PID = 0
					_ = im.saveStateLocked()
					im.logger.Error("Instance crashed", "strategy", instance.StrategyName, "id", instance.ID)
				}
				im.mu.Unlock()
				return
			}

			im.mu.Lock()
			instance.LastStatusCheck = time.Now()
			im.mu.Unlock()
		}
	}
}

func signalProcess(cmd *exec.Cmd, pid int, sig os.Signal) error {
	if cmd != nil && cmd.Process != nil {
		return cmd.Process.Signal(sig)
	}
	if pid <= 0 {
		return fmt.Errorf("invalid pid %d", pid)
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return proc.Signal(sig)
}

// killProcess force-kills via Cmd or PID. Nil Cmd (reattached) is supported.
func killProcess(cmd *exec.Cmd, pid int) error {
	if cmd != nil && cmd.Process != nil {
		if err := cmd.Process.Kill(); err == nil {
			return nil
		}
	}
	if pid <= 0 {
		return fmt.Errorf("instance has no valid process reference")
	}
	if !processAlive(pid) {
		return nil
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return proc.Kill()
}

// waitForExit waits until the process exits or timeout elapses.
func waitForExit(cmd *exec.Cmd, pid int, timeout time.Duration) error {
	if cmd != nil && cmd.Process != nil {
		done := make(chan error, 1)
		go func() { done <- cmd.Wait() }()
		select {
		case <-done:
			return nil
		case <-time.After(timeout):
			return fmt.Errorf("timeout waiting for process to exit")
		}
	}

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !processAlive(pid) {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !processAlive(pid) {
		return nil
	}
	return fmt.Errorf("timeout waiting for process to exit")
}
