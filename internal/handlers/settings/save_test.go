package settings

import (
	"fmt"
	"testing"

	"github.com/wisp-trading/sdk/pkg/types/config"
	"github.com/wisp-trading/sdk/pkg/types/connector"
)

// stubConfig is a minimal Configuration for save tests.
type stubConfig struct {
	added   *config.Connector
	updated *config.Connector
}

func (s *stubConfig) LoadSettings(string) (*config.Settings, error) { return &config.Settings{}, nil }
func (s *stubConfig) GetConnectors() ([]config.Connector, error)    { return nil, nil }
func (s *stubConfig) GetEnabledConnectors() ([]config.Connector, error) {
	return nil, nil
}
func (s *stubConfig) SaveSettings(*config.Settings) error { return nil }
func (s *stubConfig) AddConnector(c config.Connector) error {
	s.added = &c
	return nil
}
func (s *stubConfig) UpdateConnector(c config.Connector) error {
	s.updated = &c
	return nil
}
func (s *stubConfig) RemoveConnector(string) error       { return nil }
func (s *stubConfig) EnableConnector(string, bool) error { return nil }

type stubConnectorSvc struct {
	fields []string
	valid  error
}

func (s *stubConnectorSvc) GetMatchingConnectors() (map[connector.ExchangeName]config.Connector, error) {
	return nil, nil
}
func (s *stubConnectorSvc) ValidateConnectorConfig(connector.ExchangeName, config.Connector) error {
	return s.valid
}
func (s *stubConnectorSvc) MapToSDKConfig(config.Connector) (connector.Config, error) {
	return nil, nil
}
func (s *stubConnectorSvc) GetConnectorConfigsForStrategy([]string) (map[connector.ExchangeName]connector.Config, error) {
	return nil, nil
}
func (s *stubConnectorSvc) GetRequiredCredentialFields(string) []string { return s.fields }

func TestSaveConnector_RequiresDiscoveredFields(t *testing.T) {
	cfg := &stubConfig{}
	svc := &stubConnectorSvc{fields: []string{"private_key", "account_address"}}
	m := &ConnectorFormModel{
		config:       cfg,
		connectorSvc: svc,
		connector: config.Connector{
			Name:        "hyperliquid",
			Enabled:     true,
			Credentials: map[string]string{"private_key": "x"}, // missing account_address
		},
	}
	err := m.saveConnector()
	if err == nil {
		t.Fatal("expected missing field error")
	}
}

func TestSaveConnector_RunsValidate(t *testing.T) {
	cfg := &stubConfig{}
	svc := &stubConnectorSvc{
		fields: []string{"api_key", "api_secret"},
		valid:  fmt.Errorf("sdk says no"),
	}
	m := &ConnectorFormModel{
		config:       cfg,
		connectorSvc: svc,
		connector: config.Connector{
			Name:    "bybit",
			Enabled: true,
			Credentials: map[string]string{
				"api_key":    "k",
				"api_secret": "s",
			},
		},
	}
	err := m.saveConnector()
	if err == nil || err.Error() != "sdk says no" {
		t.Fatalf("expected validate error, got %v", err)
	}
}

func TestSaveConnector_SuccessAdd(t *testing.T) {
	cfg := &stubConfig{}
	svc := &stubConnectorSvc{fields: []string{"api_key"}}
	m := &ConnectorFormModel{
		config:       cfg,
		connectorSvc: svc,
		isEditMode:   false,
		connector: config.Connector{
			Name:        "bybit",
			Enabled:     true,
			Credentials: map[string]string{"api_key": "k"},
		},
	}
	if err := m.saveConnector(); err != nil {
		t.Fatal(err)
	}
	if cfg.added == nil || cfg.added.Name != "bybit" {
		t.Fatalf("not added: %+v", cfg.added)
	}
}
