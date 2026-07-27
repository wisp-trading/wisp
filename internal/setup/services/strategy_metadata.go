package services

import "strings"

// StrategyMetadata is the init template descriptor (folder name under strategies/).
// Scaffold always writes the same starter body; Name is only the directory label.
type StrategyMetadata struct {
	Name        string
	DisplayName string
	Description string
	Type        string
	Icon        string
	SDKExample  string // strategies/<name>/ folder
}

// FetchAvailableStrategies returns the only supported template: starter standalone.
// Remote/local SDK example scraping was removed (hardcoded paths + broken URLs).
func FetchAvailableStrategies() ([]StrategyMetadata, error) {
	return []StrategyMetadata{StarterTemplate()}, nil
}

// StarterTemplate is main.go + StartStandalone + Wait; keys via Settings.
func StarterTemplate() StrategyMetadata {
	return StrategyMetadata{
		Name:        "starter",
		DisplayName: "Starter (standalone)",
		Description: "main.go + StartStandalone + Wait — keys via Settings (~/.wisp/connectors.yml)",
		Type:        "starter",
		Icon:        "🚀",
		SDKExample:  "starter",
	}
}

// GetDefaultIcon returns a default icon for a strategy type label.
func GetDefaultIcon(strategyType string) string {
	iconMap := map[string]string{
		"starter":        "🚀",
		"momentum":       "📈",
		"mean_reversion": "📉",
		"arbitrage":      "💱",
		"grid":           "⚡",
		"grid_trading":   "⚡",
		"technical":      "📊",
		"volume":         "📦",
	}
	if icon, ok := iconMap[strings.ToLower(strategyType)]; ok {
		return icon
	}
	return "🎯"
}
