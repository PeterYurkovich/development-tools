package k8s

import (
	"context"
	"fmt"

	storagev1 "k8s.io/api/storage/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DefaultStorageClassAnnotation = "storageclass.kubernetes.io/is-default-class"
	DefaultStorageClassName       = "gp3-csi" // AWS default
)

// GetDefaultStorageClass returns the cluster's default StorageClass name
func GetDefaultStorageClass(ctx context.Context, kubeClient client.Client) (string, error) {
	storageClassList := &storagev1.StorageClassList{}
	if err := kubeClient.List(ctx, storageClassList); err != nil {
		return "", fmt.Errorf("failed to list storage classes: %w", err)
	}

	// Look for annotated default
	for _, sc := range storageClassList.Items {
		if sc.Annotations[DefaultStorageClassAnnotation] == "true" {
			return sc.Name, nil
		}
	}

	// If no default found, return first available
	if len(storageClassList.Items) > 0 {
		return storageClassList.Items[0].Name, nil
	}

	// Fallback to common default
	return DefaultStorageClassName, nil
}
