package tabs

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wisp-trading/sdk/pkg/types/monitoring"
	"github.com/wisp-trading/wisp/internal/ui"
)

// AssetListModel displays assets for selected exchange
type AssetListModel struct {
	exchangeName  string
	marketType    string
	marketViews   *monitoring.MarketViews
	items         []assetItem
	selectedIndex int
}

type assetItem struct {
	displayName string
	pair        string
	marketID    string
	slug        string
	outcomes    []monitoring.PredictionOutcomeView
}

type selectAssetMsg struct {
	item assetItem
}

func NewAssetListModel(exchangeName, marketType string, marketViews *monitoring.MarketViews) *AssetListModel {
	m := &AssetListModel{
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
				return m, func() tea.Msg {
					return selectAssetMsg{item: m.items[m.selectedIndex]}
				}
			}
		case "esc":
			return m, func() tea.Msg {
				return backToExchangeListMsg{}
			}
		}
	}
	return m, nil
}

func (m *AssetListModel) View() string {
	var b strings.Builder
	title := fmt.Sprintf("%s — %s", strings.ToUpper(m.exchangeName), m.marketType)
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
	case "spot":
		for _, spot := range m.marketViews.Spot {
			if spot.Exchange == m.exchangeName {
				m.items = append(m.items, assetItem{
					displayName: spot.Pair,
					pair:        spot.Pair,
				})
			}
		}
	case "perp":
		for _, perp := range m.marketViews.Perp {
			if perp.Exchange == m.exchangeName {
				m.items = append(m.items, assetItem{
					displayName: perp.Pair,
					pair:        perp.Pair,
				})
			}
		}
	case "prediction":
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
