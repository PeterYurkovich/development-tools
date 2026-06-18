# Storage Provider Package

Abstraction layer for object storage backends used by observability components (Loki, Tempo, ACM).

## Overview

This package provides a unified interface for deploying and managing object storage backends. Currently implements MinIO for development/testing. The interface design allows for future addition of production storage providers.

## Quick Start

### Deploy MinIO for Logging

```go
import (
    "github.com/observability-ui/development-tools/pkg/storage"
    "github.com/observability-ui/development-tools/pkg/executor"
)

config := storage.ProviderConfig{
    Type:                   storage.StorageTypeMinio,
    Namespace:              "minio",
    BucketName:             "loki",
    StorageSize:            "10Gi",
    StorageClass:           "",
    UseDefaultStorageClass: true,
}

provider, err := storage.NewProvider(config)
if err != nil {
    return err
}

exec := executor.NewExecutor()
secretName, err := provider.Deploy(ctx, k8sClient, exec)
if err != nil {
    return err
}

// Use secretName when creating LokiStack CR
```

### Cleanup Storage

```go
exec := executor.NewExecutor()
err := provider.Delete(ctx, k8sClient, exec)
```

## Component Configurations

### Logging (Loki)
- Namespace: `minio`
- Bucket: `loki`
- Storage: `10Gi`

### Tracing (Tempo)
- Namespace: `openshift-tracing`
- Bucket: `tempo`
- Storage: `10Gi`

### ACM (Thanos)
- Namespace: `open-cluster-management-observability`
- Bucket: `thanos`
- Storage: `1Gi`
- Storage Class: `gp3-csi`

## Interface

```go
type StorageProvider interface {
    GetType() StorageType
    Deploy(ctx context.Context, client client.Client, exec *executor.Executor) (string, error)
    Delete(ctx context.Context, client client.Client, exec *executor.Executor) error
    GetSecretName() string
    GetEndpoint() string
    GetBucketName() string
}
```

## MinIO Implementation

### What Deploy() Does

1. Creates namespace if it doesn't exist
2. Creates PersistentVolumeClaim with configured size
3. Creates Deployment (single replica, Recreate strategy)
4. Creates Service (ClusterIP on port 9000)
5. Creates Secret with S3-compatible credentials
6. Waits for deployment to become ready (timeout: 5 minutes)

Each step sends progress updates via the executor pattern for real-time feedback.

### Secret Format

Generic S3-compatible format that works with Loki, Tempo, and ACM:

```yaml
apiVersion: v1
kind: Secret
type: Opaque
data:
  access_key_id: bWluaW8=           # "minio"
  access_key_secret: bWluaW8xMjM=   # "minio123"
  bucketname: <base64-encoded>
  endpoint: <base64-encoded>        # http://minio.<namespace>.svc:9000
```

## Adding New Storage Providers

The interface design allows adding new storage backends in the future. To add a provider:
1. Define new `StorageType` constant
2. Implement `StorageProvider` interface
3. Add case to `NewProvider()` factory function

## Design Philosophy

- **Development-focused**: MinIO for easy local/dev deployments
- **Production-ready interface**: Easy to swap MinIO for real S3/Azure
- **Single source of truth**: Replaces scattered YAML files
- **Type-safe**: No YAML templates, all Go structs
- **Wait for ready**: Ensures resources are operational before returning
