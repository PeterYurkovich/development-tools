# Agent Context: obstool Development

> **Quick Start for AI Agents**: This file provides essential context for understanding the obstool project. Read this first, then navigate to specific documents as needed.

---

## Project Overview

**Goal**: Migrate development-tools repository from multi-technology stack (bash, make, just, js, yaml) to unified Go CLI tool named `obstool`

**Current Phase**: Implementation (Foundation & Core Commands)

**Tool Name**: `obstool` (observability tooling)

**Target Users**: OpenShift Observability UI team members working with OCP clusters

---

## Current Status

### Completed
- ✅ Architecture and design decisions finalized
- ✅ Go module research completed (18 CRD types)
- ✅ Migration plan documented
- ✅ TODO breakdown created with atomic tasks
- ✅ Team feedback incorporated
- ✅ Foundation implementation (Go module, core packages)
- ✅ TUI framework (Bubble Tea components, mode detection)
- ✅ Business logic decoupling (channel-based architecture)
- ✅ First commands implemented (update/cleanup monitoring)
- ✅ OLM utilities (subscription, catalogsource, CSV waiting, operatorgroup, IDMS)
- ✅ Storage provider interface (MinIO abstraction for Loki/Tempo/ACM)

### In Progress
- 🔄 Deploy COO commands (bundle, fbc, stage, operatorhub)

### Not Started
- ⏸️ Deploy commands (logging, tracing, dashboards)
- ⏸️ Users commands (create, rbac)
- ⏸️ Additional cleanup commands (COO, logging, tracing, ACM, all)

---

## Key Architectural Decisions

**Framework & Libraries**
- CLI Framework: **Cobra** (industry standard for k8s tooling)
- TUI Library: **Bubble Tea** + **Huh** (for interactive mode and forms)
- Logging: **charmbracelet/log** (structured logging with color support)
- K8s Client: **controller-runtime directly** (no abstraction layer)
- Config: **Type-safe Go structs** (no Viper/config files)

