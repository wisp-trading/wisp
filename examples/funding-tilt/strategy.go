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
// Default dry_run=true. Live only when:
//   - control: WISP_RUN_MODE=live and not WISP_DRY_RUN (requires arm_live)
//   - CLI: FUNDING_TILT_LIVE=1 without WISP_DRY_RUN
//
// Live inventory is re-seeded from the exchange; emits await execution result.
type FundingTilt struct {
	strategy.BaseStrategy
	k wisptypes.Wisp

	exchange   connector.ExchangeName
	params     Params
	dryRun     bool
	pos        Side
	posSeeded  bool
	posUnknown bool
	inFlight   bool
	tick       time.Duration
	interval   string
	limit      int
	configDir  string
	maxDataAge time.Duration
}

// NewFundingTilt builds the strategy. configDir is used to load config.yml params.
func NewFundingTilt(k wisptypes.Wisp, configDir string) strategy.Strategy {
	params, tick, interval, limit, err := LoadParamsFromConfigDir(configDir)
	if err != nil {
		// Hard-fail is safer for live; still allow construct — Start logs and uses defaults.
		params = DefaultParams()
		tick = 30 * time.Second
		interval = "15m"
		limit = 60
	}

	s := &FundingTilt{
		k:          k,
		exchange:   connector.ExchangeName("hyperliquid"),
		params:     params,
		dryRun:     resolveDryRun(),
		pos:        Flat,
		tick:       tick,
		interval:   interval,
		limit:      limit,
		configDir:  configDir,
		maxDataAge: 90 * time.Second,
	}

	if err != nil {
		// Will log after logger available in run
		_ = err
	}

	if os.Getenv("FUNDING_TILT_SAFE") == "1" {
		s.params = clampParams(Params{
			FundingEntry:  numerical.NewFromFloat(0.5),
			FundingExit:   numerical.NewFromFloat(0.25),
			RSIOverbought: numerical.NewFromFloat(99.5),
			RSIOversold:   numerical.NewFromFloat(0.5),
			RSIMid:        numerical.NewFromFloat(50),
			Size:          numerical.NewFromFloat(0.001),
		})
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

	// Warm-up for ingestors + optional REST-less position feed
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
			// Reconcile inventory every tick when live (or until first seed).
			if !s.dryRun || !s.posSeeded {
				s.seedPosition(pair)
			}
			s.tickOnce(pair)
		}
	}
}

func (s *FundingTilt) seedPosition(pair portfolio.Pair) {
	pos, ok := s.k.Perp().Position(s.exchange, pair)
	if !ok || pos == nil {
		// No row after warm-up: treat as flat (HL often omits flat accounts).
		// Live used to wait forever — that bricked first entry on flat wallets.
		s.pos = Flat
		s.posSeeded = true
		s.posUnknown = false
		s.k.Log().Info("funding-tilt seed", "pos", "flat", "source", "no_position_row")
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

func (s *FundingTilt) exitQty(pair portfolio.Pair) numerical.Decimal {
	qty := s.params.Size
	max := numerical.NewFromFloat(HardMaxSizeBTC)
	if qty.GreaterThan(max) {
		qty = max
	}
	// Prefer exchange size so we don't overshoot / flip after partial fills.
	if pos, ok := s.k.Perp().Position(s.exchange, pair); ok && pos != nil {
		ex := pos.Size
		if ex.IsNegative() {
			ex = ex.Neg()
		}
		if !ex.IsZero() && ex.LessThan(qty) {
			return ex
		}
	}
	return qty
}

func (s *FundingTilt) enterQty() numerical.Decimal {
	qty := s.params.Size
	max := numerical.NewFromFloat(HardMaxSizeBTC)
	if qty.GreaterThan(max) {
		qty = max
	}
	return qty
}

func (s *FundingTilt) tickOnce(pair portfolio.Pair) {
	if s.inFlight {
		s.status("in_trade", "order in flight", map[string]string{"pos": sideName(s.pos)})
		return
	}

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

	// Stale kline gate: last bar open time should be recent enough for the interval.
	if s.maxDataAge > 0 && len(klines) > 0 {
		last := klines[len(klines)-1]
		if !last.OpenTime.IsZero() && time.Since(last.OpenTime) > s.maxDataAge+15*time.Minute {
			s.status("error", "stale klines — halt emit", map[string]string{
				"last_open": last.OpenTime.UTC().Format(time.RFC3339),
			})
			return
		}
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
	var qty numerical.Decimal
	switch act {
	case ActionExitLong, ActionExitShort:
		qty = s.exitQty(pair)
	default:
		qty = s.enterQty()
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

	s.inFlight = true
	defer func() { s.inFlight = false }()

	// Emit and await — only mutate local inventory on success.
	var buildErr error
	var emitOK bool
	var emitErr error

	switch act {
	case ActionEnterLong, ActionExitShort:
		sig, err := s.k.Perp().Signal(s.GetName()).Buy(pair, s.exchange, qty).Build()
		if err != nil {
			buildErr = err
			break
		}
		res, ok := s.k.Perp().Emit(sig).AwaitWithTimeout(45 * time.Second)
		if !ok {
			emitErr = fmt.Errorf("execution timeout")
		} else if !res.Success {
			emitErr = res.Error
			if emitErr == nil {
				emitErr = fmt.Errorf("execution failed")
			}
		} else {
			emitOK = true
		}
	case ActionEnterShort:
		sig, err := s.k.Perp().Signal(s.GetName()).SellShort(pair, s.exchange, qty).Build()
		if err != nil {
			buildErr = err
			break
		}
		res, ok := s.k.Perp().Emit(sig).AwaitWithTimeout(45 * time.Second)
		if !ok {
			emitErr = fmt.Errorf("execution timeout")
		} else if !res.Success {
			emitErr = res.Error
			if emitErr == nil {
				emitErr = fmt.Errorf("execution failed")
			}
		} else {
			emitOK = true
		}
	case ActionExitLong:
		sig, err := s.k.Perp().Signal(s.GetName()).Sell(pair, s.exchange, qty).Build()
		if err != nil {
			buildErr = err
			break
		}
		res, ok := s.k.Perp().Emit(sig).AwaitWithTimeout(45 * time.Second)
		if !ok {
			emitErr = fmt.Errorf("execution timeout")
		} else if !res.Success {
			emitErr = res.Error
			if emitErr == nil {
				emitErr = fmt.Errorf("execution failed")
			}
		} else {
			emitOK = true
		}
	}

	if buildErr != nil {
		s.k.Log().Info("order build failed", "err", buildErr.Error())
		s.status("error", buildErr.Error(), fields)
		return
	}
	if !emitOK {
		msg := "emit failed"
		if emitErr != nil {
			msg = emitErr.Error()
		}
		s.k.Log().Info("order emit failed — reseed", "err", msg)
		s.seedPosition(pair)
		fields["pos"] = sideName(s.pos)
		s.status("error", msg, fields)
		return
	}

	s.applyLocal(act)
	// Prefer exchange truth after fill.
	s.seedPosition(pair)
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
