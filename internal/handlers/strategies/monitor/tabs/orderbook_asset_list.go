package tabs

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	predictionConnector "github.com/wisp-trading/sdk/pkg/markets/prediction/types/connector"
	"github.com/wisp-trading/sdk/pkg/types/connector"
	"github.com/wisp-trading/sdk/pkg/types/monitoring"
	"github.com/wisp-trading/sdk/pkg/types/portfolio"
	"github.com/wisp-trading/wisp/internal/ui"
)

// AssetListModel displays assets for selected exchange
type AssetListModel struct {
	source        string
	exchangeName  connector.ExchangeName
	marketType    connector.MarketType
	marketViews   *monitoring.MarketViews
	items         []assetItem
	selectedIndex int
}

type assetItem struct {
	displayName string
	pair        portfolio.Pair
	marketID    predictionConnector.MarketID
	slug        string
	outcomes    []monitoring.PredictionOutcomeView
}

type selectAssetMsg struct {
	source string
	item   assetItem
}

func NewAssetListModel(source string, exchangeName connector.ExchangeName, marketType connector.MarketType, marketViews *monitoring.MarketViews) *AssetListModel {
	m := &AssetListModel{
		source:        source,
		exchangeName:  exchangeName,
		marketType:    marketType,
		marketViews:   marketViews,
		selectedIndex: 0,
	}
	m.buildAssetList()
	return m
}

func (m *AssetListModel) Init() tea.Cmd {
	return nil
}

func (m *AssetListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.selectedIndex > 0 {
				m.selectedIndex--
			}
		case "down", "j":
			if m.selectedIndex < len(m.items)-1 {
				m.selectedIndex++
			}
		case "enter":
			if len(m.items) > 0 {
				source := m.source
				item := m.items[m.selectedIndex]
				return m, func() tea.Msg {
					return selectAssetMsg{source: source, item: item}
				}
			}
		case "esc":
			source := m.source
			return m, func() tea.Msg {
				return backToExchangeListMsg{source: source}
			}
		}
	}
	return m, nil
}

func (m *AssetListModel) View() string {
	var b strings.Builder
	title := fmt.Sprintf("%s — %s", strings.ToUpper(m.exchangeName.String()), m.marketType)
	b.WriteString(ui.SectionHeaderStyle.Render(title))
	b.WriteString("\n\n")

	if len(m.items) == 0 {
		b.WriteString(ui.SubtitleStyle.Render("No markets available"))
		b.WriteString("\n\n")
		b.WriteString(ui.HelpStyle.Render("[Esc] Back"))
		return b.String()
	}

	for i, item := range m.items {
		if i == m.selectedIndex {
			cursor := ui.SelectedItemStyle.Render("▶ ")
			b.WriteString(fmt.Sprintf("%s%s\n", cursor, ui.SelectedItemStyle.Render(item.displayName)))
		} else {
			b.WriteString(fmt.Sprintf("  %s\n", ui.ItemStyle.Render(item.displayName)))
		}
	}

	b.WriteString("\n")
	b.WriteString(ui.HelpStyle.Render("[↑↓/jk] Navigate • [Enter] Select • [Esc] Back"))
	return b.String()
}

func (m *AssetListModel) buildAssetList() {
	m.items = []assetItem{}
	switch m.marketType {
	case connector.MarketTypeSpot:
		for _, spot := range m.marketViews.Spot {
			if spot.Exchange == m.exchangeName {
				m.items = append(m.items, assetItem{
					displayName: spot.Pair.Symbol(),
					pair:        spot.Pair,
				})
			}
		}
	case connector.MarketTypePerp:
		for _, perp := range m.marketViews.Perp {
			if perp.Exchange == m.exchangeName {
				m.items = append(m.items, assetItem{
					displayName: perp.Pair.Symbol(),
					pair:        perp.Pair,
				})
			}
		}
	case connector.MarketTypePrediction:
		for _, pred := range m.marketViews.Prediction {
			if pred.Exchange == m.exchangeName {
				m.items = append(m.items, assetItem{
					displayName: pred.Slug,
					marketID:    pred.MarketID,
					slug:        pred.Slug,
					outcomes:    pred.Outcomes,
				})
			}
		}
	}
}
