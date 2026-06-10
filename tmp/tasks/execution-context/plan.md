# Execution Context Package Implementation Plan

**Task**: Create execution context package  
**Status**: Planning  
**Date**: 2026-06-10

---

## Overview

Implement the execution context package that provides a unified way to pass shared state through command execution. This avoids passing multiple parameters (`isTUI`, `client`, `ctx`) separately to every function.

---

## Task Breakdown

### TODO.md Reference

**Task** (lines 50-55):
```
### Execution Context
- [ ] **Create execution context package**
  - [ ] Create `pkg/context/context.go`
  - [ ] Implement `ExecutionContext` struct with Context, Client, IsTUI fields
  - [ ] Implement `NewExecutionContext()` constructor
  - Blocked by: Implement k8s client package
```

**Dependencies**: ✅ k8s client package (complete)

---

## Architecture Context

### From go-migration-plan.md (Updated)

**Original Pattern** (with Version field - now removed):
```go
type ExecutionContext struct {
    Context context.Context  // Go context for cancellation
    Client  client.Client    // Kubernetes client
    Version *VersionInfo     // Cluster version info (REMOVED)
    IsTUI   bool            // Running in TUI mode vs CLI mode
}
```

**Updated Pattern** (simplified):
```go
type ExecutionContext struct {
    Context context.Context  // Go context for cancellation
    Client  client.Client    // Kubernetes client
    IsTUI   bool            // Running in TUI mode vs CLI mode
}
```

### Purpose

**Problem**: Without execution context, every function needs multiple parameters:
```go
// Bad: repetitive parameters
func deployLogging(ctx context.Context, client client.Client, isTUI bool, cfg LoggingConfig) error
func createLokiStack(ctx context.Context, client client.Client, isTUI bool, config LokiConfig) error
func waitForReady(ctx context.Context, client client.Client, isTUI bool, name string) error
```

**Solution**: Single context parameter:
```go
// Good: single context parameter
func deployLogging(execCtx *ExecutionContext, cfg LoggingConfig) error
func createLokiStack(execCtx *ExecutionContext, config LokiConfig) error
func waitForReady(execCtx *ExecutionContext, name string) error
```

### Benefits

1. **Cleaner signatures**: Single parameter instead of 3+
2. **Consistent pattern**: All operations receive ExecutionContext
3. **Easy to extend**: Add new shared state without changing all signatures
4. **Mode-aware**: Functions can check `execCtx.IsTUI` for behavior
5. **Context propagation**: Go context for cancellation built-in

---

## Implementation Plan

### File: `pkg/context/context.go`

**Package name**: `context` (note: conflicts with stdlib `context`, use carefully)

**Alternative package name**: `execcontext` or `exec` to avoid confusion with stdlib

**Decision**: Use `context` as package name since:
- Import path will be `github.com/observability-ui/development-tools/pkg/context`
- Can import as `execctx` if needed: `import execctx "github.com/observability-ui/development-tools/pkg/context"`
- Follows the architectural design documents

### Structure

```go
package context

import (
    "context"
    
    "sigs.k8s.io/controller-runtime/pkg/client"
)

// ExecutionContext holds shared state for command execution
type ExecutionContext struct {
    Context context.Context
    Client  client.Client
    IsTUI   bool
}

// NewExecutionContext creates a new execution context
func NewExecutionContext(ctx context.Context, kubeClient client.Client, isTUI bool) *ExecutionContext {
    return &ExecutionContext{
        Context: ctx,
        Client:  kubeClient,
        IsTUI:   isTUI,
    }
}
```

### Implementation Details

#### 1. ExecutionContext Struct

**Fields**:
- `Context context.Context` - Standard Go context for cancellation, deadlines, values
- `Client client.Client` - Kubernetes client from controller-runtime
- `IsTUI bool` - Running in TUI mode (interactive) vs CLI mode (flags-only)

**Why these fields?**:
- `Context`: Required for all Kubernetes API calls, signal handling, timeouts
- `Client`: Every operation needs to talk to the cluster
- `IsTUI`: Mode-aware behavior (silent vs interactive, progress display, etc.)

