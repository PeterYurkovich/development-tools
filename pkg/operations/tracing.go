package operations

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/observability-ui/development-tools/internal/constants"
	"github.com/observability-ui/development-tools/pkg/executor"
	"github.com/observability-ui/development-tools/pkg/k8s"
	"github.com/observability-ui/development-tools/pkg/operators"
	"github.com/observability-ui/development-tools/pkg/operators/otel"
	"github.com/observability-ui/development-tools/pkg/operators/tempo"
	"github.com/observability-ui/development-tools/pkg/resources"
	"github.com/observability-ui/development-tools/pkg/storage"
)

const (
	StepTracingEnsureTracingNamespace = iota
	StepTracingEnsureTempoOperatorNamespace
	StepTracingEnsureTempoOperatorGroup
	StepTracingDeployTempoOperator
	StepTracingEnsureOTelOperatorNamespace
	StepTracingEnsureOTelOperatorGroup
	StepTracingDeployOTelOperator
	StepTracingWaitForTempoCSV
	StepTracingWaitForOTelCSV
	StepTracingDeployMinIOStart
	StepTracingDeployTempoStack
	StepTracingWaitForTempoStack
	StepTracingEnableUserWorkloadMonitoring
	StepTracingCreateTraceReaderRBAC
	StepTracingDeployPlatformCollector
	StepTracingWaitForPlatformCollector
	StepTracingDeployUserCollector
	StepTracingWaitForUserCollector
	StepTracingDeploySignalApps
	StepTracingDeployUIPlugin
)

type DeployTracingConfig struct {
	TempoChannel                string
	OTelChannel                 string
	StorageClassName            string
	DeployMinIO                 bool
	DeployTempoStack            bool
	EnableUserWorkloadMonitoring bool
	DeployCollectors            bool
	DeploySignals               bool
	DeployUIPlugin              bool
}

