package browse

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/donderom/bubblon"
	"github.com/wisp-trading/sdk/pkg/types/config"
	"github.com/wisp-trading/wisp/internal/ui"
	strategyTypes "github.com/wisp-trading/wisp/pkg/strategy"
)

type StrategyListView interface {
	tea.Model
}

type strategyListView struct {
	strategies []config.Strategy
	cursor     int
	pageSize   int
	pageNum    int // 1-based
	width      int
	height     int
	// listTop is the content line offset of the first strategy row (for mouse hit testing)
	listTop         int
	loadErr         error
	compileService  strategyTypes.CompileService
	strategyService config.StrategyConfig
	detailFactory   StrategyDetailViewFactory
}

func newStrategyListView(
	compileService strategyTypes.CompileService,
	strategyService config.StrategyConfig,
	detailFactory StrategyDetailViewFactory,
) tea.Model {
	view := &strategyListView{
		compileService:  compileService,
		strategyService: strategyService,
		detailFactory:   detailFactory,
		pageSize:        10,
		pageNum:         1,
		width:           80,
		height:          24,
	}
	view.refresh()
	return view
}

func (m *strategyListView) refresh() {
	list, err := m.strategyService.FindStrategies()
	m.loadErr = err
	if err != nil {
		m.strategies = nil
		return
	}
	m.strategies = list
}

func (m *strategyListView) Init() tea.Cmd {
	return nil
}

func (m *strategyListView) totalPages() int {
	if len(m.strategies) == 0 || m.pageSize <= 0 {
		return 1
	}
	n := (len(m.strategies) + m.pageSize - 1) / m.pageSize
	if n < 1 {
		return 1
	}
	return n
}

func (m *strategyListView) pageStart() int {
	return (m.pageNum - 1) * m.pageSize
}

func (m *strategyListView) pageEnd() int {
	end := m.pageStart() + m.pageSize
	if end > len(m.strategies) {
		end = len(m.strategies)
	}
	return end
}

func (m *strategyListView) syncPageFromCursor() {
	if m.pageSize <= 0 || len(m.strategies) == 0 {
		m.pageNum = 1
		return
	}
	m.pageNum = m.cursor/m.pageSize + 1
	if m.pageNum > m.totalPages() {
		m.pageNum = m.totalPages()
	}
	if m.pageNum < 1 {
		m.pageNum = 1
	}
}

func (m *strategyListView) clampCursorToPage() {
	start, end := m.pageStart(), m.pageEnd()
	if start >= end {
		m.cursor = 0
		return
	}
	if m.cursor < start {
		m.cursor = start
	}
	if m.cursor >= end {
		m.cursor = end - 1
	}
}

func (m *strategyListView) recomputePageSize() {
	// header ~4 lines, footer ~3, box padding
	avail := m.height - 10
	if avail < 5 {
		avail = 5
	}
	if avail > 20 {
		avail = 20
	}
	m.pageSize = avail
	m.syncPageFromCursor()
	m.clampCursorToPage()
}

func (m *strategyListView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.recomputePageSize()
		return m, nil

	case tea.MouseMsg:
		if msg.Action != tea.MouseActionPress || msg.Button != tea.MouseButtonLeft {
			return m, nil
		}
		// Click on a strategy row (listTop is 1-based content line of first item)
		row := msg.Y - m.listTop
		if row < 0 || row >= m.pageEnd()-m.pageStart() {
			return m, nil
		}
		m.cursor = m.pageStart() + row
		if m.cursor >= 0 && m.cursor < len(m.strategies) {
			detailView := m.detailFactory(&m.strategies[m.cursor])
			return m, bubblon.Open(detailView)
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q", "esc":
			return m, bubblon.Cmd(bubblon.Close())
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.syncPageFromCursor()
			}
		case "down", "j":
			if m.cursor < len(m.strategies)-1 {
				m.cursor++
				m.syncPageFromCursor()
			}
		case "left", "h", "pgup", "ctrl+u":
			if m.pageNum > 1 {
				m.pageNum--
				m.clampCursorToPage()
			}
		case "right", "l", "pgdown", "ctrl+d":
			if m.pageNum < m.totalPages() {
				m.pageNum++
				m.clampCursorToPage()
			}
		case "home", "g":
			m.cursor = 0
			m.pageNum = 1
		case "end", "G":
			if len(m.strategies) > 0 {
				m.cursor = len(m.strategies) - 1
				m.syncPageFromCursor()
			}
		case "enter", " ":
			if len(m.strategies) == 0 {
				return m, nil
			}
			selectedStrat := &m.strategies[m.cursor]
			detailView := m.detailFactory(selectedStrat)
			return m, bubblon.Open(detailView)
		case "r":
			m.refresh()
			if m.cursor >= len(m.strategies) && len(m.strategies) > 0 {
				m.cursor = len(m.strategies) - 1
			}
			if len(m.strategies) == 0 {
				m.cursor = 0
			}
			m.syncPageFromCursor()
			m.clampCursorToPage()
		}
	}
	return m, nil
}

func (m *strategyListView) View() string {
	if m.loadErr != nil {
		return ui.BoxStyle.Render(
			ui.TitleStyle.Render("STRATEGIES") + "\n\n" +
				ui.ErrorBoxStyle.Render("❌ "+m.loadErr.Error()) + "\n\n" +
				ui.MutedStyle.Render("Run wisp from a project root (directory with ./strategies).") + "\n" +
				ui.MutedStyle.Render("r retry   q back"),
		)
	}
	if len(m.strategies) == 0 {
		return ui.BoxStyle.Render(
			ui.TitleStyle.Render("STRATEGIES") + "\n\n" +
				ui.SubtitleStyle.Render("No strategies in ./strategies.") + "\n\n" +
				ui.MutedStyle.Render("Create New Project from the menu, or:") + "\n" +
				ui.MutedStyle.Render("  wisp init my-bot") + "\n\n" +
				ui.MutedStyle.Render("Then: Settings → keys · here → Start Live · Monitor → Stop") + "\n\n" +
				ui.MutedStyle.Render("r refresh   q back"),
		)
	}

	var content string
	content += ui.TitleStyle.Render("STRATEGIES") + "\n"
	content += ui.MutedStyle.Render(
		fmt.Sprintf("%d strategies  ·  ↑↓ move  ←→ page  ↵ open  click row  r refresh  q back", len(m.strategies)),
	) + "\n\n"

	// listTop: approximate Y of first item after box border + title + subtitle + blank
	// Used for mouse; absolute coords depend on terminal — we store relative content line.
	m.listTop = 4 // title, subtitle, blank → items start around line 4 inside box

	start, end := m.pageStart(), m.pageEnd()
	for i := start; i < end; i++ {
		strat := m.strategies[i]
		exchanges := fmt.Sprintf("[%v]", strat.Exchanges)
		line := fmt.Sprintf("%s %s", strat.Name, exchanges)
		if i == m.cursor {
			content += ui.StrategyNameSelectedStyle.Render("▶ "+line) + "\n"
		} else {
			content += ui.StrategyNameStyle.Render("  "+line) + "\n"
		}
	}

	content += "\n" + ui.SubtitleStyle.Render(
		fmt.Sprintf("Page %d/%d  (%d–%d of %d)",
			m.pageNum, m.totalPages(), start+1, end, len(m.strategies)),
	)

	return ui.BoxStyle.Render(content)
}
