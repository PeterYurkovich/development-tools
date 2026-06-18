# Plan: Implement OLM Utilities

**Task ID**: Supporting Infrastructure → OLM Utilities  
**Status**: ⏸️ Planning  
**Date**: 2026-06-18  
**Dependencies**: Implement k8s client package (✅ Complete)

---

## Overview

Implement the `pkg/operators/` package that provides utilities for working with the Operator Lifecycle Manager (OLM). These utilities will be used across all COO deployment methods (bundle, FBC, stage, operatorhub) and other operator deployments (logging, tracing, etc.).

This is foundational infrastructure needed before implementing any `deploy coo` or `cleanup coo` commands.

---

## Goals

1. **Create Subscription management utilities** for installing and managing operators via OLM
2. **Create helper functions** for waiting on CSV (ClusterServiceVersion) installation
3. **Create helper functions** for creating and managing CatalogSources
4. **Provide clean, reusable API** for all COO deployment methods
5. **Support progress reporting** via the executor/channel pattern for both CLI and TUI modes

---

## Non-Goals

- COO-specific deployment logic (belongs in `pkg/operators/coo/`)
- Bundle building or image management (developer workflow, not runtime)
- IDMS (ImageDigestMirrorSet) management (will be in `pkg/operators/coo/bundle.go`)
- OperatorGroup management (typically pre-existing in `openshift-operators`)
- Approval strategies beyond automatic (manual approval out of scope)

---

## Architecture Context

From CONTEXT.md and go-migration-plan.md:

**Key Decisions**:
- Use controller-runtime client directly (no abstraction)
- Business logic sends progress updates via channels
- OLM utilities provide reusable functions for Subscription, CSV, CatalogSource operations
- Mode-aware: works with both CLI and TUI via executor pattern

**Directory Structure**:
```
pkg/operators/
├── olm.go           # Generic OLM utilities (wait for CSV, check status)
├── subscription.go  # Subscription creation and management
└── coo/             # COO-specific deployment methods (future)
    ├── bundle.go
    ├── fbc.go
    ├── stage.go
    └── operatorhub.go
```

**Existing Related Files**:
- `pkg/k8s/scheme.go` - Already registers OLM schemes (Subscription, CSV, CatalogSource)
- `pkg/executor/executor.go` - Channel-based progress reporting
- `pkg/output/cli.go` - CLI output handler

---

## Detailed Design

### 1. File: `pkg/operators/subscription.go`

**Purpose**: Subscription creation and management utilities

**Key Functions**:

```go
package operators

import (
    "context"
    "fmt"
    
    operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "sigs.k8s.io/controller-runtime/pkg/client"
)

// SubscriptionConfig holds configuration for creating a Subscription
type SubscriptionConfig struct {
    Name             string
    Namespace        string
    Channel          string
    PackageName      string
    CatalogSource    string
    CatalogNamespace string
    StartingCSV      string  // Optional: specific version to start from
    InstallMode      string  // "AllNamespaces" or "OwnNamespace"
}

// CreateSubscription creates an OLM Subscription resource
func CreateSubscription(ctx context.Context, kubeClient client.Client, config SubscriptionConfig) error

// GetSubscription retrieves a Subscription by name
func GetSubscription(ctx context.Context, kubeClient client.Client, name, namespace string) (*operatorsv1alpha1.Subscription, error)

// DeleteSubscription deletes a Subscription
func DeleteSubscription(ctx context.Context, kubeClient client.Client, name, namespace string) error

// GetSubscriptionCSV returns the CSV name referenced by a Subscription
func GetSubscriptionCSV(ctx context.Context, kubeClient client.Client, name, namespace string) (string, error)
```

**Implementation Details**:

1. **CreateSubscription**:
   - Creates Subscription with proper owner references and labels
   - Sets approval strategy to Automatic
   - Handles InstallPlanApproval automatically
   - Returns error with context on failure

2. **GetSubscription**:
   - Uses client.Get() with proper error handling
   - Returns not-found vs other errors distinctly

