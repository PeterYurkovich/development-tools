# TUI Input Collection - Implementation

**Date**: 2026-06-12  
**Status**: ✅ Complete

## Overview

Implemented interactive input collection for TUI mode. When a command is run in a terminal with missing flags, it now shows an interactive form to collect those values instead of erroring.

## What Was Implemented

### 1. Text Input Component (`pkg/tui/input.go`)

**Single-field text input** with cursor support and validation.

**Features**:
- Text entry with cursor navigation
- Left/Right arrow keys to move cursor
- Home/End to jump to start/end
- Backspace to delete
- Placeholder text when empty
- Enter to submit
- Esc to cancel

**Functions**:
- `NewTextInputModel(prompt, placeholder)` - Create input model
- `IsCancelled()` - Check if user cancelled
- `Value()` - Get entered value

**Visual Example**:
```
Enter monitoring plugin image

  quay.io/observability-ui/monitoring-plugin:latest

enter: submit • esc: cancel
```

### 2. Form Component (`pkg/tui/form.go`)

**Multi-field form** for collecting multiple inputs.

**Features**:
- Multiple fields in sequence
- Tab/Enter to move to next field
- Shift+Tab to move to previous field
- Same text editing as single input
- Submit when last field confirmed
- Returns map of field names to values

**Functions**:
- `NewFormModel(title, fields)` - Create form
- `IsCancelled()` - Check if user cancelled
- `Values()` - Get map of all field values

**Visual Example**:
```
Deploy Logging Stack

> Namespace: logging
  Stack Size: small
  Data Model: otel

tab/enter: next field • shift+tab: previous • esc: cancel
```

### 3. Integration in Update Monitoring

**File**: `cmd/update/monitoring.go`

Added `collectImageInput()` function that:
1. Shows text input TUI for image
2. Returns entered value or error if cancelled
3. Proceeds with operation using collected value

**Flow**:
```
./obstool update monitoring (no --image flag)
         ↓
In terminal? YES + Flags missing? YES
         ↓
TUI mode activated
         ↓
collectImageInput() shows input form
         ↓
User enters: quay.io/observability-ui/monitoring-plugin:latest
         ↓
Proceeds with update using entered image
         ↓
Shows progress TUI with operations
```

## Code Structure

### Text Input Model
```go
type TextInputModel struct {
    BaseModel
    prompt      string      // Question to ask
    value       string      // User input
    cursor      int         // Cursor position
    placeholder string      // Hint text
    done        bool        // Submission complete
    cancelled   bool        // User cancelled
}
```

### Form Model
```go
type FormModel struct {
    BaseModel
    title     string      // Form title
    fields    []Field     // List of fields
    current   int         // Currently active field
    cursorPos int         // Cursor in current field
    done      bool        // Form submitted
    cancelled bool        // User cancelled
}

type Field struct {
    Name        string  // Field identifier
    Prompt      string  // Field label
    Placeholder string  // Hint text
    Value       string  // User input
}
```

## Usage Pattern

### Simple Input (Single Value)
```go
func collectImageInput() (string, error) {
    model := tui.NewTextInputModel(
        "Enter monitoring plugin image",
        "e.g., quay.io/observability-ui/monitoring-plugin:latest",
    )
    
    program := tea.NewProgram(model)
    finalModel, err := program.Run()
    if err != nil {
        return "", err
    }
    
    m := finalModel.(tui.TextInputModel)
    if m.IsCancelled() {
        return "", fmt.Errorf("input cancelled")
    }
    
    return m.Value(), nil
}
```

### Multi-Field Form
```go
func collectDeployInputs() (map[string]string, error) {
    fields := []tui.Field{
        {Name: "namespace", Prompt: "Namespace", Placeholder: "logging"},
        {Name: "size", Prompt: "Stack Size", Placeholder: "small, medium, large"},
        {Name: "data-model", Prompt: "Data Model", Placeholder: "otel or viaq"},
    }
    
    model := tui.NewFormModel("Deploy Logging Stack", fields)
    program := tea.NewProgram(model)
    
    finalModel, err := program.Run()
    if err != nil {
        return nil, err
    }
    
    m := finalModel.(tui.FormModel)
    if m.IsCancelled() {
        return nil, fmt.Errorf("input cancelled")
    }
    
    return m.Values(), nil
}
```

## Complete Flow Example

### Terminal + No Flags

```bash
$ ./obstool update monitoring
```

**Step 1**: Mode detection
```go
useTUI, err := mode.DetermineMode(cmd, []string{"image"})
// Returns: (true, nil) - terminal + missing flags
```

**Step 2**: TUI mode activated
```go
return runUpdateMonitoringTUI(cmd)
```

**Step 3**: Collect missing input
```
Enter monitoring plugin image

  [cursor here]

enter: submit • esc: cancel
```

