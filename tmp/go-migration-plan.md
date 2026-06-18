# Development Tools Repository Go Migration Plan

> **🤖 AI Agents**: This document contains detailed architecture, patterns, and code examples. For quick context, see [CONTEXT.md](./CONTEXT.md) first.

**Repository**: development-tools (OpenShift Observability UI Team)  
**Date**: June 8, 2026 (Updated: June 18, 2026)  
**Status**: Implementation in progress - Foundation complete, commands being added  
**Scope**: Convert from multi-technology stack (bash, make, just, js, yaml) to Go-based CLI  
**Purpose**: Detailed architecture reference and implementation patterns

> **📌 IMPLEMENTATION UPDATE (June 18, 2026)**: Foundation and TUI framework complete. Business logic now uses channel-based decoupling pattern. See `tmp/tasks/business-logic-decoupling/` for details on the implemented architecture.

---

## Quick Navigation for Agents

**Looking for...**
- Architecture diagram → [Section: Target Architecture (Proposed)](#target-architecture-proposed)
- Execution Context pattern → [Section: Execution Context Pattern](#execution-context-pattern)
- Directory structure → [Section: Target Architecture](#target-architecture-proposed)
- CLI examples → [Section: User Interaction](#decision-points-for-team-consideration)
- Version detection → [Section: Decision 3 - Version Detection](#decision-points-for-team-consideration)
- Mode detection (CLI vs TUI) → [Section: Decision 2 - User Interaction](#decision-points-for-team-consideration)
- Resource structure → [Key Architecture Changes](#key-architecture-changes)
- COO deployment methods → Architecture diagram, `deploy/coo.go` section
- Testing approach → [Section: Decision 6 - Testing Strategy](#decision-points-for-team-consideration)

---

## Executive Summary

The development-tools repository has grown organically to serve the team's needs but now suffers from:
- **Technology Fragmentation**: 6+ different technologies (Bash, Make, Just, JS, YAML, Kustomize)
- **Duplicate Actions**: Multiple ways to accomplish the same tasks
- **Maintainability Issues**: Difficult to extend and maintain
- **Lack of Type Safety**: No compile-time guarantees
- **Limited Error Handling**: Inconsistent error propagation

### Recommended Approach

**Convert to a single Go-based CLI tool** that:
1. Consolidates all operations into one binary
2. Provides type-safe interactions with Kubernetes resources
3. Leverages Go modules for CRD definitions
4. Enables better testing and error handling
5. Maintains backward compatibility where possible

---

## Current State Analysis

### Repository Statistics
- **Total Shell Scripts**: 58
- **YAML Manifests**: 78
- **Justfiles**: 2
- **Makefiles**: 2
- **TypeScript Files**: 3 (colorize utility)
- **Go Files**: 0 (currently none)
- **Major Functional Areas**: 15

### Technology Stack (Current)
```
Bash (73%)  ████████████████████████████████
YAML (17%)  ████████
Just (5%)   ██
Make (3%)   █
TS (2%)     ▌
```

### Primary Operations Performed

| Operation Type | Count | Technologies Used |
|---------------|-------|-------------------|
| Resource Creation (oc apply/create) | 260+ | Bash, YAML heredocs |
| Resource Deletion | 30+ | Bash |
| Version Detection | 20+ | Bash, jq, bc |
| Interactive Prompts | 40+ | Bash read |
| Wait/Polling | 15+ | Bash loops, sleep |
| Image Building | 5+ | Bash, operator-sdk, podman |
| RBAC Management | 100+ | YAML, Bash |
| User Management | 10+ | Bash, htpasswd |

---

## Target State: Go-Based Architecture

### CLI Name: `obstool`

**Decision**: Tool will be named `obstool` (observability tooling)

### Architecture Overview

```
obstool (main binary)
├── cmd/
│   ├── root.go                 # Root command
│   ├── version.go             # Version command
│   ├── deploy/
│   │   ├── deploy.go          # Deploy command group (TUI selection when no subcommand)
│   │   ├── all.go             # Deploy all components (flag mode only)
│   │   ├── coo.go             # Deploy COO (4 methods: bundle, fbc, stage, operatorhub)
│   │   ├── logging.go         # Deploy logging stack
│   │   ├── tracing.go         # Deploy tracing stack
│   │   ├── dashboards.go      # Deploy dashboards
│   │   ├── monitoring.go      # Deploy monitoring plugin
│   │   ├── acm.go             # Deploy ACM observability
│   │   └── korrel8r.go        # Deploy korrel8r
│   ├── users/
│   │   ├── users.go           # User management group
│   │   ├── create.go          # Create test users
│   │   └── rbac.go            # Apply RBAC
│   ├── update/
│   │   ├── update.go          # Update command group
│   │   ├── monitoring.go      # Scale up monitoring components
│   │   └── coo.go             # Update COO in place
│   └── cleanup/
│       ├── cleanup.go         # Cleanup command group
│       ├── monitoring.go      # Scale down monitoring components
│       ├── coo.go             # Cleanup COO
│       ├── logging.go         # Cleanup logging stack
│       ├── tracing.go         # Cleanup tracing stack
│       ├── acm.go             # Cleanup ACM
│       └── all.go             # Cleanup all components (flag mode only)
├── pkg/
│   ├── k8s/
│   │   ├── client.go          # Kubernetes client wrapper (controller-runtime)
│   │   ├── connection.go      # Connection management (kubeconfig only)
│   │   └── version.go         # Version detection
│   ├── context/
│   │   └── context.go         # Execution context (mode, client, version, etc.)
│   ├── resources/
│   │   ├── uiplugin.go        # UIPlugin CRs (using COO's abstraction)
│   │   ├── lokistack.go       # LokiStack resources
│   │   ├── clusterlogforwarder.go  # ClusterLogForwarder resources
│   │   ├── tempostack.go      # TempoStack resources
│   │   ├── otel.go            # OpenTelemetry resources
│   │   ├── perses.go          # Perses datasource and global resources
│   │   ├── rbac.go            # RBAC resources
│   │   ├── minio.go           # MinIO deployment resources
│   │   └── dashboards/        # Dashboard definitions (1 file per dashboard)
│   │       ├── node-exporter.go
│   │       ├── prometheus.go
│   │       ├── thanos.go
│   │       └── ... (30+ dashboard files)
│   ├── operators/
│   │   ├── olm.go             # OLM operations
│   │   ├── subscription.go    # Subscription management
│   │   ├── coo/               # COO-specific deployment
│   │   │   ├── bundle.go      # Bundle deployment
│   │   │   ├── fbc.go         # FBC deployment
│   │   │   ├── stage.go       # Stage registry deployment
│   │   │   ├── operatorhub.go # OperatorHub deployment
│   │   │   └── update.go      # In-place update
│   │   └── bundle.go          # Generic bundle operations
│   ├── users/
│   │   ├── htpasswd.go        # htpasswd management
│   │   └── oauth.go           # OAuth configuration
│   ├── storage/
│   │   ├── provider.go        # Storage provider interface
│   │   └── minio.go           # MinIO implementation (to be deprecated)
│   ├── tui/
│   │   ├── deploy.go          # Deploy selection TUI (Bubble Tea)
│   │   ├── progress.go        # Progress display
│   │   ├── models.go          # TUI models
│   │   └── styles.go          # Styling
│   └── config/
│       └── config.go          # Configuration as Go struct (type-safe)
└── internal/
    └── version/
        ├── version.go         # Version detection logic
        └── compare.go         # Version comparison
```

**Key Architecture Changes**:
1. **No `qe/` directory** - COO deployment methods consolidated in `deploy/coo.go` and `operators/coo/`
2. **No `demo/` directory** - Demo workflows handled by combining existing commands
3. **Added `update/` directory** - For COO in-place updates and scaling up components
4. **Cleanup mirrors deploy** - Same structure with cleanup commands
5. **TUI package** - Using Bubble Tea for interactive mode
6. **No templates directory** - All resources defined as Go structs
7. **Storage provider interface** - For future MinIO deprecation
8. **Dashboards subdirectory** - Large number of dashboard files (1 per dashboard)
9. **Context package** - Execution context with mode, client, version to avoid passing isTUI everywhere
10. **Flat resources structure** - All resource files at top level except dashboards/
11. **Cleanup all flag-only** - `obstool cleanup all` only available with full flags

### Execution Context Pattern

To avoid passing `isTUI` and other state everywhere, all operations use a shared context:

```go
package context

import (
    "context"
    "sigs.k8s.io/controller-runtime/pkg/client"
)

// ExecutionContext holds shared state for command execution
type ExecutionContext struct {
    Context context.Context  // Go context for cancellation
    Client  client.Client    // Kubernetes client
    Version *VersionInfo     // Cluster version info
    IsTUI   bool            // Running in TUI mode vs CLI mode
}

// NewExecutionContext creates a new execution context
func NewExecutionContext(ctx context.Context, client client.Client, version *VersionInfo, isTUI bool) *ExecutionContext {
    return &ExecutionContext{
        Context: ctx,
        Client:  client,
        Version: version,
        IsTUI:   isTUI,
    }
}

// Usage in commands:
func deployLogging(execCtx *ExecutionContext, cfg LoggingConfig) error {
    if execCtx.IsTUI {
        // TUI mode behavior
        return deployLoggingTUI(execCtx, cfg)
    }
    // CLI mode behavior
    return deployLoggingCLI(execCtx, cfg)
}
```

---

## Code Style Standards

**Comments**:
- Minimal to none - prefer self-documenting code
- Only add comments for non-obvious business logic or complex algorithms
- No package/function documentation comments unless exported and truly necessary
- Exception: Complex algorithms or non-obvious business rules may warrant explanation

**Variable Naming**:
- No 1-2 letter variable names except standard Go idioms
- ✅ Acceptable short names: `err`, `ctx`, `ok`
- ❌ Avoid: `c`, `e`, `i`, `j`, `k`, `x`, `y`, `s`, `r`
- Use descriptive names: `client`, `config`, `namespace`, `index`, `count`, `result`

**Example - Good**:
```go
func NewClient(ctx context.Context, kubeconfigPath string) (*Client, error) {
    config, err := getKubeConfig(kubeconfigPath)
    if err != nil {
        return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
    }
    return &Client{config: config}, nil
}
```

**Example - Avoid**:
```go
// NewClient creates a new client (unnecessary comment)
func NewClient(c context.Context, p string) (*Client, error) {
    // Get config (obvious from code)
    cfg, e := getKubeConfig(p)
    if e != nil {
        return nil, fmt.Errorf("failed to load kubeconfig: %w", e)
    }
    return &Client{config: cfg}, nil
}
```

---

## Decision Points for Team Consideration

### 1. CLI Framework Choice

**Decision**: **Cobra**

**Rationale**: Industry standard for Kubernetes tooling (kubectl, oc, helm), excellent documentation, and built-in shell completion support.

**Example Cobra Command Structure**:
```go
// cmd/deploy/logging.go
var loggingCmd = &cobra.Command{
    Use:   "logging",
    Short: "Deploy the logging stack",
    Long:  "Deploys LokiStack, ClusterLogForwarder, and logging UIPlugin",
    RunE: func(cmd *cobra.Command, args []string) error {
        return deployLogging(cmd.Context())
    },
}

func init() {
    deployCmd.AddCommand(loggingCmd)
    loggingCmd.Flags().String("data-model", "", "Data model: otel or viaq")
    loggingCmd.Flags().Bool("skip-ui-plugin", false, "Skip UIPlugin deployment")
}
```

---

### 2. User Interaction Approach

**Decision**: **Flags + TUI Hybrid**

**Implementation**:
- **With all required flags**: Run in CLI mode (non-interactive)
- **Missing flags**: Launch TUI mode using Bubble Tea
- **Non-interactive environment**: Fail early with clear error messages

**Rationale**: Provides automation-friendly CLI while offering intuitive TUI for interactive use. No prompts - either full flag support or full TUI.

**Example Implementation**:
```go
import (
    tea "github.com/charmbracelet/bubbletea"
    "os"
)

func deployLogging(cmd *cobra.Command, args []string) error {
    // Check if all required flags are provided
    if hasAllRequiredFlags(cmd) {
        // CLI mode - direct execution
        return executeDeployLogging(cmd.Context(), getFlagsConfig(cmd))
    }
    
    // Check if we're in a non-interactive environment
    if !isTerminal() {
        return fmt.Errorf("missing required flags; run with --help to see required flags")
    }
    
    // TUI mode - interactive selection
    model := NewDeployLoggingModel()
    p := tea.NewProgram(model)
    finalModel, err := p.Run()
    if err != nil {
        return err
    }
    
    // Execute with TUI-selected configuration
    return executeDeployLogging(cmd.Context(), finalModel.(DeployLoggingModel).Config)
}

func isTerminal() bool {
    return term.IsTerminal(int(os.Stdin.Fd()))
}
```

**TUI Selection Example** (for `obstool deploy` with no subcommand):
```go
type DeploySelectionModel struct {
    choices  []string
    selected map[int]bool
    cursor   int
}

func (m DeploySelectionModel) View() string {
    s := "Select components to deploy (space to select, enter to confirm):\n\n"
    
    for i, choice := range m.choices {
        cursor := " "
        if m.cursor == i {
            cursor = ">"
        }
        
        checked := " "
        if m.selected[i] {
            checked = "x"
        }
        
        s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
    }
    
    s += "\nPress q to quit.\n"
    return s
}
```

**Special Case - Deploy All**:
- `obstool deploy all --flag1=val1 --flag2=val2`: CLI mode, deploy everything
- `obstool deploy`: TUI mode with multi-select including "Select All" option

**Optional TUI Enhancement** (for monitoring):
```go
import tea "github.com/charmbracelet/bubbletea"

type deployModel struct {
    operations []Operation
    current    int
    err        error
}

func (m deployModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Handle deployment progress updates
}

func (m deployModel) View() string {
    s := "Deploying Observability Stack\n\n"
    for i, op := range m.operations {
        checkbox := " "
        if op.Done {
            checkbox = "✓"
        } else if i == m.current {
            checkbox = "⋯"
        }
        s += fmt.Sprintf("[%s] %s\n", checkbox, op.Name)
    }
    return s
}
```

---

### 3. Kubernetes Client Strategy

**Decision**: **controller-runtime/pkg/client directly (no abstraction)**

**Rationale**: 
- Simpler API than client-go
- Same library used by operators (team familiarity)
- Can rewrite later if needed (no premature abstraction)
- Reduces complexity for contributors

**Example Client Setup**:
```go
package k8s

import (
    "context"
    "time"
    "os"
    "path/filepath"
    "k8s.io/client-go/rest"
    "k8s.io/client-go/tools/clientcmd"
    "sigs.k8s.io/controller-runtime/pkg/client"
)

type Client struct {
    client.Client
    config  *rest.Config
}

func NewClient(ctx context.Context, kubeconfigPath string) (*Client, error) {
    // Only kubeconfig file authentication (no in-cluster config)
    config, err := getKubeConfig(kubeconfigPath)
    if err != nil {
        return nil, err
    }
    
    // Configure client timeouts and rate limits
    config.Timeout = 30 * time.Second  // API request timeout
    config.QPS = 50                    // Queries per second (rate limit)
    config.Burst = 100                 // Burst allowance (max concurrent)
    
    scheme := runtime.NewScheme()
    if err := registerSchemes(scheme); err != nil {
        return nil, err
    }
    
    c, err := client.New(config, client.Options{Scheme: scheme})
    if err != nil {
        return nil, err
    }
    
    return &Client{
        Client: c,
        config: config,
    }, nil
}

func getKubeConfig(kubeconfigPath string) (*rest.Config, error) {
    // Priority: explicit flag > KUBECONFIG env > ~/.kube/config
    if kubeconfigPath != "" {
        return clientcmd.BuildConfigFromFlags("", kubeconfigPath)
    }
    
    if kubeconfigEnv := os.Getenv("KUBECONFIG"); kubeconfigEnv != "" {
        return clientcmd.BuildConfigFromFlags("", kubeconfigEnv)
    }
    
    home, _ := os.UserHomeDir()
    defaultPath := filepath.Join(home, ".kube", "config")
    return clientcmd.BuildConfigFromFlags("", defaultPath)
}
```

---

### 4. Version Detection & Conditional Logic

**Decision**: **Centralized version detection**

**Rationale**: Based on observability-operator PR #1100 pattern. Version-specific logic is important for UIPlugin/ConsolePlugin compatibility across OpenShift versions.

**Note**: The ConsolePlugin version-specific code shown in PR #1100 is an example of issues to watch for. `obstool` will use UIPlugin from COO which provides abstraction over ConsolePlugin, so we won't need to handle those version differences directly.

```go
package version

import (
    "context"
    "golang.org/x/mod/semver"
    configv1 "github.com/openshift/api/config/v1"
    "sigs.k8s.io/controller-runtime/pkg/client"
)

type VersionInfo struct {
    OpenShiftVersion string
    KubernetesVersion string
}

func Detect(ctx context.Context, c client.Client) (*VersionInfo, error) {
    cv := &configv1.ClusterVersion{}
    key := client.ObjectKey{Name: "version"}
    if err := c.Get(ctx, key, cv); err != nil {
        return nil, err
    }
    
    return &VersionInfo{
        OpenShiftVersion: cv.Status.Desired.Version,
    }, nil
}

func (v *VersionInfo) IsOCP419OrNewer() bool {
    return compareVersion(v.OpenShiftVersion, "4.19") >= 0
}

func (v *VersionInfo) IsOCP417To418() bool {
    return compareVersion(v.OpenShiftVersion, "4.17") >= 0 &&
           compareVersion(v.OpenShiftVersion, "4.19") < 0
}

func compareVersion(current, target string) int {
    if !strings.HasPrefix(current, "v") {
        current = "v" + current
    }
    if !strings.HasPrefix(target, "v") {
        target = "v" + target
    }
    return semver.Compare(current, target)
}
```

**Usage in Code** (version-aware resource creation):
```go
func (c *Client) CreateUIPlugin(ctx context.Context, plugin *UIPluginConfig) error {
    // UIPlugin from COO handles ConsolePlugin version differences internally
    // We just need to be aware of OCP version for other resource decisions
    
    uiPlugin := &uiv1alpha1.UIPlugin{
        ObjectMeta: metav1.ObjectMeta{
            Name: plugin.Name,
        },
        Spec: plugin.Spec,
    }
    
    return c.Create(ctx, uiPlugin)
}
```

---

### 5. Configuration Management

**Decision**: **Type-safe Go struct for configuration**

**Rationale**: Maintain type safety throughout the codebase. Avoid runtime configuration errors.

**Example Implementation**:
```go
package config

// Config is a type-safe configuration struct
// Exposed as a package-level variable for easy access
var Default = Config{
    DefaultBundle:    "quay.io/openshift-observability-ui/observability-ui-operator-bundle:latest",
    DefaultNamespace: "openshift-cluster-observability-operator",
    Kubeconfig:       "", // Will use standard kubeconfig resolution
    Registry: RegistryConfig{
        Default:  "quay",
        QuayOrg:  "openshift-observability-ui",
        StageURL: "registry.stage.redhat.io",
    },
    Demo: DemoConfig{
        UsersCount:      6,
        DefaultPassword: "password",
    },
}

type Config struct {
    DefaultBundle    string
    DefaultNamespace string
    Kubeconfig       string
    Registry         RegistryConfig
    Demo             DemoConfig
}

type RegistryConfig struct {
    Default  string
    QuayOrg  string
    StageURL string
}

type DemoConfig struct {
    UsersCount      int
    DefaultPassword string
}

// Usage in code
func getDefaultBundle() string {
    return config.Default.DefaultBundle
}

// Allow runtime override if needed
func SetDefaultBundle(bundle string) {
    config.Default.DefaultBundle = bundle
}
```

**Benefits**:
- Type safety at compile time
- IDE autocomplete
- Easy to test
- No runtime config parsing errors
- Can still be overridden programmatically if needed

---

### 6. Error Handling Strategy

**Decision**: **Mode-aware error handling**

**Rationale**: Error display and bubbling should differ based on whether running in TUI or flag mode.

```go
package errors

import (
    "fmt"
)

type ObstoolError struct {
    Op      string
    Err     error
    Context map[string]string
}

func (e *ObstoolError) Error() string {
    return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

func Wrap(op string, err error) error {
    return &ObstoolError{
        Op:  op,
        Err: err,
    }
}

// Mode-aware error handling
func HandleError(err error, isTUIMode bool) {
    if isTUIMode {
        // In TUI mode, display error in the TUI
        // Don't immediately exit, allow user to see error and decide
        displayErrorInTUI(err)
    } else {
        // In CLI mode, print to stderr and exit
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}

// Usage in commands:
func deployLogging(cmd *cobra.Command, args []string) error {
    isTUI := !hasAllRequiredFlags(cmd)
    
    if err := executeDeployment(); err != nil {
        HandleError(Wrap("deploy logging", err), isTUI)
        return err
    }
    return nil
}
```

---

### 7. Testing Strategy

**Decision**: **Minimal testing - unit tests only for critical code**

**Rationale**: This is a development tooling repository for the team. Low barrier to contribution is prioritized. Everyone should be able to debug and fix issues directly.

**Approach**:
- **No automated CI/CD testing** (no GitHub Actions)
- **No E2E tests** - too brittle and difficult to maintain
- **Unit tests only for critical logic**:
  - Version comparison functions
  - Resource struct construction
  - Client abstraction layer
- **Manual testing** is expected and acceptable

**Example Unit Test** (for critical version logic):
```go
func TestVersionComparison(t *testing.T) {
    tests := []struct {
        name     string
        current  string
        target   string
        expected int
    }{
        {"newer", "4.19.0", "4.17.0", 1},
        {"older", "4.17.0", "4.19.0", -1},
        {"equal", "4.18.0", "4.18.0", 0},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := compareVersion(tt.current, tt.target)
            if result != tt.expected {
                t.Errorf("compareVersion(%s, %s) = %d, want %d", 
                    tt.current, tt.target, result, tt.expected)
            }
        })
    }
}
```

**Testing Philosophy**:
- If it's critical enough to need testing, write a unit test
- Everything else: manual verification
- Keep tests simple and maintainable
- Contributors should be empowered to fix, not blocked by test failures

---

### 8. Waiting & Polling Strategy

**Decision**: **Mode-aware waiting**

**Rationale**: Waiting behavior should differ based on TUI vs CLI mode. No separate wait utilities needed.

**Implementation**:
- **TUI Mode**: Show real-time progress updates in the TUI
- **CLI Mode**: Silent waiting with optional progress dots/spinner
- **No shared wait.go utility** - handled inline based on mode

```go
package operators

import (
    "context"
    "time"
    tea "github.com/charmbracelet/bubbletea"
)

// TUI Mode - waiting shows progress
type subscriptionWaitModel struct {
    name      string
    namespace string
    status    string
    done      bool
}

func (m subscriptionWaitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case subscriptionStatusMsg:
        m.status = msg.status
        if msg.ready {
            m.done = true
            return m, tea.Quit
        }
        return m, checkSubscriptionStatus(m.name, m.namespace)
    }
    return m, nil
}

// CLI Mode - silent polling
func waitForSubscriptionReadyCLI(ctx context.Context, client k8s.ClientInterface, name, namespace string) error {
    timeout := time.After(10 * time.Minute)
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-timeout:
            return fmt.Errorf("timeout waiting for subscription %s", name)
        case <-ticker.C:
            ready, err := checkSubscriptionReady(ctx, client, name, namespace)
            if err != nil {
                return err
            }
            if ready {
                return nil
            }
        }
    }
}
```

---

### 9. Logging & Output

**Decision**: **Mode-aware output**

**Rationale**: Output handling should differ based on TUI vs CLI mode.

**Implementation**:
- **TUI Mode**: Output goes into the TUI display, no console logs
- **CLI Mode**: Structured logging to console with colors

```go
package output

import (
    "fmt"
    "github.com/fatih/color"
)

type OutputMode int

const (
    ModeCLI OutputMode = iota
    ModeTUI
)

type Handler struct {
    mode OutputMode
}

func NewHandler(isTUI bool) *Handler {
    mode := ModeCLI
    if isTUI {
        mode = ModeTUI
    }
    return &Handler{mode: mode}
}

func (h *Handler) Info(msg string, args ...interface{}) {
    if h.mode == ModeCLI {
        fmt.Printf(msg+"\n", args...)
    }
    // In TUI mode, messages are handled by the TUI itself
}

func (h *Handler) Success(msg string, args ...interface{}) {
    if h.mode == ModeCLI {
        green := color.New(color.FgGreen).SprintFunc()
        fmt.Printf("%s %s\n", green("✓"), fmt.Sprintf(msg, args...))
    }
}

func (h *Handler) Error(msg string, args ...interface{}) {
    if h.mode == ModeCLI {
        red := color.New(color.FgRed).SprintFunc()
        fmt.Fprintf(os.Stderr, "%s %s\n", red("✗"), fmt.Sprintf(msg, args...))
    }
}

**Usage in Commands**:
```go
func deployLogging(cmd *cobra.Command, args []string) error {
    isTUI := !hasAllRequiredFlags(cmd)
    output := output.NewHandler(isTUI)
    
    output.Info("Deploying Logging Stack...")
    
    if err := createLokiStack(ctx); err != nil {
        output.Error("Failed to deploy LokiStack: %v", err)
        return err
    }
    output.Success("LokiStack deployed")
    
    if err := createClusterLogForwarder(ctx); err != nil {
        output.Error("Failed to deploy ClusterLogForwarder: %v", err)
        return err
    }
    output.Success("ClusterLogForwarder deployed")
    
    return nil
}
```

---

### 10. Resource Definitions

**Decision**: **No templates - all resources as Go structs**

**Rationale**: Maintain type safety, enable IDE support, avoid runtime template errors.

**Implementation**: Define resources as Go functions returning typed structs:

```go
package resources

import (
    lokiv1 "github.com/grafana/loki/operator/apis/loki/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type LokiStackConfig struct {
    Name      string
    Namespace string
    Size      string  // 1x.extra-small, 1x.small, etc.
    Storage   StorageConfig
}

type StorageConfig struct {
    Provider    string // "minio", future: "s3", "azure", etc.
    SecretName  string
    BucketName  string
}

func NewLokiStack(cfg LokiStackConfig) *lokiv1.LokiStack {
    return &lokiv1.LokiStack{
        ObjectMeta: metav1.ObjectMeta{
            Name:      cfg.Name,
            Namespace: cfg.Namespace,
        },
        Spec: lokiv1.LokiStackSpec{
            Size: lokiv1.LokiStackSizeType(cfg.Size),
            Storage: lokiv1.ObjectStorageSpec{
                Secret: lokiv1.ObjectStorageSecretSpec{
                    Name: cfg.Storage.SecretName,
                    Type: lokiv1.ObjectStorageSecretS3,
                },
            },
            StorageClassName: "gp3-csi",
            Template: &lokiv1.LokiTemplateSpec{
                Compactor: &lokiv1.LokiComponentSpec{
                    Replicas: 1,
                },
            },
        },
    }
}

func RenderLokiStack(params LokiStackParams) (string, error) {
    tmpl, err := template.New("lokistack").Parse(lokiStackTemplate)
    if err != nil {
        return "", err
    }
    
    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, params); err != nil {
        return "", err
    }
    
    return buf.String(), nil
}
```

**Allow File Override**:
```bash
# Use embedded template
devtool deploy logging

# Use custom template
devtool deploy logging --template ./custom-lokistack.yaml
```

---

## Migration Phases

### Phase 1: Foundation (Weeks 1-2)

**Goal**: Establish Go project structure and core utilities

**Tasks**:
1. ✅ Initialize Go module
2. ✅ Set up Cobra CLI framework
3. ✅ Implement Kubernetes client wrapper
4. ✅ Add version detection logic
5. ✅ Create configuration management
6. ✅ Implement logging/UI utilities
7. ✅ Set up testing framework

**Deliverables**:
- `devtool version` command working
- Basic CLI structure
- Connection to cluster validated
- Version detection working

**Success Criteria**:
- Can connect to OpenShift cluster
- Can detect cluster version
- Basic flags and config working

---

### Phase 2: Core Operations (Weeks 3-5)

**Goal**: Implement most common operations

**Priority Order** (based on usage frequency):
1. **Monitoring Plugin Management** (`devtool monitoring scale`)
   - Scale up/down CMO
   - Update plugin image
   - High usage, relatively simple
   
2. **User Management** (`devtool users create`)
   - Create htpasswd users
   - Apply OAuth config
   - Frequently used for demos
   
3. **Dashboard Deployment** (`devtool deploy dashboards`)
   - Deploy Perses dashboards
   - Configure datasources
   - Moderate complexity

**Deliverables**:
- `devtool monitoring scale up/down`
- `devtool users create`
- `devtool deploy dashboards`

**Success Criteria**:
- Can replace `./monitoring-plugin/scale.sh`
- Can replace `./users/create-htpsswd-auth.sh`
- Can replace `./dashboards-manifests/deploy-dashboards.sh`

---

### Phase 3: Complex Deployments (Weeks 6-9)

**Goal**: Implement full stack deployments

**Priority Order**:
1. **Logging Stack** (`devtool deploy logging`)
   - LokiStack
   - ClusterLogForwarder (OTEL/ViaQ)
   - MinIO
   - UIPlugin
   
2. **Tracing Stack** (`devtool deploy tracing`)
   - TempoStack
   - OpenTelemetry Collectors
   - UIPlugin
   - Demo apps
   
3. **Korrel8r/NetObserv** (`devtool deploy korrel8r`)
   - Network Observability
   - FlowCollector
   - LokiStack (netobserv)

**Deliverables**:
- `devtool deploy logging`
- `devtool deploy tracing`
- `devtool deploy korrel8r`
- `devtool deploy all` (orchestration)

**Success Criteria**:
- Can replace Makefiles in tracing-manifests/
- Can replace Makefile in korrel8r-manifests/
- Interactive data model selection works
- Version-specific deployments work

---

### Phase 4: QE & Advanced Features (Weeks 10-12)

**Goal**: QE testing workflows and advanced features

**Tasks**:
1. **QE Bundle Installation** (`devtool qe bundle`)
   - Interactive bundle selection
   - IDMS creation
   - Version-specific logic
   
2. **QE FBC Installation** (`devtool qe fbc`)
   - FBC catalog source creation
   - Subscription management
   
3. **Demo Orchestration** (`devtool demo setup`)
   - Full demo environment
   - RBAC setup for 6 users
   - Dashboard provisioning

**Deliverables**:
- `devtool qe bundle install`
- `devtool qe fbc install`
- `devtool demo setup`

**Success Criteria**:
- Can replace qe-release-testing/ scripts
- Can replace perses/demo/master.sh
- Interactive prompts polished

---

### Phase 5: Migration & Documentation (Weeks 13-14)

**Goal**: Complete migration and documentation

**Tasks**:
1. ✅ Update README with Go instructions
2. ✅ Create migration guide for team
3. ✅ Add shell completion support
4. ✅ Create man pages
5. ✅ Archive old scripts (don't delete yet)
6. ✅ Team training session
7. ✅ Update CI/CD to use new tool

**Deliverables**:
- Complete documentation
- Migration guide
- Shell completion installed
- Old scripts marked deprecated

**Success Criteria**:
- Team can use new tool for daily work
- Documentation covers all use cases
- CI/CD migrated

---

### Phase 6: Refinement & Removal (Weeks 15-16)

**Goal**: Polish and remove old scripts

**Tasks**:
1. ✅ Gather feedback from team
2. ✅ Address pain points
3. ✅ Performance optimization
4. ✅ Remove old bash scripts
5. ✅ Remove old Makefiles/Justfiles
6. ✅ Final documentation review

**Deliverables**:
- Polished tool
- Old scripts removed
- Clean repository

**Success Criteria**:
- No usage of old scripts
- Repository only has Go code
- Team satisfied with new tool

---

## Go Modules Required

Based on the CRD research, here are the required Go modules:

### Core Dependencies
```go
// go.mod
module github.com/YOUR_ORG/devtools

go 1.22

require (
    // CLI Framework
    github.com/spf13/cobra v1.10.2
    
    // Kubernetes Client (NO VIPER - using type-safe structs)
    k8s.io/client-go v0.36.0
    k8s.io/apimachinery v0.36.0
    sigs.k8s.io/controller-runtime v0.21.3
    
    // OpenShift APIs
    github.com/openshift/api v0.0.0-20260605005319
    github.com/openshift/cluster-logging-operator // For ClusterLogForwarder
    
    // Operator Framework
    github.com/operator-framework/api v0.43.0
    
    // Observability CRDs (as needed)
    github.com/grafana/loki/operator v0.0.0-20260101000000-xxxxx
    github.com/grafana/tempo-operator v0.20.0
    github.com/open-telemetry/opentelemetry-operator v0.148.0
    github.com/netobserv/network-observability-operator v1.7.0
    github.com/prometheus-operator/prometheus-operator v0.91.0
    github.com/perses/perses v0.53.1
    github.com/rhobs/perses-operator v0.1.10
    github.com/rhobs/observability-operator v0.0.0-20260101000000-xxxxx
    github.com/stolostron/multicluster-observability-operator v0.0.0-20260101000000-xxxxx
    
    // TUI (SELECTED - not optional)
    github.com/charmbracelet/bubbletea v1.3.10
    github.com/charmbracelet/lipgloss v1.1.0
    github.com/charmbracelet/huh v1.0.0  // Forms with paste support
    github.com/charmbracelet/bubbles v1.0.0
    github.com/charmbracelet/log v1.0.0  // Structured logging
    
    // Terminal utilities
    github.com/muesli/termenv // For color profiles
    golang.org/x/term v0.44.0  // Terminal detection
    
    // Utilities
    github.com/pkg/errors v0.9.1
    github.com/sirupsen/logrus v1.9.3
    golang.org/x/mod v0.15.0 // For semver
    
    // Testing
    github.com/stretchr/testify v1.8.4
    sigs.k8s.io/controller-runtime/pkg/envtest v0.17.0
)
```

### Version-Specific Module Selection

See the detailed [Go Modules Research document](/tmp/opencode/crd_go_modules_research.md) for:
- Version compatibility matrix
- When to use forked vs upstream modules
- Decision trees for module selection
- Troubleshooting common issues

**Key Pattern**:
```go
// For OpenShift 4.17-4.18 Console resources
import osRhobsv1 "github.com/rhobs/openshift-api/console/v1"

// For OpenShift 4.19+ Console resources  
import osv1 "github.com/openshift/api/console/v1"

// Runtime selection
if version.IsOCP419OrNewer() {
    // Use osv1
} else {
    // Use osRhobsv1
}
```

---

## Connection Requirements

### Kubernetes Connection

**Authentication Methods Supported**:
1. **Kubeconfig File** (default: `~/.kube/config`)
2. **In-Cluster Config** (when running as pod)
3. **Service Account Token** (explicit path)

**Implementation**:
```go
func getKubeConfig(kubeconfigPath string) (*rest.Config, error) {
    // Priority: explicit flag > KUBECONFIG env > default location
    if kubeconfigPath != "" {
        return clientcmd.BuildConfigFromFlags("", kubeconfigPath)
    }
    
    kubeconfigEnv := os.Getenv("KUBECONFIG")
    if kubeconfigEnv != "" {
        return clientcmd.BuildConfigFromFlags("", kubeconfigEnv)
    }
    
    // Try in-cluster config first (for running as pod)
    if config, err := rest.InClusterConfig(); err == nil {
        return config, nil
    }
    
    // Fall back to default location
    home, _ := os.UserHomeDir()
    defaultPath := filepath.Join(home, ".kube", "config")
    return clientcmd.BuildConfigFromFlags("", defaultPath)
}
```

**Timeout Configuration**:
```go
config.Timeout = 30 * time.Second // API request timeout
config.QPS = 50                   // Queries per second
config.Burst = 100                // Burst allowance
```

---

## Alternatives & Trade-offs

### Alternative 1: Keep Bash, Add Testing

**Pros**:
- No migration needed
- Team already familiar
- Quick wins with testing

**Cons**:
- Still fragmented
- Hard to test thoroughly
- No type safety
- Doesn't solve root problems

**Verdict**: ❌ Doesn't address core issues

---

### Alternative 2: Operator-SDK

**Pros**:
- Built for Kubernetes operators
- Good scaffolding

**Cons**:
- Overkill for a CLI tool
- Operator-centric, not CLI-centric
- Heavy dependency

**Verdict**: ❌ Wrong tool for the job

---

### Alternative 3: Shell + Go Hybrid

**Pros**:
- Incremental migration
- Low risk

**Cons**:
- Still fragmented
- Maintains technical debt
- Complex to maintain both

**Verdict**: ⚠️ Possible transition strategy, not end goal

---

### Alternative 4: Python

**Pros**:
- Rich ecosystem
- Good for scripting

**Cons**:
- Not Kubernetes ecosystem standard
- Dependency management challenges
- Runtime requirement
- Team not Python-focused

**Verdict**: ❌ Wrong choice for Kubernetes tooling

---

### Recommended Approach: Full Go Migration

**Why Go is the Right Choice**:
1. ✅ **Kubernetes Native**: All K8s tooling is in Go
2. ✅ **Type Safety**: Compile-time guarantees
3. ✅ **Single Binary**: No runtime dependencies
4. ✅ **CRD Integration**: Native support for all CRDs
5. ✅ **Team Alignment**: Team works on Go operators
6. ✅ **Ecosystem**: Rich Kubernetes Go libraries
7. ✅ **Performance**: Fast execution
8. ✅ **Testing**: Excellent testing tools

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Migration takes longer than expected | Medium | Medium | Phased approach, parallel run |
| Team resistance to new tool | Low | Low | Early involvement, training |
| Loss of functionality | Low | High | Comprehensive requirements gathering |
| Bugs in new implementation | Medium | Medium | Extensive testing, gradual rollout |
| Breaking changes during migration | Medium | High | Keep old scripts until fully validated |
| Performance issues | Low | Low | Performance testing, benchmarking |

---

## Success Metrics

### Technical Metrics
- ✅ 100% functional parity with bash scripts
- ✅ 70%+ test coverage
- ✅ <5 second startup time
- ✅ All CRD types supported
- ✅ Version detection working for OCP 4.11-4.19+

### Team Adoption Metrics
- ✅ 100% team using new tool within 4 weeks
- ✅ Zero bash scripts used after 8 weeks
- ✅ Positive feedback from team
- ✅ Faster development velocity

### Quality Metrics
- ✅ <5 bugs reported in first month
- ✅ All major operations working
- ✅ CI/CD fully migrated

---

## Timeline Summary

| Phase | Duration | Completion Date | Key Deliverable |
|-------|----------|-----------------|-----------------|
| Phase 1: Foundation | 2 weeks | Week 2 | Core framework ready |
| Phase 2: Core Operations | 3 weeks | Week 5 | Common commands working |
| Phase 3: Complex Deployments | 4 weeks | Week 9 | Full stack deployments |
| Phase 4: QE & Advanced | 3 weeks | Week 12 | QE workflows complete |
| Phase 5: Migration & Docs | 2 weeks | Week 14 | Team migrated |
| Phase 6: Refinement | 2 weeks | Week 16 | Old scripts removed |

**Total Timeline**: 16 weeks (4 months)

---

## Next Steps

### Immediate Actions (This Week)

1. **Team Review** (Required)
   - [ ] Review this plan with the team
   - [ ] Gather feedback and concerns
   - [ ] Vote on key decisions:
     - [ ] CLI framework (Cobra vs urfave/cli)
     - [ ] Interactive mode (prompts vs flags-only vs TUI)
     - [ ] Timeline approval
   
2. **Tool Name Decision** (Required)
   - [ ] Choose final name: `devtool`, `obsvtool`, or other?
   
3. **Repo Setup** (If approved)
   - [ ] Initialize Go module
   - [ ] Set up GitHub Actions for CI
   - [ ] Create project board

### Week 1 Tasks (If Approved)

1. **Setup**
   - [ ] Run `go mod init`
   - [ ] Install Cobra CLI generator
   - [ ] Set up directory structure
   
2. **Proof of Concept**
   - [ ] Implement `devtool version` command
   - [ ] Implement basic cluster connection
   - [ ] Implement version detection
   
3. **Team Demo**
   - [ ] Show working POC
   - [ ] Get feedback
   - [ ] Adjust plan if needed

---

## Questions for Team Discussion

### 1. Naming
- What should we name the CLI tool?
  - `devtool`
  - `obsvtool`
  - `o11y-dev`
  - Other suggestions?

### 2. Interactive Mode
- How important is interactive mode vs flags-only?
  - Always interactive by default?
  - Interactive only when flags missing?
  - Flags-only, no prompts?

### 3. TUI Enhancement
- Do we want a TUI for long-running operations?
  - Nice to have but not required?
  - Critical for UX?
  - Not needed?

### 4. Timeline
- Is 16 weeks realistic?
  - Too aggressive?
  - Can we go faster?
  - Need more time?

### 5. Parallel Development
- Should we:
  - Freeze new bash script development now?
  - Continue bash scripts until Go version ready?
  - Hybrid approach?

### 6. Repository Location
- Where should this live?
  - Same repo (development-tools)?
  - New dedicated repo?
  - Within observability-operator repo?

---

## Conclusion

This plan provides a comprehensive roadmap for converting the development-tools repository from a fragmented multi-technology stack to a unified Go-based CLI tool. The phased approach minimizes risk while delivering value incrementally.

**Key Benefits**:
- 🎯 Single tool, single language
- 🔒 Type safety and compile-time guarantees
- 🧪 Testable and maintainable
- 🚀 Better error handling and user experience
- 📚 Native Kubernetes integration
- 🔄 Version-aware deployments

**Recommended Next Step**: Team review and decision on key choices (framework, naming, timeline).

---

## Appendices

- **Appendix A**: [Detailed Go Modules Research](/tmp/opencode/crd_go_modules_research.md)
- **Appendix B**: Connection Patterns (see perses login example)
- **Appendix C**: Observability Operator PR #1100 Analysis (version-specific CRDs)

**Document Version**: 1.0  
**Last Updated**: June 8, 2026  
**Next Review**: After team feedback
