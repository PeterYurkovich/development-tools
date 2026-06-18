package operations

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/observability-ui/development-tools/internal/constants"
	"github.com/observability-ui/development-tools/pkg/executor"
	"github.com/observability-ui/development-tools/pkg/k8s"
	"github.com/observability-ui/development-tools/pkg/operators"
)

const (
	StepDeleteCOOSubscription = iota
	StepDeleteCOOCSV
	StepDeleteCOOCatalogSource
	StepDeleteCOOIDMS
	StepDeleteCOOOperatorGroup
	StepDeleteCOONamespace
)

type CleanupCOOConfig struct {
	DeleteNamespace     bool
	DeleteOperatorGroup bool
	Confirm             bool
}

func CleanupCOO(ctx context.Context, kubeClient client.Client,
	config CleanupCOOConfig, exec *executor.Executor) error {
	defer exec.Close()

	var csvName string

	stepName := "Delete Subscription"
	exec.SendUpdate(StepDeleteCOOSubscription, executor.StatusInProgress, stepName)
	exec.SendLog(StepDeleteCOOSubscription, fmt.Sprintf("Looking for subscription %s in namespace %s", constants.COOOperatorName, constants.COONamespace))

	subscription, err := operators.GetSubscription(ctx, kubeClient, constants.COOOperatorName, constants.COONamespace)
	if err != nil {
		if errors.IsNotFound(err) {
			exec.SendLog(StepDeleteCOOSubscription, "Subscription not found (already deleted or never created)")
			exec.SendUpdate(StepDeleteCOOSubscription, executor.StatusComplete, stepName)
		} else {
			exec.SendUpdateWithError(StepDeleteCOOSubscription, executor.StatusFailed, stepName, err)
			return err
		}
	} else {
		csvName = subscription.Status.InstalledCSV
		exec.SendLog(StepDeleteCOOSubscription, fmt.Sprintf("Found subscription with CSV: %s", csvName))

		if err := operators.DeleteSubscription(ctx, kubeClient, constants.COOOperatorName, constants.COONamespace); err != nil {
			exec.SendUpdateWithError(StepDeleteCOOSubscription, executor.StatusFailed, stepName, err)
			return err
		}
		exec.SendLog(StepDeleteCOOSubscription, "Subscription deleted")
		exec.SendUpdate(StepDeleteCOOSubscription, executor.StatusComplete, stepName)
	}

	stepName = "Delete ClusterServiceVersion"
	exec.SendUpdate(StepDeleteCOOCSV, executor.StatusInProgress, stepName)

	if csvName != "" {
		exec.SendLog(StepDeleteCOOCSV, fmt.Sprintf("Deleting CSV: %s", csvName))
		if err := operators.DeleteCSV(ctx, kubeClient, csvName, constants.COONamespace); err != nil {
			exec.SendUpdateWithError(StepDeleteCOOCSV, executor.StatusFailed, stepName, err)
			return err
		}
		exec.SendLog(StepDeleteCOOCSV, "CSV deleted")
	} else {
		exec.SendLog(StepDeleteCOOCSV, "No CSV to delete")
	}
	exec.SendUpdate(StepDeleteCOOCSV, executor.StatusComplete, stepName)

	stepName = "Delete CatalogSource"
	exec.SendUpdate(StepDeleteCOOCatalogSource, executor.StatusInProgress, stepName)
	exec.SendLog(StepDeleteCOOCatalogSource, fmt.Sprintf("Checking for CatalogSource: %s", constants.COOCatalogName))

	_, err = operators.GetCatalogSource(ctx, kubeClient, constants.COOCatalogName, constants.MarketplaceNamespace)
	if err != nil {
		if errors.IsNotFound(err) {
			exec.SendLog(StepDeleteCOOCatalogSource, "CatalogSource not found (may have used OperatorHub)")
			exec.SendUpdate(StepDeleteCOOCatalogSource, executor.StatusComplete, stepName)
		} else {
			exec.SendUpdateWithError(StepDeleteCOOCatalogSource, executor.StatusFailed, stepName, err)
			return err
		}
	} else {
		if err := operators.DeleteCatalogSource(ctx, kubeClient, constants.COOCatalogName, constants.MarketplaceNamespace); err != nil {
			exec.SendUpdateWithError(StepDeleteCOOCatalogSource, executor.StatusFailed, stepName, err)
			return err
		}
		exec.SendLog(StepDeleteCOOCatalogSource, "CatalogSource deleted")
		exec.SendUpdate(StepDeleteCOOCatalogSource, executor.StatusComplete, stepName)
	}

	stepName = "Delete ImageDigestMirrorSets"
	exec.SendUpdate(StepDeleteCOOIDMS, executor.StatusInProgress, stepName)
	
	idmsDeleted := false
	for _, idmsName := range []string{constants.IDMSCOOQuay, constants.IDMSCOOStage} {
		exec.SendLog(StepDeleteCOOIDMS, fmt.Sprintf("Checking for IDMS: %s", idmsName))
		exists, err := operators.ImageDigestMirrorSetExists(ctx, kubeClient, idmsName)
		if err != nil {
			exec.SendLog(StepDeleteCOOIDMS, fmt.Sprintf("Error checking IDMS %s: %v", idmsName, err))
			continue
		}
		
		if exists {
			exec.SendLog(StepDeleteCOOIDMS, fmt.Sprintf("Deleting IDMS: %s", idmsName))
			if err := operators.DeleteImageDigestMirrorSet(ctx, kubeClient, idmsName); err != nil {
				exec.SendLog(StepDeleteCOOIDMS, fmt.Sprintf("Error deleting IDMS %s: %v", idmsName, err))
			} else {
				exec.SendLog(StepDeleteCOOIDMS, fmt.Sprintf("Deleted IDMS: %s", idmsName))
				idmsDeleted = true
			}
		} else {
			exec.SendLog(StepDeleteCOOIDMS, fmt.Sprintf("IDMS %s not found", idmsName))
		}
	}
	
	if idmsDeleted {
		exec.SendLog(StepDeleteCOOIDMS, "One or more IDMS deleted")
	} else {
		exec.SendLog(StepDeleteCOOIDMS, "No IDMS found to delete")
	}
	exec.SendUpdate(StepDeleteCOOIDMS, executor.StatusComplete, stepName)

	if config.DeleteOperatorGroup {
		stepName = "Delete OperatorGroup"
		exec.SendUpdate(StepDeleteCOOOperatorGroup, executor.StatusInProgress, stepName)
		exec.SendLog(StepDeleteCOOOperatorGroup, fmt.Sprintf("Deleting OperatorGroup: %s", constants.COOOperatorGroupName))

		if err := operators.DeleteOperatorGroup(ctx, kubeClient, constants.COOOperatorGroupName, constants.COONamespace); err != nil {
			exec.SendLog(StepDeleteCOOOperatorGroup, fmt.Sprintf("Error deleting OperatorGroup: %v", err))
		} else {
			exec.SendLog(StepDeleteCOOOperatorGroup, "OperatorGroup deleted")
		}
		exec.SendUpdate(StepDeleteCOOOperatorGroup, executor.StatusComplete, stepName)
	}

	if config.DeleteNamespace {
		stepName = "Delete Namespace"
		exec.SendUpdate(StepDeleteCOONamespace, executor.StatusInProgress, stepName)
		exec.SendLog(StepDeleteCOONamespace, fmt.Sprintf("Deleting namespace: %s", constants.COONamespace))

		if err := k8s.DeleteNamespace(ctx, kubeClient, constants.COONamespace); err != nil {
			exec.SendUpdateWithError(StepDeleteCOONamespace, executor.StatusFailed, stepName, err)
			return err
		}
		exec.SendLog(StepDeleteCOONamespace, "Namespace deleted")
		exec.SendUpdate(StepDeleteCOONamespace, executor.StatusComplete, stepName)
	}

	return nil
}
