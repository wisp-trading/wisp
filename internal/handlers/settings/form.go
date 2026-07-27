package settings

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/donderom/bubblon"
	"github.com/wisp-trading/connectors/pkg/connectors/types"
	"github.com/wisp-trading/sdk/pkg/types/config"
	"github.com/wisp-trading/sdk/pkg/types/connector"
	"github.com/wisp-trading/wisp/internal/router"
	"github.com/wisp-trading/wisp/internal/ui"
)

// ConnectorFormModel represents the connector detail/edit view
type ConnectorFormModel struct {
	form          *huh.Form
	connector     config.Connector
	config        config.Configuration
	connectorSvc  config.ConnectorService
	router        router.Router
	deleteFactory DeleteConfirmViewFactory
	isEditMode    bool
	originalName  string
	err           error

	// UI state
	showingDetail bool // true = show detail view, false = show edit form
	// confirmExit is a modal over the edit form (ctrl+x / esc).
	confirmExit   bool
	confirmCursor int // 0 = stay, 1 = leave

	// Form field values
	exchangeName       string
	network            string
	enabled            bool
	credentials        map[string]string
	credentialPointers map[string]*string // Pointers to actual form input values
	assets             []string
}

// NewConnectorFormView creates a new connector form view with Huh forms
func NewConnectorFormView(
	config config.Configuration,
	connectorSvc config.ConnectorService,
	r router.Router,
	deleteFactory DeleteConfirmViewFactory,
	connectorName string,
	isEdit bool,
) tea.Model {
	m := &ConnectorFormModel{
		config:        config,
		connectorSvc:  connectorSvc,
		router:        r,
		deleteFactory: deleteFactory,
		isEditMode:    isEdit,
		originalName:  connectorName,
		credentials:   make(map[string]string),
		enabled:       true,
	}

	if isEdit && connectorName != "" {
		// Load existing connector
		connectorList, err := config.GetConnectors()
		if err != nil {
			m.err = err
			return m
		}

		for _, conn := range connectorList {
			if conn.Name == connectorName {
				m.connector = conn
				m.exchangeName = conn.Name
				m.network = conn.Network
				m.enabled = conn.Enabled
				m.assets = conn.Assets
				m.credentials = conn.Credentials
				break
			}
		}

		if m.connector.Name == "" {
			m.err = fmt.Errorf("connector '%s' not found", connectorName)
			return m
		}

		// Validate we have required data before showing detail view
		if m.connector.Name == "" || m.exchangeName == "" {
			m.err = fmt.Errorf("invalid connector data")
			return m
		}

		// Show detail view first for editing
		m.showingDetail = true
	} else {
		// Adding new - set exchange name if provided (from list selection)
		if connectorName != "" {
			m.exchangeName = connectorName
		}

		// Always go to form for new connectors (never detail view)
		m.showingDetail = false
		m.form = m.buildForm()
	}

	return m
}

