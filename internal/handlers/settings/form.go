package settings

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/donderom/bubblon"
	"github.com/wisp-trading/sdk/pkg/types/config"
	"github.com/wisp-trading/sdk/pkg/types/connector"
	"github.com/wisp-trading/wisp/internal/router"
	"github.com/wisp-trading/wisp/internal/ui"
)

// ConnectorFormModel: detail card + native Bubble Tea credential editor.
// Fields from ConnectorService.GetRequiredCredentialFields (NewConfig).
type ConnectorFormModel struct {
	connector     config.Connector
	config        config.Configuration
	connectorSvc  config.ConnectorService
	router        router.Router
	deleteFactory DeleteConfirmViewFactory
	isEditMode    bool
	err           error

	showingDetail bool
	confirmExit   bool
	confirmCursor int // 0 stay, 1 leave

	exchangeName string
	network      string
	enabled      bool
	fieldNames   []string
	inputs       []textinput.Model
	// focus: 0..n-1 inputs, n = enabled toggle, n+1 = Save button
	focus int
	width int
}

// NewConnectorFormView creates detail or add/edit credential screen.
func NewConnectorFormView(
	cfg config.Configuration,
	connectorSvc config.ConnectorService,
	r router.Router,
	deleteFactory DeleteConfirmViewFactory,
	connectorName string,
	isEdit bool,
) tea.Model {
	m := &ConnectorFormModel{
		config:        cfg,
		connectorSvc:  connectorSvc,
		router:        r,
		deleteFactory: deleteFactory,
		isEditMode:    isEdit,
		enabled:       true,
		network:       "mainnet",
		width:         72,
	}

	if isEdit && connectorName != "" {
		list, err := cfg.GetConnectors()
		if err != nil {
			m.err = err
			return m
		}
		for _, conn := range list {
			if conn.Name == connectorName {
				m.connector = conn
				m.exchangeName = conn.Name
				m.network = conn.Network
				if m.network == "" {
					m.network = "mainnet"
				}
				m.enabled = conn.Enabled
				break
			}
		}
		if m.exchangeName == "" {
			m.err = fmt.Errorf("connector '%s' not found", connectorName)
			return m
		}
		m.showingDetail = true
		return m
	}

	m.exchangeName = connectorName
	m.showingDetail = false
	m.buildInputs()
	return m
}

func (m *ConnectorFormModel) buildInputs() {
	m.fieldNames = m.connectorSvc.GetRequiredCredentialFields(m.exchangeName)
	m.inputs = make([]textinput.Model, 0, len(m.fieldNames))
	creds := m.connector.Credentials
	if creds == nil {
		creds = map[string]string{}
	}
	for _, name := range m.fieldNames {
		ti := textinput.New()
		ti.Placeholder = formatFieldName(name)
		ti.CharLimit = 512
		ti.Width = 48
		if isSecretField(name) {
			ti.EchoMode = textinput.EchoPassword
			ti.EchoCharacter = '•'
		}
		if v, ok := creds[name]; ok {
			ti.SetValue(v)
		}
		m.inputs = append(m.inputs, ti)
	}
	m.focus = 0
	m.blurAll()
	if len(m.inputs) > 0 {
		m.inputs[0].Focus()
	}
}

func (m *ConnectorFormModel) blurAll() {
	for i := range m.inputs {
		m.inputs[i].Blur()
	}
}

func (m *ConnectorFormModel) focusIndex(i int) {
	m.blurAll()
	m.focus = i
	if i >= 0 && i < len(m.inputs) {
		m.inputs[i].Focus()
	}
}

func (m *ConnectorFormModel) maxFocus() int {
	// fields + enabled + save
	return len(m.inputs) + 1
}

func isSecretField(name string) bool {
	n := strings.ToLower(name)
	return strings.Contains(n, "key") || strings.Contains(n, "secret") || strings.Contains(n, "password") || strings.Contains(n, "passphrase")
}

func formatFieldName(field string) string {
	parts := strings.Split(field, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, " ")
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (m *ConnectorFormModel) Init() tea.Cmd {
	if !m.showingDetail && len(m.inputs) > 0 {
		return textinput.Blink
	}
	return nil
}

func (m *ConnectorFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = msg.Width
		if m.width > 80 {
			m.width = 80
		}
		if m.width < 40 {
			m.width = 40
		}
	}

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

	if m.confirmExit {
		return m.updateConfirmExit(msg)
	}

	if m.showingDetail {
		return m.updateDetail(msg)
	}

	return m.updateEditor(msg)
}

