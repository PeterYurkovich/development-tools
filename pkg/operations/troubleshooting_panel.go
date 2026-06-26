package operations

import (
	"context"
	"fmt"

	lokiv1 "github.com/grafana/loki/operator/api/loki/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/observability-ui/development-tools/internal/constants"
	"github.com/observability-ui/development-tools/pkg/executor"
	"github.com/observability-ui/development-tools/pkg/k8s"
	netobservOperator "github.com/observability-ui/development-tools/pkg/operators/netobserv"
	"github.com/observability-ui/development-tools/pkg/operators"
	"github.com/observability-ui/development-tools/pkg/resources"
	"github.com/observability-ui/development-tools/pkg/storage"
)

const (
	StepTSPEnsureNetObservNS = iota
	StepTSPEnsureOperatorGroup
	StepTSPCreateSubscription
	StepTSPWaitForCSV
	StepTSPDeployMinIO
	StepTSPEnsureNetObservWorkloadNS
	StepTSPCreateLokiStack
	StepTSPWaitForLokiStack
	StepTSPCreateFlowCollector
	StepTSPWaitForNetObservPlugin
	StepTSPDeployUIPlugin
)

type DeployTroubleshootingPanelConfig struct {
	NetObservChannel    string
	StorageClassName    string
	DeployFlowCollector bool
	DeployUIPlugin      bool
}

