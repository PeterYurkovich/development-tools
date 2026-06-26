package resources

import (
	"context"
	"fmt"

	lokiv1 "github.com/grafana/loki/operator/api/loki/v1"
	tempov1alpha1 "github.com/grafana/tempo-operator/api/tempo/v1alpha1"
	otelv1beta1 "github.com/open-telemetry/opentelemetry-operator/apis/v1beta1"
	uipluginv1alpha1 "github.com/rhobs/observability-operator/pkg/apis/uiplugin/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func DeleteUIPlugin(ctx context.Context, kubeClient client.Client, name string) error {
	plugin := &uipluginv1alpha1.UIPlugin{
		ObjectMeta: metav1.ObjectMeta{Name: name},
	}

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
	lokiStack := &lokiv1.LokiStack{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
	}

	err := kubeClient.Delete(ctx, lokiStack)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete LokiStack %s/%s: %w", namespace, name, err)
	}

	return nil
}

func DeleteTempoStack(ctx context.Context, kubeClient client.Client, name, namespace string) error {
	tempoStack := &tempov1alpha1.TempoStack{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
	}

	err := kubeClient.Delete(ctx, tempoStack)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete TempoStack %s/%s: %w", namespace, name, err)
	}

	return nil
}

func DeleteOTelCollector(ctx context.Context, kubeClient client.Client, name, namespace string) error {
	collector := &otelv1beta1.OpenTelemetryCollector{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
	}

	err := kubeClient.Delete(ctx, collector)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete OpenTelemetryCollector %s/%s: %w", namespace, name, err)
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
