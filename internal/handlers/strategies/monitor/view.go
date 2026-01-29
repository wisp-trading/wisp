package monitor

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/wisp-trading/sdk/pkg/types/monitoring"
)

// MonitorViewFactory creates a new monitor view
type MonitorViewFactory func() tea.Model

// NewMonitorViewFactory creates the factory for monitor views
func NewMonitorViewFactory(querier monitoring.ViewQuerier) MonitorViewFactory {
	return func() tea.Model {
		return NewInstanceListModel(querier)
	}
}
