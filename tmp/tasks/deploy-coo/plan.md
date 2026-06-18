# Implementation Plan: Deploy Command Group & Deploy COO

**Status**: Ready for Implementation  
**Created**: 2026-06-18  
**Implements**: [TASKS.md](../../TASKS.md) - Deploy Command Group & Deploy COO Command

---

## Overview

Implement two related tasks:
1. **Deploy command group** - Parent `obstool deploy` command with TUI component selection
2. **Deploy COO command** - Four deployment methods (bundle, fbc, stage, operatorhub) for Cluster Observability Operator

This replaces bash scripts in `qe-release-testing/{bundle,fbc,stage}/` and provides unified Go-based deployment.

---

## Requirements

### Functional Requirements

**Deploy Command Group:**
- When run without subcommand (`obstool deploy`), shows TUI multi-select for components
- Initially shows only "COO" option (other commands add themselves when implemented)
- When run with subcommand (`obstool deploy coo`), delegates to that subcommand
- No flags needed for parent command

**Deploy COO - Four Methods:**

1. **Bundle Method** (`--method=bundle`)
   - Uses operator-sdk binary via `exec.Command`
   - Creates IDMS based on registry type (quay or stage)
   - Detects OCP version for security context (4.19+ uses restricted)
   - Equivalent to: `operator-sdk run bundle <url> --install-mode AllNamespaces`

2. **FBC Method** (`--method=fbc`)
   - Creates IDMS for quay
   - Creates CatalogSource with FBC image
   - Creates Subscription
   - Waits for CSV to be ready

3. **Stage Method** (`--method=stage`)
   - Same as FBC but with stage-specific IDMS (includes brew registry)
   - Used for stage registry testing

4. **OperatorHub Method** (`--method=operatorhub`)
   - Simplest - just creates Subscription to default catalog
   - No IDMS or CatalogSource needed
   - Uses channel flag (default: stable)

**Common Behavior (All Methods):**
- Creates namespace with monitoring label: `openshift.io/cluster-monitoring: "true"`
- Creates OperatorGroup in namespace
- Waits for CSV to reach Succeeded phase with timeout
- Progress tracking in both CLI and TUI modes
- Idempotent operations

**Deferred:**
- ⚠️ **Scheduler patching** removed from this implementation (see Deferred Items section)
- ⚠️ **Perses namespace** NOT created by COO (left to dashboards deployment)

### Flags

**Deploy COO Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--method` | string | - | Yes | Deployment method: bundle, fbc, stage, operatorhub |
| `--bundle-url` | string | - | If bundle | Bundle image URL |
| `--fbc-url` | string | - | If fbc/stage | FBC image URL |
| `--channel` | string | "stable" | No | Subscription channel (operatorhub method) |
| `--namespace` | string | "openshift-cluster-observability-operator" | No | Namespace for COO |
| `--registry-type` | string | "quay" | If bundle/fbc | Registry type: quay or stage (affects IDMS) |

**Required Flags by Method:**
- Bundle: `--method`, `--bundle-url`
- FBC: `--method`, `--fbc-url`
- Stage: `--method`, `--fbc-url`
- OperatorHub: `--method` only

### Mode Support

**CLI Mode** (all required flags present):
```bash
obstool deploy coo --method=bundle --bundle-url=quay.io/...
obstool deploy coo --method=fbc --fbc-url=quay.io/...
obstool deploy coo --method=stage --fbc-url=registry.stage.redhat.io/...
obstool deploy coo --method=operatorhub --channel=stable
```

**TUI Mode** (missing required flags):
```bash
obstool deploy coo
# Shows form to select method and collect method-specific inputs
```

---

## Architecture

### Command Flow - Deploy Command Group

```
cmd/deploy/deploy.go
  ↓
No subcommand? → Show TUI component selection
  ├─ Show multi-select with: [ ] COO
  ├─ User selects components
  └─ Execute selected commands
  ↓
Subcommand provided? → Cobra delegates to subcommand
```

### Command Flow - Deploy COO

```
cmd/deploy/coo.go
  ↓
Mode Detection (CLI vs TUI)
  ↓
Validate inputs (method-specific flags)
  ↓
pkg/operations/coo.go::DeployCOO()
  ├─ Step 1: Create namespace with monitoring label
  ├─ Step 2: Create OperatorGroup
  ├─ Step 3: Deploy via method (delegate to pkg/operators/coo/)
  │    ├─ bundle.go::DeployBundle()
  │    ├─ fbc.go::DeployFBC()
  │    ├─ stage.go::DeployStage()
  │    └─ operatorhub.go::DeployOperatorHub()
  └─ Step 4: Wait for CSV to be ready
  ↓
Output results (CLI handler or TUI progress)
```

### Executor Pattern

**Step Constants (operations/coo.go):**
```go
const (
    StepEnsureNamespace = iota
    StepEnsureOperatorGroup
    StepDeployMethod
    StepWaitForCSV
)
```

**Progress Updates:**
- Step 1: "Create namespace openshift-cluster-observability-operator"
- Step 2: "Create OperatorGroup"
- Step 3: "Deploy COO ({method})"
  - Sub-logs for method-specific steps (IDMS, CatalogSource, etc.)
- Step 4: "Wait for operator to be ready"
  - Uses existing CSV countdown from OLM utilities

---

## Implementation Details

### 1. File Structure

**New Files:**
```
cmd/deploy/
├── deploy.go                 # Deploy command group
└── coo.go                    # COO deployment command

pkg/operations/
└── coo.go                    # DeployCOO business logic

pkg/operators/coo/
├── bundle.go                 # Bundle method implementation
├── fbc.go                    # FBC method implementation
├── stage.go                  # Stage method implementation
└── operatorhub.go            # OperatorHub method implementation

internal/constants/
└── coo.go                    # COO-specific constants

pkg/k8s/
└── namespace.go              # Namespace creation with labels helper
```

**Modified/Enhanced Files:**
```
pkg/operators/idms.go         # Add COO-specific IDMS functions
cmd/root.go                   # Register deploy command group
tmp/TASKS.md                  # Add deferred scheduler patching task
```

---

### 2. Constants (`internal/constants/coo.go`)

```go
package constants

