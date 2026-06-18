# Storage Provider Interface - Implementation Plan

## Overview

Create a storage provider abstraction layer to handle object storage deployment for Loki, Tempo, and ACM observability stacks. This will standardize MinIO deployments and enable future migration to managed storage services.

---

## Design Decisions

Based on user feedback:

1. ✅ **Secret format**: Generic S3-compatible format for all components
2. ✅ **Wait behavior**: `Deploy()` waits for MinIO deployment to be ready
3. ✅ **Namespace handling**: Create namespace if it doesn't exist
4. ✅ **Image versioning**: Use same image for all components (`quay.io/minio/minio:latest`)
5. ✅ **Storage size**: Fixed per component (Logging: 10Gi, Tracing: 10Gi, ACM: 1Gi)
6. ✅ **No factory functions**: Deploy commands construct providers directly with config

---

## Architecture

### Storage Provider Interface

```go
type StorageProvider interface {
    GetType() StorageType
    Deploy(ctx context.Context, client client.Client) (string, error)
    Delete(ctx context.Context, client client.Client) error
    GetSecretName() string
    GetEndpoint() string
    GetBucketName() string
}
```

### MinIO Provider

Single implementation that:
- Creates namespace if missing
- Deploys PVC, Deployment, Service, Secret
- Waits for deployment to be ready (uses rollout status check)
- Returns secret name for use in LokiStack/TempoStack/ACM configs

### Usage Pattern

```go
// In deploy logging command
storageConfig := storage.ProviderConfig{
    Type:         storage.StorageTypeMinio,
    Namespace:    "minio",
    BucketName:   "loki",
    StorageSize:  "10Gi",
    StorageClass: storageClass,
}

provider := storage.NewProvider(storageConfig)
secretName, err := provider.Deploy(ctx, client)
// Use secretName when creating LokiStack
```

---

## Component-Specific Configurations

| Component | Namespace | Storage Size | Bucket | Credentials |
|-----------|-----------|--------------|--------|-------------|
| Logging | `minio` | 10Gi | `loki` | minio/minio123 |
| Tracing | `openshift-tracing` | 10Gi | `tempo` | minio/minio123 |
| ACM | `open-cluster-management-observability` | 1Gi | `thanos` | minio/minio123 |

All use: `quay.io/minio/minio:latest`

---

## Implementation Steps

### 1. Create Interface (`pkg/storage/provider.go`)

```go
package storage

import (
    "context"
    "sigs.k8s.io/controller-runtime/pkg/client"
)

type StorageType string

const (
    StorageTypeMinio StorageType = "minio"
    StorageTypeS3    StorageType = "s3"
    StorageTypeAzure StorageType = "azure"
    StorageTypeGCS   StorageType = "gcs"
)

type StorageProvider interface {
    GetType() StorageType
    Deploy(ctx context.Context, client client.Client) (string, error)
    Delete(ctx context.Context, client client.Client) error
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
        return nil, fmt.Errorf("storage type %s not implemented", config.Type)
    }
}
```

### 2. Implement MinIO Provider (`pkg/storage/minio.go`)

**Key responsibilities**:
- Create namespace if missing
- Deploy all resources (PVC, Deployment, Service, Secret)
- Wait for deployment rollout using controller-runtime
- Generic S3-compatible secret format
- Cleanup all resources on delete

**Resource creation**:
1. `createOrGetNamespace()` - Create namespace if doesn't exist
2. `createPVC()` - 10Gi/1Gi based on config
3. `createDeployment()` - Single replica, Recreate strategy
4. `createService()` - ClusterIP on port 9000
5. `createSecret()` - Generic S3 format
6. `waitForReady()` - Poll deployment status until ready

**Wait implementation**:
```go
func (m *MinioProvider) waitForReady(ctx context.Context, client client.Client) error {
    deployment := &appsv1.Deployment{}
    key := types.NamespacedName{Name: "minio", Namespace: m.config.Namespace}
    
    timeout := 5 * time.Minute
    interval := 2 * time.Second
    
    return wait.PollImmediate(interval, timeout, func() (bool, error) {
        if err := client.Get(ctx, key, deployment); err != nil {
            return false, err
        }
        
        return deployment.Status.ReadyReplicas == 1, nil
    })
}
```

### 3. Update Config Package

No changes needed - storage class already exists in `pkg/config/config.go`.

---

## Files to Create

```
pkg/storage/
├── provider.go          # Interface + NewProvider factory (1 file)
└── minio.go            # MinIO implementation (1 file)
```

Only 2 files total - keep it simple!

---

## Testing Plan

Manual testing steps:

1. **Deploy MinIO for logging**:
   ```go
   config := ProviderConfig{
       Type:         StorageTypeMinio,
       Namespace:    "minio",
       BucketName:   "loki",
       StorageSize:  "10Gi",
       StorageClass: "",
       UseDefaultStorageClass: true,
   }
   provider := NewProvider(config)
   secretName, err := provider.Deploy(ctx, client)
   ```
   Verify:
   - Namespace `minio` created
   - PVC, Deployment, Service, Secret all created
   - Deployment becomes ready
   - Secret has correct S3 format

2. **Delete MinIO**:
   ```go
   err := provider.Delete(ctx, client)
   ```
   Verify all resources removed

3. **Repeat for tracing** (namespace: `openshift-tracing`, bucket: `tempo`)

4. **Repeat for ACM** (namespace: `open-cluster-management-observability`, bucket: `thanos`, size: `1Gi`)

---

## Dependencies

**Blocks**:
- Deploy logging command
- Deploy tracing command
- Deploy ACM command

**Required packages** (already in go.mod):
- `k8s.io/api/apps/v1`
- `k8s.io/api/core/v1`
- `k8s.io/apimachinery/pkg/api/resource`
- `k8s.io/apimachinery/pkg/util/wait`
- `sigs.k8s.io/controller-runtime/pkg/client`

---

## Generic S3 Secret Format

All secrets will use this format:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: <bucket>-object-storage
  namespace: <namespace>
type: Opaque
data:
  access_key_id: <base64>
  access_key_secret: <base64>
  bucketname: <base64>
  endpoint: <base64>
```

This works for:
- LokiStack (expects `access_key_id`, `access_key_secret`, `bucketname`, `endpoint`)
- TempoStack (expects same format)
- ACM/Thanos (can be adapted if needed)

---

## Success Criteria

✅ Interface compiles without errors  
✅ MinIO provider creates namespace if missing  
✅ All 4 resources created (PVC, Deployment, Service, Secret)  
✅ Deploy() waits for MinIO to be ready  
✅ Secret uses generic S3-compatible format  
✅ Delete() removes all resources cleanly  
✅ Same image used for all components  
✅ Storage size is configurable per component  
✅ No 1-2 letter variable names  
✅ Minimal comments (only where needed)  

---

## Future Enhancements (Out of Scope)

- S3 provider (use existing AWS S3 buckets)
- Azure Blob provider
- GCS provider
- Credential injection from external secret managers

---

**Ready to implement**: Awaiting approval to proceed.
