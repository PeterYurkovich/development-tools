# TUI Framework, Mode Detection, and Output Handling - Implementation

**Date**: 2026-06-12  
**Status**: ✅ Complete

## Overview

Implemented the complete TUI (Terminal User Interface) framework, mode detection utilities, and output handling infrastructure. These three components work together to provide a seamless experience for both interactive (TUI) and non-interactive (CLI) usage.

## What Was Implemented

### 1. TUI Package (`pkg/tui/`)

#### **`styles.go`** - Lipgloss Styling
Centralized styling definitions using lipgloss:
- `TitleStyle` - Bold magenta for titles
- `SelectedStyle` - Bold purple for selected items
- `CheckedStyle` - Green for checked items
- `ErrorStyle` - Bold red for errors
- `SuccessStyle` - Bold green for success messages
- `InfoStyle` - Cyan for info messages
- `ProgressStyle` - Purple for progress indicators
- `HelpStyle` - Gray for help text

#### **`models.go`** - Base Model Interface
- `Model` interface extending `tea.Model` with `Error()` and `Result()` methods
- `BaseModel` struct providing common functionality for all TUI models
- Helper methods: `SetError()`, `SetResult()`

#### **`deploy.go`** - Component Selection TUI
Interactive multi-select checkbox list:
- Navigate with arrow keys or `j`/`k`
- Toggle selection with `space`
- Toggle all with `a`
- Confirm with `enter`
- Cancel with `q` or `ctrl+c`
- Returns list of selected components

**Features**:
- Cursor indicator (`>`)
- Checkbox visual (`[ ]` or `[✓]`)
- Color-coded selection
- Keyboard shortcuts help text

#### **`progress.go`** - Operation Progress TUI
Real-time progress display for long-running operations:
- Shows list of operations with status icons
- Status types: Pending (○), In Progress (⋯), Complete (✓), Failed (✗)
- Color-coded status (gray, purple, green, red)
- Error display for failed operations
- Message-based updates via `OperationUpdateMsg`

**Features**:
- `SendOperationUpdate()` helper for sending progress updates
- Automatically quits when all operations complete
- Tracks errors for failed operations

### 2. Mode Detection Package (`pkg/mode/`)

#### **`detect.go`** - CLI vs TUI Detection
Utilities to determine whether to run in CLI or TUI mode:

**`IsTerminal()`**
- Checks if running in an interactive terminal
- Uses `golang.org/x/term` for cross-platform detection
- Returns `false` for pipes, redirects, non-TTY environments

**`HasAllRequiredFlags(cmd, requiredFlags)`**
- Checks if all required flags have been provided by the user
- Uses Cobra's `flag.Changed` to verify user actually set the flag
- Returns `false` if any required flag is missing or not changed

**`ShouldUseTUI(cmd, requiredFlags)`**
- Combines terminal detection and flag checking
- Returns `true` if terminal is interactive AND required flags are missing
- Returns `false` if not a terminal or all flags are present

**Usage Pattern**:
```go
requiredFlags := []string{"namespace", "data-model"}
if mode.ShouldUseTUI(cmd, requiredFlags) {
    // Launch TUI
    return runWithTUI(cmd)
}
// Run in CLI mode with flags
return runWithFlags(cmd)
```

### 3. Output Package (`pkg/output/`)

#### **`output.go`** - Mode-Aware Output Handler
Context-aware output that adapts to CLI vs TUI mode:

**`Handler` struct**
- Contains context with TUI mode flag
- Provides consistent interface for all output

**Methods**:
- `Info(message)` - Informational messages (cyan in TUI, plain in CLI)
- `Success(message)` - Success messages (green ✓ in TUI, plain in CLI)
- `Error(message)` - Error messages (red ✗ in TUI, "Error:" prefix in CLI)
- `Progress(message)` - Progress updates (purple ⋯ in TUI, plain in CLI)

**Behavior**:
- **TUI mode**: Colored output with Unicode symbols (✓, ✗, ⋯)
- **CLI mode**: Plain text, no colors, simple prefixes
- Errors always go to `stderr`, other output to `stdout`

### 4. Demo Command (`cmd/demo.go`)

Development/testing command to showcase TUI components:

**`obstool demo selection`**
- Demonstrates multi-select component selection
- Shows all observability components as choices
- Displays selected components after confirmation

**`obstool demo progress`**
- Demonstrates progress tracking
- Simulates deploying logging stack with 6 operations
- Shows real-time progress updates with delays

## File Structure

```
pkg/
├── mode/
│   └── detect.go          # Terminal and flag detection
├── output/
│   └── output.go          # Mode-aware output handler
└── tui/
    ├── deploy.go          # Component selection TUI
    ├── example_usage.md   # Usage examples and patterns
    ├── models.go          # Base model interface
    ├── progress.go        # Progress tracking TUI
    └── styles.go          # Lipgloss styles

cmd/
└── demo.go                # Demo commands for development/testing
```

