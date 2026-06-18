# obstool Architecture

CLI tool for OpenShift Observability team development workflows.

**Name**: `obstool` (observability tooling)  
**Purpose**: Unified Go-based replacement for fragmented bash/make/just/js/yaml tooling  
**Status**: Implementation in progress - Foundation complete, commands being added

---

## 1. Overview (50 lines)

### Problem Statement

Legacy repository suffered from:
- **Technology Fragmentation**: 6+ technologies (Bash 73%, YAML 17%, Just 5%, Make 3%, TS 2%)
- **Duplicate Actions**: Multiple ways to accomplish same tasks
- **Maintainability Issues**: 58 shell scripts, difficult to extend
- **No Type Safety**: Runtime errors, inconsistent validation
- **Limited Error Handling**: Inconsistent error propagation

### Solution

Single Go-based CLI providing:
- Type-safe Kubernetes resource interactions
- Unified binary with no runtime dependencies
- Better error handling and testing
- Native CRD support via Go modules
- Version-aware deployments (OCP 4.17-4.19+)

### Design Principles

1. **Kubernetes Native**: Use controller-runtime directly, no abstraction layers
2. **Type Safety**: All resources as Go structs, no YAML templates
3. **Mode Awareness**: CLI mode (all flags) vs TUI mode (missing flags)
4. **Business Logic Decoupling**: Single implementation for both modes via channels
5. **Minimal Testing**: Unit tests for critical code only, no CI/CD overhead
6. **Low Barrier**: Simple patterns, easy contribution, manual testing acceptable
7. **Version Awareness**: Detect OCP version, conditional resource creation
8. **Flat Resources**: All resource files at top level except dashboards/

### Core Architecture Pattern

```
Business Logic (pkg/operations/)
    ↓ sends ProgressUpdate via channel
    ├─→ CLI Handler (pkg/output/cli.go) - displays with logger
    └─→ TUI Handler (command-specific) - forwards to Bubble Tea
```

Enables:
- Write business logic once
- CLI and TUI guaranteed consistent
- Easy testing without display mocking
- Scalable to many commands

---

## 2. Directory Structure (100 lines)

