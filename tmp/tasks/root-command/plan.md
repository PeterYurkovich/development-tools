# Plan: Implement Root Command

**Task ID**: Foundation → Root Command  
**Status**: Planning → Implementation  
**Date**: 2026-06-12  
**Dependencies**: ✅ Add core dependencies (complete), ✅ Create execution context package (complete)

---

## Overview

Implement the root Cobra command for obstool, including global flags, execution context creation, and the version subcommand. This is the entry point for the entire CLI tool.

---

## Goals

1. **Create root command** with Cobra framework
2. **Add global flags** (--kubeconfig)
3. **Implement PersistentPreRun** to create execution context and attach to command context
4. **Create version command** as first subcommand
5. **Set up main.go** to execute the root command

---

## Non-Goals

- ❌ No other subcommands yet (deploy, cleanup, etc.) - those come later
- ❌ No TUI mode detection yet - will add when implementing actual commands
- ❌ No comprehensive help text - keep it simple for now

---

## Architecture Context

### From TODO.md (Lines 65-76)

**Tasks**:
- Create `cmd/root.go`
- Set up Cobra root command with global flags
- Add `--kubeconfig` flag
- Implement execution context creation in PersistentPreRun
- Create `cmd/version.go` for version command

### From go-migration-plan.md

**Root command responsibilities**:
- Global flag parsing (--kubeconfig)
- Create execution context in PersistentPreRun
- Attach context to command for subcommands to access
- Provide help and version information

---

## Implementation Plan

### File 1: `cmd/root.go`

**Purpose**: Main root command with global flags and execution context setup

**Structure**:
```go
package cmd

import (
    "context"
    "fmt"
    "os"
    
    "github.com/spf13/cobra"
    execctx "github.com/observability-ui/development-tools/pkg/context"
    "github.com/observability-ui/development-tools/pkg/k8s"
)

var rootCmd = &cobra.Command{
    Use:   "obstool",
    Short: "Observability tooling for OpenShift",
    Long:  "A unified CLI tool for managing OpenShift observability components",
}

func Execute() error {
    return rootCmd.Execute()
}

func init() {
    // Global flags
    rootCmd.PersistentFlags().String("kubeconfig", "", "Path to kubeconfig file")
}
```

**Key Components**:

1. **Root Command Definition**:
   - Use: `obstool`
   - Short description
   - Long description

2. **Global Flags**:
   - `--kubeconfig` (string, optional)

3. **PersistentPreRunE**:
   - Create Kubernetes client
   - Create execution context
   - Store in command context for subcommands

4. **Execute() Function**:
   - Entry point called from main.go
   - Returns error for proper exit codes

### File 2: `cmd/version.go`

**Purpose**: Version subcommand

**Structure**:
```go
package cmd

import (
    "fmt"
    
    "github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
    Use:   "version",
    Short: "Print the version of obstool",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("obstool version 0.1.0")
    },
}

func init() {
    rootCmd.AddCommand(versionCmd)
}
```

**Version Info**:
- Start with simple hardcoded version: `0.1.0`
- Can be enhanced later with build-time variables

### File 3: `main.go` (update)

**Purpose**: Entry point that calls root command

**Current state**: May already exist from initial setup
**Update needed**: Call `cmd.Execute()`

```go
package main

import (
    "os"
    
    "github.com/observability-ui/development-tools/cmd"
)

func main() {
    if err := cmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

---

## Execution Context Pattern

### Creating Context in PersistentPreRunE

```go
func init() {
    rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
        // Skip for version command (doesn't need k8s client)
        if cmd.Name() == "version" {
            return nil
        }
        
        // Get kubeconfig flag
        kubeconfigPath, _ := cmd.Flags().GetString("kubeconfig")
        
        // Create Kubernetes client
        kubeClient, err := k8s.NewClient(cmd.Context(), kubeconfigPath)
        if err != nil {
            return fmt.Errorf("failed to create kubernetes client: %w", err)
        }
        
        // For now, assume CLI mode (isTUI will be determined per command)
        isTUI := false
        
        // Create execution context
        execCtx := execctx.NewExecutionContext(cmd.Context(), kubeClient.Client, isTUI)
        
        // Store in command context
        ctx := context.WithValue(cmd.Context(), execCtxKey{}, execCtx)
        cmd.SetContext(ctx)
        
        return nil
    }
}

