package coo

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/observability-ui/development-tools/internal/constants"
	"github.com/observability-ui/development-tools/pkg/executor"
	"github.com/observability-ui/development-tools/pkg/operators"
)

const (
	StepOperatorHubCreateSubscription = iota
)

func DeployOperatorHub(ctx context.Context, kubeClient client.Client,
	channel string, exec *executor.Executor, baseStep int) error {

	stepName := "Create Subscription to OperatorHub"
	exec.SendUpdate(baseStep+StepOperatorHubCreateSubscription, executor.StatusInProgress, stepName)
	exec.SendLog(baseStep+StepOperatorHubCreateSubscription, fmt.Sprintf("Creating subscription with channel: %s", channel))

	err := operators.CreateSubscription(ctx, kubeClient, operators.SubscriptionConfig{
		Name:             constants.COOOperatorName,
		Namespace:        constants.COONamespace,
		Channel:          channel,
		PackageName:      constants.COOPackageName,
		CatalogSource:    "redhat-operators",
		CatalogNamespace: "openshift-marketplace",
	})
	if err != nil {
		exec.SendUpdateWithError(baseStep+StepOperatorHubCreateSubscription, executor.StatusFailed, stepName, err)
		return fmt.Errorf("failed to create Subscription: %w", err)
	}

	exec.SendLog(baseStep+StepOperatorHubCreateSubscription, "OperatorHub subscription created successfully")
	exec.SendUpdate(baseStep+StepOperatorHubCreateSubscription, executor.StatusComplete, stepName)
	return nil
}