**Why not more fields?**:
- Configuration: Passed as explicit parameters to keep functions testable
- Version: Not needed (version checks removed)
- Logger: Can be added later if needed
- Output handler: Can be created from `IsTUI` flag

#### 2. Constructor Function

**Signature**:
```go
func NewExecutionContext(ctx context.Context, kubeClient client.Client, isTUI bool) *ExecutionContext
```

**Parameters**:
- `ctx context.Context` - Base Go context (usually from command)
- `kubeClient client.Client` - Initialized Kubernetes client
- `isTUI bool` - Mode flag

**Returns**: Pointer to ExecutionContext (allows nil checks if needed)

**Why constructor?**:
- Ensures all fields are initialized
- Single point to create context
- Can add validation/defaults in future
- Follows Go best practices

---

## Usage Patterns

### 1. Creating ExecutionContext (in cmd/root.go)

```go
// PersistentPreRun for root command
func setupExecutionContext(cmd *cobra.Command) (*ExecutionContext, error) {
    ctx := cmd.Context()
    
    // Create Kubernetes client
    kubeconfigPath := cmd.Flag("kubeconfig").Value.String()
    kubeClient, err := k8s.NewClient(ctx, kubeconfigPath)
    if err != nil {
        return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
    }
    
    // Determine mode
    isTUI := !hasAllRequiredFlags(cmd)
    
    // Create execution context
    execCtx := context.NewExecutionContext(ctx, kubeClient.Client, isTUI)
    
    return execCtx, nil
}
```

### 2. Using ExecutionContext in Commands

```go
// cmd/deploy/logging.go
func deployLoggingCmd(cmd *cobra.Command, args []string) error {
    execCtx, err := getExecutionContext(cmd)
    if err != nil {
        return err
    }
    
    cfg := LoggingConfig{
        Namespace: cmd.Flag("namespace").Value.String(),
        DataModel: cmd.Flag("data-model").Value.String(),
    }
    
    return deployLogging(execCtx, cfg)
}

func deployLogging(execCtx *ExecutionContext, cfg LoggingConfig) error {
    if execCtx.IsTUI {
        return deployLoggingTUI(execCtx, cfg)
    }
    return deployLoggingCLI(execCtx, cfg)
}
```

### 3. Using ExecutionContext in Resource Operations

```go
// pkg/resources/lokistack.go
func CreateLokiStack(execCtx *ExecutionContext, config LokiStackConfig) error {
    lokiStack := NewLokiStack(config)
    
    if err := execCtx.Client.Create(execCtx.Context, lokiStack); err != nil {
        return fmt.Errorf("failed to create LokiStack: %w", err)
    }
    
    if execCtx.IsTUI {
        // Show success in TUI
    }
    
    return nil
}
```

### 4. Mode-Aware Operations

```go
func waitForReady(execCtx *ExecutionContext, name, namespace string) error {
    if execCtx.IsTUI {
        // Show progress in TUI
        return waitWithTUIProgress(execCtx, name, namespace)
    }
    
    // Silent polling in CLI mode
    return waitSilently(execCtx, name, namespace)
}
```

---

## Package Import Consideration

### Potential Conflict

The package name `context` conflicts with Go's standard library `context`.

**Options**:

1. **Keep package name as `context`**, import with alias when needed:
   ```go
   import (
       stdctx "context"
       execctx "github.com/observability-ui/development-tools/pkg/context"
   )
   ```

2. **Rename package to `execcontext`**:
   ```go
   package execcontext
   
   import "github.com/observability-ui/development-tools/pkg/execcontext"
   ```

3. **Rename package to `exec`**:
   ```go
   package exec
   
   import "github.com/observability-ui/development-tools/pkg/exec"
   ```

**Recommendation**: **Keep as `context`** because:
- Architectural docs use `pkg/context`
- When used, import path is fully qualified: `github.com/observability-ui/development-tools/pkg/context`
- Can alias if needed: `import execctx "github.com/.../pkg/context"`
- The `ExecutionContext` type name makes it clear what it is

