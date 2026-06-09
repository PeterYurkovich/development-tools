# Documentation Index

> **🤖 START HERE**: Quick guide to all documentation files for AI agents.

---

## File Purpose Overview

### 🎯 **CONTEXT.md** - START HERE
**Purpose**: Primary entry point for AI agents  
**Contains**: Project overview, current status, key decisions, navigation guide, critical patterns  
**When to read**: Always read this FIRST when starting work  
**Agent Action**: Load into context before any implementation work

---

### 📋 **TODO.md** - Task List
**Purpose**: Implementation task breakdown with status tracking  
**Contains**: Atomic tasks, subtasks, dependencies ("blocked by"), status markers  
**When to read**: When looking for what to work on next  
**Agent Action**: 
- Check for unblocked tasks `[ ]`
- Mark in-progress `[~]` when starting
- Mark complete `[x]` when done
- Update subtasks independently

**Format**:
```markdown
- [ ] **Parent task**
  - [x] Completed subtask
  - [~] In-progress subtask  
  - [ ] Todo subtask
  - Blocked by: Some other task
```

---

### 🏗️ **go-migration-plan.md** - Architecture Reference
**Purpose**: Detailed architecture, design patterns, code examples  
**Contains**: 
- Current state analysis
- Target architecture diagram
- 10 key decisions with rationale
- Execution Context pattern
- Mode detection patterns
- Directory structure
- Code examples for each pattern

**When to read**: 
- When implementing new commands/packages
- When unsure about architectural patterns
- When need code examples
- When making design decisions

**Agent Action**: Reference specific sections as needed (use Quick Navigation at top)

---

### 📦 **crd_go_modules_research.md** - CRD Module Reference
**Purpose**: Go module lookup for Kubernetes CRDs  
**Contains**:
- 18 CRD types with module paths
- Version-specific requirements
- Import examples
- Compatibility notes
- Quick lookup table at top

**When to read**:
- When working with Kubernetes resources
- When adding imports for CRD types
- When encountering version-specific issues
- When implementing resource creation functions

**Agent Action**: Use Quick CRD Lookup table, then read detailed section if needed

**Critical Info**:
- ⚠️ Console: Different modules for OCP 4.17-4.18 vs 4.19+
- ⚠️ ClusterLogForwarder: From cluster-logging-operator, NOT observability-operator

---

### 📄 **README.md** - Executive Summary
**Purpose**: High-level overview and decisions summary  
**Contains**:
- Executive summary
- Key findings
- Decisions made (summary)
- Document structure overview
- Success criteria

**When to read**: 
- When need high-level context
- When explaining project to others
- When checking decision summary

**Agent Action**: Reference for quick decision lookup after reading CONTEXT.md

---

### 📝 **UPDATES.md** - Change Log
**Purpose**: Track all changes based on team feedback  
**Contains**:
- Summary of changes
- Key decisions incorporated
- Files updated
- Validation checklist

**When to read**:
- When investigating why a decision was made
- When checking what changed from original plan
- When validating implementation matches decisions

**Agent Action**: Reference when uncertain about decision rationale

---

### 📑 **INDEX.md** - This File
**Purpose**: Navigation guide for all documentation  
**Contains**: File descriptions and when to use each

---

## Recommended Agent Workflows

### 🆕 Starting Fresh on Project
1. Read **CONTEXT.md** in full → load into context
2. Skim **TODO.md** to see available tasks
3. Reference **go-migration-plan.md** for patterns you'll need
4. Keep **crd_go_modules_research.md** handy for CRD work

### 🔨 Implementing a Specific Task
1. Check **TODO.md** → find task and mark `[~]`
2. Check "Blocked by" → ensure dependencies complete
3. Reference **CONTEXT.md** → review execution context pattern
4. Reference **go-migration-plan.md** → find relevant code examples
5. If using CRDs → check **crd_go_modules_research.md**
6. Complete task → mark `[x]` in **TODO.md**

### ❓ Answering "How do I...?"
- "How do I structure a command?" → **go-migration-plan.md** Section on CLI Framework
- "How do I handle mode detection?" → **CONTEXT.md** Common Implementation Patterns
- "What module for LokiStack?" → **crd_go_modules_research.md** Quick Lookup
- "What tasks are available?" → **TODO.md**
- "Why this decision?" → **README.md** or **UPDATES.md**

### 🐛 Debugging/Clarification
1. Check **CONTEXT.md** → Critical Gotchas section
2. Check **go-migration-plan.md** → relevant decision section
3. Check **crd_go_modules_research.md** → version compatibility
4. Ask user if still unclear

### 📊 Reporting Status
- Current progress → Count `[x]` vs `[ ]` in **TODO.md**
- What's done → Check **CONTEXT.md** Current Status
- What's next → Check **TODO.md** for unblocked `[ ]` tasks

---

## File Size Reference

| File | Lines | Purpose | Read Time |
|------|-------|---------|-----------|
| CONTEXT.md | ~300 | Quick start | 3-5 min |
| TODO.md | ~530 | Task list | Scan only |
| go-migration-plan.md | ~1440 | Deep reference | 20+ min (use Quick Nav) |
| crd_go_modules_research.md | ~980 | CRD lookup | 15+ min (use Quick Lookup) |
| README.md | ~410 | Summary | 5-10 min |
| UPDATES.md | ~410 | Change log | Reference only |

---

## Critical Information Summary

**Must Know Before Implementing**:
1. **Execution Context First**: Always pass `*ExecutionContext` as first parameter
2. **Flat Resources**: All resources at `pkg/resources/*.go` except dashboards
3. **No Abstraction**: Use controller-runtime directly
4. **Mode Aware**: Check `execCtx.IsTUI` for CLI vs TUI behavior
5. **Version Specific**: Console and other CRDs need version detection

**Common Mistakes to Avoid**:
- ❌ Using abstraction layer over controller-runtime
- ❌ Creating nested folders in pkg/resources/ (except dashboards/)
- ❌ Forgetting ExecutionContext parameter
- ❌ Using wrong module for ClusterLogForwarder
- ❌ Not checking OCP version for Console CRD
- ❌ Creating `obstool demo` command

---

## Document Update Protocol

When updating documentation:
1. Update the relevant detailed file (go-migration-plan.md, etc.)
2. Update CONTEXT.md if it affects critical info
3. Update TODO.md if it affects tasks
4. Add entry to UPDATES.md with date and change summary
5. Update README.md if it affects high-level decisions

---

**Last Updated**: 2026-06-09  
**Version**: 1.0
