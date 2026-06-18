# Mode Detection Logic - Final Clarification

**Date**: 2026-06-12  
**Status**: ✅ Complete

## The Two Modes

1. **TUI mode** - Interactive, visual, asks for missing inputs
2. **CLI mode** - Non-interactive, flags-based, plain output

## Mode Selection Logic

### Rule 1: Not in Terminal → CLI Mode Required
```
Not in Terminal (pipe, redirect, CI/CD)
└─→ All flags present? 
    ├─→ YES: CLI mode ✅
    └─→ NO:  ERROR ❌
```

**Examples**:
```bash
# ✅ Works - all flags provided
echo | ./obstool update monitoring --image=quay.io/test/image:v1

# ❌ Error - missing flags
echo | ./obstool update monitoring
# Error: not running in terminal and missing required flags: --image
```

### Rule 2: In Terminal + Missing Flags → TUI Mode
```
In Terminal + Missing Flags
└─→ TUI mode (ask for flags interactively) ✅
```

**Example**:
```bash
# ✅ Would show TUI to collect --image value
./obstool update monitoring
# (TUI input collection not yet implemented)
```

### Rule 3: In Terminal + All Flags → CLI Mode
```
In Terminal + All Flags Present
└─→ CLI mode (user chose to provide flags) ✅
```

**Example**:
```bash
# ✅ CLI mode with structured output
./obstool update monitoring --image=quay.io/test/image:v1
```

## Implementation

### Function: `DetermineMode(cmd, requiredFlags) (useTUI bool, err error)`

```go
func DetermineMode(cmd *cobra.Command, requiredFlags []string) (useTUI bool, err error) {
    inTerminal := IsTerminal()
    hasAllFlags := HasAllRequiredFlags(cmd, requiredFlags)
    
    // Rule 1: Not in terminal
    if !inTerminal {
        if !hasAllFlags {
            // Missing flags → ERROR
            return false, fmt.Errorf("not running in terminal and missing required flags: ...")
        }
        // All flags → CLI mode
        return false, nil
    }
    
    // Rule 2: In terminal, missing flags → TUI
    if !hasAllFlags {
        return true, nil
    }
    
    // Rule 3: In terminal, all flags → CLI
    return false, nil
}
```

### Usage Pattern

```go
func runCommand(cmd *cobra.Command, args []string) error {
    requiredFlags := []string{"image"}
    
    useTUI, err := mode.DetermineMode(cmd, requiredFlags)
    if err != nil {
        return err  // Not in terminal and missing flags
    }
    
    if useTUI {
        return runCommandTUI(cmd)   // Interactive mode
    }
    
    return runCommandCLI(cmd)       // Flags mode
}
```

## Decision Matrix

| Environment | Flags | Result | Mode |
|-------------|-------|--------|------|
| Terminal | All present | `(false, nil)` | CLI |
| Terminal | Missing | `(true, nil)` | TUI |
| Non-terminal | All present | `(false, nil)` | CLI |
| Non-terminal | Missing | `(false, error)` | ERROR |

## Why This Makes Sense

### For Users
- **Terminal + no flags** → Interactive (friendly)
- **Terminal + flags** → Fast CLI (power users)
- **Pipe/script + flags** → Automation works
- **Pipe/script + no flags** → Clear error (safe)

### For Automation
- Scripts must provide flags (explicit)
- No hanging on interactive prompts
- Reproducible (same flags = same result)
- Clear errors when misconfigured

### For Developers
- Single function determines mode
- Clear true/false return for TUI vs CLI
- Error case is explicit
- Easy to test and reason about

## Testing

```bash
# Test 1: Non-terminal + missing flags → ERROR
echo | ./obstool update monitoring
# ✅ Error: not running in terminal and missing required flags: --image

# Test 2: Non-terminal + all flags → CLI
echo | ./obstool update monitoring --image=test
# ✅ Would run in CLI mode (needs cluster)

# Test 3: Terminal + all flags → CLI
./obstool update monitoring --image=test
# ✅ Would run in CLI mode (needs cluster)

# Test 4: Terminal + missing flags → TUI
./obstool update monitoring
# ✅ Would run in TUI mode (TUI input not yet implemented)
```

## Future: TUI Input Collection

When a command runs in TUI mode with missing flags, it should:

1. Show interactive form/selection to collect missing values
2. Validate inputs
3. Proceed with operation using collected values

**Example** (future implementation):
```
Update Monitoring Plugin

Please provide the required information:

  Image: [                                              ]
         ▲ Enter monitoring plugin image URL

  [Continue]  [Cancel]

Use tab to navigate, enter to submit
```

## Code Changes

### New Functions
- `DetermineMode()` - Clear name, returns `useTUI bool`
- `GetMissingFlags()` - Helper to get list of missing flags

### Deprecated Functions  
- `MustUseFlags()` - Confusing name, replaced by `DetermineMode()`
- `ShouldUseTUI()` - Old version, kept for compatibility

### Updated Commands
- `cmd/update/monitoring.go` - Uses `DetermineMode()`
- Clear if/else: `if useTUI { ... } else { ... }`

## Summary

**Crystal clear logic**:
1. Not terminal → must have flags or error
2. Terminal + missing → TUI asks for them
3. Terminal + present → CLI uses them

**Function signature**:
```go
DetermineMode(cmd, requiredFlags) (useTUI bool, err error)
```

**Usage**:
```go
if useTUI {
    return runCommandTUI(cmd)
}
return runCommandCLI(cmd)
```

Simple, clear, correct! ✅
