package resources

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func DeleteUIPlugin(ctx context.Context, kubeClient client.Client, name string) error {
	plugin := &unstructured.Unstructured{}
	plugin.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "observability.openshift.io",
		Version: "v1alpha1",
		Kind:    "UIPlugin",
	})
	plugin.SetName(name)

	err := kubeClient.Delete(ctx, plugin)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete UIPlugin %s: %w", name, err)
	}

	return nil
}

func DeleteClusterLogForwarder(ctx context.Context, kubeClient client.Client, name, namespace string) error {
	clf := &unstructured.Unstructured{}
	clf.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "observability.openshift.io",
		Version: "v1",
		Kind:    "ClusterLogForwarder",
	})
	clf.SetName(name)
	clf.SetNamespace(namespace)

	err := kubeClient.Delete(ctx, clf)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete ClusterLogForwarder %s/%s: %w", namespace, name, err)
	}

	return nil
}

func DeleteLokiStack(ctx context.Context, kubeClient client.Client, name, namespace string) error {
	lokistack := &unstructured.Unstructured{}
	lokistack.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "loki.grafana.com",
		Version: "v1",
		Kind:    "LokiStack",
	})
	lokistack.SetName(name)
	lokistack.SetNamespace(namespace)

	err := kubeClient.Delete(ctx, lokistack)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete LokiStack %s/%s: %w", namespace, name, err)
	}

	return nil
}

func DeleteServiceAccount(ctx context.Context, kubeClient client.Client, name, namespace string) error {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	err := kubeClient.Delete(ctx, sa)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete ServiceAccount %s/%s: %w", namespace, name, err)
	}

	return nil
}
