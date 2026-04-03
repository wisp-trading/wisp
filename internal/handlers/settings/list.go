package settings

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/donderom/bubblon"
	"github.com/wisp-trading/connectors/pkg/connectors/types"
	"github.com/wisp-trading/sdk/pkg/types/config"
	"github.com/wisp-trading/wisp/internal/router"
	"github.com/wisp-trading/wisp/internal/ui"
)

// ConnectorListModel represents the settings list view
type ConnectorListModel struct {
	configured         []config.Connector // Already configured connectors
	available          []string           // Available but not configured
	cursor             int
	inAvailableSection bool // true if cursor is in "add new" section
	config             config.Configuration
	connectorSvc       config.ConnectorService
	router             router.Router
	formFactory        ConnectorFormViewFactory
	deleteFactory      DeleteConfirmViewFactory
	err                error
	successMsg         string
}

// NewSettingsListView creates a new settings list view
func NewSettingsListView(
	cfg config.Configuration,
	connectorSvc config.ConnectorService,
	r router.Router,
	formFactory ConnectorFormViewFactory,
	deleteFactory DeleteConfirmViewFactory,
) tea.Model {
	return &ConnectorListModel{
		config:        cfg,
		connectorSvc:  connectorSvc,
		router:        r,
		formFactory:   formFactory,
		deleteFactory: deleteFactory,
		configured:    []config.Connector{},
		available:     []string{},
	}
}

func (m *ConnectorListModel) Init() tea.Cmd {
	// Load configured connectors
	connectorList, err := m.config.GetConnectors()
	if err != nil {
		m.err = err
		return nil
	}
	m.configured = connectorList

	// Get available connectors from SDK
	allAvailable := types.AllConnectors

	// Filter out already configured ones
	configuredMap := make(map[string]bool)
	for _, c := range m.configured {
		configuredMap[c.Name] = true
	}

	m.available = []string{}
	for _, name := range allAvailable {
		if !configuredMap[string(name)] {
			m.available = append(m.available, string(name))
		}
	}

	return nil
}

func (m *ConnectorListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			// Navigate back to main menu
			return m, m.router.Back()
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				// Update section tracking
				m.inAvailableSection = m.cursor >= len(m.configured)
			}
		case "down", "j":
			totalItems := len(m.configured) + len(m.available)
			if m.cursor < totalItems-1 {
				m.cursor++
				// Update section tracking
				m.inAvailableSection = m.cursor >= len(m.configured)
			}
		case "enter":
			if m.inAvailableSection {
				// Add new connector from available list
				availableIndex := m.cursor - len(m.configured)
				if availableIndex >= 0 && availableIndex < len(m.available) {
					selectedExchange := m.available[availableIndex]
					createView := m.formFactory(selectedExchange, false)
					return m, bubblon.Open(createView)
				}
			} else {
				// Edit configured connector
				if m.cursor < len(m.configured) {
					selectedConnectorName := m.configured[m.cursor].Name
					editView := m.formFactory(selectedConnectorName, true)
					return m, bubblon.Open(editView)
				}
			}
		case "d":
			// Delete only works on configured connectors
			if !m.inAvailableSection && m.cursor < len(m.configured) {
				selectedConnectorName := m.configured[m.cursor].Name
				deleteView := m.deleteFactory(selectedConnectorName)
				return m, bubblon.Open(deleteView)
			}
		case " ":
			// Toggle only works on configured connectors
			if !m.inAvailableSection && m.cursor < len(m.configured) {
				connectorName := m.configured[m.cursor].Name
				newState := !m.configured[m.cursor].Enabled
				if err := m.config.EnableConnector(connectorName, newState); err != nil {
					m.err = err
				} else {
					// Reload - calls Init() which refreshes both lists
					m.Init()
					m.successMsg = "Connector updated"
				}
			}
		}
	}
	return m, nil
}

