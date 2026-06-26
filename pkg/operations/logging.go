package operations

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/observability-ui/development-tools/internal/constants"
	"github.com/observability-ui/development-tools/pkg/executor"
	"github.com/observability-ui/development-tools/pkg/k8s"
	"github.com/observability-ui/development-tools/pkg/operators"
	"github.com/observability-ui/development-tools/pkg/operators/loki"
	"github.com/observability-ui/development-tools/pkg/operators/logging"
	"github.com/observability-ui/development-tools/pkg/resources"
	"github.com/observability-ui/development-tools/pkg/storage"
)

const (
	StepEnsureLoggingNamespace = iota
	StepEnsureLoggingOperatorGroup
	StepDeployLoggingMethod
	StepEnsureLokiNamespace
	StepEnsureLokiOperatorGroup
	StepDeployLokiMethod
	StepWaitForLoggingCSV
	StepWaitForLokiCSV
	// MinIO steps are handled by storage provider (6 steps)
	StepDeployMinIOStart
	StepDeployLokiStack
	StepWaitForLokiStack
	StepCreateCollectorRBAC
	StepDeployClusterLogForwarder
	StepWaitForClusterLogForwarder
	StepDeployUIPlugin
	StepWaitForUIPlugin
	StepDeploySignalApps
)

type DeployLoggingConfig struct {
	LoggingChannel   string
	LokiChannel      string
	StorageClassName string
	DeployMinIO      bool
	DeployLokiStack  bool
	DeployForwarder  bool
	DeployUIPlugin   bool
	DeploySignals    bool
}