**In practice**, most files will look like:
```go
package deploy

import (
    "context"  // stdlib context
    
    execctx "github.com/observability-ui/development-tools/pkg/context"
    "github.com/observability-ui/development-tools/pkg/k8s"
)

func deployLogging(execCtx *execctx.ExecutionContext, cfg LoggingConfig) error {
    ctx := execCtx.Context  // stdlib context.Context
    client := execCtx.Client
    // ...
}
```

---

## File Structure

### Complete File: `pkg/context/context.go`

```go
package context

import (
    "context"
    
    "sigs.k8s.io/controller-runtime/pkg/client"
)

type ExecutionContext struct {
    Context context.Context
    Client  client.Client
    IsTUI   bool
}

func NewExecutionContext(ctx context.Context, kubeClient client.Client, isTUI bool) *ExecutionContext {
    return &ExecutionContext{
        Context: ctx,
        Client:  kubeClient,
        IsTUI:   isTUI,
    }
}
```

**That's it!** Very simple, minimal implementation.

---

## Code Style Compliance

### Comments
- ✅ Type comment: Brief, describes purpose
- ✅ Constructor comment: Brief, describes what it does
- ❌ No field comments: Names are self-documenting

### Variable Naming
- ✅ `ctx` - Standard Go idiom for context.Context
- ✅ `kubeClient` - Descriptive, not just `c` or `client`
- ✅ `isTUI` - Boolean prefix convention
- ✅ `execCtx` - Standard abbreviation for ExecutionContext in usage

### Minimal Code
- ✅ Simple constructor, no validation (fail fast elsewhere)
- ✅ No getters/setters
- ✅ Exported fields (direct access is fine)
- ✅ No unnecessary abstractions

---

## Testing Strategy

### Unit Tests

Per project philosophy: **Minimal testing**

This package is **too simple to need tests**:
- No logic, just a struct and constructor
- No error cases
- No edge cases

**Decision**: Skip unit tests for this package.

### Integration Testing

Will be tested implicitly when:
- Root command creates ExecutionContext
- Commands use ExecutionContext
- Manual testing of actual operations

---

## Dependencies

### Go Modules Required

**Already in go.mod**:
- ✅ `context` - Go standard library (no import needed in go.mod)
- ✅ `sigs.k8s.io/controller-runtime` - For `client.Client`

**No new dependencies needed**.

---

## Integration Points

### 1. Root Command (cmd/root.go)

Will use ExecutionContext:
```go
// In PersistentPreRun
execCtx := context.NewExecutionContext(ctx, kubeClient, isTUI)
cmd.SetContext(context.WithValue(ctx, "execCtx", execCtx))
```

Or simpler pattern - store in command context:
```go
type execCtxKey struct{}

func setExecutionContext(cmd *cobra.Command, execCtx *ExecutionContext) {
    ctx := context.WithValue(cmd.Context(), execCtxKey{}, execCtx)
    cmd.SetContext(ctx)
}

func getExecutionContext(cmd *cobra.Command) (*ExecutionContext, error) {
    execCtx, ok := cmd.Context().Value(execCtxKey{}).(*ExecutionContext)
    if !ok {
        return nil, fmt.Errorf("execution context not found")
    }
    return execCtx, nil
}
```

### 2. All Deploy Commands

Will receive ExecutionContext as first parameter:
```go
func deployLogging(execCtx *ExecutionContext, cfg LoggingConfig) error
func deployTracing(execCtx *ExecutionContext, cfg TracingConfig) error
```

### 3. Resource Creation Functions

Will use ExecutionContext for client access:
```go
func CreateLokiStack(execCtx *ExecutionContext, cfg LokiStackConfig) error
```

### 4. Output Package (Future)

Will check `execCtx.IsTUI` for output mode:
```go
func NewOutputHandler(execCtx *ExecutionContext) *Handler
```

---

## Implementation Steps

### Step 1: Create directory
```bash
mkdir -p pkg/context
```

### Step 2: Create context.go
```bash
touch pkg/context/context.go
```

### Step 3: Implement the file
- Add package declaration
- Add imports
- Add ExecutionContext struct
- Add NewExecutionContext constructor

