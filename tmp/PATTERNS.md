# obstool Code Patterns

Required patterns and code style for obstool development.

> **Quick Reference**: This is THE authoritative source for how to write code in obstool. See [README.md](./README.md) for overview, [ARCHITECTURE.md](./ARCHITECTURE.md) for system design.

---

## Code Style

### Comments

❌ **Avoid**:
- Redundant comments that just restate the code
- Package/function documentation unless exported and truly necessary
- Obvious comments like `// Create client` before `client := NewClient()`

✅ **Only add when**:
- Business logic is non-obvious
- Complex algorithms need explanation
- Important gotchas or constraints

### Variable Naming

❌ **Avoid 1-2 letter variables**:
- `c`, `e`, `i`, `j`, `k`, `x`, `y`, `s`, `r`

✅ **Acceptable short names** (standard Go idioms):
- `err` (errors)
- `ctx` (context.Context)
- `ok` (boolean checks)

✅ **Use descriptive names**:
- `client`, `config`, `namespace`, `deployment`
- `index`, `count`, `result`
- `kubeClient`, `storageProvider`, `secretName`

**Example**:
```go
// ❌ Avoid
for i, r := range resources {
    c.Create(ctx, r)
}

// ✅ Good
for index, resource := range resources {
    kubeClient.Create(ctx, resource)
}
```

---

## Required Patterns

### Multi-Step Functions: Executor Pattern

**Rule**: Any function performing multiple operations **MUST** support the executor pattern.

**Why**: Enables consistent progress tracking in both CLI and TUI modes.

**Implementation Requirements**:

1. **Accept executor parameter**: `*executor.Executor`
2. **Define step constants** at package level:
   ```go
   const (
       StepOne = iota
       StepTwo
       StepThree
   )
   ```
3. **Send progress updates** for each step:
   ```go
   exec.SendUpdate(StepOne, executor.StatusInProgress, "Step description")
   ```
4. **Send detailed logs** within steps:
   ```go
   exec.SendLog(StepOne, "Detailed progress message")
   ```
5. **Send errors with context**:
   ```go
   exec.SendUpdateWithError(StepOne, executor.StatusFailed, "Step description", err)
   ```
6. **Mark steps complete**:
   ```go
   exec.SendUpdate(StepOne, executor.StatusComplete, "Step description")
   ```

**Complete Example**:
```go
const (
    StepCreateNamespace = iota
    StepDeployResources
    StepWaitForReady
)

func DeployLogging(ctx context.Context, client client.Client, 
                   config LoggingConfig, exec *executor.Executor) error {
    defer exec.Close()
    
    // Step 1: Create namespace
    stepName := "Create namespace"
    exec.SendUpdate(StepCreateNamespace, executor.StatusInProgress, stepName)
    exec.SendLog(StepCreateNamespace, fmt.Sprintf("Creating namespace: %s", config.Namespace))
    
    err := createNamespace(ctx, client, config.Namespace)
    if err != nil {
        exec.SendUpdateWithError(StepCreateNamespace, executor.StatusFailed, stepName, err)
        return err
    }
    exec.SendUpdate(StepCreateNamespace, executor.StatusComplete, stepName)
    
    // Step 2: Deploy resources
    stepName = "Deploy resources"
    exec.SendUpdate(StepDeployResources, executor.StatusInProgress, stepName)
    exec.SendLog(StepDeployResources, "Creating LokiStack and dependencies")
    
    err = deployResources(ctx, client, config)
    if err != nil {
        exec.SendUpdateWithError(StepDeployResources, executor.StatusFailed, stepName, err)
        return err
    }
    exec.SendUpdate(StepDeployResources, executor.StatusComplete, stepName)
    
    // Step 3: Wait for ready
    stepName = "Wait for ready"
    exec.SendUpdate(StepWaitForReady, executor.StatusInProgress, stepName)
    exec.SendLog(StepWaitForReady, "Waiting for resources to become ready")
    
    err = waitForReady(ctx, client, config.Namespace)
    if err != nil {
        exec.SendUpdateWithError(StepWaitForReady, executor.StatusFailed, stepName, err)
        return err
    }
    exec.SendUpdate(StepWaitForReady, executor.StatusComplete, stepName)
    
    return nil
}
```

**Examples in Codebase**:
- `pkg/operations/monitoring.go` - Update/cleanup monitoring operations
- `pkg/storage/minio.go` - Storage provider deploy/delete

---

### Ensure Functions

**Rule**: Functions named `Ensure*` that create resources idempotently must return `(bool, error)`.

**Return Values**:
- `(true, nil)` - Resource was created
- `(false, nil)` - Resource already existed
- `(false, err)` - Error occurred

**Why**: Allows callers to know whether action was taken or resource already existed.

**Example**:
```go
func EnsureNamespace(ctx context.Context, client client.Client, name string) (bool, error) {
    namespace := &corev1.Namespace{
        ObjectMeta: metav1.ObjectMeta{Name: name},
    }
    
    err := client.Create(ctx, namespace)
    if err != nil {
        if errors.IsAlreadyExists(err) {
            return false, nil  // Already exists
        }
        return false, err  // Error
    }
    
    return true, nil  // Created
}

// Usage
created, err := EnsureNamespace(ctx, client, "minio")
if err != nil {
    return err
}
if created {
    log.Info("Created namespace minio")
} else {
    log.Info("Namespace minio already exists")
}
```

