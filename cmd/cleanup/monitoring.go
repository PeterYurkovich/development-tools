package cleanup

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/observability-ui/development-tools/internal/constants"
	execctx "github.com/observability-ui/development-tools/pkg/context"
	"github.com/observability-ui/development-tools/pkg/k8s"
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

	if err := k8s.ScaleDeployment(ctx, kubeClient, constants.CMODeployment, constants.MonitoringNamespace, 1); err != nil {
		return fmt.Errorf("failed to scale up CMO: %w", err)
	}
	fmt.Printf("Scaled up %s\n", constants.CMODeployment)

	return nil
}
