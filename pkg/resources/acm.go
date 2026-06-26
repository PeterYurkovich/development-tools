package resources

import (
	"context"
	"fmt"
	"time"

	mcov1beta2 "github.com/stolostron/multicluster-observability-operator/operators/multiclusterobservability/api/v1beta2"
	mcoshared "github.com/stolostron/multicluster-observability-operator/operators/multiclusterobservability/api/shared"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/observability-ui/development-tools/internal/constants"
)

var multiClusterHubGVK = schema.GroupVersionKind{
	Group:   "operator.open-cluster-management.io",
	Version: "v1",
	Kind:    "MultiClusterHub",
}

func CreateMultiClusterHub(ctx context.Context, kubeClient client.Client) error {
	mch := &unstructured.Unstructured{}
	mch.SetGroupVersionKind(multiClusterHubGVK)
	mch.SetName(constants.ACMMultiClusterHubName)
	mch.SetNamespace(constants.ACMNamespace)

	err := kubeClient.Create(ctx, mch)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create MultiClusterHub: %w", err)
	}
	return nil
}

func WaitForMultiClusterHub(ctx context.Context, kubeClient client.Client, logFn func(string)) error {
	mch := &unstructured.Unstructured{}
	mch.SetGroupVersionKind(multiClusterHubGVK)

	key := client.ObjectKey{
		Name:      constants.ACMMultiClusterHubName,
		Namespace: constants.ACMNamespace,
	}

	pollInterval := 15 * time.Second
	timeout := 15 * time.Minute
	attempts := 0

	return wait.PollUntilContextTimeout(ctx, pollInterval, timeout, true, func(ctx context.Context) (bool, error) {
		attempts++

		if err := kubeClient.Get(ctx, key, mch); err != nil {
			if errors.IsNotFound(err) {
				return false, nil
			}
			return false, fmt.Errorf("failed to get MultiClusterHub: %w", err)
		}

		conditions, found, err := unstructured.NestedSlice(mch.Object, "status", "conditions")
		if err != nil || !found {
			if attempts%4 == 0 {
				logFn("Waiting for MultiClusterHub status conditions...")
			}
			return false, nil
		}

		for _, condRaw := range conditions {
			cond, ok := condRaw.(map[string]interface{})
			if !ok {
				continue
			}
			condType, _, _ := unstructured.NestedString(cond, "type")
			condStatus, _, _ := unstructured.NestedString(cond, "status")
			if condType == "Complete" && condStatus == "True" {
				logFn("MultiClusterHub is ready")
				return true, nil
			}
		}

		if attempts%4 == 0 {
			logFn(fmt.Sprintf("Still waiting for MultiClusterHub to be ready (%dm elapsed)...", int(time.Duration(attempts)*pollInterval/time.Minute)))
		}

		return false, nil
	})
}

func CreateMultiClusterObservability(ctx context.Context, kubeClient client.Client) error {
	mco := &mcov1beta2.MultiClusterObservability{
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.ACMMCOName,
		},
		Spec: mcov1beta2.MultiClusterObservabilitySpec{
			ObservabilityAddonSpec: &mcoshared.ObservabilityAddonSpec{},
			StorageConfig: &mcov1beta2.StorageConfig{
				MetricObjectStorage: &mcoshared.PreConfiguredStorage{
					Name: constants.ACMMinIOSecretName,
					Key:  constants.ACMMinIOSecretKey,
				},
			},
		},
	}

	err := kubeClient.Create(ctx, mco)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create MultiClusterObservability: %w", err)
	}
	return nil
}