const (
    COONamespace         = "openshift-cluster-observability-operator"
    COOOperatorName      = "cluster-observability-operator"
    COOPackageName       = "cluster-observability-operator"
    COOCatalogName       = "observability-operator"
    MarketplaceNamespace = "openshift-marketplace"
    DefaultCOOChannel    = "stable"
    
    // IDMS names
    IDMSCOOQuay  = "idms-coo-quay"
    IDMSCOOStage = "idms-coo-stage"
    
    // OperatorGroup
    COOOperatorGroupName = "observability-operator-group"
)
```

---

### 3. Deploy Command Group (`cmd/deploy/deploy.go`)

**Purpose**: Parent command with TUI component selection

```go
package deploy

import (
    "github.com/charmbracelet/huh"
    "github.com/spf13/cobra"
)

var DeployCmd = &cobra.Command{
    Use:   "deploy",
    Short: "Deploy observability components",
    Long: `Deploy observability components to the cluster.

Run without subcommand for interactive component selection.
Run with subcommand to deploy specific component directly.

Available components:
  - coo: Cluster Observability Operator
  - logging: Logging stack (LokiStack + ClusterLogForwarder)
  - tracing: Tracing stack (TempoStack + OTEL)
  - dashboards: Perses dashboards
  - monitoring: Monitoring plugin
  - acm: ACM observability
  - korrel8r: Korrel8r + NetObserv`,
    RunE: runDeploy,
}

func runDeploy(cmd *cobra.Command, args []string) error {
    // If subcommand was used, Cobra handles it automatically
    // If no subcommand, show TUI selection
    return runDeployTUI(cmd)
}

func runDeployTUI(cmd *cobra.Command) error {
    var selectedComponents []string
    
    form := huh.NewForm(
        huh.NewGroup(
            huh.NewMultiSelect[string]().
                Title("Select components to deploy").
                Description("Components will be deployed sequentially in order selected").
                Options(
                    huh.NewOption("Cluster Observability Operator", "coo"),
                    // Future: Other options added when those commands implemented
                ).
                Value(&selectedComponents),
        ),
    )
    
    if err := form.Run(); err != nil {
        return err
    }
    
    if len(selectedComponents) == 0 {
        fmt.Println("No components selected")
        return nil
    }
    
    // Execute selected components sequentially
    for _, component := range selectedComponents {
        switch component {
        case "coo":
            if err := runDeployCOO(cmd, []string{}); err != nil {
                return fmt.Errorf("failed to deploy COO: %w", err)
            }
        // Future: Other cases when implemented
        }
    }
    
    return nil
}
```

**Notes:**
- Initially only shows COO option
- Other deploy commands will add their options when implemented
- Parent command has no flags
- Delegates to subcommand RunE functions

---

### 4. Deploy COO Command (`cmd/deploy/coo.go`)

**Purpose**: CLI entry point with mode detection and method-specific TUI forms

```go
package deploy

import (
    "fmt"
    "strconv"
    
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/huh"
    "github.com/spf13/cobra"
    
    "github.com/observability-ui/development-tools/internal/constants"
    execctx "github.com/observability-ui/development-tools/pkg/context"
    "github.com/observability-ui/development-tools/pkg/executor"
    "github.com/observability-ui/development-tools/pkg/mode"
    "github.com/observability-ui/development-tools/pkg/operations"
    "github.com/observability-ui/development-tools/pkg/output"
    "github.com/observability-ui/development-tools/pkg/tui"
)

type DeployCOOConfig struct {
    Method       string
    Namespace    string
    BundleURL    string
    FBCURL       string
    Channel      string
    RegistryType string
}

var cooCmd = &cobra.Command{
    Use:   "coo",
    Short: "Deploy Cluster Observability Operator",
    Long: `Deploy Cluster Observability Operator using one of four methods:

  bundle:      Deploy from bundle image using operator-sdk
  fbc:         Deploy from File-Based Catalog (FBC) image
  stage:       Deploy from stage registry FBC
  operatorhub: Deploy from default OperatorHub catalog

Method-specific requirements:
  bundle:      --bundle-url required, --registry-type optional (quay|stage)
  fbc:         --fbc-url required, --registry-type optional (quay|stage)
  stage:       --fbc-url required (uses stage IDMS automatically)
  operatorhub: --channel optional (default: stable)

All methods create namespace and OperatorGroup, then wait for CSV to be ready.`,
    Example: `  # Deploy from bundle
  obstool deploy coo --method=bundle --bundle-url=quay.io/rhobs/coo-bundle:v0.3.6

  # Deploy from FBC with quay registry
  obstool deploy coo --method=fbc --fbc-url=quay.io/rhobs/coo-catalog:latest

  # Deploy from stage registry
  obstool deploy coo --method=stage --fbc-url=registry.stage.redhat.io/...

  # Deploy from OperatorHub
  obstool deploy coo --method=operatorhub --channel=stable

  # Interactive mode (TUI prompts for inputs)
  obstool deploy coo`,
    RunE: runDeployCOO,
}

func init() {
    cooCmd.Flags().String("method", "", "Deployment method: bundle, fbc, stage, operatorhub")
    cooCmd.Flags().String("bundle-url", "", "Bundle image URL (bundle method)")
    cooCmd.Flags().String("fbc-url", "", "FBC image URL (fbc/stage methods)")
    cooCmd.Flags().String("channel", "stable", "Subscription channel (operatorhub method)")
    cooCmd.Flags().String("namespace", constants.COONamespace, "Namespace for COO")
    cooCmd.Flags().String("registry-type", "quay", "Registry type: quay or stage (bundle/fbc methods)")
    
    DeployCmd.AddCommand(cooCmd)
}

func runDeployCOO(cmd *cobra.Command, args []string) error {
    method, _ := cmd.Flags().GetString("method")
    
    requiredFlags := getRequiredFlagsForMethod(method)
    useTUI, err := mode.DetermineMode(cmd, requiredFlags)
    if err != nil {
        return err
    }
    
    if useTUI {
        return runDeployCOOTUI(cmd)
    }
    return runDeployCOOCLI(cmd)
}

func getRequiredFlagsForMethod(method string) []string {
    switch method {
    case "bundle":
        return []string{"method", "bundle-url"}
    case "fbc":
        return []string{"method", "fbc-url"}
    case "stage":
        return []string{"method", "fbc-url"}
    case "operatorhub":
        return []string{"method"}
    default:
        return []string{"method"}
    }
}

