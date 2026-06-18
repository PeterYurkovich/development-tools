package resources

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type LoggingUIPluginConfig struct {
	Name          string
	LokiStackName string
	LogsLimit     int
	Timeout       string
	Schema        string
}

// CreateLoggingUIPlugin creates a Logging UIPlugin CR
func CreateLoggingUIPlugin(ctx context.Context, kubeClient client.Client, config LoggingUIPluginConfig) error {
	plugin := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "observability.openshift.io/v1alpha1",
			"kind":       "UIPlugin",
			"metadata": map[string]interface{}{
				"name": config.Name,
			},
			"spec": map[string]interface{}{
				"type": "Logging",
				"logging": map[string]interface{}{
					"lokiStack": map[string]interface{}{
						"name": config.LokiStackName,
					},
					"logsLimit":             config.LogsLimit,
					"timeout":               config.Timeout,
					"schema":                config.Schema,
					"showTimezoneSelector":  true,
				},
			},
		},
	}

	err := kubeClient.Create(ctx, plugin)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create UIPlugin: %w", err)
	}
	return nil
}

// GetUIPlugin retrieves a UIPlugin CR
func GetUIPlugin(ctx context.Context, kubeClient client.Client, name string) (*unstructured.Unstructured, error) {
	plugin := &unstructured.Unstructured{}
	plugin.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "observability.openshift.io",
		Version: "v1alpha1",
		Kind:    "UIPlugin",
	})

	key := client.ObjectKey{Name: name}
	err := kubeClient.Get(ctx, key, plugin)
	if err != nil {
		return nil, err
	}

	return plugin, nil
}
