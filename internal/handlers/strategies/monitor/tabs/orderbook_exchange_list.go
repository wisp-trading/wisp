package tabs

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wisp-trading/sdk/pkg/types/connector"
	"github.com/wisp-trading/sdk/pkg/types/monitoring"
	"github.com/wisp-trading/wisp/internal/ui"
)

// ExchangeListModel displays a list of exchanges to select from
type ExchangeListModel struct {
	source        string
	querier       monitoring.ViewQuerier
	instanceID    string
	marketViews   *monitoring.MarketViews
	exchanges     []exchangeItem
	selectedIndex int
	loading       bool
	err           error
}

type exchangeItem struct {
	name       connector.ExchangeName
	marketType connector.MarketType
	count      int
}

// NewExchangeListModel creates a new exchange list model
func NewExchangeListModel(source string, querier monitoring.ViewQuerier, instanceID string) *ExchangeListModel {
	return &ExchangeListModel{
		source:        source,
		querier:       querier,
		instanceID:    instanceID,
		loading:       true,
		selectedIndex: 0,
	}
}

// Messages
type exchangeMarketsMsg struct {
	source  string
	markets *monitoring.MarketViews
	err     error
}

type selectExchangeMsg struct {
	source       string
	exchangeName connector.ExchangeName
	marketType   connector.MarketType
	marketViews  *monitoring.MarketViews
}

func (m *ExchangeListModel) Init() tea.Cmd {
	return m.fetchMarkets()
}

func (m *ExchangeListModel) fetchMarkets() tea.Cmd {
	source := m.source
	return func() tea.Msg {
		markets, err := m.querier.QueryMarkets(m.instanceID)
		return exchangeMarketsMsg{source: source, markets: markets, err: err}
	}
}

func (m *ExchangeListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case exchangeMarketsMsg:
		if msg.source != m.source {
			return m, nil
		}
		m.loading = false
		m.err = msg.err
		if msg.err == nil && msg.markets != nil {
			m.marketViews = msg.markets
			m.buildExchangeList()
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.selectedIndex > 0 {
				m.selectedIndex--
			}
			return m, nil

		case "down", "j":
			if m.selectedIndex < len(m.exchanges)-1 {
				m.selectedIndex++
			}
			return m, nil

		case "enter":
			if len(m.exchanges) > 0 {
				selected := m.exchanges[m.selectedIndex]
				return m, func() tea.Msg {
					return selectExchangeMsg{
						source:       m.source,
						exchangeName: selected.name,
						marketType:   selected.marketType,
						marketViews:  m.marketViews,
					}
				}
			}
			return m, nil

		case "r":
			m.loading = true
			m.err = nil
			m.exchanges = nil
			return m, m.fetchMarkets()
		}
	}
	return m, nil
}

func (m *ExchangeListModel) View() string {
	var b strings.Builder

	b.WriteString(ui.SectionHeaderStyle.Render("EXCHANGES"))
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(ui.SubtitleStyle.Render("Loading exchanges..."))
		return b.String()
	}

	if m.err != nil {
		b.WriteString(ui.ErrorBoxStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		b.WriteString("\n\n")
		b.WriteString(ui.HelpStyle.Render("[R] Retry"))
		return b.String()
	}

	if len(m.exchanges) == 0 {
		b.WriteString(ui.SubtitleStyle.Render("No exchanges configured"))
		b.WriteString("\n\n")
		b.WriteString(ui.HelpStyle.Render("[R] Refresh"))
		return b.String()
	}

	// Render exchange list
	for i, ex := range m.exchanges {
		var line string
		if i == m.selectedIndex {
			cursor := ui.SelectedItemStyle.Render("▶ ")
			badge := m.renderMarketTypeBadge(ex.marketType)
			line = fmt.Sprintf("%s%s  %s  %s",
				cursor,
				ui.SelectedItemStyle.Render(ex.name.String()),
				badge,
				ui.SubtitleStyle.Render(fmt.Sprintf(
					"%d %s",
					ex.count,
					m.pluralizeMarkets(ex.marketType, ex.count),
				)))
		} else {
			badge := m.renderMarketTypeBadge(ex.marketType)
			line = fmt.Sprintf("  %s  %s  %s",
				ui.ItemStyle.Render(ex.name.String()),
				badge,
				ui.SubtitleStyle.Render(fmt.Sprintf("%d %s", ex.count, m.pluralizeMarkets(ex.marketType, ex.count))))
		}
		b.WriteString(line)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(ui.HelpStyle.Render("[↑↓/jk] Navigate • [Enter] Select • [R] Refresh"))

	return b.String()
}

func (m *ExchangeListModel) buildExchangeList() {
	// Build a map of exchange -> market type -> count
	exchangeMap := make(map[connector.ExchangeName]map[connector.MarketType]int)

	// Process spot markets
	for _, spot := range m.marketViews.Spot {
		if exchangeMap[spot.Exchange] == nil {
			exchangeMap[spot.Exchange] = make(map[connector.MarketType]int)
		}
		exchangeMap[spot.Exchange][connector.MarketTypeSpot]++
	}

	// Process perp markets
	for _, perp := range m.marketViews.Perp {
		if exchangeMap[perp.Exchange] == nil {
			exchangeMap[perp.Exchange] = make(map[connector.MarketType]int)
		}
		exchangeMap[perp.Exchange][connector.MarketTypePerp]++
	}

	// Process prediction markets
	for _, pred := range m.marketViews.Prediction {
		if exchangeMap[pred.Exchange] == nil {
			exchangeMap[pred.Exchange] = make(map[connector.MarketType]int)
		}
		exchangeMap[pred.Exchange][connector.MarketTypePrediction]++
	}

	// Convert to list
	m.exchanges = []exchangeItem{}
	for exchange, marketTypes := range exchangeMap {
		for marketType, count := range marketTypes {
			m.exchanges = append(m.exchanges, exchangeItem{
				name:       exchange,
				marketType: marketType,
				count:      count,
			})
		}
	}
}

func (m *ExchangeListModel) renderMarketTypeBadge(marketType connector.MarketType) string {
	switch marketType {
	case connector.MarketTypeSpot:
		return ui.NetworkBadgeStyle.Render("[spot]")
	case connector.MarketTypePerp:
		return ui.NetworkBadgeWarningStyle.Render("[perp]")
	case connector.MarketTypePrediction:
		return ui.StatusRunningStyle.Render("[pred]")
	default:
		return ui.SubtitleStyle.Render("[?]")
	}
}

func (m *ExchangeListModel) pluralizeMarkets(marketType connector.MarketType, count int) string {
	switch marketType {
	case connector.MarketTypeSpot, connector.MarketTypePerp:
		if count == 1 {
			return "pair"
		}
		return "pairs"
	case connector.MarketTypePrediction:
		if count == 1 {
			return "market"
		}
		return "markets"
	default:
		return "items"
	}
}
