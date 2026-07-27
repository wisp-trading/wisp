package handlers

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wisp-trading/sdk/pkg/types/config"
	"github.com/wisp-trading/sdk/pkg/types/connector"
	"github.com/wisp-trading/wisp/internal/handlers/settings"
	"github.com/wisp-trading/wisp/internal/router"
)

// Lightweight walkthrough of model state machines (not full alt-screen).
// Documents expected keyboard paths for the three core jobs.

type wtRouter struct {
	backs int
	navs  []router.Route
}

func (w *wtRouter) Init() tea.Cmd                                  { return nil }
func (w *wtRouter) Update(msg tea.Msg) (tea.Model, tea.Cmd)        { return w, nil }
func (w *wtRouter) View() string                                   { return "" }
func (w *wtRouter) RegisterRoute(router.Route, router.ViewFactory) {}
func (w *wtRouter) NavigateTo(r router.Route) tea.Cmd {
	w.navs = append(w.navs, r)
	return nil
}
func (w *wtRouter) Back() tea.Cmd            { w.backs++; return nil }
func (w *wtRouter) SetInitialView(tea.Model) {}

type wtScaffold struct {
	created string
}

func (s *wtScaffold) CreateProject(name string) error {
	s.created = name
	return nil
}
func (s *wtScaffold) CreateProjectWithStrategy(name, _ string) error {
	return s.CreateProject(name)
}

func TestWalkthrough_MenuOpensCoreRoutes(t *testing.T) {
	r := &wtRouter{}
	m := mainMenuModel{
		choices:  []string{"Strategies", "Monitor", "Settings", "Create New Project", "Help"},
		router:   r,
		scaffold: &wtScaffold{},
	}
	// Strategies
	m.cursor = 0
	_, _ = m.activate()
	// Monitor
	m.cursor = 1
	_, _ = m.activate()
	// Settings
	m.cursor = 2
	_, _ = m.activate()
	if len(r.navs) != 3 {
		t.Fatalf("navs=%v", r.navs)
	}
	if r.navs[0] != router.RouteStrategyList || r.navs[2] != router.RouteSettingsList {
		t.Fatalf("unexpected routes %v", r.navs)
	}
}

func TestWalkthrough_SettingsListEscBack(t *testing.T) {
	r := &wtRouter{}
	// Use real list model constructor
	list := settings.NewSettingsListView(
		&emptyCfg{},
		&emptySvc{},
		r,
		func(string, bool) tea.Model { return nil },
		func(string) tea.Model { return nil },
	)
	// Backspace must NOT leave (form-cancel key; accidental exits)
	_, _ = list.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if r.backs != 0 {
		t.Fatalf("backspace should not leave settings, backs=%d", r.backs)
	}
	updated, _ := list.Update(tea.KeyMsg{Type: tea.KeyEsc})
	_ = updated
	if r.backs != 1 {
		t.Fatalf("esc should back out of settings, backs=%d", r.backs)
	}
	// q also
	r.backs = 0
	_, _ = list.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if r.backs != 1 {
		t.Fatalf("q should back, backs=%d", r.backs)
	}
}

type emptyCfg struct{}

func (emptyCfg) LoadSettings(string) (*config.Settings, error) { return &config.Settings{}, nil }
func (emptyCfg) GetConnectors() ([]config.Connector, error)    { return nil, nil }
func (emptyCfg) GetEnabledConnectors() ([]config.Connector, error) {
	return nil, nil
}
func (emptyCfg) SaveSettings(*config.Settings) error    { return nil }
func (emptyCfg) AddConnector(config.Connector) error    { return nil }
func (emptyCfg) UpdateConnector(config.Connector) error { return nil }
func (emptyCfg) RemoveConnector(string) error           { return nil }
func (emptyCfg) EnableConnector(string, bool) error     { return nil }

type emptySvc struct{}

func (emptySvc) GetMatchingConnectors() (map[connector.ExchangeName]config.Connector, error) {
	return nil, nil
}
func (emptySvc) ValidateConnectorConfig(connector.ExchangeName, config.Connector) error {
	return nil
}
func (emptySvc) MapToSDKConfig(config.Connector) (connector.Config, error) { return nil, nil }
func (emptySvc) GetConnectorConfigsForStrategy([]string) (map[connector.ExchangeName]connector.Config, error) {
	return nil, nil
}
func (emptySvc) GetRequiredCredentialFields(string) []string { return nil }
