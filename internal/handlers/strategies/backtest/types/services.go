package types

import (
	"github.com/wisp-trading/sdk/pkg/types/config"
)

type AnalyzeService interface {
	AnalyzeResults(path string) error
}

type BacktestService interface {
	RunInteractive() error
	ExecuteBacktest(cfg *config.Settings) error
}