type execCtxKey struct{}

func getExecutionContext(cmd *cobra.Command) (*execctx.ExecutionContext, error) {
    execCtx, ok := cmd.Context().Value(execCtxKey{}).(*execctx.ExecutionContext)
    if !ok {
        return nil, fmt.Errorf("execution context not found in command context")
    }
    return execCtx, nil
}
```

---

## Implementation Steps

### Step 1: Check existing cmd/ structure
```bash
ls -la cmd/
```

### Step 2: Create cmd/root.go

**Content**:
- Package declaration
- Import statements
- Root command definition
- Global flags
- PersistentPreRunE for execution context setup
- Execute() function
- Helper functions for context management

### Step 3: Create cmd/version.go

**Content**:
- Package declaration
- Version command definition
- Simple version output
- Register with root command in init()

### Step 4: Update/Create main.go

**Content**:
- Call cmd.Execute()
- Proper exit code handling

### Step 5: Test compilation
```bash
go build .
```

### Step 6: Test execution
```bash
./obstool --help
./obstool version
```

---

## Testing Strategy

**Manual testing**:
1. `obstool --help` - shows help
2. `obstool version` - shows version
3. `obstool --kubeconfig=/path/to/config version` - accepts flag but doesn't use it for version
4. `obstool --invalid-flag` - shows error

**No unit tests** per minimal testing philosophy.

---

## Success Criteria

✅ **Files Created**:
- `cmd/root.go` exists
- `cmd/version.go` exists
- `main.go` updated/created

✅ **Compilation**:
- `go build .` succeeds
- Binary `obstool` is created

✅ **Functionality**:
- `obstool --help` works
- `obstool version` displays version
- `--kubeconfig` flag is recognized

✅ **Code Quality**:
- Minimal comments
- Descriptive variable names
- Follows project patterns

✅ **Integration Ready**:
- Execution context setup works
- Ready for subcommands to be added

---

## Dependencies

**Already in go.mod**:
- ✅ `github.com/spf13/cobra` 
- ✅ `sigs.k8s.io/controller-runtime`

**No new dependencies needed**.

---

## Future Enhancements

**Not implementing now**:
- Build-time version information (ldflags)
- Shell completion generation
- Config file support
- TUI mode detection (per command)

**Will add when needed**:
- Subcommand groups (deploy/, cleanup/, etc.)
- Mode detection logic
- Output handlers

---

## Files Summary

### New Files
1. `cmd/root.go` (~80-100 lines)
2. `cmd/version.go` (~20 lines)
3. `main.go` (~15 lines) - if not exists, or update existing

### Modified Files
1. `tmp/TODO.md` - mark tasks complete

---

## Timeline Estimate

**Total Effort**: ~30 minutes

- Step 1 (check structure): 2 minutes
- Step 2 (root.go): 15 minutes
- Step 3 (version.go): 5 minutes
- Step 4 (main.go): 3 minutes
- Step 5-6 (testing): 5 minutes

---

## References

- **TODO.md**: Lines 65-76 (root command task)
- **CONTEXT.md**: Execution context pattern, directory structure
- **go-migration-plan.md**: Lines 277-294 (Cobra command examples)
- **pkg/context/context.go**: ExecutionContext definition
- **pkg/k8s/client.go**: Client creation

---

## Implementation Approved

Ready to proceed with implementation.

**Plan Status**: ✅ Complete - proceeding with implementation  
**Blockers**: None  
**Next Task After This**: Implement TUI framework or deploy commands
