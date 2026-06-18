# Implementation: OLM Utilities

**Task ID**: Supporting Infrastructure → OLM Utilities  
**Status**: ✅ Complete  
**Date**: 2026-06-18  
**Implementation Time**: ~2 hours

---

## Summary

Successfully implemented a comprehensive OLM utilities package with 5 modules providing reusable functions for operator lifecycle management. All utilities integrate with the executor pattern for progress reporting and support both CLI and TUI modes.

---

## Files Created

### 1. `pkg/operators/subscription.go` (115 lines)

**Functions Implemented**:
- `CreateSubscription` - Creates OLM Subscription with automatic approval
- `GetSubscription` - Retrieves Subscription by name
- `DeleteSubscription` - Deletes Subscription (idempotent)
- `GetSubscriptionCSV` - Returns installed CSV name from Subscription
- `GetSubscriptionStatus` - Returns Subscription status
- `GetSubscriptionConditions` - Returns Subscription conditions

**Key Features**:
- Automatic InstallPlanApproval
- Proper labeling for package tracking
- Optional StartingCSV support
- Idempotent delete operation
- Clear error messages with context

### 2. `pkg/operators/catalogsource.go` (124 lines)

**Functions Implemented**:
- `CreateCatalogSource` - Creates OLM CatalogSource
- `GetCatalogSource` - Retrieves CatalogSource by name
- `DeleteCatalogSource` - Deletes CatalogSource (idempotent)
- `WaitForCatalogSourceReady` - Waits for CatalogSource READY state with executor integration

**Key Features**:
- GRPC source type support
- Image pull secrets support
- **TUI countdown**: Shows "next check in X seconds" during polling
- Executor integration for progress reporting
- 5 minute default timeout
- 5 second poll interval

**TUI Enhancement**:
```go
exec.SendLog(stepIndex, fmt.Sprintf("CatalogSource not ready (state: %s), next check in %d seconds",
    getConnectionState(state), int(remainingTime.Seconds())))
```

### 3. `pkg/operators/olm.go` (155 lines)

**Functions Implemented**:
- `WaitForCSVSucceeded` - Waits for CSV to reach Succeeded phase with executor integration
- `GetCSV` - Retrieves ClusterServiceVersion by name
- `DeleteCSV` - Deletes CSV (idempotent)
- `ListCSVsByPackage` - Lists all CSVs for a package
- `GetCSVPhase` - Returns current CSV phase

**Key Features**:
- **TUI countdown**: Shows "next check in X seconds" during CSV installation
- Full CSV phase tracking (Pending, Installing, Succeeded, Failed)
- Detailed error messages with phase and reason
- Executor integration with progress updates
- 10 minute default timeout
- 5 second poll interval

**TUI Enhancement**:
```go
exec.SendLog(stepIndex, fmt.Sprintf("%s, next check in %d seconds", statusMsg, nextCheckSeconds))
```

**Constants Defined**:
- `DefaultCSVTimeout = 10 * time.Minute`
- `CSVPollInterval = 5 * time.Second`
- `OpenShiftOperatorsNamespace = "openshift-operators"`
- `OpenShiftMarketplaceNamespace = "openshift-marketplace"`

### 4. `pkg/operators/operatorgroup.go` (118 lines)

**Functions Implemented**:
- `CreateOperatorGroup` - Creates OLM OperatorGroup
- `GetOperatorGroup` - Retrieves OperatorGroup by name
- `DeleteOperatorGroup` - Deletes OperatorGroup (idempotent)
- `ListOperatorGroups` - Lists all OperatorGroups in namespace
- `OperatorGroupExists` - Checks if any OperatorGroup exists in namespace
- `EnsureOperatorGroup` - Creates OperatorGroup only if none exists; returns (created bool, error)

**Key Features**:
- Target namespace configuration
- Managed resource labeling
- Existence check before creation
- Idempotent operations
- `EnsureOperatorGroup` returns bool indicating if new group was created (follows Ensure* pattern)

### 5. `pkg/operators/idms.go` (96 lines)

**Functions Implemented**:
- `CreateImageDigestMirrorSet` - Creates IDMS for image mirroring
- `GetImageDigestMirrorSet` - Retrieves IDMS by name
- `DeleteImageDigestMirrorSet` - Deletes IDMS (idempotent)
- `ListImageDigestMirrorSets` - Lists all IDMS resources
- `ImageDigestMirrorSetExists` - Checks if IDMS exists

**Key Features**:
- Support for multiple mirrors per source
- Cluster-scoped resource handling
- Managed resource labeling
- Proper type handling for ImageMirror

---

## Schema Registration

Updated `pkg/k8s/scheme.go` to register OperatorGroup types:

**Added Import**:
```go
operatorsv1 "github.com/operator-framework/api/pkg/operators/v1"
```

**Added Scheme Registration**:
```go
if err := operatorsv1.AddToScheme(scheme); err != nil {
    return fmt.Errorf("failed to register olm operators/v1 scheme: %w", err)
}
```

---

## Constants Organization

Per approved plan, constants are defined at the top of each file (not in a separate constants.go):

**catalogsource.go**:
- `DefaultCatalogSourceTimeout = 5 * time.Minute`
- `CatalogSourcePollInterval = 5 * time.Second`

**olm.go**:
- `DefaultCSVTimeout = 10 * time.Minute`
- `CSVPollInterval = 5 * time.Second`
- `OpenShiftOperatorsNamespace = "openshift-operators"`
- `OpenShiftMarketplaceNamespace = "openshift-marketplace"`

**subscription.go**:
- `ApprovalAutomatic = "Automatic"`
- `ApprovalManual = "Manual"`

---

## Executor Integration Pattern

All wait functions integrate with the executor pattern:

**Function Signature**:
```go
func WaitForCSVSucceeded(
    ctx context.Context,
    kubeClient client.Client,
    csvName string,
    namespace string,
    timeout time.Duration,
    exec *executor.Executor,  // Always present
    stepIndex int,            // Step index for progress updates
) error
```

**Progress Updates**:
- `StatusInProgress` - When starting wait
- `StatusComplete` - When resource ready/succeeded
- `StatusFailed` - On timeout or failure
- Log messages - For intermediate status updates with countdown

**TUI Countdown Feature**:
Both wait functions show countdown to next poll:
```go
nextCheckSeconds := int(PollInterval.Seconds())
exec.SendLog(stepIndex, fmt.Sprintf("Status: %s, next check in %d seconds", status, nextCheckSeconds))
```

---

## Error Handling

All functions follow consistent error handling:

**Error Wrapping**:
```go
return fmt.Errorf("failed to create subscription %s/%s: %w", namespace, name, err)
```

**Idempotent Deletes**:
```go
if errors.IsNotFound(err) {
    return nil  // Already deleted, not an error
}
```

**Not Found vs Other Errors**:
```go
if errors.IsNotFound(err) {
    return nil, fmt.Errorf("resource %s/%s not found", namespace, name)
}
return nil, fmt.Errorf("failed to get resource: %w", err)
```

---

## Usage Examples

### Example 1: Deploy Operator via Bundle

```go
package coo

import (
    "github.com/observability-ui/development-tools/pkg/operators"
    "github.com/observability-ui/development-tools/pkg/executor"
)

const (
    StepEnsureOperatorGroup = iota
    StepCreateCatalogSource
    StepWaitCatalogSource
    StepCreateSubscription
    StepWaitCSV
)

func DeployBundle(ctx context.Context, client client.Client, bundleImage string, exec *executor.Executor) error {
    defer exec.Close()
    
    // Ensure OperatorGroup exists
    exec.SendUpdate(StepEnsureOperatorGroup, executor.StatusInProgress, "Ensuring OperatorGroup")
    
    ogConfig := operators.OperatorGroupConfig{
        Name:             "global-operators",
        Namespace:        operators.OpenShiftOperatorsNamespace,
        TargetNamespaces: []string{},  // AllNamespaces
    }
    
    created, err := operators.EnsureOperatorGroup(ctx, client, ogConfig)
    if err != nil {
        exec.SendUpdateWithError(StepEnsureOperatorGroup, executor.StatusFailed, "Ensuring OperatorGroup", err)
        return err
    }
    
    if created {
        exec.SendLog(StepEnsureOperatorGroup, "Created new OperatorGroup")
    } else {
        exec.SendLog(StepEnsureOperatorGroup, "Using existing OperatorGroup")
    }
    exec.SendUpdate(StepEnsureOperatorGroup, executor.StatusComplete, "Ensuring OperatorGroup")
    
    // Create CatalogSource
    exec.SendUpdate(StepCreateCatalogSource, executor.StatusInProgress, "Creating CatalogSource")
    
    catalogConfig := operators.CatalogSourceConfig{
        Name:        "observability-bundle",
        Namespace:   operators.OpenShiftOperatorsNamespace,
        DisplayName: "Observability Operator Bundle",
        Publisher:   "Red Hat",
        SourceType:  "grpc",
        Image:       bundleImage,
    }
    
    err := operators.CreateCatalogSource(ctx, client, catalogConfig)
    if err != nil {
        exec.SendUpdateWithError(StepCreateCatalogSource, executor.StatusFailed, "Creating CatalogSource", err)
        return err
    }
    exec.SendUpdate(StepCreateCatalogSource, executor.StatusComplete, "Creating CatalogSource")
    
    // Wait for CatalogSource (with TUI countdown)
    err = operators.WaitForCatalogSourceReady(ctx, client, "observability-bundle", 
        operators.OpenShiftOperatorsNamespace, operators.DefaultCatalogSourceTimeout, exec, StepWaitCatalogSource)
    if err != nil {
        return err
    }
    
    // Create Subscription
    exec.SendUpdate(StepCreateSubscription, executor.StatusInProgress, "Creating Subscription")
    
    subConfig := operators.SubscriptionConfig{
        Name:             "observability-operator",
        Namespace:        operators.OpenShiftOperatorsNamespace,
        Channel:          "development",
        PackageName:      "observability-operator",
        CatalogSource:    "observability-bundle",
        CatalogNamespace: operators.OpenShiftOperatorsNamespace,
    }
    
    err = operators.CreateSubscription(ctx, client, subConfig)
    if err != nil {
        exec.SendUpdateWithError(StepCreateSubscription, executor.StatusFailed, "Creating Subscription", err)
        return err
    }
    exec.SendUpdate(StepCreateSubscription, executor.StatusComplete, "Creating Subscription")
    
    // Wait for CSV (with TUI countdown)
    csvName, _ := operators.GetSubscriptionCSV(ctx, client, "observability-operator", operators.OpenShiftOperatorsNamespace)
    err = operators.WaitForCSVSucceeded(ctx, client, csvName, operators.OpenShiftOperatorsNamespace, 
        operators.DefaultCSVTimeout, exec, StepWaitCSV)
    
    return err
}
```

