package resources

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ClusterLogForwarderConfig struct {
	Name             string
	Namespace        string
	ServiceAccount   string
	LokiStackName    string
	LokiStackNS      string
}

// CreateClusterLogForwarder creates a ClusterLogForwarder CR
func CreateClusterLogForwarder(ctx context.Context, kubeClient client.Client, config ClusterLogForwarderConfig) error {
	clf := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "observability.openshift.io/v1",
			"kind":       "ClusterLogForwarder",
			"metadata": map[string]interface{}{
				"name":      config.Name,
				"namespace": config.Namespace,
			},
			"spec": map[string]interface{}{
				"serviceAccount": map[string]interface{}{
					"name": config.ServiceAccount,
				},
				"outputs": []interface{}{
					map[string]interface{}{
						"name": "default-lokistack",
						"type": "lokiStack",
						"lokiStack": map[string]interface{}{
							"target": map[string]interface{}{
								"name":      config.LokiStackName,
								"namespace": config.LokiStackNS,
							},
							"authentication": map[string]interface{}{
								"token": map[string]interface{}{
									"from": "serviceAccount",
								},
							},
						},
						"tls": map[string]interface{}{
							"ca": map[string]interface{}{
								"key":           "service-ca.crt",
								"configMapName": "openshift-service-ca.crt",
							},
						},
					},
				},
				"pipelines": []interface{}{
					map[string]interface{}{
						"name": "default-logstore",
						"inputRefs": []interface{}{
							"application",
							"infrastructure",
						},
						"outputRefs": []interface{}{
							"default-lokistack",
						},
					},
				},
			},
		},
	}

	err := kubeClient.Create(ctx, clf)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create ClusterLogForwarder: %w", err)
	}
	return nil
}

// GetClusterLogForwarder retrieves a ClusterLogForwarder CR
func GetClusterLogForwarder(ctx context.Context, kubeClient client.Client, name, namespace string) (*unstructured.Unstructured, error) {
	clf := &unstructured.Unstructured{}
	clf.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "observability.openshift.io",
		Version: "v1",
		Kind:    "ClusterLogForwarder",
	})

	key := client.ObjectKey{Name: name, Namespace: namespace}
	err := kubeClient.Get(ctx, key, clf)
	if err != nil {
		return nil, err
	}

	return clf, nil
}