// buildForm creates the Huh form focused on credentials
func (m *ConnectorFormModel) buildForm() *huh.Form {
	var groups []*huh.Group

	// If no exchange name set, show selector (this should rarely happen)
	if m.exchangeName == "" {
		availableExchanges := types.AllConnectors
		exchangeOptions := make([]huh.Option[string], len(availableExchanges))
		for i, ex := range availableExchanges {
			exchangeOptions[i] = huh.NewOption(string(ex), string(ex))
		}

		groups = append(groups, huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select Exchange").
				Options(exchangeOptions...).
				Value(&m.exchangeName),
		))
	}

	// Field names from connector NewConfig() via existing GetRequiredCredentialFields.
	requiredFields := m.connectorSvc.GetRequiredCredentialFields(m.exchangeName)

	var credFields []huh.Field
	titleEmoji := "➕"
	titleText := "Add Connector"
	if m.isEditMode {
		titleEmoji = "✏️"
		titleText = "Edit Connector"
	}

	if len(requiredFields) == 0 {
		// Do not invent api_key/api_secret — empty means unregistered or no fields.
		m.credentialPointers = map[string]*string{}
		credFields = append(credFields,
			huh.NewNote().
				Title(fmt.Sprintf("%s  %s", titleEmoji, m.exchangeName)).
				Description("No credential fields discovered (connector missing or NewConfig empty)."),
		)
		return huh.NewForm(huh.NewGroup(credFields...)).WithTheme(huh.ThemeCharm())
	}

	credFields = append(credFields,
		huh.NewNote().
			Title(fmt.Sprintf("%s  %s", titleEmoji, m.exchangeName)).
			Description(titleText),
	)

	credentialValues := make(map[string]*string)

	for _, fieldName := range requiredFields {
		// Allocate a string pointer for this field
		fieldValue := ""

		// If editing, pre-fill with existing value
		if m.isEditMode && len(m.credentials) > 0 {
			if existing, exists := m.credentials[fieldName]; exists {
				fieldValue = existing
			}
		}

		// Store pointer to this field's value
		credentialValues[fieldName] = &fieldValue

		// Build description
		fieldDesc := fmt.Sprintf("Enter your %s", formatFieldName(fieldName))
		if m.isEditMode && len(m.credentials) > 0 {
			if existing, exists := m.credentials[fieldName]; exists && len(existing) > 3 {
				masked := existing[:3] + strings.Repeat("•", minInt(len(existing)-3, 20))
				fieldDesc = fmt.Sprintf("Current: %s", masked)
			}
		}

		// Determine echo mode (mask secrets/keys, show addresses plainly)
		echoMode := huh.EchoModeNormal
		if strings.Contains(strings.ToLower(fieldName), "key") ||
			strings.Contains(strings.ToLower(fieldName), "secret") {
			echoMode = huh.EchoModePassword
		}

		credFields = append(credFields,
			huh.NewInput().
				Title(formatFieldName(fieldName)).
				Description(fieldDesc).
				Placeholder("...").
				EchoMode(echoMode).
				Value(credentialValues[fieldName]),
		)
	}

	// After form completes, we'll copy values from credentialValues to m.credentials
	// Store the map so we can access it in Update
	m.credentialPointers = credentialValues

	// Only show enable toggle if editing (less prominent)
	if m.isEditMode {
		credFields = append(credFields,
			huh.NewConfirm().
				Title("Enabled?").
				Value(&m.enabled),
		)
	}

	groups = append(groups, huh.NewGroup(credFields...))

	// Set defaults
	if m.network == "" {
		m.network = "mainnet" // Default, but we don't ask about it
	}
	if !m.isEditMode {
		m.enabled = true // Default to enabled for new connectors
	}

	km := huh.NewDefaultKeyMap()
	// Esc / ctrl+x abort the form (default Quit is only ctrl+c).
	// Parent model intercepts these first for a confirm dialog when possible.
	km.Quit = key.NewBinding(
		key.WithKeys("ctrl+x", "ctrl+c", "esc"),
		key.WithHelp("ctrl+x", "cancel"),
	)
	return huh.NewForm(groups...).
		WithTheme(huh.ThemeCharm()).
		WithShowHelp(true).
		WithShowErrors(true).
		WithKeyMap(km)
}

// minInt returns the minimum of two ints
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// formatFieldName converts snake_case to Title Case for display
func formatFieldName(field string) string {
	// Replace underscores with spaces and capitalize each word
	parts := strings.Split(field, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, " ")
}

func (m *ConnectorFormModel) Init() tea.Cmd {
	if m.form != nil {
		return m.form.Init()
	}
	return nil
}