### Example 2: Deploy Operator via OperatorHub

```go
func DeployOperatorHub(ctx context.Context, client client.Client, exec *executor.Executor) error {
    defer exec.Close()
    
    // No CatalogSource needed - use default catalog
    exec.SendUpdate(0, executor.StatusInProgress, "Creating Subscription")
    
    subConfig := operators.SubscriptionConfig{
        Name:             "observability-operator",
        Namespace:        operators.OpenShiftOperatorsNamespace,
        Channel:          "stable",
        PackageName:      "observability-operator",
        CatalogSource:    "redhat-operators",
        CatalogNamespace: operators.OpenShiftMarketplaceNamespace,
    }
    
    err := operators.CreateSubscription(ctx, client, subConfig)
    if err != nil {
        exec.SendUpdateWithError(0, executor.StatusFailed, "Creating Subscription", err)
        return err
    }
    exec.SendUpdate(0, executor.StatusComplete, "Creating Subscription")
    
    // Wait for installation
    csvName, _ := operators.GetSubscriptionCSV(ctx, client, "observability-operator", operators.OpenShiftOperatorsNamespace)
    return operators.WaitForCSVSucceeded(ctx, client, csvName, operators.OpenShiftOperatorsNamespace, 
        operators.DefaultCSVTimeout, exec, 1)
}
```

### Example 3: Cleanup Operator

```go
func CleanupOperator(ctx context.Context, client client.Client, exec *executor.Executor) error {
    defer exec.Close()
    
    namespace := operators.OpenShiftOperatorsNamespace
    
    // Step 1: Delete Subscription
    exec.SendUpdate(0, executor.StatusInProgress, "Deleting Subscription")
    err := operators.DeleteSubscription(ctx, client, "observability-operator", namespace)
    if err != nil {
        exec.SendUpdateWithError(0, executor.StatusFailed, "Deleting Subscription", err)
        return err
    }
    exec.SendUpdate(0, executor.StatusComplete, "Deleting Subscription")
    
    // Step 2: Delete CSV
    exec.SendUpdate(1, executor.StatusInProgress, "Deleting CSV")
    csvName, _ := operators.GetSubscriptionCSV(ctx, client, "observability-operator", namespace)
    if csvName != "" {
        err = operators.DeleteCSV(ctx, client, csvName, namespace)
        if err != nil {
            exec.SendUpdateWithError(1, executor.StatusFailed, "Deleting CSV", err)
            return err
        }
    }
    exec.SendUpdate(1, executor.StatusComplete, "Deleting CSV")
    
    // Step 3: Delete CatalogSource (if custom)
    exec.SendUpdate(2, executor.StatusInProgress, "Deleting CatalogSource")
    err = operators.DeleteCatalogSource(ctx, client, "observability-bundle", namespace)
    if err != nil {
        exec.SendUpdateWithError(2, executor.StatusFailed, "Deleting CatalogSource", err)
        return err
    }
    exec.SendUpdate(2, executor.StatusComplete, "Deleting CatalogSource")
    
    return nil
}
```

