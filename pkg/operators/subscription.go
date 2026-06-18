package operators

import (
	"context"
	"fmt"

	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ApprovalAutomatic = "Automatic"
	ApprovalManual    = "Manual"
)

type SubscriptionConfig struct {
	Name             string
	Namespace        string
	Channel          string
	PackageName      string
	CatalogSource    string
	CatalogNamespace string
	StartingCSV      string
	InstallMode      string
}

func CreateSubscription(ctx context.Context, kubeClient client.Client, config SubscriptionConfig) error {
	subscription := &operatorsv1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.Name,
			Namespace: config.Namespace,
			Labels: map[string]string{
				"operators.coreos.com/package": config.PackageName,
			},
		},
		Spec: &operatorsv1alpha1.SubscriptionSpec{
			Channel:                config.Channel,
			Package:                config.PackageName,
			CatalogSource:          config.CatalogSource,
			CatalogSourceNamespace: config.CatalogNamespace,
			InstallPlanApproval:    operatorsv1alpha1.Approval(ApprovalAutomatic),
		},
	}

	if config.StartingCSV != "" {
		subscription.Spec.StartingCSV = config.StartingCSV
	}

	err := kubeClient.Create(ctx, subscription)
	if err != nil {
		return fmt.Errorf("failed to create subscription %s/%s: %w", config.Namespace, config.Name, err)
	}

	return nil
}

func GetSubscription(ctx context.Context, kubeClient client.Client, name, namespace string) (*operatorsv1alpha1.Subscription, error) {
	subscription := &operatorsv1alpha1.Subscription{}
	key := client.ObjectKey{Name: name, Namespace: namespace}

	err := kubeClient.Get(ctx, key, subscription)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, fmt.Errorf("subscription %s/%s not found", namespace, name)
		}
		return nil, fmt.Errorf("failed to get subscription %s/%s: %w", namespace, name, err)
	}

	return subscription, nil
}

func DeleteSubscription(ctx context.Context, kubeClient client.Client, name, namespace string) error {
	subscription := &operatorsv1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	err := kubeClient.Delete(ctx, subscription)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete subscription %s/%s: %w", namespace, name, err)
	}

	return nil
}

func GetSubscriptionCSV(ctx context.Context, kubeClient client.Client, name, namespace string) (string, error) {
	subscription, err := GetSubscription(ctx, kubeClient, name, namespace)
	if err != nil {
		return "", err
	}

	if subscription.Status.InstalledCSV == "" {
		return "", nil
	}

	return subscription.Status.InstalledCSV, nil
}

func GetSubscriptionStatus(ctx context.Context, kubeClient client.Client, name, namespace string) (*operatorsv1alpha1.SubscriptionStatus, error) {
	subscription, err := GetSubscription(ctx, kubeClient, name, namespace)
	if err != nil {
		return nil, err
	}

	return &subscription.Status, nil
}

func GetSubscriptionConditions(ctx context.Context, kubeClient client.Client, name, namespace string) ([]operatorsv1alpha1.SubscriptionCondition, error) {
	subscription, err := GetSubscription(ctx, kubeClient, name, namespace)
	if err != nil {
		return nil, err
	}

	return subscription.Status.Conditions, nil
}
