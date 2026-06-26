package resources

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/observability-ui/development-tools/internal/constants"
)

var flowCollectorGVK = schema.GroupVersionKind{
	Group:   "flows.netobserv.io",
	Version: "v1beta2",
	Kind:    "FlowCollector",
}

func CreateNetObservLokiStackSecret(ctx context.Context, kubeClient client.Client) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "minio",
			Namespace: constants.NetObservNamespace,
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			"access_key_id":     "minio",
			"access_key_secret": "minio123",
			"bucketnames":       "loki",
			"endpoint":          "http://minio.minio.svc:9000",
		},
	}

	err := kubeClient.Create(ctx, secret)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create netobserv MinIO secret: %w", err)
	}
	return nil
}

func CreateFlowCollector(ctx context.Context, kubeClient client.Client) error {
	fc := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "flows.netobserv.io/v1beta2",
			"kind":       "FlowCollector",
			"metadata": map[string]interface{}{
				"name": constants.FlowCollectorName,
			},
			"spec": map[string]interface{}{
				"namespace":       constants.NetObservNamespace,
				"deploymentModel": "Direct",
				"agent": map[string]interface{}{
					"type": "eBPF",
				},
				"loki": map[string]interface{}{
					"enable": true,
					"mode":   "LokiStack",
					"lokiStack": map[string]interface{}{
						"name": constants.NetObservLokiStackName,
					},
				},
				"consolePlugin": map[string]interface{}{
					"enable": true,
				},
			},
		},
	}
	fc.SetGroupVersionKind(flowCollectorGVK)

	err := kubeClient.Create(ctx, fc)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create FlowCollector: %w", err)
	}
	return nil
}

func WaitForNetObservPlugin(ctx context.Context, kubeClient client.Client, logFn func(string)) error {
	deployment := &appsv1.Deployment{}
	key := types.NamespacedName{
		Name:      constants.NetObservPluginDeployment,
		Namespace: constants.NetObservNamespace,
	}

	pollInterval := 5 * time.Second
	timeout := 10 * time.Minute
	attempts := 0

	return wait.PollUntilContextTimeout(ctx, pollInterval, timeout, true, func(ctx context.Context) (bool, error) {
		attempts++

		if err := kubeClient.Get(ctx, key, deployment); err != nil {
			if errors.IsNotFound(err) {
				if attempts%6 == 0 {
					logFn("Waiting for netobserv-plugin deployment to appear...")
				}
				return false, nil
			}
			return false, fmt.Errorf("failed to get netobserv-plugin deployment: %w", err)
		}

		if deployment.Status.ReadyReplicas >= 1 {
			logFn("netobserv-plugin is ready")
			return true, nil
		}

		if attempts%6 == 0 {
			logFn(fmt.Sprintf("Waiting for netobserv-plugin to be ready (ready: %d/1)...", deployment.Status.ReadyReplicas))
		}
		return false, nil
	})
}
