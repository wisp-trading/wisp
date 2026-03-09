package tabs

import (
	"fmt"
	"math"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wisp-trading/sdk/pkg/types/connector"
	"github.com/wisp-trading/sdk/pkg/types/monitoring"
	"github.com/wisp-trading/wisp/internal/ui"
)

// Available kline intervals and their display labels
var klineIntervals = []string{"1m", "5m", "15m", "1h", "4h", "1d"}

// KlinesViewModel displays a bar chart of kline (OHLCV) data for a given exchange/pair
type KlinesViewModel struct {
	querier      monitoring.ViewQuerier
	instanceID   string
	exchangeName connector.ExchangeName
	item         assetItem

	// Data
	klines  []connector.Kline
	err     error
	loading bool

	// Settings
	intervalIndex int // index into klineIntervals
	limit         int

	// Update tracking
	lastUpdate time.Time
}

// NewKlinesViewModel creates a new klines chart view
func NewKlinesViewModel(
	querier monitoring.ViewQuerier,
	instanceID string,
	exchangeName connector.ExchangeName,
	item assetItem,
) *KlinesViewModel {
	return &KlinesViewModel{
		querier:       querier,
		instanceID:    instanceID,
		exchangeName:  exchangeName,
		item:          item,
		intervalIndex: 3, // default: 1h
		limit:         60,
		loading:       true,
	}
}

// Messages
type klinesDataMsg struct {
	klines []connector.Kline
	err    error
}

type klinesTickMsg time.Time

func (m *KlinesViewModel) Init() tea.Cmd {
	return tea.Batch(m.fetchData(), m.tick())
}

func (m *KlinesViewModel) tick() tea.Cmd {
	return tea.Tick(30*time.Second, func(t time.Time) tea.Msg {
		return klinesTickMsg(t)
	})
}

func (m *KlinesViewModel) currentInterval() string {
	return klineIntervals[m.intervalIndex]
}

func (m *KlinesViewModel) fetchData() tea.Cmd {
	exchange := m.exchangeName
	pair := m.item.pair
	interval := m.currentInterval()
	limit := m.limit
	instanceID := m.instanceID

	return func() tea.Msg {
		klines, err := m.querier.QueryKlines(
			instanceID,
			exchange,
			pair,
			interval,
			limit,
		)
		return klinesDataMsg{klines: klines, err: err}
	}
}

func (m *KlinesViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case klinesDataMsg:
		m.loading = false
		m.err = msg.err
		m.klines = msg.klines
		if msg.err == nil {
			m.lastUpdate = time.Now()
		}
		return m, nil

	case klinesTickMsg:
		return m, tea.Batch(m.fetchData(), m.tick())

	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			// Previous interval
			if m.intervalIndex > 0 {
				m.intervalIndex--
				m.loading = true
				m.klines = nil
				return m, m.fetchData()
			}
			return m, nil

		case "right", "l":
			// Next interval
			if m.intervalIndex < len(klineIntervals)-1 {
				m.intervalIndex++
				m.loading = true
				m.klines = nil
				return m, m.fetchData()
			}
			return m, nil

		case "+", "=":
			// More candles
			if m.limit < 200 {
				m.limit += 20
				m.loading = true
				return m, m.fetchData()
			}
			return m, nil

		case "-":
			// Fewer candles
			if m.limit > 20 {
				m.limit -= 20
				m.loading = true
				return m, m.fetchData()
			}
			return m, nil

		case "r":
			m.loading = true
			return m, m.fetchData()

		case "esc":
			return m, func() tea.Msg {
				return backToExchangeListMsg{source: "klines"}
			}
		}
	}
	return m, nil
}

func (m *KlinesViewModel) View() string {
	var b strings.Builder

	b.WriteString(m.renderHeader())
	b.WriteString("\n\n")

	if m.loading && m.klines == nil {
		b.WriteString(ui.SubtitleStyle.Render("Loading klines..."))
		b.WriteString("\n\n")
		b.WriteString(ui.HelpStyle.Render("[←→] Interval • [+/-] Candles • [Esc] Back"))
		return b.String()
	}

	if m.err != nil {
		b.WriteString(ui.ErrorBoxStyle.Render(fmt.Sprintf("Error fetching klines: %v", m.err)))
		b.WriteString("\n\n")
		b.WriteString(ui.HelpStyle.Render("[R] Retry • [←→] Interval • [Esc] Back"))
		return b.String()
	}

	if len(m.klines) == 0 {
		b.WriteString(ui.SubtitleStyle.Render("No kline data available for this interval"))
		b.WriteString("\n\n")
		b.WriteString(ui.HelpStyle.Render("[←→] Change Interval • [Esc] Back"))
		return b.String()
	}

	b.WriteString(m.renderIntervalSelector())
	b.WriteString("\n\n")
	b.WriteString(m.renderChart())
	b.WriteString("\n\n")
	b.WriteString(m.renderStats())
	b.WriteString("\n\n")
	b.WriteString(ui.HelpStyle.Render("[←→/hl] Interval • [+/-] Candles • [R] Refresh • [Esc] Back"))

	return b.String()
}

