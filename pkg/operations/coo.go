package operations

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/observability-ui/development-tools/internal/constants"
	"github.com/observability-ui/development-tools/pkg/executor"
	"github.com/observability-ui/development-tools/pkg/k8s"
	"github.com/observability-ui/development-tools/pkg/operators"
	"github.com/observability-ui/development-tools/pkg/operators/coo"
)

const (
	StepEnsureNamespace = iota
	StepEnsureOperatorGroup
	StepDeployMethod
	StepWaitForCSV
)

type DeployCOOConfig struct {
	Method       string
	BundleURL    string
	FBCURL       string
	Channel      string
	RegistryType string
}

func DeployCOO(ctx context.Context, kubeClient client.Client,
	config DeployCOOConfig, exec *executor.Executor) error {
	defer exec.Close()

	stepName := fmt.Sprintf("Create namespace %s", constants.COONamespace)
	exec.SendUpdate(StepEnsureNamespace, executor.StatusInProgress, stepName)
	exec.SendLog(StepEnsureNamespace, "Ensuring namespace exists with monitoring label")

	created, err := k8s.EnsureNamespaceWithLabels(ctx, kubeClient, constants.COONamespace, map[string]string{
		"openshift.io/cluster-monitoring": "true",
	})
	if err != nil {
		exec.SendUpdateWithError(StepEnsureNamespace, executor.StatusFailed, stepName, err)
		return err
	}
	if created {
		exec.SendLog(StepEnsureNamespace, "Namespace created")
	} else {
		exec.SendLog(StepEnsureNamespace, "Namespace already exists")
	}
	exec.SendUpdate(StepEnsureNamespace, executor.StatusComplete, stepName)

	stepName = "Create OperatorGroup"
	exec.SendUpdate(StepEnsureOperatorGroup, executor.StatusInProgress, stepName)
	exec.SendLog(StepEnsureOperatorGroup, "Ensuring OperatorGroup exists")

	created, err = operators.EnsureOperatorGroup(ctx, kubeClient, operators.OperatorGroupConfig{
		Name:             constants.COOOperatorGroupName,
		Namespace:        constants.COONamespace,
		TargetNamespaces: []string{constants.COONamespace},
	})
	if err != nil {
		exec.SendUpdateWithError(StepEnsureOperatorGroup, executor.StatusFailed, stepName, err)
		return err
	}
	if created {
		exec.SendLog(StepEnsureOperatorGroup, "OperatorGroup created")
	} else {
		exec.SendLog(StepEnsureOperatorGroup, "OperatorGroup already exists")
	}
	exec.SendUpdate(StepEnsureOperatorGroup, executor.StatusComplete, stepName)

	switch config.Method {
	case "bundle":
		err = coo.DeployBundle(ctx, kubeClient, config.BundleURL, config.RegistryType, exec, StepDeployMethod)
	case "fbc":
		err = coo.DeployFBC(ctx, kubeClient, config.FBCURL, exec, StepDeployMethod)
	case "stage":
		err = coo.DeployStage(ctx, kubeClient, config.FBCURL, exec, StepDeployMethod)
	case "operatorhub":
		err = coo.DeployOperatorHub(ctx, kubeClient, config.Channel, exec, StepDeployMethod)
	default:
		err = fmt.Errorf("unknown deployment method: %s", config.Method)
	}

	if err != nil {
		return err
	}

	stepName = "Wait for operator to be ready"
	exec.SendUpdate(StepWaitForCSV, executor.StatusInProgress, stepName)
	exec.SendLog(StepWaitForCSV, "Waiting for ClusterServiceVersion to reach Succeeded phase")

	subscription, err := operators.GetSubscription(ctx, kubeClient, constants.COOOperatorName, constants.COONamespace)
	if err != nil {
		exec.SendUpdateWithError(StepWaitForCSV, executor.StatusFailed, stepName, err)
		return err
	}

	err = operators.WaitForCSVSucceeded(ctx, kubeClient, subscription.Name, constants.COONamespace, 0, exec, StepWaitForCSV)
	if err != nil {
		exec.SendUpdateWithError(StepWaitForCSV, executor.StatusFailed, stepName, err)
		return err
	}
	exec.SendUpdate(StepWaitForCSV, executor.StatusComplete, stepName)

	return nil
}
