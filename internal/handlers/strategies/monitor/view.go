package monitor

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/wisp-trading/sdk/pkg/types/monitoring"
	"github.com/wisp-trading/wisp/pkg/live"
)

// MonitorViewFactory creates a new monitor view
type MonitorViewFactory func() tea.Model

// NewMonitorViewFactory creates the factory for monitor views.
// InstanceManager is required so Stop uses the supervisor contract
// (HTTP /shutdown → wait → SIGKILL) and updates instance state.
func NewMonitorViewFactory(
	querier monitoring.ViewQuerier,
	manager live.InstanceManager,
) MonitorViewFactory {
	return func() tea.Model {
		return NewInstanceListModel(querier, manager)
	}
}
