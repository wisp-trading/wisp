package services

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateProjectWritesStandaloneStarter(t *testing.T) {
	dir := t.TempDir()
	cwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	svc := NewScaffoldService()
	if err := svc.CreateProject("demo_bot"); err != nil {
		t.Fatal(err)
	}

	main := filepath.Join("demo_bot", "strategies", "starter", "main.go")
	if _, err := os.Stat(main); err != nil {
		t.Fatalf("missing main.go: %v", err)
	}
	cfg := filepath.Join("demo_bot", "strategies", "starter", "config.yml")
	if _, err := os.Stat(cfg); err != nil {
		t.Fatal(err)
	}
	// No credential files in project
	for _, bad := range []string{
		filepath.Join("demo_bot", "exchanges.yml"),
		filepath.Join("demo_bot", "wisp.yml"),
	} {
		if _, err := os.Stat(bad); err == nil {
			t.Fatalf("should not create %s", bad)
		}
	}
	body, _ := os.ReadFile(main)
	if !contains(string(body), "StartStandalone") || !contains(string(body), "Wait") {
		t.Fatal("main.go must use StartStandalone + Wait")
	}
	cfgBody, _ := os.ReadFile(cfg)
	if contains(string(cfgBody), "instruments:") {
		t.Fatal("config.yml must not list unused instruments — domain = connector MarketType")
	}
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})())
}
