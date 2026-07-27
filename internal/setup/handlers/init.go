package handlers

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wisp-trading/wisp/internal/setup/services"
	"github.com/wisp-trading/wisp/internal/setup/types"
)

// InitHandler handles the init command
type initHandler struct {
	scaffoldService types.ScaffoldService
}

func NewInitHandler(scaffoldService types.ScaffoldService) types.InitHandler {
	return &initHandler{
		scaffoldService: scaffoldService,
	}
}

func (h *initHandler) Handle(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		strategyExample, projectName, err := RunInitTUI()
		if err != nil {
			return err
		}
		if err := h.scaffoldService.CreateProjectWithStrategy(projectName, strategyExample); err != nil {
			return err
		}
		_, _ = fmt.Fprint(cmd.OutOrStdout(), services.FormatProjectCreatedMsg(projectName, strategyExample))
		return nil
	}

	name := args[0]
	if err := h.scaffoldService.CreateProject(name); err != nil {
		return err
	}
	_, _ = fmt.Fprint(cmd.OutOrStdout(), services.FormatProjectCreatedMsg(name, "starter"))
	return nil
}

func (h *initHandler) HandleWithStrategy(strategyExample, name string) error {
	return h.scaffoldService.CreateProjectWithStrategy(name, strategyExample)
}
