package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/wisp-trading/sdk/pkg/types/connector"
	"github.com/wisp-trading/sdk/pkg/types/portfolio"
	"github.com/wisp-trading/sdk/pkg/types/strategy"
	wisptypes "github.com/wisp-trading/sdk/pkg/types/wisp"
	"github.com/wisp-trading/sdk/pkg/types/wisp/numerical"
)

// FundingTilt fades crowded Hyperliquid perp positioning using funding + RSI.
//
// Thesis: positive funding + overbought RSI → shorts are paid by crowded longs;
// inverse for negative funding. Single-venue (no spot hedge).
//
// Default dry_run=true. Live only when:
//   - control: WISP_RUN_MODE=live and not WISP_DRY_RUN (registry live arm)
//   - CLI: FUNDING_TILT_LIVE=1 without WISP_DRY_RUN
//
// Position is seeded from the exchange on start to avoid double-entry after restart.
type FundingTilt struct {
	strategy.BaseStrategy
	k wisptypes.Wisp

	exchange   connector.ExchangeName
	params     Params
	dryRun     bool
	pos        Side
	posSeeded  bool
	posUnknown bool // live: refuse emit until position known
	tick       time.Duration
	interval   string
	limit      int
	configDir  string
}

// NewFundingTilt builds the strategy. configDir is used to load config.yml params.
func NewFundingTilt(k wisptypes.Wisp, configDir string) strategy.Strategy {
	params, tick, interval, limit, err := LoadParamsFromConfigDir(configDir)
	if err != nil {
		// Fall back to defaults; log once Start has logger.
		params = DefaultParams()
		tick = 30 * time.Second
		interval = "15m"
		limit = 60
	}

	s := &FundingTilt{
		k:         k,
		exchange:  connector.ExchangeName("hyperliquid"),
		params:    params,
		dryRun:    resolveDryRun(),
		pos:       Flat,
		tick:      tick,
		interval:  interval,
		limit:     limit,
		configDir: configDir,
	}

	// Data-flow / monitor tests: thresholds no normal market should hit.
	if os.Getenv("FUNDING_TILT_SAFE") == "1" {
		s.params = Params{
			FundingEntry:  numerical.NewFromFloat(0.5),
			FundingExit:   numerical.NewFromFloat(0.25),
			RSIOverbought: numerical.NewFromFloat(99.5),
			RSIOversold:   numerical.NewFromFloat(0.5),
			RSIMid:        numerical.NewFromFloat(50),
			Size:          numerical.NewFromFloat(0.001),
		}
		s.params = clampParams(s.params)
		s.tick = 15 * time.Second
	}

	s.BaseStrategy = *strategy.NewBaseStrategy(strategy.BaseStrategyConfig{
		Name: "funding-tilt",
	})
	return s
}

func (s *FundingTilt) Start(ctx context.Context) error {
	return s.StartWithRunner(ctx, s.run)
}

func (s *FundingTilt) run(ctx context.Context) {
	pair := s.k.Pair(s.k.Asset("BTC"), s.k.Asset("USD"))
	s.k.Perp().WatchPair(s.exchange, pair)

	s.k.Log().Info("funding-tilt started",
		"exchange", string(s.exchange),
		"pair", pair.Symbol(),
		"dry_run", s.dryRun,
		"size", s.params.Size.String(),
		"hard_max_size", fmt.Sprintf("%g", HardMaxSizeBTC),
		"funding_entry", s.params.FundingEntry.String(),
		"config_dir", s.configDir,
	)

	// Warm-up for ingestors + position feed
	time.Sleep(2 * time.Second)
	s.seedPosition(pair)

	t := time.NewTicker(s.tick)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			s.k.Log().Info("funding-tilt stopped")
			return
		case <-t.C:
			if !s.posSeeded {
				s.seedPosition(pair)
			}
			s.tickOnce(pair)
		}
	}
}

func (s *FundingTilt) seedPosition(pair portfolio.Pair) {
	pos, ok := s.k.Perp().Position(s.exchange, pair)
	if !ok || pos == nil {
		// No row: treat as flat for dry_run; live holds emit until seen once.
		if s.dryRun {
			s.pos = Flat
			s.posSeeded = true
			s.posUnknown = false
			s.k.Log().Info("funding-tilt seed", "pos", "flat", "source", "no_position_row")
			return
		}
		s.posUnknown = true
		s.k.Log().Info("funding-tilt seed waiting", "reason", "no exchange position yet")
		return
	}

	side := SideFromExchange(pos.Size, string(pos.Side))
	s.pos = side
	s.posSeeded = true
	s.posUnknown = false
	s.k.Log().Info("funding-tilt seed",
		"pos", sideName(side),
		"size", pos.Size.String(),
		"side", string(pos.Side),
	)
}

