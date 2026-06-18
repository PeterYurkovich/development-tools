# Tasks Directory

This directory contains task-specific planning and implementation documentation for the obstool development effort.

## Structure

Each task gets its own subdirectory with standardized documentation:

```
tasks/
├── {task-name}/
│   ├── plan.md            # Detailed implementation plan
│   └── implementation.md  # Post-implementation summary
└── README.md              # This file
```

## Workflow

### 1. Planning Phase
- Create a new folder: `tasks/{task-name}/`
- Create `plan.md` with:
  - Overview and goals
  - Architecture context
  - Detailed design
  - Implementation steps
  - Success criteria
  - Open questions/decisions
- Request approval before proceeding

### 2. Implementation Phase
- Execute the plan
- Follow code patterns from [PATTERNS.md](../PATTERNS.md):
  - **Multi-step functions MUST use executor pattern**
  - **Minimal to no comments** (code should be self-documenting)
  - **No 1-2 letter variable names** (except `err`, `ctx`, `ok`)
- Update [TASKS.md](../TASKS.md) to track progress
- Mark subtasks as complete

### 3. Documentation Phase
- Create `implementation.md` with:
  - Summary of what was implemented
  - Files created/modified
  - Dependencies added
  - Compilation/testing status
  - Next steps
  - References

## Naming Convention

Task folder names should:
- Be descriptive and concise
- Use kebab-case (lowercase with hyphens)
- Match or closely relate to the task name in [TASKS.md](../TASKS.md)

**Examples**:
- `k8s-client-package` - Implement k8s client package
- `execution-context` - Create execution context package
- `root-command` - Implement root command

## Task Status

See [TASKS.md](../TASKS.md) for current task status and breakdown.

**Completed** (examples):
- ✅ k8s-client-package, execution-context, config-package
- ✅ root-command, tui-framework
- ✅ update-cleanup-monitoring, business-logic-decoupling
- ✅ storage-provider-interface

**Full list**: See [TASKS.md](../TASKS.md)

---

**Note**: This tasks directory is for planning and documentation. Actual code goes in the main project structure (`cmd/`, `pkg/`, `internal/`).
