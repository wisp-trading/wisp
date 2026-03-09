package tabs

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/wisp-trading/sdk/pkg/types/connector"
	"github.com/wisp-trading/sdk/pkg/types/monitoring"
)

// KlinesModel is the coordinator for klines navigation: exchange → asset → chart
type KlinesModel struct {
	querier    monitoring.ViewQuerier
	instanceID string

	// Navigation state
	currentView string // "exchange", "asset", or "klines"

	// Sub-models
	exchangeList *ExchangeListModel
	assetList    *AssetListModel
	klinesView   *KlinesViewModel

	// State carried between views
	selectedExchange   connector.ExchangeName
	selectedMarketType connector.MarketType
	marketViews        *monitoring.MarketViews
}

// NewKlinesModel creates a new klines coordinator
func NewKlinesModel(querier monitoring.ViewQuerier, instanceID string) *KlinesModel {
	return &KlinesModel{
		querier:      querier,
		instanceID:   instanceID,
		currentView:  "exchange",
		exchangeList: NewExchangeListModel("klines", querier, instanceID),
	}
}

func (m *KlinesModel) Init() tea.Cmd {
	return m.exchangeList.Init()
}

// IsChartView returns true when the klines chart is being displayed,
// so the parent can route left/right to interval switching instead of tab navigation.
func (m *KlinesModel) IsChartView() bool {
	return m.currentView == "klines"
}

func (m *KlinesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case selectExchangeMsg:
		if msg.source != "klines" {
			return m, nil
		}
		m.selectedExchange = msg.exchangeName
		m.selectedMarketType = msg.marketType
		m.marketViews = msg.marketViews
		m.assetList = NewAssetListModel("klines", msg.exchangeName, msg.marketType, msg.marketViews)
		m.currentView = "asset"
		return m, m.assetList.Init()

	case selectAssetMsg:
		if msg.source != "klines" {
			return m, nil
		}
		m.klinesView = NewKlinesViewModel(
			m.querier,
			m.instanceID,
			m.selectedExchange,
			msg.item,
		)
		m.currentView = "klines"
		return m, m.klinesView.Init()

	case backToExchangeListMsg:
		if msg.source != "klines" {
			return m, nil
		}
		m.currentView = "exchange"
		m.assetList = nil
		m.klinesView = nil
		return m, m.exchangeList.fetchMarkets()
	}

	// Forward messages to active sub-model
	var cmd tea.Cmd
	switch m.currentView {
	case "exchange":
		if m.exchangeList != nil {
			_, cmd = m.exchangeList.Update(msg)
		}
	case "asset":
		if m.assetList != nil {
			_, cmd = m.assetList.Update(msg)
		}
	case "klines":
		if m.klinesView != nil {
			_, cmd = m.klinesView.Update(msg)
		}
	}

	return m, cmd
}

func (m *KlinesModel) View() string {
	switch m.currentView {
	case "exchange":
		if m.exchangeList != nil {
			return m.exchangeList.View()
		}
	case "asset":
		if m.assetList != nil {
			return m.assetList.View()
		}
	case "klines":
		if m.klinesView != nil {
			return m.klinesView.View()
		}
	}
	return "Loading..."
}