func runDeployCOOCLI(cmd *cobra.Command) error {
    config, err := getConfigFromFlags(cmd)
    if err != nil {
        return err
    }
    
    ctx := cmd.Context()
    ctx = execctx.WithTUI(ctx, false)
    
    kubeClient, err := execctx.GetClient(ctx)
    if err != nil {
        return err
    }
    
    handler := output.NewCLIHandler()
    exec := executor.NewExecutor()
    
    go operations.DeployCOO(ctx, kubeClient, config, exec)
    
    for update := range exec.UpdateCh {
        if err := handler.HandleUpdate(update); err != nil {
            return err
        }
    }
    
    return nil
}

func runDeployCOOTUI(cmd *cobra.Command) error {
    ctx := cmd.Context()
    ctx = execctx.WithTUI(ctx, true)
    
    kubeClient, err := execctx.GetClient(ctx)
    if err != nil {
        return err
    }
    
    // Collect inputs via forms
    config, err := collectCOOInput(cmd)
    if err != nil {
        return err
    }
    
    operationsList := []string{
        fmt.Sprintf("Create namespace %s", config.Namespace),
        "Create OperatorGroup",
        fmt.Sprintf("Deploy COO (%s)", config.Method),
        "Wait for operator to be ready",
    }
    
    model := tui.NewProgressModel("Deploying Cluster Observability Operator", operationsList)
    program := tea.NewProgram(model)
    
    exec := executor.NewExecutor()
    
    go operations.DeployCOO(ctx, kubeClient, config, exec)
    
    go func() {
        for update := range exec.UpdateCh {
            if update.Message != "" {
                continue
            }
            
            program.Send(tui.OperationUpdateMsg{
                Index:  update.Index,
                Status: convertStatus(update.Status),
                Error:  update.Error,
            })
        }
    }()
    
    finalModel, err := program.Run()
    if err != nil {
        return err
    }
    
    model = finalModel.(tui.ProgressModel)
    if model.Error() != nil {
        return model.Error()
    }
    
    return nil
}

func getConfigFromFlags(cmd *cobra.Command) (operations.DeployCOOConfig, error) {
    method, _ := cmd.Flags().GetString("method")
    if method == "" {
        return operations.DeployCOOConfig{}, fmt.Errorf("--method is required")
    }
    
    config := operations.DeployCOOConfig{
        Method:       method,
        Namespace:    cmd.Flags().Lookup("namespace").Value.String(),
        Channel:      cmd.Flags().Lookup("channel").Value.String(),
        RegistryType: cmd.Flags().Lookup("registry-type").Value.String(),
    }
    
    switch method {
    case "bundle":
        bundleURL, _ := cmd.Flags().GetString("bundle-url")
        if bundleURL == "" {
            return config, fmt.Errorf("--bundle-url is required for bundle method")
        }
        config.BundleURL = bundleURL
    case "fbc":
        fbcURL, _ := cmd.Flags().GetString("fbc-url")
        if fbcURL == "" {
            return config, fmt.Errorf("--fbc-url is required for fbc method")
        }
        config.FBCURL = fbcURL
    case "stage":
        fbcURL, _ := cmd.Flags().GetString("fbc-url")
        if fbcURL == "" {
            return config, fmt.Errorf("--fbc-url is required for stage method")
        }
        config.FBCURL = fbcURL
        config.RegistryType = "stage" // Override
    case "operatorhub":
        // No additional required flags
    default:
        return config, fmt.Errorf("unknown method: %s (must be: bundle, fbc, stage, operatorhub)", method)
    }
    
    return config, nil
}

func collectCOOInput(cmd *cobra.Command) (operations.DeployCOOConfig, error) {
    var config operations.DeployCOOConfig
    
    // Step 1: Select method
    methodOptions := []huh.Option[string]{
        huh.NewOption("Bundle (operator-sdk)", "bundle"),
        huh.NewOption("FBC (File-Based Catalog)", "fbc"),
        huh.NewOption("Stage Registry", "stage"),
        huh.NewOption("OperatorHub (default catalog)", "operatorhub"),
    }
    
    methodForm := huh.NewForm(
        huh.NewGroup(
            huh.NewSelect[string]().
                Title("Select deployment method").
                Description("Choose how to deploy Cluster Observability Operator").
                Options(methodOptions...).
                Value(&config.Method),
        ),
    )
    
    if err := methodForm.Run(); err != nil {
        return config, err
    }
    
    // Step 2: Method-specific inputs
    switch config.Method {
    case "bundle":
        bundleForm := huh.NewForm(
            huh.NewGroup(
                huh.NewInput().
                    Title("Bundle image URL").
                    Placeholder("quay.io/rhobs/coo-bundle:v0.3.6").
                    Value(&config.BundleURL).
                    Validate(func(s string) error {
                        if s == "" {
                            return fmt.Errorf("bundle URL cannot be empty")
                        }
                        return nil
                    }),
                huh.NewSelect[string]().
                    Title("Registry type").
                    Description("Affects IDMS configuration").
                    Options(
                        huh.NewOption("Quay", "quay"),
                        huh.NewOption("Stage", "stage"),
                    ).
                    Value(&config.RegistryType),
            ),
        )
        if err := bundleForm.Run(); err != nil {
            return config, err
        }
        
    case "fbc":
        fbcForm := huh.NewForm(
            huh.NewGroup(
                huh.NewInput().
                    Title("FBC image URL").
                    Placeholder("quay.io/rhobs/coo-catalog:latest").
                    Value(&config.FBCURL).
                    Validate(func(s string) error {
                        if s == "" {
                            return fmt.Errorf("FBC URL cannot be empty")
                        }
                        return nil
                    }),
                huh.NewSelect[string]().
                    Title("Registry type").
                    Options(
                        huh.NewOption("Quay", "quay"),
                        huh.NewOption("Stage", "stage"),
                    ).
                    Value(&config.RegistryType),
            ),
        )
        if err := fbcForm.Run(); err != nil {
            return config, err
        }
        
    case "stage":
        stageForm := huh.NewForm(
            huh.NewGroup(
                huh.NewInput().
                    Title("Stage FBC image URL").
                    Placeholder("registry.stage.redhat.io/...").
                    Value(&config.FBCURL).
                    Validate(func(s string) error {
                        if s == "" {
                            return fmt.Errorf("FBC URL cannot be empty")
                        }
                        return nil
                    }),
            ),
        )
        if err := stageForm.Run(); err != nil {
            return config, err
        }
        config.RegistryType = "stage"
        
    case "operatorhub":
        channelForm := huh.NewForm(
            huh.NewGroup(
                huh.NewInput().
                    Title("Subscription channel").
                    Placeholder("stable").
                    Value(&config.Channel).
                    Validate(func(s string) error {
                        if s == "" {
                            return fmt.Errorf("channel cannot be empty")
                        }
                        return nil
                    }),
            ),
        )
        if err := channelForm.Run(); err != nil {
            return config, err
        }
    }
    
    config.Namespace = constants.COONamespace
    return config, nil
}

