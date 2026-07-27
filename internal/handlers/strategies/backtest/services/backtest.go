package services

import (
	"fmt"

	"github.com/wisp-trading/sdk/pkg/types/config"
	"github.com/wisp-trading/wisp/internal/handlers/strategies/backtest/types"
)

// backtestService — placeholder. Live strategies are the supported product path.
type backtestService struct{}

func NewBacktestService() types.BacktestService {
	return &backtestService{}
}

func (s *backtestService) RunInteractive() error {
	return fmt.Errorf(`backtest is not available in this build

Supported today:
  wisp                      # TUI: project · keys · Start Live · Monitor
  wisp init my-bot          # scaffold standalone strategy

Use live strategies via: Strategies → Start Live`)
}

func (s *backtestService) ExecuteBacktest(_ *config.Settings) error {
	return s.RunInteractive()
}
