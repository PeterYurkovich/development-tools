package resources

import (
	"context"
	"fmt"

	lokiv1 "github.com/grafana/loki/operator/api/loki/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type LokiStackConfig struct {
	Name             string
	Namespace        string
	Size             string
	StorageClassName string
	SecretName       string
	SourceNamespace  string
}

func CopySecretToNamespace(ctx context.Context, kubeClient client.Client, secretName, sourceNS, targetNS string) error {
	sourceSecret := &corev1.Secret{}
	key := client.ObjectKey{Name: secretName, Namespace: sourceNS}

	if err := kubeClient.Get(ctx, key, sourceSecret); err != nil {
		return fmt.Errorf("failed to get secret %s/%s: %w", sourceNS, secretName, err)
	}

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

func CreateLokiStack(ctx context.Context, kubeClient client.Client, config LokiStackConfig) error {
	if config.SourceNamespace != "" && config.SourceNamespace != config.Namespace {
		if err := CopySecretToNamespace(ctx, kubeClient, config.SecretName, config.SourceNamespace, config.Namespace); err != nil {
			return err
		}
	}

	lokiStack := &lokiv1.LokiStack{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.Name,
			Namespace: config.Namespace,
		},
		Spec: lokiv1.LokiStackSpec{
			Size: lokiv1.LokiStackSizeType(config.Size),
			Storage: lokiv1.ObjectStorageSpec{
				Schemas: []lokiv1.ObjectStorageSchema{{
					Version:       lokiv1.ObjectStorageSchemaV12,
					EffectiveDate: "2022-06-01",
				}},
				Secret: lokiv1.ObjectStorageSecretSpec{
					Name: config.SecretName,
					Type: lokiv1.ObjectStorageSecretS3,
				},
			},
			StorageClassName: config.StorageClassName,
			Tenants: &lokiv1.TenantsSpec{
				Mode: lokiv1.OpenshiftLogging,
			},
		},
	}

	err := kubeClient.Create(ctx, lokiStack)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create LokiStack: %w", err)
	}
	return nil
}

func GetLokiStack(ctx context.Context, kubeClient client.Client, name, namespace string) (*lokiv1.LokiStack, error) {
	lokiStack := &lokiv1.LokiStack{}
	key := client.ObjectKey{Name: name, Namespace: namespace}

	err := kubeClient.Get(ctx, key, lokiStack)
	if err != nil {
		return nil, err
	}

	return lokiStack, nil
}
