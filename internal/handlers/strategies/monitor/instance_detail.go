package monitor

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wisp-trading/sdk/pkg/types/monitoring"
	"github.com/wisp-trading/wisp/internal/handlers/strategies/monitor/tabs"
	"github.com/wisp-trading/wisp/internal/ui"
)

// Tab represents a detail view tab
type Tab int

const (
	TabOverview Tab = iota
	TabPositions
	TabStatus
	TabMarketData
	TabPnL
	TabProfiling
)

var tabNames = []string{
	"Overview",
	"Positions",
	"Status",
	"Market Data",
	"PnL",
	"Profiling",
}

// instanceDetailModel shows detailed view of a single instance
type instanceDetailModel struct {
	ui.BaseModel // Embed for common key handling
	querier      monitoring.ViewQuerier
	instanceID   string
	activeTab    Tab
	width        int
	height       int

	// Tab models - each manages its own data
	overviewTab   *tabs.OverviewModel
	positionsTab  *tabs.PositionsModel
	statusTab     *tabs.StatusModel
	marketDataTab *tabs.MarketDataModel
	pnlTab        *tabs.PnLModel
	profilingTab  *tabs.ProfilingModel
}

// NewInstanceDetailModel creates a detail view for an instance
func NewInstanceDetailModel(querier monitoring.ViewQuerier, instanceID string) tea.Model {
	return &instanceDetailModel{
		BaseModel:     ui.BaseModel{IsRoot: false},
		querier:       querier,
		instanceID:    instanceID,
		activeTab:     TabOverview,
		overviewTab:   tabs.NewOverviewModel(querier, instanceID),
		positionsTab:  tabs.NewPositionsModel(querier, instanceID),
		statusTab:     tabs.NewStatusModel(querier, instanceID),
		marketDataTab: tabs.NewMarketDataModel(querier, instanceID),
		pnlTab:        tabs.NewPnLModel(querier, instanceID),
		profilingTab:  tabs.NewProfilingModel(querier, instanceID),
	}
}

func (m *instanceDetailModel) Init() tea.Cmd {
	// Initialize all tabs
	return tea.Batch(
		m.overviewTab.Init(),
		m.positionsTab.Init(),
		m.statusTab.Init(),
		m.marketDataTab.Init(),
		m.pnlTab.Init(),
		m.profilingTab.Init(),
	)
}

func (m *instanceDetailModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if handled, cmd := m.BaseModel.HandleCommonKeys(msg); handled {
			return m, cmd
		}

		switch msg.String() {
		case "tab":
			if m.activeTab < TabProfiling {
				m.activeTab++
			} else {
				m.activeTab = TabOverview
			}
			return m, nil

		case "shift+tab":
			if m.activeTab > TabOverview {
				m.activeTab--
			} else {
				m.activeTab = TabProfiling
			}
			return m, nil

		case "1":
			m.activeTab = TabOverview
			return m, nil
		case "2":
			m.activeTab = TabPositions
			return m, nil
		case "3":
			m.activeTab = TabStatus
			return m, nil
		case "4":
			m.activeTab = TabMarketData
			return m, nil
		case "5":
			m.activeTab = TabPnL
			return m, nil
		case "6":
			m.activeTab = TabProfiling
			return m, nil
		}

		// Forward key messages only to active tab
		var cmd tea.Cmd
		switch m.activeTab {
		case TabOverview:
			_, cmd = m.overviewTab.Update(msg)
		case TabPositions:
			_, cmd = m.positionsTab.Update(msg)
		case TabStatus:
			_, cmd = m.statusTab.Update(msg)
		case TabMarketData:
			_, cmd = m.marketDataTab.Update(msg)
		case TabPnL:
			_, cmd = m.pnlTab.Update(msg)
		case TabProfiling:
			_, cmd = m.profilingTab.Update(msg)
		}
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)
	}

	// Forward ALL other messages (ticks, data, etc.) to ALL tabs so each
	// tab's polling cycle continues even when it is not the active one.
	var cmd tea.Cmd

	_, cmd = m.overviewTab.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	_, cmd = m.positionsTab.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	_, cmd = m.statusTab.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	_, cmd = m.marketDataTab.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	_, cmd = m.pnlTab.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	_, cmd = m.profilingTab.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *instanceDetailModel) View() string {
	var b strings.Builder

	// Header
	b.WriteString(m.renderHeader())
	b.WriteString("\n")

	// Tabs
	b.WriteString(m.renderTabs())
	b.WriteString("\n\n")

	// Content from active tab
	switch m.activeTab {
	case TabOverview:
		b.WriteString(m.overviewTab.View())
	case TabPositions:
		b.WriteString(m.positionsTab.View())
	case TabStatus:
		b.WriteString(m.statusTab.View())
	case TabMarketData:
		b.WriteString(m.marketDataTab.View())
	case TabPnL:
		b.WriteString(m.pnlTab.View())
	case TabProfiling:
		b.WriteString(m.profilingTab.View())
	}

	// Help
	b.WriteString("\n\n")
	helpText := "[Tab/S-Tab] Switch Tab • [1-6] Jump • [R] Refresh • [Q] Back"
	switch m.activeTab {
	case TabStatus:
		helpText = "[Tab/S-Tab] Switch Tab • [↑↓] Scroll Log • [Enter] Expand Fields • [R] Refresh • [Q] Back"
	case TabMarketData:
		helpText = "[Tab/S-Tab] Switch Tab • [[ ]] Switch View • [←→] Navigate • [R] Refresh • [Q] Back"
	}
	b.WriteString(ui.HelpStyle.Render(helpText))

	return b.String()
}

func (m *instanceDetailModel) renderHeader() string {
	title := ui.TitleStyle.Render(strings.ToUpper(m.instanceID))
	return title
}

func (m *instanceDetailModel) renderTabs() string {
	var renderedTabs []string
	for i, name := range tabNames {
		if Tab(i) == m.activeTab {
			renderedTabs = append(renderedTabs, TabActiveStyle.Render(fmt.Sprintf("[%s]", name)))
		} else {
			renderedTabs = append(renderedTabs, TabStyle.Render(name))
		}
	}
	return strings.Join(renderedTabs, "  ")
}
