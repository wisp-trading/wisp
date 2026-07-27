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
	m.err = nil
	m.successMsg = ""
	m.reload(true)
	return nil
}

// reload re-reads ~/.wisp/connectors.yml and rebuilds configured/available lists.
// clearErr: true on Init / explicit refresh; false on soft View refresh (keep action errors).
func (m *ConnectorListModel) reload(jumpToAddIfEmpty bool) {
	m.reloadWithOpts(jumpToAddIfEmpty, true)
}

func (m *ConnectorListModel) softReload() {
	m.reloadWithOpts(false, false)
}

func (m *ConnectorListModel) reloadWithOpts(jumpToAddIfEmpty, clearErr bool) {
	connectorList, err := m.config.GetConnectors()
	if err != nil {
		// Still show available exchanges so user can add the first key.
		m.err = err
		m.configured = []config.Connector{}
	} else {
		m.configured = connectorList
		if clearErr {
			m.err = nil
		}
	}

	configuredMap := make(map[string]bool)
	for _, c := range m.configured {
		configuredMap[c.Name] = true
	}

	m.available = []string{}
	for _, name := range types.AllConnectors {
		if !configuredMap[string(name)] {
			m.available = append(m.available, string(name))
		}
	}

	total := len(m.configured) + len(m.available)
	if total == 0 {
		m.cursor = 0
		m.inAvailableSection = false
		return
	}
	if m.cursor >= total {
		m.cursor = total - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	m.inAvailableSection = m.cursor >= len(m.configured)

	if jumpToAddIfEmpty && len(m.configured) == 0 && len(m.available) > 0 {
		m.cursor = 0
		m.inAvailableSection = true
	}
}

func (m *ConnectorListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Soft-refresh after form/delete close (list stays mounted under bubblon stack).
		prevSuccess := m.successMsg
		m.reload(false)
		// Keep flash messages until next non-refresh key; r sets its own.
		if msg.String() != "r" {
			m.successMsg = ""
		} else {
			m.successMsg = prevSuccess
		}

		switch msg.String() {
		case "q", "esc":
			// Deliberate leave only — not Backspace/Ctrl+X (form-cancel keys).
			return m, m.router.Back()
		case "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.inAvailableSection = m.cursor >= len(m.configured)
			}
		case "down", "j":
			totalItems := len(m.configured) + len(m.available)
			if m.cursor < totalItems-1 {
				m.cursor++
				m.inAvailableSection = m.cursor >= len(m.configured)
			}
		case "enter":
			if m.inAvailableSection {
				availableIndex := m.cursor - len(m.configured)
				if availableIndex >= 0 && availableIndex < len(m.available) {
					selectedExchange := m.available[availableIndex]
					createView := m.formFactory(selectedExchange, false)
					return m, bubblon.Open(createView)
				}
			} else if m.cursor < len(m.configured) {
				selectedConnectorName := m.configured[m.cursor].Name
				editView := m.formFactory(selectedConnectorName, true)
				return m, bubblon.Open(editView)
			}
		case "d":
			if !m.inAvailableSection && m.cursor < len(m.configured) {
				selectedConnectorName := m.configured[m.cursor].Name
				deleteView := m.deleteFactory(selectedConnectorName)
				return m, bubblon.Open(deleteView)
			}
		case " ", "t":
			// Toggle only works on configured connectors
			if !m.inAvailableSection && m.cursor < len(m.configured) {
				connectorName := m.configured[m.cursor].Name
				newState := !m.configured[m.cursor].Enabled
				if err := m.config.EnableConnector(connectorName, newState); err != nil {
					m.err = err
				} else {
					m.reload(false)
					if newState {
						m.successMsg = connectorName + " enabled"
					} else {
						m.successMsg = connectorName + " disabled"
					}
				}
			}
		case "r":
			m.reload(false)
			m.successMsg = "Refreshed"
		}
	}
	return m, nil
}

func (m *ConnectorListModel) View() string {
	// When form/delete pops, View re-renders before the next key — keep list current.
	// Soft: do not clear action errors (e.g. failed toggle) until the next key.
	m.softReload()

	var content strings.Builder

	title := ui.TitleStyle.Render("⚙️  Exchange keys")
	content.WriteString(title)
	content.WriteString("\n")
	content.WriteString(ui.MutedStyle.Render("Stored in ~/.wisp/connectors.yml — shared by all strategies"))
	content.WriteString("\n\n")

	if m.err != nil {
		errorBox := ui.ErrorBoxStyle.Width(68).Render("❌ " + m.err.Error())
		content.WriteString(errorBox)
		content.WriteString("\n\n")
	}

	if m.successMsg != "" {
		content.WriteString(ui.StatusReadyStyle.Render("✓ " + m.successMsg))
		content.WriteString("\n\n")
	}

	if len(m.configured) == 0 && m.err == nil {
		content.WriteString(ui.MutedStyle.Render("No keys yet. Select an exchange below and press Enter."))
		content.WriteString("\n\n")
	}

	sectionHeader := ui.SectionHeaderStyle.Render("📋 CONFIGURED")
	content.WriteString(sectionHeader)
	content.WriteString("\n\n")

	if len(m.configured) == 0 {
		content.WriteString(ui.MutedStyle.Render("   (none)"))
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
			"%s/%s  %s Add  %s Refresh  %s Back",
			ui.KeyHintStyle.Render("↑"),
			ui.KeyHintStyle.Render("↓"),
			ui.KeyHintStyle.Render("Enter"),
			ui.KeyHintStyle.Render("r"),
			ui.KeyHintStyle.Render("q/Esc"),
		)
	}

	return fmt.Sprintf(
		"%s/%s  %s Edit  %s Delete  %s Toggle  %s Refresh  %s Back",
		ui.KeyHintStyle.Render("↑"),
		ui.KeyHintStyle.Render("↓"),
		ui.KeyHintStyle.Render("Enter"),
		ui.KeyHintStyle.Render("d"),
		ui.KeyHintStyle.Render("Space"),
		ui.KeyHintStyle.Render("r"),
		ui.KeyHintStyle.Render("q/Esc"),
	)
}
