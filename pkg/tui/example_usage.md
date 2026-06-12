# TUI Framework Usage Examples

## Deploy Selection TUI

```go
package main

import (
    "fmt"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/observability-ui/development-tools/pkg/tui"
)

func main() {
    choices := []string{
        "cluster-observability-operator",
        "logging",
        "tracing",
        "dashboards",
        "monitoring",
    }
    
    model := tui.NewDeploySelectionModel(choices)
    program := tea.NewProgram(model)
    
    finalModel, err := program.Run()
    if err != nil {
        panic(err)
    }
    
    if result := finalModel.(tui.DeploySelectionModel).Result(); result != nil {
        selected := result.([]string)
        fmt.Printf("Selected components: %v\n", selected)
    }
}
```

## Progress TUI

```go
package main

import (
    "fmt"
    "time"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/observability-ui/development-tools/pkg/tui"
)

func main() {
    operations := []string{
        "Creating namespace",
        "Deploying LokiStack",
        "Deploying ClusterLogForwarder",
        "Deploying UIPlugin",
    }
    
    model := tui.NewProgressModel("Deploying Logging Stack", operations)
    program := tea.NewProgram(model)
    
    // Simulate operations in background
    go func() {
        for i := range operations {
            time.Sleep(1 * time.Second)
            program.Send(tui.OperationUpdateMsg{
                Index:  i,
                Status: tui.OperationInProgress,
            })
            
            time.Sleep(2 * time.Second)
            program.Send(tui.OperationUpdateMsg{
                Index:  i,
                Status: tui.OperationComplete,
            })
        }
    }()
    
    finalModel, err := program.Run()
    if err != nil {
        panic(err)
    }
    
    if finalModel.(tui.ProgressModel).Error() != nil {
        fmt.Printf("Error: %v\n", finalModel.(tui.ProgressModel).Error())
    }
}
```

## Mode Detection + TUI Integration

```go
package deploy

import (
    "fmt"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/spf13/cobra"
    
    "github.com/observability-ui/development-tools/pkg/mode"
    "github.com/observability-ui/development-tools/pkg/tui"
    execctx "github.com/observability-ui/development-tools/pkg/context"
)

var loggingCmd = &cobra.Command{
    Use:   "logging",
    Short: "Deploy logging stack",
    RunE:  runDeployLogging,
}

func init() {
    loggingCmd.Flags().String("namespace", "", "Namespace for logging stack")
    loggingCmd.Flags().String("data-model", "", "Data model (otel or viaq)")
    loggingCmd.Flags().Bool("skip-ui-plugin", false, "Skip deploying UI plugin")
}

func runDeployLogging(cmd *cobra.Command, args []string) error {
    requiredFlags := []string{"namespace", "data-model"}
    
    // Default: TUI mode (when in terminal without all flags)
    // Override: CLI mode (when all flags provided)
    mustUseFlags, err := mode.MustUseFlags(cmd, requiredFlags)
    if err != nil {
        // Not in terminal and missing flags
        return err
    }
    
    if mustUseFlags {
        // CLI mode - all flags present or not in terminal
        return runDeployLoggingCLI(cmd)
    }
    
    // TUI mode - in terminal, missing flags (interactive)
    return runDeployLoggingTUI(cmd)
}

func runDeployLoggingCLI(cmd *cobra.Command) error {
    ctx := cmd.Context()
    ctx = execctx.WithTUI(ctx, false)
    
    namespace, _ := cmd.Flags().GetString("namespace")
    dataModel, _ := cmd.Flags().GetString("data-model")
    skipUIPlugin, _ := cmd.Flags().GetBool("skip-ui-plugin")
    
    return deployLogging(ctx, namespace, dataModel, skipUIPlugin)
}

func runDeployLoggingTUI(cmd *cobra.Command) error {
    // TUI mode - collect missing values interactively
    // This is a simplified example - real implementation would use
    // a custom TUI model to collect the configuration
    
    choices := []string{"otel", "viaq"}
    model := tui.NewDeploySelectionModel(choices)
    program := tea.NewProgram(model)
    
    finalModel, err := program.Run()
    if err != nil {
        return err
    }
    
    // Extract selected values and deploy
    ctx := cmd.Context()
    ctx = execctx.WithTUI(ctx, true)
    
    // ... rest of TUI-specific logic
    return nil
}
```

## Output Handler Usage

```go
package deploy

import (
    "context"
    "github.com/observability-ui/development-tools/pkg/output"
    execctx "github.com/observability-ui/development-tools/pkg/context"
)

func deployLogging(ctx context.Context, namespace, dataModel string, skipUIPlugin bool) error {
    out := output.NewHandler(ctx)
    
    out.Info(fmt.Sprintf("Deploying logging to namespace: %s", namespace))
    
    out.Progress("Creating LokiStack...")
    if err := createLokiStack(ctx, namespace); err != nil {
        out.Error(fmt.Sprintf("Failed to create LokiStack: %v", err))
        return err
    }
    out.Success("LokiStack created")
    
    out.Progress("Creating ClusterLogForwarder...")
    if err := createCLF(ctx, namespace, dataModel); err != nil {
        out.Error(fmt.Sprintf("Failed to create ClusterLogForwarder: %v", err))
        return err
    }
    out.Success("ClusterLogForwarder created")
    
    if !skipUIPlugin {
        out.Progress("Deploying UI plugin...")
        if err := deployUIPlugin(ctx); err != nil {
            out.Error(fmt.Sprintf("Failed to deploy UI plugin: %v", err))
            return err
        }
        out.Success("UI plugin deployed")
    }
    
    return nil
}
```
