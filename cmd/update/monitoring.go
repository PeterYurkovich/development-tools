package update

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/observability-ui/development-tools/internal/constants"
	execctx "github.com/observability-ui/development-tools/pkg/context"
	"github.com/observability-ui/development-tools/pkg/executor"
	"github.com/observability-ui/development-tools/pkg/mode"
	"github.com/observability-ui/development-tools/pkg/operations"
	"github.com/observability-ui/development-tools/pkg/output"
	"github.com/observability-ui/development-tools/pkg/tui"
)

var monitoringCmd = &cobra.Command{
	Use:   "monitoring",
	Short: "Update monitoring plugin image",
	Long:  "Scale down CMO to allow updating the monitoring plugin image",
	RunE:  runUpdateMonitoring,
}

func init() {
	monitoringCmd.Flags().String("image", "", "Monitoring plugin image to use")
	UpdateCmd.AddCommand(monitoringCmd)
}

func runUpdateMonitoring(cmd *cobra.Command, args []string) error {
	requiredFlags := []string{"image"}

	useTUI, err := mode.DetermineMode(cmd, requiredFlags)
	if err != nil {
		return err
	}

	if useTUI {
		return runUpdateMonitoringTUI(cmd)
	}

	return runUpdateMonitoringCLI(cmd)
}

func runUpdateMonitoringCLI(cmd *cobra.Command) error {
	image, _ := cmd.Flags().GetString("image")
	if image == "" {
		return fmt.Errorf("--image flag is required")
	}

	ctx := cmd.Context()
	ctx = execctx.WithTUI(ctx, false)

	kubeClient, err := execctx.GetClient(ctx)
	if err != nil {
		return err
	}

	handler := output.NewCLIHandler()
	exec := executor.NewExecutor()

	go operations.UpdateMonitoring(ctx, kubeClient, operations.UpdateMonitoringConfig{
		Image: image,
	}, exec)

	for update := range exec.UpdateCh {
		if err := handler.HandleUpdate(update); err != nil {
			return err
		}
	}

	return nil
}

func runUpdateMonitoringTUI(cmd *cobra.Command) error {
	ctx := cmd.Context()
	ctx = execctx.WithTUI(ctx, true)

	kubeClient, err := execctx.GetClient(ctx)
	if err != nil {
		return err
	}

	image, _ := cmd.Flags().GetString("image")
	if image == "" {
		image, err = collectImageInput()
		if err != nil {
			return err
		}
	}

	operationsList := []string{
		fmt.Sprintf("Scale down %s", constants.CMODeployment),
		fmt.Sprintf("Update %s image to %s", constants.PluginDeployment, image),
	}

	model := tui.NewProgressModel("Updating Monitoring Plugin", operationsList)
	program := tea.NewProgram(model)

	exec := executor.NewExecutor()

	go operations.UpdateMonitoring(ctx, kubeClient, operations.UpdateMonitoringConfig{
		Image: image,
	}, exec)

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

func collectImageInput() (string, error) {
	var image string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Enter monitoring plugin image").
				Placeholder("quay.io/observability-ui/monitoring-plugin:latest").
				Value(&image).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("image cannot be empty")
					}
					return nil
				}),
		),
	)

	err := form.Run()
	if err != nil {
		return "", err
	}

	return image, nil
}
