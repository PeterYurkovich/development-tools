package operations

import (
	"context"
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/observability-ui/development-tools/internal/constants"
	"github.com/observability-ui/development-tools/pkg/executor"
	"github.com/observability-ui/development-tools/pkg/k8s"
	"github.com/observability-ui/development-tools/pkg/operators"
	"github.com/observability-ui/development-tools/pkg/resources"
	"github.com/observability-ui/development-tools/pkg/storage"
)

const (
	StepDeleteLoggingUIPlugin = iota
	StepDeleteLoggingClusterLogForwarder
	StepDeleteLoggingLokiStack
	StepDeleteLoggingMinIO
	StepDeleteLoggingCollectorRBAC
	StepDeleteLoggingOperatorSubscription
	StepDeleteLoggingOperatorCSV
	StepDeleteLokiOperatorSubscription
	StepDeleteLokiOperatorCSV
	StepDeleteLoggingNamespaces
)

type CleanupLoggingConfig struct {
	DeleteOperators  bool
	DeleteNamespaces bool
	DeleteMinIO      bool
	Confirm          bool
}

func CleanupLogging(ctx context.Context, kubeClient client.Client,
	config CleanupLoggingConfig, exec *executor.Executor) error {
	defer exec.Close()

	stepName := "Delete Logging UIPlugin"
	exec.SendUpdate(StepDeleteLoggingUIPlugin, executor.StatusInProgress, stepName)
	exec.SendLog(StepDeleteLoggingUIPlugin, fmt.Sprintf("Deleting UIPlugin: %s", constants.LoggingUIPluginName))

	if err := resources.DeleteUIPlugin(ctx, kubeClient, constants.LoggingUIPluginName); err != nil {
		exec.SendLog(StepDeleteLoggingUIPlugin, fmt.Sprintf("Error deleting UIPlugin: %v", err))
	} else {
		exec.SendLog(StepDeleteLoggingUIPlugin, "UIPlugin deleted (or already absent)")
	}
	exec.SendUpdate(StepDeleteLoggingUIPlugin, executor.StatusComplete, stepName)

	stepName = "Delete ClusterLogForwarder"
	exec.SendUpdate(StepDeleteLoggingClusterLogForwarder, executor.StatusInProgress, stepName)
	exec.SendLog(StepDeleteLoggingClusterLogForwarder, fmt.Sprintf("Deleting ClusterLogForwarder: %s", constants.ClusterLogForwarderName))

	if err := resources.DeleteClusterLogForwarder(ctx, kubeClient, constants.ClusterLogForwarderName, constants.LoggingNamespace); err != nil {
		exec.SendLog(StepDeleteLoggingClusterLogForwarder, fmt.Sprintf("Error deleting ClusterLogForwarder: %v", err))
	} else {
		exec.SendLog(StepDeleteLoggingClusterLogForwarder, "ClusterLogForwarder deleted (or already absent)")
	}
	exec.SendUpdate(StepDeleteLoggingClusterLogForwarder, executor.StatusComplete, stepName)

	stepName = "Delete LokiStack"
	exec.SendUpdate(StepDeleteLoggingLokiStack, executor.StatusInProgress, stepName)
	exec.SendLog(StepDeleteLoggingLokiStack, fmt.Sprintf("Deleting LokiStack: %s", constants.LokiStackName))

	if err := resources.DeleteLokiStack(ctx, kubeClient, constants.LokiStackName, constants.LoggingNamespace); err != nil {
		exec.SendLog(StepDeleteLoggingLokiStack, fmt.Sprintf("Error deleting LokiStack: %v", err))
	} else {
		exec.SendLog(StepDeleteLoggingLokiStack, "LokiStack deleted (or already absent)")
	}
	exec.SendUpdate(StepDeleteLoggingLokiStack, executor.StatusComplete, stepName)

	if config.DeleteMinIO {
		stepName = "Delete MinIO resources"
		exec.SendUpdate(StepDeleteLoggingMinIO, executor.StatusInProgress, stepName)
		exec.SendLog(StepDeleteLoggingMinIO, "Deleting MinIO deployment, service, PVC, and secrets")

		minioNamespace := "minio"
		storageConfig := storage.ProviderConfig{
			Type:                   storage.StorageTypeMinio,
			Namespace:              minioNamespace,
			BucketName:             "loki",
			StorageSize:            "10Gi",
			UseDefaultStorageClass: true,
		}

		provider, err := storage.NewProvider(storageConfig)
		if err != nil {
			exec.SendLog(StepDeleteLoggingMinIO, fmt.Sprintf("Error creating storage provider: %v", err))
		} else {
			if err := provider.Delete(ctx, kubeClient, exec); err != nil {
				exec.SendLog(StepDeleteLoggingMinIO, fmt.Sprintf("Error deleting MinIO: %v", err))
			} else {
				exec.SendLog(StepDeleteLoggingMinIO, "MinIO resources deleted")
			}
		}
		exec.SendUpdate(StepDeleteLoggingMinIO, executor.StatusComplete, stepName)
	}

	stepName = "Delete collector RBAC"
	exec.SendUpdate(StepDeleteLoggingCollectorRBAC, executor.StatusInProgress, stepName)
	exec.SendLog(StepDeleteLoggingCollectorRBAC, "Deleting collector ServiceAccount and ClusterRoleBindings")

	if err := resources.DeleteServiceAccount(ctx, kubeClient, constants.CollectorServiceAccount, constants.LoggingNamespace); err != nil {
		exec.SendLog(StepDeleteLoggingCollectorRBAC, fmt.Sprintf("Error deleting ServiceAccount: %v", err))
	} else {
		exec.SendLog(StepDeleteLoggingCollectorRBAC, "ServiceAccount deleted")
	}

	rbacRoles := []string{
		"logging-collector-logs-writer",
		"collect-application-logs",
		"collect-infrastructure-logs",
		"collect-audit-logs",
	}
	for _, roleName := range rbacRoles {
		crbName := fmt.Sprintf("%s-%s", constants.CollectorServiceAccount, roleName)
		if err := deleteClusterRoleBinding(ctx, kubeClient, crbName); err != nil {
			exec.SendLog(StepDeleteLoggingCollectorRBAC, fmt.Sprintf("Error deleting ClusterRoleBinding %s: %v", crbName, err))
		} else {
			exec.SendLog(StepDeleteLoggingCollectorRBAC, fmt.Sprintf("ClusterRoleBinding %s deleted", crbName))
		}
	}

	exec.SendUpdate(StepDeleteLoggingCollectorRBAC, executor.StatusComplete, stepName)

	if config.DeleteOperators {
		var loggingCSV, lokiCSV string

		stepName = "Delete logging operator Subscription"
		exec.SendUpdate(StepDeleteLoggingOperatorSubscription, executor.StatusInProgress, stepName)
		exec.SendLog(StepDeleteLoggingOperatorSubscription, fmt.Sprintf("Looking for subscription: %s", constants.LoggingOperatorName))

		subscription, err := operators.GetSubscription(ctx, kubeClient, constants.LoggingOperatorName, constants.LoggingNamespace)
		if err != nil {
			if errors.IsNotFound(err) {
				exec.SendLog(StepDeleteLoggingOperatorSubscription, "Subscription not found")
			} else {
				exec.SendLog(StepDeleteLoggingOperatorSubscription, fmt.Sprintf("Error getting subscription: %v", err))
			}
		} else {
			loggingCSV = subscription.Status.InstalledCSV
			exec.SendLog(StepDeleteLoggingOperatorSubscription, fmt.Sprintf("Found subscription with CSV: %s", loggingCSV))
			if err := operators.DeleteSubscription(ctx, kubeClient, constants.LoggingOperatorName, constants.LoggingNamespace); err != nil {
				exec.SendLog(StepDeleteLoggingOperatorSubscription, fmt.Sprintf("Error deleting subscription: %v", err))
			} else {
				exec.SendLog(StepDeleteLoggingOperatorSubscription, "Subscription deleted")
			}
		}
		exec.SendUpdate(StepDeleteLoggingOperatorSubscription, executor.StatusComplete, stepName)

		stepName = "Delete logging operator CSV"
		exec.SendUpdate(StepDeleteLoggingOperatorCSV, executor.StatusInProgress, stepName)
		if loggingCSV != "" {
			exec.SendLog(StepDeleteLoggingOperatorCSV, fmt.Sprintf("Deleting CSV: %s", loggingCSV))
			if err := operators.DeleteCSV(ctx, kubeClient, loggingCSV, constants.LoggingNamespace); err != nil {
				exec.SendLog(StepDeleteLoggingOperatorCSV, fmt.Sprintf("Error deleting CSV: %v", err))
			} else {
				exec.SendLog(StepDeleteLoggingOperatorCSV, "CSV deleted")
			}
		} else {
			exec.SendLog(StepDeleteLoggingOperatorCSV, "No CSV to delete")
		}
		exec.SendUpdate(StepDeleteLoggingOperatorCSV, executor.StatusComplete, stepName)

		stepName = "Delete loki operator Subscription"
		exec.SendUpdate(StepDeleteLokiOperatorSubscription, executor.StatusInProgress, stepName)
		exec.SendLog(StepDeleteLokiOperatorSubscription, fmt.Sprintf("Looking for subscription: %s", constants.LokiOperatorName))

		lokiSubscription, err := operators.GetSubscription(ctx, kubeClient, constants.LokiOperatorName, constants.LokiNamespace)
		if err != nil {
			if errors.IsNotFound(err) {
				exec.SendLog(StepDeleteLokiOperatorSubscription, "Subscription not found")
			} else {
				exec.SendLog(StepDeleteLokiOperatorSubscription, fmt.Sprintf("Error getting subscription: %v", err))
			}
		} else {
			lokiCSV = lokiSubscription.Status.InstalledCSV
			exec.SendLog(StepDeleteLokiOperatorSubscription, fmt.Sprintf("Found subscription with CSV: %s", lokiCSV))
			if err := operators.DeleteSubscription(ctx, kubeClient, constants.LokiOperatorName, constants.LokiNamespace); err != nil {
				exec.SendLog(StepDeleteLokiOperatorSubscription, fmt.Sprintf("Error deleting subscription: %v", err))
			} else {
				exec.SendLog(StepDeleteLokiOperatorSubscription, "Subscription deleted")
			}
		}
		exec.SendUpdate(StepDeleteLokiOperatorSubscription, executor.StatusComplete, stepName)

		stepName = "Delete loki operator CSV"
		exec.SendUpdate(StepDeleteLokiOperatorCSV, executor.StatusInProgress, stepName)
		if lokiCSV != "" {
			exec.SendLog(StepDeleteLokiOperatorCSV, fmt.Sprintf("Deleting CSV: %s", lokiCSV))
			if err := operators.DeleteCSV(ctx, kubeClient, lokiCSV, constants.LokiNamespace); err != nil {
				exec.SendLog(StepDeleteLokiOperatorCSV, fmt.Sprintf("Error deleting CSV: %v", err))
			} else {
				exec.SendLog(StepDeleteLokiOperatorCSV, "CSV deleted")
			}
		} else {
			exec.SendLog(StepDeleteLokiOperatorCSV, "No CSV to delete")
		}
		exec.SendUpdate(StepDeleteLokiOperatorCSV, executor.StatusComplete, stepName)
	}

	if config.DeleteNamespaces {
		stepName = "Delete namespaces"
		exec.SendUpdate(StepDeleteLoggingNamespaces, executor.StatusInProgress, stepName)

		for _, ns := range []string{constants.LoggingNamespace, constants.LokiNamespace, "minio"} {
			exec.SendLog(StepDeleteLoggingNamespaces, fmt.Sprintf("Deleting namespace: %s", ns))
			if err := k8s.DeleteNamespace(ctx, kubeClient, ns); err != nil {
				exec.SendLog(StepDeleteLoggingNamespaces, fmt.Sprintf("Error deleting namespace %s: %v", ns, err))
			} else {
				exec.SendLog(StepDeleteLoggingNamespaces, fmt.Sprintf("Namespace %s deleted", ns))
			}
		}

		exec.SendUpdate(StepDeleteLoggingNamespaces, executor.StatusComplete, stepName)
	}

	return nil
}

func deleteClusterRoleBinding(ctx context.Context, kubeClient client.Client, name string) error {
	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	err := kubeClient.Delete(ctx, crb)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	return nil
}
