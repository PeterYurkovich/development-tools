package update

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/observability-ui/development-tools/internal/constants"
	execctx "github.com/observability-ui/development-tools/pkg/context"
	"github.com/observability-ui/development-tools/pkg/k8s"
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
	image, _ := cmd.Flags().GetString("image")
	if image == "" {
		return fmt.Errorf("--image flag is required")
	}

	ctx := cmd.Context()
	kubeClient, err := execctx.GetClient(ctx)
	if err != nil {
		return err
	}

	if err := k8s.ScaleDeployment(ctx, kubeClient, constants.CMODeployment, constants.MonitoringNamespace, 0); err != nil {
		return fmt.Errorf("failed to scale down CMO: %w", err)
	}
	fmt.Printf("Scaled down %s\n", constants.CMODeployment)

	if err := k8s.UpdateDeploymentImage(ctx, kubeClient, constants.PluginDeployment, constants.MonitoringNamespace, image); err != nil {
		return fmt.Errorf("failed to update monitoring plugin image: %w", err)
	}
	fmt.Printf("Updated %s image to %s\n", constants.PluginDeployment, image)

	return nil
}
