package tabs

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wisp-trading/sdk/pkg/types/monitoring"
	"github.com/wisp-trading/wisp/internal/ui"
)

// marketDataSubTab represents the active sub-tab within Market Data
type marketDataSubTab int

const (
	marketDataOrderbook marketDataSubTab = iota
	marketDataKlines
	marketDataTrades
)

var marketDataSubTabNames = []string{"Orderbook", "Klines", "Trades"}

// MarketDataModel combines Orderbook, Klines, and Trades under one top-level tab.
// Sub-tab switching uses [ and ] so that ←→ is always free for content navigation
// (klines interval switching, drill-down list navigation, etc.).
type MarketDataModel struct {
	querier    monitoring.ViewQuerier
	instanceID string

	activeSubTab marketDataSubTab

	orderbookTab *OrderbookModel
	klinesTab    *KlinesModel
	tradesTab    *TradesModel
}

// NewMarketDataModel creates the Market Data tab.
func NewMarketDataModel(querier monitoring.ViewQuerier, instanceID string) *MarketDataModel {
	return &MarketDataModel{
		querier:      querier,
		instanceID:   instanceID,
		activeSubTab: marketDataOrderbook,
		orderbookTab: NewOrderbookModel(querier, instanceID),
		klinesTab:    NewKlinesModel(querier, instanceID),
		tradesTab:    NewTradesModel(querier, instanceID),
	}
}

func (m *MarketDataModel) Init() tea.Cmd {
	return tea.Batch(
		m.orderbookTab.Init(),
		m.klinesTab.Init(),
		m.tradesTab.Init(),
	)
}

func (m *MarketDataModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// [ and ] switch sub-tabs — no conflict with ←→ used by content views
		case "[":
			if m.activeSubTab > 0 {
				m.activeSubTab--
			}
			return m, nil
		case "]":
			if int(m.activeSubTab) < len(marketDataSubTabNames)-1 {
				m.activeSubTab++
			}
			return m, nil
		}

		// All other keys go to the active sub-tab only
		var cmd tea.Cmd
		switch m.activeSubTab {
		case marketDataOrderbook:
			_, cmd = m.orderbookTab.Update(msg)
		case marketDataKlines:
			_, cmd = m.klinesTab.Update(msg)
		case marketDataTrades:
			_, cmd = m.tradesTab.Update(msg)
		}
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)
	}

	// Ticks and data messages go to ALL sub-tabs so polling continues in the background
	var cmd tea.Cmd
	_, cmd = m.orderbookTab.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	_, cmd = m.klinesTab.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	_, cmd = m.tradesTab.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *MarketDataModel) View() string {
	var b strings.Builder

	b.WriteString(m.renderSubTabs())
	b.WriteString("\n\n")

	switch m.activeSubTab {
	case marketDataOrderbook:
		b.WriteString(m.orderbookTab.View())
	case marketDataKlines:
		b.WriteString(m.klinesTab.View())
	case marketDataTrades:
		b.WriteString(m.tradesTab.View())
	}

	return b.String()
}

func (m *MarketDataModel) renderSubTabs() string {
	var parts []string
	for i, name := range marketDataSubTabNames {
		if marketDataSubTab(i) == m.activeSubTab {
			parts = append(parts, marketDataSubTabActiveStyle.Render(fmt.Sprintf("[%s]", name)))
		} else {
			parts = append(parts, marketDataSubTabStyle.Render(name))
		}
	}
	hint := ui.SubtitleStyle.Render("  [ [ ] ] to switch view")
	return strings.Join(parts, "  ") + hint
}
