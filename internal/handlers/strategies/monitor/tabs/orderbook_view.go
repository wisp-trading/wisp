package tabs

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wisp-trading/sdk/pkg/types/connector"
	"github.com/wisp-trading/sdk/pkg/types/monitoring"
	"github.com/wisp-trading/wisp/internal/ui"
)

// OrderbookViewModel displays live orderbook data with outcome selection for prediction markets
type OrderbookViewModel struct {
	querier      monitoring.ViewQuerier
	instanceID   string
	exchangeName connector.ExchangeName
	marketType   connector.MarketType
	item         assetItem

	// Orderbook data
	orderbook *connector.OrderBook
	depth     int
	loading   bool
	err       error

	// Prediction market outcome selection
	selectedOutcomeIndex int

	// Live update tracking
	lastUpdate  time.Time
	updateCount int
	showPulse   bool
	lastBestBid float64
	lastBestAsk float64
}

func NewOrderbookViewModel(
	querier monitoring.ViewQuerier,
	instanceID string,
	exchangeName connector.ExchangeName,
	marketType connector.MarketType,
	item assetItem,
) *OrderbookViewModel {
	return &OrderbookViewModel{
		querier:              querier,
		instanceID:           instanceID,
		exchangeName:         exchangeName,
		marketType:           marketType,
		item:                 item,
		depth:                10,
		loading:              true,
		selectedOutcomeIndex: 0,
		updateCount:          0,
	}
}

type orderbookViewDataMsg struct {
	orderbook *connector.OrderBook
	err       error
}

type orderbookViewTickMsg time.Time
type orderbookViewPulseOffMsg struct{}

func (m *OrderbookViewModel) Init() tea.Cmd {
	return tea.Batch(m.fetchData(), m.tick())
}

func (m *OrderbookViewModel) tick() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return orderbookViewTickMsg(t)
	})
}

func (m *OrderbookViewModel) pulseOff() tea.Cmd {
	return tea.Tick(150*time.Millisecond, func(t time.Time) tea.Msg {
		return orderbookViewPulseOffMsg{}
	})
}

func (m *OrderbookViewModel) fetchData() tea.Cmd {
	return func() tea.Msg {
		var orderbook *connector.OrderBook
		var err error

		switch m.marketType {
		case connector.MarketTypeSpot, connector.MarketTypePerp:
			orderbook, err = m.querier.QueryOrderbook(m.instanceID, m.exchangeName, m.item.pair)
		case connector.MarketTypePrediction:
			if len(m.item.outcomes) == 0 {
				return orderbookViewDataMsg{err: fmt.Errorf("no outcomes available")}
			}
			outcome := m.item.outcomes[m.selectedOutcomeIndex]

			orderbook, err = m.querier.QueryPredictionOrderbook(m.instanceID, m.item.marketID, outcome.OutcomeID)
		}

		return orderbookViewDataMsg{orderbook: orderbook, err: err}
	}
}

func (m *OrderbookViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case orderbookViewDataMsg:
		m.loading = false
		m.err = msg.err
		if msg.err == nil && msg.orderbook != nil {
			changed := m.hasDataChanged(msg.orderbook)
			m.orderbook = msg.orderbook
			m.lastUpdate = time.Now()
			m.updateCount++

			if changed {
				m.showPulse = true
				if len(msg.orderbook.Bids) > 0 {
					m.lastBestBid, _ = msg.orderbook.Bids[0].Price.Float64()
				}
				if len(msg.orderbook.Asks) > 0 {
					m.lastBestAsk, _ = msg.orderbook.Asks[0].Price.Float64()
				}
				return m, m.pulseOff()
			}
		}
		return m, nil

	case orderbookViewPulseOffMsg:
		m.showPulse = false
		return m, nil

	case orderbookViewTickMsg:
		return m, tea.Batch(m.fetchData(), m.tick())

	case tea.KeyMsg:
		switch msg.String() {
		case "d":
			switch m.depth {
			case 5:
				m.depth = 10
			case 10:
				m.depth = 20
			default:
				m.depth = 5
			}
			return m, nil

		case "tab", "n":
			if m.marketType == "prediction" && len(m.item.outcomes) > 0 {
				m.selectedOutcomeIndex = (m.selectedOutcomeIndex + 1) % len(m.item.outcomes)
				m.loading = true
				m.orderbook = nil
				m.updateCount = 0
				return m, m.fetchData()
			}
			return m, nil

		case "shift+tab", "p":
			if m.marketType == "prediction" && len(m.item.outcomes) > 0 {
				m.selectedOutcomeIndex--
				if m.selectedOutcomeIndex < 0 {
					m.selectedOutcomeIndex = len(m.item.outcomes) - 1
				}
				m.loading = true
				m.orderbook = nil
				m.updateCount = 0
				return m, m.fetchData()
			}
			return m, nil

		case "esc":
			return m, func() tea.Msg {
				return backToExchangeListMsg{source: "orderbook"}
			}
		}
	}
	return m, nil
}

