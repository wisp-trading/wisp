package handlers

import (
	"github.com/spf13/cobra"
	"github.com/wisp-trading/wisp/internal/handlers/strategies/backtest/types"
)

// analyzeHandler handles the analyze command
type analyzeHandler struct {
	analyzeService types.AnalyzeService
}

func NewAnalyzeHandler(analyzeService types.AnalyzeService) types.AnalyzeHandler {
	return &analyzeHandler{
		analyzeService: analyzeService,
	}
}

func (h *analyzeHandler) Handle(cmd *cobra.Command, args []string) error {
	resultsPath, _ := cmd.Flags().GetString("path")
	if resultsPath == "" {
		resultsPath = "./results"
	}

	return h.analyzeService.AnalyzeResults(resultsPath)
}
