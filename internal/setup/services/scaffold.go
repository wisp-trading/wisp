package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
	"github.com/wisp-trading/wisp/internal/setup/types"
)

// scaffolder creates a minimal local project: strategies/<name>/ with main.go + config.yml.
// Credentials are never written here — use wisp TUI Settings → ~/.wisp/connectors.yml.
type scaffolder struct {
	fs afero.Fs
}

func NewScaffoldService() types.ScaffoldService {
	return &scaffolder{fs: afero.NewOsFs()}
}

// CreateProject scaffolds a starter strategy project under the given directory name.
func (s *scaffolder) CreateProject(name string) error {
	return s.CreateProjectWithStrategy(name, "starter")
}

// CreateProjectWithStrategy creates the project. strategyExample is the strategy folder name
// (defaults to "starter" when empty).
func (s *scaffolder) CreateProjectWithStrategy(name, strategyExample string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("project name cannot be empty")
	}
	if strategyExample == "" {
		strategyExample = "starter"
	}
	// Sanitize strategy folder
	strategyExample = strings.ReplaceAll(strategyExample, " ", "_")

	// Quiet: no stdout (TUI hosts this path). CLI init prints its own summary.
	if exists, _ := afero.DirExists(s.fs, name); exists {
		return fmt.Errorf("directory '%s' already exists", name)
	}

	strategyDir := filepath.Join(name, "strategies", strategyExample)
	if err := os.MkdirAll(strategyDir, 0o755); err != nil {
		return fmt.Errorf("create strategy dir: %w", err)
	}

	if err := s.writeStrategyFiles(name, strategyExample, strategyDir); err != nil {
		_ = os.RemoveAll(name)
		return err
	}

	if err := s.writeProjectRoot(name, strategyExample); err != nil {
		_ = os.RemoveAll(name)
		return err
	}
	return nil
}

func (s *scaffolder) writeStrategyFiles(project, strategy, strategyDir string) error {
	module := "github.com/example/" + project + "/strategies/" + strategy

	files := map[string]string{
		"go.mod": fmt.Sprintf(`module %s

go 1.26

require (
	github.com/wisp-trading/connectors v0.1.3
	github.com/wisp-trading/sdk v0.1.10
	go.uber.org/fx v1.24.0
)
`, module),
		"config.yml": fmt.Sprintf(`name: %s
description: Starter strategy (standalone binary)
# Domain (spot/perp/…) comes from the connector MarketType, not this file.
exchanges:
  - hyperliquid
assets:
  hyperliquid:
    - base: BTC
      quote: USD
parameters:
  dry_run: true
`, strategy),
		"main.go": `package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/wisp-trading/connectors/pkg/connectors"
	"github.com/wisp-trading/sdk/pkg/types/runtime"
	"github.com/wisp-trading/sdk/pkg/types/strategy"
	"github.com/wisp-trading/sdk/wisp"
	"go.uber.org/fx"
)

func main() {
	configDir := flag.String("config", ".", "strategy config directory")
	// Empty --wisp → ~/.wisp/connectors.yml (set keys via: wisp → Settings)
	wispPath := flag.String("wisp", "", "connector settings path (default: ~/.wisp/connectors.yml)")
	flag.Parse()

	ctx := context.Background()
	var (
		rt    runtime.Runtime
		strat strategy.Strategy
	)

	app := fx.New(
		connectors.Module,
		wisp.Module,
		fx.Provide(NewStrategy),
		fx.Populate(&rt, &strat),
		fx.NopLogger,
	)

	if err := app.Start(ctx); err != nil {
		log.Fatalf("fx start: %v", err)
	}
	defer func() { _ = app.Stop(ctx) }()

	if err := rt.StartStandalone(strat, *configDir, *wispPath); err != nil {
		log.Fatalf("StartStandalone: %v", err)
	}

	log.Println("running — Ctrl+C or Monitor → Stop")
	if err := rt.Wait(); err != nil {
		log.Printf("Wait: %v", err)
		os.Exit(1)
	}
}
`,
		"strategy.go": fmt.Sprintf(`package main

import (
	"context"
	"time"

	"github.com/wisp-trading/sdk/pkg/types/strategy"
	"github.com/wisp-trading/sdk/pkg/types/wisp"
)

// Strategy is a minimal self-directed strategy (no orders).
type Strategy struct {
	strategy.BaseStrategy
	k wisp.Wisp
}

func NewStrategy(k wisp.Wisp) strategy.Strategy {
	s := &Strategy{k: k}
	s.BaseStrategy = *strategy.NewBaseStrategy(strategy.BaseStrategyConfig{
		Name: "%s",
	})
	return s
}

func (s *Strategy) Start(ctx context.Context) error {
	return s.StartWithRunner(ctx, s.run)
}

func (s *Strategy) run(ctx context.Context) {
	s.k.Log().Info("strategy running")
	t := time.NewTicker(30 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			s.k.Log().Info("strategy stopped")
			return
		case <-t.C:
			s.k.Log().Info("heartbeat")
		}
	}
}
`, strategy),
	}

	for name, body := range files {
		path := filepath.Join(strategyDir, name)
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", name, err)
		}
	}
	return nil
}

func (s *scaffolder) writeProjectRoot(name, strategyExample string) error {
	readme := fmt.Sprintf(`# %s

Wisp strategy project.

## Quick start

1. **Keys** (once, global): run `+"`wisp`"+` → **Settings** → add Hyperliquid (or other) credentials  
   Saved to `+"`~/.wisp/connectors.yml`"+` — shared by all strategies.
2. **Deps:** `+"`cd strategies/%s && go mod tidy`"+`
3. **Run:** from that directory: `+"`go run .`"+`  
   Or from project root: `+"`wisp`"+` → **Strategies** → Start Live / **Monitor** → Stop

## Layout

`+"```"+`
%s/
└── strategies/
    └── %s/
        ├── main.go      # StartStandalone + Wait
        ├── strategy.go
        ├── config.yml   # exchanges/assets only (no secrets)
        └── go.mod
`+"```"+`

Do not put API keys in this repo.
`, name, strategyExample, name, strategyExample)

	if err := os.WriteFile(filepath.Join(name, "README.md"), []byte(readme), 0o644); err != nil {
		return err
	}

	gitignore := `# Never commit secrets
.env
.env.*

# Strategy build artifacts
strategies/*/*
!strategies/*/*.go
!strategies/*/*.yml
!strategies/*/go.mod
!strategies/*/go.sum
strategies/*/*~
*.so
*.exe

.DS_Store
.idea/
.vscode/
`
	if err := os.WriteFile(filepath.Join(name, ".gitignore"), []byte(gitignore), 0o644); err != nil {
		return err
	}
	return nil
}

// FormatProjectCreatedMsg is the post-init summary for CLI (not used under TUI).
func FormatProjectCreatedMsg(name, strategyExample string) string {
	if strategyExample == "" {
		strategyExample = "starter"
	}
	strat := filepath.Join(name, "strategies", strategyExample)
	return fmt.Sprintf(`✅ Project created: ./%s

Next:
  1. wisp                         # Settings → add exchange keys → ~/.wisp/connectors.yml
  2. cd %s && go mod tidy
  3. go run .                     # or: wisp → Strategies → Start Live
  4. wisp                         # Monitor → select → Stop
`, name, strat)
}
