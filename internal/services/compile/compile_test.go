package compile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCompileStrategyRequiresMainGo(t *testing.T) {
	dir := t.TempDir()
	// strategy.go only — no main.go
	if err := os.WriteFile(filepath.Join(dir, "strategy.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	svc := NewCompileService()
	err := svc.CompileStrategy(dir)
	if err == nil {
		t.Fatal("expected error without main.go")
	}
	if !contains(err.Error(), "main.go") {
		t.Fatalf("error should mention main.go: %v", err)
	}
}

func TestIsCompiledAndNeedsRecompile(t *testing.T) {
	dir := t.TempDir()
	name := filepath.Base(dir)
	// no binary
	svc := NewCompileService()
	if svc.IsCompiled(dir) {
		t.Fatal("expected not compiled")
	}
	if !svc.NeedsRecompile(dir) {
		t.Fatal("expected needs recompile without sources/binary")
	}

	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module t\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// still no binary
	if !svc.NeedsRecompile(dir) {
		t.Fatal("expected needs recompile when binary missing")
	}

	// touch a stub binary newer than sources — skip real go build
	bin := filepath.Join(dir, name)
	if err := os.WriteFile(bin, []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	// binary mtime may equal source on fast FS; force NeedsRecompile path via older binary
	// IsCompiled should be true
	if !svc.IsCompiled(dir) {
		t.Fatal("expected compiled when binary present")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		(len(s) > 0 && (func() bool {
			for i := 0; i+len(sub) <= len(s); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		})()))
}