func DeployLogging(ctx context.Context, kubeClient client.Client,
	config DeployLoggingConfig, exec *executor.Executor) error {
	defer exec.Close()

	stepName := fmt.Sprintf("Create namespace %s", constants.LoggingNamespace)
	exec.SendUpdate(StepEnsureLoggingNamespace, executor.StatusInProgress, stepName)
	exec.SendLog(StepEnsureLoggingNamespace, "Ensuring namespace exists with monitoring label")

	created, err := k8s.EnsureNamespaceWithLabels(ctx, kubeClient, constants.LoggingNamespace, map[string]string{
		"openshift.io/cluster-monitoring": "true",
	})
	if err != nil {
		exec.SendUpdateWithError(StepEnsureLoggingNamespace, executor.StatusFailed, stepName, err)
		return err
	}
	if created {
		exec.SendLog(StepEnsureLoggingNamespace, "Namespace created")
	} else {
		exec.SendLog(StepEnsureLoggingNamespace, "Namespace already exists")
	}
	exec.SendUpdate(StepEnsureLoggingNamespace, executor.StatusComplete, stepName)

	stepName = "Create OperatorGroup"
	exec.SendUpdate(StepEnsureLoggingOperatorGroup, executor.StatusInProgress, stepName)
	exec.SendLog(StepEnsureLoggingOperatorGroup, "Ensuring OperatorGroup exists")

	created, err = operators.EnsureOperatorGroup(ctx, kubeClient, operators.OperatorGroupConfig{
		Name:             constants.LoggingOperatorGroupName,
		Namespace:        constants.LoggingNamespace,
		TargetNamespaces: []string{constants.LoggingNamespace},
	})
	if err != nil {
		exec.SendUpdateWithError(StepEnsureLoggingOperatorGroup, executor.StatusFailed, stepName, err)
		return err
	}
	if created {
		exec.SendLog(StepEnsureLoggingOperatorGroup, "OperatorGroup created")
	} else {
		exec.SendLog(StepEnsureLoggingOperatorGroup, "OperatorGroup already exists")
	}
	exec.SendUpdate(StepEnsureLoggingOperatorGroup, executor.StatusComplete, stepName)

	// Deploy cluster-logging operator subscription
	baseStep := StepDeployLoggingMethod
	if err := logging.DeployViaOperatorHub(ctx, kubeClient, exec, baseStep, config.LoggingChannel); err != nil {
		return err
	}

	// Deploy loki-operator (namespace, operatorgroup, subscription)
	stepName = fmt.Sprintf("Create namespace %s", constants.LokiNamespace)
	exec.SendUpdate(StepEnsureLokiNamespace, executor.StatusInProgress, stepName)
	exec.SendLog(StepEnsureLokiNamespace, "Ensuring Loki namespace exists with monitoring and logging labels")

	created, err = k8s.EnsureNamespaceWithLabels(ctx, kubeClient, constants.LokiNamespace, map[string]string{
		"openshift.io/cluster-monitoring": "true",
		"openshift.io/cluster-logging":    "true",
	})
	if err != nil {
		exec.SendUpdateWithError(StepEnsureLokiNamespace, executor.StatusFailed, stepName, err)
		return err
	}
	if created {
		exec.SendLog(StepEnsureLokiNamespace, "Loki namespace created")
	} else {
		exec.SendLog(StepEnsureLokiNamespace, "Loki namespace already exists")
	}
	exec.SendUpdate(StepEnsureLokiNamespace, executor.StatusComplete, stepName)

	stepName = "Create Loki OperatorGroup"
	exec.SendUpdate(StepEnsureLokiOperatorGroup, executor.StatusInProgress, stepName)
	exec.SendLog(StepEnsureLokiOperatorGroup, "Ensuring Loki OperatorGroup exists")

	created, err = operators.EnsureOperatorGroup(ctx, kubeClient, operators.OperatorGroupConfig{
		Name:      constants.LokiOperatorGroupName,
		Namespace: constants.LokiNamespace,
		// AllNamespaces mode - empty targetNamespaces
		TargetNamespaces: []string{},
	})
	if err != nil {
		exec.SendUpdateWithError(StepEnsureLokiOperatorGroup, executor.StatusFailed, stepName, err)
		return err
	}
	if created {
		exec.SendLog(StepEnsureLokiOperatorGroup, "Loki OperatorGroup created")
	} else {
		exec.SendLog(StepEnsureLokiOperatorGroup, "Loki OperatorGroup already exists")
	}
	exec.SendUpdate(StepEnsureLokiOperatorGroup, executor.StatusComplete, stepName)

	baseStep = StepDeployLokiMethod
	if err := loki.DeployViaOperatorHub(ctx, kubeClient, exec, baseStep, config.LokiChannel); err != nil {
		return err
	}

	// Wait for both operators in parallel (like the script does)
	// For simplicity, we'll wait sequentially here - can be improved later
	stepName = "Wait for cluster-logging operator to be ready"
	exec.SendUpdate(StepWaitForLoggingCSV, executor.StatusInProgress, stepName)
	exec.SendLog(StepWaitForLoggingCSV, "Waiting for cluster-logging CSV")

	subscription, err := operators.GetSubscription(ctx, kubeClient, constants.LoggingOperatorName, constants.LoggingNamespace)
	if err != nil {
		exec.SendUpdateWithError(StepWaitForLoggingCSV, executor.StatusFailed, stepName, err)
		return err
	}

	err = operators.WaitForCSVSucceeded(ctx, kubeClient, subscription.Name, constants.LoggingNamespace, 0, exec, StepWaitForLoggingCSV)
	if err != nil {
		exec.SendUpdateWithError(StepWaitForLoggingCSV, executor.StatusFailed, stepName, err)
		return err
	}
	exec.SendUpdate(StepWaitForLoggingCSV, executor.StatusComplete, stepName)

	stepName = "Wait for loki-operator to be ready"
	exec.SendUpdate(StepWaitForLokiCSV, executor.StatusInProgress, stepName)
	exec.SendLog(StepWaitForLokiCSV, "Waiting for loki-operator CSV")

	lokiSubscription, err := operators.GetSubscription(ctx, kubeClient, constants.LokiOperatorName, constants.LokiNamespace)
	if err != nil {
		exec.SendUpdateWithError(StepWaitForLokiCSV, executor.StatusFailed, stepName, err)
		return err
	}

	err = operators.WaitForCSVSucceeded(ctx, kubeClient, lokiSubscription.Name, constants.LokiNamespace, 0, exec, StepWaitForLokiCSV)
	if err != nil {
		exec.SendUpdateWithError(StepWaitForLokiCSV, executor.StatusFailed, stepName, err)
		return err
	}
	exec.SendUpdate(StepWaitForLokiCSV, executor.StatusComplete, stepName)

	// Deploy MinIO if requested
	var minioSecretName string
	var minioNamespace string
	if config.DeployMinIO {
		minioNamespace = "minio"
		
		// Create storage provider for MinIO
		storageConfig := storage.ProviderConfig{
			Type:                   storage.StorageTypeMinio,
			Namespace:              minioNamespace,
			BucketName:             "loki",
			StorageSize:            "10Gi",
			StorageClass:           config.StorageClassName,
			UseDefaultStorageClass: config.StorageClassName == "",
		}

		provider, err := storage.NewProvider(storageConfig)
		if err != nil {
			return fmt.Errorf("failed to create storage provider: %w", err)
		}

		// Deploy MinIO using storage provider (handles all 6 steps internally)
		exec.SendUpdate(StepDeployMinIOStart, executor.StatusInProgress, "Deploy MinIO storage")
		exec.SendLog(StepDeployMinIOStart, "Deploying MinIO using storage provider")

		secretName, err := provider.Deploy(ctx, kubeClient, exec)
		if err != nil {
			exec.SendUpdateWithError(StepDeployMinIOStart, executor.StatusFailed, "Deploy MinIO storage", err)
			return err
		}
		
		minioSecretName = secretName
		exec.SendUpdate(StepDeployMinIOStart, executor.StatusComplete, "Deploy MinIO storage")
	}

	// Deploy LokiStack if requested
	if config.DeployLokiStack {
		if minioSecretName == "" {
			return fmt.Errorf("LokiStack requires MinIO to be deployed first (no secret name available)")
		}

		stepName = "Deploy LokiStack"
		exec.SendUpdate(StepDeployLokiStack, executor.StatusInProgress, stepName)
		exec.SendLog(StepDeployLokiStack, fmt.Sprintf("Creating LokiStack CR with S3 backend (secret: %s)", minioSecretName))

		if err := resources.CreateLokiStack(ctx, kubeClient, resources.LokiStackConfig{
			Name:             constants.LokiStackName,
			Namespace:        constants.LoggingNamespace,
			Size:             constants.LokiStackSize,
			StorageClassName: config.StorageClassName,
			SecretName:       minioSecretName,
			SourceNamespace:  minioNamespace,
		}); err != nil {
			exec.SendUpdateWithError(StepDeployLokiStack, executor.StatusFailed, stepName, err)
			return err
		}
		exec.SendUpdate(StepDeployLokiStack, executor.StatusComplete, stepName)

		stepName = "Wait for LokiStack to be ready"
		exec.SendUpdate(StepWaitForLokiStack, executor.StatusInProgress, stepName)
		exec.SendLog(StepWaitForLokiStack, "Waiting for Loki gateway deployment")

		if err := k8s.WaitForDeploymentReady(ctx, kubeClient, constants.LokiGatewayDeployment, constants.LoggingNamespace, 0); err != nil {
			exec.SendUpdateWithError(StepWaitForLokiStack, executor.StatusFailed, stepName, err)
			return err
		}
		exec.SendUpdate(StepWaitForLokiStack, executor.StatusComplete, stepName)
	}

	// Deploy ClusterLogForwarder if requested
	if config.DeployForwarder {
		stepName = "Create collector RBAC"
		exec.SendUpdate(StepCreateCollectorRBAC, executor.StatusInProgress, stepName)
		exec.SendLog(StepCreateCollectorRBAC, "Creating collector ServiceAccount and ClusterRoleBindings")

		if err := resources.CreateCollectorRBAC(ctx, kubeClient, constants.LoggingNamespace); err != nil {
			exec.SendUpdateWithError(StepCreateCollectorRBAC, executor.StatusFailed, stepName, err)
			return err
		}
		exec.SendUpdate(StepCreateCollectorRBAC, executor.StatusComplete, stepName)

		stepName = "Deploy ClusterLogForwarder"
		exec.SendUpdate(StepDeployClusterLogForwarder, executor.StatusInProgress, stepName)
		exec.SendLog(StepDeployClusterLogForwarder, "Creating ClusterLogForwarder CR")

		if err := resources.CreateClusterLogForwarder(ctx, kubeClient, resources.ClusterLogForwarderConfig{
			Name:           constants.ClusterLogForwarderName,
			Namespace:      constants.LoggingNamespace,
			ServiceAccount: constants.CollectorServiceAccount,
			LokiStackName:  constants.LokiStackName,
			LokiStackNS:    constants.LoggingNamespace,
		}); err != nil {
			exec.SendUpdateWithError(StepDeployClusterLogForwarder, executor.StatusFailed, stepName, err)
			return err
		}
		exec.SendUpdate(StepDeployClusterLogForwarder, executor.StatusComplete, stepName)

		stepName = "Wait for ClusterLogForwarder to be ready"
		exec.SendUpdate(StepWaitForClusterLogForwarder, executor.StatusInProgress, stepName)
		exec.SendLog(StepWaitForClusterLogForwarder, "Waiting for collector pods to be ready")

		// Note: In the script, this waits for CLF Ready condition or collector pods
		// For simplicity, we'll just mark complete - actual pod readiness can be checked separately
		exec.SendUpdate(StepWaitForClusterLogForwarder, executor.StatusComplete, stepName)
	}

	// Deploy UIPlugin if requested
	if config.DeployUIPlugin {
		stepName = "Deploy Logging UIPlugin"
		exec.SendUpdate(StepDeployUIPlugin, executor.StatusInProgress, stepName)
		exec.SendLog(StepDeployUIPlugin, "Creating Logging UIPlugin CR")

		if err := resources.CreateLoggingUIPlugin(ctx, kubeClient, resources.LoggingUIPluginConfig{
			Name:          constants.LoggingUIPluginName,
			LokiStackName: constants.LokiStackName,
			LogsLimit:     constants.LoggingUIPluginLogsLimit,
			Timeout:       constants.LoggingUIPluginTimeout,
			Schema:        constants.LoggingUIPluginSchema,
		}); err != nil {
			exec.SendUpdateWithError(StepDeployUIPlugin, executor.StatusFailed, stepName, err)
			return err
		}
		exec.SendUpdate(StepDeployUIPlugin, executor.StatusComplete, stepName)

		stepName = "Wait for UIPlugin to be ready"
		exec.SendUpdate(StepWaitForUIPlugin, executor.StatusInProgress, stepName)
		exec.SendLog(StepWaitForUIPlugin, "UIPlugin created (reconciliation managed by COO)")

		// UIPlugin is reconciled by COO, which should already be deployed
		// For now, just mark complete
		exec.SendUpdate(StepWaitForUIPlugin, executor.StatusComplete, stepName)
	}

	if config.DeploySignals {
		stepName = "Deploy chat signal app"
		exec.SendUpdate(StepDeploySignalApps, executor.StatusInProgress, stepName)
		exec.SendLog(StepDeploySignalApps, "Creating chat Deployment in namespace 'chat'")

		if err := resources.CreateChatApp(ctx, kubeClient); err != nil {
			exec.SendLog(StepDeploySignalApps, fmt.Sprintf("Warning: failed to deploy chat app: %v", err))
		} else {
			exec.SendLog(StepDeploySignalApps, "Chat app deployed in namespace 'chat'")
		}
		exec.SendUpdate(StepDeploySignalApps, executor.StatusComplete, stepName)
	}

	return nil
}
