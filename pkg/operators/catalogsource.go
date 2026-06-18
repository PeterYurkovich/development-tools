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
	DefaultCatalogSourceTimeout = 5 * time.Minute
	CatalogSourcePollInterval   = 5 * time.Second
)

type CatalogSourceConfig struct {
	Name        string
	Namespace   string
	DisplayName string
	Publisher   string
	SourceType  string
	Image       string
	Secrets     []string
}

func CreateCatalogSource(ctx context.Context, kubeClient client.Client, config CatalogSourceConfig) error {
	catalogSource := &operatorsv1alpha1.CatalogSource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.Name,
			Namespace: config.Namespace,
			Labels: map[string]string{
				"obstool.observability.openshift.io/managed": "true",
			},
		},
		Spec: operatorsv1alpha1.CatalogSourceSpec{
			SourceType:  operatorsv1alpha1.SourceType(config.SourceType),
			Image:       config.Image,
			DisplayName: config.DisplayName,
			Publisher:   config.Publisher,
		},
	}

	if len(config.Secrets) > 0 {
		catalogSource.Spec.Secrets = config.Secrets
	}

	err := kubeClient.Create(ctx, catalogSource)
	if err != nil {
		return fmt.Errorf("failed to create catalogsource %s/%s: %w", config.Namespace, config.Name, err)
	}

	return nil
}

func GetCatalogSource(ctx context.Context, kubeClient client.Client, name, namespace string) (*operatorsv1alpha1.CatalogSource, error) {
	catalogSource := &operatorsv1alpha1.CatalogSource{}
	key := client.ObjectKey{Name: name, Namespace: namespace}

	err := kubeClient.Get(ctx, key, catalogSource)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, fmt.Errorf("catalogsource %s/%s not found", namespace, name)
		}
		return nil, fmt.Errorf("failed to get catalogsource %s/%s: %w", namespace, name, err)
	}

	return catalogSource, nil
}

func DeleteCatalogSource(ctx context.Context, kubeClient client.Client, name, namespace string) error {
	catalogSource := &operatorsv1alpha1.CatalogSource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	err := kubeClient.Delete(ctx, catalogSource)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete catalogsource %s/%s: %w", namespace, name, err)
	}

	return nil
}

func WaitForCatalogSourceReady(
	ctx context.Context,
	kubeClient client.Client,
	name string,
	namespace string,
	timeout time.Duration,
	exec *executor.Executor,
	stepIndex int,
) error {
	if timeout == 0 {
		timeout = DefaultCatalogSourceTimeout
	}

	timeoutChan := time.After(timeout)
	ticker := time.NewTicker(CatalogSourcePollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled while waiting for catalogsource %s/%s", namespace, name)

		case <-timeoutChan:
			return fmt.Errorf("timeout waiting for catalogsource %s/%s to be ready after %v", namespace, name, timeout)

		case <-ticker.C:
			catalogSource, err := GetCatalogSource(ctx, kubeClient, name, namespace)
			if err != nil {
				if exec != nil {
					exec.SendLog(stepIndex, fmt.Sprintf("Error checking catalogsource: %v", err))
				}
				continue
			}

			state := catalogSource.Status.GRPCConnectionState
			if state != nil && state.LastObservedState == "READY" {
				return nil
			}

			if exec != nil {
				remainingTime := time.Until(time.Now().Add(CatalogSourcePollInterval))
				exec.SendLog(stepIndex, fmt.Sprintf("CatalogSource not ready (state: %s), next check in %d seconds",
					getConnectionState(state), int(remainingTime.Seconds())))
			}
		}
	}
}

func getConnectionState(state *operatorsv1alpha1.GRPCConnectionState) string {
	if state == nil {
		return "unknown"
	}
	if state.LastObservedState == "" {
		return "initializing"
	}
	return state.LastObservedState
}
