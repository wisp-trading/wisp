package main

import (
	"context"
	"time"

	"github.com/wisp-trading/sdk/pkg/types/strategy"
	"github.com/wisp-trading/sdk/pkg/types/wisp"
)

// ReferenceStrategy is a minimal self-directed strategy for packaging demos.
// It does not place orders — it only proves the process lifecycle.
type ReferenceStrategy struct {
	strategy.BaseStrategy
	k wisp.Wisp
}

func NewReferenceStrategy(k wisp.Wisp) strategy.Strategy {
	s := &ReferenceStrategy{k: k}
	s.BaseStrategy = *strategy.NewBaseStrategy(strategy.BaseStrategyConfig{
		Name: "reference-standalone",
	})
	return s
}

func (s *ReferenceStrategy) Start(ctx context.Context) error {
	return s.StartWithRunner(ctx, s.run)
}

func (s *ReferenceStrategy) run(ctx context.Context) {
	s.k.Log().Info("reference-standalone: running (Ctrl+C or POST /shutdown to stop)")
	t := time.NewTicker(30 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			s.k.Log().Info("reference-standalone: context done")
			return
		case <-t.C:
			s.k.Log().Info("reference-standalone: heartbeat")
		}
	}
}
