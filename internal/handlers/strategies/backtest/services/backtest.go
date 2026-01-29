package services

import (
	"github.com/wisp-trading/sdk/pkg/types/config"
	"github.com/wisp-trading/wisp/internal/handlers/strategies/backtest/types"
)

// backtestService handles backtest operations
type backtestService struct{}

func NewBacktestService() types.BacktestService {
	return &backtestService{}
}

func (s *backtestService) RunInteractive() error {
	//cfg, err := interactive.InteractiveMode()
	//if err != nil {
	//	return err
	//}
	//return s.ExecuteBacktest(cfg)

	return nil
}

func (s *backtestService) ExecuteBacktest(cfg *config.Settings) error {
	// TODO: Implement actual backtest execution
	return nil
}
