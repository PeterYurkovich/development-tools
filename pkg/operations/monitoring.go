package operations

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/observability-ui/development-tools/internal/constants"
	"github.com/observability-ui/development-tools/pkg/executor"
	"github.com/observability-ui/development-tools/pkg/k8s"
	"github.com/observability-ui/development-tools/pkg/resources"
)

const (
	StepScaleDownCMO = iota
	StepUpdatePluginImage
)

const (
	StepScaleUpCMO = iota
)

type UpdateMonitoringConfig struct {
	Image string
}

func UpdateMonitoring(ctx context.Context, kubeClient client.Client, config UpdateMonitoringConfig, exec *executor.Executor) error {
	defer exec.Close()

	stepName := fmt.Sprintf("Scale down %s", constants.CMODeployment)
	exec.SendUpdate(StepScaleDownCMO, executor.StatusInProgress, stepName)

	exec.SendLog(StepScaleDownCMO, "Updating deployment replicas to 0")
	err := k8s.ScaleDeployment(ctx, kubeClient, constants.CMODeployment, constants.MonitoringNamespace, 0)
	if err != nil {
		exec.SendUpdateWithError(StepScaleDownCMO, executor.StatusFailed, stepName, err)
		return err
	}

	exec.SendLog(StepScaleDownCMO, "Deployment scaled down successfully")
	exec.SendUpdate(StepScaleDownCMO, executor.StatusComplete, stepName)

	stepName = fmt.Sprintf("Update %s image to %s", constants.PluginDeployment, config.Image)
	exec.SendUpdate(StepUpdatePluginImage, executor.StatusInProgress, stepName)

	exec.SendLog(StepUpdatePluginImage, fmt.Sprintf("Patching deployment with new image: %s", config.Image))
	err = k8s.UpdateDeploymentImage(ctx, kubeClient, constants.PluginDeployment, constants.MonitoringNamespace, config.Image)
	if err != nil {
		exec.SendUpdateWithError(StepUpdatePluginImage, executor.StatusFailed, stepName, err)
		return err
	}

	exec.SendLog(StepUpdatePluginImage, "Image updated successfully")
	exec.SendUpdate(StepUpdatePluginImage, executor.StatusComplete, stepName)

	return nil
}

const (
	StepDeployMonitoringUIPlugin = iota
)

type DeployMonitoringConfig struct {
	EnablePerses                bool
	EnableACM                   bool
	EnableClusterHealthAnalyzer bool
}

func DeployMonitoring(ctx context.Context, kubeClient client.Client, config DeployMonitoringConfig, exec *executor.Executor) error {
	defer exec.Close()

	stepName := "Deploy Monitoring UIPlugin"
	exec.SendUpdate(StepDeployMonitoringUIPlugin, executor.StatusInProgress, stepName)
	exec.SendLog(StepDeployMonitoringUIPlugin, fmt.Sprintf("Creating UIPlugin %s", constants.MonitoringUIPluginName))

	pluginConfig := resources.MonitoringUIPluginConfig{
		EnablePerses:                config.EnablePerses,
		EnableACM:                   config.EnableACM,
		EnableClusterHealthAnalyzer: config.EnableClusterHealthAnalyzer,
	}

	if err := resources.CreateMonitoringUIPlugin(ctx, kubeClient, pluginConfig); err != nil {
		exec.SendUpdateWithError(StepDeployMonitoringUIPlugin, executor.StatusFailed, stepName, err)
		return err
	}

	exec.SendUpdate(StepDeployMonitoringUIPlugin, executor.StatusComplete, stepName)
	return nil
}

type CleanupMonitoringConfig struct{}

func CleanupMonitoring(ctx context.Context, kubeClient client.Client, config CleanupMonitoringConfig, exec *executor.Executor) error {
	defer exec.Close()

	stepName := fmt.Sprintf("Scale up %s", constants.CMODeployment)
	exec.SendUpdate(StepScaleUpCMO, executor.StatusInProgress, stepName)

	exec.SendLog(StepScaleUpCMO, "Restoring CMO to normal state (replicas: 1)")
	err := k8s.ScaleDeployment(ctx, kubeClient, constants.CMODeployment, constants.MonitoringNamespace, 1)
	if err != nil {
		exec.SendUpdateWithError(StepScaleUpCMO, executor.StatusFailed, stepName, err)
		return err
	}

	exec.SendLog(StepScaleUpCMO, "CMO will reconcile and restore monitoring plugin")
	exec.SendUpdate(StepScaleUpCMO, executor.StatusComplete, stepName)

	return nil
}
