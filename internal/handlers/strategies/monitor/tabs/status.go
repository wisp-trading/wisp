package tabs

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wisp-trading/sdk/pkg/types/monitoring"
	"github.com/wisp-trading/wisp/internal/ui"
)

// StatusModel shows live strategy status (current snapshot) and a scrollable log of past snapshots.
// It polls QueryStatus every 1 s for the live panel and QueryStatusLog every 5 s for the log.
type StatusModel struct {
	querier    monitoring.ViewQuerier
	instanceID string

	// Live panel data
	live    []monitoring.StrategyStatusView
	liveErr error

	// Log data
	log    []monitoring.StrategyStatusView
	logErr error

	// UI state
	loading        bool
	logCursor      int  // selected row in the log
	showFieldsPane bool // expand fields for selected log entry
}

// NewStatusModel creates the status tab model.
func NewStatusModel(querier monitoring.ViewQuerier, instanceID string) *StatusModel {
	return &StatusModel{
		querier:    querier,
		instanceID: instanceID,
		loading:    true,
	}
}

// ── messages ────────────────────────────────────────────────────────────────

type statusLiveMsg struct {
	views []monitoring.StrategyStatusView
	err   error
}

type statusLogMsg struct {
	views []monitoring.StrategyStatusView
	err   error
}

type statusLiveTickMsg time.Time
type statusLogTickMsg time.Time

// ── lifecycle ────────────────────────────────────────────────────────────────

func (m *StatusModel) Init() tea.Cmd {
	return tea.Batch(
		m.fetchLive(),
		m.fetchLog(),
		m.tickLive(),
		m.tickLog(),
	)
}

func (m *StatusModel) tickLive() tea.Cmd {
	return tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
		return statusLiveTickMsg(t)
	})
}

func (m *StatusModel) tickLog() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return statusLogTickMsg(t)
	})
}

func (m *StatusModel) fetchLive() tea.Cmd {
	instanceID := m.instanceID
	return func() tea.Msg {
		views, err := m.querier.QueryStatus(instanceID)
		return statusLiveMsg{views: views, err: err}
	}
}

func (m *StatusModel) fetchLog() tea.Cmd {
	instanceID := m.instanceID
	return func() tea.Msg {
		views, err := m.querier.QueryStatusLog(instanceID)
		return statusLogMsg{views: views, err: err}
	}
}

// ── update ────────────────────────────────────────────────────────────────

func (m *StatusModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case statusLiveMsg:
		m.loading = false
		m.liveErr = msg.err
		m.live = msg.views
		return m, nil

	case statusLogMsg:
		m.logErr = msg.err
		prevLen := len(m.log)
		m.log = msg.views
		// Auto-scroll: if cursor was at the bottom, keep it there
		if len(m.log) > 0 && m.logCursor == prevLen-1 {
			m.logCursor = len(m.log) - 1
		}
		// Also pin cursor to bottom if nothing was selected yet
		if prevLen == 0 && len(m.log) > 0 {
			m.logCursor = len(m.log) - 1
		}
		return m, nil

	case statusLiveTickMsg:
		return m, tea.Batch(m.fetchLive(), m.tickLive())

	case statusLogTickMsg:
		return m, tea.Batch(m.fetchLog(), m.tickLog())

	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			m.loading = true
			return m, tea.Batch(m.fetchLive(), m.fetchLog())

		case "up", "k":
			if m.logCursor > 0 {
				m.logCursor--
				m.showFieldsPane = false
			}
			return m, nil

		case "down", "j":
			if m.logCursor < len(m.log)-1 {
				m.logCursor++
				m.showFieldsPane = false
			}
			return m, nil

		case "enter", " ":
			// Toggle expanded fields for selected log entry
			if len(m.log) > 0 {
				m.showFieldsPane = !m.showFieldsPane
			}
			return m, nil

		case "G":
			// Jump to bottom (newest)
			if len(m.log) > 0 {
				m.logCursor = len(m.log) - 1
				m.showFieldsPane = false
			}
			return m, nil

		case "g":
			// Jump to top (oldest)
			m.logCursor = 0
			m.showFieldsPane = false
			return m, nil
		}
	}
	return m, nil
}

// ── view ────────────────────────────────────────────────────────────────────

func (m *StatusModel) View() string {
	var b strings.Builder

	// ── Live status panels ───────────────────────────────────────────────────
	b.WriteString(ui.StrategyNameStyle.Render("LIVE STATUS"))
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(ui.SubtitleStyle.Render("Loading status..."))
		b.WriteString("\n\n")
	} else if m.liveErr != nil {
		b.WriteString(ui.StatusErrorStyle.Render(fmt.Sprintf("Error: %v", m.liveErr)))
		b.WriteString("\n\n")
	} else if len(m.live) == 0 {
		b.WriteString(ui.SubtitleStyle.Render("waiting for status…"))
		b.WriteString("\n\n")
	} else {
		for _, sv := range m.live {
			b.WriteString(renderStatusPanel(sv))
			b.WriteString("\n")
		}
	}

	// ── History log ─────────────────────────────────────────────────────────
	b.WriteString(ui.StrategyNameStyle.Render("STATUS LOG"))
	b.WriteString("\n\n")

	if m.logErr != nil {
		b.WriteString(ui.StatusErrorStyle.Render(fmt.Sprintf("Log error: %v", m.logErr)))
		b.WriteString("\n")
	}

	if len(m.log) == 0 {
		b.WriteString(ui.SubtitleStyle.Render("No log entries yet"))
	} else {
		b.WriteString(m.renderLog())
	}

	b.WriteString("\n\n")
	helpParts := []string{
		ui.KeyHintStyle.Render("↑↓") + " navigate",
		ui.KeyHintStyle.Render("Enter") + " expand fields",
		ui.KeyHintStyle.Render("G") + " newest",
		ui.KeyHintStyle.Render("g") + " oldest",
		ui.KeyHintStyle.Render("R") + " refresh",
	}
	b.WriteString(ui.HelpStyle.Render(strings.Join(helpParts, " • ")))

	return b.String()
}

