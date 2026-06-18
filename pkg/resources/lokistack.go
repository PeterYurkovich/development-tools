package resources

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var lokiStackGVR = schema.GroupVersionResource{
	Group:    "loki.grafana.com",
	Version:  "v1",
	Resource: "lokistacks",
}

type LokiStackConfig struct {
	Name             string
	Namespace        string
	Size             string
	StorageClassName string
	SecretName       string // Secret name created by storage provider
	SourceNamespace  string // Namespace where the original secret was created (e.g., "minio")
}

// CopySecretToNamespace copies a secret from one namespace to another
func CopySecretToNamespace(ctx context.Context, kubeClient client.Client, secretName, sourceNS, targetNS string) error {
	// Get the source secret
	sourceSecret := &corev1.Secret{}
	key := client.ObjectKey{Name: secretName, Namespace: sourceNS}
	
	if err := kubeClient.Get(ctx, key, sourceSecret); err != nil {
		return fmt.Errorf("failed to get secret %s/%s: %w", sourceNS, secretName, err)
	}

	// Create a copy in the target namespace
	targetSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: targetNS,
		},
		Type: sourceSecret.Type,
		Data: sourceSecret.Data,
	}

	err := kubeClient.Create(ctx, targetSecret)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create secret in %s: %w", targetNS, err)
	}

	return nil
}

// CreateLokiStack creates a LokiStack CR configured with S3-compatible storage backend
func CreateLokiStack(ctx context.Context, kubeClient client.Client, config LokiStackConfig) error {
	// Copy secret from source namespace (e.g., minio) to LokiStack namespace
	if config.SourceNamespace != "" && config.SourceNamespace != config.Namespace {
		if err := CopySecretToNamespace(ctx, kubeClient, config.SecretName, config.SourceNamespace, config.Namespace); err != nil {
			return err
		}
	}

	// LokiStack CR
	lokiStack := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "loki.grafana.com/v1",
			"kind":       "LokiStack",
			"metadata": map[string]interface{}{
				"name":      config.Name,
				"namespace": config.Namespace,
			},
			"spec": map[string]interface{}{
				"size": config.Size,
				"storage": map[string]interface{}{
					"schemas": []interface{}{
						map[string]interface{}{
							"version":       "v12",
							"effectiveDate": "2022-06-01",
						},
					},
					"secret": map[string]interface{}{
						"name": config.SecretName,
						"type": "s3",
					},
				},
				"storageClassName": config.StorageClassName,
				"tenants": map[string]interface{}{
					"mode": "openshift-logging",
				},
			},
		},
	}

	err := kubeClient.Create(ctx, lokiStack)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create LokiStack: %w", err)
	}
	return nil
}

// GetLokiStack retrieves a LokiStack CR
func GetLokiStack(ctx context.Context, kubeClient client.Client, name, namespace string) (*unstructured.Unstructured, error) {
	lokiStack := &unstructured.Unstructured{}
	lokiStack.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "loki.grafana.com",
		Version: "v1",
		Kind:    "LokiStack",
	})

	key := client.ObjectKey{Name: name, Namespace: namespace}
	err := kubeClient.Get(ctx, key, lokiStack)
	if err != nil {
		return nil, err
	}

	return lokiStack, nil
}