func (m *OrderbookViewModel) hasDataChanged(newOB *connector.OrderBook) bool {
	if m.orderbook == nil {
		return true
	}
	if len(newOB.Bids) > 0 && len(m.orderbook.Bids) > 0 {
		newBid, _ := newOB.Bids[0].Price.Float64()
		if newBid != m.lastBestBid {
			return true
		}
	}
	if len(newOB.Asks) > 0 && len(m.orderbook.Asks) > 0 {
		newAsk, _ := newOB.Asks[0].Price.Float64()
		if newAsk != m.lastBestAsk {
			return true
		}
	}
	return false
}

func (m *OrderbookViewModel) View() string {
	var b strings.Builder

	b.WriteString(m.renderHeader())
	b.WriteString("\n\n")

	if m.loading && m.orderbook == nil && m.err == nil {
		b.WriteString(ui.SubtitleStyle.Render("Loading orderbook..."))
		return b.String()
	}

	if m.orderbook == nil {
		if m.err != nil {
			b.WriteString(ui.SubtitleStyle.Render(fmt.Sprintf("No data: %v", m.err)))
		} else {
			b.WriteString(ui.SubtitleStyle.Render("No orderbook data available"))
		}
		b.WriteString("\n\n")
		b.WriteString(ui.HelpStyle.Render("[D] Depth • [Esc] Back"))
		return b.String()
	}

	b.WriteString(m.renderOrderbook())
	b.WriteString("\n\n")

	// Help text varies by market type
	if m.marketType == "prediction" {
		b.WriteString(ui.HelpStyle.Render("[Tab] Next Outcome • [D] Toggle Depth • [Esc] Back"))
	} else {
		b.WriteString(ui.HelpStyle.Render("[D] Toggle Depth • [Esc] Back"))
	}

	return b.String()
}

func (m *OrderbookViewModel) renderHeader() string {
	var header strings.Builder

	// Title with asset/market info
	var title string
	switch m.marketType {
	case connector.MarketTypeSpot, connector.MarketTypePerp:
		title = fmt.Sprintf("ORDERBOOK - %s @ %s", m.item.pair, m.exchangeName)
	case connector.MarketTypePrediction:
		title = fmt.Sprintf("ORDERBOOK - %s", m.item.slug)
	}
	header.WriteString(ui.StrategyNameStyle.Render(title))
	header.WriteString("  ")

	// Live indicator with pulse
	if m.showPulse {
		header.WriteString(ui.StatusRunningStyle.Render("◉ LIVE"))
	} else if !m.lastUpdate.IsZero() {
		header.WriteString(ui.StatusReadyStyle.Render("● LIVE"))
	}

	// Update stats
	if !m.lastUpdate.IsZero() {
		ago := time.Since(m.lastUpdate)
		if ago < time.Second {
			header.WriteString(ui.SubtitleStyle.Render("  <1s ago"))
		} else {
			header.WriteString(ui.SubtitleStyle.Render(fmt.Sprintf("  %ds ago", int(ago.Seconds()))))
		}
		header.WriteString(ui.SubtitleStyle.Render(fmt.Sprintf("  |  %d updates", m.updateCount)))
	}

	header.WriteString(ui.SubtitleStyle.Render(fmt.Sprintf("  |  Depth: %d", m.depth)))

	// Prediction market outcome selector
	if m.marketType == "prediction" && len(m.item.outcomes) > 0 {
		header.WriteString("\n\n")
		header.WriteString(ui.LabelStyle.Render("Outcome: "))
		for i, outcome := range m.item.outcomes {
			if i == m.selectedOutcomeIndex {
				header.WriteString(ui.SelectedItemStyle.Render(fmt.Sprintf("[%s ▶]", outcome.Name)))
			} else {
				header.WriteString(ui.SubtitleStyle.Render(fmt.Sprintf("  [%s]", outcome.Name)))
			}
			header.WriteString("  ")
		}
	}

	return header.String()
}