```
obstool/
├── cmd/                            # Cobra command definitions
│   ├── root.go                     # Root command + global flags (--kubeconfig, --namespace)
│   ├── version.go                  # Version command (cluster version info)
│   ├── deploy/
│   │   ├── deploy.go               # Deploy command group (TUI selection when no subcommand)
│   │   ├── all.go                  # Deploy all components (flag mode only)
│   │   ├── coo.go                  # Deploy COO (4 methods: bundle, fbc, stage, operatorhub)
│   │   ├── logging.go              # Deploy logging stack (LokiStack + ClusterLogForwarder + UIPlugin)
│   │   ├── tracing.go              # Deploy tracing stack (TempoStack + OTEL + UIPlugin)
│   │   ├── dashboards.go           # Deploy dashboards (Perses datasources + dashboards)
│   │   ├── monitoring.go           # Deploy monitoring plugin (UIPlugin only)
│   │   ├── acm.go                  # Deploy ACM observability (MultiClusterObservability)
│   │   └── korrel8r.go             # Deploy korrel8r (NetObserv + FlowCollector)
│   ├── users/
│   │   ├── users.go                # User management command group
│   │   ├── create.go               # Create test users (htpasswd + OAuth)
│   │   └── rbac.go                 # Apply RBAC scenarios (6 predefined roles)
│   ├── update/
│   │   ├── update.go               # Update command group
│   │   ├── monitoring.go           # Scale down CMO, update plugin image
│   │   └── coo.go                  # Update COO in-place (new version)
│   └── cleanup/
│       ├── cleanup.go              # Cleanup command group
│       ├── monitoring.go           # Scale up CMO (restore plugin)
│       ├── coo.go                  # Cleanup COO (subscription, CSV, CRDs)
│       ├── logging.go              # Cleanup logging stack
│       ├── tracing.go              # Cleanup tracing stack
│       ├── acm.go                  # Cleanup ACM observability
│       └── all.go                  # Cleanup all components (flag mode only)
├── pkg/
│   ├── k8s/
│   │   ├── client.go               # Kubernetes client wrapper (controller-runtime)
│   │   ├── connection.go           # Connection management (kubeconfig only, timeout 30s, QPS 50)
│   │   └── version.go              # Version detection (ClusterVersion CR)
│   ├── context/
│   │   └── execcontext.go          # Execution context (context.WithValue pattern for client, version, TUI)
│   ├── executor/
│   │   └── executor.go             # Channel-based progress updates (ProgressUpdate struct)
│   ├── operations/
│   │   ├── monitoring.go           # Business logic: monitoring operations (scale up/down, update)
│   │   ├── coo.go                  # Business logic: COO deployment (4 methods)
│   │   ├── logging.go              # Business logic: logging stack deployment
│   │   ├── tracing.go              # Business logic: tracing stack deployment
│   │   ├── dashboards.go           # Business logic: dashboard deployment
│   │   ├── users.go                # Business logic: user creation
│   │   └── rbac.go                 # Business logic: RBAC application
│   ├── resources/                  # Resource definitions (flat structure)
│   │   ├── uiplugin.go             # UIPlugin CRs (using COO's abstraction)
│   │   ├── lokistack.go            # LokiStack resources (from loki-operator)
│   │   ├── clusterlogforwarder.go  # ClusterLogForwarder (from cluster-logging-operator)
│   │   ├── tempostack.go           # TempoStack resources (from tempo-operator)
│   │   ├── otel.go                 # OpenTelemetry collectors (from otel-operator)
│   │   ├── perses.go               # Perses datasource and global resources
│   │   ├── rbac.go                 # RBAC resources (ClusterRole, RoleBinding)
│   │   ├── minio.go                # MinIO deployment resources (to be deprecated)
│   │   └── dashboards/             # Dashboard definitions (30+ files, 1 per dashboard)
│   │       ├── node-exporter.go
│   │       ├── prometheus.go
│   │       ├── thanos.go
│   │       └── ...
│   ├── operators/
│   │   ├── olm.go                  # OLM operations (wait for CSV ready)
│   │   ├── subscription.go         # Subscription management (create, wait, delete)
│   │   ├── catalogsource.go        # CatalogSource management (create, wait)
│   │   ├── operatorgroup.go        # OperatorGroup management (ensure exists)
│   │   ├── idms.go                 # ImageDigestMirrorSet (for stage/internal registries)
│   │   ├── coo/                    # COO-specific deployment methods
│   │   │   ├── bundle.go           # Bundle deployment (OLM bundle install)
│   │   │   ├── fbc.go              # FBC deployment (file-based catalog)
│   │   │   ├── stage.go            # Stage registry deployment (registry.stage.redhat.io)
│   │   │   ├── operatorhub.go      # OperatorHub deployment (default OCP catalog)
│   │   │   └── update.go           # In-place update (change subscription version)
│   │   └── bundle.go               # Generic bundle operations (extract, parse)
│   ├── users/
│   │   ├── htpasswd.go             # htpasswd management (create file, encode)
│   │   └── oauth.go                # OAuth configuration (update identity provider)
│   ├── storage/
│   │   ├── provider.go             # Storage provider interface (future abstraction)
│   │   └── minio.go                # MinIO implementation (create deployment, service, secret)
│   ├── tui/
│   │   ├── deploy.go               # Deploy selection TUI (Bubble Tea multi-select)
│   │   ├── progress.go             # Progress display (general component)
│   │   ├── models.go               # TUI models (state management)
│   │   └── styles.go               # Styling (colors, layout)
│   ├── output/
│   │   ├── cli.go                  # CLI output handler (general for all commands)
│   │   └── handler.go              # Handler interface
│   ├── mode/
│   │   └── detect.go               # Mode detection (terminal check, flag completeness)
│   └── config/
│       └── config.go               # Type-safe configuration (no Viper, Go structs only)
└── internal/
    ├── version/
    │   ├── version.go              # Version detection logic (OCP version from ClusterVersion)
    │   └── compare.go              # Version comparison (semver)
    └── constants/
        └── constants.go            # Shared constants (namespaces, labels, timeouts)
```