## Dependencies Added

- `golang.org/x/term` v0.44.0 - Terminal detection (added)
- `golang.org/x/sys` v0.46.0 - System calls (upgraded from v0.44.0)

## Integration with Existing Code

### Execution Context
The TUI framework integrates with the existing execution context pattern:

```go
// Set TUI mode in context
ctx = execctx.WithTUI(ctx, true)

// Output handler uses context to determine mode
out := output.NewHandler(ctx)
out.Info("Running in TUI mode")  // Shows colored output with icon
```

### Command Pattern
Commands check mode and dispatch accordingly:

```go
func runCommand(cmd *cobra.Command, args []string) error {
    requiredFlags := []string{"flag1", "flag2"}
    
    if mode.ShouldUseTUI(cmd, requiredFlags) {
        ctx := execctx.WithTUI(cmd.Context(), true)
        return runCommandTUI(ctx, cmd)
    }
    
    ctx := execctx.WithTUI(cmd.Context(), false)
    return runCommandCLI(ctx, cmd)
}
```

## Usage Examples

### Deploy Selection

```go
choices := []string{"coo", "logging", "tracing"}
model := tui.NewDeploySelectionModel(choices)
program := tea.NewProgram(model)

finalModel, err := program.Run()
if err != nil {
    return err
}

selected := finalModel.(tui.DeploySelectionModel).Result().([]string)
// Process selected components
```

### Progress Tracking

```go
operations := []string{
    "Creating namespace",
    "Deploying resources",
    "Waiting for ready",
}

model := tui.NewProgressModel("Deploying", operations)
program := tea.NewProgram(model)

// In background goroutine:
program.Send(tui.OperationUpdateMsg{
    Index:  0,
    Status: tui.OperationInProgress,
})

// Later:
program.Send(tui.OperationUpdateMsg{
    Index:  0,
    Status: tui.OperationComplete,
})
```

### Output Handling

```go
out := output.NewHandler(ctx)

out.Info("Starting deployment...")
out.Progress("Creating resources...")
// Do work
out.Success("Resources created")

if err != nil {
    out.Error(fmt.Sprintf("Failed: %v", err))
    return err
}
```

## Testing

Manual testing with demo commands:

```bash
# Test selection TUI
./obstool demo selection

# Test progress TUI
./obstool demo progress

# Verify help text
./obstool demo --help
./obstool demo selection --help
./obstool demo progress --help
```

## Code Style

✅ **Minimal comments** - Code is self-documenting  
✅ **Descriptive variable names** - No 1-2 letter variables (except `s` for string building, `m` for model)  
✅ **Consistent patterns** - All TUI models follow same structure  
✅ **Error handling** - Errors propagated through Result/Error interface  

## TODO Integration

These tasks from `tmp/TODO.md` are now complete:

- [x] Create TUI package structure (models.go, styles.go)
- [x] Implement deploy selection TUI
- [x] Implement progress TUI
- [x] Implement mode detection utilities
- [x] Create output package

## Next Steps

Ready to implement commands that use these frameworks:

1. **Deploy commands** - Can now use TUI for interactive selection
2. **Update/Cleanup commands** - Can use progress tracking for multi-step operations
3. **Users commands** - Can use mode detection for flag vs TUI modes

The foundation is complete and ready for actual command implementation!

## Key Decisions

### Why Bubble Tea?
- Industry-standard TUI library for Go
- Elm-inspired architecture (model-view-update)
- Used by many popular CLI tools (gh, glow, etc.)
- Excellent keyboard handling and rendering

### Why Separate Mode Detection?
- Single responsibility - mode logic isolated from commands
- Reusable across all commands
- Easy to test and modify
- Clear decision boundary (CLI vs TUI)

### Why Output Handler?
- Consistent output formatting across all commands
- Mode-aware without duplicating logic in every command
- Easy to extend with new output types
- Centralized styling decisions

### Why Demo Command?
- Allows testing TUI components without cluster access
- Helps contributors understand the framework
- Visual verification of styles and behavior
- Can be removed before production release

## Files Created

1. `pkg/tui/styles.go` - 35 lines
2. `pkg/tui/models.go` - 28 lines
3. `pkg/tui/deploy.go` - 103 lines
4. `pkg/tui/progress.go` - 138 lines
5. `pkg/tui/example_usage.md` - Documentation
6. `pkg/mode/detect.go` - 33 lines
7. `pkg/output/output.go` - 51 lines
8. `cmd/demo.go` - 114 lines

**Total**: ~502 lines of code + documentation

---

**Implementation Status**: ✅ Complete and ready for use  
**Build Status**: ✅ Compiles without errors  
**Integration Status**: ✅ Works with existing execution context pattern
