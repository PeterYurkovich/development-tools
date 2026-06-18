package operators

import (
	"context"
	"fmt"

	operatorsv1 "github.com/operator-framework/api/pkg/operators/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type OperatorGroupConfig struct {
	Name             string
	Namespace        string
	TargetNamespaces []string
}

func CreateOperatorGroup(ctx context.Context, kubeClient client.Client, config OperatorGroupConfig) error {
	operatorGroup := &operatorsv1.OperatorGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.Name,
			Namespace: config.Namespace,
			Labels: map[string]string{
				"obstool.observability.openshift.io/managed": "true",
			},
		},
		Spec: operatorsv1.OperatorGroupSpec{
			TargetNamespaces: config.TargetNamespaces,
		},
	}

	err := kubeClient.Create(ctx, operatorGroup)
	if err != nil {
		return fmt.Errorf("failed to create operatorgroup %s/%s: %w", config.Namespace, config.Name, err)
	}

	return nil
}

func GetOperatorGroup(ctx context.Context, kubeClient client.Client, name, namespace string) (*operatorsv1.OperatorGroup, error) {
	operatorGroup := &operatorsv1.OperatorGroup{}
	key := client.ObjectKey{Name: name, Namespace: namespace}

	err := kubeClient.Get(ctx, key, operatorGroup)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, fmt.Errorf("operatorgroup %s/%s not found", namespace, name)
		}
		return nil, fmt.Errorf("failed to get operatorgroup %s/%s: %w", namespace, name, err)
	}

	return operatorGroup, nil
}

func DeleteOperatorGroup(ctx context.Context, kubeClient client.Client, name, namespace string) error {
	operatorGroup := &operatorsv1.OperatorGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	err := kubeClient.Delete(ctx, operatorGroup)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete operatorgroup %s/%s: %w", namespace, name, err)
	}

	return nil
}

func ListOperatorGroups(ctx context.Context, kubeClient client.Client, namespace string) (*operatorsv1.OperatorGroupList, error) {
	operatorGroupList := &operatorsv1.OperatorGroupList{}

	listOptions := &client.ListOptions{
		Namespace: namespace,
	}

	err := kubeClient.List(ctx, operatorGroupList, listOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list operatorgroups in namespace %s: %w", namespace, err)
	}

	return operatorGroupList, nil
}

func OperatorGroupExists(ctx context.Context, kubeClient client.Client, namespace string) (bool, error) {
	list, err := ListOperatorGroups(ctx, kubeClient, namespace)
	if err != nil {
		return false, err
	}

	return len(list.Items) > 0, nil
}

func EnsureOperatorGroup(ctx context.Context, kubeClient client.Client, config OperatorGroupConfig) (bool, error) {
	exists, err := OperatorGroupExists(ctx, kubeClient, config.Namespace)
	if err != nil {
		return false, err
	}

	if exists {
		return false, nil
	}

	err = CreateOperatorGroup(ctx, kubeClient, config)
	if err != nil {
		return false, err
	}

	return true, nil
}