func convertStatus(status executor.UpdateStatus) tui.OperationStatus {
    switch status {
    case executor.StatusPending:
        return tui.OperationPending
    case executor.StatusInProgress:
        return tui.OperationInProgress
    case executor.StatusComplete:
        return tui.OperationComplete
    case executor.StatusFailed:
        return tui.OperationFailed
    default:
        return tui.OperationPending
    }
}
```

---

### 5. Business Logic (`pkg/operations/coo.go`)

**Purpose**: Orchestrate COO deployment, delegate to method-specific implementations

```go
package operations

import (
    "context"
    "fmt"
    
    "sigs.k8s.io/controller-runtime/pkg/client"
    
    "github.com/observability-ui/development-tools/internal/constants"
    "github.com/observability-ui/development-tools/pkg/executor"
    "github.com/observability-ui/development-tools/pkg/k8s"
    "github.com/observability-ui/development-tools/pkg/operators"
    "github.com/observability-ui/development-tools/pkg/operators/coo"
)

const (
    StepEnsureNamespace = iota
    StepEnsureOperatorGroup
    StepDeployMethod
    StepWaitForCSV
)

type DeployCOOConfig struct {
    Method       string
    Namespace    string
    BundleURL    string
    FBCURL       string
    Channel      string
    RegistryType string
}

func DeployCOO(ctx context.Context, kubeClient client.Client, 
               config DeployCOOConfig, exec *executor.Executor) error {
    defer exec.Close()
    
    // Step 1: Ensure namespace with monitoring label
    stepName := fmt.Sprintf("Create namespace %s", config.Namespace)
    exec.SendUpdate(StepEnsureNamespace, executor.StatusInProgress, stepName)
    exec.SendLog(StepEnsureNamespace, "Ensuring namespace exists with monitoring label")
    
    created, err := k8s.EnsureNamespaceWithLabels(ctx, kubeClient, config.Namespace, map[string]string{
        "openshift.io/cluster-monitoring": "true",
    })
    if err != nil {
        exec.SendUpdateWithError(StepEnsureNamespace, executor.StatusFailed, stepName, err)
        return err
    }
    if created {
        exec.SendLog(StepEnsureNamespace, "Namespace created")
    } else {
        exec.SendLog(StepEnsureNamespace, "Namespace already exists")
    }
    exec.SendUpdate(StepEnsureNamespace, executor.StatusComplete, stepName)
    
    // Step 2: Ensure OperatorGroup
    stepName = "Create OperatorGroup"
    exec.SendUpdate(StepEnsureOperatorGroup, executor.StatusInProgress, stepName)
    exec.SendLog(StepEnsureOperatorGroup, "Ensuring OperatorGroup exists")
    
    created, err = operators.EnsureOperatorGroup(ctx, kubeClient, config.Namespace, constants.COOOperatorGroupName)
    if err != nil {
        exec.SendUpdateWithError(StepEnsureOperatorGroup, executor.StatusFailed, stepName, err)
        return err
    }
    if created {
        exec.SendLog(StepEnsureOperatorGroup, "OperatorGroup created")
    } else {
        exec.SendLog(StepEnsureOperatorGroup, "OperatorGroup already exists")
    }
    exec.SendUpdate(StepEnsureOperatorGroup, executor.StatusComplete, stepName)
    
    // Step 3: Deploy via method
    stepName = fmt.Sprintf("Deploy COO (%s)", config.Method)
    exec.SendUpdate(StepDeployMethod, executor.StatusInProgress, stepName)
    
    switch config.Method {
    case "bundle":
        err = coo.DeployBundle(ctx, kubeClient, config, exec)
    case "fbc":
        err = coo.DeployFBC(ctx, kubeClient, config, exec)
    case "stage":
        err = coo.DeployStage(ctx, kubeClient, config, exec)
    case "operatorhub":
        err = coo.DeployOperatorHub(ctx, kubeClient, config, exec)
    default:
        err = fmt.Errorf("unknown deployment method: %s", config.Method)
    }
    
    if err != nil {
        exec.SendUpdateWithError(StepDeployMethod, executor.StatusFailed, stepName, err)
        return err
    }
    exec.SendUpdate(StepDeployMethod, executor.StatusComplete, stepName)
    
    // Step 4: Wait for CSV to be ready
    stepName = "Wait for operator to be ready"
    exec.SendUpdate(StepWaitForCSV, executor.StatusInProgress, stepName)
    exec.SendLog(StepWaitForCSV, "Waiting for ClusterServiceVersion to reach Succeeded phase")
    
    err = operators.WaitForCSV(ctx, kubeClient, config.Namespace, constants.COOOperatorName, exec, StepWaitForCSV)
    if err != nil {
        exec.SendUpdateWithError(StepWaitForCSV, executor.StatusFailed, stepName, err)
        return err
    }
    exec.SendUpdate(StepWaitForCSV, executor.StatusComplete, stepName)
    
    return nil
}
```

---

### 6. Method Implementations

#### Bundle Method (`pkg/operators/coo/bundle.go`)

**Purpose**: Deploy using operator-sdk run bundle command

```go
package coo

import (
    "context"
    "fmt"
    "os/exec"
    
    "sigs.k8s.io/controller-runtime/pkg/client"
    
    "github.com/observability-ui/development-tools/pkg/executor"
    "github.com/observability-ui/development-tools/pkg/k8s"
    "github.com/observability-ui/development-tools/pkg/operators"
    "github.com/observability-ui/development-tools/pkg/operations"
)