func (m *OrderbookViewModel) renderOrderbook() string {
	var b strings.Builder

	maxQty := m.calculateMaxQuantity()
	if maxQty == 0 {
		maxQty = 1
	}

	barWidth := 30
	isPrediction := m.marketType == "prediction"

	// Use pulse style for the labels if we just updated
	currentAskStyle := lipgloss.NewStyle().Foreground(ui.ColorDanger)
	currentBidStyle := lipgloss.NewStyle().Foreground(ui.ColorSuccess)
	if m.showPulse {
		currentAskStyle = ui.StatusRunningStyle
		currentBidStyle = ui.StatusRunningStyle
	}

	// Asks header
	b.WriteString(currentAskStyle.Render("                              ASKS"))
	b.WriteString("\n")

	asksToShow := m.depth
	if asksToShow > len(m.orderbook.Asks) {
		asksToShow = len(m.orderbook.Asks)
	}

	for i := asksToShow - 1; i >= 0; i-- {
		level := m.orderbook.Asks[i]
		price, _ := level.Price.Float64()
		qty, _ := level.Quantity.Float64()

		bar := renderDepthBar(qty, maxQty, barWidth)
		if isPrediction {
			row := fmt.Sprintf("  %12.4f  %s  %8.4f", price, currentAskStyle.Render(bar), qty)
			b.WriteString(row)
		} else {
			row := fmt.Sprintf("  %12.2f  %s  %8.4f", price, currentAskStyle.Render(bar), qty)
			b.WriteString(row)
		}
		b.WriteString("\n")
	}

	// Spread
	if len(m.orderbook.Asks) > 0 && len(m.orderbook.Bids) > 0 {
		bestAsk, _ := m.orderbook.Asks[0].Price.Float64()
		bestBid, _ := m.orderbook.Bids[0].Price.Float64()
		spread := bestAsk - bestBid
		spreadPct := (spread / bestBid) * 100

		var spreadLine string
		if isPrediction {
			spreadLine = fmt.Sprintf("  ──────────── SPREAD: %.4f pts (%.3f%%) ────────────", spread, spreadPct)
		} else {
			spreadLine = fmt.Sprintf("  ──────────── SPREAD: $%.2f (%.3f%%) ────────────", spread, spreadPct)
		}
		b.WriteString(ui.SubtitleStyle.Render(spreadLine))
		b.WriteString("\n")
	}

	// Bids
	bidsToShow := m.depth
	if bidsToShow > len(m.orderbook.Bids) {
		bidsToShow = len(m.orderbook.Bids)
	}

	for i := 0; i < bidsToShow; i++ {
		level := m.orderbook.Bids[i]
		price, _ := level.Price.Float64()
		qty, _ := level.Quantity.Float64()

		bar := renderDepthBar(qty, maxQty, barWidth)
		if isPrediction {
			row := fmt.Sprintf("  %12.4f  %s  %8.4f", price, currentBidStyle.Render(bar), qty)
			b.WriteString(row)
		} else {
			row := fmt.Sprintf("  %12.2f  %s  %8.4f", price, currentBidStyle.Render(bar), qty)
			b.WriteString(row)
		}
		b.WriteString("\n")
	}

	// Bids footer
	b.WriteString(currentBidStyle.Render("                              BIDS"))
	b.WriteString("\n\n")

	// Mid price info
	if len(m.orderbook.Asks) > 0 && len(m.orderbook.Bids) > 0 {
		bestAsk, _ := m.orderbook.Asks[0].Price.Float64()
		bestBid, _ := m.orderbook.Bids[0].Price.Float64()
		midPrice := (bestAsk + bestBid) / 2

		if isPrediction {
			fmt.Fprintf(&b, "Mid: %.4f   Bid: %.4f   Ask: %.4f",
				midPrice, bestBid, bestAsk)
		} else {
			fmt.Fprintf(&b, "Mid: $%.2f   Bid: $%.2f   Ask: $%.2f",
				midPrice, bestBid, bestAsk)
		}
	}

	return b.String()
}

func (m *OrderbookViewModel) calculateMaxQuantity() float64 {
	maxQty := 0.0
	for i := 0; i < m.depth && i < len(m.orderbook.Asks); i++ {
		qty, _ := m.orderbook.Asks[i].Quantity.Float64()
		if qty > maxQty {
			maxQty = qty
		}
	}
	for i := 0; i < m.depth && i < len(m.orderbook.Bids); i++ {
		qty, _ := m.orderbook.Bids[i].Quantity.Float64()
		if qty > maxQty {
			maxQty = qty
		}
	}
	return maxQty
}

func renderDepthBar(qty, maxQty float64, width int) string {
	barLen := int((qty / maxQty) * float64(width))
	if barLen < 1 {
		barLen = 1
	}
	return strings.Repeat("█", barLen) + strings.Repeat("░", width-barLen)
}
