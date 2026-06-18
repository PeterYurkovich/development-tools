package k8s

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func EnsureNamespaceWithLabels(ctx context.Context, kubeClient client.Client,
	name string, labels map[string]string) (bool, error) {
	namespace := &corev1.Namespace{}
	key := client.ObjectKey{Name: name}

	err := kubeClient.Get(ctx, key, namespace)
	if err == nil {
		return false, nil
	}

	if !errors.IsNotFound(err) {
		return false, err
	}

	namespace = &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
	}

	if err := kubeClient.Create(ctx, namespace); err != nil {
		return false, err
	}

	return true, nil
}

func DeleteNamespace(ctx context.Context, kubeClient client.Client, name string) error {
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	err := kubeClient.Delete(ctx, namespace)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	return nil
}
