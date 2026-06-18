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
	StepFBCCreateIDMS = iota
	StepFBCCreateCatalogSource
	StepFBCCreateSubscription
)

func DeployFBC(ctx context.Context, kubeClient client.Client,
	fbcURL string, exec *executor.Executor, baseStep int) error {

	stepName := "Create ImageDigestMirrorSet for Quay"
	exec.SendUpdate(baseStep+StepFBCCreateIDMS, executor.StatusInProgress, stepName)
	exec.SendLog(baseStep+StepFBCCreateIDMS, "Ensuring IDMS for quay registry")

	err := operators.EnsureIDMSQuay(ctx, kubeClient)
	if err != nil {
		exec.SendUpdateWithError(baseStep+StepFBCCreateIDMS, executor.StatusFailed, stepName, err)
		return fmt.Errorf("failed to create IDMS: %w", err)
	}

	exec.SendUpdate(baseStep+StepFBCCreateIDMS, executor.StatusComplete, stepName)

	stepName = "Create CatalogSource"
	exec.SendUpdate(baseStep+StepFBCCreateCatalogSource, executor.StatusInProgress, stepName)
	exec.SendLog(baseStep+StepFBCCreateCatalogSource, fmt.Sprintf("Creating catalog with image: %s", fbcURL))

	err = operators.CreateCatalogSource(ctx, kubeClient, operators.CatalogSourceConfig{
		Name:        constants.COOCatalogName,
		Namespace:   constants.MarketplaceNamespace,
		Image:       fbcURL,
		DisplayName: "COO FBC Catalog",
	})
	if err != nil {
		exec.SendUpdateWithError(baseStep+StepFBCCreateCatalogSource, executor.StatusFailed, stepName, err)
		return fmt.Errorf("failed to create CatalogSource: %w", err)
	}

	exec.SendUpdate(baseStep+StepFBCCreateCatalogSource, executor.StatusComplete, stepName)

	stepName = "Create Subscription"
	exec.SendUpdate(baseStep+StepFBCCreateSubscription, executor.StatusInProgress, stepName)
	exec.SendLog(baseStep+StepFBCCreateSubscription, "Creating subscription to FBC catalog")

	err = operators.CreateSubscription(ctx, kubeClient, operators.SubscriptionConfig{
		Name:             constants.COOOperatorName,
		Namespace:        constants.COONamespace,
		Channel:          constants.DefaultCOOChannel,
		PackageName:      constants.COOPackageName,
		CatalogSource:    constants.COOCatalogName,
		CatalogNamespace: constants.MarketplaceNamespace,
	})
	if err != nil {
		exec.SendUpdateWithError(baseStep+StepFBCCreateSubscription, executor.StatusFailed, stepName, err)
		return fmt.Errorf("failed to create Subscription: %w", err)
	}

	exec.SendLog(baseStep+StepFBCCreateSubscription, "FBC deployment initiated successfully")
	exec.SendUpdate(baseStep+StepFBCCreateSubscription, executor.StatusComplete, stepName)
	return nil
}
