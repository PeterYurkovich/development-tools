package k8s

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ScaleDeployment(ctx context.Context, kubeClient client.Client, name, namespace string, replicas int32) error {
	deployment := &appsv1.Deployment{}
	key := client.ObjectKey{Namespace: namespace, Name: name}

	if err := kubeClient.Get(ctx, key, deployment); err != nil {
		return err
	}

	deployment.Spec.Replicas = &replicas

	return kubeClient.Update(ctx, deployment)
}

func UpdateDeploymentImage(ctx context.Context, kubeClient client.Client, name, namespace, image string) error {
	deployment := &appsv1.Deployment{}
	key := client.ObjectKey{Namespace: namespace, Name: name}

	if err := kubeClient.Get(ctx, key, deployment); err != nil {
		return fmt.Errorf("failed to get deployment: %w", err)
	}

	if len(deployment.Spec.Template.Spec.Containers) == 0 {
		return fmt.Errorf("no containers found in deployment %s/%s", namespace, name)
	}

	deployment.Spec.Template.Spec.Containers[0].Image = image

	if err := kubeClient.Update(ctx, deployment); err != nil {
		return fmt.Errorf("failed to update deployment: %w", err)
	}

	return nil
}