### Key Structural Decisions

1. **No qe/ directory** - COO deployment methods in `operators/coo/`
2. **No demo/ directory** - Demo workflows via command composition
3. **Flat pkg/resources/** - All resource files except dashboards/ subdirectory
4. **Added update/ commands** - COO updates, monitoring scale operations
5. **Cleanup mirrors deploy** - Same structure for consistency
6. **Storage provider interface** - Future MinIO deprecation support
7. **Context package** - Execution context with client, version, TUI flag
8. **Executor package** - Channel-based progress updates for mode decoupling

---

## 3. Architectural Patterns (200 lines)

### Executor Pattern (Multi-Step Functions)

**Purpose**: Enable both CLI and TUI modes with consistent progress tracking.

**Rules**:
- Any function performing multiple operations MUST accept `*executor.Executor`
- Define step constants at package level (`const StepOne = iota`)
- Send progress updates via `exec.SendUpdate(step, status, description)`
- Send detailed logs via `exec.SendLog(step, message)`
- Send errors via `exec.SendUpdateWithError(step, status, description, err)`
- Always mark steps complete: `exec.SendUpdate(step, executor.StatusComplete, description)`

**Example**:

```go
// Define steps
const (
    StepCreateNamespace = iota
    StepCreateResources
    StepWaitForReady
)

// Multi-step function with executor
func DeployComponent(ctx context.Context, client client.Client, 
                    config Config, exec *executor.Executor) error {
    defer exec.Close()
    
    stepName := "Create namespace"
    exec.SendUpdate(StepCreateNamespace, executor.StatusInProgress, stepName)
    exec.SendLog(StepCreateNamespace, "Ensuring namespace exists")
    
    err := createNamespace(ctx, client, config.Namespace)
    if err != nil {
        exec.SendUpdateWithError(StepCreateNamespace, executor.StatusFailed, stepName, err)
        return err
    }
    exec.SendUpdate(StepCreateNamespace, executor.StatusComplete, stepName)
    
    stepName = "Create resources"
    exec.SendUpdate(StepCreateResources, executor.StatusInProgress, stepName)
    exec.SendLog(StepCreateResources, fmt.Sprintf("Creating %d resources", len(config.Resources)))
    
    err = createResources(ctx, client, config.Resources)
    if err != nil {
        exec.SendUpdateWithError(StepCreateResources, executor.StatusFailed, stepName, err)
        return err
    }
    exec.SendUpdate(StepCreateResources, executor.StatusComplete, stepName)
    
    return nil
}
```

### Business Logic Decoupling Pattern

**Architecture**: Business logic sends progress updates via Go channels to display handlers.

**Flow**:
```
Business Logic (pkg/operations/)
    ↓ ProgressUpdate via exec.UpdateCh
    ├─→ CLI Mode: output.CLIHandler receives updates, logs with charmbracelet/log
    └─→ TUI Mode: Command-specific Bubble Tea model receives updates
```

**Implementation**:

**Business Logic** (pkg/operations/):
```go
func ExecuteMonitoring(ctx context.Context, client client.Client, 
                      config MonitoringConfig, exec *executor.Executor) error {
    defer exec.Close()
    
    exec.SendUpdate(StepScaleDown, executor.StatusInProgress, "Scale down CMO")
    // ... do work ...
    exec.SendUpdate(StepScaleDown, executor.StatusComplete, "Scale down CMO")
    
    return nil
}
```

**CLI Mode** (uses general handler):
```go
func runCLIMode(ctx context.Context, client client.Client, config Config) error {
    handler := output.NewCLIHandler()
    exec := executor.NewExecutor()
    
    go operations.ExecuteMonitoring(ctx, client, config, exec)
    
    for update := range exec.UpdateCh {
        if err := handler.HandleUpdate(update); err != nil {
            return err
        }
    }
    return nil
}
```

**TUI Mode** (command-specific components):
```go
func runTUIMode(ctx context.Context, client client.Client, config Config) error {
    model := tui.NewProgressModel("Monitoring Operations", operations)
    program := tea.NewProgram(model)
    exec := executor.NewExecutor()
    
    go operations.ExecuteMonitoring(ctx, client, config, exec)
    
    go func() {
        for update := range exec.UpdateCh {
            if update.Message != "" {
                continue  // TUI ignores detailed log messages
            }
            program.Send(tui.OperationUpdateMsg{
                Step:   update.Step,
                Status: update.Status,
                Name:   update.Name,
                Error:  update.Error,
            })
        }
    }()
    
    _, err := program.Run()
    return err
}
```

**Benefits**:
- Business logic written once
- CLI and TUI guaranteed consistent
- Easy to test (no display mocking)
- Scalable to many commands

### Execution Context Pattern

**Purpose**: Pass shared state (client, version, TUI flag) without cluttering function signatures.

**Implementation**:
```go
// pkg/context/execcontext.go
func WithClient(ctx context.Context, client client.Client) context.Context {
    return context.WithValue(ctx, clientKey, client)
}

func GetClient(ctx context.Context) (client.Client, error) {
    client, ok := ctx.Value(clientKey).(client.Client)
    if !ok {
        return nil, fmt.Errorf("client not found in context")
    }
    return client, nil
}

func WithTUI(ctx context.Context, isTUI bool) context.Context {
    return context.WithValue(ctx, tuiKey, isTUI)
}

func IsTUI(ctx context.Context) bool {
    isTUI, ok := ctx.Value(tuiKey).(bool)
    return ok && isTUI
}
```

**Usage**:
```go
// Set values in command
ctx = execctx.WithClient(ctx, kubeClient)
ctx = execctx.WithVersion(ctx, versionInfo)
ctx = execctx.WithTUI(ctx, isTUI)

// Retrieve in operations
client, err := execctx.GetClient(ctx)
version, err := execctx.GetVersion(ctx)
isTUI := execctx.IsTUI(ctx)
```

### Mode Detection Pattern

**Purpose**: Determine whether to run in CLI (non-interactive) or TUI (interactive) mode.

**Rules**:
- **All required flags present** → CLI mode (direct execution)
- **Missing flags + terminal** → TUI mode (interactive selection)
- **Missing flags + non-terminal** → Error with helpful message

**Implementation**:
```go
import "golang.org/x/term"

func deployLogging(cmd *cobra.Command, args []string) error {
    // Check if all required flags are provided
    if hasAllRequiredFlags(cmd) {
        // CLI mode - direct execution
        return runCLIMode(cmd.Context(), getFlagsConfig(cmd))
    }
    
    // Check if we're in a non-interactive environment
    if !term.IsTerminal(int(os.Stdin.Fd())) {
        return fmt.Errorf("missing required flags; run with --help to see required flags")
    }
    
    // TUI mode - interactive selection
    return runTUIMode(cmd.Context())
}
```

### Version Detection Pattern

**Purpose**: Detect OpenShift version for conditional resource creation.

**Key Use Case**: Console CRD requires different modules for OCP 4.17-4.18 vs 4.19+.

**Implementation**:
```go
// pkg/k8s/version.go
func DetectVersion(ctx context.Context, c client.Client) (*VersionInfo, error) {
    cv := &configv1.ClusterVersion{}
    key := client.ObjectKey{Name: "version"}
    if err := c.Get(ctx, key, cv); err != nil {
        return nil, err
    }
    
    return &VersionInfo{
        OpenShiftVersion: cv.Status.Desired.Version,
    }, nil
}

// internal/version/compare.go
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

**Usage**:
```go
func CreateUIPlugin(ctx context.Context, plugin *UIPluginConfig) error {
    version, err := execctx.GetVersion(ctx)
    if err != nil {
        return err
    }
    
    // UIPlugin from COO handles ConsolePlugin version differences internally
    // We just need to be aware of OCP version for other resource decisions
    
    uiPlugin := &uiv1alpha1.UIPlugin{
        ObjectMeta: metav1.ObjectMeta{
            Name: plugin.Name,
        },
        Spec: plugin.Spec,
    }
    
    return client.Create(ctx, uiPlugin)
}
```

### Resource Definition Pattern

**Purpose**: Define all resources as type-safe Go structs, no YAML templates.

**Benefits**:
- Type safety at compile time
- IDE autocomplete
- No runtime template errors
- Easy to test

**Example**:
```go
// pkg/resources/lokistack.go
type LokiStackConfig struct {
    Name      string
    Namespace string
    Size      string  // 1x.extra-small, 1x.small, etc.
    Storage   StorageConfig
}

type StorageConfig struct {
    Provider    string // "minio", future: "s3", "azure"
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
```

### Ensure Functions Pattern

**Purpose**: Idempotent resource creation with caller feedback.

**Rules**:
- Functions named `Ensure*` must return `(bool, error)`
- Return `(true, nil)` when resource was created
- Return `(false, nil)` when resource already existed
- Return `(false, err)` on error

**Example**:
```go
func EnsureOperatorGroup(ctx context.Context, client client.Client, 
                        namespace string) (bool, error) {
    og := &operatorsv1.OperatorGroup{}
    key := client.ObjectKey{Namespace: namespace, Name: "observability"}
    
    err := client.Get(ctx, key, og)
    if err == nil {
        return false, nil  // Already exists
    }
    
    if !errors.IsNotFound(err) {
        return false, err  // Unexpected error
    }
    
    // Create new OperatorGroup
    og = &operatorsv1.OperatorGroup{
        ObjectMeta: metav1.ObjectMeta{
            Name:      "observability",
            Namespace: namespace,
        },
        Spec: operatorsv1.OperatorGroupSpec{
            TargetNamespaces: []string{namespace},
        },
    }
    
    if err := client.Create(ctx, og); err != nil {
        return false, err
    }
    
    return true, nil  // Created new resource
}
```

---

## 4. Technology Stack (100 lines)

### Core Dependencies

**CLI Framework**:
- **Cobra** (`github.com/spf13/cobra v1.10.2`)
- Industry standard for Kubernetes tooling (kubectl, oc, helm)
- Built-in shell completion support
- Excellent documentation
- Nested command structure

**TUI Library**:
- **Bubble Tea** (`github.com/charmbracelet/bubbletea v1.3.10`)
- Elm-inspired architecture (Model-Update-View)
- Used for progress display and multi-select
- **Huh** (`github.com/charmbracelet/huh v1.0.0`)
- Forms with paste support (input validation, multi-select)
- **Lipgloss** (`github.com/charmbracelet/lipgloss v1.1.0`)
- Styling and layout
- **Bubbles** (`github.com/charmbracelet/bubbles v1.0.0`)
- Reusable components (spinner, progress bar)

**Logging**:
- **charmbracelet/log** (`github.com/charmbracelet/log v1.0.0`)
- Structured logging with color support
- Integrates with Bubble Tea (non-interfering)
- Levels: debug, info, warn, error

**Kubernetes Client**:
- **controller-runtime** (`sigs.k8s.io/controller-runtime v0.21.3`)
- Simpler API than client-go
- Same library used by operators (team familiarity)
- No abstraction layer (use directly)
- Timeout: 30s, QPS: 50, Burst: 100

**OpenShift APIs**:
- **openshift/api** (`github.com/openshift/api v0.0.0-20260605005319`)
- ClusterVersion, Console, Route, OAuth
- **cluster-logging-operator** (`github.com/openshift/cluster-logging-operator`)
- ClusterLogForwarder CRD (observability/v1)

**Operator Framework**:
- **operator-framework/api** (`github.com/operator-framework/api v0.43.0`)
- Subscription, CatalogSource, ClusterServiceVersion, OperatorGroup

**Observability CRDs**:
- **loki-operator** (`github.com/grafana/loki/operator`)
- LokiStack CRD
- **tempo-operator** (`github.com/grafana/tempo-operator v0.20.0`)
- TempoStack CRD
- **otel-operator** (`github.com/open-telemetry/opentelemetry-operator v0.148.0`)
- OpenTelemetryCollector CRD
- **netobserv-operator** (`github.com/netobserv/network-observability-operator v1.7.0`)
- FlowCollector CRD
- **perses-operator** (`github.com/rhobs/perses-operator v0.1.10`)
- PersesDashboard, PersesDatasource CRDs
- **observability-operator** (`github.com/rhobs/observability-operator`)
- UIPlugin CRD (abstraction over ConsolePlugin)
- **acm-operator** (`github.com/stolostron/multicluster-observability-operator`)
- MultiClusterObservability CRD

**Utilities**:
- **golang.org/x/term** (`v0.44.0`)
- Terminal detection (IsTerminal)
- **golang.org/x/mod** (`v0.15.0`)
- Semver comparison
- **muesli/termenv** (`github.com/muesli/termenv`)
- Color profile detection

**Testing** (minimal):
- **testify** (`github.com/stretchr/testify v1.8.4`)
- Assertions and mocking
- **envtest** (`sigs.k8s.io/controller-runtime/pkg/envtest v0.17.0`)
- Kubernetes API server for integration tests

### Technology Rationale

**Why Cobra?**
- Kubernetes ecosystem standard
- Nested commands (deploy/logging, cleanup/all)
- Flag parsing and validation
- Shell completion generation

**Why Bubble Tea?**
- Modern TUI framework
- Type-safe Elm architecture
- Good documentation and examples
- Active community support
- Used by charm.sh tools (gum, mods)

**Why controller-runtime?**
- Team familiar with operators
- Simpler than client-go
- Typed client interface
- Scheme management built-in
- Can refactor later if needed

**Why Go structs for config?**
- Type safety at compile time
- IDE autocomplete
- No runtime config parsing
- Easy to test
- Can override programmatically

**Why no Viper?**
- Config is simple (defaults in Go structs)
- No external config file needed
- Reduces dependencies
- Type safety maintained

**Why no in-cluster config?**
- Tool is for local development
- Always run from workstation
- Kubeconfig auth only
- Simpler implementation

---

## 5. Key Decisions (100 lines)

| Decision | Rationale | Impact |
|----------|-----------|--------|
| **CLI Framework: Cobra** | Industry standard for K8s tooling, excellent docs, shell completion | Standard patterns, easy contribution |
| **TUI Library: Bubble Tea** | Modern framework, type-safe, good examples | Interactive mode polished UX |
| **Logging: charmbracelet/log** | Structured logging, color support, TUI-compatible | Clean output in both modes |
| **K8s Client: controller-runtime directly** | Team familiar, simpler API, no abstraction overhead | Fast implementation, refactor later if needed |
| **Config: Go structs (no Viper)** | Type safety, compile-time errors, IDE support | No runtime config errors |
| **Resource Definitions: Go structs (no YAML)** | Type safety, no template errors, CRD modules provide types | Compile-time validation |
| **Testing: Minimal (critical code only)** | Low barrier to contribution, manual testing acceptable | Fast iteration, easy fixes |
| **Mode Detection: Flags + TUI hybrid** | Automation-friendly CLI, intuitive TUI for interactive | Best of both worlds |
| **Business Logic: Channel-based decoupling** | Write once, both modes consistent, easy testing | Scalable architecture |
| **Execution Context: context.WithValue** | Avoid passing state everywhere, clean signatures | Simpler function signatures |
| **Version Detection: Centralized** | Handle OCP version differences in one place | Consistent version-aware logic |
| **Timeout: 30s API, 10m operations** | Balance responsiveness vs slow clusters | Fails fast but allows rollouts |
| **QPS/Burst: 50/100** | High throughput for bulk operations | Fast resource creation |
| **Flat Resources: pkg/resources/*.go** | Easy to find, no deep nesting (except dashboards/) | Simple navigation |
| **Dashboards: Subdirectory** | 30+ files, isolate from other resources | Clean resources/ directory |
| **No Demo Command** | Use command composition instead | Simpler CLI structure |
| **Cleanup All: Flag mode only** | Prevent accidental deletions | Safety requirement |
| **No In-Cluster Config** | Tool for local development only | Simpler auth flow |
| **Storage Provider Interface** | Future MinIO deprecation | Easy swap to S3/Azure |
| **UIPlugin from COO** | Abstracts ConsolePlugin version differences | No version-specific UI code |
| **ClusterLogForwarder: cluster-logging-operator** | Correct module (not observability-operator) | Proper CRD imports |
| **Console CRD: Version-specific modules** | OCP 4.17-4.18 vs 4.19+ require different imports | Runtime version detection |
| **Ensure Functions: (bool, error)** | Caller knows if resource was created or existed | Informative feedback |
| **Executor Pattern: Required for multi-step** | Enables progress tracking in both modes | Consistent UX |
| **No Parallel Development Freeze** | Bash scripts continue during Go development | Zero risk migration |

### Critical Architecture Constraints

1. **Authentication**: Kubeconfig file only (no in-cluster config)
2. **Cleanup All**: Flag mode only with `--confirm=yes` flag
3. **Dashboards**: 30+ files in `pkg/resources/dashboards/`, 1 per file
4. **ClusterLogForwarder**: From `cluster-logging-operator` NOT `observability-operator`
5. **Console CRD**: Different modules for OCP 4.17-4.18 vs 4.19+
6. **No Templates**: All resources as Go structs
7. **No Demo Command**: Achieved via command composition

---

## 6. Migration Strategy (50 lines)

### Phases Overview

**Phase 1: Foundation (Weeks 1-2)**
- Initialize Go module
- Set up Cobra framework
- Implement K8s client wrapper
- Add version detection
- Create config management
- Implement executor pattern

**Phase 2: Core Operations (Weeks 3-5)**
- Monitoring plugin management (scale up/down, update)
- User management (create users, apply RBAC)
- Dashboard deployment (Perses datasources + dashboards)

**Phase 3: Complex Deployments (Weeks 6-9)**
- Logging stack (LokiStack + ClusterLogForwarder + UIPlugin)
- Tracing stack (TempoStack + OTEL + UIPlugin)
- Korrel8r/NetObserv (FlowCollector + LokiStack)

**Phase 4: COO & Advanced Features (Weeks 10-12)**
- COO deployment (4 methods: bundle, fbc, stage, operatorhub)
- COO update in-place
- ACM observability

**Phase 5: Migration & Documentation (Weeks 13-14)**
- Update README
- Create migration guide
- Shell completion
- Archive old scripts

**Phase 6: Refinement & Removal (Weeks 15-16)**
- Gather feedback
- Address pain points
- Remove old bash scripts

### Parallel Development Strategy

**No Freeze**: Bash scripts continue to work during Go development
- Team can use bash for production work
- Go tool developed in parallel
- Rebase on bash script changes
- No timeline pressure
- Focus on quality and maintainability

**Validation**: New command implemented → manual testing → feedback → iterate
- No automated CI/CD required
- Manual testing acceptable
- Team testing preferred
- Quick fixes encouraged

**Migration**: 100% functional parity → team training → gradual adoption → script removal
- No forced switch
- Demonstrate value first
- Support bash and Go during transition
- Remove bash only after team satisfied
