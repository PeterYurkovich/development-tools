package k8s

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DeploymentPollInterval   = 5 * time.Second
	DefaultDeploymentTimeout = 10 * time.Minute
)

// WaitForDeploymentReady waits for a deployment to become available
func WaitForDeploymentReady(ctx context.Context, kubeClient client.Client, name, namespace string, timeout time.Duration) error {
	if timeout == 0 {
		timeout = DefaultDeploymentTimeout
	}

	timeoutChan := time.After(timeout)
	ticker := time.NewTicker(DeploymentPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled while waiting for deployment %s/%s", namespace, name)

		case <-timeoutChan:
			return fmt.Errorf("timeout waiting for deployment %s/%s to be ready after %v", namespace, name, timeout)

		case <-ticker.C:
			deployment := &appsv1.Deployment{}
			key := client.ObjectKey{Name: name, Namespace: namespace}

			err := kubeClient.Get(ctx, key, deployment)
			if err != nil {
				if errors.IsNotFound(err) {
					// Deployment doesn't exist yet, continue waiting
					continue
				}
				return fmt.Errorf("error getting deployment %s/%s: %w", namespace, name, err)
			}

			// Check if deployment is available
			for _, condition := range deployment.Status.Conditions {
				if condition.Type == appsv1.DeploymentAvailable && condition.Status == "True" {
					return nil
				}
			}
		}
	}
}