func DeployTroubleshootingPanel(ctx context.Context, kubeClient client.Client,
	config DeployTroubleshootingPanelConfig, exec *executor.Executor) error {
	defer exec.Close()

	channel := config.NetObservChannel
	if channel == "" {
		channel = constants.NetObservDefaultChannel
	}

	stepName := fmt.Sprintf("Create namespace %s", constants.NetObservOperatorNamespace)
	exec.SendUpdate(StepTSPEnsureNetObservNS, executor.StatusInProgress, stepName)
	exec.SendLog(StepTSPEnsureNetObservNS, "Ensuring Network Observability operator namespace exists")

	created, err := k8s.EnsureNamespaceWithLabels(ctx, kubeClient, constants.NetObservOperatorNamespace, nil)
	if err != nil {
		exec.SendUpdateWithError(StepTSPEnsureNetObservNS, executor.StatusFailed, stepName, err)
		return err
	}
	if created {
		exec.SendLog(StepTSPEnsureNetObservNS, fmt.Sprintf("Namespace %s created", constants.NetObservOperatorNamespace))
	} else {
		exec.SendLog(StepTSPEnsureNetObservNS, fmt.Sprintf("Namespace %s already exists", constants.NetObservOperatorNamespace))
	}
	exec.SendUpdate(StepTSPEnsureNetObservNS, executor.StatusComplete, stepName)

	if err := netobservOperator.DeployViaOperatorHub(ctx, kubeClient, exec, StepTSPEnsureOperatorGroup, channel); err != nil {
		return err
	}

	stepName = "Wait for netobserv-operator to be ready"
	exec.SendUpdate(StepTSPWaitForCSV, executor.StatusInProgress, stepName)
	exec.SendLog(StepTSPWaitForCSV, "Waiting for Network Observability operator CSV to reach Succeeded phase")

	netobservSubscription, err := operators.GetSubscription(ctx, kubeClient, constants.NetObservOperatorName, constants.NetObservOperatorNamespace)
	if err != nil {
		exec.SendUpdateWithError(StepTSPWaitForCSV, executor.StatusFailed, stepName, err)
		return err
	}

	if err := operators.WaitForCSVSucceeded(ctx, kubeClient, netobservSubscription.Name, constants.NetObservOperatorNamespace, 0, exec, StepTSPWaitForCSV); err != nil {
		exec.SendUpdateWithError(StepTSPWaitForCSV, executor.StatusFailed, stepName, err)
		return err
	}
	exec.SendUpdate(StepTSPWaitForCSV, executor.StatusComplete, stepName)

	if config.DeployFlowCollector {
		storageConfig := storage.ProviderConfig{
			Type:                   storage.StorageTypeMinio,
			Namespace:              constants.MinioNamespace,
			BucketName:             "loki",
			StorageSize:            "10Gi",
			StorageClass:           config.StorageClassName,
			UseDefaultStorageClass: config.StorageClassName == "",
		}

		provider, err := storage.NewProvider(storageConfig)
		if err != nil {
			return fmt.Errorf("failed to create storage provider: %w", err)
		}

		exec.SendUpdate(StepTSPDeployMinIO, executor.StatusInProgress, "Deploy MinIO storage")
		exec.SendLog(StepTSPDeployMinIO, fmt.Sprintf("Deploying MinIO in %s namespace", constants.MinioNamespace))

		if _, err := provider.Deploy(ctx, kubeClient, exec); err != nil {
			exec.SendUpdateWithError(StepTSPDeployMinIO, executor.StatusFailed, "Deploy MinIO storage", err)
			return err
		}
		exec.SendUpdate(StepTSPDeployMinIO, executor.StatusComplete, "Deploy MinIO storage")

		stepName = fmt.Sprintf("Create namespace %s", constants.NetObservNamespace)
		exec.SendUpdate(StepTSPEnsureNetObservWorkloadNS, executor.StatusInProgress, stepName)
		exec.SendLog(StepTSPEnsureNetObservWorkloadNS, "Ensuring netobserv workload namespace exists")

		created, err = k8s.EnsureNamespaceWithLabels(ctx, kubeClient, constants.NetObservNamespace, map[string]string{
			"openshift.io/cluster-monitoring": "true",
		})
		if err != nil {
			exec.SendUpdateWithError(StepTSPEnsureNetObservWorkloadNS, executor.StatusFailed, stepName, err)
			return err
		}
		if created {
			exec.SendLog(StepTSPEnsureNetObservWorkloadNS, fmt.Sprintf("Namespace %s created", constants.NetObservNamespace))
		} else {
			exec.SendLog(StepTSPEnsureNetObservWorkloadNS, fmt.Sprintf("Namespace %s already exists", constants.NetObservNamespace))
		}
		exec.SendUpdate(StepTSPEnsureNetObservWorkloadNS, executor.StatusComplete, stepName)

		stepName = fmt.Sprintf("Create LokiStack %s", constants.NetObservLokiStackName)
		exec.SendUpdate(StepTSPCreateLokiStack, executor.StatusInProgress, stepName)
		exec.SendLog(StepTSPCreateLokiStack, "Creating MinIO secret and LokiStack for network observability")

		if err := resources.CreateNetObservLokiStackSecret(ctx, kubeClient); err != nil {
			exec.SendUpdateWithError(StepTSPCreateLokiStack, executor.StatusFailed, stepName, err)
			return err
		}

		if err := resources.CreateLokiStack(ctx, kubeClient, resources.LokiStackConfig{
			Name:            constants.NetObservLokiStackName,
			Namespace:       constants.NetObservNamespace,
			Size:            constants.NetObservLokiStackSize,
			SecretName:      "minio",
			StorageClassName: config.StorageClassName,
			TenantMode:      lokiv1.OpenshiftNetwork,
		}); err != nil {
			exec.SendUpdateWithError(StepTSPCreateLokiStack, executor.StatusFailed, stepName, err)
			return err
		}
		exec.SendUpdate(StepTSPCreateLokiStack, executor.StatusComplete, stepName)

		stepName = fmt.Sprintf("Wait for LokiStack %s to be ready", constants.NetObservLokiStackName)
		exec.SendUpdate(StepTSPWaitForLokiStack, executor.StatusInProgress, stepName)
		exec.SendLog(StepTSPWaitForLokiStack, fmt.Sprintf("Waiting for %s deployment in %s", constants.NetObservLokiGateway, constants.NetObservNamespace))

		if err := k8s.WaitForDeploymentReady(ctx, kubeClient, constants.NetObservLokiGateway, constants.NetObservNamespace, 0); err != nil {
			exec.SendUpdateWithError(StepTSPWaitForLokiStack, executor.StatusFailed, stepName, err)
			return err
		}
		exec.SendUpdate(StepTSPWaitForLokiStack, executor.StatusComplete, stepName)

		stepName = fmt.Sprintf("Deploy FlowCollector %s", constants.FlowCollectorName)
		exec.SendUpdate(StepTSPCreateFlowCollector, executor.StatusInProgress, stepName)
		exec.SendLog(StepTSPCreateFlowCollector, "Creating FlowCollector with LokiStack backend and eBPF agent")

		if err := resources.CreateFlowCollector(ctx, kubeClient); err != nil {
			exec.SendUpdateWithError(StepTSPCreateFlowCollector, executor.StatusFailed, stepName, err)
			return err
		}
		exec.SendUpdate(StepTSPCreateFlowCollector, executor.StatusComplete, stepName)

		stepName = "Wait for netobserv-plugin to be ready"
		exec.SendUpdate(StepTSPWaitForNetObservPlugin, executor.StatusInProgress, stepName)
		exec.SendLog(StepTSPWaitForNetObservPlugin, "Waiting for Network Observability console plugin deployment")

		logFn := func(msg string) { exec.SendLog(StepTSPWaitForNetObservPlugin, msg) }
		if err := resources.WaitForNetObservPlugin(ctx, kubeClient, logFn); err != nil {
			exec.SendUpdateWithError(StepTSPWaitForNetObservPlugin, executor.StatusFailed, stepName, err)
			return err
		}
		exec.SendUpdate(StepTSPWaitForNetObservPlugin, executor.StatusComplete, stepName)
	}

	if config.DeployUIPlugin {
		stepName = "Deploy TroubleshootingPanel UIPlugin"
		exec.SendUpdate(StepTSPDeployUIPlugin, executor.StatusInProgress, stepName)
		exec.SendLog(StepTSPDeployUIPlugin, fmt.Sprintf("Creating UIPlugin %s", constants.TroubleshootingPanelUIPluginName))

		if err := resources.CreateTroubleshootingPanelUIPlugin(ctx, kubeClient); err != nil {
			exec.SendUpdateWithError(StepTSPDeployUIPlugin, executor.StatusFailed, stepName, err)
			return err
		}
		exec.SendUpdate(StepTSPDeployUIPlugin, executor.StatusComplete, stepName)
	}

	return nil
}
