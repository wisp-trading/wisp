package compile

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/wisp-trading/wisp/pkg/strategy"
)

// compileService builds standalone strategy binaries for live execution.
// Requires main.go (StartStandalone + Wait). Plugin/.so packaging is removed.
type compileService struct{}

func NewCompileService() strategy.CompileService {
	return &compileService{}
}

// CompileStrategy builds a standalone binary; requires main.go.
func (s *compileService) CompileStrategy(strategyPath string) error {
	if !isStandalone(strategyPath) {
		return fmt.Errorf(
			"strategy at %s has no main.go — standalone packaging is required (main + StartStandalone + Wait); plugin .so support has been removed",
			strategyPath,
		)
	}
	return s.compileBinary(strategyPath)
}

func isStandalone(strategyPath string) bool {
	_, err := os.Stat(filepath.Join(strategyPath, "main.go"))
	return err == nil
}

func binaryPath(strategyPath string) string {
	return filepath.Join(strategyPath, filepath.Base(strategyPath))
}

func compileSourceModTime(strategyPath string) (time.Time, error) {
	var newest time.Time
	entries, err := os.ReadDir(strategyPath)
	if err != nil {
		return newest, err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if filepath.Ext(name) != ".go" && name != "go.mod" && name != "go.sum" {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.ModTime().After(newest) {
			newest = info.ModTime()
		}
	}
	if newest.IsZero() {
		return newest, fmt.Errorf("no go sources in %s", strategyPath)
	}
	return newest, nil
}

func (s *compileService) compileBinary(strategyPath string) error {
	name := filepath.Base(strategyPath)
	out := binaryPath(strategyPath)

	srcTime, err := compileSourceModTime(strategyPath)
	if err != nil {
		return err
	}
	if info, err := os.Stat(out); err == nil && info.ModTime().After(srcTime) {
		return nil
	}

	_ = os.Remove(out)

	// Quiet build — TUI shows its own progress; only surface command failures.
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = strategyPath
	if outBytes, err := tidyCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go mod tidy failed: %s", string(outBytes))
	}

	cmd := exec.Command("go", "build", "-o", name, ".")
	cmd.Dir = strategyPath
	if outBytes, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("standalone build failed: %s", string(outBytes))
	}
	return nil
}

// PreCompileStrategies scans and compiles all strategies with main.go.
func (s *compileService) PreCompileStrategies(strategiesDir string) map[string]error {
	errors := make(map[string]error)

	entries, err := os.ReadDir(strategiesDir)
	if err != nil {
		return errors
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		strategyPath := filepath.Join(strategiesDir, entry.Name())
		configPath := filepath.Join(strategyPath, "config.yml")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			continue
		}
		if err := s.CompileStrategy(strategyPath); err != nil {
			errors[entry.Name()] = err
		}
	}
	return errors
}

// IsCompiled reports whether the standalone binary is present.
func (s *compileService) IsCompiled(strategyPath string) bool {
	_, err := os.Stat(binaryPath(strategyPath))
	return err == nil
}

// NeedsRecompile reports whether sources are newer than the binary.
func (s *compileService) NeedsRecompile(strategyPath string) bool {
	if !isStandalone(strategyPath) {
		return true
	}
	srcTime, err := compileSourceModTime(strategyPath)
	if err != nil {
		return true
	}
	info, err := os.Stat(binaryPath(strategyPath))
	if err != nil {
		return true
	}
	return srcTime.After(info.ModTime())
}