func (s *FundingTilt) tickOnce(pair portfolio.Pair) {
	price, ok := s.k.Perp().Price(s.exchange, pair)
	if !ok {
		s.status("scanning", "waiting for mark price", map[string]string{"pair": pair.Symbol()})
		return
	}

	fr, ok := s.k.Perp().FundingRate(s.exchange, pair)
	if !ok || fr == nil {
		s.status("scanning", "waiting for funding rate", map[string]string{
			"price": price.StringFixed(2),
		})
		return
	}
	funding := fr.CurrentRate

	klines := s.k.Perp().Klines(s.exchange, pair, s.interval, s.limit)
	if len(klines) < 16 {
		s.status("scanning", "warming klines for RSI", map[string]string{
			"klines":  fmt.Sprintf("%d", len(klines)),
			"price":   price.StringFixed(2),
			"funding": funding.String(),
		})
		return
	}

	rsi, err := s.k.Indicators().RSI(klines, 14)
	if err != nil {
		s.k.Log().Info("RSI error", "err", err.Error())
		return
	}

	snap := Snapshot{Funding: funding, RSI: rsi, Price: price}
	act := Decide(s.pos, snap, s.params)

	fields := map[string]string{
		"price":   price.StringFixed(2),
		"funding": funding.String(),
		"rsi":     rsi.StringFixed(2),
		"pos":     sideName(s.pos),
		"dry_run": fmt.Sprintf("%v", s.dryRun),
		"action":  actionName(act),
		"size":    s.params.Size.String(),
	}

	switch act {
	case ActionNone:
		s.status("scanning",
			fmt.Sprintf("scan funding=%s rsi=%s pos=%s", funding.String(), rsi.StringFixed(2), sideName(s.pos)),
			fields)
	case ActionEnterLong:
		s.execute(pair, ActionEnterLong, "enter long (neg funding + oversold)", fields)
	case ActionEnterShort:
		s.execute(pair, ActionEnterShort, "enter short (pos funding + overbought)", fields)
	case ActionExitLong:
		s.execute(pair, ActionExitLong, "exit long", fields)
	case ActionExitShort:
		s.execute(pair, ActionExitShort, "exit short", fields)
	}
}

func (s *FundingTilt) execute(pair portfolio.Pair, act Action, reason string, fields map[string]string) {
	qty := s.params.Size
	max := numerical.NewFromFloat(HardMaxSizeBTC)
	if qty.GreaterThan(max) {
		qty = max
	}

	s.k.Log().Info("funding-tilt signal",
		"action", actionName(act),
		"reason", reason,
		"dry_run", s.dryRun,
		"qty", qty.String(),
	)

	if s.dryRun {
		s.applyLocal(act)
		fields["pos"] = sideName(s.pos)
		s.status("in_trade", "DRY "+reason, fields)
		return
	}

	if s.posUnknown || !s.posSeeded {
		s.k.Log().Info("refusing live emit until position seeded")
		s.status("error", "live blocked: position not seeded", fields)
		return
	}

	var buildErr error
	switch act {
	case ActionEnterLong, ActionExitShort:
		sig, err := s.k.Perp().Signal(s.GetName()).Buy(pair, s.exchange, qty).Build()
		if err != nil {
			buildErr = err
			break
		}
		s.k.Perp().Emit(sig)
		s.applyLocal(act)
	case ActionEnterShort:
		sig, err := s.k.Perp().Signal(s.GetName()).SellShort(pair, s.exchange, qty).Build()
		if err != nil {
			buildErr = err
			break
		}
		s.k.Perp().Emit(sig)
		s.applyLocal(act)
	case ActionExitLong:
		sig, err := s.k.Perp().Signal(s.GetName()).Sell(pair, s.exchange, qty).Build()
		if err != nil {
			buildErr = err
			break
		}
		s.k.Perp().Emit(sig)
		s.applyLocal(act)
	}

	if buildErr != nil {
		s.k.Log().Info("order build failed", "err", buildErr.Error())
		s.status("error", buildErr.Error(), fields)
		return
	}
	fields["pos"] = sideName(s.pos)
	s.status("in_trade", reason, fields)
}

func (s *FundingTilt) applyLocal(act Action) {
	switch act {
	case ActionEnterLong:
		s.pos = Long
	case ActionEnterShort:
		s.pos = Short
	case ActionExitLong, ActionExitShort:
		s.pos = Flat
	}
}

func (s *FundingTilt) status(phase strategy.StatusPhase, summary string, fields map[string]string) {
	s.EmitStatus(strategy.StrategyStatus{
		Phase:   phase,
		Summary: summary,
		Fields:  fields,
		At:      time.Now(),
	})
}

func sideName(s Side) string {
	switch s {
	case Long:
		return "long"
	case Short:
		return "short"
	default:
		return "flat"
	}
}

func actionName(a Action) string {
	switch a {
	case ActionEnterLong:
		return "enter_long"
	case ActionEnterShort:
		return "enter_short"
	case ActionExitLong:
		return "exit_long"
	case ActionExitShort:
		return "exit_short"
	default:
		return "none"
	}
}
