package tabs

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/wisp-trading/sdk/pkg/types/monitoring"
)

// OrderbookModel is the coordinator for the three-level orderbook navigation
type OrderbookModel struct {
	querier    monitoring.ViewQuerier
	instanceID string

	// Navigation state
	currentView string // "exchange", "asset", or "orderbook"

	// Sub-models
	exchangeList  *ExchangeListModel
	assetList     *AssetListModel
	orderbookView *OrderbookViewModel

	// State carried between views
	selectedExchange   string
	selectedMarketType string
	marketViews        *monitoring.MarketViews
}

// NewOrderbookModel creates a new orderbook coordinator
func NewOrderbookModel(querier monitoring.ViewQuerier, instanceID string) *OrderbookModel {
	return &OrderbookModel{
		querier:      querier,
		instanceID:   instanceID,
		currentView:  "exchange",
		exchangeList: NewExchangeListModel(querier, instanceID),
	}
}

func (m *OrderbookModel) Init() tea.Cmd {
	return m.exchangeList.Init()
}

func (m *OrderbookModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle navigation messages
	switch msg := msg.(type) {
	case selectExchangeMsg:
		// Transition from exchange list to asset list
		m.selectedExchange = msg.exchangeName
		m.selectedMarketType = msg.marketType
		m.marketViews = msg.marketViews
		m.assetList = NewAssetListModel(msg.exchangeName, msg.marketType, msg.marketViews)
		m.currentView = "asset"
		return m, m.assetList.Init()

	case selectAssetMsg:
		// Transition from asset list to orderbook view
		m.orderbookView = NewOrderbookViewModel(m.querier, m.instanceID, m.selectedExchange, m.selectedMarketType, msg.item)
		m.currentView = "orderbook"
		return m, m.orderbookView.Init()

	case backToExchangeListMsg:
		// Go back to exchange list (works from both asset list and orderbook view)
		m.currentView = "exchange"
		m.assetList = nil
		m.orderbookView = nil
		// Re-fetch exchange list to ensure fresh data
		return m, m.exchangeList.fetchMarkets()
	}

	// Forward messages to active view
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
	case "orderbook":
		if m.orderbookView != nil {
			_, cmd = m.orderbookView.Update(msg)
		}
	}

	return m, cmd
}

func (m *OrderbookModel) View() string {
	switch m.currentView {
	case "exchange":
		if m.exchangeList != nil {
			return m.exchangeList.View()
		}
	case "asset":
		if m.assetList != nil {
			return m.assetList.View()
		}
	case "orderbook":
		if m.orderbookView != nil {
			return m.orderbookView.View()
		}
	}
	return "Loading..."
}