func DeployBundle(ctx context.Context, kubeClient client.Client, 
                 config operations.DeployCOOConfig, exec *executor.Executor) error {
    
    // Check if operator-sdk is available
    _, err := exec.LookPath("operator-sdk")
    if err != nil {
        return fmt.Errorf("operator-sdk binary not found in PATH. Install from: https://sdk.operatorframework.io/docs/installation/")
    }
    
    // Create IDMS based on registry type
    exec.SendLog(operations.StepDeployMethod, "Creating ImageDigestMirrorSet")
    err = createIDMS(ctx, kubeClient, config.RegistryType)
    if err != nil {
        return fmt.Errorf("failed to create IDMS: %w", err)
    }
    
    // Detect OCP version for security context
    exec.SendLog(operations.StepDeployMethod, "Detecting OCP version")
    version, err := k8s.DetectVersion(ctx, kubeClient)
    if err != nil {
        return fmt.Errorf("failed to detect OCP version: %w", err)
    }
    
    // Build operator-sdk command
    exec.SendLog(operations.StepDeployMethod, fmt.Sprintf("Running operator-sdk bundle for %s", config.BundleURL))
    
    args := []string{"run", "bundle", config.BundleURL, 
        "--install-mode", "AllNamespaces",
        "--namespace", config.Namespace}
    
    if version.IsOCP419OrNewer() {
        args = append(args, "--security-context-config", "restricted")
        exec.SendLog(operations.StepDeployMethod, "Using restricted security context (OCP 4.19+)")
    } else {
        exec.SendLog(operations.StepDeployMethod, "Using default security context (OCP <4.19)")
    }
    
    cmd := exec.CommandContext(ctx, "operator-sdk", args...)
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("operator-sdk run bundle failed: %w\nOutput: %s", err, string(output))
    }
    
    exec.SendLog(operations.StepDeployMethod, "Bundle deployment initiated successfully")
    return nil
}

func createIDMS(ctx context.Context, kubeClient client.Client, registryType string) error {
    switch registryType {
    case "quay":
        return operators.EnsureIDMSQuay(ctx, kubeClient)
    case "stage":
        return operators.EnsureIDMSStage(ctx, kubeClient)
    default:
        return fmt.Errorf("unknown registry type: %s (must be quay or stage)", registryType)
    }
}
```

#### FBC Method (`pkg/operators/coo/fbc.go`)

**Purpose**: Deploy using FBC CatalogSource + Subscription

```go
package coo

import (
    "context"
    "fmt"
    
    "sigs.k8s.io/controller-runtime/pkg/client"
    
    "github.com/observability-ui/development-tools/internal/constants"
    "github.com/observability-ui/development-tools/pkg/executor"
    "github.com/observability-ui/development-tools/pkg/operators"
    "github.com/observability-ui/development-tools/pkg/operations"
)

func DeployFBC(ctx context.Context, kubeClient client.Client,
              config operations.DeployCOOConfig, exec *executor.Executor) error {
    
    // Create IDMS for quay
    exec.SendLog(operations.StepDeployMethod, "Creating ImageDigestMirrorSet for Quay")
    err := operators.EnsureIDMSQuay(ctx, kubeClient)
    if err != nil {
        return fmt.Errorf("failed to create IDMS: %w", err)
    }
    
    // Create CatalogSource
    exec.SendLog(operations.StepDeployMethod, "Creating CatalogSource")
    err = operators.CreateCatalogSource(ctx, kubeClient, operators.CatalogSourceConfig{
        Name:        constants.COOCatalogName,
        Namespace:   constants.MarketplaceNamespace,
        Image:       config.FBCURL,
        DisplayName: "COO FBC Catalog",
    })
    if err != nil {
        return fmt.Errorf("failed to create CatalogSource: %w", err)
    }
    
    // Wait for CatalogSource to be ready
    exec.SendLog(operations.StepDeployMethod, "Waiting for CatalogSource to be ready")
    err = operators.WaitForCatalogSource(ctx, kubeClient, constants.COOCatalogName, constants.MarketplaceNamespace)
    if err != nil {
        return fmt.Errorf("CatalogSource not ready: %w", err)
    }
    
    // Create Subscription
    exec.SendLog(operations.StepDeployMethod, "Creating Subscription")
    err = operators.CreateSubscription(ctx, kubeClient, operators.SubscriptionConfig{
        Name:             constants.COOOperatorName,
        Namespace:        config.Namespace,
        Channel:          constants.DefaultCOOChannel,
        PackageName:      constants.COOPackageName,
        CatalogSource:    constants.COOCatalogName,
        CatalogNamespace: constants.MarketplaceNamespace,
    })
    if err != nil {
        return fmt.Errorf("failed to create Subscription: %w", err)
    }
    
    exec.SendLog(operations.StepDeployMethod, "FBC deployment initiated successfully")
    return nil
}
```

#### Stage Method (`pkg/operators/coo/stage.go`)

**Purpose**: Deploy using stage registry FBC (includes brew registry IDMS)

```go
package coo

import (
    "context"
    "fmt"
    
    "sigs.k8s.io/controller-runtime/pkg/client"
    
    "github.com/observability-ui/development-tools/internal/constants"
    "github.com/observability-ui/development-tools/pkg/executor"
    "github.com/observability-ui/development-tools/pkg/operators"
    "github.com/observability-ui/development-tools/pkg/operations"
)

func DeployStage(ctx context.Context, kubeClient client.Client,
                config operations.DeployCOOConfig, exec *executor.Executor) error {
    
    // Create stage-specific IDMS (includes brew registry)
    exec.SendLog(operations.StepDeployMethod, "Creating ImageDigestMirrorSet for Stage (with brew registry)")
    err := operators.EnsureIDMSStageWithBrew(ctx, kubeClient)
    if err != nil {
        return fmt.Errorf("failed to create IDMS: %w", err)
    }
    
    // Create CatalogSource
    exec.SendLog(operations.StepDeployMethod, "Creating CatalogSource")
    err = operators.CreateCatalogSource(ctx, kubeClient, operators.CatalogSourceConfig{
        Name:        constants.COOCatalogName,
        Namespace:   constants.MarketplaceNamespace,
        Image:       config.FBCURL,
        DisplayName: "COO Stage Catalog",
    })
    if err != nil {
        return fmt.Errorf("failed to create CatalogSource: %w", err)
    }
    
    // Wait for CatalogSource to be ready
    exec.SendLog(operations.StepDeployMethod, "Waiting for CatalogSource to be ready")
    err = operators.WaitForCatalogSource(ctx, kubeClient, constants.COOCatalogName, constants.MarketplaceNamespace)
    if err != nil {
        return fmt.Errorf("CatalogSource not ready: %w", err)
    }
    
    // Create Subscription
    exec.SendLog(operations.StepDeployMethod, "Creating Subscription")
    err = operators.CreateSubscription(ctx, kubeClient, operators.SubscriptionConfig{
        Name:             constants.COOOperatorName,
        Namespace:        config.Namespace,
        Channel:          constants.DefaultCOOChannel,
        PackageName:      constants.COOPackageName,
        CatalogSource:    constants.COOCatalogName,
        CatalogNamespace: constants.MarketplaceNamespace,
    })
    if err != nil {
        return fmt.Errorf("failed to create Subscription: %w", err)
    }
    
    exec.SendLog(operations.StepDeployMethod, "Stage deployment initiated successfully")
    return nil
}
```

#### OperatorHub Method (`pkg/operators/coo/operatorhub.go`)

**Purpose**: Deploy from default OperatorHub catalog (simplest method)

```go
package coo

