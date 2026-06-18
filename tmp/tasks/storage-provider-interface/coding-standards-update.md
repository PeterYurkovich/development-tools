# Coding Standards Update - Executor Pattern Requirement

## Summary

Updated obstool coding standards to **require** the executor pattern for all multi-step functions, ensuring consistent progress tracking across CLI and TUI modes.

---

## Changes Made

### 1. Created New Document: `CODING_STANDARDS.md`

Comprehensive coding standards reference document covering:

**Code Style**:
- Comments (minimal to none)
- Variable naming (no 1-2 letter names)
- Error handling (wrap with context)

**Required Patterns**:
- ✅ **Multi-step functions MUST use executor pattern**
- ✅ Ensure* functions return `(bool, error)`
- ✅ Business logic decoupling via channels
- ✅ Execution context pattern

**Architecture**:
- File organization
- Resource definitions
- Command structure
- Version-specific code

**Examples**:
- Complete executor pattern implementation
- Good vs avoid examples
- Quick checklist for code review

### 2. Updated `CONTEXT.md`

**Added to Code Style section**:
```markdown
**Code Style** (See CODING_STANDARDS.md for complete reference)
- Comments: **Minimal to none** (prefer self-documenting code)
- Variables: **No 1-2 letter names** (...)
- Multi-step functions: **MUST use executor pattern** (progress tracking for CLI/TUI modes)
```

**Added to Document Navigation**:
```markdown
### For Coding Standards
→ **[CODING_STANDARDS.md](./CODING_STANDARDS.md)** - Required patterns, style guide, executor pattern, examples
```

**Added to Common Implementation Patterns**:
```markdown
**Multi-Step Functions**:
- Any function performing multiple operations MUST support the executor pattern
- Accept `*executor.Executor` as a parameter
- Define step constants (`const StepOne = iota`)
- Send progress updates via `exec.SendUpdate(step, status, description)`
- Send detailed logs via `exec.SendLog(step, message)`
- Send errors via `exec.SendUpdateWithError(step, status, description, err)`
- Examples: `pkg/operations/monitoring.go`, `pkg/storage/minio.go`
```

### 3. Updated `go-migration-plan.md`

**Added to Code Style Standards section**:
```markdown
**Multi-Step Functions (Executor Pattern)**:
- Any function performing multiple operations MUST support the executor pattern
- Accept `*executor.Executor` as a parameter
- Define step constants at package level: `const (StepOne = iota; StepTwo; ...)`
- Send progress updates: `exec.SendUpdate(step, executor.StatusInProgress, "Description")`
- Send detailed logs: `exec.SendLog(step, "Detailed message")`
- Send errors with context: `exec.SendUpdateWithError(step, executor.StatusFailed, "Description", err)`
- Always mark steps complete: `exec.SendUpdate(step, executor.StatusComplete, "Description")`
- Pattern enables both CLI and TUI modes with consistent progress tracking
- Examples: `pkg/operations/monitoring.go`, `pkg/storage/minio.go`
```

**Added complete executor pattern example**:
```go
const (
    StepCreateNamespace = iota
    StepCreateResources
    StepWaitForReady
)

func DeployComponent(ctx context.Context, client client.Client, 
                    config Config, exec *executor.Executor) error {
    stepName := "Create namespace"
    exec.SendUpdate(StepCreateNamespace, executor.StatusInProgress, stepName)
    exec.SendLog(StepCreateNamespace, "Ensuring namespace exists")
    
    err := createNamespace(ctx, client, config.Namespace)
    if err != nil {
        exec.SendUpdateWithError(StepCreateNamespace, executor.StatusFailed, stepName, err)
        return err
    }
    exec.SendUpdate(StepCreateNamespace, executor.StatusComplete, stepName)
    
    // ... more steps
    
    return nil
}
```

---

## What This Means for Future Development

### ✅ Required for All New Code

**Any function that performs multiple operations MUST**:
1. Accept `*executor.Executor` parameter
2. Define step constants at package level
3. Send progress updates at each step
4. Send detailed logs for debugging
5. Send errors with context
6. Mark steps complete

### Examples of Multi-Step Functions

**Requires Executor Pattern**:
- ✅ Deploy operations (logging, tracing, ACM)
- ✅ Storage provider operations
- ✅ Update operations (COO, monitoring)
- ✅ Cleanup operations (all components)
- ✅ User creation and RBAC setup
- ✅ Any operation with 2+ distinct steps

**Does NOT Require Executor** (single operations):
- ❌ Simple resource creation helpers
- ❌ Single k8s API calls
- ❌ Utility functions
- ❌ Version detection

### Code Review Checklist

Before approving PR, verify:
- [ ] Multi-step functions have executor parameter
- [ ] Step constants defined at package level
- [ ] Progress updates sent for each step
- [ ] Errors sent via `SendUpdateWithError()`
- [ ] Steps marked complete
- [ ] Log messages provide useful context
- [ ] No 1-2 letter variable names
- [ ] Minimal comments

---

## Benefits of This Standard

### For Users
✅ Consistent progress feedback in CLI mode  
✅ Real-time updates in TUI mode  
✅ Clear error messages with context  
✅ Visibility into what's happening  

### For Developers
✅ Single pattern to follow  
✅ Business logic separate from display  
✅ Easy to test (no UI mocking)  
✅ Scales to many commands  

### For Maintainers
✅ Consistent codebase  
✅ Easy to review PRs  
✅ Clear expectations  
✅ Examples to reference  

---

## Reference Examples in Codebase

**Complete Implementations**:
1. **`pkg/operations/monitoring.go`**
   - Update monitoring: 2 steps
   - Cleanup monitoring: 1 step
   - Shows error handling pattern

2. **`pkg/storage/minio.go`**
   - Deploy: 6 steps with wait loop
   - Delete: 4 steps
   - Shows periodic status updates during wait

**Command Integration**:
3. **`cmd/update/monitoring.go`**
   - Shows CLI vs TUI mode detection
   - Executor setup and handling
   - Progress consumption pattern

---

## Documentation Files Updated

1. ✅ **`tmp/CODING_STANDARDS.md`** (NEW)
   - Comprehensive reference
   - Examples and anti-patterns
   - Quick checklist

2. ✅ **`tmp/CONTEXT.md`**
   - Added executor pattern requirement
   - Added coding standards link
   - Updated navigation

3. ✅ **`tmp/go-migration-plan.md`**
   - Added executor pattern to code style
   - Added complete example
   - Emphasized requirement

---

## Next Steps for New Features

When implementing new commands:

1. **Read** `CODING_STANDARDS.md` first
2. **Reference** existing examples (`pkg/operations/monitoring.go`, `pkg/storage/minio.go`)
3. **Define** step constants for your operation
4. **Implement** with executor pattern from the start
5. **Test** in both CLI and TUI modes
6. **Verify** against checklist before PR

---

## Impact on Existing Code

**Already Compliant**:
- ✅ `pkg/operations/monitoring.go`
- ✅ `pkg/storage/minio.go`

**Future Work** (when implementing):
- Deploy commands (logging, tracing, ACM, dashboards)
- Update/cleanup commands (COO, logging, tracing, ACM)
- Users commands (create, rbac)
- OLM utilities

All new implementations will follow this standard from the start.

---

**Status**: ✅ Complete  
**Enforced**: Required for all new multi-step functions  
**Examples**: Available in codebase  
**Documentation**: Complete and comprehensive
