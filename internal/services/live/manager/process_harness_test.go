package manager_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/wisp-trading/sdk/pkg/types/config"
	"github.com/wisp-trading/sdk/pkg/types/logging"
	"github.com/wisp-trading/wisp/internal/services/live/manager"
	"github.com/wisp-trading/wisp/pkg/live"
)

type memoryState struct {
	instances []*live.Instance
}

func (m *memoryState) Save(instances []*live.Instance) error {
	m.instances = append([]*live.Instance(nil), instances...)
	return nil
}

func (m *memoryState) Load() ([]*live.Instance, error) {
	return append([]*live.Instance(nil), m.instances...), nil
}

func (m *memoryState) GetPath() string { return ":memory:" }

type fixtureSpawner struct{ bin string }

func (f *fixtureSpawner) Spawn(_ context.Context, _ *config.Strategy) (*exec.Cmd, error) {
	return exec.Command(f.bin), nil
}

func (f *fixtureSpawner) AttachMonitor(_ *live.Instance) error { return nil }

func buildSignalFixture(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("signal harness is Unix-oriented")
	}
	dir := t.TempDir()
	src := filepath.Join(dir, "main.go")
	body := `package main
import (
  "os"
  "os/signal"
  "syscall"
)
func main() {
  ch := make(chan os.Signal, 1)
  signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
  <-ch
}
`
	if err := os.WriteFile(src, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(dir, "fixture")
	cmd := exec.Command("go", "build", "-o", bin, src)
	cmd.Env = append(os.Environ(), "GOWORK=off")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build fixture: %v\n%s", err, out)
	}
	return bin
}

// Real process start/stop without exchange keys: spawn → Stop (SIGINT) → exits.
func TestProcessHarnessStopViaSignal(t *testing.T) {
	bin := buildSignalFixture(t)
	im := manager.NewInstanceManager(&memoryState{}, &fixtureSpawner{bin: bin}, logging.NewNoOpLogger(), nil)

	inst, err := im.Start(context.Background(), &config.Strategy{Name: "fixture-harness", Path: t.TempDir()}, t.TempDir())
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if inst.PID == 0 {
		t.Fatal("expected non-zero PID")
	}
	if err := exec.Command("kill", "-0", fmt.Sprintf("%d", inst.PID)).Run(); err != nil {
		t.Fatalf("fixture not alive: %v", err)
	}

	done := make(chan error, 1)
	go func() { done <- im.Stop(inst.ID) }()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Stop: %v", err)
		}
	case <-time.After(15 * time.Second):
		_ = im.Kill(inst.ID)
		t.Fatal("Stop timed out")
	}

	running, err := im.List(live.StatusRunning)
	if err != nil {
		t.Fatal(err)
	}
	if len(running) > 0 {
		t.Fatalf("instance still running after Stop: %+v", running)
	}
}