func (m *KlinesViewModel) renderHeader() string {
	var h strings.Builder

	title := fmt.Sprintf("KLINES - %s @ %s", strings.ToUpper(m.item.pair.Symbol()), strings.ToUpper(string(m.exchangeName)))
	h.WriteString(ui.StrategyNameStyle.Render(title))
	h.WriteString("  ")

	if !m.lastUpdate.IsZero() {
		h.WriteString(ui.StatusReadyStyle.Render("● LIVE"))
		ago := time.Since(m.lastUpdate)
		if ago < time.Second {
			h.WriteString(ui.SubtitleStyle.Render("  <1s ago"))
		} else {
			h.WriteString(ui.SubtitleStyle.Render(fmt.Sprintf("  %ds ago", int(ago.Seconds()))))
		}
	}

	return h.String()
}

func (m *KlinesViewModel) renderIntervalSelector() string {
	var b strings.Builder
	b.WriteString(ui.LabelStyle.Render("Interval: "))
	for i, iv := range klineIntervals {
		if i == m.intervalIndex {
			b.WriteString(ui.SelectedItemStyle.Render(fmt.Sprintf("[%s]", iv)))
		} else {
			b.WriteString(ui.SubtitleStyle.Render(fmt.Sprintf(" %s ", iv)))
		}
		if i < len(klineIntervals)-1 {
			b.WriteString(ui.SubtitleStyle.Render(" · "))
		}
	}
	b.WriteString(ui.SubtitleStyle.Render(fmt.Sprintf("   (%d candles)", m.limit)))
	return b.String()
}

// renderChart renders an ASCII bar chart with OHLC bars
func (m *KlinesViewModel) renderChart() string {
	if len(m.klines) == 0 {
		return ""
	}

	const chartHeight = 20
	const maxBars = 80

	klines := m.klines
	if len(klines) > maxBars {
		klines = klines[len(klines)-maxBars:]
	}

	// Find price range
	minPrice, maxPrice := math.MaxFloat64, -math.MaxFloat64
	for _, k := range klines {
		if k.Low < minPrice {
			minPrice = k.Low
		}
		if k.High > maxPrice {
			maxPrice = k.High
		}
	}

	priceRange := maxPrice - minPrice
	if priceRange == 0 {
		priceRange = 1
	}

	// Build chart grid: rows (price levels) x cols (candles)
	// Each row is a price level, each col is a candle
	numCols := len(klines)
	grid := make([][]rune, chartHeight)
	colors := make([][]bool, chartHeight) // true = bullish (green), false = bearish (red)
	for i := range grid {
		grid[i] = make([]rune, numCols)
		colors[i] = make([]bool, numCols)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}

	priceToRow := func(price float64) int {
		// row 0 = top (highest price), row chartHeight-1 = bottom (lowest price)
		normalized := (price - minPrice) / priceRange
		row := chartHeight - 1 - int(normalized*float64(chartHeight-1))
		if row < 0 {
			row = 0
		}
		if row >= chartHeight {
			row = chartHeight - 1
		}
		return row
	}

	for col, k := range klines {
		bullish := k.Close >= k.Open

		highRow := priceToRow(k.High)
		lowRow := priceToRow(k.Low)
		bodyTop := priceToRow(math.Max(k.Open, k.Close))
		bodyBottom := priceToRow(math.Min(k.Open, k.Close))

		// Draw wick
		for row := highRow; row <= lowRow; row++ {
			if grid[row][col] == ' ' {
				grid[row][col] = '│'
				colors[row][col] = bullish
			}
		}

		// Draw body (overwrite wick)
		for row := bodyTop; row <= bodyBottom; row++ {
			grid[row][col] = '█'
			colors[row][col] = bullish
		}

		// If open == close (doji), draw a dash
		if bodyTop == bodyBottom {
			grid[bodyTop][col] = '─'
			colors[bodyTop][col] = bullish
		}
	}

	// Render the grid with price axis
	var b strings.Builder
	priceStep := priceRange / float64(chartHeight-1)

	bullStyle := lipgloss.NewStyle().Foreground(ui.ColorSuccess)
	bearStyle := lipgloss.NewStyle().Foreground(ui.ColorDanger)
	wickStyle := lipgloss.NewStyle().Foreground(ui.ColorMuted)

	for row := 0; row < chartHeight; row++ {
		// Price label on the left
		price := maxPrice - float64(row)*priceStep
		label := formatKlinePrice(price)
		b.WriteString(lipgloss.NewStyle().Foreground(ui.ColorMuted).Width(12).Align(lipgloss.Right).Render(label))
		b.WriteString(" │")

		// Chart columns
		for col := 0; col < numCols; col++ {
			ch := grid[row][col]
			if ch == ' ' {
				b.WriteRune(' ')
				continue
			}
			isBullish := colors[row][col]
			switch ch {
			case '│':
				b.WriteString(wickStyle.Render("│"))
			case '█':
				if isBullish {
					b.WriteString(bullStyle.Render("█"))
				} else {
					b.WriteString(bearStyle.Render("█"))
				}
			case '─':
				if isBullish {
					b.WriteString(bullStyle.Render("─"))
				} else {
					b.WriteString(bearStyle.Render("─"))
				}
			default:
				b.WriteRune(ch)
			}
		}
		b.WriteString("\n")
	}

	// X-axis line
	b.WriteString(lipgloss.NewStyle().Foreground(ui.ColorMuted).Width(12).Render(""))
	b.WriteString(" └")
	b.WriteString(lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(strings.Repeat("─", numCols)))
	b.WriteString("\n")

	// Time labels: show a few timestamps along the bottom
	if len(klines) > 0 {
		b.WriteString(m.renderTimeAxis(klines, numCols))
	}

	return b.String()
}

