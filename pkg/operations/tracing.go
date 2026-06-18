package operations

import (
	"context"
	"fmt"

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
)

type DeployTracingConfig struct {
	TempoChannel     string
	OTelChannel      string
	StorageClassName string
	DeployMinIO      bool
	DeployTempoStack bool
}

func DeployTracing(ctx context.Context, kubeClient client.Client,
	config DeployTracingConfig, exec *executor.Executor) error {
	defer exec.Close()

	// Create tracing namespace (where TempoStack and collectors will run)
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

	// Deploy Tempo Operator
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

	baseStep := StepTracingDeployTempoOperator
	if err := tempo.DeployViaOperatorHub(ctx, kubeClient, exec, baseStep, config.TempoChannel); err != nil {
		return err
	}

	// Deploy OpenTelemetry Operator
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

	baseStep = StepTracingDeployOTelOperator
	if err := otel.DeployViaOperatorHub(ctx, kubeClient, exec, baseStep, config.OTelChannel); err != nil {
		return err
	}

	// Wait for both operators
	stepName = "Wait for tempo-product operator to be ready"
	exec.SendUpdate(StepTracingWaitForTempoCSV, executor.StatusInProgress, stepName)
	exec.SendLog(StepTracingWaitForTempoCSV, "Waiting for Tempo operator CSV")

	tempoSubscription, err := operators.GetSubscription(ctx, kubeClient, constants.TempoOperatorName, constants.TempoOperatorNS)
	if err != nil {
		exec.SendUpdateWithError(StepTracingWaitForTempoCSV, executor.StatusFailed, stepName, err)
		return err
	}

	err = operators.WaitForCSVSucceeded(ctx, kubeClient, tempoSubscription.Name, constants.TempoOperatorNS, 0, exec, StepTracingWaitForTempoCSV)
	if err != nil {
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

	err = operators.WaitForCSVSucceeded(ctx, kubeClient, otelSubscription.Name, constants.OTelOperatorNS, 0, exec, StepTracingWaitForOTelCSV)
	if err != nil {
		exec.SendUpdateWithError(StepTracingWaitForOTelCSV, executor.StatusFailed, stepName, err)
		return err
	}
	exec.SendUpdate(StepTracingWaitForOTelCSV, executor.StatusComplete, stepName)

	// Deploy MinIO if requested
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

	// Deploy TempoStack if requested
	if config.DeployTempoStack {
		if minioSecretName == "" {
			return fmt.Errorf("TempoStack requires MinIO to be deployed first (no secret name available)")
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

		// TempoStack creates a gateway deployment
		if err := k8s.WaitForDeploymentReady(ctx, kubeClient, "tempo-platform-gateway", constants.TracingNamespace, 0); err != nil {
			exec.SendUpdateWithError(StepTracingWaitForTempoStack, executor.StatusFailed, stepName, err)
			return err
		}
		exec.SendUpdate(StepTracingWaitForTempoStack, executor.StatusComplete, stepName)
	}

	return nil
}
