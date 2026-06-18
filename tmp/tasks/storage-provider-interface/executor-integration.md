# Storage Provider - Executor Pattern Integration

## Summary

Updated storage provider implementation to use the executor pattern for progress updates, consistent with the rest of obstool's architecture.

---

## Changes Made

### 1. Interface Updated (`pkg/storage/provider.go`)

**Before:**
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

**After:**
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

### 2. MinIO Implementation Updated (`pkg/storage/minio.go`)

**Added Step Constants:**
```go
const (
    StepCreateNamespace = iota
    StepCreatePVC
    StepCreateDeployment
    StepCreateService
    StepCreateSecret
    StepWaitForReady
)

const (
    StepDeleteSecret = iota
    StepDeleteService
    StepDeleteDeployment
    StepDeletePVC
)
```

**Deploy Progress Updates:**
- ✅ Step 0: Create namespace (with log)
- ✅ Step 1: Create PVC (with size info)
- ✅ Step 2: Create Deployment (with bucket info)
- ✅ Step 3: Create Service (with port info)
- ✅ Step 4: Create Secret (with secret name)
- ✅ Step 5: Wait for ready (with periodic status updates every 10 seconds)

**Delete Progress Updates:**
- ✅ Step 0: Delete Secret
- ✅ Step 1: Delete Service
- ✅ Step 2: Delete Deployment
- ✅ Step 3: Delete PVC

**Error Handling:**
- All steps send error updates with context via `SendUpdateWithError()`
- Errors include which step failed and the error details

---

## Usage Examples

### Deploy with Progress Tracking (CLI Mode)

```go
import (
    "github.com/observability-ui/development-tools/pkg/storage"
    "github.com/observability-ui/development-tools/pkg/executor"
    "github.com/observability-ui/development-tools/pkg/output"
)

config := storage.ProviderConfig{
    Type:                   storage.StorageTypeMinio,
    Namespace:              "minio",
    BucketName:             "loki",
    StorageSize:            "10Gi",
    StorageClass:           "",
    UseDefaultStorageClass: true,
}

provider, _ := storage.NewProvider(config)
exec := executor.NewExecutor()
handler := output.NewCLIHandler()

go func() {
    for update := range exec.UpdateCh {
        handler.HandleUpdate(update)
    }
}()

secretName, err := provider.Deploy(ctx, client, exec)
```

**Output:**
```
⏳ Create namespace minio
   Ensuring namespace minio exists
✓ Create namespace minio

⏳ Create PersistentVolumeClaim
   Creating PVC with size 10Gi
✓ Create PersistentVolumeClaim

⏳ Create MinIO deployment
   Deploying MinIO for bucket: loki
✓ Create MinIO deployment

⏳ Create MinIO service
   Creating service on port 9000
✓ Create MinIO service

⏳ Create storage credentials secret
   Creating secret: loki-object-storage
✓ Create storage credentials secret

⏳ Wait for MinIO to be ready
   Waiting for deployment to become ready (timeout: 5m0s)
   Still waiting... (ready: 0/1)
   MinIO deployment is ready
✓ Wait for MinIO to be ready
```

### Deploy with TUI Mode

```go
model := tui.NewProgressModel("Deploy MinIO Storage", 
    func() error {
        provider, _ := storage.NewProvider(config)
        exec := executor.NewExecutor()
        
        go func() {
            for update := range exec.UpdateCh {
                if update.Message != "" {
                    continue // TUI ignores log messages
                }
                program.Send(tui.OperationUpdateMsg{
                    Index:  update.Index,
                    Status: update.Status,
                    Step:   update.Step,
                    Error:  update.Error,
                })
            }
        }()
        
        secretName, err := provider.Deploy(ctx, client, exec)
        return err
    })

program := tea.NewProgram(model)
program.Run()
```

---

## Progress Update Flow

### Deploy Operation

| Step | Index | Status | Message |
|------|-------|--------|---------|
| Create namespace | 0 | InProgress → Complete | "Create namespace minio" |
| Create PVC | 1 | InProgress → Complete | "Create PersistentVolumeClaim" |
| Create Deployment | 2 | InProgress → Complete | "Create MinIO deployment" |
| Create Service | 3 | InProgress → Complete | "Create MinIO service" |
| Create Secret | 4 | InProgress → Complete | "Create storage credentials secret" |
| Wait for ready | 5 | InProgress → Complete | "Wait for MinIO to be ready" |

### Delete Operation

| Step | Index | Status | Message |
|------|-------|--------|---------|
| Delete Secret | 0 | InProgress → Complete | "Delete storage credentials secret" |
| Delete Service | 1 | InProgress → Complete | "Delete MinIO service" |
| Delete Deployment | 2 | InProgress → Complete | "Delete MinIO deployment" |
| Delete PVC | 3 | InProgress → Complete | "Delete PersistentVolumeClaim" |

---

## Log Messages

### Deploy Logs
- **Step 0**: `"Ensuring namespace <namespace> exists"`
- **Step 1**: `"Creating PVC with size <size>"`
- **Step 2**: `"Deploying MinIO for bucket: <bucket>"`
- **Step 3**: `"Creating service on port 9000"`
- **Step 4**: `"Creating secret: <secret-name>"`
- **Step 5**: 
  - `"Waiting for deployment to become ready (timeout: 5m0s)"`
  - `"Still waiting... (ready: 0/1)"` (every 10 seconds)
  - `"MinIO deployment is ready"`

### Delete Logs
- **Step 0**: `"Deleting secret: <secret-name>"`
- **Step 1**: `"Deleting service"`
- **Step 2**: `"Deleting deployment"`
- **Step 3**: `"Deleting PVC and associated storage"`

---

## Benefits

✅ **Consistent UX**: Same progress pattern as monitoring commands  
✅ **Real-time feedback**: Users see what's happening at each step  
✅ **Error context**: Clear indication of which step failed  
✅ **TUI support**: Works seamlessly with Bubble Tea progress display  
✅ **CLI support**: Works with CLIHandler for terminal output  
✅ **Testable**: Business logic separate from display logic  

---

## File Statistics

**Updated Files:**
- `pkg/storage/provider.go`: 43 lines (was 41)
- `pkg/storage/minio.go`: 403 lines (was 311)
- **Total**: 446 lines (was 352)

**Growth**: +94 lines for comprehensive progress tracking

---

## Verification

✅ Code compiles without errors  
✅ Full obstool binary builds successfully  
✅ Interface matches executor pattern used in operations  
✅ All steps have clear names and log messages  
✅ Error handling integrated with executor  

---

## Next Steps

When implementing deploy commands (logging, tracing, ACM), the storage deployment will automatically provide:
- Step-by-step progress in CLI mode
- Real-time progress updates in TUI mode
- Detailed logging for debugging
- Clear error reporting

Example from deploy logging command:
```go
func deployLogging(ctx context.Context, client client.Client, config Config, exec *executor.Executor) error {
    defer exec.Close()
    
    // Storage deployment (6 steps with progress)
    provider := storage.NewProvider(storageConfig)
    secretName, err := provider.Deploy(ctx, client, exec)
    if err != nil {
        return err
    }
    
    // Continue with LokiStack deployment...
}
```