func (m *ConnectorFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Error banner
	if m.err != nil {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc", "enter":
				m.err = nil
				return m, nil
			case "q", "ctrl+x":
				return m, m.leaveForm()
			}
		}
		return m, nil
	}

	// Modal: confirm discard / leave edit form
	if m.confirmExit {
		return m.updateConfirmExit(msg)
	}

	// Detail card
	if m.showingDetail {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "q", "esc", "ctrl+x", "backspace":
				return m, m.router.Back()
			case "e", "enter":
				m.showingDetail = false
				m.form = m.buildForm()
				return m, m.form.Init()
			case " ":
				m.connector.Enabled = !m.connector.Enabled
				if err := m.config.UpdateConnector(m.connector); err != nil {
					m.err = err
					return m, nil
				}
				return m, nil
			case "d":
				deleteView := m.deleteFactory(m.connector.Name)
				return m, bubblon.Open(deleteView)
			}
		}
		return m, nil
	}

	// Edit form: intercept leave keys before huh swallows them
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "ctrl+x", "esc":
			m.confirmExit = true
			m.confirmCursor = 0 // default Stay
			return m, nil
		case "ctrl+c":
			// Same as leave — confirm first (don't hard-quit mid-edit)
			m.confirmExit = true
			m.confirmCursor = 0
			return m, nil
		}
	}

	if m.form == nil {
		return m, nil
	}

	var cmd tea.Cmd
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	if m.form.State == huh.StateCompleted {
		for fieldName, valuePtr := range m.credentialPointers {
			if valuePtr != nil {
				m.credentials[fieldName] = *valuePtr
			}
		}
		m.connector = config.Connector{
			Name:        m.exchangeName,
			Network:     m.network,
			Enabled:     m.enabled,
			Assets:      m.assets,
			Credentials: m.credentials,
		}
		if err := m.saveConnector(); err != nil {
			m.err = fmt.Errorf("%w\n\nEsc: fix form  ·  Ctrl+X: leave without saving", err)
			m.form.State = huh.StateNormal
			return m, nil
		}
		return m, m.router.Back()
	}

	// huh Quit (if it still fires) → same confirm
	if m.form.State == huh.StateAborted {
		m.form.State = huh.StateNormal
		m.confirmExit = true
		m.confirmCursor = 0
		return m, nil
	}

	return m, cmd
}

func (m *ConnectorFormModel) updateConfirmExit(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h", "up", "k":
			m.confirmCursor = 0
		case "right", "l", "down", "j", "tab":
			m.confirmCursor = 1
		case "enter", " ":
			if m.confirmCursor == 1 {
				m.confirmExit = false
				return m, m.leaveForm()
			}
			// Stay — resume form
			m.confirmExit = false
			return m, nil
		case "n", "N", "esc":
			m.confirmExit = false
			return m, nil
		case "y", "Y", "ctrl+x":
			m.confirmExit = false
			return m, m.leaveForm()
		}
	}
	return m, nil
}

// leaveForm exits the form: detail view if editing, list if adding.
func (m *ConnectorFormModel) leaveForm() tea.Cmd {
	m.confirmExit = false
	m.err = nil
	if m.isEditMode {
		m.showingDetail = true
		m.form = nil
		return nil
	}
	return m.router.Back()
}

func (m *ConnectorFormModel) saveConnector() error {
	// Required keys from the same discovery path as the form.
	for _, key := range m.connectorSvc.GetRequiredCredentialFields(m.connector.Name) {
		if strings.TrimSpace(m.connector.Credentials[key]) == "" {
			return fmt.Errorf("credential '%s' cannot be empty", formatFieldName(key))
		}
	}

	// Full connector validation (MapToSDKConfig + Config.Validate).
	if err := m.connectorSvc.ValidateConnectorConfig(
		connector.ExchangeName(m.connector.Name),
		m.connector,
	); err != nil {
		return err
	}

	if m.isEditMode {
		return m.config.UpdateConnector(m.connector)
	}
	return m.config.AddConnector(m.connector)
}

func (m *ConnectorFormModel) View() string {
	if m.err != nil {
		return ui.ErrorBoxStyle.Render("❌ "+m.err.Error()) + "\n\n" +
			ui.MutedStyle.Render("Esc fix  ·  Ctrl+X leave")
	}

	if m.showingDetail {
		return m.renderDetailView()
	}

	base := "Loading..."
	if m.form != nil {
		base = m.form.View()
	}
	base += "\n" + ui.MutedStyle.Render("Tab next field  ·  Enter submit  ·  Ctrl+X / Esc cancel")

	if m.confirmExit {
		return m.renderConfirmExit(base)
	}
	return base
}

