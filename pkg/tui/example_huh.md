# TUI Input Collection with Huh

We use [huh](https://github.com/charmbracelet/huh) for forms and [bubbles](https://github.com/charmbracelet/bubbles) for input components.

## Why Huh?

- ✅ **Paste support** - Handles pasted text correctly
- ✅ **Validation** - Built-in validation framework
- ✅ **Accessibility** - Screen reader support
- ✅ **Battle-tested** - Used in production by many CLI tools
- ✅ **Rich components** - Input, Select, MultiSelect, Confirm, etc.
- ✅ **Theming** - Customizable styles

## Simple Input

```go
package main

import (
    "fmt"
    "github.com/charmbracelet/huh"
)

func main() {
    var name string
    
    form := huh.NewForm(
        huh.NewGroup(
            huh.NewInput().
                Title("What's your name?").
                Placeholder("John Doe").
                Value(&name),
        ),
    )
    
    err := form.Run()
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Hello, %s!\n", name)
}
```

## Input with Validation

```go
func collectImageInput() (string, error) {
    var image string
    
    form := huh.NewForm(
        huh.NewGroup(
            huh.NewInput().
                Title("Enter monitoring plugin image").
                Placeholder("quay.io/observability-ui/monitoring-plugin:latest").
                Value(&image).
                Validate(func(s string) error {
                    if s == "" {
                        return fmt.Errorf("image cannot be empty")
                    }
                    if !strings.Contains(s, ":") {
                        return fmt.Errorf("image must include a tag")
                    }
                    return nil
                }),
        ),
    )
    
    err := form.Run()
    if err != nil {
        return "", err
    }
    
    return image, nil
}
```

## Multi-Field Form

```go
func collectDeployInputs() (namespace, size, dataModel string, err error) {
    form := huh.NewForm(
        huh.NewGroup(
            huh.NewInput().
                Title("Namespace").
                Placeholder("logging").
                Value(&namespace).
                Validate(func(s string) error {
                    if s == "" {
                        return fmt.Errorf("namespace is required")
                    }
                    return nil
                }),
            
            huh.NewSelect[string]().
                Title("Stack Size").
                Options(
                    huh.NewOption("Small", "small"),
                    huh.NewOption("Medium", "medium"),
                    huh.NewOption("Large", "large"),
                ).
                Value(&size),
            
            huh.NewSelect[string]().
                Title("Data Model").
                Options(
                    huh.NewOption("OpenTelemetry", "otel"),
                    huh.NewOption("ViaQ", "viaq"),
                ).
                Value(&dataModel),
        ),
    )
    
    err = form.Run()
    if err != nil {
        return "", "", "", err
    }
    
    return namespace, size, dataModel, nil
}
```

## Selection (Single)

```go
func selectComponent() (string, error) {
    var component string
    
    form := huh.NewForm(
        huh.NewGroup(
            huh.NewSelect[string]().
                Title("Which component to deploy?").
                Options(
                    huh.NewOption("Cluster Observability Operator", "coo"),
                    huh.NewOption("Logging", "logging"),
                    huh.NewOption("Tracing", "tracing"),
                    huh.NewOption("Dashboards", "dashboards"),
                    huh.NewOption("Monitoring", "monitoring"),
                ).
                Value(&component),
        ),
    )
    
    err := form.Run()
    if err != nil {
        return "", err
    }
    
    return component, nil
}
```

## Multi-Select

```go
func selectComponents() ([]string, error) {
    var components []string
    
    form := huh.NewForm(
        huh.NewGroup(
            huh.NewMultiSelect[string]().
                Title("Select components to deploy").
                Options(
                    huh.NewOption("COO", "coo"),
                    huh.NewOption("Logging", "logging"),
                    huh.NewOption("Tracing", "tracing"),
                    huh.NewOption("Dashboards", "dashboards"),
                    huh.NewOption("Monitoring", "monitoring"),
                ).
                Value(&components),
        ),
    )
    
    err := form.Run()
    if err != nil {
        return nil, err
    }
    
    return components, nil
}
```

## Confirmation

```go
func confirmDeletion() (bool, error) {
    var confirmed bool
    
    form := huh.NewForm(
        huh.NewGroup(
            huh.NewConfirm().
                Title("Delete all observability components?").
                Description("This action cannot be undone.").
                Affirmative("Yes, delete").
                Negative("No, cancel").
                Value(&confirmed),
        ),
    )
    
    err := form.Run()
    if err != nil {
        return false, err
    }
    
    return confirmed, nil
}
```

## Conditional Fields

```go
func collectCOOInputs() (method, image, version string, err error) {
    form := huh.NewForm(
        huh.NewGroup(
            huh.NewSelect[string]().
                Title("Install method").
                Options(
                    huh.NewOption("Bundle", "bundle"),
                    huh.NewOption("FBC", "fbc"),
                    huh.NewOption("Stage", "stage"),
                    huh.NewOption("OperatorHub", "operatorhub"),
                ).
                Value(&method),
        ),
        
        // Conditional group - only for bundle method
        huh.NewGroup(
            huh.NewInput().
                Title("Bundle image").
                Placeholder("quay.io/...").
                Value(&image).
                Validate(func(s string) error {
                    if s == "" {
                        return fmt.Errorf("bundle image is required")
                    }
                    return nil
                }),
        ).WithHideFunc(func() bool {
            return method != "bundle"
        }),
        
        // Conditional group - only for stage method
        huh.NewGroup(
            huh.NewInput().
                Title("Version").
                Placeholder("v0.3.0").
                Value(&version),
        ).WithHideFunc(func() bool {
            return method != "stage"
        }),
    )
    
    err = form.Run()
    if err != nil {
        return "", "", "", err
    }
    
    return method, image, version, nil
}
```

## Integration with Command

```go
package deploy

import (
    "fmt"
    "github.com/charmbracelet/huh"
    "github.com/spf13/cobra"
    
    "github.com/observability-ui/development-tools/pkg/mode"
    execctx "github.com/observability-ui/development-tools/pkg/context"
)

var loggingCmd = &cobra.Command{
    Use:   "logging",
    Short: "Deploy logging stack",
    RunE:  runDeployLogging,
}

func init() {
    loggingCmd.Flags().String("namespace", "", "Namespace")
    loggingCmd.Flags().String("size", "", "Stack size")
    loggingCmd.Flags().String("data-model", "", "Data model")
}

func runDeployLogging(cmd *cobra.Command, args []string) error {
    requiredFlags := []string{"namespace", "size", "data-model"}
    
    useTUI, err := mode.DetermineMode(cmd, requiredFlags)
    if err != nil {
        return err
    }
    
    if useTUI {
        return runDeployLoggingTUI(cmd)
    }
    
    return runDeployLoggingCLI(cmd)
}

func runDeployLoggingTUI(cmd *cobra.Command) error {
    var namespace, size, dataModel string
    
    // Collect missing values
    form := huh.NewForm(
        huh.NewGroup(
            huh.NewInput().
                Title("Namespace").
                Placeholder("logging").
                Value(&namespace).
                Validate(func(s string) error {
                    if s == "" {
                        return fmt.Errorf("namespace is required")
                    }
                    return nil
                }),
            
            huh.NewSelect[string]().
                Title("Stack Size").
                Options(
                    huh.NewOption("Small (development)", "small"),
                    huh.NewOption("Medium (staging)", "medium"),
                    huh.NewOption("Large (production)", "large"),
                ).
                Value(&size),
            
            huh.NewSelect[string]().
                Title("Data Model").
                Options(
                    huh.NewOption("OpenTelemetry (recommended)", "otel"),
                    huh.NewOption("ViaQ (legacy)", "viaq"),
                ).
                Value(&dataModel),
        ),
    )
    
    err := form.Run()
    if err != nil {
        return err
    }
    
    // Deploy with collected values
    ctx := cmd.Context()
    ctx = execctx.WithTUI(ctx, true)
    
    return deployLogging(ctx, namespace, size, dataModel)
}
```

## Available Components

From `huh`:
- ✅ **Input** - Text input with paste support
- ✅ **Text** - Multi-line text area
- ✅ **Select** - Single selection from list
- ✅ **MultiSelect** - Multiple selections
- ✅ **Confirm** - Yes/No confirmation
- ✅ **FilePicker** - File selection
- ✅ **Note** - Display information

## Features

### Input
- Paste support (Ctrl+V)
- Character masking (for passwords)
- Validation
- Placeholders
- Suggestions/autocomplete

### Select
- Keyboard navigation
- Filtering/search
- Custom rendering
- Descriptions

### MultiSelect
- Space to toggle
- Select all/none
- Limit selection count

### Validation
- Real-time validation
- Custom error messages
- Async validation support

## Styling

```go
theme := huh.ThemeBase16()

form := huh.NewForm(
    huh.NewGroup(
        huh.NewInput().
            Title("Name").
            Value(&name),
    ),
).WithTheme(theme)
```

## Error Handling

```go
err := form.Run()
if err == huh.ErrUserAborted {
    return fmt.Errorf("input cancelled")
}
if err != nil {
    return fmt.Errorf("form error: %w", err)
}
```

## Keyboard Shortcuts

- **Tab** - Next field
- **Shift+Tab** - Previous field
- **Enter** - Submit/Select
- **Esc** - Cancel
- **↑/↓** - Navigate options
- **Space** - Toggle selection (MultiSelect)
- **Ctrl+C** - Quit
- **Ctrl+V** - Paste (Input fields)

## References

- [Huh GitHub](https://github.com/charmbracelet/huh)
- [Bubbles GitHub](https://github.com/charmbracelet/bubbles)
- [Huh Examples](https://github.com/charmbracelet/huh/tree/main/examples)
