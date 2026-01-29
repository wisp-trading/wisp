package settings

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/wisp-trading/sdk/pkg/types/config"
	"github.com/wisp-trading/wisp/internal/router"
	"go.uber.org/fx"
)

var Module = fx.Module("settings",
	fx.Provide(
		// TUI Views (factories)
		NewSettingsListViewFactory,
		NewConnectorFormViewFactory,
		NewDeleteConfirmViewFactory,
	),
)

// Factory types for DI
type SettingsListViewFactory func() tea.Model
type ConnectorFormViewFactory func(connectorName string, isEdit bool) tea.Model
type DeleteConfirmViewFactory func(connectorName string) tea.Model

// NewSettingsListViewFactory creates a factory function for the settings list view
func NewSettingsListViewFactory(
	config config.Configuration,
	connectorSvc config.ConnectorService,
	r router.Router,
	formFactory ConnectorFormViewFactory,
	deleteFactory DeleteConfirmViewFactory,
) SettingsListViewFactory {
	return func() tea.Model {
		return NewSettingsListView(config, connectorSvc, r, formFactory, deleteFactory)
	}
}

// NewConnectorFormViewFactory creates a factory function for the connector form view
func NewConnectorFormViewFactory(
	config config.Configuration,
	connectorSvc config.ConnectorService,
	r router.Router,
	deleteFactory DeleteConfirmViewFactory,
) ConnectorFormViewFactory {
	return func(connectorName string, isEdit bool) tea.Model {
		return NewConnectorFormView(config, connectorSvc, r, deleteFactory, connectorName, isEdit)
	}
}

// NewDeleteConfirmViewFactory creates a factory function for the delete confirmation view
func NewDeleteConfirmViewFactory(
	config config.Configuration,
	r router.Router,
) DeleteConfirmViewFactory {
	return func(connectorName string) tea.Model {
		return NewDeleteConfirmView(config, r, connectorName)
	}
}