### Step 4: Verify compilation
```bash
go build ./pkg/context/...
```

### Step 5: Update TODO.md
Mark task as complete:
```
- [x] **Create execution context package**
  - [x] Create `pkg/context/context.go`
  - [x] Implement `ExecutionContext` struct with Context, Client, IsTUI fields
  - [x] Implement `NewExecutionContext()` constructor
```

---

## Success Criteria

✅ **Compilation**:
- `go build ./pkg/context/...` succeeds
- No compilation errors
- No import errors

✅ **Code Quality**:
- Minimal comments (only type/function level)
- Descriptive variable names
- Follows project patterns from k8s client package
- Simple, clean implementation

✅ **Integration Ready**:
- Can be imported by other packages
- Type exported and accessible
- Constructor function available
- Ready for root command implementation

✅ **Documentation**:
- Clear struct field names
- Constructor is obvious
- Pattern matches architectural docs

---

## Files Summary

### New Files Created
1. `pkg/context/context.go` (~25-30 lines total)

### Modified Files
1. `tmp/TODO.md` - Mark execution context task as complete

### No Changes Needed
- No dependencies to add to go.mod
- No scheme registration needed
- No tests to create (too simple)

---

## Timeline Estimate

**Total Effort**: ~15-20 minutes

- Step 1-2 (setup): 2 minutes
- Step 3 (implementation): 5 minutes
- Step 4 (compilation): 2 minutes
- Step 5 (documentation): 5 minutes
- Buffer: 5 minutes

---

## Future Extensions

This package can be extended later if needed:

**Possible additions**:
- `execCtx.Logger` - Structured logger
- `execCtx.Output` - Output handler
- `execCtx.DryRun` - Dry-run flag
- Methods for common operations

**Not needed now**:
- Keep it simple
- Add only when there's a clear need
- Don't over-engineer

---

## Comparison to Other Projects

### kubectl (kubernetes/kubectl)
- Uses command context heavily
- Stores config, client factory, etc.
- More complex (supports multiple contexts)

### helm
- Uses single global config object
- Different pattern (config-centric)

### obstool approach
- Hybrid: lightweight context + explicit configs
- Best of both: shared state + explicit parameters
- Simpler than kubectl, more structured than pure globals

---

## Notes

- **Very simple package** - don't overthink it
- **Critical foundation piece** - all commands will use it
- **Enables clean patterns** - mode-aware operations, consistent signatures
- **Easy to extend** - can add fields later without breaking existing code
- **No version field** - simplified after removing version detection

---

## Questions and Clarifications

### Q1: Should ExecutionContext be a pointer or value?
**Answer**: Pointer (`*ExecutionContext`) because:
- Consistent with Go conventions for "large" structs
- Allows nil checks if needed
- More efficient to pass around
- Can be extended later

### Q2: Should we validate fields in constructor?
**Answer**: No, keep it simple:
- Let caller handle validation
- Fail fast where it matters (in actual operations)
- Simpler code, easier to understand

### Q3: Should we embed context.Context instead of having it as a field?
**Answer**: No, keep as explicit field:
- More clear what's happening
- Easier to understand for newcomers
- Avoids method conflicts
- Consistent with architectural docs

### Q4: Package naming conflict with stdlib context?
**Answer**: Keep as `pkg/context`, alias on import if needed:
- Follows architectural design
- Import path makes it clear: `github.com/.../pkg/context`
- Can use alias: `import execctx "github.com/.../pkg/context"`

---

## References

### Documentation
- **TODO.md**: Lines 50-55 (execution context task)
- **CONTEXT.md**: Lines 90-109 (execution context pattern)
- **go-migration-plan.md**: Lines 186-225 (execution context pattern - needs update to remove Version)

### Code References
- **pkg/k8s/client.go**: Pattern for client usage
- Future reference: **cmd/root.go**: Will create and use ExecutionContext

### External
- Go context package: https://pkg.go.dev/context
- controller-runtime client: https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client

---

**Ready for Implementation**: ✅ Yes  
**Blockers**: None  
**Next Task After This**: Implement root command (`cmd/root.go`)
