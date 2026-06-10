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
- Follow code style guidelines:
  - **Minimal to no comments** (code should be self-documenting)
  - **No 1-2 letter variable names** (except `err`, `ctx`, `ok`)
- Update TODO.md to track progress
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
- Match or closely relate to the task name in TODO.md

**Examples**:
- `k8s-client-package` - Implement k8s client package
- `execution-context` - Create execution context package
- `root-command` - Implement root command

## Completed Tasks

- ✅ **k8s-client-package**: Kubernetes client implementation with kubeconfig loading and scheme registration

## In Progress

(None currently)

## Planned

- Execution context package
- Root command
- TUI framework
- And more per TODO.md

---

**Note**: This tasks directory is for planning and documentation. Actual code goes in the main project structure (`cmd/`, `pkg/`, `internal/`).