**Step 4**: User types
```
Enter monitoring plugin image

  quay.io/observability-ui/monitoring-plugin:latest

enter: submit • esc: cancel
```

**Step 5**: User presses Enter

**Step 6**: Progress TUI
```
Updating Monitoring Plugin

✓ Scale down cluster-monitoring-operator
⋯ Update monitoring-plugin image to quay.io/...

ctrl+c/q: cancel
```

**Step 7**: Complete
```
Updating Monitoring Plugin

✓ Scale down cluster-monitoring-operator
✓ Update monitoring-plugin image to quay.io/...
```

### Terminal + All Flags

```bash
$ ./obstool update monitoring --image=quay.io/test/image:v1
```

**Flow**:
1. Mode detection → CLI mode (all flags present)
2. Skip input collection (already have image)
3. Use output.Handler for structured text
4. No TUI components shown

## Keyboard Shortcuts

### Text Input
| Key | Action |
|-----|--------|
| Type | Enter text at cursor |
| ← → | Move cursor left/right |
| Home | Jump to start |
| End | Jump to end |
| Backspace | Delete character before cursor |
| Enter | Submit value |
| Esc | Cancel input |

### Form
| Key | Action |
|-----|--------|
| Type | Enter text in current field |
| Tab / Enter / ↓ | Next field |
| Shift+Tab / ↑ | Previous field |
| ← → | Move cursor in field |
| Home / End | Jump to start/end of field |
| Backspace | Delete character |
| Enter (last field) | Submit form |
| Esc | Cancel form |

## Files Created

1. **`pkg/tui/input.go`** (110 lines)
   - TextInputModel
   - Cursor-based text entry
   - Single value collection

2. **`pkg/tui/form.go`** (131 lines)
   - FormModel
   - Multi-field input
   - Field navigation

3. **`pkg/tui/example_input.md`**
   - Usage examples
   - Integration patterns
   - Keyboard shortcuts

## Files Modified

1. **`cmd/update/monitoring.go`**
   - Added `collectImageInput()`
   - Integrated into TUI flow
   - Removed "not yet implemented" error

## Testing

### Manual Test Scenarios

**Test 1**: Terminal + Missing Flag
```bash
$ ./obstool update monitoring
# → Shows text input TUI
# → Enter image URL
# → Shows progress TUI
# → Completes operation
```

**Test 2**: Terminal + Flag Present
```bash
$ ./obstool update monitoring --image=test
# → CLI mode (no input TUI)
# → Uses provided flag
# → Shows structured output
```

**Test 3**: Non-Terminal + Missing Flag
```bash
$ echo | ./obstool update monitoring
# → Error: not running in terminal and missing required flags: --image
```

**Test 4**: Cancel Input
```bash
$ ./obstool update monitoring
# → Shows input TUI
# → Press Esc
# → Error: input cancelled
```

## Code Statistics

**New Code**:
- `pkg/tui/input.go`: 110 lines
- `pkg/tui/form.go`: 131 lines
- Total: 241 lines

**Modified Code**:
- `cmd/update/monitoring.go`: +18 lines

**Total TUI Package**: ~520 lines across 7 files

## Benefits

### For Users
✅ **No documentation needed** - Form shows what's required  
✅ **Visual feedback** - Clear prompts and placeholders  
✅ **Error prevention** - Can't submit empty values  
✅ **Discoverability** - Learn available options interactively  

### For Automation
✅ **Flags still work** - Can skip TUI entirely  
✅ **Predictable** - Same flags always produce same result  
✅ **Scriptable** - No prompts in non-terminal  

### For Developers
✅ **Reusable components** - TextInput and Form work for any command  
✅ **Consistent UX** - Same input pattern everywhere  
✅ **Easy integration** - Just call collectInput() function  

## Next Steps

### Future Enhancements

1. **Input Validation**
   - Validate image format before submission
   - Show error message on invalid input
   - Allow retry without restarting

2. **Default Values**
   - Pre-populate with common values
   - Show most recently used
   - Suggest based on context

3. **Advanced Inputs**
   - Selection list (dropdown style)
   - Multi-select checkboxes
   - Yes/No confirmations

4. **Help Text**
   - Context-sensitive help
   - Show valid options
   - Examples on request

### Commands to Enhance

1. ✅ `update monitoring` - Input collection implemented
2. ⏸️ `deploy logging` - Use FormModel for 3+ fields
3. ⏸️ `deploy tracing` - Use FormModel
4. ⏸️ `deploy coo` - Selection + input combination
5. ⏸️ `users create` - Numeric input with validation

---

**Status**: ✅ Complete and working  
**Integration**: ✅ Integrated into update monitoring  
**Testing**: ✅ Manual testing complete  
**Ready for**: Reuse in other commands
