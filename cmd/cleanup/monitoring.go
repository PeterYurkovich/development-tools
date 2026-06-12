package cleanup

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/observability-ui/development-tools/internal/constants"
	execctx "github.com/observability-ui/development-tools/pkg/context"
	"github.com/observability-ui/development-tools/pkg/k8s"
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
	ctx := cmd.Context()
	kubeClient, err := execctx.GetClient(ctx)
	if err != nil {
		return err
	}

	isTUI := output.IsTerminal()
	ctx = execctx.WithTUI(ctx, isTUI)

	if isTUI {
		return runCleanupMonitoringTUI(ctx, kubeClient)
	}

	return runCleanupMonitoringCLI(ctx, kubeClient)
}

func runCleanupMonitoringCLI(ctx context.Context, kubeClient client.Client) error {
	out := output.NewHandler(ctx)

	out.Info("Restoring monitoring to normal state")

	out.Progress("Scaling up CMO...")
	if err := k8s.ScaleDeployment(ctx, kubeClient, constants.CMODeployment, constants.MonitoringNamespace, 1); err != nil {
		out.Error(fmt.Sprintf("Failed to scale up CMO: %v", err))
		return err
	}
	out.Success(fmt.Sprintf("Scaled up %s (will reconcile monitoring plugin)", constants.CMODeployment))

	return nil
}

func runCleanupMonitoringTUI(ctx context.Context, kubeClient client.Client) error {
	operations := []string{
		fmt.Sprintf("Scale up %s", constants.CMODeployment),
	}

	model := tui.NewProgressModel("Restoring Monitoring", operations)
	program := tea.NewProgram(model)

	go func() {
		program.Send(tui.OperationUpdateMsg{Index: 0, Status: tui.OperationInProgress})

		err := k8s.ScaleDeployment(ctx, kubeClient, constants.CMODeployment, constants.MonitoringNamespace, 1)
		if err != nil {
			program.Send(tui.OperationUpdateMsg{Index: 0, Status: tui.OperationFailed, Error: err})
			return
		}
		program.Send(tui.OperationUpdateMsg{Index: 0, Status: tui.OperationComplete})
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
