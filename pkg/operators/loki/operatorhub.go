package loki

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/observability-ui/development-tools/internal/constants"
	"github.com/observability-ui/development-tools/pkg/executor"
	"github.com/observability-ui/development-tools/pkg/operators"
)

const (
	StepCreateSubscription = iota
)

func DeployViaOperatorHub(ctx context.Context, kubeClient client.Client,
	exec *executor.Executor, baseStep int, channel string) error {

	stepName := "Create Subscription to OperatorHub"
	exec.SendUpdate(baseStep+StepCreateSubscription, executor.StatusInProgress, stepName)

	if err := operators.CreateSubscription(ctx, kubeClient, operators.SubscriptionConfig{
		Name:             constants.LokiOperatorName,
		Namespace:        constants.LokiNamespace,
		PackageName:      constants.LokiPackageName,
		Channel:          channel,
		CatalogSource:    constants.LokiCatalogSource,
		CatalogNamespace: constants.MarketplaceNamespace,
	}); err != nil {
		exec.SendUpdateWithError(baseStep+StepCreateSubscription, executor.StatusFailed, stepName, err)
		return err
	}

	exec.SendUpdate(baseStep+StepCreateSubscription, executor.StatusComplete, stepName)
	return nil
}
