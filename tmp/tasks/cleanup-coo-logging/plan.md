# Cleanup COO and Cleanup Logging - Implementation Plan

## Overview

Implement `obstool cleanup coo` and `obstool cleanup logging` commands to remove deployed operators and their resources.

## Task Breakdown

### Cleanup COO Command

**What it does**: Remove Cluster Observability Operator and related resources

**Resources to delete** (in order):
1. Delete Subscription (if exists)
2. Delete CSV (ClusterServiceVersion) (if exists)
3. Delete CatalogSource (if created by us - check for our catalog names)
4. Delete ImageDigestMirrorSet (if created - check for `idms-coo-quay` or `idms-coo-stage`)
5. Optionally: Delete OperatorGroup (if desired)
6. Optionally: Delete namespace (if desired - probably should NOT delete by default)

**Flags**:
- `--confirm` (bool): Skip confirmation prompt (default: false)
- `--delete-namespace` (bool): Also delete namespace (default: false)

**TUI vs CLI**:
- Both modes supported
- TUI shows progress for each deletion step
- CLI mode requires all flags OR prompts for confirmation

**Steps** (executor pattern):
1. Get current Subscription (if exists, extract CSV name)
2. Delete Subscription
3. Delete CSV (use name from subscription status)
4. Delete CatalogSource (if name matches our constants)
5. Delete IDMS (check both quay and stage names)
6. Delete OperatorGroup (optional)
7. Delete namespace (optional, if flag set)

### Cleanup Logging Command

**What it does**: Remove logging stack (operators and deployed resources)

**Resources to delete** (in order):
1. Delete UIPlugin (logging)
2. Delete ClusterLogForwarder
3. Delete LokiStack
4. Delete MinIO resources (Deployment, Service, PVC, Secret)
5. Delete collector RBAC (ServiceAccount, ClusterRoleBindings)
6. Optionally: Delete operators (Subscriptions, CSVs)
7. Optionally: Delete namespaces

**Flags**:
- `--confirm` (bool): Skip confirmation prompt (default: false)
- `--delete-operators` (bool): Also delete operators (default: false)
- `--delete-namespaces` (bool): Also delete namespaces (default: false)
- `--delete-minio` (bool): Delete MinIO resources (default: true)

**TUI vs CLI**:
- Both modes supported
- TUI shows progress for each deletion step

**Steps** (executor pattern):
1. Delete UIPlugin (logging)
2. Delete ClusterLogForwarder
3. Delete LokiStack
4. Delete MinIO resources (if flag set)
5. Delete collector RBAC
6. Delete logging operator (if flag set)
7. Delete loki operator (if flag set)
8. Delete namespaces (if flag set)

## Files to Create/Modify

### New Files

1. **cmd/cleanup/coo.go**
   - Cobra command definition
   - Flag definitions
   - CLI and TUI mode handling
   - Input collection (TUI)

2. **cmd/cleanup/logging.go**
   - Cobra command definition
   - Flag definitions
   - CLI and TUI mode handling
   - Input collection (TUI)

3. **pkg/operations/cleanup_coo.go**
   - CleanupCOOConfig struct
   - CleanupCOO() function with executor pattern
   - Step constants

4. **pkg/operations/cleanup_logging.go**
   - CleanupLoggingConfig struct
   - CleanupLogging() function with executor pattern
   - Step constants

5. **pkg/resources/delete.go** (helper functions)
   - DeleteUIPlugin()
   - DeleteClusterLogForwarder()
   - DeleteLokiStack()
   - DeleteCatalogSource() (might already exist)
   - DeleteIDMS() (might already exist)

### Modified Files

1. **pkg/operators/catalogsource.go**
   - Add DeleteCatalogSource() if not exists

2. **pkg/operators/idms.go**
   - Add DeleteIDMS() if not exists

3. **pkg/storage/minio.go**
   - Add Delete() method to MinIO provider

## Implementation Details

### Cleanup COO

```go
// Step constants
const (
    StepDeleteSubscription = iota
    StepDeleteCSV
    StepDeleteCatalogSource
    StepDeleteIDMS
    StepDeleteOperatorGroup
    StepDeleteNamespace
)

type CleanupCOOConfig struct {
    DeleteNamespace     bool
    DeleteOperatorGroup bool
    Confirm             bool
}

func CleanupCOO(ctx context.Context, kubeClient client.Client, 
                config CleanupCOOConfig, exec *executor.Executor) error {
    defer exec.Close()
    
    // 1. Get subscription to find CSV name
    // 2. Delete subscription
    // 3. Delete CSV
    // 4. Delete CatalogSource (check if ours)
    // 5. Delete IDMS (check both quay/stage)
    // 6. Optional: Delete OperatorGroup
    // 7. Optional: Delete namespace
    
    return nil
}
```

### Cleanup Logging

```go
// Step constants
const (
    StepDeleteUIPlugin = iota
    StepDeleteClusterLogForwarder
    StepDeleteLokiStack
    StepDeleteMinIO
    StepDeleteCollectorRBAC
    StepDeleteLoggingOperator
    StepDeleteLokiOperator
    StepDeleteNamespaces
)

type CleanupLoggingConfig struct {
    DeleteOperators  bool
    DeleteNamespaces bool
    DeleteMinIO      bool
    Confirm          bool
}

func CleanupLogging(ctx context.Context, kubeClient client.Client,
                    config CleanupLoggingConfig, exec *executor.Executor) error {
    defer exec.Close()
    
    // 1. Delete UIPlugin
    // 2. Delete ClusterLogForwarder
    // 3. Delete LokiStack
    // 4. Delete MinIO (if flag set)
    // 5. Delete collector RBAC
    // 6. Delete operators (if flag set)
    // 7. Delete namespaces (if flag set)
    
    return nil
}
```

## Error Handling

- **Resource not found**: Log but don't fail (already deleted is OK)
- **Deletion errors**: Report error but continue with other resources
- **Timeout errors**: Allow configurable timeout for waiting on deletion

## Testing Strategy

1. **Manual testing**: Deploy then cleanup
2. **Partial cleanup**: Test when some resources don't exist
3. **Flag combinations**: Test various flag combinations

## Success Criteria

- [x] Commands follow existing patterns (monitoring cleanup)
- [x] Executor pattern used for all multi-step operations
- [x] Both CLI and TUI modes work
- [x] Gracefully handles missing resources
- [x] Proper error messages
- [x] Updates TASKS.md when complete

## Open Questions

1. **Should we delete namespaces by default?** 
   - NO - require explicit flag
   
2. **Should we wait for deletion to complete?**
   - For CSVs: YES - wait with timeout
   - For CRs: Log deletion but don't wait (controller handles it)
   
3. **What if operators are from OperatorHub (not deployed by us)?**
   - Check CatalogSource name before deletion
   - Only delete if matches our constants

## Timeline

- Estimated: 2-3 hours total
  - Plan: 30 min ✅
  - Implement COO cleanup: 45 min
  - Implement Logging cleanup: 1 hour
  - Testing: 30 min
  - Documentation: 15 min
