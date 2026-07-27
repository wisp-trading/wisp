package main

import "github.com/wisp-trading/sdk/pkg/types/wisp/numerical"

// Side is the strategy's inventory stance.
type Side int

const (
	Flat Side = iota
	Long
	Short
)

// Snapshot is one evaluation of market inputs (pure — no I/O).
type Snapshot struct {
	Funding numerical.Decimal // HL current funding rate (hourly)
	RSI     numerical.Decimal
	Price   numerical.Decimal
}

// Params are tunable thresholds (from config.yml parameters when wired).
type Params struct {
	FundingEntry   numerical.Decimal // |funding| must exceed to enter
	FundingExit    numerical.Decimal // |funding| below this allows RSI-only exit
	RSIOverbought  numerical.Decimal
	RSIOversold    numerical.Decimal
	RSIMid         numerical.Decimal
	Size           numerical.Decimal
}

// Action is what the strategy wants to do this tick.
type Action int

const (
	ActionNone Action = iota
	ActionEnterLong
	ActionEnterShort
	ActionExitLong
	ActionExitShort
)

// Decide implements funding-tilt + RSI mean-reversion:
//
//	Crowded longs  (funding >> 0 + RSI high)  → short
//	Crowded shorts (funding << 0 + RSI low)   → long
//	Exit when RSI returns through mid, or funding collapses toward zero.
//
// This is the popular single-venue "funding as sentiment" play on HL perps
// (full cash-and-carry needs spot; we only have perp here).
func Decide(pos Side, s Snapshot, p Params) Action {
	switch pos {
	case Flat:
		// Fade extreme positive funding + overbought
		if s.Funding.GreaterThan(p.FundingEntry) && s.RSI.GreaterThanOrEqual(p.RSIOverbought) {
			return ActionEnterShort
		}
		// Fade extreme negative funding + oversold
		if s.Funding.LessThan(p.FundingEntry.Neg()) && s.RSI.LessThanOrEqual(p.RSIOversold) {
			return ActionEnterLong
		}
	case Long:
		// Take profit / neutralize when RSI recovered or funding no longer pays us
		if s.RSI.GreaterThanOrEqual(p.RSIMid) {
			return ActionExitLong
		}
		if s.Funding.GreaterThan(p.FundingExit) {
			// funding flipped back positive while we're long — stop paying
			return ActionExitLong
		}
	case Short:
		if s.RSI.LessThanOrEqual(p.RSIMid) {
			return ActionExitShort
		}
		if s.Funding.LessThan(p.FundingExit.Neg()) {
			return ActionExitShort
		}
	}
	return ActionNone
}

func DefaultParams() Params {
	return Params{
		FundingEntry:  numerical.NewFromFloat(0.00005),
		FundingExit:   numerical.NewFromFloat(0.00001),
		RSIOverbought: numerical.NewFromFloat(70),
		RSIOversold:   numerical.NewFromFloat(30),
		RSIMid:        numerical.NewFromFloat(50),
		Size:          numerical.NewFromFloat(0.001),
	}
}
