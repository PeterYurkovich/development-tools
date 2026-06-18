# Business Logic Decoupling Proposal

**Purpose**: Decouple business logic from display logic using Go channels  
**Problem**: Current implementation duplicates business logic in CLI and TUI modes  
**Solution**: Single business logic function sends progress updates via channels  
**Date**: 2026-06-18

---

## Table of Contents

1. [Problem Analysis](#problem-analysis)
2. [Proposed Architecture](#proposed-architecture)
3. [Implementation Pattern](#implementation-pattern)
4. [Code Examples](#code-examples)
5. [Migration Strategy](#migration-strategy)
6. [Benefits & Trade-offs](#benefits--trade-offs)

---

## Problem Analysis

### Current Implementation (update monitoring)

**File**: `cmd/update/monitoring.go`

```go
func runUpdateMonitoringCLI(cmd *cobra.Command) error {
    image, _ := cmd.Flags().GetString("image")
    // ...
    
    out := output.NewHandler(ctx)
    
    // BUSINESS LOGIC DUPLICATED HERE
    out.Progress("Scaling down CMO...")
    if err := k8s.ScaleDeployment(ctx, kubeClient, constants.CMODeployment, constants.MonitoringNamespace, 0); err != nil {
        out.Error(fmt.Sprintf("Failed to scale down CMO: %v", err))
        return err
    }
    out.Success(fmt.Sprintf("Scaled down %s", constants.CMODeployment))
    
    out.Progress("Updating monitoring plugin image...")
    if err := k8s.UpdateDeploymentImage(ctx, kubeClient, constants.PluginDeployment, constants.MonitoringNamespace, image); err != nil {
        out.Error(fmt.Sprintf("Failed to update image: %v", err))
        return err
    }
    out.Success(fmt.Sprintf("Updated %s image to %s", constants.PluginDeployment, image))
    
    return nil
}

func runUpdateMonitoringTUI(cmd *cobra.Command) error {
    // ... input collection ...
    
    operations := []string{
        fmt.Sprintf("Scale down %s", constants.CMODeployment),
        fmt.Sprintf("Update %s image to %s", constants.PluginDeployment, image),
    }
    
    model := tui.NewProgressModel("Updating Monitoring Plugin", operations)
    program := tea.NewProgram(model)
    
    // SAME BUSINESS LOGIC DUPLICATED HERE
    go func() {
        program.Send(tui.OperationUpdateMsg{Index: 0, Status: tui.OperationInProgress})
        
        err := k8s.ScaleDeployment(ctx, kubeClient, constants.CMODeployment, constants.MonitoringNamespace, 0)
        if err != nil {
            program.Send(tui.OperationUpdateMsg{Index: 0, Status: tui.OperationFailed, Error: err})
            return
        }
        program.Send(tui.OperationUpdateMsg{Index: 0, Status: tui.OperationComplete})
        
        program.Send(tui.OperationUpdateMsg{Index: 1, Status: tui.OperationInProgress})
        
        err = k8s.UpdateDeploymentImage(ctx, kubeClient, constants.PluginDeployment, constants.MonitoringNamespace, image)
        if err != nil {
            program.Send(tui.OperationUpdateMsg{Index: 1, Status: tui.OperationFailed, Error: err})
            return
        }
        program.Send(tui.OperationUpdateMsg{Index: 1, Status: tui.OperationComplete})
    }()
    
    finalModel, err := program.Run()
    // ...
}
```

### Problems

1. **Code Duplication**: Same K8s operations in two places
2. **Maintenance Burden**: Bug fixes need to be applied twice
3. **Inconsistency Risk**: TUI and CLI could diverge over time
4. **Scalability**: With 10+ commands, this becomes 20+ functions
5. **Testing Complexity**: Must test business logic twice

---

## Proposed Architecture

### High-Level Design

```
┌─────────────────────────────────────────────────────────┐
│                     Command Layer                        │
│  (runUpdateMonitoring, runCleanupMonitoring, etc.)      │
└───────────────────┬─────────────────────────────────────┘
                    │
        ┌───────────┴──────────┐
        │                      │
        ▼                      ▼
┌──────────────┐      ┌──────────────┐
│  TUI Runner  │      │  CLI Runner  │
│              │      │              │
│ - Creates    │      │ - Creates    │
│   progress   │      │   output     │
│   model      │      │   handler    │
│ - Receives   │      │ - Receives   │
│   updates    │      │   updates    │
│ - Displays   │      │ - Displays   │
└──────┬───────┘      └──────┬───────┘
       │                     │
       │   ProgressUpdate    │
       │      Channel        │
       └──────────┬──────────┘
                  │
                  ▼
      ┌────────────────────────┐
      │   Business Logic       │
      │                        │
      │ - Single function      │
      │ - Sends updates to ch  │
      │ - No display code      │
      │ - Fully testable       │
      └────────────────────────┘
```

### Channel Protocol

Business logic sends **ProgressUpdate** messages:

```go
type ProgressUpdate struct {
    Step    string        // Operation description
    Status  UpdateStatus  // Pending, InProgress, Complete, Failed
    Error   error         // Error if failed (only set when StatusFailed)
    Index   int           // Step index (for ordered operations)
    Message string        // Log message for progress updates (optional)
}

type UpdateStatus int

const (
    StatusPending UpdateStatus = iota
    StatusInProgress
    StatusComplete
    StatusFailed
)
```

**Key Fields**:
- **Step**: Human-readable operation name (e.g., "Scale down cluster-monitoring-operator")
- **Status**: Current state of the operation
- **Error**: Only populated when Status is StatusFailed
- **Index**: Numeric step identifier (use enumerated constants, not raw numbers)
- **Message**: Optional log message for progress details (e.g., "Waiting for pods to terminate...")

**Message Usage**:
- **CLI Mode**: Messages are displayed as log output
- **TUI Mode**: Messages can be ignored or shown as sub-text
- **Example**: During a long operation, send StatusInProgress multiple times with different Messages

---

## Implementation Pattern

### Step 1: Define Progress Update Types

**File**: `pkg/executor/executor.go` (new package)

```go
package executor

type UpdateStatus int

const (
    StatusPending UpdateStatus = iota
    StatusInProgress
    StatusComplete
    StatusFailed
)

type ProgressUpdate struct {
    Step    string
    Status  UpdateStatus
    Error   error
    Index   int
    Message string
}

type Executor struct {
    UpdateCh chan ProgressUpdate
}

func NewExecutor() *Executor {
    return &Executor{
        UpdateCh: make(chan ProgressUpdate, 10),
    }
}

func (e *Executor) SendUpdate(index int, status UpdateStatus, step string) {
    e.UpdateCh <- ProgressUpdate{
        Index:  index,
        Status: status,
        Step:   step,
    }
}

func (e *Executor) SendUpdateWithError(index int, status UpdateStatus, step string, err error) {
    e.UpdateCh <- ProgressUpdate{
        Index:  index,
        Status: status,
        Step:   step,
        Error:  err,
    }
}

func (e *Executor) SendLog(index int, message string) {
    e.UpdateCh <- ProgressUpdate{
        Index:   index,
        Status:  StatusInProgress,
        Message: message,
    }
}

func (e *Executor) Close() {
    close(e.UpdateCh)
}
```

**Key Methods**:
- `SendUpdate()`: Send status change (InProgress, Complete, etc.)
- `SendUpdateWithError()`: Send status change with error (for Failed status)
- `SendLog()`: Send progress log message (CLI displays immediately, TUI can ignore)
- `Close()`: Signal completion (must be called when done)

### Step 2: Refactor Business Logic

**File**: `pkg/operations/monitoring.go` (new package)

```go
package operations

import (
    "context"
    "fmt"
    
    "github.com/observability-ui/development-tools/internal/constants"
    "github.com/observability-ui/development-tools/pkg/executor"
    "github.com/observability-ui/development-tools/pkg/k8s"
    "sigs.k8s.io/controller-runtime/pkg/client"
)

// Step indexes - enumerate for clarity
const (
    StepScaleDownCMO = iota
    StepUpdatePluginImage
)

type UpdateMonitoringConfig struct {
    Image string
}

func UpdateMonitoring(ctx context.Context, kubeClient client.Client, config UpdateMonitoringConfig, exec *executor.Executor) error {
    defer exec.Close()
    
    // Step 0: Scale down CMO
    stepName := fmt.Sprintf("Scale down %s", constants.CMODeployment)
    exec.SendUpdate(StepScaleDownCMO, executor.StatusInProgress, stepName)
    
    exec.SendLog(StepScaleDownCMO, "Updating deployment replicas to 0")
    err := k8s.ScaleDeployment(ctx, kubeClient, constants.CMODeployment, constants.MonitoringNamespace, 0)
    if err != nil {
        exec.SendUpdateWithError(StepScaleDownCMO, executor.StatusFailed, stepName, err)
        return err
    }
    
    exec.SendLog(StepScaleDownCMO, "Deployment scaled down successfully")
    exec.SendUpdate(StepScaleDownCMO, executor.StatusComplete, stepName)
    
    // Step 1: Update plugin image
    stepName = fmt.Sprintf("Update %s image to %s", constants.PluginDeployment, config.Image)
    exec.SendUpdate(StepUpdatePluginImage, executor.StatusInProgress, stepName)
    
    exec.SendLog(StepUpdatePluginImage, fmt.Sprintf("Patching deployment with new image: %s", config.Image))
    err = k8s.UpdateDeploymentImage(ctx, kubeClient, constants.PluginDeployment, constants.MonitoringNamespace, config.Image)
    if err != nil {
        exec.SendUpdateWithError(StepUpdatePluginImage, executor.StatusFailed, stepName, err)
        return err
    }
    
    exec.SendLog(StepUpdatePluginImage, "Image updated successfully")
    exec.SendUpdate(StepUpdatePluginImage, executor.StatusComplete, stepName)
    
    return nil
}
```

**Best Practices Demonstrated**:
1. ✅ **Enumerate step indexes**: `StepScaleDownCMO` instead of raw `0`
2. ✅ **Send log messages**: `SendLog()` for progress details (CLI displays, TUI ignores)
3. ✅ **Use helper methods**: `SendUpdate()` vs `SendUpdateWithError()` for clarity
4. ✅ **Close in defer**: Ensures channel is always closed
5. ✅ **Descriptive step names**: Clear human-readable descriptions

### Step 3: TUI Runner

**File**: `cmd/update/monitoring.go` (refactored)

```go
func runUpdateMonitoringTUI(cmd *cobra.Command) error {
    ctx := cmd.Context()
    ctx = execctx.WithTUI(ctx, true)
    
    kubeClient, err := execctx.GetClient(ctx)
    if err != nil {
        return err
    }
    
    image, _ := cmd.Flags().GetString("image")
    if image == "" {
        image, err = collectImageInput()
        if err != nil {
            return err
        }
    }
    
    // Create executor with channel
    exec := executor.NewExecutor()
    
    // Create TUI progress model
    operations := []string{
        "Scale down cluster-monitoring-operator",
        "Update monitoring-plugin image",
    }
    model := tui.NewProgressModel("Updating Monitoring Plugin", operations)
    program := tea.NewProgram(model)
    
    // Run business logic in goroutine
    go operations.UpdateMonitoring(ctx, kubeClient, operations.UpdateMonitoringConfig{
        Image: image,
    }, exec)
    
    // Forward channel updates to TUI
    go func() {
        for update := range exec.UpdateCh {
            // Skip log messages - TUI only cares about status changes
            if update.Message != "" {
                continue
            }
            
            program.Send(tui.OperationUpdateMsg{
                Index:  update.Index,
                Status: convertStatus(update.Status),
                Error:  update.Error,
            })
        }
    }()
    
    finalModel, err := program.Run()
    if err != nil {
        return err
    }
    
    m := finalModel.(tui.ProgressModel)
    return m.Error()
}

func convertStatus(status executor.UpdateStatus) tui.OperationStatus {
    switch status {
    case executor.StatusPending:
        return tui.OperationPending
    case executor.StatusInProgress:
        return tui.OperationInProgress
    case executor.StatusComplete:
        return tui.OperationComplete
    case executor.StatusFailed:
        return tui.OperationFailed
    default:
        return tui.OperationPending
    }
}
```

### Step 4: CLI Runner

**File**: `cmd/update/monitoring.go` (refactored)

```go
func runUpdateMonitoringCLI(cmd *cobra.Command) error {
    image, _ := cmd.Flags().GetString("image")
    if image == "" {
        return fmt.Errorf("--image flag is required")
    }
    
    ctx := cmd.Context()
    ctx = execctx.WithTUI(ctx, false)
    
    kubeClient, err := execctx.GetClient(ctx)
    if err != nil {
        return err
    }
    
    out := output.NewHandler(ctx)
    
    // Create executor with channel
    exec := executor.NewExecutor()
    
    // Run business logic in goroutine
    go operations.UpdateMonitoring(ctx, kubeClient, operations.UpdateMonitoringConfig{
        Image: image,
    }, exec)
    
    // Display updates from channel
    for update := range exec.UpdateCh {
        // Handle log messages
        if update.Message != "" {
            out.Info(update.Message)
            continue
        }
        
        // Handle status changes
        switch update.Status {
        case executor.StatusInProgress:
            out.Progress(update.Step + "...")
        case executor.StatusComplete:
            out.Success(update.Step)
        case executor.StatusFailed:
            out.Error(fmt.Sprintf("%s: %v", update.Step, update.Error))
            return update.Error
        }
    }
    
    return nil
}
```

**Key Changes**:
1. ✅ **Handle log messages**: Check for `update.Message` and display with `Info()`
2. ✅ **Separate concerns**: Log messages vs status updates
3. ✅ **Continue after logs**: Log messages don't affect status flow

---

## Code Examples

### Before: Duplicated Logic

**CLI Version** (50 lines):
```go
func runUpdateMonitoringCLI(cmd *cobra.Command) error {
    // Get image from flags
    // Create client
    // Create output handler
    
    // BUSINESS LOGIC START
    out.Progress("Scaling down CMO...")
    if err := k8s.ScaleDeployment(...); err != nil {
        out.Error(...)
        return err
    }
    out.Success(...)
    
    out.Progress("Updating image...")
    if err := k8s.UpdateDeploymentImage(...); err != nil {
        out.Error(...)
        return err
    }
    out.Success(...)
    // BUSINESS LOGIC END
    
    return nil
}
```

**TUI Version** (80 lines):
```go
func runUpdateMonitoringTUI(cmd *cobra.Command) error {
    // Collect input if needed
    // Create client
    // Create progress model
    
    go func() {
        // BUSINESS LOGIC START (duplicated)
        program.Send(...)
        err := k8s.ScaleDeployment(...)
        if err != nil {
            program.Send(...)
            return
        }
        program.Send(...)
        
        program.Send(...)
        err = k8s.UpdateDeploymentImage(...)
        if err != nil {
            program.Send(...)
            return
        }
        program.Send(...)
        // BUSINESS LOGIC END
    }()
    
    program.Run()
}
```

**Total**: 130 lines, business logic duplicated

### After: Decoupled Logic

**Business Logic** (30 lines):
```go
func UpdateMonitoring(ctx context.Context, client client.Client, config UpdateMonitoringConfig, exec *executor.Executor) error {
    defer exec.Close()
    
    exec.SendUpdate(0, StatusInProgress, nil, "Scale down CMO")
    err := k8s.ScaleDeployment(ctx, client, ...)
    if err != nil {
        exec.SendUpdate(0, StatusFailed, err, "Scale down CMO")
        return err
    }
    exec.SendUpdate(0, StatusComplete, nil, "Scale down CMO")
    
    exec.SendUpdate(1, StatusInProgress, nil, "Update image")
    err = k8s.UpdateDeploymentImage(ctx, client, ...)
    if err != nil {
        exec.SendUpdate(1, StatusFailed, err, "Update image")
        return err
    }
    exec.SendUpdate(1, StatusComplete, nil, "Update image")
    
    return nil
}
```

**CLI Runner** (25 lines):
```go
func runUpdateMonitoringCLI(cmd *cobra.Command) error {
    // Setup
    exec := executor.NewExecutor()
    
    // Run business logic
    go operations.UpdateMonitoring(ctx, client, config, exec)
    
    // Display updates
    for update := range exec.UpdateCh {
        displayUpdate(update, out)
    }
    
    return nil
}
```

**TUI Runner** (30 lines):
```go
func runUpdateMonitoringTUI(cmd *cobra.Command) error {
    // Setup
    exec := executor.NewExecutor()
    
    // Run business logic
    go operations.UpdateMonitoring(ctx, client, config, exec)
    
    // Forward to TUI
    go func() {
        for update := range exec.UpdateCh {
            program.Send(convertToTUIUpdate(update))
        }
    }()
    
    return program.Run()
}
```

**Total**: 85 lines, business logic written once

**Savings**: 45 lines, no duplication

---

## Migration Strategy

### Phase 1: Create Infrastructure

**New Files**:
1. `pkg/executor/executor.go` - Executor and ProgressUpdate types
2. `pkg/operations/` - New package for business logic

**Estimated Time**: 1 hour

### Phase 2: Refactor Existing Commands

**Order** (easiest to hardest):
1. ✅ `cleanup monitoring` - Single operation
2. ✅ `update monitoring` - Two operations
3. `cleanup coo` - Multiple operations
4. `deploy logging` - Complex with conditional steps
5. `deploy coo` - Most complex with method switching

**Per Command**: ~30 minutes

### Phase 3: Establish Pattern for New Commands

**Template**: New commands follow the decoupled pattern from the start

**Benefits**:
- No duplication in new commands
- Easier testing
- Consistent structure

---

## Benefits & Trade-offs

### Benefits

| Benefit | Description |
|---------|-------------|
| **Single Source of Truth** | Business logic written once |
| **Easier Testing** | Test business logic without display concerns |
| **Consistency** | TUI and CLI always execute same logic |
| **Maintainability** | Bug fixes applied once |
| **Scalability** | Pattern works for any number of commands |
| **Flexibility** | Easy to add new display modes (e.g., JSON output) |
| **Separation of Concerns** | Clean architecture |

### Trade-offs

| Trade-off | Mitigation |
|-----------|------------|
| **More files** | Organized into clear packages |
| **Channel overhead** | Negligible - buffered channels, few messages |
| **Complexity** | Well-documented pattern, reusable template |
| **Learning curve** | Go channels primer provided |

### Performance Impact

**Negligible**:
- Channels are very fast (millions of ops/sec)
- Buffered channels prevent blocking
- Few messages per command (2-10 typically)
- K8s API calls dominate execution time

**Benchmark** (estimated):
- Channel send/receive: ~100ns
- K8s API call: ~50-500ms
- Ratio: Channel overhead is 0.0002% of total time

---

## Testing Benefits

### Before: Hard to Test

```go
// Can't test business logic without mocking output.Handler and TUI
func TestUpdateMonitoringCLI(t *testing.T) {
    // Must mock entire CLI execution
    // Must parse stdout to verify behavior
    // Fragile
}
```

### After: Easy to Test

```go
func TestUpdateMonitoring(t *testing.T) {
    ctx := context.Background()
    client := fake.NewClientBuilder().Build()
    exec := executor.NewExecutor()
    
    go operations.UpdateMonitoring(ctx, client, operations.UpdateMonitoringConfig{
        Image: "test:v1",
    }, exec)
    
    updates := []executor.ProgressUpdate{}
    for update := range exec.UpdateCh {
        updates = append(updates, update)
    }
    
    // Verify sequence of updates
    assert.Equal(t, 4, len(updates))  // 2 in-progress, 2 complete
    assert.Equal(t, executor.StatusInProgress, updates[0].Status)
    assert.Equal(t, executor.StatusComplete, updates[1].Status)
    // etc.
}
```

**Benefits**:
- No mocking of display layer
- Direct verification of business logic
- Fast (no actual K8s calls needed with fake client)
- Clear test structure

---

## Recommended Implementation Order

### Step 1: Create Infrastructure (30 min)

1. Create `pkg/executor/executor.go`:
   ```go
   type ProgressUpdate struct { ... }
   type Executor struct { ... }
   func NewExecutor() *Executor { ... }
   func (e *Executor) SendUpdate(...) { ... }
   func (e *Executor) Close() { ... }
   ```

2. Create `pkg/operations/` package

### Step 2: Migrate Cleanup Monitoring (30 min)

**Why first?** Simplest - single operation

1. Create `pkg/operations/monitoring.go`
2. Extract business logic to `CleanupMonitoring()` function
3. Refactor `cmd/cleanup/monitoring.go` to use channels
4. Test both CLI and TUI modes
5. Verify no regressions

### Step 3: Migrate Update Monitoring (30 min)

**Why second?** Two operations, already familiar from cleanup

1. Add `UpdateMonitoring()` to `pkg/operations/monitoring.go`
2. Refactor `cmd/update/monitoring.go` to use channels
3. Test both CLI and TUI modes
4. Verify no regressions

### Step 4: Document Pattern (15 min)

1. Create template in `tmp/docs/command-implementation-template.md`
2. Show step-by-step how to create new command with decoupled logic
3. Update CONTEXT.md with pattern reference

### Step 5: Apply to Future Commands

All new commands follow the template:
- Business logic in `pkg/operations/`
- Command files use `executor.Executor`
- No duplication

---

## Example: Full Refactored Command

### Structure

```
pkg/operations/monitoring.go  - Business logic
cmd/update/monitoring.go      - Command (uses operations)
cmd/cleanup/monitoring.go     - Command (uses operations)
```

### Business Logic

**File**: `pkg/operations/monitoring.go`

```go
package operations

import (
    "context"
    "fmt"
    
    "github.com/observability-ui/development-tools/internal/constants"
    "github.com/observability-ui/development-tools/pkg/executor"
    "github.com/observability-ui/development-tools/pkg/k8s"
    "sigs.k8s.io/controller-runtime/pkg/client"
)

type UpdateMonitoringConfig struct {
    Image string
}

func UpdateMonitoring(ctx context.Context, kubeClient client.Client, config UpdateMonitoringConfig, exec *executor.Executor) error {
    defer exec.Close()
    
    stepIndex := 0
    
    // Step 0: Scale down CMO
    exec.SendUpdate(stepIndex, executor.StatusInProgress, nil, 
        fmt.Sprintf("Scale down %s", constants.CMODeployment))
    
    err := k8s.ScaleDeployment(ctx, kubeClient, constants.CMODeployment, constants.MonitoringNamespace, 0)
    if err != nil {
        exec.SendUpdate(stepIndex, executor.StatusFailed, err, 
            fmt.Sprintf("Scale down %s", constants.CMODeployment))
        return err
    }
    
    exec.SendUpdate(stepIndex, executor.StatusComplete, nil, 
        fmt.Sprintf("Scale down %s", constants.CMODeployment))
    
    stepIndex++
    
    // Step 1: Update plugin image
    exec.SendUpdate(stepIndex, executor.StatusInProgress, nil, 
        fmt.Sprintf("Update %s image to %s", constants.PluginDeployment, config.Image))
    
    err = k8s.UpdateDeploymentImage(ctx, kubeClient, constants.PluginDeployment, constants.MonitoringNamespace, config.Image)
    if err != nil {
        exec.SendUpdate(stepIndex, executor.StatusFailed, err, 
            fmt.Sprintf("Update %s image", constants.PluginDeployment))
        return err
    }
    
    exec.SendUpdate(stepIndex, executor.StatusComplete, nil, 
        fmt.Sprintf("Update %s image to %s", constants.PluginDeployment, config.Image))
    
    return nil
}

type CleanupMonitoringConfig struct{}

func CleanupMonitoring(ctx context.Context, kubeClient client.Client, config CleanupMonitoringConfig, exec *executor.Executor) error {
    defer exec.Close()
    
    exec.SendUpdate(0, executor.StatusInProgress, nil, 
        fmt.Sprintf("Scale up %s", constants.CMODeployment))
    
    err := k8s.ScaleDeployment(ctx, kubeClient, constants.CMODeployment, constants.MonitoringNamespace, 1)
    if err != nil {
        exec.SendUpdate(0, executor.StatusFailed, err, 
            fmt.Sprintf("Scale up %s", constants.CMODeployment))
        return err
    }
    
    exec.SendUpdate(0, executor.StatusComplete, nil, 
        fmt.Sprintf("Scale up %s", constants.CMODeployment))
    
    return nil
}
```

### Command Implementation

**File**: `cmd/update/monitoring.go`

```go
package update

import (
    "fmt"
    
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/huh"
    "github.com/spf13/cobra"
    
    execctx "github.com/observability-ui/development-tools/pkg/context"
    "github.com/observability-ui/development-tools/pkg/executor"
    "github.com/observability-ui/development-tools/pkg/mode"
    "github.com/observability-ui/development-tools/pkg/operations"
    "github.com/observability-ui/development-tools/pkg/output"
    "github.com/observability-ui/development-tools/pkg/tui"
)

var monitoringCmd = &cobra.Command{
    Use:   "monitoring",
    Short: "Update monitoring plugin image",
    Long:  "Scale down CMO to allow updating the monitoring plugin image",
    RunE:  runUpdateMonitoring,
}

func init() {
    monitoringCmd.Flags().String("image", "", "Monitoring plugin image to use")
    UpdateCmd.AddCommand(monitoringCmd)
}

func runUpdateMonitoring(cmd *cobra.Command, args []string) error {
    requiredFlags := []string{"image"}
    
    useTUI, err := mode.DetermineMode(cmd, requiredFlags)
    if err != nil {
        return err
    }
    
    if useTUI {
        return runUpdateMonitoringTUI(cmd)
    }
    
    return runUpdateMonitoringCLI(cmd)
}

func runUpdateMonitoringCLI(cmd *cobra.Command) error {
    image, _ := cmd.Flags().GetString("image")
    if image == "" {
        return fmt.Errorf("--image flag is required")
    }
    
    ctx := cmd.Context()
    ctx = execctx.WithTUI(ctx, false)
    
    kubeClient, err := execctx.GetClient(ctx)
    if err != nil {
        return err
    }
    
    out := output.NewHandler(ctx)
    exec := executor.NewExecutor()
    
    // Run business logic in goroutine
    go operations.UpdateMonitoring(ctx, kubeClient, operations.UpdateMonitoringConfig{
        Image: image,
    }, exec)
    
    // Display updates from channel
    for update := range exec.UpdateCh {
        switch update.Status {
        case executor.StatusInProgress:
            out.Progress(update.Step + "...")
        case executor.StatusComplete:
            out.Success(update.Step)
        case executor.StatusFailed:
            out.Error(fmt.Sprintf("%s: %v", update.Step, update.Error))
            return update.Error
        }
    }
    
    return nil
}

func runUpdateMonitoringTUI(cmd *cobra.Command) error {
    ctx := cmd.Context()
    ctx = execctx.WithTUI(ctx, true)
    
    kubeClient, err := execctx.GetClient(ctx)
    if err != nil {
        return err
    }
    
    image, _ := cmd.Flags().GetString("image")
    if image == "" {
        image, err = collectImageInput()
        if err != nil {
            return err
        }
    }
    
    exec := executor.NewExecutor()
    
    // Note: We need to know the operation names upfront for TUI
    // This could be improved by having operations return their steps
    operations := []string{
        "Scale down cluster-monitoring-operator",
        "Update monitoring-plugin image",
    }
    
    model := tui.NewProgressModel("Updating Monitoring Plugin", operations)
    program := tea.NewProgram(model)
    
    // Run business logic in goroutine
    go operations.UpdateMonitoring(ctx, kubeClient, operations.UpdateMonitoringConfig{
        Image: image,
    }, exec)
    
    // Forward channel updates to TUI
    go func() {
        for update := range exec.UpdateCh {
            program.Send(tui.OperationUpdateMsg{
                Index:  update.Index,
                Status: convertStatus(update.Status),
                Error:  update.Error,
            })
        }
    }()
    
    finalModel, err := program.Run()
    if err != nil {
        return err
    }
    
    m := finalModel.(tui.ProgressModel)
    return m.Error()
}

func convertStatus(status executor.UpdateStatus) tui.OperationStatus {
    switch status {
    case executor.StatusPending:
        return tui.OperationPending
    case executor.StatusInProgress:
        return tui.OperationInProgress
    case executor.StatusComplete:
        return tui.OperationComplete
    case executor.StatusFailed:
        return tui.OperationFailed
    default:
        return tui.OperationPending
    }
}

func collectImageInput() (string, error) {
    var image string
    
    form := huh.NewForm(
        huh.NewGroup(
            huh.NewInput().
                Title("Enter monitoring plugin image").
                Placeholder("quay.io/observability-ui/monitoring-plugin:latest").
                Value(&image).
                Validate(func(s string) error {
                    if s == "" {
                        return fmt.Errorf("image cannot be empty")
                    }
                    return nil
                }),
        ),
    )
    
    err := form.Run()
    if err != nil {
        return "", err
    }
    
    return image, nil
}
```

---

## Summary

### The Pattern

1. **Business logic** in `pkg/operations/` sends updates via channel
2. **CLI runner** receives updates, displays with `output.Handler`
3. **TUI runner** receives updates, forwards to Bubble Tea program
4. **Zero duplication** of business logic

### Key Files

```
pkg/executor/executor.go          - Channel infrastructure
pkg/operations/monitoring.go      - Business logic (DRY)
cmd/update/monitoring.go          - Command (uses operations)
cmd/cleanup/monitoring.go         - Command (uses operations)
```

### Migration Checklist

- [ ] Create `pkg/executor/` package
- [ ] Create `pkg/operations/` package
- [ ] Migrate `cleanup monitoring` business logic
- [ ] Migrate `update monitoring` business logic
- [ ] Test both commands in CLI and TUI modes
- [ ] Document pattern as template
- [ ] Apply to future commands

### Expected Results

- ✅ Business logic written once per command
- ✅ Easy to test (no display mocking)
- ✅ Consistent behavior across modes
- ✅ Scalable to many commands
- ✅ Clean separation of concerns

---

**Next Steps**: Review this proposal, then implement Step 1 (Create Infrastructure) to validate the approach.
