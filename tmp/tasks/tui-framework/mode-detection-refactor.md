# Mode Detection Refactor - Inverted Logic

**Date**: 2026-06-12  
**Change**: Inverted mode detection logic from "ShouldUseTUI" to "MustUseFlags"

## Problem

The original `ShouldUseTUI()` function checked if TUI should be used, which put the logic backwards:
- TUI mode was the "special case"
- CLI mode was the "default"
- Commands had to check "should I use TUI" instead of "should I use CLI"

This is backwards because:
1. **TUI should be the default** when running interactively in a terminal
2. **Flags are the override** to skip interactive mode
3. The question should be "must I use flags?" not "should I use TUI?"

## Solution

Renamed and refactored to `MustUseFlags()` with inverted logic:

### New Function: `MustUseFlags(cmd, requiredFlags) (bool, error)`

**Returns**:
- `(true, nil)` - Must use CLI/flags mode (flags provided OR not in terminal)
- `(false, nil)` - Should use TUI mode (in terminal, missing flags → default interactive)
- `(true, error)` - Cannot run (not in terminal AND missing required flags)

**Logic**:
```go
func MustUseFlags(cmd *cobra.Command, requiredFlags []string) (bool, error) {
    if !IsTerminal() {
        // Not in terminal - MUST use flags
        if !HasAllRequiredFlags(cmd, requiredFlags) {
            // Missing flags → error
            return true, fmt.Errorf("not running in terminal and missing required flags: %s", ...)
        }
        // All flags present → CLI mode
        return true, nil
    }
    
    // In terminal - check if user provided flags (CLI mode override)
    return HasAllRequiredFlags(cmd, requiredFlags), nil
}
```

## Usage Pattern

### Before (Wrong)
```go
if mode.ShouldUseTUI(cmd, requiredFlags) {
    return runCommandTUI(cmd)
}
return runCommandCLI(cmd)
```

This reads as: "Special case: use TUI. Default: use CLI."

### After (Correct)
```go
mustUseFlags, err := mode.MustUseFlags(cmd, requiredFlags)
if err != nil {
    return err  // Not in terminal and missing flags
}

if mustUseFlags {
    return runCommandCLI(cmd)  // Flags provided or non-interactive
}

return runCommandTUI(cmd)  // Default: interactive
```

This reads as: "Default: TUI. Override: CLI when flags provided."

## Behavior Matrix

| Environment | Flags Provided | Result | Mode |
|-------------|---------------|--------|------|
| Terminal | ✅ All | `(true, nil)` | CLI - user wants non-interactive |
| Terminal | ❌ Missing | `(false, nil)` | TUI - default interactive |
| Non-terminal (pipe) | ✅ All | `(true, nil)` | CLI - automation mode |
| Non-terminal (pipe) | ❌ Missing | `(true, error)` | Error - can't prompt |

## Examples

### Example 1: Interactive User (Default → TUI)
```bash
$ ./obstool update monitoring
# In terminal, no flags
# → TUI mode shows interactive progress
```

**Flow**:
1. `IsTerminal()` → `true`
2. `HasAllRequiredFlags()` → `false` (--image not provided)
3. `MustUseFlags()` → `(false, nil)`
4. ✅ Runs in TUI mode (default interactive behavior)

### Example 2: Power User with Flags (Override → CLI)
```bash
$ ./obstool update monitoring --image=quay.io/test/image:v1
# In terminal, all flags provided
# → CLI mode with structured output
```

**Flow**:
1. `IsTerminal()` → `true`
2. `HasAllRequiredFlags()` → `true` (--image provided)
3. `MustUseFlags()` → `(true, nil)`
4. ✅ Runs in CLI mode (user explicitly provided flags)

### Example 3: Automation/Script (CLI Required)
```bash
$ ./obstool update monitoring --image=quay.io/test/image:v1 | tee log.txt
# Not a terminal, all flags provided
# → CLI mode for automation
```

**Flow**:
1. `IsTerminal()` → `false` (piped)
2. `HasAllRequiredFlags()` → `true`
3. `MustUseFlags()` → `(true, nil)`
4. ✅ Runs in CLI mode (non-interactive environment)

### Example 4: Error - No Terminal, No Flags
```bash
$ echo | ./obstool update monitoring
# Not a terminal, missing flags
# → Error with clear message
```

**Flow**:
1. `IsTerminal()` → `false` (piped)
2. `HasAllRequiredFlags()` → `false` (--image not provided)
3. `MustUseFlags()` → `(true, error)`
4. ❌ Returns error: "not running in terminal and missing required flags: --image"

## Error Message

When not in terminal and flags are missing:
```
Error: not running in terminal and missing required flags: --image
Usage:
  obstool update monitoring [flags]

Flags:
  -h, --help           help for monitoring
      --image string   Monitoring plugin image to use
```

Clear, actionable, tells user exactly what's needed.

## Philosophy

### Design Principles

1. **Interactive by Default**
   - When a human runs the command, show TUI
   - More discoverable, user-friendly
   - Visual feedback and validation

2. **Automation-Friendly**
   - Flags available for all operations
   - Scripts/CI get predictable CLI output
   - No prompts that could hang

3. **Explicit Override**
   - Providing flags opts into CLI mode
   - Clear intent: "I know what I want, skip interaction"
   - Power users can be efficient

4. **Safe Defaults**
   - Non-interactive environments MUST provide flags
   - Can't accidentally prompt in CI/CD
   - Fail fast with clear error messages

## Command-Specific Configuration

Each command defines its own required flags:

```go
// Simple command - one flag
requiredFlags := []string{"image"}

// Complex command - multiple flags
requiredFlags := []string{"namespace", "data-model", "size"}

// Safety command - requires confirmation
requiredFlags := []string{"confirm"}
```

The mode detection logic is the same for all commands, but each command decides what flags are required for non-interactive use.

## Benefits

### For End Users
- ✅ Better UX - interactive by default
- ✅ Discoverable - can explore without docs
- ✅ Flexible - can use flags when needed

### For Automation
- ✅ Predictable - always same behavior with same flags
- ✅ Safe - errors in non-interactive instead of hanging
- ✅ Debuggable - clear error messages

### For Developers
- ✅ Consistent - same pattern for all commands
- ✅ Readable - "must use flags" is clearer than "should use TUI"
- ✅ Maintainable - single source of truth for mode logic

## Files Changed

1. **`pkg/mode/detect.go`**
   - Added `MustUseFlags()` function
   - Removed `ShouldUseTUI()` (deprecated)
   - Added clear error messages

2. **`cmd/update/monitoring.go`**
   - Changed from `ShouldUseTUI()` to `MustUseFlags()`
   - Inverted if/else logic
   - Added error handling

3. **`pkg/mode/README.md`**
   - Added comprehensive documentation
   - Examples for all scenarios
   - Design rationale

4. **`pkg/tui/example_usage.md`**
   - Updated examples to use `MustUseFlags()`
   - Clearer comments

## Migration Path

For future commands, use this pattern:

```go
func runCommand(cmd *cobra.Command, args []string) error {
    requiredFlags := []string{"flag1", "flag2"}
    
    mustUseFlags, err := mode.MustUseFlags(cmd, requiredFlags)
    if err != nil {
        return err
    }
    
    if mustUseFlags {
        return runCommandCLI(cmd)
    }
    
    return runCommandTUI(cmd)
}
```

Old `ShouldUseTUI()` function is still present for backwards compatibility but should not be used in new code.

---

**Status**: ✅ Complete  
**Impact**: Better UX, clearer code, safer automation  
**Breaking**: No (old function still exists)