func (m *KlinesViewModel) renderTimeAxis(klines []connector.Kline, numCols int) string {
	const labelWidth = 8
	const prefix = 14 // price label width + " │"

	// Place labels at ~every 15 columns
	step := 15
	if numCols < 30 {
		step = 10
	}

	row := make([]byte, prefix+numCols)
	for i := range row {
		row[i] = ' '
	}

	for col := 0; col < numCols; col += step {
		k := klines[col]
		ts := k.OpenTime
		var label string
		switch m.currentInterval() {
		case "1d":
			label = ts.Format("01/02")
		case "4h", "1h":
			label = ts.Format("01/02")
		default:
			label = ts.Format("15:04")
		}

		pos := prefix + col
		for i, ch := range []byte(label) {
			if pos+i < len(row) {
				row[pos+i] = ch
			}
		}
	}

	_ = labelWidth
	return ui.SubtitleStyle.Render(string(row)) + "\n"
}

func (m *KlinesViewModel) renderStats() string {
	if len(m.klines) == 0 {
		return ""
	}

	last := m.klines[len(m.klines)-1]

	open := last.Open
	high := last.High
	low := last.Low
	close_ := last.Close
	vol := last.Volume

	change := close_ - open
	changePct := 0.0
	if open != 0 {
		changePct = (change / open) * 100
	}

	var changeStr string
	if change >= 0 {
		changeStr = ui.PnLProfitStyle.Render(fmt.Sprintf("+%s (%.2f%%)", formatKlinePrice(change), changePct))
	} else {
		changeStr = ui.PnLLossStyle.Render(fmt.Sprintf("%s (%.2f%%)", formatKlinePrice(change), changePct))
	}

	// Compute overall range stats
	var minLow, maxHigh float64 = math.MaxFloat64, -math.MaxFloat64
	totalVol := 0.0
	for _, k := range m.klines {
		if k.Low < minLow {
			minLow = k.Low
		}
		if k.High > maxHigh {
			maxHigh = k.High
		}
		totalVol += k.Volume
	}

	var b strings.Builder
	b.WriteString(ui.SectionHeaderStyle.Render("LAST CANDLE"))
	b.WriteString("\n")

	row1 := fmt.Sprintf("%s %s   %s %s   %s %s   %s %s   %s %s",
		ui.LabelStyle.Render("O:"), ui.ValueStyle.Render(formatKlinePrice(open)),
		ui.LabelStyle.Render("H:"), ui.PnLProfitStyle.Render(formatKlinePrice(high)),
		ui.LabelStyle.Render("L:"), ui.PnLLossStyle.Render(formatKlinePrice(low)),
		ui.LabelStyle.Render("C:"), ui.ValueStyle.Render(formatKlinePrice(close_)),
		ui.LabelStyle.Render("Chg:"), changeStr,
	)
	b.WriteString(row1)
	b.WriteString("\n")

	row2 := fmt.Sprintf("%s %s   %s %s – %s   %s %s",
		ui.LabelStyle.Render("Vol:"), ui.SubtitleStyle.Render(formatVolume(vol)),
		ui.LabelStyle.Render("Range:"), ui.SubtitleStyle.Render(formatKlinePrice(minLow)),
		ui.SubtitleStyle.Render(formatKlinePrice(maxHigh)),
		ui.LabelStyle.Render("Total Vol:"), ui.SubtitleStyle.Render(formatVolume(totalVol)),
	)
	b.WriteString(row2)

	return b.String()
}

// formatKlinePrice formats a price value sensibly regardless of magnitude
func formatKlinePrice(price float64) string {
	abs := math.Abs(price)
	switch {
	case abs == 0:
		return "0.00"
	case abs >= 10000:
		return fmt.Sprintf("%.0f", price)
	case abs >= 1000:
		return fmt.Sprintf("%.1f", price)
	case abs >= 1:
		return fmt.Sprintf("%.2f", price)
	case abs >= 0.01:
		return fmt.Sprintf("%.4f", price)
	default:
		return fmt.Sprintf("%.6f", price)
	}
}

// formatVolume formats a volume number with K/M suffixes
func formatVolume(vol float64) string {
	switch {
	case vol >= 1_000_000:
		return fmt.Sprintf("%.2fM", vol/1_000_000)
	case vol >= 1_000:
		return fmt.Sprintf("%.2fK", vol/1_000)
	default:
		return fmt.Sprintf("%.2f", vol)
	}
}
