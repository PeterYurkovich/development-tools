# Output Handler Architecture Options

**Purpose**: Evaluate output handler design now that business logic is decoupled via channels  
**Context**: With channels separating business/display logic, should output handling be unified or split?  
**Date**: 2026-06-18

---

## Table of Contents

1. [Current State](#current-state)
2. [The Question](#the-question)
3. [Option 1: Unified Output Handler](#option-1-unified-output-handler)
4. [Option 2: Split CLI/TUI Handlers](#option-2-split-clitui-handlers)
5. [Option 3: CLI Handler + Command-Specific TUI](#option-3-cli-handler--command-specific-tui)
6. [Option 4: Channel-Based Display Adapters](#option-4-channel-based-display-adapters)
7. [Comparison Matrix](#comparison-matrix)
8. [Recommendation](#recommendation)

---

## Current State

### Existing Output Handler

**File**: `pkg/output/output.go`

```go
type Handler struct {
    ctx    context.Context
    logger *log.Logger
}

func NewHandler(ctx context.Context) *Handler {
    logger := log.NewWithOptions(os.Stderr, log.Options{
        ReportTimestamp: false,
        ReportCaller:    false,
    })
    
    if execctx.IsTUI(ctx) {
        logger.SetColorProfile(termenv.TrueColor)
    } else {
        logger.SetColorProfile(termenv.Ascii)
    }
    
    return &Handler{ctx: ctx, logger: logger}
}

func (h *Handler) Info(message string) { ... }
func (h *Handler) Success(message string) { ... }
func (h *Handler) Error(message string) { ... }
func (h *Handler) Progress(message string) { ... }
```

**Current Usage**:
- Both CLI and TUI modes use the same Handler
- Handler checks `execctx.IsTUI()` to decide formatting
- Used for simple text output

### With Channels (New Pattern)

```go
// Business logic sends updates
exec.SendUpdate(StepScaleDownCMO, StatusInProgress, "Scale down CMO")
exec.SendLog(StepScaleDownCMO, "Updating deployment replicas to 0")

// CLI mode - receives and displays
for update := range exec.UpdateCh {
    if update.Message != "" {
        out.Info(update.Message)  // Log message
        continue
    }
    
    switch update.Status {
        case StatusInProgress:
            out.Progress(update.Step + "...")
        case StatusComplete:
            out.Success(update.Step)
    }
}

// TUI mode - receives and forwards to Bubble Tea
for update := range exec.UpdateCh {
    if update.Message != "" {
        continue  // Ignore log messages
    }
    
    program.Send(tui.OperationUpdateMsg{...})
}
```

**Observation**: CLI and TUI handle updates very differently now.

---

## The Question

With business logic decoupled via channels:

1. **Should we keep a unified `output.Handler`** that tries to handle both CLI and TUI?
2. **Or split into separate handlers** optimized for each mode?
3. **Or use a general CLI handler** but command-specific TUI components?
4. **Or create display adapters** that convert channels to appropriate output?

---

## Option 1: Unified Output Handler

### Description

Keep current `output.Handler` that checks mode internally.

### Structure

```go
// pkg/output/output.go
type Handler struct {
    ctx    context.Context
    logger *log.Logger
}

func NewHandler(ctx context.Context) *Handler { ... }

func (h *Handler) Info(message string) {
    if execctx.IsTUI(h.ctx) {
        h.logger.Info(message)  // With colors
    } else {
        h.logger.Print(message)  // Plain
    }
}

func (h *Handler) HandleUpdate(update executor.ProgressUpdate) {
    if update.Message != "" {
        h.Info(update.Message)
        return
    }
    
    switch update.Status {
    case executor.StatusInProgress:
        h.Progress(update.Step)
    case executor.StatusComplete:
        h.Success(update.Step)
    case executor.StatusFailed:
        h.Error(fmt.Sprintf("%s: %v", update.Step, update.Error))
    }
}
```

### Usage in CLI

```go
func runUpdateMonitoringCLI(cmd *cobra.Command) error {
    // ...
    out := output.NewHandler(ctx)
    exec := executor.NewExecutor()
    
    go operations.UpdateMonitoring(ctx, kubeClient, config, exec)
    
    for update := range exec.UpdateCh {
        out.HandleUpdate(update)
    }
    
    return nil
}
```

### Pros

✅ **Simple**: One handler for everything  
✅ **Familiar**: Similar to current pattern  
✅ **Less code**: Single implementation  

### Cons

❌ **Mode mixing**: TUI and CLI logic in same code  
❌ **Limited flexibility**: Can't easily customize per-command  
❌ **TUI awkward**: TUI mode still needs separate Bubble Tea handling  
❌ **Not actually unified**: TUI commands bypass this handler anyway  

### Verdict

**Poor fit** - The unified handler doesn't actually unify anything since TUI commands need Bubble Tea programs regardless. The mode check becomes dead code in TUI paths.

---

## Option 2: Split CLI/TUI Handlers

### Description

Separate handlers for CLI and TUI modes.

### Structure

```go
// pkg/output/cli.go
type CLIHandler struct {
    logger *log.Logger
}

func NewCLIHandler() *CLIHandler {
    logger := log.NewWithOptions(os.Stderr, log.Options{
        ReportTimestamp: false,
        ReportCaller:    false,
    })
    logger.SetColorProfile(termenv.Ascii)
    
    return &CLIHandler{logger: logger}
}

func (h *CLIHandler) HandleUpdate(update executor.ProgressUpdate) error {
    if update.Message != "" {
        h.logger.Print(update.Message)
        return nil
    }
    
    switch update.Status {
    case executor.StatusInProgress:
        h.logger.Print(update.Step + "...")
    case executor.StatusComplete:
        h.logger.Info("✓ " + update.Step)
    case executor.StatusFailed:
        h.logger.Error(fmt.Sprintf("✗ %s: %v", update.Step, update.Error))
        return update.Error
    }
    
    return nil
}

// pkg/tui/handler.go
type TUIHandler struct {
    program *tea.Program
}

func NewTUIHandler(operations []string) (*TUIHandler, error) {
    model := NewProgressModel("Title", operations)
    program := tea.NewProgram(model)
    return &TUIHandler{program: program}, nil
}

func (h *TUIHandler) HandleUpdate(update executor.ProgressUpdate) {
    // Ignore log messages
    if update.Message != "" {
        return
    }
    
    h.program.Send(OperationUpdateMsg{
        Index:  update.Index,
        Status: convertStatus(update.Status),
        Error:  update.Error,
    })
}

func (h *TUIHandler) Run() error {
    finalModel, err := h.program.Run()
    if err != nil {
        return err
    }
    
    m := finalModel.(ProgressModel)
    return m.Error()
}
```

### Usage

```go
// CLI
func runUpdateMonitoringCLI(cmd *cobra.Command) error {
    handler := output.NewCLIHandler()
    exec := executor.NewExecutor()
    
    go operations.UpdateMonitoring(ctx, client, config, exec)
    
    for update := range exec.UpdateCh {
        if err := handler.HandleUpdate(update); err != nil {
            return err
        }
    }
    
    return nil
}

// TUI
func runUpdateMonitoringTUI(cmd *cobra.Command) error {
    operations := []string{
        "Scale down cluster-monitoring-operator",
        "Update monitoring-plugin image",
    }
    
    handler, err := tui.NewTUIHandler(operations)
    if err != nil {
        return err
    }
    
    exec := executor.NewExecutor()
    
    go operations.UpdateMonitoring(ctx, client, config, exec)
    
    go func() {
        for update := range exec.UpdateCh {
            handler.HandleUpdate(update)
        }
    }()
    
    return handler.Run()
}
```

### Pros

✅ **Clean separation**: CLI and TUI don't mix  
✅ **Optimized**: Each handler does one thing well  
✅ **Simpler code**: No mode checks  
✅ **Similar interface**: Both have `HandleUpdate()`  

### Cons

❌ **More files**: Two handler implementations  
❌ **TUI inflexible**: Generic TUI handler might not fit all commands  
❌ **Duplication risk**: Similar patterns in both handlers  

### Verdict

**Better** - Clean separation is good, but TUI handler might be too generic.

---

## Option 3: CLI Handler + Command-Specific TUI

### Description

General CLI handler for all commands, but TUI components customized per-command.

### Structure

```go
// pkg/output/cli.go - GENERAL CLI HANDLER
type CLIHandler struct {
    logger *log.Logger
}

func NewCLIHandler() *CLIHandler {
    // Same as Option 2
}

func (h *CLIHandler) HandleUpdate(update executor.ProgressUpdate) error {
    // Same as Option 2 - works for all commands
}

// pkg/tui/progress.go - GENERIC TUI COMPONENT
type ProgressModel struct { ... }
func NewProgressModel(title string, operations []string) *ProgressModel { ... }

// pkg/tui/deploy.go - COMMAND-SPECIFIC TUI
type DeploySelectionModel struct { ... }
func NewDeploySelectionModel(choices []string) *DeploySelectionModel { ... }

// pkg/tui/users.go - COMMAND-SPECIFIC TUI (potential)
type UserCreationModel struct { ... }
```

### Usage

```go
// CLI - Same handler for ALL commands
func runUpdateMonitoringCLI(cmd *cobra.Command) error {
    handler := output.NewCLIHandler()  // General handler
    exec := executor.NewExecutor()
    
    go operations.UpdateMonitoring(ctx, client, config, exec)
    
    for update := range exec.UpdateCh {
        if err := handler.HandleUpdate(update); err != nil {
            return err
        }
    }
    
    return nil
}

func runCleanupLoggingCLI(cmd *cobra.Command) error {
    handler := output.NewCLIHandler()  // SAME handler, different command
    exec := executor.NewExecutor()
    
    go operations.CleanupLogging(ctx, client, config, exec)
    
    for update := range exec.UpdateCh {
        if err := handler.HandleUpdate(update); err != nil {
            return err
        }
    }
    
    return nil
}

// TUI - Command-specific components
func runUpdateMonitoringTUI(cmd *cobra.Command) error {
    // Use generic progress model
    model := tui.NewProgressModel("Updating Monitoring", operations)
    program := tea.NewProgram(model)
    
    exec := executor.NewExecutor()
    go operations.UpdateMonitoring(ctx, client, config, exec)
    
    go func() {
        for update := range exec.UpdateCh {
            if update.Message != "" {
                continue
            }
            program.Send(tui.OperationUpdateMsg{...})
        }
    }()
    
    return runProgram(program)
}

func runDeployTUI(cmd *cobra.Command) error {
    // Use command-specific selection model
    model := tui.NewDeploySelectionModel([]string{
        "COO", "Logging", "Tracing",
    })
    program := tea.NewProgram(model)
    
    finalModel, _ := program.Run()
    selected := finalModel.Result().([]string)
    
    // Then use progress model for deployment...
}
```

### Pros

✅ **CLI reusable**: One handler for all commands  
✅ **TUI flexible**: Each command can customize UX  
✅ **Best of both**: General where possible, specific where needed  
✅ **Matches reality**: TUI is inherently more visual/custom  
✅ **Simple CLI path**: No per-command CLI code  

### Cons

❌ **TUI repetition**: Each TUI command has similar forwarding logic  
❌ **Two patterns**: CLI is simple, TUI is custom  

### Verdict

**Strong option** - Matches the natural differences between CLI (simple, consistent) and TUI (rich, customizable).

---

## Option 4: Channel-Based Display Adapters

### Description

Create adapter pattern that converts channels to appropriate display.

### Structure

```go
// pkg/display/adapter.go
type Adapter interface {
    HandleUpdates(updateCh <-chan executor.ProgressUpdate) error
}

// pkg/display/cli_adapter.go
type CLIAdapter struct {
    logger *log.Logger
}

func NewCLIAdapter() *CLIAdapter {
    // ...
}

func (a *CLIAdapter) HandleUpdates(updateCh <-chan executor.ProgressUpdate) error {
    for update := range updateCh {
        if update.Message != "" {
            a.logger.Print(update.Message)
            continue
        }
        
        switch update.Status {
        case executor.StatusInProgress:
            a.logger.Print(update.Step + "...")
        case executor.StatusComplete:
            a.logger.Info("✓ " + update.Step)
        case executor.StatusFailed:
            a.logger.Error(fmt.Sprintf("✗ %s: %v", update.Step, update.Error))
            return update.Error
        }
    }
    return nil
}

// pkg/display/tui_adapter.go
type TUIAdapter struct {
    program *tea.Program
}

func NewTUIAdapter(model tea.Model) *TUIAdapter {
    return &TUIAdapter{
        program: tea.NewProgram(model),
    }
}

func (a *TUIAdapter) HandleUpdates(updateCh <-chan executor.ProgressUpdate) error {
    go func() {
        for update := range updateCh {
            if update.Message != "" {
                continue
            }
            a.program.Send(tui.OperationUpdateMsg{...})
        }
    }()
    
    finalModel, err := a.program.Run()
    if err != nil {
        return err
    }
    
    return finalModel.(tui.Model).Error()
}
```

### Usage

```go
// CLI
func runUpdateMonitoringCLI(cmd *cobra.Command) error {
    adapter := display.NewCLIAdapter()
    exec := executor.NewExecutor()
    
    go operations.UpdateMonitoring(ctx, client, config, exec)
    
    return adapter.HandleUpdates(exec.UpdateCh)
}

// TUI
func runUpdateMonitoringTUI(cmd *cobra.Command) error {
    model := tui.NewProgressModel("Updating Monitoring", operations)
    adapter := display.NewTUIAdapter(model)
    exec := executor.NewExecutor()
    
    go operations.UpdateMonitoring(ctx, client, config, exec)
    
    return adapter.HandleUpdates(exec.UpdateCh)
}
```

### Pros

✅ **Unified interface**: Both implement `Adapter`  
✅ **Clean abstraction**: Channel → Display adapter  
✅ **Simple usage**: One method call to handle all updates  
✅ **Testable**: Can create mock adapter  

### Cons

❌ **Extra abstraction**: Another layer to understand  
❌ **TUI still custom**: Model creation is command-specific  
❌ **Overkill?**: May be more complex than needed  

### Verdict

**Nice pattern** - Clean abstraction, but might be overengineering for this use case.

---

## Comparison Matrix

| Criteria | Option 1:<br/>Unified | Option 2:<br/>Split | Option 3:<br/>CLI General + TUI Custom | Option 4:<br/>Adapters |
|----------|----------|---------|---------|----------|
| **Simplicity** | 🟢 Simple | 🟡 Moderate | 🟡 Moderate | 🔴 Complex |
| **CLI Reusability** | 🟢 One handler | 🟢 One handler | 🟢 One handler | 🟢 One handler |
| **TUI Flexibility** | 🔴 Limited | 🟡 Generic only | 🟢 Full | 🟢 Full |
| **Code Clarity** | 🔴 Mixed concerns | 🟢 Clear | 🟢 Clear | 🟢 Very clear |
| **Maintenance** | 🟡 Mode checks | 🟢 Separated | 🟢 Separated | 🟡 Extra layer |
| **Matches Reality** | 🔴 Forces unity | 🟡 Okay | 🟢 Natural fit | 🟡 Okay |
| **Per-Command Code** | 🟢 Minimal | 🟢 Minimal | 🟡 TUI custom | 🟢 Minimal |

**Legend**: 🟢 Good | 🟡 Okay | 🔴 Poor

---

## Recommendation

### **Recommended: Option 3 - CLI Handler + Command-Specific TUI**

#### Why?

1. **Matches Natural Differences**:
   - CLI is inherently consistent - same log format for all commands
   - TUI is inherently custom - each command may want different UX

2. **Pragmatic**:
   - Reuses CLI handler across all commands (DRY)
   - Allows TUI customization where it adds value
   - No forced abstraction where it doesn't fit

3. **Clear Separation**:
   - CLI path is simple and generic
   - TUI path is explicit about customization

4. **Already Partially Implemented**:
   - Current `pkg/tui/progress.go` - Generic progress component
   - Current `pkg/tui/deploy.go` - Command-specific selection component
   - This is the pattern we're already following!

#### Implementation

**CLI Handler** (pkg/output/cli.go):
```go
package output

import (
    "github.com/charmbracelet/log"
    "github.com/muesli/termenv"
    "github.com/observability-ui/development-tools/pkg/executor"
)

type CLIHandler struct {
    logger *log.Logger
}

func NewCLIHandler() *CLIHandler {
    logger := log.NewWithOptions(os.Stderr, log.Options{
        ReportTimestamp: false,
        ReportCaller:    false,
    })
    logger.SetColorProfile(termenv.Ascii)
    
    return &CLIHandler{logger: logger}
}

func (h *CLIHandler) HandleUpdate(update executor.ProgressUpdate) error {
    // Handle log messages
    if update.Message != "" {
        h.logger.Print(update.Message)
        return nil
    }
    
    // Handle status updates
    switch update.Status {
    case executor.StatusInProgress:
        h.logger.Print(update.Step + "...")
    case executor.StatusComplete:
        h.logger.Info("✓ " + update.Step)
    case executor.StatusFailed:
        h.logger.Error(fmt.Sprintf("✗ %s: %v", update.Step, update.Error))
        return update.Error
    }
    
    return nil
}
```

**TUI Components** (pkg/tui/):
- `progress.go` - Generic progress tracker (use for most commands)
- `deploy.go` - Multi-select component (command-specific)
- `models.go` - Base interfaces
- Future: Add command-specific components as needed

**Usage Pattern**:

```go
// ALL CLI commands use same pattern
func runAnyCLICommand(cmd *cobra.Command) error {
    handler := output.NewCLIHandler()
    exec := executor.NewExecutor()
    
    go operations.DoSomething(ctx, client, config, exec)
    
    for update := range exec.UpdateCh {
        if err := handler.HandleUpdate(update); err != nil {
            return err
        }
    }
    
    return nil
}

// TUI commands choose appropriate components
func runAnyTUICommand(cmd *cobra.Command) error {
    // Use generic progress for simple operations
    model := tui.NewProgressModel("Doing Something", operations)
    
    // OR use custom component for complex UX
    model := tui.NewCustomComponentModel(...)
    
    program := tea.NewProgram(model)
    exec := executor.NewExecutor()
    
    go operations.DoSomething(ctx, client, config, exec)
    
    go func() {
        for update := range exec.UpdateCh {
            if update.Message != "" {
                continue  // TUI ignores logs
            }
            program.Send(tui.OperationUpdateMsg{...})
        }
    }()
    
    finalModel, _ := program.Run()
    return finalModel.(tui.Model).Error()
}
```

#### Migration Path

1. **Create `pkg/output/cli.go`** with `CLIHandler`
2. **Update all CLI command functions** to use `CLIHandler.HandleUpdate()`
3. **Keep existing TUI components** (progress.go, deploy.go)
4. **Document pattern** in command template
5. **Remove old unified handler** once migration complete

---

## Summary

### The Answer

**Use Option 3: CLI Handler + Command-Specific TUI**

**CLI**:
- General `CLIHandler` in `pkg/output/cli.go`
- Used by all commands
- Simple `HandleUpdate()` method
- Consistent formatting

**TUI**:
- Generic components in `pkg/tui/` (progress, etc.)
- Command-specific components as needed
- Each command chooses appropriate TUI component
- Customizable UX per command

**Why This Works**:
- CLI is naturally uniform → One handler fits all
- TUI is naturally custom → Per-command flexibility
- Matches existing code structure
- Simple to implement
- Clear separation of concerns

**Next Steps**:
1. Create `pkg/output/cli.go` with `CLIHandler`
2. Refactor monitoring commands to use new pattern
3. Document in command implementation template
4. Apply to all future commands

---

**Conclusion**: Split the handlers, but keep CLI general and allow TUI to be command-specific. This matches the natural characteristics of each mode and the patterns we're already using.
