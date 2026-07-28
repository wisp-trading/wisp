package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/wisp-trading/sdk/pkg/types/wisp/numerical"
	"gopkg.in/yaml.v3"
)

// HardMaxSize is the absolute ceiling on order size (BTC base).
// config.yml cannot raise size above this without a code change.
const HardMaxSizeBTC = 0.01

// strategyYAML is the subset of config.yml we care about.
type strategyYAML struct {
	Parameters map[string]interface{} `yaml:"parameters"`
}

// LoadParamsFromConfigDir reads config.yml parameters and applies hard max size.
func LoadParamsFromConfigDir(configDir string) (Params, time.Duration, string, int, error) {
	p := DefaultParams()
	tick := 30 * time.Second
	interval := "15m"
	limit := 60

	path := filepath.Join(configDir, "config.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return clampParams(p), tick, interval, limit, nil
		}
		return p, tick, interval, limit, err
	}
	var y strategyYAML
	if err := yaml.Unmarshal(data, &y); err != nil {
		return p, tick, interval, limit, fmt.Errorf("parse config.yml: %w", err)
	}
	params := y.Parameters
	if params == nil {
		return clampParams(p), tick, interval, limit, nil
	}

	if v, ok := paramDecimal(params, "funding_entry"); ok {
		p.FundingEntry = v
	}
	if v, ok := paramDecimal(params, "funding_exit"); ok {
		p.FundingExit = v
	}
	if v, ok := paramDecimal(params, "rsi_overbought"); ok {
		p.RSIOverbought = v
	}
	if v, ok := paramDecimal(params, "rsi_oversold"); ok {
		p.RSIOversold = v
	}
	if v, ok := paramDecimal(params, "rsi_mid"); ok {
		p.RSIMid = v
	}
	if v, ok := paramDecimal(params, "size"); ok {
		p.Size = v
	}
	if v, ok := paramString(params, "kline_interval"); ok && v != "" {
		interval = v
	}
	if v, ok := paramInt(params, "kline_limit"); ok && v > 0 {
		limit = v
	}
	if v, ok := paramInt(params, "tick_seconds"); ok && v > 0 {
		tick = time.Duration(v) * time.Second
	}

	return clampParams(p), tick, interval, limit, nil
}

func clampParams(p Params) Params {
	max := numerical.NewFromFloat(HardMaxSizeBTC)
	if p.Size.GreaterThan(max) {
		p.Size = max
	}
	if p.Size.IsNegative() || p.Size.IsZero() {
		p.Size = numerical.NewFromFloat(0.001)
	}
	return p
}

func paramDecimal(m map[string]interface{}, key string) (numerical.Decimal, bool) {
	raw, ok := m[key]
	if !ok || raw == nil {
		return numerical.Decimal{}, false
	}
	switch v := raw.(type) {
	case string:
		d, err := numerical.NewFromString(v)
		if err != nil {
			return numerical.Decimal{}, false
		}
		return d, true
	case float64:
		return numerical.NewFromFloat(v), true
	case int:
		return numerical.NewFromFloat(float64(v)), true
	default:
		d, err := numerical.NewFromString(fmt.Sprint(v))
		if err != nil {
			return numerical.Decimal{}, false
		}
		return d, true
	}
}

func paramString(m map[string]interface{}, key string) (string, bool) {
	raw, ok := m[key]
	if !ok || raw == nil {
		return "", false
	}
	s, ok := raw.(string)
	return s, ok
}

func paramInt(m map[string]interface{}, key string) (int, bool) {
	raw, ok := m[key]
	if !ok || raw == nil {
		return 0, false
	}
	switch v := raw.(type) {
	case int:
		return v, true
	case float64:
		return int(v), true
	case string:
		var n int
		_, err := fmt.Sscanf(v, "%d", &n)
		return n, err == nil
	default:
		return 0, false
	}
}

// resolveDryRun: live only when deliberately armed.
// Control injects WISP_DRY_RUN=1 for dry_run registry starts.
// CLI: FUNDING_TILT_LIVE=1 without WISP_DRY_RUN.
// Control live: WISP_RUN_MODE=live and not WISP_DRY_RUN.
func resolveDryRun() bool {
	if os.Getenv("WISP_DRY_RUN") == "1" {
		return true
	}
	if os.Getenv("WISP_CONTROL_STARTED") == "1" {
		return os.Getenv("WISP_RUN_MODE") != "live"
	}
	// Standalone CLI
	return os.Getenv("FUNDING_TILT_LIVE") != "1"
}
