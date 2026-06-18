package cleanup

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/observability-ui/development-tools/internal/constants"
	execctx "github.com/observability-ui/development-tools/pkg/context"
	"github.com/observability-ui/development-tools/pkg/executor"
	"github.com/observability-ui/development-tools/pkg/operations"
	"github.com/observability-ui/development-tools/pkg/output"
	"github.com/observability-ui/development-tools/pkg/tui"
)

var monitoringCmd = &cobra.Command{
	Use:   "monitoring",
	Short: "Restore monitoring to normal state",
	Long:  "Scale up CMO to restore monitoring plugin to its managed state",
	RunE:  runCleanupMonitoring,
}

func init() {
	CleanupCmd.AddCommand(monitoringCmd)
}

func runCleanupMonitoring(cmd *cobra.Command, args []string) error {
	isTUI := output.IsTerminal()

	if isTUI {
		return runCleanupMonitoringTUI(cmd)
	}

	return runCleanupMonitoringCLI(cmd)
}

func runCleanupMonitoringCLI(cmd *cobra.Command) error {
	ctx := cmd.Context()
	ctx = execctx.WithTUI(ctx, false)

	kubeClient, err := execctx.GetClient(ctx)
	if err != nil {
		return err
	}

	handler := output.NewCLIHandler()
	exec := executor.NewExecutor()

	go operations.CleanupMonitoring(ctx, kubeClient, operations.CleanupMonitoringConfig{}, exec)

	for update := range exec.UpdateCh {
		if err := handler.HandleUpdate(update); err != nil {
			return err
		}
	}

	return nil
}

func runCleanupMonitoringTUI(cmd *cobra.Command) error {
	ctx := cmd.Context()
	ctx = execctx.WithTUI(ctx, true)

	kubeClient, err := execctx.GetClient(ctx)
	if err != nil {
		return err
	}

	operationsList := []string{
		fmt.Sprintf("Scale up %s", constants.CMODeployment),
	}

	model := tui.NewProgressModel("Restoring Monitoring", operationsList)
	program := tea.NewProgram(model)

	exec := executor.NewExecutor()

	go operations.CleanupMonitoring(ctx, kubeClient, operations.CleanupMonitoringConfig{}, exec)

	go func() {
		for update := range exec.UpdateCh {
			if update.Message != "" {
				continue
			}

			program.Send(tui.OperationUpdateMsg{
				Index:  update.Index,
				Status: convertStatus(update.Status),
				Error:  update.Error,
			})
		}
	}()

	finalModel, err := program.Run()
	if err != nil {
		return err
	}

	m := finalModel.(tui.ProgressModel)
	if m.Error() != nil {
		return m.Error()
	}

	return nil
}

func convertStatus(status executor.UpdateStatus) tui.OperationStatus {
	switch status {
	case executor.StatusPending:
		return tui.OperationPending
	case executor.StatusInProgress:
		return tui.OperationInProgress
	case executor.StatusComplete:
		return tui.OperationComplete
	case executor.StatusFailed:
		return tui.OperationFailed
	default:
		return tui.OperationPending
	}
}
