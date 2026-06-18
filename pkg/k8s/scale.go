package k8s

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ScaleDeployment scales a deployment to the specified number of replicas
func ScaleDeployment(ctx context.Context, kubeClient client.Client, name, namespace string, replicas int32) error {
	deployment := &appsv1.Deployment{}
	key := client.ObjectKey{Name: name, Namespace: namespace}

	if err := kubeClient.Get(ctx, key, deployment); err != nil {
		return fmt.Errorf("failed to get deployment %s/%s: %w", namespace, name, err)
	}

	deployment.Spec.Replicas = &replicas

	if err := kubeClient.Update(ctx, deployment); err != nil {
		return fmt.Errorf("failed to scale deployment %s/%s to %d replicas: %w", namespace, name, replicas, err)
	}

	return nil
}

// UpdateDeploymentImage updates the image of the first container in a deployment
func UpdateDeploymentImage(ctx context.Context, kubeClient client.Client, name, namespace, image string) error {
	deployment := &appsv1.Deployment{}
	key := client.ObjectKey{Name: name, Namespace: namespace}

	if err := kubeClient.Get(ctx, key, deployment); err != nil {
		return fmt.Errorf("failed to get deployment %s/%s: %w", namespace, name, err)
	}

	if len(deployment.Spec.Template.Spec.Containers) == 0 {
		return fmt.Errorf("deployment %s/%s has no containers", namespace, name)
	}

	deployment.Spec.Template.Spec.Containers[0].Image = image

	if err := kubeClient.Update(ctx, deployment); err != nil {
		return fmt.Errorf("failed to update deployment %s/%s image: %w", namespace, name, err)
	}

	return nil
}