---

## Architecture Patterns

### Business Logic Decoupling

**Rule**: Business logic must be decoupled from display logic using executor channels.

**Structure**:
```
Business Logic (pkg/operations/)
    ↓ sends ProgressUpdate via channel
    ├─→ CLI Handler (pkg/output/cli.go) - displays with logger
    └─→ TUI Handler (command-specific) - forwards to Bubble Tea
```

**Benefits**:
- ✅ Business logic written once (no duplication)
- ✅ CLI and TUI guaranteed consistent
- ✅ Easy to test (no display mocking)
- ✅ Scalable to many commands

### Execution Context Pattern

**Rule**: Use context.Context with values for shared state (client, mode, version).

**Implementation**:
```go
// Set values in context
ctx = execctx.WithClient(ctx, kubeClient)
ctx = execctx.WithTUI(ctx, isTUI)

// Retrieve values from context
client, err := execctx.GetClient(ctx)
isTUI := execctx.IsTUI(ctx)
```

---

## File Organization

### Resource Definitions

**Location**: `pkg/resources/{resource_type}.go` (flat structure)

**Exception**: Dashboards go in `pkg/resources/dashboards/` (30+ files)

**Rules**:
- Use Go structs, not YAML templates
- Return typed objects: `*lokiv1.LokiStack`, not `map[string]interface{}`
- Handle version differences: Use `execCtx.Version` for conditional logic

**Example**:
```go
// pkg/resources/lokistack.go
package resources

func NewLokiStack(namespace, secretName, storageClass string) *lokiv1.LokiStack {
    return &lokiv1.LokiStack{
        ObjectMeta: metav1.ObjectMeta{
            Name:      "logging",
            Namespace: namespace,
        },
        Spec: lokiv1.LokiStackSpec{
            Size: lokiv1.SizeOneXExtraSmall,
            Storage: lokiv1.ObjectStorageSpec{
                Secret: lokiv1.ObjectStorageSecretSpec{
                    Name: secretName,
                    Type: lokiv1.ObjectStorageSecretS3,
                },
            },
            StorageClassName: lokiv1.StorageClassName(storageClass),
        },
    }
}
```

### Command Structure

**Location**: `cmd/{category}/{command}.go`

**Pattern**:
1. Define flags
2. Mode detection (CLI vs TUI)
3. Create execution context
4. Call operations function with executor
5. Handle errors mode-aware

**Example**:
```go
// cmd/deploy/logging.go
var loggingCmd = &cobra.Command{
    Use:   "logging",
    Short: "Deploy logging stack",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Get config from flags
        config := getLoggingConfig(cmd)
        
        // Determine mode
        isTUI := shouldUseTUI(cmd)
        
        if isTUI {
            return deployLoggingTUI(cmd.Context(), config)
        }
        return deployLoggingCLI(cmd.Context(), config)
    },
}
```

---

## Testing

**Philosophy**: Minimal testing

**What to test**:
- Critical business logic
- Complex algorithms
- Version-specific behavior

**What NOT to test**:
- Simple CRUD operations
- Display logic (decoupled via executor)
- Integration tests (no CI/CD, manual testing)

---

## Error Handling

### Error Wrapping

**Always wrap errors** with context using `fmt.Errorf` with `%w`:

```go
// ✅ Good
err := doSomething()
if err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}

// ❌ Avoid
if err != nil {
    return err  // No context
}
```

### Error Messages

**Use descriptive, actionable messages**:

```go
// ✅ Good
return fmt.Errorf("failed to create namespace %s: %w", namespace, err)

// ❌ Avoid
return fmt.Errorf("error: %w", err)
```

---

## Version-Specific Code

**Rule**: Check cluster version for version-dependent behavior.

**Console CRD Example**:
```go
// OCP 4.17-4.18 vs 4.19+
if version.IsOCP419OrNewer() {
    import osv1 "github.com/openshift/api/console/v1"
    // Use osv1
} else {
    import osRhobsv1 "github.com/rhobs/openshift-api/console/v1"
    // Use osRhobsv1
}
```

---

## Quick Checklist

Before submitting code, verify:

- [ ] Multi-step functions use executor pattern
- [ ] No 1-2 letter variable names (except `err`, `ctx`, `ok`)
- [ ] Minimal comments (only where truly needed)
- [ ] Ensure* functions return `(bool, error)`
- [ ] Errors wrapped with context (`%w`)
- [ ] Step constants defined for executor pattern
- [ ] Progress updates sent for each step
- [ ] Resources defined as Go structs (not YAML)
- [ ] Files in correct directory structure

---

## Examples to Reference

**Best Practice Examples**:
- `pkg/operations/monitoring.go` - Complete executor pattern
- `pkg/storage/minio.go` - Multi-step function with progress
- `cmd/update/monitoring.go` - Command structure with mode detection
- `pkg/executor/executor.go` - Executor interface definition

---

**Last Updated**: 2026-06-18  
**Questions?** Check [CONTEXT.md](./CONTEXT.md) or [go-migration-plan.md](./go-migration-plan.md)