import (
    "context"
    "fmt"
    
    "sigs.k8s.io/controller-runtime/pkg/client"
    
    "github.com/observability-ui/development-tools/internal/constants"
    "github.com/observability-ui/development-tools/pkg/executor"
    "github.com/observability-ui/development-tools/pkg/operators"
    "github.com/observability-ui/development-tools/pkg/operations"
)

func DeployOperatorHub(ctx context.Context, kubeClient client.Client,
                      config operations.DeployCOOConfig, exec *executor.Executor) error {
    
    // Simplest method - just create subscription to default catalog
    exec.SendLog(operations.StepDeployMethod, "Creating Subscription to OperatorHub")
    
    err := operators.CreateSubscription(ctx, kubeClient, operators.SubscriptionConfig{
        Name:             constants.COOOperatorName,
        Namespace:        config.Namespace,
        Channel:          config.Channel,
        PackageName:      constants.COOPackageName,
        CatalogSource:    "redhat-operators",
        CatalogNamespace: "openshift-marketplace",
    })
    if err != nil {
        return fmt.Errorf("failed to create Subscription: %w", err)
    }
    
    exec.SendLog(operations.StepDeployMethod, "OperatorHub subscription created successfully")
    return nil
}
```

---

### 7. Supporting Infrastructure

#### Namespace Helper (`pkg/k8s/namespace.go` - NEW)

**Purpose**: Create namespace with custom labels

```go
package k8s

import (
    "context"
    
    corev1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/api/errors"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "sigs.k8s.io/controller-runtime/pkg/client"
)

// EnsureNamespaceWithLabels creates a namespace with specified labels if it doesn't exist
// Returns (true, nil) if created, (false, nil) if already exists, (false, err) on error
func EnsureNamespaceWithLabels(ctx context.Context, kubeClient client.Client, 
                               name string, labels map[string]string) (bool, error) {
    namespace := &corev1.Namespace{}
    key := client.ObjectKey{Name: name}
    
    err := kubeClient.Get(ctx, key, namespace)
    if err == nil {
        return false, nil // Already exists
    }
    
    if !errors.IsNotFound(err) {
        return false, err
    }
    
    namespace = &corev1.Namespace{
        ObjectMeta: metav1.ObjectMeta{
            Name:   name,
            Labels: labels,
        },
    }
    
    if err := kubeClient.Create(ctx, namespace); err != nil {
        return false, err
    }
    
    return true, nil
}
```

#### Enhanced IDMS Utilities (`pkg/operators/idms.go` - ENHANCE)

**Add new functions for COO-specific IDMS configurations:**

```go
// EnsureIDMSQuay creates IDMS for quay.io registry
// Mirrors: quay.io/redhat-user-workloads/cluster-observabilit-tenant/cluster-observability-operator
//       -> registry.redhat.io/cluster-observability-operator
func EnsureIDMSQuay(ctx context.Context, kubeClient client.Client) error {
    // Implementation using existing EnsureIDMS helper
}

// EnsureIDMSStage creates IDMS for stage registry
// Mirrors: registry.stage.redhat.io -> registry.redhat.io
func EnsureIDMSStage(ctx context.Context, kubeClient client.Client) error {
    // Implementation
}