func (m *ConnectorFormModel) updateDetail(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+x", "backspace":
			return m, m.router.Back()
		case "e", "enter":
			m.showingDetail = false
			m.buildInputs()
			return m, textinput.Blink
		case " ":
			m.connector.Enabled = !m.connector.Enabled
			if err := m.config.UpdateConnector(m.connector); err != nil {
				m.err = err
			}
			return m, nil
		case "d":
			return m, bubblon.Open(m.deleteFactory(m.connector.Name))
		}
	}
	return m, nil
}

func (m *ConnectorFormModel) updateEditor(msg tea.Msg) (tea.Model, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "ctrl+x", "esc":
			m.confirmExit = true
			m.confirmCursor = 0
			return m, nil
		case "ctrl+c":
			m.confirmExit = true
			m.confirmCursor = 0
			return m, nil
		case "tab", "down":
			next := m.focus + 1
			if next > m.maxFocus() {
				next = 0
			}
			m.focusIndex(next)
			return m, textinput.Blink
		case "shift+tab", "up":
			prev := m.focus - 1
			if prev < 0 {
				prev = m.maxFocus()
			}
			m.focusIndex(prev)
			return m, textinput.Blink
		case "enter":
			switch {
			case m.focus == len(m.inputs):
				m.enabled = !m.enabled
				return m, nil
			case m.focus == m.maxFocus():
				return m, m.trySave()
			case m.focus < len(m.inputs):
				next := m.focus + 1
				m.focusIndex(next)
				return m, textinput.Blink
			}
		case " ":
			if m.focus == len(m.inputs) {
				m.enabled = !m.enabled
				return m, nil
			}
			// else space goes to textinput
		}
	}

	// Forward to focused text input
	if m.focus >= 0 && m.focus < len(m.inputs) {
		var cmd tea.Cmd
		m.inputs[m.focus], cmd = m.inputs[m.focus].Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *ConnectorFormModel) trySave() tea.Cmd {
	creds := make(map[string]string, len(m.fieldNames))
	for i, name := range m.fieldNames {
		creds[name] = strings.TrimSpace(m.inputs[i].Value())
	}
	m.connector = config.Connector{
		Name:        m.exchangeName,
		Network:     m.network,
		Enabled:     m.enabled,
		Assets:      m.connector.Assets,
		Credentials: creds,
	}
	if err := m.saveConnector(); err != nil {
		m.err = fmt.Errorf("%w\n\nEsc: fix form  ·  Ctrl+X: leave without saving", err)
		return nil
	}
	return m.router.Back()
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
			m.confirmExit = false
			return m, textinput.Blink
		case "n", "N", "esc":
			m.confirmExit = false
			return m, textinput.Blink
		case "y", "Y", "ctrl+x":
			m.confirmExit = false
			return m, m.leaveForm()
		}
	}
	return m, nil
}

func (m *ConnectorFormModel) leaveForm() tea.Cmd {
	m.confirmExit = false
	m.err = nil
	if m.isEditMode {
		m.showingDetail = true
		m.inputs = nil
		m.fieldNames = nil
		return nil
	}
	return m.router.Back()
}

func (m *ConnectorFormModel) saveConnector() error {
	for _, key := range m.connectorSvc.GetRequiredCredentialFields(m.connector.Name) {
		if strings.TrimSpace(m.connector.Credentials[key]) == "" {
			return fmt.Errorf("credential '%s' cannot be empty", formatFieldName(key))
		}
	}
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
	base := m.renderEditor()
	if m.confirmExit {
		return m.renderConfirmExit(base)
	}
	return base
}