// renderStatusPanel renders a single live status panel box.
func renderStatusPanel(sv monitoring.StrategyStatusView) string {
	var b strings.Builder

	// Header line: "StrategyName ── phase"
	nameStr := ui.StrategyNameStyle.Render(sv.StrategyName)
	phaseStr := phaseStyle(sv.Phase).Render(sv.Phase)
	header := nameStr + "  " + phaseStr

	// Body
	if sv.Summary == "" && sv.Phase == "" {
		b.WriteString(ui.SubtitleStyle.Render("waiting for status…"))
	} else {
		b.WriteString(ui.TextStyle.Render(sv.Summary))
		if len(sv.Fields) > 0 {
			b.WriteString("\n\n")
			b.WriteString(renderFieldsTable(sv.Fields))
		}
	}

	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(phaseBorderColor(sv.Phase)).
		Padding(1, 2).
		Width(70)

	return panelStyle.Render(header + "\n" + b.String())
}

// renderLog renders the scrollable history list.
func (m *StatusModel) renderLog() string {
	var b strings.Builder

	// Fixed prefix width: "  HH:MM:SS  " (10) + phase column (13) + space = 25 chars
	// Summary gets the rest of the terminal width; we use a generous cap so long
	// error messages are never silently cut off.
	const timeCol = 10
	const phaseCol = 13
	const prefixWidth = timeCol + 1 + phaseCol + 1 // "HH:MM:SS  PHASE        "

	// Header row
	header := fmt.Sprintf("  %-*s %-*s %s", timeCol, "TIME", phaseCol, "PHASE", "SUMMARY")
	b.WriteString(tableHeaderStyle.Render(header))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(strings.Repeat("─", 80)))
	b.WriteString("\n")

	const visibleRows = 20
	start := 0
	if len(m.log) > visibleRows {
		start = len(m.log) - visibleRows
		if m.logCursor < start {
			start = m.logCursor
		}
		if m.logCursor > start+visibleRows-1 {
			start = m.logCursor - visibleRows + 1
		}
	}
	end := start + visibleRows
	if end > len(m.log) {
		end = len(m.log)
	}

	for i := start; i < end; i++ {
		entry := m.log[i]
		timeStr := entry.At.Format("15:04:05")
		phaseStr := phaseStyle(entry.Phase).Width(phaseCol).Render(entry.Phase)
		selected := i == m.logCursor

		if selected {
			// Selected row: always show the full summary, highlighted
			firstLine := fmt.Sprintf("%-*s %s %s", timeCol, timeStr, phaseStr, entry.Summary)
			b.WriteString(lipgloss.NewStyle().Foreground(ui.ColorPrimary).Bold(true).Render("▶ " + firstLine))
			b.WriteString("\n")

			// Fields pane inline when expanded
			if m.showFieldsPane && len(entry.Fields) > 0 {
				fieldsContent := renderFieldsTable(entry.Fields)
				b.WriteString(lipgloss.NewStyle().Foreground(ui.ColorMuted).PaddingLeft(prefixWidth + 2).Render(fieldsContent))
				b.WriteString("\n")
			}
		} else {
			// Unselected: show full summary — no truncation
			row := fmt.Sprintf("  %-*s %s %s", timeCol, timeStr, phaseStr, entry.Summary)
			b.WriteString(phaseStyle(entry.Phase).Faint(true).Render(row))
			b.WriteString("\n")
		}
	}

	// Scroll position indicator
	indicator := fmt.Sprintf("  [%d / %d]  ↑↓ scroll • Enter expand fields", m.logCursor+1, len(m.log))
	b.WriteString(ui.SubtitleStyle.Render(indicator))

	return b.String()
}

// renderFieldsTable renders Fields map as a compact two-column key/value table.
func renderFieldsTable(fields map[string]string) string {
	if len(fields) == 0 {
		return ""
	}

	// Sort keys for deterministic output
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var rows []string
	for i := 0; i < len(keys); i += 2 {
		k1 := keys[i]
		v1 := fields[k1]
		row := fmt.Sprintf("%s  %s",
			ui.LabelStyle.Render(k1),
			ui.ValueStyle.Render(v1),
		)
		if i+1 < len(keys) {
			k2 := keys[i+1]
			v2 := fields[k2]
			row += fmt.Sprintf("    %s  %s",
				ui.LabelStyle.Render(k2),
				ui.ValueStyle.Render(v2),
			)
		}
		rows = append(rows, row)
	}

	return strings.Join(rows, "\n")
}

// phaseStyle returns a lipgloss style based on the phase string.
// error → red, in_trade → green, scanning → yellow, else → muted.
func phaseStyle(phase string) lipgloss.Style {
	lower := strings.ToLower(phase)
	switch {
	case strings.Contains(lower, "error"):
		return ui.StatusDangerStyle
	case strings.Contains(lower, "in_trade"):
		return ui.StatusReadyStyle
	case strings.Contains(lower, "scanning"):
		return ui.StatusRunningStyle
	default:
		return ui.SubtitleStyle
	}
}

// phaseBorderColor returns a border color based on the phase string.
func phaseBorderColor(phase string) lipgloss.Color {
	lower := strings.ToLower(phase)
	switch {
	case strings.Contains(lower, "error"):
		return ui.ColorDanger
	case strings.Contains(lower, "in_trade"):
		return ui.ColorSuccess
	case strings.Contains(lower, "scanning"):
		return ui.ColorWarning
	default:
		return ui.ColorPrimary
	}
}
