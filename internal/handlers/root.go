package handlers

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/wisp-trading/wisp/internal/handlers/settings"
	"github.com/wisp-trading/wisp/internal/handlers/strategies"
	backtesting "github.com/wisp-trading/wisp/internal/handlers/strategies/backtest/types"
	"github.com/wisp-trading/wisp/internal/handlers/strategies/browse"
	"github.com/wisp-trading/wisp/internal/handlers/strategies/monitor"
	"github.com/wisp-trading/wisp/internal/router"
	setup "github.com/wisp-trading/wisp/internal/setup/types"
)

type RootHandler interface {
	Handle(cmd *cobra.Command, args []string) error
}

type rootHandler struct {
	strategyBrowser      strategies.StrategyBrowser
	initHandler          setup.InitHandler
	scaffold             setup.ScaffoldService
	backtestHandler      backtesting.BacktestHandler
	analyzeHandler       backtesting.AnalyzeHandler
	monitorViewFactory   monitor.MonitorViewFactory
	strategyListFactory  browse.StrategyListViewFactory
	settingsListFactory  settings.SettingsListViewFactory
	connectorFormFactory settings.ConnectorFormViewFactory
	deleteConfirmFactory settings.DeleteConfirmViewFactory
	router               router.Router
}

func NewRootHandler(
	strategyBrowser strategies.StrategyBrowser,
	initHandler setup.InitHandler,
	scaffold setup.ScaffoldService,
	backtestHandler backtesting.BacktestHandler,
	analyzeHandler backtesting.AnalyzeHandler,
	monitorViewFactory monitor.MonitorViewFactory,
	strategyListFactory browse.StrategyListViewFactory,
	settingsListFactory settings.SettingsListViewFactory,
	connectorFormFactory settings.ConnectorFormViewFactory,
	deleteConfirmFactory settings.DeleteConfirmViewFactory,
	r router.Router,
) RootHandler {
	r.RegisterRoute(router.RouteMonitor, func() tea.Model {
		return monitorViewFactory()
	})

	r.RegisterRoute(router.RouteStrategyList, func() tea.Model {
		return strategyListFactory()
	})

	r.RegisterRoute(router.RouteSettingsList, func() tea.Model {
		return settingsListFactory()
	})

	r.RegisterRoute(router.RouteSettingsCreate, func() tea.Model {
		return connectorFormFactory("", false)
	})

	r.RegisterRoute(router.RouteSettingsEdit, func() tea.Model {
		return connectorFormFactory("", true)
	})

	r.RegisterRoute(router.RouteSettingsDelete, func() tea.Model {
		return deleteConfirmFactory("")
	})

	return &rootHandler{
		strategyBrowser:      strategyBrowser,
		initHandler:          initHandler,
		scaffold:             scaffold,
		backtestHandler:      backtestHandler,
		analyzeHandler:       analyzeHandler,
		monitorViewFactory:   monitorViewFactory,
		strategyListFactory:  strategyListFactory,
		settingsListFactory:  settingsListFactory,
		connectorFormFactory: connectorFormFactory,
		deleteConfirmFactory: deleteConfirmFactory,
		router:               r,
	}
}

func (h *rootHandler) Handle(cmd *cobra.Command, args []string) error {
	cliMode, _ := cmd.Flags().GetBool("cli")
	if cliMode || len(args) > 0 {
		return cmd.Help()
	}
	return h.runMainMenu()
}

func (h *rootHandler) runMainMenu() error {
	m := mainMenuModel{
		choices: []string{
			"Strategies",
			"Monitor",
			"Settings",
			"Create New Project",
			"Help",
		},
		router:   h.router,
		scaffold: h.scaffold,
	}

	h.router.SetInitialView(m)
	p := tea.NewProgram(h.router, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
