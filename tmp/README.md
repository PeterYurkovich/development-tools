# obstool - OpenShift Observability Development Tool

**Single Go CLI replacing bash/make/just/yaml tooling for OCP observability development**

**Goal**: Migrate development-tools repository from multi-technology stack to unified Go CLI tool named `obstool`

**Target Users**: OpenShift Observability UI team members working with OCP clusters

---

## Current Status

### ✅ Completed
- Architecture and design decisions finalized
- Go module research completed (18 CRD types)
- Migration plan documented
- Foundation implementation (Go module, core packages, k8s client)
- TUI framework (Bubble Tea + Huh components, mode detection)
- Business logic decoupling (channel-based architecture)
- First commands implemented (update/cleanup monitoring)
- OLM utilities (subscription, catalogsource, CSV waiting, operatorgroup, IDMS)
- Storage provider interface (MinIO abstraction for Loki/Tempo/ACM)

### 🔄 In Progress
- Deploy COO commands (bundle, fbc, stage, operatorhub)

### ⏸️ Not Started
- Deploy commands (logging, tracing, dashboards, monitoring, ACM, korrel8r)
- Users commands (create, rbac)
- Additional cleanup commands (COO, logging, tracing, ACM, all)
- Build & release (Makefile, completions, releases)

---

## Quick Start

### For AI Agents
1. **Read this file** (10 min) - Project overview, status, navigation, critical gotchas
2. **Check [TASKS.md](./TASKS.md)** (2 min) - Find available work
3. **Read [PATTERNS.md](./PATTERNS.md)** (10 min) - Learn required code patterns
4. **Check [ARCHITECTURE.md](./ARCHITECTURE.md)** (optional) - Understand system design
5. **Start implementing** - Follow executor pattern for all multi-step functions

### For Developers
1. Read this file - Project overview and status
2. Review [ARCHITECTURE.md](./ARCHITECTURE.md) - Understand system design
3. Check [TASKS.md](./TASKS.md) - Find what needs to be built
4. Read [PATTERNS.md](./PATTERNS.md) - Code style and required patterns
5. Use [REFERENCE.md](./REFERENCE.md) - Look up CRDs, commands, troubleshooting

### For Stakeholders
1. Read "Current Status" section above
2. See [TASKS.md](./TASKS.md) for implementation progress
3. Review [ARCHITECTURE.md](./ARCHITECTURE.md) for technical decisions

---

## Navigation Guide

| File | Purpose | Read When |
|------|---------|-----------|
| **[TASKS.md](./TASKS.md)** | Task breakdown & tracking | Finding what to implement |
| **[ARCHITECTURE.md](./ARCHITECTURE.md)** | System design & decisions | Understanding structure, making design choices |
| **[PATTERNS.md](./PATTERNS.md)** | Code style & patterns | Before coding, during PR review |
| **[REFERENCE.md](./REFERENCE.md)** | CRDs, commands, troubleshooting | Looking up technical details |
| **[tasks/](./tasks/)** | Task-specific plans | Working on specific implementation |

---

## Critical Gotchas

⚠️ **Must-Know Before Implementing**:

1. **Multi-step functions MUST use executor pattern**
   - Any function with 2+ operations requires `*executor.Executor` parameter
   - Send progress updates at each step
   - See [PATTERNS.md](./PATTERNS.md#multi-step-functions-executor-pattern)

2. **ClusterLogForwarder from cluster-logging-operator**
   - NOT from observability-operator
   - Module: `github.com/openshift/cluster-logging-operator/apis/observability/v1`
   - See [REFERENCE.md](./REFERENCE.md#crd-modules)

3. **Console CRD has version-specific modules**
   - OCP 4.17-4.18: `github.com/rhobs/openshift-api/console/v1`
   - OCP 4.19+: `github.com/openshift/api/console/v1`
   - See [REFERENCE.md](./REFERENCE.md#version-specific-crds)

4. **Resources: flat structure except dashboards**
   - `pkg/resources/*.go` (flat)
   - `pkg/resources/dashboards/*.go` (subdirectory for 30+ files)
   - See [ARCHITECTURE.md](./ARCHITECTURE.md#directory-structure)

5. **No 1-2 letter variable names**
   - ✅ Acceptable: `err`, `ctx`, `ok`
   - ❌ Avoid: `c`, `e`, `i`, `j`, `k`, `x`, `y`
   - See [PATTERNS.md](./PATTERNS.md#variable-naming)

6. **Cleanup all: flag mode only**
   - Requires `--confirm=yes` flag
   - No TUI mode for cleanup all
   - See [ARCHITECTURE.md](./ARCHITECTURE.md#constraints)

7. **Authentication: kubeconfig only**
   - Will NOT run in-cluster
   - Uses kubeconfig file exclusively
   - See [ARCHITECTURE.md](./ARCHITECTURE.md#technology-stack)

8. **No obstool demo command**
   - Use command composition instead
   - Workflows via chaining commands
   - See [ARCHITECTURE.md](./ARCHITECTURE.md#constraints)

9. **Minimal comments**
   - Code should be self-documenting
   - Only add for non-obvious business logic
   - See [PATTERNS.md](./PATTERNS.md#comments)

10. **Ensure* functions return (bool, error)**
    - `(true, nil)` = created
    - `(false, nil)` = already existed
    - `(false, err)` = error occurred
    - See [PATTERNS.md](./PATTERNS.md#ensure-functions)

---

## Key Architectural Decisions

**Framework & Libraries**:
- CLI Framework: **Cobra** (industry standard for k8s tooling)
- TUI Library: **Bubble Tea** + **Huh** (for interactive mode and forms)
- Logging: **charmbracelet/log** (structured logging with color support)
- K8s Client: **controller-runtime directly** (no abstraction layer)
- Config: **Type-safe Go structs** (no Viper/config files)

**Execution Patterns**:
- Mode Detection: **Flags + TUI hybrid** (CLI when all flags present, TUI when missing)
- State Management: **Execution Context pattern** (context.Context with values)
- Business Logic: **Channel-based decoupling** (business logic sends ProgressUpdate via channels)
- Resource Structure: **Flat pkg/resources/** (except dashboards/ subdirectory)
- Output Handling: **CLI general + TUI custom** (CLIHandler for all commands)

**Quality Philosophy**:
- Testing: **Minimal** (unit tests for critical code only, no CI/CD, no E2E)
- Timeline: **Not a constraint** (quality over speed)
- Migration: **Parallel run** (bash scripts continue until Go version ready)

See [ARCHITECTURE.md](./ARCHITECTURE.md) for complete details and rationale.

---

## Quick Command Reference

```bash
# Version info
obstool version

# Update monitoring plugin
obstool update monitoring --image=quay.io/my-org/monitoring-plugin:latest

# Cleanup (restore CMO, removes plugin updates)
obstool cleanup monitoring

# Create test users
obstool users create --count=6 --password=password

# Deploy components
obstool deploy logging
obstool deploy tracing --size=1x.small
obstool deploy coo --method=bundle
obstool deploy dashboards

# Cleanup components
obstool cleanup logging
obstool cleanup all --confirm=yes  # Flag mode only
```

See [REFERENCE.md](./REFERENCE.md#command-reference) for complete command list with all flags.

---

## Technology Stack

| Component | Technology | Reason |
|-----------|-----------|--------|
| CLI Framework | Cobra | Industry standard for Kubernetes tooling |
| TUI Library | Bubble Tea + Huh | Best-in-class terminal UI with form support |
| Logging | charmbracelet/log | Structured logging with color support |
| K8s Client | controller-runtime | Direct access, no abstraction overhead |
| Config | Go structs | Type-safe, no config files needed |

---

## Directory Structure (Brief)

```
obstool/
├── cmd/              # Cobra commands (deploy, cleanup, update, users, version)
├── pkg/              # Reusable packages
│   ├── config/       # Type-safe config structs
│   ├── context/      # Execution context pattern
│   ├── executor/     # Channel-based progress updates
│   ├── k8s/          # Kubernetes client wrapper
│   ├── mode/         # Mode detection (CLI vs TUI)
│   ├── operations/   # Business logic (decoupled from display)
│   ├── operators/    # OLM utilities, COO deployment
│   ├── output/       # CLI output handler
│   ├── resources/    # CRD definitions (flat, except dashboards/)
│   ├── storage/      # Storage provider interface
│   ├── tui/          # Bubble Tea TUI components
│   └── users/        # User/RBAC utilities
└── internal/         # Private packages
    ├── constants/    # Shared constants
    └── version/      # Version detection & comparison
```

See [ARCHITECTURE.md](./ARCHITECTURE.md#directory-structure) for complete annotated structure.

---

## Getting Help

### Can't Find Something?

**Question** → **Where to Look**:
- What needs to be built? → [TASKS.md](./TASKS.md)
- How to code? → [PATTERNS.md](./PATTERNS.md)
- System design? → [ARCHITECTURE.md](./ARCHITECTURE.md)
- CRD modules? → [REFERENCE.md](./REFERENCE.md)
- Command syntax? → [REFERENCE.md](./REFERENCE.md)
- Blocked on task? → [ARCHITECTURE.md](./ARCHITECTURE.md) "When Blocked" section

### Still Stuck?

1. Check [PATTERNS.md](./PATTERNS.md) for coding patterns
2. Look at existing examples:
   - `pkg/operations/monitoring.go` - Executor pattern example
   - `pkg/storage/minio.go` - Multi-step function example
   - `cmd/update/monitoring.go` - Command structure
3. Review task-specific plan in `tasks/{task-name}/plan.md`

---

## Migration Strategy

**Approach**: Parallel development, no timeline pressure

1. **Bash scripts continue** to work during Go development
2. **Rebase Go code** when bash scripts change
3. **No timeline constraints** - focus on quality and maintainability
4. **Iterative implementation** - start with high-value commands
5. **100% feature parity** before deprecating bash

Current bash scripts: 58 scripts, 78 YAML files  
Target: Single Go binary (`obstool`)

---

## Working with This Codebase

### Before Implementation
1. Read this file (overview, status, gotchas)
2. Check [TASKS.md](./TASKS.md) for available tasks
3. Read [PATTERNS.md](./PATTERNS.md) for code standards
4. Review [ARCHITECTURE.md](./ARCHITECTURE.md) if making design decisions

### During Implementation
- Update [TASKS.md](./TASKS.md): Mark `[~]` in progress, `[x]` when complete
- Follow patterns from [PATTERNS.md](./PATTERNS.md)
- Use executor pattern for multi-step functions
- Check [REFERENCE.md](./REFERENCE.md) for CRD modules
- Create task folder: `tasks/{task-name}/plan.md` → get approval → implement → `implementation.md`

### When Blocked
- Check if dependency task is complete in [TASKS.md](./TASKS.md)
- Review decision rationale in [ARCHITECTURE.md](./ARCHITECTURE.md)
- Verify CRD module compatibility in [REFERENCE.md](./REFERENCE.md)
- Check existing examples in codebase

---

## Success Criteria

### Technical
- ✅ 100% functional parity with bash scripts
- ✅ All CRD types supported (18 types)
- ✅ Version detection working for OCP 4.11-4.19+
- ✅ Both CLI and TUI modes functional

### Quality
- ✅ Type-safe Go code
- ✅ Consistent progress tracking (executor pattern)
- ✅ Clean architecture (business logic decoupled)
- ✅ Maintainable (single source of truth)

### Adoption
- ✅ Team using new tool
- ✅ Bash scripts deprecated
- ✅ Faster development velocity

---

**Last Updated**: 2026-06-18  
**Version**: 2.0 (restructured documentation)  
**Status**: Active development - Foundation complete, commands in progress
