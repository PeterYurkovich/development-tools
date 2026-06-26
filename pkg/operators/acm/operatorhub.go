package acm

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/observability-ui/development-tools/internal/constants"
	"github.com/observability-ui/development-tools/pkg/executor"
	"github.com/observability-ui/development-tools/pkg/operators"
)

const (
	StepCreateOperatorGroup = iota
	StepCreateSubscription
)

func DeployViaOperatorHub(ctx context.Context, kubeClient client.Client,
	exec *executor.Executor, baseStep int, channel string) error {

	stepName := "Create OperatorGroup og-global"
	exec.SendUpdate(baseStep+StepCreateOperatorGroup, executor.StatusInProgress, stepName)
	exec.SendLog(baseStep+StepCreateOperatorGroup, "Ensuring OperatorGroup exists in open-cluster-management namespace")

	created, err := operators.EnsureOperatorGroup(ctx, kubeClient, operators.OperatorGroupConfig{
		Name:             constants.ACMOperatorGroup,
		Namespace:        constants.ACMNamespace,
		TargetNamespaces: []string{constants.ACMNamespace},
	})
	if err != nil {
		exec.SendUpdateWithError(baseStep+StepCreateOperatorGroup, executor.StatusFailed, stepName, err)
		return err
	}
	if created {
		exec.SendLog(baseStep+StepCreateOperatorGroup, "OperatorGroup created")
	} else {
		exec.SendLog(baseStep+StepCreateOperatorGroup, "OperatorGroup already exists")
	}
	exec.SendUpdate(baseStep+StepCreateOperatorGroup, executor.StatusComplete, stepName)

	stepName = "Create Subscription for advanced-cluster-management"
	exec.SendUpdate(baseStep+StepCreateSubscription, executor.StatusInProgress, stepName)
	exec.SendLog(baseStep+StepCreateSubscription, "Creating ACM subscription from redhat-operators catalog")

	if err := operators.CreateSubscription(ctx, kubeClient, operators.SubscriptionConfig{
		Name:             constants.ACMOperatorName,
		Namespace:        constants.ACMNamespace,
		PackageName:      constants.ACMOperatorName,
		Channel:          channel,
		CatalogSource:    constants.ACMCatalogSource,
		CatalogNamespace: constants.MarketplaceNamespace,
	}); err != nil {
		exec.SendUpdateWithError(baseStep+StepCreateSubscription, executor.StatusFailed, stepName, err)
		return err
	}
	exec.SendUpdate(baseStep+StepCreateSubscription, executor.StatusComplete, stepName)

	return nil
}
