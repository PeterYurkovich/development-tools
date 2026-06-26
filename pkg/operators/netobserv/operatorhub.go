package netobserv

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

	stepName := "Create OperatorGroup openshift-netobserv-operator-hack"
	exec.SendUpdate(baseStep+StepCreateOperatorGroup, executor.StatusInProgress, stepName)
	exec.SendLog(baseStep+StepCreateOperatorGroup, "Ensuring OperatorGroup exists in openshift-netobserv-operator namespace")

	created, err := operators.EnsureOperatorGroup(ctx, kubeClient, operators.OperatorGroupConfig{
		Name:             constants.NetObservOperatorGroup,
		Namespace:        constants.NetObservOperatorNamespace,
		TargetNamespaces: []string{},
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

	stepName = "Create Subscription for netobserv-operator"
	exec.SendUpdate(baseStep+StepCreateSubscription, executor.StatusInProgress, stepName)
	exec.SendLog(baseStep+StepCreateSubscription, "Creating Network Observability subscription from redhat-operators catalog")

	if err := operators.CreateSubscription(ctx, kubeClient, operators.SubscriptionConfig{
		Name:             constants.NetObservOperatorName,
		Namespace:        constants.NetObservOperatorNamespace,
		PackageName:      constants.NetObservOperatorName,
		Channel:          channel,
		CatalogSource:    constants.NetObservCatalogSource,
		CatalogNamespace: constants.MarketplaceNamespace,
	}); err != nil {
		exec.SendUpdateWithError(baseStep+StepCreateSubscription, executor.StatusFailed, stepName, err)
		return err
	}
	exec.SendUpdate(baseStep+StepCreateSubscription, executor.StatusComplete, stepName)

	return nil
}