// EnsureIDMSStageWithBrew creates IDMS for stage with brew registry
// Two mirrors:
// 1. registry.stage.redhat.io -> registry.redhat.io
// 2. brew.registry.redhat.io/rh-osbs/iib -> registry-proxy.engineering.redhat.com/rh-osbs/iib
func EnsureIDMSStageWithBrew(ctx context.Context, kubeClient client.Client) error {
    // Implementation with multiple mirrors in single IDMS
}
```

---

## Deferred Items

### Scheduler Patching

**Decision**: Removed from initial COO deployment implementation

**Bash Script Behavior**:
```bash
kubectl patch Scheduler cluster --type='json' -p '[{ "op": "replace", "path": "/spec/mastersSchedulable", "value": true }]'
```

**Reason for Deferral**:
- Unclear if this should be:
  - Part of COO deployment specifically
  - Global cluster preparation step (separate command?)
  - Optional via flag
  - Part of all operator deployments
- Needs architectural decision

**Action Required**:
- Add task to TASKS.md under new "Deferred Items" section
- Determine correct placement
- Implement in appropriate location

**Note**: This will be documented in TASKS.md update

---

## Error Handling

### Validation Errors

| Scenario | Error Message | Action |
|----------|---------------|--------|
| Missing method | "--method is required" | Exit, show help |
| Unknown method | "unknown method: X (must be: bundle, fbc, stage, operatorhub)" | Exit |
| Bundle without URL | "--bundle-url is required for bundle method" | Exit |
| FBC without URL | "--fbc-url is required for fbc/stage methods" | Exit |
| operator-sdk not found | "operator-sdk binary not found in PATH. Install from: ..." | Exit |

### Runtime Errors

| Scenario | Handling | User Impact |
|----------|----------|-------------|
| Namespace create fails | Fail at Step 1, return error | Command fails, no changes |
| OperatorGroup create fails | Fail at Step 2, return error | Namespace exists, no operator |
| IDMS create fails | Fail in method, return error | Partial state (namespace + OG) |
| CatalogSource create fails | Fail in method, return error | Partial state |
| Subscription create fails | Fail in method, return error | CatalogSource exists, no subscription |
| CSV never ready | Timeout at Step 4, return error | Subscription exists, CSV failed/pending |
| operator-sdk command fails | Show output, return error | Bundle not deployed |

### Edge Cases

**Namespace already exists:**
- Log: "Namespace already exists"
- Continue (idempotent)
- Success

**OperatorGroup already exists:**
- Log: "OperatorGroup already exists"
- Continue (idempotent)
- Success

**IDMS already exists:**
- IDMS creation should be idempotent in utilities
- Skip if exists
- Success

**CatalogSource already exists:**
- Error (user should clean up first)
- Suggest: `oc delete catalogsource <name> -n openshift-marketplace`

**Subscription already exists:**
- Error (user should clean up first)
- Suggest: `obstool cleanup coo` or `oc delete subscription <name>`

---

## Testing Strategy

### Manual Testing Sequence

**Prerequisites**:
- operator-sdk installed (for bundle method)
- Access to test OCP cluster
- Valid bundle/FBC URLs

**Test Cases**:

#### 1. Deploy Command Group
- [ ] `obstool deploy` shows TUI with COO option
- [ ] Selecting COO executes COO deployment (TUI mode)
- [ ] `obstool deploy coo` bypasses parent TUI

#### 2. Deploy COO - OperatorHub (Simplest)
- [ ] CLI: `obstool deploy coo --method=operatorhub`
- [ ] TUI: `obstool deploy coo` → select operatorhub → enter channel
- [ ] Namespace created with `openshift.io/cluster-monitoring: "true"` label
- [ ] OperatorGroup created
- [ ] Subscription created to redhat-operators catalog
- [ ] CSV becomes Succeeded within timeout
- [ ] COO operator pod running

#### 3. Deploy COO - FBC
- [ ] CLI: `obstool deploy coo --method=fbc --fbc-url=<url> --registry-type=quay`
- [ ] TUI: Form collects FBC URL and registry type
- [ ] IDMS created for quay
- [ ] CatalogSource created
- [ ] CatalogSource pod running and READY
- [ ] Subscription created
- [ ] CSV becomes Succeeded
- [ ] COO operator pod running

#### 4. Deploy COO - Stage
- [ ] CLI: `obstool deploy coo --method=stage --fbc-url=<stage-url>`
- [ ] Stage IDMS created (verify both mirrors: stage + brew)
- [ ] CatalogSource created
- [ ] Rest same as FBC
- [ ] CSV becomes Succeeded

#### 5. Deploy COO - Bundle
- [ ] operator-sdk detected (or error shown)
- [ ] CLI: `obstool deploy coo --method=bundle --bundle-url=<url> --registry-type=quay`
- [ ] OCP version detected correctly
- [ ] IDMS created for quay
- [ ] operator-sdk command executed
- [ ] On OCP 4.19+: `--security-context-config restricted` argument present
- [ ] On OCP <4.19: No security context argument
- [ ] CSV becomes Succeeded

#### 6. Error Cases
- [ ] operator-sdk not found → Clear error with install instructions
- [ ] Invalid bundle URL → Error from operator-sdk
- [ ] Invalid FBC URL → CatalogSource fails to pull
- [ ] CSV timeout → Shows timeout error with troubleshooting
- [ ] Namespace already exists → Idempotent, succeeds
- [ ] OperatorGroup already exists → Idempotent, succeeds

#### 7. Progress Tracking
- [ ] CLI mode: Logs show all 4 steps
- [ ] CLI mode: Sub-logs show method-specific details (IDMS, CatalogSource, etc.)
- [ ] TUI mode: Progress bar updates for each step
- [ ] TUI mode: CSV countdown visible during wait
- [ ] TUI mode: Success message when complete
- [ ] TUI mode: Error message on failure

#### 8. Idempotency
- [ ] Run same command twice
- [ ] Second run: Namespace/OperatorGroup already exist
- [ ] Second run: Fails gracefully if subscription exists (expected)

---

## Dependencies

### Go Modules (New)
- None (all required modules already present)

### Go Modules (Existing)
- `github.com/operator-framework/api` ✅ (Subscription, OperatorGroup, CSV)
- `github.com/openshift/api/config/v1` ✅ (IDMS)
- `github.com/spf13/cobra` ✅
- `github.com/charmbracelet/bubbletea` ✅
- `github.com/charmbracelet/huh` ✅
- `sigs.k8s.io/controller-runtime/pkg/client` ✅

### External Binary Dependencies
- `operator-sdk` - **Required for bundle method only**
  - Check with: `exec.LookPath("operator-sdk")`
  - Error message includes installation link
  - Other methods work without it

### Blocked By
- Nothing! All foundation complete:
  - ✅ Executor pattern
  - ✅ OLM utilities (subscription, catalogsource, operatorgroup, IDMS, CSV waiting)
  - ✅ Mode detection
  - ✅ TUI framework
  - ✅ Output handling

---

## Implementation Checklist

### Phase 1: Deploy Command Group
- [ ] Create `cmd/deploy/deploy.go`
  - [ ] Define parent command
  - [ ] Implement `runDeployTUI()` with component selection
  - [ ] Initially show only COO option
- [ ] Register in `cmd/root.go`
- [ ] Test: `obstool deploy` shows TUI
- [ ] Test: `obstool deploy coo` delegates to coo subcommand

### Phase 2: Constants & Helpers
- [ ] Create `internal/constants/coo.go`
  - [ ] Define namespace, operator name, catalog name, etc.
- [ ] Create `pkg/k8s/namespace.go`
  - [ ] Implement `EnsureNamespaceWithLabels()`
  - [ ] Test namespace creation with labels
- [ ] Enhance `pkg/operators/idms.go`
  - [ ] Implement `EnsureIDMSQuay()`
  - [ ] Implement `EnsureIDMSStage()`
  - [ ] Implement `EnsureIDMSStageWithBrew()`
  - [ ] Test IDMS creation

### Phase 3: Method Implementations (Parallel)
- [ ] Create `pkg/operators/coo/` directory
- [ ] Implement `pkg/operators/coo/operatorhub.go` (simplest)
  - [ ] Test standalone
- [ ] Implement `pkg/operators/coo/fbc.go`
  - [ ] Test standalone
- [ ] Implement `pkg/operators/coo/stage.go`
  - [ ] Test standalone
- [ ] Implement `pkg/operators/coo/bundle.go`
  - [ ] Check for operator-sdk binary
  - [ ] Build command with version detection
  - [ ] Test standalone

### Phase 4: Business Logic
- [ ] Create `pkg/operations/coo.go`
  - [ ] Define step constants
  - [ ] Implement `DeployCOO()` orchestration
  - [ ] Integrate executor pattern
  - [ ] Test progress updates

### Phase 5: Command Implementation
- [ ] Create `cmd/deploy/coo.go`
  - [ ] Define flags
  - [ ] Implement mode detection
  - [ ] Implement `runDeployCOOCLI()`
  - [ ] Implement `runDeployCOOTUI()`
  - [ ] Implement `collectCOOInput()` forms
  - [ ] Implement `getConfigFromFlags()` validation
- [ ] Add to `cmd/deploy/deploy.go`
- [ ] Test all 4 methods in CLI mode
- [ ] Test all 4 methods in TUI mode

### Phase 6: Documentation & Polish
- [ ] Update `tmp/TASKS.md`
  - [ ] Mark deploy command group as complete
  - [ ] Mark deploy coo as complete
  - [ ] Add deferred scheduler patching task
- [ ] Verify help text: `obstool deploy --help`
- [ ] Verify help text: `obstool deploy coo --help`
- [ ] Update examples in help text

### Phase 7: Integration Testing
- [ ] Test on real OCP cluster (all 4 methods)
- [ ] Test error cases
- [ ] Test progress tracking (CLI + TUI)
- [ ] Test idempotency
- [ ] Verify CSV waits with countdown

---

## Success Criteria

### Deploy Command Group ✅
- [ ] Shows TUI component selection when run without subcommand
- [ ] Only COO shown initially (design for future expansion)
- [ ] Delegates to subcommands correctly
- [ ] No errors or crashes

### Deploy COO ✅

**Functional:**
- [ ] All 4 methods work (bundle, fbc, stage, operatorhub)
- [ ] Both CLI and TUI modes functional
- [ ] Namespace created with monitoring label
- [ ] OperatorGroup created
- [ ] Method-specific resources created correctly
- [ ] CSV reaches Succeeded state
- [ ] Idempotent for namespace/OperatorGroup

**Quality:**
- [ ] Follows executor pattern for progress tracking
- [ ] Proper error handling with descriptive messages
- [ ] No 1-2 letter variable names (except err, ctx, ok)
- [ ] Minimal comments (code self-documenting)
- [ ] Consistent with existing command patterns
- [ ] 100% feature parity with bash scripts (except scheduler - deferred)

**User Experience:**
- [ ] Clear progress in both CLI and TUI
- [ ] CSV countdown visible in TUI
- [ ] Helpful error messages (operator-sdk not found, invalid URLs, etc.)
- [ ] Command help text complete and clear

---

## TASKS.md Updates Required

### New Section: Deferred Items

```markdown
## Deferred Items