func (m *ConnectorListModel) View() string {
	var content strings.Builder

	// Title
	title := ui.TitleStyle.Render("⚙️  Connector Configuration")
	content.WriteString(title)
	content.WriteString("\n\n")

	// Error message if any
	if m.err != nil {
		errorBox := ui.ErrorBoxStyle.
			Width(68).
			Render("❌ " + m.err.Error())
		content.WriteString(errorBox)
		content.WriteString("\n\n")
	}

	// Success message if any
	if m.successMsg != "" {
		successMsg := ui.StatusReadyStyle.Render("✓ " + m.successMsg)
		content.WriteString(successMsg)
		content.WriteString("\n\n")
	}

	// Section 1: Configured Connectors
	sectionHeader := ui.SectionHeaderStyle.Render("📋 CONFIGURED CONNECTORS")
	content.WriteString(sectionHeader)
	content.WriteString("\n\n")

	if len(m.configured) == 0 {
		emptyMsg := ui.MutedStyle.Render("   No connectors configured yet")
		content.WriteString(emptyMsg)
		content.WriteString("\n")
	} else {
		for i, conn := range m.configured {
			isSelected := m.cursor == i && !m.inAvailableSection
			content.WriteString(m.renderConfiguredConnector(conn, isSelected))
		}
	}

	content.WriteString("\n")

	// Section 2: Add New Connector
	addHeader := ui.SectionHeaderStyle.Render("➕ ADD NEW CONNECTOR")
	content.WriteString(addHeader)
	content.WriteString("\n\n")

	if len(m.available) == 0 {
		emptyMsg := ui.MutedStyle.Render("   All available connectors are configured")
		content.WriteString(emptyMsg)
		content.WriteString("\n")
	} else {
		for i, name := range m.available {
			globalIndex := len(m.configured) + i
			isSelected := m.cursor == globalIndex
			content.WriteString(m.renderAvailableConnector(name, isSelected))
		}
	}

	content.WriteString("\n")

	// Help text (context-aware)
	helpText := m.getHelpText()
	help := ui.HelpStyle.Render(helpText)
	content.WriteString(help)

	return content.String()
}

func (m *ConnectorListModel) renderConfiguredConnector(conn config.Connector, selected bool) string {
	var itemStyle lipgloss.Style
	var nameStyle lipgloss.Style

	if selected {
		itemStyle = ui.StrategyItemSelectedStyle
		nameStyle = ui.StrategyNameSelectedStyle
	} else {
		itemStyle = ui.StrategyItemStyle
		nameStyle = ui.StrategyNameStyle
	}

	// Build content
	var content strings.Builder

	// Name with network badge
	name := nameStyle.Render(conn.Name)
	if conn.Network != "" {
		var networkBadge string
		if conn.Network == "testnet" {
			networkBadge = ui.NetworkBadgeWarningStyle.Render(" [" + conn.Network + "]")
		} else {
			networkBadge = ui.NetworkBadgeStyle.Render(" [" + conn.Network + "]")
		}
		name += networkBadge
	}
	content.WriteString(name)
	content.WriteString("\n")

	// Status indicator
	var statusText string
	if conn.Enabled {
		statusText = ui.StatusReadyStyle.Render("● ENABLED")
	} else {
		statusText = ui.StatusDisabledStyle.Render("○ DISABLED")
	}
	content.WriteString(statusText)

	return itemStyle.Render(content.String())
}

func (m *ConnectorListModel) renderAvailableConnector(name string, selected bool) string {
	cursor := "  "
	style := ui.MutedStyle.Italic(false)

	if selected {
		cursor = ui.SelectedItemStyle.Render("▶ ")
		style = ui.SelectedItemStyle
	}

	bullet := ui.SectionHeaderStyle.Bold(false).Render("• ")

	line := cursor + bullet + style.Render(name) + "\n"
	return line
}

func (m *ConnectorListModel) getHelpText() string {
	if m.inAvailableSection {
		return fmt.Sprintf(
			"%s/%s Navigate  %s Add Connector  %s Back",
			ui.KeyHintStyle.Render("↑"),
			ui.KeyHintStyle.Render("↓"),
			ui.KeyHintStyle.Render("Enter"),
			ui.KeyHintStyle.Render("q"),
		)
	}

	return fmt.Sprintf(
		"%s/%s Navigate  %s Edit  %s Delete  %s Toggle  %s Back",
		ui.KeyHintStyle.Render("↑"),
		ui.KeyHintStyle.Render("↓"),
		ui.KeyHintStyle.Render("Enter"),
		ui.KeyHintStyle.Render("d"),
		ui.KeyHintStyle.Render("Space"),
		ui.KeyHintStyle.Render("q"),
	)
}
