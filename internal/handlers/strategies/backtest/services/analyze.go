package services

import (
	"fmt"

	"github.com/wisp-trading/wisp/internal/handlers/strategies/backtest/types"
)

// analyzeService — placeholder until backtest results exist.
type analyzeService struct{}

func NewAnalyzeService() types.AnalyzeService {
	return &analyzeService{}
}

func (s *analyzeService) AnalyzeResults(_ string) error {
	return fmt.Errorf(`analyze is not available in this build (no backtest results pipeline yet)

Supported today: wisp → Strategies → Start Live · Monitor`)
}
