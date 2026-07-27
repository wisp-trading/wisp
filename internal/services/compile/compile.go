package compile

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/wisp-trading/wisp/pkg/strategy"
)

// compileService builds strategy packages for live execution.
// Blessed path: standalone binary (main.go present).
// Legacy: Go plugin .so when only strategy.go exists without main.
type compileService struct{}

func NewCompileService() strategy.CompileService {
	return &compileService{}
}

// CompileStrategy builds a standalone binary when main.go is present,
// otherwise falls back to the legacy plugin .so path.
func (s *compileService) CompileStrategy(strategyPath string) error {
	if isStandalone(strategyPath) {
		return s.compileBinary(strategyPath)
	}
	return s.compilePlugin(strategyPath)
}

func isStandalone(strategyPath string) bool {
	_, err := os.Stat(filepath.Join(strategyPath, "main.go"))
	return err == nil
}

func binaryPath(strategyPath string) string {
	return filepath.Join(strategyPath, filepath.Base(strategyPath))
}

func soPath(strategyPath string) string {
	name := filepath.Base(strategyPath)
	return filepath.Join(strategyPath, name+".so")
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

	fmt.Printf("🔨 Compiling standalone strategy %s...\n", name)
	fmt.Printf("  📦 go mod tidy...\n")
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = strategyPath
	if outBytes, err := tidyCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go mod tidy failed: %s", string(outBytes))
	}

	fmt.Printf("  🔧 go build -o %s .\n", name)
	cmd := exec.Command("go", "build", "-o", name, ".")
	cmd.Dir = strategyPath
	if outBytes, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("standalone build failed: %s", string(outBytes))
	}

	fmt.Printf("✅ Built standalone binary %s\n\n", out)
	return nil
}

// compilePlugin is the legacy .so path (demoted).
func (s *compileService) compilePlugin(strategyPath string) error {
	strategyName := filepath.Base(strategyPath)
	strategyGoPath := filepath.Join(strategyPath, "strategy.go")
	so := soPath(strategyPath)

	if _, err := os.Stat(strategyGoPath); os.IsNotExist(err) {
		return fmt.Errorf("strategy.go not found (and no main.go for standalone)")
	}

	goInfo, err := os.Stat(strategyGoPath)
	if err != nil {
		return err
	}
	if soInfo, err := os.Stat(so); err == nil && soInfo.ModTime().After(goInfo.ModTime()) {
		return nil
	}

	_ = os.Remove(so)

	fmt.Printf("🔨 Compiling %s strategy (legacy plugin)...\n", strategyName)
	fmt.Printf("  ⚠️  Plugin packaging is legacy; prefer main.go + StartStandalone + Wait\n")

	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = strategyPath
	if out, err := tidyCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to download dependencies: %s", string(out))
	}

	outputFileName := strategyName + ".so"
	cmd := exec.Command("go", "build", "-a", "-buildmode=plugin", "-o", outputFileName, "strategy.go")
	cmd.Dir = strategyPath
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("compilation failed: %s", string(out))
	}

	fmt.Printf("✅ Compiled %s.so successfully\n\n", strategyName)
	return nil
}

// PreCompileStrategies scans and compiles all strategies in the strategies directory.
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

// IsCompiled reports whether a binary or legacy .so is present.
func (s *compileService) IsCompiled(strategyPath string) bool {
	if _, err := os.Stat(binaryPath(strategyPath)); err == nil {
		return true
	}
	_, err := os.Stat(soPath(strategyPath))
	return err == nil
}

// NeedsRecompile reports whether sources are newer than the artifact.
func (s *compileService) NeedsRecompile(strategyPath string) bool {
	srcTime, err := compileSourceModTime(strategyPath)
	if err != nil {
		return true
	}
	var art string
	if isStandalone(strategyPath) {
		art = binaryPath(strategyPath)
	} else {
		art = soPath(strategyPath)
	}
	info, err := os.Stat(art)
	if err != nil {
		return true
	}
	return srcTime.After(info.ModTime())
}
