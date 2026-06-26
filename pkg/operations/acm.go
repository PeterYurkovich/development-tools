package operations

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/observability-ui/development-tools/internal/constants"
	"github.com/observability-ui/development-tools/pkg/executor"
	"github.com/observability-ui/development-tools/pkg/k8s"
	acmOperator "github.com/observability-ui/development-tools/pkg/operators/acm"
	"github.com/observability-ui/development-tools/pkg/operators"
	"github.com/observability-ui/development-tools/pkg/resources"
	"github.com/observability-ui/development-tools/pkg/storage"
)

const (
	StepACMEnsureHubNamespace = iota
	StepACMEnsureOperatorGroup
	StepACMCreateSubscription
	StepACMWaitForCSV
	StepACMEnsureObsNamespace
	StepACMDeployMinIO
	StepACMCreateMultiClusterHub
	StepACMWaitForMultiClusterHub
	StepACMCreateMCO
)

type DeployACMConfig struct {
	ACMChannel          string
	StorageClassName    string
	SkipACMInstall      bool
	SkipMultiClusterHub bool
}

func DeployACM(ctx context.Context, kubeClient client.Client,
	config DeployACMConfig, exec *executor.Executor) error {
	defer exec.Close()

	if !config.SkipACMInstall {
		stepName := fmt.Sprintf("Create namespace %s", constants.ACMNamespace)
		exec.SendUpdate(StepACMEnsureHubNamespace, executor.StatusInProgress, stepName)
		exec.SendLog(StepACMEnsureHubNamespace, "Ensuring ACM hub namespace exists")

		created, err := k8s.EnsureNamespaceWithLabels(ctx, kubeClient, constants.ACMNamespace, nil)
		if err != nil {
			exec.SendUpdateWithError(StepACMEnsureHubNamespace, executor.StatusFailed, stepName, err)
			return err
		}
		if created {
			exec.SendLog(StepACMEnsureHubNamespace, fmt.Sprintf("Namespace %s created", constants.ACMNamespace))
		} else {
			exec.SendLog(StepACMEnsureHubNamespace, fmt.Sprintf("Namespace %s already exists", constants.ACMNamespace))
		}
		exec.SendUpdate(StepACMEnsureHubNamespace, executor.StatusComplete, stepName)

		channel := config.ACMChannel
		if channel == "" {
			channel = constants.ACMDefaultChannel
		}

		if err := acmOperator.DeployViaOperatorHub(ctx, kubeClient, exec, StepACMEnsureOperatorGroup, channel); err != nil {
			return err
		}

		stepName = "Wait for advanced-cluster-management operator to be ready"
		exec.SendUpdate(StepACMWaitForCSV, executor.StatusInProgress, stepName)
		exec.SendLog(StepACMWaitForCSV, "Waiting for ACM CSV to reach Succeeded phase")

		acmSubscription, err := operators.GetSubscription(ctx, kubeClient, constants.ACMOperatorName, constants.ACMNamespace)
		if err != nil {
			exec.SendUpdateWithError(StepACMWaitForCSV, executor.StatusFailed, stepName, err)
			return err
		}

		if err := operators.WaitForCSVSucceeded(ctx, kubeClient, acmSubscription.Name, constants.ACMNamespace, 0, exec, StepACMWaitForCSV); err != nil {
			exec.SendUpdateWithError(StepACMWaitForCSV, executor.StatusFailed, stepName, err)
			return err
		}
		exec.SendUpdate(StepACMWaitForCSV, executor.StatusComplete, stepName)
	}

	stepName := fmt.Sprintf("Create namespace %s", constants.ACMObservabilityNamespace)
	exec.SendUpdate(StepACMEnsureObsNamespace, executor.StatusInProgress, stepName)
	exec.SendLog(StepACMEnsureObsNamespace, "Ensuring ACM observability namespace exists")

	created, err := k8s.EnsureNamespaceWithLabels(ctx, kubeClient, constants.ACMObservabilityNamespace, nil)
	if err != nil {
		exec.SendUpdateWithError(StepACMEnsureObsNamespace, executor.StatusFailed, stepName, err)
		return err
	}
	if created {
		exec.SendLog(StepACMEnsureObsNamespace, fmt.Sprintf("Namespace %s created", constants.ACMObservabilityNamespace))
	} else {
		exec.SendLog(StepACMEnsureObsNamespace, fmt.Sprintf("Namespace %s already exists", constants.ACMObservabilityNamespace))
	}
	exec.SendUpdate(StepACMEnsureObsNamespace, executor.StatusComplete, stepName)

	storageConfig := storage.ProviderConfig{
		Type:                   storage.StorageTypeMinio,
		Namespace:              constants.ACMObservabilityNamespace,
		BucketName:             constants.ACMMinIOBucket,
		StorageSize:            "1Gi",
		StorageClass:           config.StorageClassName,
		UseDefaultStorageClass: config.StorageClassName == "",
		SecretFormat:           storage.SecretFormatThanos,
	}

	provider, err := storage.NewProvider(storageConfig)
	if err != nil {
		return fmt.Errorf("failed to create storage provider: %w", err)
	}

	exec.SendUpdate(StepACMDeployMinIO, executor.StatusInProgress, "Deploy MinIO storage")
	exec.SendLog(StepACMDeployMinIO, fmt.Sprintf("Deploying MinIO in %s with Thanos S3 secret format", constants.ACMObservabilityNamespace))

	if _, err := provider.Deploy(ctx, kubeClient, exec); err != nil {
		exec.SendUpdateWithError(StepACMDeployMinIO, executor.StatusFailed, "Deploy MinIO storage", err)
		return err
	}
	exec.SendUpdate(StepACMDeployMinIO, executor.StatusComplete, "Deploy MinIO storage")

	if !config.SkipMultiClusterHub {
		stepName = "Create MultiClusterHub"
		exec.SendUpdate(StepACMCreateMultiClusterHub, executor.StatusInProgress, stepName)
		exec.SendLog(StepACMCreateMultiClusterHub, fmt.Sprintf("Creating MultiClusterHub %s", constants.ACMMultiClusterHubName))

		if err := resources.CreateMultiClusterHub(ctx, kubeClient); err != nil {
			exec.SendUpdateWithError(StepACMCreateMultiClusterHub, executor.StatusFailed, stepName, err)
			return err
		}
		exec.SendUpdate(StepACMCreateMultiClusterHub, executor.StatusComplete, stepName)

		stepName = "Wait for MultiClusterHub to be ready"
		exec.SendUpdate(StepACMWaitForMultiClusterHub, executor.StatusInProgress, stepName)
		exec.SendLog(StepACMWaitForMultiClusterHub, "Waiting for MultiClusterHub Complete=True (up to 15 minutes)...")

		logFn := func(msg string) {
			exec.SendLog(StepACMWaitForMultiClusterHub, msg)
		}
		if err := resources.WaitForMultiClusterHub(ctx, kubeClient, logFn); err != nil {
			exec.SendUpdateWithError(StepACMWaitForMultiClusterHub, executor.StatusFailed, stepName, err)
			return err
		}
		exec.SendUpdate(StepACMWaitForMultiClusterHub, executor.StatusComplete, stepName)
	}

	stepName = "Create MultiClusterObservability"
	exec.SendUpdate(StepACMCreateMCO, executor.StatusInProgress, stepName)
	exec.SendLog(StepACMCreateMCO, fmt.Sprintf("Creating MultiClusterObservability %s with thanos-object-storage backend", constants.ACMMCOName))

	if err := resources.CreateMultiClusterObservability(ctx, kubeClient); err != nil {
		exec.SendUpdateWithError(StepACMCreateMCO, executor.StatusFailed, stepName, err)
		return err
	}
	exec.SendUpdate(StepACMCreateMCO, executor.StatusComplete, stepName)

	return nil
}
