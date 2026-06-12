# Mode Detection Package

## Logic

There are **2 modes**: TUI (interactive) and CLI (non-interactive)

### Mode Selection Rules

1. **Not in terminal** → **CLI mode required**
   - If flags missing → **ERROR**
   - If all flags present → CLI mode

2. **In terminal + flags missing** → **TUI mode**
   - Ask for flags interactively

3. **In terminal + all flags present** → **CLI mode**
   - User explicitly chose to provide flags

## Functions

### `IsTerminal() bool`

Checks if running in an interactive terminal.

### `HasAllRequiredFlags(cmd, requiredFlags) bool`

Checks if all required flags have been provided.

### `GetMissingFlags(cmd, requiredFlags) []string`

Returns list of missing flag names.

### `DetermineMode(cmd, requiredFlags) (useTUI bool, err error)`

Determines which mode to use based on terminal state and flags.

**Returns**:
- `(true, nil)` → Use TUI mode (in terminal, missing flags)
- `(false, nil)` → Use CLI mode (all flags present OR in terminal with all flags)
- `(false, error)` → Cannot run (not in terminal AND missing flags)

**Logic Flow**:
```
┌─────────────────┐
│  In terminal?   │
└────┬────────┬───┘
     NO       YES
     │        │
     ▼        ▼
┌─────────┐  ┌──────────────┐
│All flags?│  │  All flags?  │
└─┬────┬──┘  └──┬────────┬──┘
  NO   YES      NO       YES
  │    │        │        │
  ▼    ▼        ▼        ▼
ERROR  CLI      TUI      CLI
```

## Usage Pattern

```go
func runCommand(cmd *cobra.Command, args []string) error {
    requiredFlags := []string{"namespace", "size"}
    
    useTUI, err := mode.DetermineMode(cmd, requiredFlags)
    if err != nil {
        return err  // Not in terminal and missing flags
    }
    
    if useTUI {
        // TUI mode - collect missing flags interactively
        return runCommandTUI(cmd)
    }
    
    // CLI mode - use provided flags
    return runCommandCLI(cmd)
}
```

## Examples

### Example 1: Terminal + Missing Flags → TUI
```bash
$ ./obstool update monitoring
# In terminal, no --image flag
# → TUI mode (would ask for image interactively)
```

**Flow**:
1. `IsTerminal()` → `true`
2. `HasAllRequiredFlags()` → `false`
3. `DetermineMode()` → `(true, nil)`
4. ✅ Runs `runCommandTUI()`

### Example 2: Terminal + All Flags → CLI
```bash
$ ./obstool update monitoring --image=quay.io/test/image:v1
# In terminal, all flags provided
# → CLI mode
```

**Flow**:
1. `IsTerminal()` → `true`
2. `HasAllRequiredFlags()` → `true`
3. `DetermineMode()` → `(false, nil)`
4. ✅ Runs `runCommandCLI()`

### Example 3: Pipe + All Flags → CLI
```bash
$ ./obstool update monitoring --image=quay.io/test/image:v1 | tee log.txt
# Not in terminal, all flags provided
# → CLI mode
```

**Flow**:
1. `IsTerminal()` → `false`
2. `HasAllRequiredFlags()` → `true`
3. `DetermineMode()` → `(false, nil)`
4. ✅ Runs `runCommandCLI()`

### Example 4: Pipe + Missing Flags → ERROR
```bash
$ echo | ./obstool update monitoring
# Not in terminal, no --image flag
# → Error
```

**Flow**:
1. `IsTerminal()` → `false`
2. `HasAllRequiredFlags()` → `false`
3. `DetermineMode()` → `(false, error)`
4. ❌ Returns error: "not running in terminal and missing required flags: --image"

## Command-Specific Configuration

Each command defines its own required flags:

```go
// Simple command
requiredFlags := []string{"image"}

// Complex command
requiredFlags := []string{"namespace", "data-model", "size"}

// Safety command
requiredFlags := []string{"confirm"}
```

## Error Messages

When not in terminal and flags missing:
```
Error: not running in terminal and missing required flags: --image
```

Multiple missing flags:
```
Error: not running in terminal and missing required flags: --namespace, --data-model, --size
```
