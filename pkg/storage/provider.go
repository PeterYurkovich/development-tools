package storage

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/observability-ui/development-tools/pkg/executor"
)

type StorageType string

const (
	StorageTypeMinio StorageType = "minio"
)

// SecretFormat controls the format of the object storage credentials secret.
type SecretFormat string

const (
	// SecretFormatGeneric produces a flat key-value secret (access_key_id, access_key_secret, bucketname, endpoint).
	SecretFormatGeneric SecretFormat = "generic"
	// SecretFormatThanos produces a single-key secret with a thanos.yaml blob for the MCO storageConfig.
	SecretFormatThanos SecretFormat = "thanos"
)

type StorageProvider interface {
	GetType() StorageType
	Deploy(ctx context.Context, client client.Client, exec *executor.Executor) (string, error)
	Delete(ctx context.Context, client client.Client, exec *executor.Executor) error
	GetSecretName() string
	GetEndpoint() string
	GetBucketName() string
}

type ProviderConfig struct {
	Type                   StorageType
	Namespace              string
	BucketName             string
	StorageSize            string
	StorageClass           string
	UseDefaultStorageClass bool
	SecretFormat           SecretFormat
}

func NewProvider(config ProviderConfig) (StorageProvider, error) {
	switch config.Type {
	case StorageTypeMinio:
		return NewMinioProvider(config), nil
	default:
		return nil, fmt.Errorf("unknown storage type: %s", config.Type)
	}
}
