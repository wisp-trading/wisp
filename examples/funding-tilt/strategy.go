package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/wisp-trading/sdk/pkg/types/connector"
	"github.com/wisp-trading/sdk/pkg/types/portfolio"
	"github.com/wisp-trading/sdk/pkg/types/strategy"
	"github.com/wisp-trading/sdk/pkg/types/wisp"
)

// FundingTilt fades crowded Hyperliquid perp positioning using funding + RSI.
//
// Thesis (2026 crypto meta): positive funding + overbought RSI → shorts are paid
// by crowded longs; inverse for negative funding. Single-venue (no spot hedge).
//
// Default dry_run=true — logs decisions and EmitStatus only. Set FUNDING_TILT_LIVE=1
// to place orders via wisp.Perp().Emit (still needs keys in ~/.wisp/connectors.yml).
type FundingTilt struct {
	strategy.BaseStrategy
	k wisp.Wisp

	exchange connector.ExchangeName
	params   Params
	dryRun   bool
	pos      Side
	tick     time.Duration
	interval string
	limit    int
}

func NewFundingTilt(k wisp.Wisp) strategy.Strategy {
	s := &FundingTilt{
		k:        k,
		exchange: connector.ExchangeName("hyperliquid"),
		params:   DefaultParams(),
		dryRun:   true,
		pos:      Flat,
		tick:     30 * time.Second,
		interval: "15m",
		limit:    60,
	}
	if os.Getenv("FUNDING_TILT_LIVE") == "1" {
		s.dryRun = false
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
		"funding_entry", s.params.FundingEntry.String(),
	)

	// Brief warm-up for ingestors
	time.Sleep(2 * time.Second)

	t := time.NewTicker(s.tick)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			s.k.Log().Info("funding-tilt stopped")
			return
		case <-t.C:
			s.tickOnce(pair)
		}
	}
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
