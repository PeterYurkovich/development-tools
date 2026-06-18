package operators

import (
	"context"
	"fmt"
	"time"

	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/observability-ui/development-tools/pkg/executor"
)

const (
	DefaultCSVTimeout = 10 * time.Minute
	CSVPollInterval   = 5 * time.Second

	OpenShiftOperatorsNamespace   = "openshift-operators"
	OpenShiftMarketplaceNamespace = "openshift-marketplace"
)

func WaitForCSVSucceeded(
	ctx context.Context,
	kubeClient client.Client,
	csvName string,
	namespace string,
	timeout time.Duration,
	exec *executor.Executor,
	stepIndex int,
) error {
	if timeout == 0 {
		timeout = DefaultCSVTimeout
	}

	if exec != nil {
		exec.SendUpdate(stepIndex, executor.StatusInProgress, fmt.Sprintf("Waiting for CSV %s", csvName))
	}

	timeoutChan := time.After(timeout)
	ticker := time.NewTicker(CSVPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			err := fmt.Errorf("context cancelled while waiting for CSV %s/%s", namespace, csvName)
			if exec != nil {
				exec.SendUpdateWithError(stepIndex, executor.StatusFailed, fmt.Sprintf("Waiting for CSV %s", csvName), err)
			}
			return err

		case <-timeoutChan:
			err := fmt.Errorf("timeout waiting for CSV %s/%s to succeed after %v", namespace, csvName, timeout)
			if exec != nil {
				exec.SendUpdateWithError(stepIndex, executor.StatusFailed, fmt.Sprintf("Waiting for CSV %s", csvName), err)
			}
			return err

		case <-ticker.C:
			csv, err := GetCSV(ctx, kubeClient, csvName, namespace)
			if err != nil {
				if exec != nil {
					exec.SendLog(stepIndex, fmt.Sprintf("Error checking CSV: %v", err))
				}
				continue
			}

			phase := csv.Status.Phase
			reason := csv.Status.Reason
			message := csv.Status.Message

			switch phase {
			case operatorsv1alpha1.CSVPhaseSucceeded:
				if exec != nil {
					exec.SendUpdate(stepIndex, executor.StatusComplete, fmt.Sprintf("Waiting for CSV %s", csvName))
				}
				return nil

			case operatorsv1alpha1.CSVPhaseFailed:
				err := fmt.Errorf("CSV %s/%s failed: %s - %s", namespace, csvName, reason, message)
				if exec != nil {
					exec.SendUpdateWithError(stepIndex, executor.StatusFailed, fmt.Sprintf("Waiting for CSV %s", csvName), err)
				}
				return err

			default:
				if exec != nil {
					nextCheckSeconds := int(CSVPollInterval.Seconds())
					statusMsg := fmt.Sprintf("CSV phase: %s", phase)
					if reason != "" {
						statusMsg = fmt.Sprintf("%s (reason: %s)", statusMsg, reason)
					}
					exec.SendLog(stepIndex, fmt.Sprintf("%s, next check in %d seconds", statusMsg, nextCheckSeconds))
				}
			}
		}
	}
}

func GetCSV(ctx context.Context, kubeClient client.Client, name, namespace string) (*operatorsv1alpha1.ClusterServiceVersion, error) {
	csv := &operatorsv1alpha1.ClusterServiceVersion{}
	key := client.ObjectKey{Name: name, Namespace: namespace}

	err := kubeClient.Get(ctx, key, csv)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, fmt.Errorf("CSV %s/%s not found", namespace, name)
		}
		return nil, fmt.Errorf("failed to get CSV %s/%s: %w", namespace, name, err)
	}

	return csv, nil
}

func DeleteCSV(ctx context.Context, kubeClient client.Client, name, namespace string) error {
	csv := &operatorsv1alpha1.ClusterServiceVersion{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	err := kubeClient.Delete(ctx, csv)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete CSV %s/%s: %w", namespace, name, err)
	}

	return nil
}

func ListCSVsByPackage(ctx context.Context, kubeClient client.Client, packageName, namespace string) (*operatorsv1alpha1.ClusterServiceVersionList, error) {
	csvList := &operatorsv1alpha1.ClusterServiceVersionList{}

	listOptions := &client.ListOptions{
		Namespace: namespace,
	}

	err := kubeClient.List(ctx, csvList, listOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list CSVs in namespace %s: %w", namespace, err)
	}

	filteredList := &operatorsv1alpha1.ClusterServiceVersionList{}
	for _, csv := range csvList.Items {
		if csv.Spec.DisplayName == packageName || csv.Labels["operators.coreos.com/packageName"] == packageName {
			filteredList.Items = append(filteredList.Items, csv)
		}
	}

	return filteredList, nil
}

func GetCSVPhase(ctx context.Context, kubeClient client.Client, name, namespace string) (operatorsv1alpha1.ClusterServiceVersionPhase, error) {
	csv, err := GetCSV(ctx, kubeClient, name, namespace)
	if err != nil {
		return "", err
	}

	return csv.Status.Phase, nil
}
