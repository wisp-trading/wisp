package settings

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wisp-trading/sdk/pkg/types/config"
	"github.com/wisp-trading/sdk/pkg/types/connector"
	"github.com/wisp-trading/wisp/internal/router"
)

type countRouter struct {
	backs int
}

func (c *countRouter) Init() tea.Cmd                                  { return nil }
func (c *countRouter) Update(msg tea.Msg) (tea.Model, tea.Cmd)        { return c, nil }
func (c *countRouter) View() string                                   { return "" }
func (c *countRouter) RegisterRoute(router.Route, router.ViewFactory) {}
func (c *countRouter) NavigateTo(router.Route) tea.Cmd                { return nil }
func (c *countRouter) Back() tea.Cmd {
	c.backs++
	return nil
}
func (c *countRouter) SetInitialView(tea.Model) {}

type fieldSvc struct{}

func (fieldSvc) GetMatchingConnectors() (map[connector.ExchangeName]config.Connector, error) {
	return nil, nil
}
func (fieldSvc) ValidateConnectorConfig(connector.ExchangeName, config.Connector) error {
	return nil
}
func (fieldSvc) MapToSDKConfig(config.Connector) (connector.Config, error) { return nil, nil }
func (fieldSvc) GetConnectorConfigsForStrategy([]string) (map[connector.ExchangeName]connector.Config, error) {
	return nil, nil
}
func (fieldSvc) GetRequiredCredentialFields(string) []string {
	return []string{"private_key", "account_address"}
}

func TestFormCtrlXShowsConfirmThenLeave(t *testing.T) {
	r := &countRouter{}
	m := &ConnectorFormModel{
		config:       &stubConfig{},
		connectorSvc: fieldSvc{},
		router:       r,
		isEditMode:   true,
		exchangeName: "hyperliquid",
		connector: config.Connector{
			Name:    "hyperliquid",
			Enabled: true,
			Credentials: map[string]string{
				"private_key":     "x",
				"account_address": "y",
			},
		},
		showingDetail: false,
	}
	m.buildInputs()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlX})
	m = updated.(*ConnectorFormModel)
	if !m.confirmExit {
		t.Fatal("expected confirmExit after ctrl+x")
	}

	m.confirmCursor = 1
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(*ConnectorFormModel)
	if m.confirmExit {
		t.Fatal("confirm should close")
	}
	if !m.showingDetail {
		t.Fatal("edit mode leave should return to detail view")
	}
}

func TestFormConfirmStayKeepsForm(t *testing.T) {
	r := &countRouter{}
	m := &ConnectorFormModel{
		config:        &stubConfig{},
		connectorSvc:  fieldSvc{},
		router:        r,
		isEditMode:    true,
		exchangeName:  "hyperliquid",
		showingDetail: false,
		confirmExit:   true,
		confirmCursor: 0,
		connector:     config.Connector{Name: "hyperliquid"},
	}
	m.buildInputs()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(*ConnectorFormModel)
	if m.confirmExit {
		t.Fatal("stay should dismiss confirm")
	}
	if m.showingDetail {
		t.Fatal("should remain on form")
	}
}

func TestNewConnectorLeaveGoesBack(t *testing.T) {
	r := &countRouter{}
	m := &ConnectorFormModel{
		config:        &stubConfig{},
		connectorSvc:  fieldSvc{},
		router:        r,
		isEditMode:    false,
		exchangeName:  "hyperliquid",
		showingDetail: false,
		confirmExit:   true,
		confirmCursor: 1,
	}
	m.buildInputs()
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if r.backs != 1 {
		t.Fatalf("expected Back() once, got %d", r.backs)
	}
}

func TestTabMovesFocus(t *testing.T) {
	m := &ConnectorFormModel{
		config:       &stubConfig{},
		connectorSvc: fieldSvc{},
		router:       &countRouter{},
		exchangeName: "hyperliquid",
		isEditMode:   false,
	}
	m.buildInputs()
	if m.focus != 0 {
		t.Fatalf("focus=%d", m.focus)
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(*ConnectorFormModel)
	if m.focus != 1 {
		t.Fatalf("after tab focus=%d want 1", m.focus)
	}
}