**Execution Patterns**
- Mode Detection: **Flags + TUI hybrid** (CLI when all flags present, TUI when missing)
- State Management: **Execution Context pattern** (context.Context with values)
- Business Logic: **Channel-based decoupling** (business logic sends ProgressUpdate via channels)
- Resource Structure: **Flat pkg/resources/** (except dashboards/ subdirectory)
- Output Handling: **CLI general + TUI custom** (CLIHandler for all commands, command-specific TUI components)

**Quality Philosophy**
- Testing: **Minimal** (unit tests for critical code only, no CI/CD, no E2E)
- Timeline: **Not a constraint** (quality over speed)
- Migration: **Parallel run** (bash scripts continue until Go version ready)

**Code Style** (See [CODING_STANDARDS.md](./CODING_STANDARDS.md) for complete reference)
- Comments: **Minimal to none** (prefer self-documenting code)
- Variables: **No 1-2 letter names** (use descriptive names: `err`, `ctx`, `client` OK; `c`, `e`, `i` not OK)
- Multi-step functions: **MUST use executor pattern** (progress tracking for CLI/TUI modes)

**Constraints**
- No `obstool demo` command (workflows via command composition)
- `cleanup all` only available in flag mode (requires all flags)
- Authentication: kubeconfig file only (will NOT run in-cluster)
- Dashboards: 30+ files in `pkg/resources/dashboards/`, 1 dashboard per file

---

## Document Navigation

### For Implementation Tasks
→ **[TODO.md](./TODO.md)** - Atomic task breakdown with dependencies and status tracking

### For Coding Standards
→ **[CODING_STANDARDS.md](./CODING_STANDARDS.md)** - Required patterns, style guide, executor pattern, examples

### For Architecture & Design
→ **[go-migration-plan.md](./go-migration-plan.md)** - Complete architecture, code patterns, execution context examples

### For Kubernetes Resources
→ **[crd_go_modules_research.md](./crd_go_modules_research.md)** - Go modules for 18 CRD types, version compatibility

### For Quick Decisions Reference
→ **[README.md](./README.md)** - Executive summary, decisions overview, success criteria

### For Change History
→ **[UPDATES.md](./UPDATES.md)** - Summary of all document updates and team feedback

---

## Critical Technical Details

### Business Logic Decoupling Pattern

**Architecture**: Business logic is decoupled from display logic using Go channels.

**Structure**:
```
Business Logic (pkg/operations/)
    ↓ sends ProgressUpdate via channel
    ├─→ CLI Handler (pkg/output/cli.go) - displays with logger
    └─→ TUI Handler (command-specific) - forwards to Bubble Tea
```

**Key Components**:

1. **Business Logic** (`pkg/operations/{command}.go`):
```go
const (
    StepOne = iota  // Enumerate steps
    StepTwo
)

func ExecuteCommand(ctx context.Context, client client.Client, 
                   config CommandConfig, exec *executor.Executor) error {
    defer exec.Close()
    
    exec.SendUpdate(StepOne, executor.StatusInProgress, "Step one")
    exec.SendLog(StepOne, "Detailed progress message")
    err := doWork(...)
    if err != nil {
        exec.SendUpdateWithError(StepOne, executor.StatusFailed, "Step one", err)
        return err
    }
    exec.SendUpdate(StepOne, executor.StatusComplete, "Step one")
    
    return nil
}
```

2. **CLI Mode** (uses general handler):
```go
handler := output.NewCLIHandler()
exec := executor.NewExecutor()

go operations.ExecuteCommand(ctx, client, config, exec)

for update := range exec.UpdateCh {
    if err := handler.HandleUpdate(update); err != nil {
        return err
    }
}
```

3. **TUI Mode** (uses command-specific components):
```go
model := tui.NewProgressModel("Title", operations)
program := tea.NewProgram(model)
exec := executor.NewExecutor()

go operations.ExecuteCommand(ctx, client, config, exec)

go func() {
    for update := range exec.UpdateCh {
        if update.Message != "" {
            continue  // TUI ignores log messages
        }
        program.Send(tui.OperationUpdateMsg{...})
    }
}()

program.Run()
```

**Benefits**:
- ✅ Business logic written once (no duplication)
- ✅ CLI and TUI guaranteed consistent
- ✅ Easy to test (no display mocking)
- ✅ Scalable to many commands

**See**: `tmp/tasks/business-logic-decoupling/` for complete documentation

### Execution Context Pattern

Context values pattern for shared state:

```go
// Set values in context
ctx = execctx.WithClient(ctx, kubeClient)
ctx = execctx.WithTUI(ctx, isTUI)

// Retrieve values from context
client, err := execctx.GetClient(ctx)
isTUI := execctx.IsTUI(ctx)
```

### Version-Specific CRDs
**Important**: Some CRDs require different modules based on OCP version:

- **Console CRD**:
  - OCP 4.17-4.18: `github.com/rhobs/openshift-api/console/v1`
  - OCP 4.19+: `github.com/openshift/api/console/v1`

- **ClusterLogForwarder**:
  - Module: `github.com/openshift/cluster-logging-operator/apis/observability/v1`
  - NOT from observability-operator

See [crd_go_modules_research.md](./crd_go_modules_research.md) for full details.

### Directory Structure
```
obstool/
├── cmd/                      # Cobra commands
│   ├── root.go              # Root command + global flags
│   ├── version.go           # Version command
│   ├── deploy/              # Deploy commands
│   ├── cleanup/             # Cleanup commands (scale down, remove resources)
│   ├── update/              # Update commands (scale up, update versions)
│   └── users/               # User management
├── pkg/
│   ├── context/             # ExecutionContext (context.WithValue pattern)
│   ├── k8s/                 # Kubernetes client
│   ├── config/              # Type-safe config structs
│   ├── executor/            # Channel-based progress updates
│   ├── operations/          # Business logic (decoupled from display)
│   ├── resources/           # CRD definitions (flat)
│   │   └── dashboards/      # Dashboard definitions (30+ files)
│   ├── operators/           # OLM utilities, COO deployment
│   ├── storage/             # Storage provider interface
│   ├── users/               # User/RBAC utilities
│   ├── tui/                 # Bubble Tea TUI components
│   ├── output/              # CLI output handler
│   └── mode/                # Mode detection utilities
└── internal/
    ├── constants/           # Shared constants
    └── version/             # Version detection & comparison
```

### Kubernetes Client Configuration
```go
timeout := 30 * time.Second
qps := float32(50)
burst := 100
```

---

## Common Implementation Patterns

### Code Style Guidelines

**Comments**:
- Minimal to none - code should be self-documenting
- Only add comments for non-obvious business logic or complex algorithms
- No package/function documentation comments unless exported and truly necessary

**Variable Naming**:
- No 1-2 letter variable names (except standard Go idioms)
- ✅ Acceptable: `err`, `ctx`, `ok`
- ❌ Avoid: `c`, `e`, `i`, `j`, `k`, `x`, `y`
- Use descriptive names: `client`, `config`, `namespace`, `index`, `count`

**Multi-Step Functions**:
- Any function performing multiple operations MUST support the executor pattern
- Accept `*executor.Executor` as a parameter
- Define step constants (`const StepOne = iota`)
- Send progress updates via `exec.SendUpdate(step, status, description)`
- Send detailed logs via `exec.SendLog(step, message)`
- Send errors via `exec.SendUpdateWithError(step, status, description, err)`
- Examples: `pkg/operations/monitoring.go`, `pkg/storage/minio.go`

**Ensure Functions**:
- Functions named `Ensure*` that create resources idempotently must return `(bool, error)`
- Return `(true, nil)` when a new resource was created
- Return `(false, nil)` when an existing resource was found
- Return `(false, err)` on error
- Example: `func EnsureOperatorGroup(...) (bool, error)` returns whether it created a new group

### Adding a New Command

1. **Create command file**: `cmd/{category}/{command}.go`
2. **Define flags**: Use Cobra's flag system
3. **Mode detection**: Check if all required flags present
4. **Create ExecutionContext**: In command's Run function
5. **Call implementation**: Pass ExecutionContext as first param
6. **Handle errors**: Mode-aware error output

### Creating Resource Definitions

1. **File location**: `pkg/resources/{resource_type}.go` (flat, except dashboards)
2. **Use Go structs**: No YAML templates
3. **Version handling**: Use `execCtx.Version` for conditional logic
4. **Return typed objects**: e.g., `*lokiv1.LokiStack`, not `map[string]interface{}`

### Mode-Aware Operations

```go
// CLI Mode (all flags present)
if !execCtx.IsTUI {
    // Silent execution, minimal output
    return executeDirectly(execCtx, cfg)
}

// TUI Mode (missing flags)
// Show interactive selection, progress
return executeWithTUI(execCtx, cfg)
```

---

## Quick Command Reference

### Planned Commands Structure

```bash
obstool version                     # Show version info
obstool deploy <component>          # Deploy component (or TUI if no component)
obstool deploy coo --method=bundle  # Deploy COO via bundle
obstool deploy logging              # Deploy logging stack
obstool deploy tracing              # Deploy tracing stack
obstool cleanup <component>         # Cleanup component
obstool cleanup all --confirm=yes   # Cleanup all (flag mode only)
obstool update monitoring --image=X # Scale down CMO, update plugin image
obstool cleanup monitoring          # Scale up CMO (restores plugin)
obstool update coo --to-version=X   # Update COO
obstool users create --count=6      # Create test users
obstool users rbac --scenario=X     # Apply RBAC scenario
```

---

## Migration Strategy

1. **Parallel Development**: Bash scripts continue to work during Go development
2. **Rebase on Changes**: Update Go code when bash scripts change
3. **No Timeline Pressure**: Focus on quality and maintainability
4. **Iterative Approach**: Start with high-value commands (monitoring, users)
5. **Feature Parity**: Ensure 100% functional parity before deprecating bash

---

## Working with This Codebase

### Before Starting Implementation

1. Read this file (CONTEXT.md) for overview
2. Check [TODO.md](./TODO.md) for available tasks
3. Review execution context pattern in [go-migration-plan.md](./go-migration-plan.md)
4. Check CRD modules in [crd_go_modules_research.md](./crd_go_modules_research.md) if working with resources

### During Implementation

- **Update TODO.md**: Mark tasks as `[~]` in progress, `[x]` when complete
- **Follow patterns**: Use ExecutionContext, flat resources structure, mode-aware logic
- **Version awareness**: Check OCP version for version-specific CRDs
- **Code style**: Minimal comments, no 1-2 letter variable names (except standard Go conventions: `err`, `ctx`)
- **Plan then implement**: Create a task folder under `./tmp/tasks/{task-name}/` with:
  - `plan.md` - Detailed implementation plan (request approval before starting)
  - `implementation.md` - Summary of what was implemented (after completion)

### When Blocked

- Check if dependency task is complete in TODO.md
- Review decision rationale in go-migration-plan.md
- Verify CRD module compatibility in crd_go_modules_research.md

---

## Critical Gotchas

⚠️ **ClusterLogForwarder**: From `cluster-logging-operator` NOT `observability-operator`  
⚠️ **Console CRD**: Different modules for OCP 4.17-4.18 vs 4.19+  
⚠️ **Dashboards**: 30+ files, keep each in separate file in dashboards/ subdirectory  
⚠️ **Cleanup All**: Flag mode only, requires `--confirm=yes`  
⚠️ **No Demo Command**: Use command composition instead  
⚠️ **Flat Resources**: Everything at `pkg/resources/*.go` except dashboards  
⚠️ **Context First**: ExecutionContext is always first parameter  

---

## Questions or Unclear Requirements?

If you encounter ambiguity:

1. Check if decision is documented in [go-migration-plan.md](./go-migration-plan.md)
2. Look for similar pattern in existing decisions
3. Prefer simple, direct implementation over abstraction
4. When in doubt, use controller-runtime directly (can refactor later)
5. Ask user for clarification on genuinely unclear requirements

---

### For Task Planning & Implementation
→ **[tasks/](./tasks/)** - Task-specific folders containing plan.md and implementation.md
  - Example: `tasks/k8s-client-package/` contains plan and implementation docs

---

**Last Updated**: 2026-06-18  
**Document Version**: 1.3 (added business logic decoupling pattern, updated current status)