func DeployTracing(ctx context.Context, kubeClient client.Client,
	config DeployTracingConfig, exec *executor.Executor) error {
	defer exec.Close()

	stepName := fmt.Sprintf("Create namespace %s", constants.TracingNamespace)
	exec.SendUpdate(StepTracingEnsureTracingNamespace, executor.StatusInProgress, stepName)
	exec.SendLog(StepTracingEnsureTracingNamespace, "Ensuring tracing namespace exists")

	created, err := k8s.EnsureNamespaceWithLabels(ctx, kubeClient, constants.TracingNamespace, nil)
	if err != nil {
		exec.SendUpdateWithError(StepTracingEnsureTracingNamespace, executor.StatusFailed, stepName, err)
		return err
	}
	if created {
		exec.SendLog(StepTracingEnsureTracingNamespace, "Tracing namespace created")
	} else {
		exec.SendLog(StepTracingEnsureTracingNamespace, "Tracing namespace already exists")
	}
	exec.SendUpdate(StepTracingEnsureTracingNamespace, executor.StatusComplete, stepName)

	stepName = fmt.Sprintf("Create namespace %s", constants.TempoOperatorNS)
	exec.SendUpdate(StepTracingEnsureTempoOperatorNamespace, executor.StatusInProgress, stepName)
	exec.SendLog(StepTracingEnsureTempoOperatorNamespace, "Ensuring Tempo operator namespace exists")

	created, err = k8s.EnsureNamespaceWithLabels(ctx, kubeClient, constants.TempoOperatorNS, nil)
	if err != nil {
		exec.SendUpdateWithError(StepTracingEnsureTempoOperatorNamespace, executor.StatusFailed, stepName, err)
		return err
	}
	if created {
		exec.SendLog(StepTracingEnsureTempoOperatorNamespace, "Tempo operator namespace created")
	} else {
		exec.SendLog(StepTracingEnsureTempoOperatorNamespace, "Tempo operator namespace already exists")
	}
	exec.SendUpdate(StepTracingEnsureTempoOperatorNamespace, executor.StatusComplete, stepName)

	stepName = "Create Tempo OperatorGroup"
	exec.SendUpdate(StepTracingEnsureTempoOperatorGroup, executor.StatusInProgress, stepName)
	exec.SendLog(StepTracingEnsureTempoOperatorGroup, "Ensuring Tempo OperatorGroup exists")

	created, err = operators.EnsureOperatorGroup(ctx, kubeClient, operators.OperatorGroupConfig{
		Name:             constants.TempoOperatorGroup,
		Namespace:        constants.TempoOperatorNS,
		TargetNamespaces: []string{},
	})
	if err != nil {
		exec.SendUpdateWithError(StepTracingEnsureTempoOperatorGroup, executor.StatusFailed, stepName, err)
		return err
	}
	if created {
		exec.SendLog(StepTracingEnsureTempoOperatorGroup, "Tempo OperatorGroup created")
	} else {
		exec.SendLog(StepTracingEnsureTempoOperatorGroup, "Tempo OperatorGroup already exists")
	}
	exec.SendUpdate(StepTracingEnsureTempoOperatorGroup, executor.StatusComplete, stepName)

	if err := tempo.DeployViaOperatorHub(ctx, kubeClient, exec, StepTracingDeployTempoOperator, config.TempoChannel); err != nil {
		return err
	}

	stepName = fmt.Sprintf("Create namespace %s", constants.OTelOperatorNS)
	exec.SendUpdate(StepTracingEnsureOTelOperatorNamespace, executor.StatusInProgress, stepName)
	exec.SendLog(StepTracingEnsureOTelOperatorNamespace, "Ensuring OpenTelemetry operator namespace exists")

	created, err = k8s.EnsureNamespaceWithLabels(ctx, kubeClient, constants.OTelOperatorNS, nil)
	if err != nil {
		exec.SendUpdateWithError(StepTracingEnsureOTelOperatorNamespace, executor.StatusFailed, stepName, err)
		return err
	}
	if created {
		exec.SendLog(StepTracingEnsureOTelOperatorNamespace, "OpenTelemetry operator namespace created")
	} else {
		exec.SendLog(StepTracingEnsureOTelOperatorNamespace, "OpenTelemetry operator namespace already exists")
	}
	exec.SendUpdate(StepTracingEnsureOTelOperatorNamespace, executor.StatusComplete, stepName)

	stepName = "Create OpenTelemetry OperatorGroup"
	exec.SendUpdate(StepTracingEnsureOTelOperatorGroup, executor.StatusInProgress, stepName)
	exec.SendLog(StepTracingEnsureOTelOperatorGroup, "Ensuring OpenTelemetry OperatorGroup exists")

	created, err = operators.EnsureOperatorGroup(ctx, kubeClient, operators.OperatorGroupConfig{
		Name:             constants.OTelOperatorGroup,
		Namespace:        constants.OTelOperatorNS,
		TargetNamespaces: []string{},
	})
	if err != nil {
		exec.SendUpdateWithError(StepTracingEnsureOTelOperatorGroup, executor.StatusFailed, stepName, err)
		return err
	}
	if created {
		exec.SendLog(StepTracingEnsureOTelOperatorGroup, "OpenTelemetry OperatorGroup created")
	} else {
		exec.SendLog(StepTracingEnsureOTelOperatorGroup, "OpenTelemetry OperatorGroup already exists")
	}
	exec.SendUpdate(StepTracingEnsureOTelOperatorGroup, executor.StatusComplete, stepName)

	if err := otel.DeployViaOperatorHub(ctx, kubeClient, exec, StepTracingDeployOTelOperator, config.OTelChannel); err != nil {
		return err
	}

	stepName = "Wait for tempo-product operator to be ready"
	exec.SendUpdate(StepTracingWaitForTempoCSV, executor.StatusInProgress, stepName)
	exec.SendLog(StepTracingWaitForTempoCSV, "Waiting for Tempo operator CSV")

	tempoSubscription, err := operators.GetSubscription(ctx, kubeClient, constants.TempoOperatorName, constants.TempoOperatorNS)
	if err != nil {
		exec.SendUpdateWithError(StepTracingWaitForTempoCSV, executor.StatusFailed, stepName, err)
		return err
	}

	if err := operators.WaitForCSVSucceeded(ctx, kubeClient, tempoSubscription.Name, constants.TempoOperatorNS, 0, exec, StepTracingWaitForTempoCSV); err != nil {
		exec.SendUpdateWithError(StepTracingWaitForTempoCSV, executor.StatusFailed, stepName, err)
		return err
	}
	exec.SendUpdate(StepTracingWaitForTempoCSV, executor.StatusComplete, stepName)

	stepName = "Wait for opentelemetry-product operator to be ready"
	exec.SendUpdate(StepTracingWaitForOTelCSV, executor.StatusInProgress, stepName)
	exec.SendLog(StepTracingWaitForOTelCSV, "Waiting for OpenTelemetry operator CSV")

	otelSubscription, err := operators.GetSubscription(ctx, kubeClient, constants.OTelOperatorName, constants.OTelOperatorNS)
	if err != nil {
		exec.SendUpdateWithError(StepTracingWaitForOTelCSV, executor.StatusFailed, stepName, err)
		return err
	}

	if err := operators.WaitForCSVSucceeded(ctx, kubeClient, otelSubscription.Name, constants.OTelOperatorNS, 0, exec, StepTracingWaitForOTelCSV); err != nil {
		exec.SendUpdateWithError(StepTracingWaitForOTelCSV, executor.StatusFailed, stepName, err)
		return err
	}
	exec.SendUpdate(StepTracingWaitForOTelCSV, executor.StatusComplete, stepName)

	var minioSecretName string
	if config.DeployMinIO {
		storageConfig := storage.ProviderConfig{
			Type:                   storage.StorageTypeMinio,
			Namespace:              constants.TracingNamespace,
			BucketName:             "tempo",
			StorageSize:            "10Gi",
			StorageClass:           config.StorageClassName,
			UseDefaultStorageClass: config.StorageClassName == "",
		}

		provider, err := storage.NewProvider(storageConfig)
		if err != nil {
			return fmt.Errorf("failed to create storage provider: %w", err)
		}

		exec.SendUpdate(StepTracingDeployMinIOStart, executor.StatusInProgress, "Deploy MinIO storage")
		exec.SendLog(StepTracingDeployMinIOStart, "Deploying MinIO in openshift-tracing namespace")

		secretName, err := provider.Deploy(ctx, kubeClient, exec)
		if err != nil {
			exec.SendUpdateWithError(StepTracingDeployMinIOStart, executor.StatusFailed, "Deploy MinIO storage", err)
			return err
		}

		minioSecretName = secretName
		exec.SendUpdate(StepTracingDeployMinIOStart, executor.StatusComplete, "Deploy MinIO storage")
	}

	if config.DeployTempoStack {
		if minioSecretName == "" {
			return fmt.Errorf("TempoStack requires MinIO to be deployed first (use --deploy-minio)")
		}

		stepName = "Deploy TempoStack"
		exec.SendUpdate(StepTracingDeployTempoStack, executor.StatusInProgress, stepName)
		exec.SendLog(StepTracingDeployTempoStack, fmt.Sprintf("Creating TempoStack CR with S3 backend (secret: %s)", minioSecretName))

		if err := resources.CreateTempoStack(ctx, kubeClient, resources.TempoStackConfig{
			Name:            constants.TempoStackName,
			Namespace:       constants.TracingNamespace,
			StorageSize:     constants.TempoStackSize,
			SecretName:      minioSecretName,
			SourceNamespace: constants.TracingNamespace,
		}); err != nil {
			exec.SendUpdateWithError(StepTracingDeployTempoStack, executor.StatusFailed, stepName, err)
			return err
		}
		exec.SendUpdate(StepTracingDeployTempoStack, executor.StatusComplete, stepName)

		stepName = "Wait for TempoStack to be ready"
		exec.SendUpdate(StepTracingWaitForTempoStack, executor.StatusInProgress, stepName)
		exec.SendLog(StepTracingWaitForTempoStack, "Waiting for Tempo gateway deployment")

		if err := k8s.WaitForDeploymentReady(ctx, kubeClient, "tempo-platform-gateway", constants.TracingNamespace, 0); err != nil {
			exec.SendUpdateWithError(StepTracingWaitForTempoStack, executor.StatusFailed, stepName, err)
			return err
		}
		exec.SendUpdate(StepTracingWaitForTempoStack, executor.StatusComplete, stepName)
	}

	if config.EnableUserWorkloadMonitoring {
		stepName = "Enable user workload monitoring"
		exec.SendUpdate(StepTracingEnableUserWorkloadMonitoring, executor.StatusInProgress, stepName)
		exec.SendLog(StepTracingEnableUserWorkloadMonitoring, "Patching cluster-monitoring-config to enable user workload monitoring")

		if err := enableUserWorkloadMonitoring(ctx, kubeClient); err != nil {
			exec.SendLog(StepTracingEnableUserWorkloadMonitoring, fmt.Sprintf("Warning: failed to enable user workload monitoring (may require cluster-admin): %v", err))
		} else {
			exec.SendLog(StepTracingEnableUserWorkloadMonitoring, "User workload monitoring enabled")
		}
		exec.SendUpdate(StepTracingEnableUserWorkloadMonitoring, executor.StatusComplete, stepName)
	}

	if config.DeployTempoStack {
		stepName = "Create trace reader RBAC"
		exec.SendUpdate(StepTracingCreateTraceReaderRBAC, executor.StatusInProgress, stepName)
		exec.SendLog(StepTracingCreateTraceReaderRBAC, "Creating ClusterRoles and ClusterRoleBindings for Tempo tenants")

		if err := resources.CreateTracingRBAC(ctx, kubeClient, constants.TracingNamespace); err != nil {
			exec.SendUpdateWithError(StepTracingCreateTraceReaderRBAC, executor.StatusFailed, stepName, err)
			return err
		}
		exec.SendUpdate(StepTracingCreateTraceReaderRBAC, executor.StatusComplete, stepName)
	}

	if config.DeployCollectors {
		stepName = "Deploy platform OpenTelemetry Collector"
		exec.SendUpdate(StepTracingDeployPlatformCollector, executor.StatusInProgress, stepName)
		exec.SendLog(StepTracingDeployPlatformCollector, "Creating platform OTelCollector CR and RBAC")

		if err := resources.CreatePlatformCollector(ctx, kubeClient, constants.TracingNamespace); err != nil {
			exec.SendUpdateWithError(StepTracingDeployPlatformCollector, executor.StatusFailed, stepName, err)
			return err
		}
		exec.SendUpdate(StepTracingDeployPlatformCollector, executor.StatusComplete, stepName)

		stepName = "Wait for platform-collector to be ready"
		exec.SendUpdate(StepTracingWaitForPlatformCollector, executor.StatusInProgress, stepName)
		exec.SendLog(StepTracingWaitForPlatformCollector, "Waiting for platform-collector deployment")

		if err := k8s.WaitForDeploymentReady(ctx, kubeClient, constants.PlatformCollectorDeployment, constants.TracingNamespace, 0); err != nil {
			exec.SendUpdateWithError(StepTracingWaitForPlatformCollector, executor.StatusFailed, stepName, err)
			return err
		}
		exec.SendUpdate(StepTracingWaitForPlatformCollector, executor.StatusComplete, stepName)

		stepName = "Deploy user OpenTelemetry Collector"
		exec.SendUpdate(StepTracingDeployUserCollector, executor.StatusInProgress, stepName)
		exec.SendLog(StepTracingDeployUserCollector, "Creating user OTelCollector CR and RBAC")

		if err := resources.CreateUserCollector(ctx, kubeClient, constants.TracingNamespace); err != nil {
			exec.SendUpdateWithError(StepTracingDeployUserCollector, executor.StatusFailed, stepName, err)
			return err
		}
		exec.SendUpdate(StepTracingDeployUserCollector, executor.StatusComplete, stepName)

		stepName = "Wait for user-collector to be ready"
		exec.SendUpdate(StepTracingWaitForUserCollector, executor.StatusInProgress, stepName)
		exec.SendLog(StepTracingWaitForUserCollector, "Waiting for user-collector deployment")

		if err := k8s.WaitForDeploymentReady(ctx, kubeClient, constants.UserCollectorDeployment, constants.TracingNamespace, 0); err != nil {
			exec.SendUpdateWithError(StepTracingWaitForUserCollector, executor.StatusFailed, stepName, err)
			return err
		}
		exec.SendUpdate(StepTracingWaitForUserCollector, executor.StatusComplete, stepName)
	}

	if config.DeploySignals {
		stepName = "Deploy signal generator apps"
		exec.SendUpdate(StepTracingDeploySignalApps, executor.StatusInProgress, stepName)
		exec.SendLog(StepTracingDeploySignalApps, "Deploying hotrod, k6-tracing, and telemetrygen")

		if err := resources.CreateHotrodApp(ctx, kubeClient); err != nil {
			exec.SendLog(StepTracingDeploySignalApps, fmt.Sprintf("Warning: failed to deploy hotrod: %v", err))
		} else {
			exec.SendLog(StepTracingDeploySignalApps, "hotrod deployed in tracing-app-hotrod")
		}

		if err := resources.CreateK6TracingApp(ctx, kubeClient); err != nil {
			exec.SendLog(StepTracingDeploySignalApps, fmt.Sprintf("Warning: failed to deploy k6-tracing: %v", err))
		} else {
			exec.SendLog(StepTracingDeploySignalApps, "k6-tracing deployed in tracing-app-k6")
		}

		if err := resources.CreateTelemetrygenApp(ctx, kubeClient); err != nil {
			exec.SendLog(StepTracingDeploySignalApps, fmt.Sprintf("Warning: failed to deploy telemetrygen: %v", err))
		} else {
			exec.SendLog(StepTracingDeploySignalApps, "telemetrygen deployed in tracing-app-telemetrygen")
		}

		exec.SendUpdate(StepTracingDeploySignalApps, executor.StatusComplete, stepName)
	}

	if config.DeployUIPlugin {
		stepName = "Deploy Distributed Tracing UIPlugin"
		exec.SendUpdate(StepTracingDeployUIPlugin, executor.StatusInProgress, stepName)
		exec.SendLog(StepTracingDeployUIPlugin, fmt.Sprintf("Creating UIPlugin %s", constants.TracingUIPluginName))

		if err := resources.CreateTracingUIPlugin(ctx, kubeClient); err != nil {
			exec.SendUpdateWithError(StepTracingDeployUIPlugin, executor.StatusFailed, stepName, err)
			return err
		}
		exec.SendUpdate(StepTracingDeployUIPlugin, executor.StatusComplete, stepName)
	}

	return nil
}

func enableUserWorkloadMonitoring(ctx context.Context, kubeClient client.Client) error {
	configMap := &corev1.ConfigMap{}
	key := client.ObjectKey{Name: "cluster-monitoring-config", Namespace: "openshift-monitoring"}

	err := kubeClient.Get(ctx, key, configMap)
	if err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("failed to get cluster-monitoring-config: %w", err)
		}

		configMap = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cluster-monitoring-config",
				Namespace: "openshift-monitoring",
			},
			Data: map[string]string{
				"config.yaml": "enableUserWorkload: true\n",
			},
		}
		if err := kubeClient.Create(ctx, configMap); err != nil {
			return fmt.Errorf("failed to create cluster-monitoring-config: %w", err)
		}
		return nil
	}

	if configMap.Data == nil {
		configMap.Data = map[string]string{}
	}
	configMap.Data["config.yaml"] = "enableUserWorkload: true\n"

	if err := kubeClient.Update(ctx, configMap); err != nil {
		return fmt.Errorf("failed to update cluster-monitoring-config: %w", err)
	}

	return nil
}