### Scheduler Patching
- **Decision**: Removed from initial COO deployment implementation
- **Reason**: Unclear if this should be:
  - Part of COO deployment
  - Global cluster preparation step
  - Optional via flag
- **Action Required**: Determine correct placement and implement
- **Reference**: Bash scripts patch with `kubectl patch Scheduler cluster --type='json' -p '[{ "op": "replace", "path": "/spec/mastersSchedulable", "value": true }]'`
- **Affects**: COO bundle deployment (and possibly others)
```

### New Section: Build & Release

```markdown
### Binary Dependencies

- [ ] **Implement binary dependency checker**
  - [ ] Create `cmd/doctor.go` command
  - [ ] Check for operator-sdk (required for bundle method)
  - [ ] Check for oc/kubectl (general requirement)
  - [ ] Provide installation instructions
  - [ ] Blocked by: None (can implement anytime)
  - [ ] Note: For current implementation, assume operator-sdk installed and error if not found
```

### Update Deploy Tasks

```markdown
### Deploy Command Group
- [~] **Implement deploy command group**
  - [~] Create `cmd/deploy/deploy.go`
  - [~] TUI shows COO option only (others add themselves when implemented)
  - Blocked by: Implement root command ✅

### Deploy COO
- [~] **Implement deploy coo command**
  - [~] Create `cmd/deploy/coo.go`
  - [~] Add `--method` flag (bundle, fbc, stage, operatorhub)
  - [~] Add method-specific flags
  - [~] Bundle method uses operator-sdk via exec.Command
  - [~] Scheduler patching DEFERRED (see Deferred Items)
  - [~] Perses namespace NOT created (left to dashboards)
  - [~] Blocked by: Implement deploy command group

  - [~] **Implement COO operatorhub deployment**
    - [~] Create `pkg/operators/coo/operatorhub.go`
    - [~] Create Subscription to default catalog
    - [~] Blocked by: Implement deploy coo command

  - [~] **Implement COO FBC deployment**
    - [~] Create `pkg/operators/coo/fbc.go`
    - [~] Create IDMS for quay
    - [~] Create CatalogSource
    - [~] Create Subscription
    - [~] Blocked by: Implement deploy coo command

  - [~] **Implement COO stage deployment**
    - [~] Create `pkg/operators/coo/stage.go`
    - [~] Create stage IDMS with brew registry
    - [~] Create CatalogSource
    - [~] Create Subscription
    - [~] Blocked by: Implement deploy coo command

  - [~] **Implement COO bundle deployment**
    - [~] Create `pkg/operators/coo/bundle.go`
    - [~] Check for operator-sdk binary
    - [~] Create IDMS based on registry type
    - [~] Execute operator-sdk run bundle
    - [~] Version-aware security context (OCP 4.19+)
    - [~] Blocked by: Implement deploy coo command
```

---

## Timeline Estimate

**Total**: ~8-12 hours development + 2-3 hours testing

- Phase 1 (Deploy Command Group): 1-2 hours
- Phase 2 (Constants & Helpers): 1-2 hours
- Phase 3 (Method Implementations): 3-4 hours
- Phase 4 (Business Logic): 1-2 hours
- Phase 5 (Command Implementation): 2-3 hours
- Phase 6 (Documentation): 30 minutes
- Phase 7 (Integration Testing): 2-3 hours

**Note**: Timeline not a constraint - focus on quality and correctness.

---

## Post-Implementation

**After completion:**
1. Update TASKS.md as specified above
2. Create `tmp/tasks/deploy-coo/implementation.md` documenting actual implementation
3. Note any deviations from plan or issues encountered
4. Document lessons learned for future deploy command implementations

---

**Ready for Implementation**: YES ✅  
**Approval Required**: User confirmation on operator-sdk approach ✅ (confirmed: exec.Command acceptable)

---

**Questions Resolved:**
1. ✅ operator-sdk dependency - Use exec.Command, error if not found
2. ✅ Scheduler patching - DEFERRED to separate task
3. ✅ Perses namespace - NOT created by COO (dashboards will create)
4. ✅ Implementation approach - Both tasks in parallel
5. ✅ Binary install task - Added to TASKS.md, defer implementation
