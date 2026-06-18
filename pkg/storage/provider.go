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
}

func NewProvider(config ProviderConfig) (StorageProvider, error) {
	switch config.Type {
	case StorageTypeMinio:
		return NewMinioProvider(config), nil
	default:
		return nil, fmt.Errorf("unknown storage type: %s", config.Type)
	}
}
