package update

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/observability-ui/development-tools/internal/constants"
	execctx "github.com/observability-ui/development-tools/pkg/context"
	"github.com/observability-ui/development-tools/pkg/k8s"
	"github.com/observability-ui/development-tools/pkg/mode"
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

	out := output.NewHandler(ctx)

	out.Info(fmt.Sprintf("Updating monitoring plugin to image: %s", image))

	out.Progress("Scaling down CMO...")
	if err := k8s.ScaleDeployment(ctx, kubeClient, constants.CMODeployment, constants.MonitoringNamespace, 0); err != nil {
		out.Error(fmt.Sprintf("Failed to scale down CMO: %v", err))
		return err
	}
	out.Success(fmt.Sprintf("Scaled down %s", constants.CMODeployment))

	out.Progress("Updating monitoring plugin image...")
	if err := k8s.UpdateDeploymentImage(ctx, kubeClient, constants.PluginDeployment, constants.MonitoringNamespace, image); err != nil {
		out.Error(fmt.Sprintf("Failed to update image: %v", err))
		return err
	}
	out.Success(fmt.Sprintf("Updated %s image to %s", constants.PluginDeployment, image))

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

	operations := []string{
		fmt.Sprintf("Scale down %s", constants.CMODeployment),
		fmt.Sprintf("Update %s image to %s", constants.PluginDeployment, image),
	}

	model := tui.NewProgressModel("Updating Monitoring Plugin", operations)
	program := tea.NewProgram(model)

	go func() {
		program.Send(tui.OperationUpdateMsg{Index: 0, Status: tui.OperationInProgress})

		err := k8s.ScaleDeployment(ctx, kubeClient, constants.CMODeployment, constants.MonitoringNamespace, 0)
		if err != nil {
			program.Send(tui.OperationUpdateMsg{Index: 0, Status: tui.OperationFailed, Error: err})
			return
		}
		program.Send(tui.OperationUpdateMsg{Index: 0, Status: tui.OperationComplete})

		program.Send(tui.OperationUpdateMsg{Index: 1, Status: tui.OperationInProgress})

		err = k8s.UpdateDeploymentImage(ctx, kubeClient, constants.PluginDeployment, constants.MonitoringNamespace, image)
		if err != nil {
			program.Send(tui.OperationUpdateMsg{Index: 1, Status: tui.OperationFailed, Error: err})
			return
		}
		program.Send(tui.OperationUpdateMsg{Index: 1, Status: tui.OperationComplete})
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