3. **DeleteSubscription**:
   - Uses client.Delete() with background deletion policy
   - Idempotent (doesn't fail if already deleted)

4. **GetSubscriptionCSV**:
   - Reads Subscription.Status.InstalledCSV
   - Returns empty string if not yet installed
   - Returns error only on actual failures (not empty status)

---

### 2. File: `pkg/operators/catalogsource.go`

**Purpose**: CatalogSource creation and management utilities

**Key Functions**:

```go
package operators

import (
    "context"
    
    operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
    "sigs.k8s.io/controller-runtime/pkg/client"
)

// CatalogSourceConfig holds configuration for creating a CatalogSource
type CatalogSourceConfig struct {
    Name        string
    Namespace   string
    DisplayName string
    Publisher   string
    SourceType  string  // "grpc"
    Image       string
    Secrets     []string  // Optional: image pull secrets
}

// CreateCatalogSource creates an OLM CatalogSource resource
func CreateCatalogSource(ctx context.Context, kubeClient client.Client, config CatalogSourceConfig) error

// GetCatalogSource retrieves a CatalogSource by name
func GetCatalogSource(ctx context.Context, kubeClient client.Client, name, namespace string) (*operatorsv1alpha1.CatalogSource, error)

// DeleteCatalogSource deletes a CatalogSource
func DeleteCatalogSource(ctx context.Context, kubeClient client.Client, name, namespace string) error

// WaitForCatalogSourceReady waits for a CatalogSource to be ready
// Returns error on timeout or failure
func WaitForCatalogSourceReady(ctx context.Context, kubeClient client.Client, name, namespace string, timeout time.Duration) error
```

**Implementation Details**:

1. **CreateCatalogSource**:
   - Creates CatalogSource with grpc source type
   - Sets proper labels for identification
   - Handles image pull secrets if provided

2. **WaitForCatalogSourceReady**:
   - Polls CatalogSource.Status.ConnectionState.LastObservedState
   - Waits for state == "READY"
   - Times out after specified duration (default: 5 minutes)
   - Returns descriptive errors

---

### 3. File: `pkg/operators/olm.go`

**Purpose**: Generic OLM utilities for CSV management and waiting

**Key Functions**:

```go
package operators

import (
    "context"
    "time"
    
    operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
    "sigs.k8s.io/controller-runtime/pkg/client"
    
    "github.com/observability-ui/development-tools/pkg/executor"
)

// WaitForCSVSucceeded waits for a CSV to reach Succeeded phase
// Reports progress via executor if provided
func WaitForCSVSucceeded(
    ctx context.Context,
    kubeClient client.Client,
    csvName string,
    namespace string,
    timeout time.Duration,
    exec *executor.Executor,  // Optional: for progress reporting
    stepIndex int,            // Step index for progress updates
) error

// GetCSV retrieves a ClusterServiceVersion by name
func GetCSV(ctx context.Context, kubeClient client.Client, name, namespace string) (*operatorsv1alpha1.ClusterServiceVersion, error)

// DeleteCSV deletes a ClusterServiceVersion
func DeleteCSV(ctx context.Context, kubeClient client.Client, name, namespace string) error

// ListCSVsByPackage returns all CSVs for a given package name
func ListCSVsByPackage(ctx context.Context, kubeClient client.Client, packageName, namespace string) (*operatorsv1alpha1.ClusterServiceVersionList, error)
```

**Implementation Details**:

1. **WaitForCSVSucceeded**:
   - Polls CSV.Status.Phase every 5 seconds
   - Waits for Phase == "Succeeded"
   - Sends progress updates via executor if provided:
     - `StatusInProgress` when starting
     - `StatusComplete` when succeeded
     - `StatusFailed` if failed or timeout
   - Returns detailed error on failure or timeout
   - Default timeout: 10 minutes

2. **GetCSV**:
   - Standard Get operation with error handling
   - Returns not-found errors clearly

3. **DeleteCSV**:
   - Deletes CSV resource
   - Idempotent operation

4. **ListCSVsByPackage**:
   - Lists CSVs matching label selector
   - Uses `operators.coreos.com/packageName=<packageName>` label

---

### 4. Integration with Executor Pattern

**How OLM utilities integrate with progress reporting**:

```go
// Example: COO bundle deployment using OLM utilities
package coo

import (
    "context"
    "time"
    
    "github.com/observability-ui/development-tools/pkg/executor"
    "github.com/observability-ui/development-tools/pkg/operators"
)

const (
    StepCreateCatalogSource = iota
    StepWaitCatalogSource
    StepCreateSubscription
    StepWaitCSV
)

func DeployBundle(
    ctx context.Context,
    kubeClient client.Client,
    bundleImage string,
    exec *executor.Executor,
) error {
    defer exec.Close()
    
    // Step 1: Create CatalogSource
    exec.SendUpdate(StepCreateCatalogSource, executor.StatusInProgress, "Creating CatalogSource")
    
    catalogConfig := operators.CatalogSourceConfig{
        Name:        "observability-operator-bundle",
        Namespace:   "openshift-operators",
        DisplayName: "Observability Operator Bundle",
        Publisher:   "Red Hat",
        SourceType:  "grpc",
        Image:       bundleImage,
    }
    
    err := operators.CreateCatalogSource(ctx, kubeClient, catalogConfig)
    if err != nil {
        exec.SendUpdateWithError(StepCreateCatalogSource, executor.StatusFailed, "Creating CatalogSource", err)
        return err
    }
    exec.SendUpdate(StepCreateCatalogSource, executor.StatusComplete, "Creating CatalogSource")
    
    // Step 2: Wait for CatalogSource
    exec.SendUpdate(StepWaitCatalogSource, executor.StatusInProgress, "Waiting for CatalogSource")
    err = operators.WaitForCatalogSourceReady(ctx, kubeClient, "observability-operator-bundle", "openshift-operators", 5*time.Minute)
    if err != nil {
        exec.SendUpdateWithError(StepWaitCatalogSource, executor.StatusFailed, "Waiting for CatalogSource", err)
        return err
    }
    exec.SendUpdate(StepWaitCatalogSource, executor.StatusComplete, "Waiting for CatalogSource")
    
    // Step 3: Create Subscription
    exec.SendUpdate(StepCreateSubscription, executor.StatusInProgress, "Creating Subscription")
    
    subConfig := operators.SubscriptionConfig{
        Name:             "observability-operator",
        Namespace:        "openshift-operators",
        Channel:          "development",
        PackageName:      "observability-operator",
        CatalogSource:    "observability-operator-bundle",
        CatalogNamespace: "openshift-operators",
        InstallMode:      "AllNamespaces",
    }
    
    err = operators.CreateSubscription(ctx, kubeClient, subConfig)
    if err != nil {
        exec.SendUpdateWithError(StepCreateSubscription, executor.StatusFailed, "Creating Subscription", err)
        return err
    }
    exec.SendUpdate(StepCreateSubscription, executor.StatusComplete, "Creating Subscription")
    
    // Step 4: Wait for CSV
    exec.SendUpdate(StepWaitCSV, executor.StatusInProgress, "Waiting for CSV installation")
    
    // Get CSV name from Subscription
    csvName, err := operators.GetSubscriptionCSV(ctx, kubeClient, "observability-operator", "openshift-operators")
    if err != nil {
        exec.SendUpdateWithError(StepWaitCSV, executor.StatusFailed, "Getting CSV name", err)
        return err
    }
    
    // Wait for CSV to succeed (utility handles progress reporting)
    err = operators.WaitForCSVSucceeded(ctx, kubeClient, csvName, "openshift-operators", 10*time.Minute, exec, StepWaitCSV)
    if err != nil {
        return err  // Error already reported via executor
    }
    
    return nil
}
```

---

### 5. Constants and Configuration

**File**: `pkg/operators/constants.go` (optional, or keep in config package)

```go
package operators

import "time"

const (
    // Default timeouts
    DefaultCatalogSourceTimeout = 5 * time.Minute
    DefaultCSVTimeout          = 10 * time.Minute
    DefaultPollInterval        = 5 * time.Second
    
    // Common namespaces
    OpenShiftOperatorsNamespace = "openshift-operators"
    OpenShiftMarketplaceNamespace = "openshift-marketplace"
    
    // Approval strategies
    ApprovalAutomatic = "Automatic"
    ApprovalManual   = "Manual"
)
```

---

## Implementation Steps

### Step 1: Create `pkg/operators/subscription.go`

Implement the 4 subscription-related functions:
- `CreateSubscription`
- `GetSubscription`
- `DeleteSubscription`
- `GetSubscriptionCSV`

### Step 2: Create `pkg/operators/catalogsource.go`

Implement the 4 catalog source-related functions:
- `CreateCatalogSource`
- `GetCatalogSource`
- `DeleteCatalogSource`
- `WaitForCatalogSourceReady`

### Step 3: Create `pkg/operators/olm.go`

Implement the generic OLM utilities:
- `WaitForCSVSucceeded` (with executor integration)
- `GetCSV`
- `DeleteCSV`
- `ListCSVsByPackage`

### Step 4: Add helper constants (optional)

Create `pkg/operators/constants.go` or add to existing config package.

### Step 5: Verify Dependencies

OLM types are already registered in `pkg/k8s/scheme.go`:
- ✅ `operatorsv1alpha1.AddToScheme(scheme)` - line 28

OLM module already in go.mod:
- ✅ `github.com/operator-framework/api v0.43.0` - line 12

Run `go mod tidy` to ensure clean state:
```bash
go mod tidy
```

### Step 6: Test Compilation

```bash
go build ./pkg/operators/...
```

---

## Usage Examples

### Example 1: Deploy Operator via OperatorHub

```go
// Simple OperatorHub subscription (no CatalogSource needed)
subConfig := operators.SubscriptionConfig{
    Name:             "observability-operator",
    Namespace:        "openshift-operators",
    Channel:          "stable",
    PackageName:      "observability-operator",
    CatalogSource:    "redhat-operators",
    CatalogNamespace: "openshift-marketplace",
    InstallMode:      "AllNamespaces",
}

err := operators.CreateSubscription(ctx, kubeClient, subConfig)
if err != nil {
    return fmt.Errorf("failed to create subscription: %w", err)
}

// Wait for CSV
csvName, err := operators.GetSubscriptionCSV(ctx, kubeClient, "observability-operator", "openshift-operators")
if err != nil {
    return err
}

err = operators.WaitForCSVSucceeded(ctx, kubeClient, csvName, "openshift-operators", 10*time.Minute, nil, 0)
if err != nil {
    return fmt.Errorf("failed waiting for CSV: %w", err)
}
```

### Example 2: Deploy Operator via Custom CatalogSource (FBC)

```go
// Create custom CatalogSource
catalogConfig := operators.CatalogSourceConfig{
    Name:        "observability-fbc",
    Namespace:   "openshift-operators",
    DisplayName: "Observability Operator FBC",
    Publisher:   "Red Hat",
    SourceType:  "grpc",
    Image:       "quay.io/openshift-observability-ui/observability-fbc:latest",
}

err := operators.CreateCatalogSource(ctx, kubeClient, catalogConfig)
if err != nil {
    return err
}

// Wait for CatalogSource ready
err = operators.WaitForCatalogSourceReady(ctx, kubeClient, "observability-fbc", "openshift-operators", 5*time.Minute)
if err != nil {
    return err
}

// Create Subscription pointing to custom catalog
subConfig := operators.SubscriptionConfig{
    Name:             "observability-operator",
    Namespace:        "openshift-operators",
    Channel:          "development",
    PackageName:      "observability-operator",
    CatalogSource:    "observability-fbc",
    CatalogNamespace: "openshift-operators",
    InstallMode:      "AllNamespaces",
}

err = operators.CreateSubscription(ctx, kubeClient, subConfig)
// ... wait for CSV as in Example 1
```

### Example 3: Cleanup Operator

```go
// Delete Subscription
err := operators.DeleteSubscription(ctx, kubeClient, "observability-operator", "openshift-operators")
if err != nil {
    return err
}

// Get CSV name before deleting
csvName, _ := operators.GetSubscriptionCSV(ctx, kubeClient, "observability-operator", "openshift-operators")
if csvName != "" {
    err := operators.DeleteCSV(ctx, kubeClient, csvName, "openshift-operators")
    if err != nil {
        return err
    }
}

// Delete CatalogSource if custom one was created
err = operators.DeleteCatalogSource(ctx, kubeClient, "observability-fbc", "openshift-operators")
if err != nil {
    return err
}
```

---

## Testing Approach

Per the "minimal testing" philosophy:

**No unit tests for this task.** Manual validation will be performed when:
1. Implementing `deploy coo` commands
2. Testing with real cluster
3. Verifying all 4 deployment methods work

**Manual Testing Checklist** (for future implementation phase):
- [ ] CatalogSource creation and waiting
- [ ] Subscription creation
- [ ] CSV installation and waiting
- [ ] Progress reporting via executor
- [ ] Cleanup operations (delete Subscription, CSV, CatalogSource)
- [ ] Error handling for timeouts
- [ ] Error handling for missing resources

---

## Success Criteria

- [ ] `pkg/operators/subscription.go` created with 4 functions
- [ ] `pkg/operators/catalogsource.go` created with 4 functions
- [ ] `pkg/operators/olm.go` created with 4 functions
- [ ] All functions follow executor pattern for progress reporting
- [ ] Code compiles without errors: `go build ./pkg/operators/...`
- [ ] Dependencies verified: `go mod tidy` completes successfully
- [ ] Clear error messages with context wrapping
- [ ] Timeouts configurable with sensible defaults
- [ ] Ready for use in COO deployment methods

---

## Follow-up Tasks

After this task is complete:

1. **Implement Deploy COO Command** (TODO.md):
   - `cmd/deploy/coo.go` - command entry point
   - Uses OLM utilities created here

2. **Implement COO Bundle Deployment** (TODO.md):
   - `pkg/operators/coo/bundle.go`
   - Uses `CreateCatalogSource`, `CreateSubscription`, `WaitForCSVSucceeded`

3. **Implement COO FBC Deployment** (TODO.md):
   - `pkg/operators/coo/fbc.go`
   - Uses same OLM utilities

4. **Implement COO Stage Deployment** (TODO.md):
   - `pkg/operators/coo/stage.go`
   - Uses same OLM utilities with different catalog

5. **Implement COO OperatorHub Deployment** (TODO.md):
   - `pkg/operators/coo/operatorhub.go`
   - Uses `CreateSubscription` only (no custom CatalogSource)

6. **Implement Cleanup COO Command** (TODO.md):
   - `cmd/cleanup/coo.go`
   - Uses delete functions from OLM utilities

---

## Decisions to Confirm

### 1. File Structure
**Proposed**: Three files (`subscription.go`, `catalogsource.go`, `olm.go`)

**Alternative**: Single `pkg/operators/olm.go` with all functions

**Recommendation**: Three files for better organization and maintainability

### 2. Executor Integration
**Proposed**: `WaitForCSVSucceeded` takes optional executor parameter

**Alternative**: No executor integration in utilities (callers handle progress)

**Recommendation**: Integrate executor in wait functions since waiting is where progress matters most

### 3. Timeout Configuration
**Proposed**: Timeouts as function parameters with constants for defaults

**Alternative**: Global config for all timeouts

**Recommendation**: Function parameters with sensible defaults for flexibility

### 4. Error Handling
**Proposed**: Wrap all errors with context using `fmt.Errorf("context: %w", err)`

**Alternative**: Return raw errors from k8s client

**Recommendation**: Wrap with context for better debugging

### 5. Idempotency
**Proposed**: Delete operations are idempotent (don't fail if not found)

**Alternative**: Return error if resource doesn't exist

**Recommendation**: Idempotent deletes for easier cleanup logic

---

## Files to Create

```
pkg/operators/
├── subscription.go     # ~150 lines (4 functions)
├── catalogsource.go    # ~150 lines (4 functions)
├── olm.go             # ~200 lines (4 functions + polling logic)
└── constants.go       # ~20 lines (optional)
```

**Total**: ~520 lines of code

---

## Estimated Effort

- **Code Writing**: 2-3 hours (3 files, ~500 lines)
- **Testing/Debugging**: 1 hour (compile + verification)
- **Documentation**: Included in plan
- **Total**: ~3-4 hours

---

## References

- **TODO.md**: Lines 409-415 (Task definition)
- **CONTEXT.md**: Business logic decoupling pattern, executor usage
- **go-migration-plan.md**: Lines 145-154 (operators package structure)
- **pkg/executor/executor.go**: Progress reporting via channels
- **pkg/k8s/scheme.go**: OLM schemes already registered
- **OLM API docs**: https://pkg.go.dev/github.com/operator-framework/api
- **Bash reference**: `coo/dev-deploy.sh` - shows operator-sdk run bundle pattern

---

## Questions for Review

1. **File Organization**: Three separate files vs single file?
2. **Executor Integration**: Should wait functions take executor param or leave that to callers?
3. **Constants File**: Separate constants.go or inline constants?
4. **Default Timeouts**: Are 5 min (CatalogSource) and 10 min (CSV) reasonable?
5. **Idempotency**: Should delete operations be idempotent?

---

## Implementation Approved

**Status**: ⏸️ Awaiting approval

Once approved, ready for implementation with clear scope and structure.
