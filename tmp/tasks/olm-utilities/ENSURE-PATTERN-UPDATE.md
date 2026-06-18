# Update: Ensure Function Pattern

**Date**: 2026-06-18  
**Requested By**: User  
**Impact**: Coding standards + 1 function signature change

---

## Change Summary

Updated `EnsureOperatorGroup` to return whether a new resource was created, and established a coding standard for all `Ensure*` functions.

---

## Function Signature Change

### Before
```go
func EnsureOperatorGroup(ctx context.Context, kubeClient client.Client, config OperatorGroupConfig) error
```

### After
```go
func EnsureOperatorGroup(ctx context.Context, kubeClient client.Client, config OperatorGroupConfig) (bool, error)
```

**Return Values**:
- `(true, nil)` - New OperatorGroup was created
- `(false, nil)` - Existing OperatorGroup was found
- `(false, err)` - Error occurred

---

## Rationale

Idempotent "Ensure" functions should communicate what action they took:
1. **Transparency**: Caller knows if action was taken or resource already existed
2. **Logging**: Allows different log messages for create vs reuse
3. **Metrics**: Can track creation vs reuse rates
4. **Debugging**: Easier to understand what happened

---

## Usage Example

```go
exec.SendUpdate(StepEnsureOperatorGroup, executor.StatusInProgress, "Ensuring OperatorGroup")

ogConfig := operators.OperatorGroupConfig{
    Name:             "global-operators",
    Namespace:        operators.OpenShiftOperatorsNamespace,
    TargetNamespaces: []string{},
}

created, err := operators.EnsureOperatorGroup(ctx, client, ogConfig)
if err != nil {
    exec.SendUpdateWithError(StepEnsureOperatorGroup, executor.StatusFailed, "Ensuring OperatorGroup", err)
    return err
}

// Caller can now take different actions based on whether resource was created
if created {
    exec.SendLog(StepEnsureOperatorGroup, "Created new OperatorGroup")
} else {
    exec.SendLog(StepEnsureOperatorGroup, "Using existing OperatorGroup")
}

exec.SendUpdate(StepEnsureOperatorGroup, executor.StatusComplete, "Ensuring OperatorGroup")
```

---

## Coding Standard Added

Added to **both** `CONTEXT.md` and `go-migration-plan.md`:

```
Ensure Functions:
- Functions named Ensure* that create resources idempotently must return (bool, error)
- Return (true, nil) when a new resource was created
- Return (false, nil) when an existing resource was found
- Return (false, err) on error
- Allows callers to know whether action was taken or resource already existed
- Example: func EnsureOperatorGroup(...) (bool, error) returns whether it created a new group
```

**Location in CONTEXT.md**: Line 258-264 (after Variable Naming section)  
**Location in go-migration-plan.md**: Line 243-249 (after Variable Naming section)

---

## Files Modified

1. **`pkg/operators/operatorgroup.go`**
   - Updated `EnsureOperatorGroup` return type
   - Returns `true` when creating new group
   - Returns `false` when existing group found

2. **`tmp/CONTEXT.md`**
   - Added "Ensure Functions" coding standard

3. **`tmp/go-migration-plan.md`**
   - Added "Ensure Functions" coding standard

4. **`tmp/tasks/olm-utilities/implementation.md`**
   - Updated function documentation
   - Added usage example showing the pattern
   - Added to Coding Standards Added section
   - Added to Lessons Learned

---

## Future Applications

This pattern should be applied to any future functions that:
- Have names starting with `Ensure*`
- Create resources idempotently
- Check if resource exists before creating

**Examples of future functions**:
- `EnsureNamespace` - Create namespace if not exists
- `EnsureSecret` - Create secret if not exists
- `EnsureServiceAccount` - Create SA if not exists
- `EnsureIDMS` - Create IDMS if not exists

---

## Verification

✅ Code compiles: `go build ./pkg/operators/...`  
✅ Full build succeeds: `go build ./...`  
✅ Function signature updated in implementation  
✅ Coding standards updated in both docs  
✅ Usage examples provided

---

## Impact

**Low Impact Change**:
- Only affects one function currently
- No existing callers yet (COO deployment not implemented)
- Establishes pattern for future development
- Improves code clarity and maintainability

**Breaking Change**: No (function not yet used in any commands)

---

## Rationale for Standard

This pattern is common in Go standard library and Kubernetes ecosystem:

**Similar Patterns**:
- `sync.Map.LoadOrStore()` - returns `(value, loaded bool)`
- `os.Mkdir()` vs `os.MkdirAll()` - MkdirAll returns nil if exists
- Kubernetes admission webhooks - return whether mutation occurred

**Benefits**:
1. Caller transparency
2. Better logging granularity
3. Enables metrics collection
4. Clearer debugging
5. Self-documenting behavior

---

## Complete

All changes implemented and verified.