func (m *ConnectorFormModel) renderConfirmExit(under string) string {
	var stay, leave string
	if m.confirmCursor == 0 {
		stay = ui.SelectedItemStyle.Render("[ Stay ]")
		leave = ui.MutedStyle.Render("  Leave  ")
	} else {
		stay = ui.MutedStyle.Render("  Stay  ")
		leave = ui.SelectedItemStyle.Render("[ Leave ]")
	}
	modal := ui.ErrorBoxStyle.Width(52).Render(
		"Leave without saving?\n\n" +
			"Unsaved credential edits will be discarded.\n\n" +
			stay + "   " + leave + "\n\n" +
			ui.MutedStyle.Render("←→ choose  ↵ confirm  y leave  n stay"),
	)
	return under + "\n\n" + modal
}

// renderDetailView shows a beautiful detail card for the connector
func (m *ConnectorFormModel) renderDetailView() string {
	// Guard: should never be here without a connector name
	if m.connector.Name == "" {
		errorBox := ui.ErrorBoxStyle.Render("❌ No connector loaded\n\nPress 'q' to go back")
		return errorBox
	}

	var content strings.Builder

	// Title
	title := ui.TitleStyle.Render("⚙️  " + m.connector.Name)
	content.WriteString(title)
	content.WriteString("\n\n")

	// Status badge
	var statusBadge string
	if m.connector.Enabled {
		statusBadge = ui.StatusReadyStyle.Render("● ENABLED")
	} else {
		statusBadge = ui.StatusDisabledStyle.Render("○ DISABLED")
	}
	content.WriteString(statusBadge)
	content.WriteString("\n\n")

	// Detail box
	var details strings.Builder

	// Exchange type
	details.WriteString(ui.LabelStyle.Render("Exchange:"))
	details.WriteString(" ")
	details.WriteString(ui.ValueStyle.Render(m.connector.Name))
	details.WriteString("\n\n")

	// Network
	details.WriteString(ui.LabelStyle.Render("Network:"))
	details.WriteString(" ")
	networkValue := m.connector.Network
	if networkValue == "" {
		networkValue = "mainnet"
	}
	var networkStyle lipgloss.Style
	if networkValue == "testnet" {
		networkStyle = ui.NetworkBadgeWarningStyle.Bold(true)
	} else {
		networkStyle = ui.ValueStyle
	}
	details.WriteString(networkStyle.Render(networkValue))
	details.WriteString("\n\n")

	// Credentials section
	details.WriteString(ui.SectionHeaderStyle.Render("Credentials"))
	details.WriteString("\n\n")

	// Same discovery path as the edit form (NewConfig JSON keys).
	requiredFields := m.connectorSvc.GetRequiredCredentialFields(m.connector.Name)

	// Show each credential field dynamically
	for _, fieldName := range requiredFields {
		fieldLabel := formatFieldName(fieldName) + ":"
		details.WriteString(ui.LabelStyle.Render(fieldLabel))
		details.WriteString(" ")

		if value, exists := m.connector.Credentials[fieldName]; exists && len(value) > 3 {
			// Mask private keys, show addresses plainly
			if strings.Contains(strings.ToLower(fieldName), "key") ||
				strings.Contains(strings.ToLower(fieldName), "secret") {
				masked := value[:3] + strings.Repeat("•", minInt(len(value)-3, 20))
				details.WriteString(ui.StatusReadyStyle.Render(masked))
			} else {
				// Show addresses/usernames plainly
				details.WriteString(ui.StatusReadyStyle.Render(value))
			}
		} else {
			details.WriteString(ui.StatusDangerStyle.Render("Not set"))
		}
		details.WriteString("\n")
	}

	content.WriteString(ui.DetailBoxStyle.Render(details.String()))
	content.WriteString("\n\n")

	// Help text
	help := fmt.Sprintf(
		"%s Edit  %s Toggle  %s Delete  %s Back",
		ui.KeyHintStyle.Render("e"),
		ui.KeyHintStyle.Render("Space"),
		ui.KeyHintStyle.Render("d"),
		ui.KeyHintStyle.Render("q"),
	)
	content.WriteString(ui.HelpStyle.Padding(0).Render(help))

	return content.String()
}
