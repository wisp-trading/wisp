package main

import (
	"testing"

	"github.com/wisp-trading/sdk/pkg/types/wisp/numerical"
)

func TestDecideEnterShortOnCrowdedLongs(t *testing.T) {
	p := DefaultParams()
	s := Snapshot{
		Funding: numerical.NewFromFloat(0.0001), // high positive
		RSI:     numerical.NewFromFloat(75),
		Price:   numerical.NewFromFloat(100000),
	}
	if got := Decide(Flat, s, p); got != ActionEnterShort {
		t.Fatalf("want EnterShort, got %v", got)
	}
}

func TestDecideEnterLongOnCrowdedShorts(t *testing.T) {
	p := DefaultParams()
	s := Snapshot{
		Funding: numerical.NewFromFloat(-0.0001),
		RSI:     numerical.NewFromFloat(25),
		Price:   numerical.NewFromFloat(100000),
	}
	if got := Decide(Flat, s, p); got != ActionEnterLong {
		t.Fatalf("want EnterLong, got %v", got)
	}
}

func TestDecideNoTradeInMiddle(t *testing.T) {
	p := DefaultParams()
	s := Snapshot{
		Funding: numerical.NewFromFloat(0.00001),
		RSI:     numerical.NewFromFloat(55),
		Price:   numerical.NewFromFloat(100000),
	}
	if got := Decide(Flat, s, p); got != ActionNone {
		t.Fatalf("want None, got %v", got)
	}
}

func TestDecideExitLongOnRSIMid(t *testing.T) {
	p := DefaultParams()
	s := Snapshot{
		Funding: numerical.NewFromFloat(-0.00002),
		RSI:     numerical.NewFromFloat(52),
		Price:   numerical.NewFromFloat(100000),
	}
	if got := Decide(Long, s, p); got != ActionExitLong {
		t.Fatalf("want ExitLong, got %v", got)
	}
}

func TestDecideExitShortOnRSIMid(t *testing.T) {
	p := DefaultParams()
	s := Snapshot{
		Funding: numerical.NewFromFloat(0.00002),
		RSI:     numerical.NewFromFloat(48),
		Price:   numerical.NewFromFloat(100000),
	}
	if got := Decide(Short, s, p); got != ActionExitShort {
		t.Fatalf("want ExitShort, got %v", got)
	}
}

func TestDecideFundingAloneNotEnough(t *testing.T) {
	p := DefaultParams()
	// High funding but RSI not extreme — no entry
	s := Snapshot{
		Funding: numerical.NewFromFloat(0.001),
		RSI:     numerical.NewFromFloat(55),
		Price:   numerical.NewFromFloat(100000),
	}
	if got := Decide(Flat, s, p); got != ActionNone {
		t.Fatalf("want None without RSI extreme, got %v", got)
	}
}