func (m *ConnectorFormModel) renderEditor() string {
	var b strings.Builder
	title := "Add exchange keys"
	if m.isEditMode {
		title = "Edit exchange keys"
	}
	b.WriteString(ui.TitleStyle.Render(fmt.Sprintf("✏️  %s — %s", title, m.exchangeName)))
	b.WriteString("\n")
	b.WriteString(ui.MutedStyle.Render("Saved to ~/.wisp/connectors.yml"))
	b.WriteString("\n\n")

	if len(m.fieldNames) == 0 {
		b.WriteString(ui.MutedStyle.Render("No credential fields for this exchange (not registered?)."))
		b.WriteString("\n\n")
	}

	for i, name := range m.fieldNames {
		label := formatFieldName(name)
		if m.focus == i {
			b.WriteString(ui.SelectedItemStyle.Render("▶ " + label))
		} else {
			b.WriteString(ui.LabelStyle.Render("  " + label))
		}
		b.WriteString("\n  ")
		b.WriteString(m.inputs[i].View())
		b.WriteString("\n\n")
	}

	// Enabled row
	enLabel := "Enabled"
	enVal := "No"
	if m.enabled {
		enVal = "Yes"
	}
	if m.focus == len(m.inputs) {
		b.WriteString(ui.SelectedItemStyle.Render(fmt.Sprintf("▶ %s: [%s]  (space/enter toggle)", enLabel, enVal)))
	} else {
		b.WriteString(ui.ItemStyle.Render(fmt.Sprintf("  %s: [%s]", enLabel, enVal)))
	}
	b.WriteString("\n\n")

	// Save
	if m.focus == m.maxFocus() {
		b.WriteString(ui.SelectedItemStyle.Render("▶ [ Save ]"))
	} else {
		b.WriteString(ui.ItemStyle.Render("  [ Save ]"))
	}
	b.WriteString("\n\n")
	b.WriteString(ui.MutedStyle.Render("↑↓/Tab fields  ·  Enter next/save  ·  Ctrl+X / Esc cancel"))
	return ui.MenuBoxStyle.Width(m.width).Render("\n" + b.String() + "\n")
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

func (m *ConnectorFormModel) renderDetailView() string {
	if m.connector.Name == "" {
		return ui.ErrorBoxStyle.Render("❌ No connector loaded\n\nPress q to go back")
	}
	var content strings.Builder
	content.WriteString(ui.TitleStyle.Render("⚙️  " + m.connector.Name))
	content.WriteString("\n\n")
	if m.connector.Enabled {
		content.WriteString(ui.StatusReadyStyle.Render("● ENABLED"))
	} else {
		content.WriteString(ui.StatusDisabledStyle.Render("○ DISABLED"))
	}
	content.WriteString("\n\n")

	var details strings.Builder
	details.WriteString(ui.LabelStyle.Render("Exchange:"))
	details.WriteString(" ")
	details.WriteString(ui.ValueStyle.Render(m.connector.Name))
	details.WriteString("\n\n")
	details.WriteString(ui.LabelStyle.Render("Network:"))
	details.WriteString(" ")
	nv := m.connector.Network
	if nv == "" {
		nv = "mainnet"
	}
	var networkStyle lipgloss.Style
	if nv == "testnet" {
		networkStyle = ui.NetworkBadgeWarningStyle.Bold(true)
	} else {
		networkStyle = ui.ValueStyle
	}
	details.WriteString(networkStyle.Render(nv))
	details.WriteString("\n\n")
	details.WriteString(ui.SectionHeaderStyle.Render("Credentials"))
	details.WriteString("\n\n")

	for _, fieldName := range m.connectorSvc.GetRequiredCredentialFields(m.connector.Name) {
		details.WriteString(ui.LabelStyle.Render(formatFieldName(fieldName) + ":"))
		details.WriteString(" ")
		if value, exists := m.connector.Credentials[fieldName]; exists && len(value) > 3 {
			if isSecretField(fieldName) {
				details.WriteString(ui.StatusReadyStyle.Render(value[:3] + strings.Repeat("•", minInt(len(value)-3, 20))))
			} else {
				details.WriteString(ui.StatusReadyStyle.Render(value))
			}
		} else if value, exists := m.connector.Credentials[fieldName]; exists && value != "" {
			details.WriteString(ui.StatusReadyStyle.Render("set"))
		} else {
			details.WriteString(ui.StatusDangerStyle.Render("Not set"))
		}
		details.WriteString("\n")
	}

	content.WriteString(ui.DetailBoxStyle.Render(details.String()))
	content.WriteString("\n\n")
	content.WriteString(ui.HelpStyle.Padding(0).Render(
		fmt.Sprintf("%s Edit  %s Toggle  %s Delete  %s Back",
			ui.KeyHintStyle.Render("e/↵"),
			ui.KeyHintStyle.Render("Space"),
			ui.KeyHintStyle.Render("d"),
			ui.KeyHintStyle.Render("q/Esc"),
		),
	))
	return content.String()
}
