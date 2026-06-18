package resources

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/observability-ui/development-tools/internal/constants"
)

type TempoStackConfig struct {
	Name             string
	Namespace        string
	StorageSize      string
	SecretName       string
	SourceNamespace  string
}

// CreateTempoStack creates a TempoStack CR configured with S3-compatible storage
func CreateTempoStack(ctx context.Context, kubeClient client.Client, config TempoStackConfig) error {
	// Copy secret from source namespace (e.g., openshift-tracing) if needed
	if config.SourceNamespace != "" && config.SourceNamespace != config.Namespace {
		if err := CopySecretToNamespace(ctx, kubeClient, config.SecretName, config.SourceNamespace, config.Namespace); err != nil {
			return err
		}
	}

	// TempoStack CR
	tempoStack := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "tempo.grafana.com/v1alpha1",
			"kind":       "TempoStack",
			"metadata": map[string]interface{}{
				"name":      config.Name,
				"namespace": config.Namespace,
			},
			"spec": map[string]interface{}{
				"storage": map[string]interface{}{
					"secret": map[string]interface{}{
						"name": config.SecretName,
						"type": "s3",
					},
				},
				"storageSize": config.StorageSize,
				"tenants": map[string]interface{}{
					"mode": "openshift",
					"authentication": []interface{}{
						map[string]interface{}{
							"tenantName": constants.PlatformTenantName,
							"tenantId":   constants.PlatformTenantID,
						},
						map[string]interface{}{
							"tenantName": constants.UserTenantName,
							"tenantId":   constants.UserTenantID,
						},
					},
				},
				"observability": map[string]interface{}{
					"tracing": map[string]interface{}{
						"otlp_http_endpoint": fmt.Sprintf("http://%s.%s:4318", constants.PlatformCollectorName, constants.TracingNamespace),
						"sampling_fraction":  "1",
					},
				},
				"template": map[string]interface{}{
					"gateway": map[string]interface{}{
						"enabled": true,
					},
					"queryFrontend": map[string]interface{}{
						"jaegerQuery": map[string]interface{}{
							"enabled": true,
							"monitorTab": map[string]interface{}{
								"enabled":            true,
								"prometheusEndpoint": "https://thanos-querier.openshift-monitoring.svc.cluster.local:9092",
							},
						},
					},
				},
			},
		},
	}

	err := kubeClient.Create(ctx, tempoStack)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create TempoStack: %w", err)
	}
	return nil
}

// GetTempoStack retrieves a TempoStack CR
func GetTempoStack(ctx context.Context, kubeClient client.Client, name, namespace string) (*unstructured.Unstructured, error) {
	tempoStack := &unstructured.Unstructured{}
	tempoStack.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "tempo.grafana.com",
		Version: "v1alpha1",
		Kind:    "TempoStack",
	})

	key := client.ObjectKey{Name: name, Namespace: namespace}
	err := kubeClient.Get(ctx, key, tempoStack)
	if err != nil {
		return nil, err
	}

	return tempoStack, nil
}
