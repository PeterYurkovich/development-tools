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
	StepStageCreateIDMS = iota
	StepStageCreateCatalogSource
	StepStageCreateSubscription
)

func DeployStage(ctx context.Context, kubeClient client.Client,
	fbcURL string, exec *executor.Executor, baseStep int) error {

	stepName := "Create ImageDigestMirrorSet for Stage"
	exec.SendUpdate(baseStep+StepStageCreateIDMS, executor.StatusInProgress, stepName)
	exec.SendLog(baseStep+StepStageCreateIDMS, "Creating IDMS for stage registry (includes brew)")

	err := operators.EnsureIDMSStageWithBrew(ctx, kubeClient)
	if err != nil {
		exec.SendUpdateWithError(baseStep+StepStageCreateIDMS, executor.StatusFailed, stepName, err)
		return fmt.Errorf("failed to create IDMS: %w", err)
	}

	exec.SendUpdate(baseStep+StepStageCreateIDMS, executor.StatusComplete, stepName)

	stepName = "Create CatalogSource"
	exec.SendUpdate(baseStep+StepStageCreateCatalogSource, executor.StatusInProgress, stepName)
	exec.SendLog(baseStep+StepStageCreateCatalogSource, fmt.Sprintf("Creating stage catalog with image: %s", fbcURL))

	err = operators.CreateCatalogSource(ctx, kubeClient, operators.CatalogSourceConfig{
		Name:        constants.COOCatalogName,
		Namespace:   constants.MarketplaceNamespace,
		Image:       fbcURL,
		DisplayName: "COO Stage Catalog",
	})
	if err != nil {
		exec.SendUpdateWithError(baseStep+StepStageCreateCatalogSource, executor.StatusFailed, stepName, err)
		return fmt.Errorf("failed to create CatalogSource: %w", err)
	}

	exec.SendUpdate(baseStep+StepStageCreateCatalogSource, executor.StatusComplete, stepName)

	stepName = "Create Subscription"
	exec.SendUpdate(baseStep+StepStageCreateSubscription, executor.StatusInProgress, stepName)
	exec.SendLog(baseStep+StepStageCreateSubscription, "Creating subscription to stage catalog")

	err = operators.CreateSubscription(ctx, kubeClient, operators.SubscriptionConfig{
		Name:             constants.COOOperatorName,
		Namespace:        constants.COONamespace,
		Channel:          constants.DefaultCOOChannel,
		PackageName:      constants.COOPackageName,
		CatalogSource:    constants.COOCatalogName,
		CatalogNamespace: constants.MarketplaceNamespace,
	})
	if err != nil {
		exec.SendUpdateWithError(baseStep+StepStageCreateSubscription, executor.StatusFailed, stepName, err)
		return fmt.Errorf("failed to create Subscription: %w", err)
	}

	exec.SendLog(baseStep+StepStageCreateSubscription, "Stage deployment initiated successfully")
	exec.SendUpdate(baseStep+StepStageCreateSubscription, executor.StatusComplete, stepName)
	return nil
}