---

## Testing

Per minimal testing philosophy, no unit tests created. Manual testing will be performed when implementing:
- `deploy coo --method=bundle`
- `deploy coo --method=fbc`
- `deploy coo --method=stage`
- `deploy coo --method=operatorhub`
- `cleanup coo`

---

## Verification

**Compilation**: ✅ All files compile successfully
```bash
go build ./pkg/operators/...
go build ./...
```

**Dependencies**: ✅ All resolved
```bash
go mod tidy
```

**Scheme Registration**: ✅ operatorsv1 and operatorsv1alpha1 registered in pkg/k8s/scheme.go

**Code Statistics**:
- Total Lines: ~608
- Files Created: 5
- Functions Implemented: 26
- Constants Defined: 7

---

## Design Decisions Implemented

1. ✅ **Three separate files** - Better organization (subscription, catalogsource, olm, operatorgroup, idms)
2. ✅ **Executor integration** - All wait functions take executor parameter
3. ✅ **Constants at file top** - No separate constants.go file
4. ✅ **Default timeouts** - 5 min for CatalogSource, 10 min for CSV
5. ✅ **Idempotent deletes** - All delete operations ignore NotFound errors
6. ✅ **TUI countdown** - Shows "next check in X seconds" during polling
7. ✅ **OperatorGroup support** - Generalized implementation with EnsureOperatorGroup
8. ✅ **IDMS support** - Generalized implementation for image mirroring
9. ✅ **Ensure* pattern** - Functions return (bool, error) indicating if resource was created

---

## Follow-up Tasks

Ready to implement (in order):

1. **Deploy COO Command** (`cmd/deploy/coo.go`)
   - Command entry point with --method flag
   - Routes to appropriate deployment method

2. **COO Bundle Deployment** (`pkg/operators/coo/bundle.go`)
   - Uses: CreateCatalogSource, WaitForCatalogSourceReady, CreateSubscription, WaitForCSVSucceeded
   - IDMS creation if needed

3. **COO FBC Deployment** (`pkg/operators/coo/fbc.go`)
   - Uses same OLM utilities with FBC catalog

4. **COO Stage Deployment** (`pkg/operators/coo/stage.go`)
   - Uses same OLM utilities with stage registry

5. **COO OperatorHub Deployment** (`pkg/operators/coo/operatorhub.go`)
   - Uses CreateSubscription only (no custom CatalogSource)

6. **Cleanup COO Command** (`cmd/cleanup/coo.go`)
   - Uses delete functions from OLM utilities

---

## Success Criteria

- ✅ 5 files created (subscription, catalogsource, olm, operatorgroup, idms)
- ✅ 26 functions implemented
- ✅ All functions integrate with executor pattern
- ✅ TUI countdown feature implemented
- ✅ Code compiles without errors
- ✅ go mod tidy completes successfully
- ✅ Clear error messages with context wrapping
- ✅ Timeouts configurable with sensible defaults
- ✅ Idempotent delete operations
- ✅ OperatorGroup management included
- ✅ IDMS management included
- ✅ Ready for use in COO deployment methods

---

## Coding Standards Added

**New Standard - Ensure Functions**:
Added to both `CONTEXT.md` and `go-migration-plan.md`:

```
Ensure Functions:
- Functions named Ensure* that create resources idempotently must return (bool, error)
- Return (true, nil) when a new resource was created
- Return (false, nil) when an existing resource was found
- Return (false, err) on error
- Allows callers to know whether action was taken or resource already existed
- Example: func EnsureOperatorGroup(...) (bool, error) returns whether it created a new group
```

**Usage Example**:
```go
created, err := operators.EnsureOperatorGroup(ctx, client, config)
if err != nil {
    return err
}

if created {
    log.Info("Created new OperatorGroup")
} else {
    log.Info("Using existing OperatorGroup")
}
```

---

## Lessons Learned

1. **ImageMirror Type**: IDMS requires `[]configv1.ImageMirror` not `[]string` for mirrors
2. **Unused Imports**: Removed unnecessary `corev1` imports
3. **OperatorGroup v1**: Uses `operatorsv1` not `operatorsv1alpha1`
4. **Scheme Registration**: Both v1 and v1alpha1 needed for complete OLM support
5. **Countdown Calculation**: Use `int(duration.Seconds())` for TUI countdown display
6. **Ensure Pattern**: Functions that idempotently create resources should return whether they created vs found existing

---

## Implementation Complete

All OLM utilities are now ready for use in COO deployment commands and other operator deployments.

**Total Implementation Time**: ~2 hours  
**Files Modified**: 6 (5 new + 1 updated scheme)  
**Lines of Code**: ~608
