# Plan: Implement k8s Client Package

**Task ID**: Foundation → Kubernetes Client  
**Status**: ✅ Complete (Implementation: [implementation.md](./implementation.md))  
**Date**: 2026-06-10  
**Dependencies**: Add core dependencies (✅ Complete)

---

## Overview

Implement the `pkg/k8s/` package that provides Kubernetes client functionality for obstool. This package will use controller-runtime directly (no abstraction layer) and support kubeconfig-based authentication only (no in-cluster config).

---

## Goals

1. **Create a robust Kubernetes client** using controller-runtime
2. **Support kubeconfig loading** with proper priority: flag > env > default
3. **Configure client parameters** (timeout, QPS, burst)
4. **Register all required schemes** for the 18+ CRD types used in obstool
5. **Provide clean API** for use throughout the application

---

## Non-Goals

- In-cluster authentication (explicitly not supported per architecture decisions)
- Client abstraction layer (use controller-runtime directly)
- Connection pooling or caching (controller-runtime handles this)
- Mocking/test utilities (minimal testing philosophy)
- Registering operator CRD schemes (Loki, Tempo, etc.) - these will be added when implementing each command

---

## Architecture Context

From CONTEXT.md and go-migration-plan.md:

**Key Decisions**:
- Use `controller-runtime/pkg/client` directly (no abstraction)
- Authentication via kubeconfig file only
- Client configuration: timeout=30s, QPS=50, burst=100
- Scheme registration for all CRDs up front

**Directory Structure**:
```
pkg/k8s/
├── client.go       # Client struct and NewClient()
├── connection.go   # Kubeconfig loading (OR combine into client.go)
└── scheme.go       # Scheme registration for all CRDs
```

**Selected structure** (simpler approach):
```
pkg/k8s/
├── client.go       # Client struct, NewClient(), kubeconfig loading
└── scheme.go       # Scheme registration (core K8s, OpenShift, OLM only)
```

---

## Detailed Design

### 1. File: `pkg/k8s/client.go`

**Purpose**: Main client creation and kubeconfig loading

**Key Components**:

```go
package k8s

import (
    "context"
    "fmt"
    "os"
    "path/filepath"
    "time"
    
    "k8s.io/client-go/rest"
    "k8s.io/client-go/tools/clientcmd"
    "k8s.io/apimachinery/pkg/runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"
)

// Client wraps the controller-runtime client with additional metadata
type Client struct {
    client.Client
    config *rest.Config
}

// NewClient creates a new Kubernetes client
// Priority: kubeconfigPath (if provided) > KUBECONFIG env > ~/.kube/config
func NewClient(ctx context.Context, kubeconfigPath string) (*Client, error)

// getKubeConfig loads kubeconfig with proper priority
func getKubeConfig(kubeconfigPath string) (*rest.Config, error)
```

**Implementation Details**:

1. **kubeconfig Loading Priority**:
   - If `kubeconfigPath` parameter is provided → use it
   - Else if `KUBECONFIG` env var is set → use it
   - Else → use `~/.kube/config`
   - Return error if none found/accessible

2. **Client Configuration**:
   ```go
   config.Timeout = 30 * time.Second
   config.QPS = 50
   config.Burst = 100
   ```

3. **Scheme Registration**:
   ```go
   scheme := runtime.NewScheme()
   if err := registerSchemes(scheme); err != nil {
       return nil, fmt.Errorf("failed to register schemes: %w", err)
   }
   ```

4. **Error Handling**:
   - Clear error messages for missing kubeconfig
   - Distinguish between "file not found" vs "invalid kubeconfig"
   - Wrap errors with context using `fmt.Errorf` with `%w`

**Questions to Consider**:
- Should we expose the `*rest.Config` for advanced use cases? (YES - store it in Client struct)
- Should we validate cluster connectivity in `NewClient()`? (NO - let first operation fail if needed)

---

### 2. File: `pkg/k8s/scheme.go`

**Purpose**: Register all Kubernetes and CRD schemes

**Key Function**:

```go
package k8s

import (
    "k8s.io/apimachinery/pkg/runtime"
    clientgoscheme "k8s.io/client-go/kubernetes/scheme"
    
    // OpenShift APIs
    configv1 "github.com/openshift/api/config/v1"
    routev1 "github.com/openshift/api/route/v1"
    consolev1 "github.com/openshift/api/console/v1"
    
    // Add more as needed...
)

// registerSchemes registers all required schemes for obstool
func registerSchemes(scheme *runtime.Scheme) error
```

**Schemes to Register** (scoped to core platform only):

**Core Kubernetes, OpenShift, and OLM** (implement now):
1. **Core Kubernetes**: `clientgoscheme.AddToScheme(scheme)` - Deployments, Services, Pods, ConfigMaps, Secrets, etc.
2. **OpenShift Config**: `configv1.AddToScheme(scheme)` - ClusterVersion, OAuth, Infrastructure, etc.
3. **OpenShift Route**: `routev1.AddToScheme(scheme)` - Route resources
4. **OpenShift Console**: `consolev1.AddToScheme(scheme)` - ConsolePlugin, ConsoleCLIDownload
5. **OLM**: `operatorsv1alpha1.AddToScheme(scheme)` - Subscription, ClusterServiceVersion, CatalogSource, OperatorGroup

**Operator CRDs** (NOT registered here - added when implementing specific commands):
- Loki (LokiStack) - added in `pkg/resources/lokistack.go`
- Tempo (TempoStack) - added in `pkg/resources/tempostack.go`
- OpenTelemetry (OpenTelemetryCollector) - added in `pkg/resources/otel.go`
- Prometheus Operator (ServiceMonitor) - added in `pkg/resources/servicemonitor.go`
- Perses (PersesDashboard) - added in `pkg/resources/perses.go`
- Cluster Logging (ClusterLogForwarder) - added in `pkg/resources/clusterlogforwarder.go`
- Network Observability (FlowCollector) - added in `pkg/resources/flowcollector.go`
- ACM (MultiClusterObservability) - added in `pkg/resources/acm.go`
- Observability Operator (UIPlugin) - added in `pkg/resources/uiplugin.go`

**Note**: Each resource implementation will be responsible for registering its own scheme. This keeps the k8s package minimal and decouples resource implementations.

**Implementation Strategy**:
```go
func registerSchemes(scheme *runtime.Scheme) error {
    // Core Kubernetes types (Pod, Service, Deployment, ConfigMap, Secret, etc.)
    if err := clientgoscheme.AddToScheme(scheme); err != nil {
        return fmt.Errorf("failed to add core kubernetes scheme: %w", err)
    }
    
    // OpenShift Config API (ClusterVersion, OAuth, Infrastructure, etc.)
    if err := configv1.AddToScheme(scheme); err != nil {
        return fmt.Errorf("failed to add openshift config scheme: %w", err)
    }
    
    // OpenShift Route API
    if err := routev1.AddToScheme(scheme); err != nil {
        return fmt.Errorf("failed to add openshift route scheme: %w", err)
    }
    
    // OpenShift Console API (ConsolePlugin, ConsoleCLIDownload)
    if err := consolev1.AddToScheme(scheme); err != nil {
        return fmt.Errorf("failed to add openshift console scheme: %w", err)
    }
    
    // OLM API (Subscription, ClusterServiceVersion, CatalogSource, OperatorGroup)
    if err := operatorsv1alpha1.AddToScheme(scheme); err != nil {
        return fmt.Errorf("failed to add olm operators scheme: %w", err)
    }
    
    return nil
}
```

**Note on Operator CRD Schemes**:
- Operator-specific schemes (Loki, Tempo, OTEL, etc.) are NOT registered here
- Each `pkg/resources/*.go` file will handle its own scheme registration
- This pattern:
  - Keeps this package minimal and focused
  - Avoids importing operator dependencies unless needed
  - Makes scheme registration explicit and colocated with resource usage
  - Allows for version-specific scheme handling per resource type

---

### 3. Dependencies to Add

**go.mod additions needed**:

Already present:
- ✅ `sigs.k8s.io/controller-runtime` (v0.24.1)
- ✅ `github.com/openshift/api` (v0.0.0-20260605005319-1194f4c62539)
- ✅ `k8s.io/client-go` (v0.36.0 - indirect, but available)
- ✅ `k8s.io/apimachinery` (v0.36.0 - indirect)

**Need to add for OLM**:
- `github.com/operator-framework/api` (for Subscription, CSV, CatalogSource)

Add later (when implementing specific commands):
- `github.com/grafana/loki/operator/apis` (for LokiStack) - added when implementing logging
- `github.com/grafana/tempo-operator/apis` (for TempoStack) - added when implementing tracing
- `github.com/open-telemetry/opentelemetry-operator/apis` (for OTEL) - added when implementing tracing
- `github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring` (for ServiceMonitor) - added when implementing monitoring
- Others as needed per crd_go_modules_research.md

---

## Implementation Steps

### Step 1: Create `pkg/k8s/client.go`

```go
package k8s

import (
    "context"
    "fmt"
    "os"
    "path/filepath"
    "time"
    
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/client-go/rest"
    "k8s.io/client-go/tools/clientcmd"
    "sigs.k8s.io/controller-runtime/pkg/client"
)

// Client wraps the controller-runtime client with configuration
type Client struct {
    client.Client
    config *rest.Config
}

// NewClient creates a new Kubernetes client using kubeconfig
// Priority: kubeconfigPath parameter > KUBECONFIG env > ~/.kube/config
func NewClient(ctx context.Context, kubeconfigPath string) (*Client, error) {
    // Load kubeconfig
    config, err := getKubeConfig(kubeconfigPath)
    if err != nil {
        return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
    }
    
    // Configure client parameters
    config.Timeout = 30 * time.Second
    config.QPS = 50
    config.Burst = 100
    
    // Create scheme and register types
    scheme := runtime.NewScheme()
    if err := registerSchemes(scheme); err != nil {
        return nil, fmt.Errorf("failed to register schemes: %w", err)
    }
    
    // Create controller-runtime client
    c, err := client.New(config, client.Options{Scheme: scheme})
    if err != nil {
        return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
    }
    
    return &Client{
        Client: c,
        config: config,
    }, nil
}

// getKubeConfig loads kubeconfig with proper priority
func getKubeConfig(kubeconfigPath string) (*rest.Config, error) {
    // Priority 1: Explicit path from flag
    if kubeconfigPath != "" {
        config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
        if err != nil {
            return nil, fmt.Errorf("failed to load kubeconfig from %s: %w", kubeconfigPath, err)
        }
        return config, nil
    }
    
    // Priority 2: KUBECONFIG environment variable
    if kubeconfigEnv := os.Getenv("KUBECONFIG"); kubeconfigEnv != "" {
        config, err := clientcmd.BuildConfigFromFlags("", kubeconfigEnv)
        if err != nil {
            return nil, fmt.Errorf("failed to load kubeconfig from KUBECONFIG env (%s): %w", kubeconfigEnv, err)
        }
        return config, nil
    }
    
    // Priority 3: Default ~/.kube/config
    home, err := os.UserHomeDir()
    if err != nil {
        return nil, fmt.Errorf("failed to get user home directory: %w", err)
    }
    
    defaultPath := filepath.Join(home, ".kube", "config")
    config, err := clientcmd.BuildConfigFromFlags("", defaultPath)
    if err != nil {
        return nil, fmt.Errorf("failed to load kubeconfig from default location (%s): %w. Provide --kubeconfig flag or set KUBECONFIG env", defaultPath, err)
    }
    
    return config, nil
}

// Config returns the underlying rest.Config
// This can be useful for creating other clients or advanced use cases
func (c *Client) Config() *rest.Config {
    return c.config
}
```

### Step 2: Add OLM dependency

```bash
go get github.com/operator-framework/api@latest
```

### Step 3: Create `pkg/k8s/scheme.go`

```go
package k8s

import (
    "fmt"
    
    "k8s.io/apimachinery/pkg/runtime"
    clientgoscheme "k8s.io/client-go/kubernetes/scheme"
    
    // OpenShift APIs
    configv1 "github.com/openshift/api/config/v1"
    routev1 "github.com/openshift/api/route/v1"
    consolev1 "github.com/openshift/api/console/v1"
    
    // OLM APIs
    operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
)

// registerSchemes registers core Kubernetes, OpenShift, and OLM schemes
// Operator-specific CRD schemes (Loki, Tempo, etc.) are registered
// in their respective resource implementation files
func registerSchemes(scheme *runtime.Scheme) error {
    // Core Kubernetes types (Pod, Service, Deployment, ConfigMap, Secret, etc.)
    if err := clientgoscheme.AddToScheme(scheme); err != nil {
        return fmt.Errorf("failed to register core kubernetes scheme: %w", err)
    }
    
    // OpenShift Config API (ClusterVersion, OAuth, Infrastructure, etc.)
    if err := configv1.AddToScheme(scheme); err != nil {
        return fmt.Errorf("failed to register openshift config/v1 scheme: %w", err)
    }
    
    // OpenShift Route API
    if err := routev1.AddToScheme(scheme); err != nil {
        return fmt.Errorf("failed to register openshift route/v1 scheme: %w", err)
    }
    
    // OpenShift Console API (ConsolePlugin, ConsoleCLIDownload)
    if err := consolev1.AddToScheme(scheme); err != nil {
        return fmt.Errorf("failed to register openshift console/v1 scheme: %w", err)
    }
    
    // OLM API (Subscription, ClusterServiceVersion, CatalogSource, OperatorGroup)
    if err := operatorsv1alpha1.AddToScheme(scheme); err != nil {
        return fmt.Errorf("failed to register olm operators/v1alpha1 scheme: %w", err)
    }
    
    return nil
}
```

### Step 4: Verify Dependencies

Run go mod tidy to ensure all dependencies are properly resolved:

```bash
go mod tidy
```

### Step 5: Test Compilation

```bash
# Navigate to project root
cd /home/pyurkovi/projects/forks/ocp/worktrees/development-tools/gentle-jackal/development-tools

# Compile to check for errors
go build ./pkg/k8s/...
```

---

## Testing Approach

Per the "minimal testing" philosophy and user direction:

**No unit tests for this task.** Manual validation will be performed when integrating with the root command and testing with a real cluster.

---

## Success Criteria

- [ ] `pkg/k8s/client.go` created with `NewClient()` function
- [ ] `pkg/k8s/scheme.go` created with core K8s, OpenShift, and OLM schemes
- [ ] OLM dependency added: `github.com/operator-framework/api`
- [ ] Code compiles without errors: `go build ./pkg/k8s/...`
- [ ] Dependencies resolved: `go mod tidy` completes successfully
- [ ] Clear error messages for missing/invalid kubeconfig
- [ ] Client properly configured (timeout=30s, QPS=50, burst=100)
- [ ] Ready for use in `pkg/context/context.go` (next task)

---

## Follow-up Tasks

After this task is complete:

1. **Scheme registration for operator CRDs** - handled per resource:
   - Each `pkg/resources/*.go` file will register its own schemes
   - Example: `pkg/resources/lokistack.go` registers Loki scheme
   - This keeps dependencies minimal and explicit

2. **Implement version detection** (next in TODO.md):
   - `internal/version/version.go`
   - Depends on this k8s client

3. **Create execution context package** (TODO.md):
   - `pkg/context/context.go`
   - Will use this k8s client

---

## Decisions Made

Based on user requirements and recommendations:

1. **Scheme Registration Scope**: ✅ Core K8s, OpenShift, and OLM only
   - Operator CRDs registered in their respective `pkg/resources/*.go` files

2. **Version-Specific Console API**: ✅ Use latest (`github.com/openshift/api/console/v1`)
   - UIPlugin from COO abstracts version differences

3. **Error on Missing Kubeconfig**: ✅ Fail fast in `NewClient()` with clear error message

4. **Cluster Connectivity Validation**: ✅ No validation in `NewClient()`
   - Keeps NewClient() fast, let first operation fail if needed

5. **File Structure**: ✅ Simple structure: `client.go` + `scheme.go`

6. **Unit Tests**: ✅ Skip for now per minimal testing philosophy

---

## Files to Create

```
pkg/k8s/
├── client.go    # ~100 lines
└── scheme.go    # ~50 lines
```

**Total**: ~150 lines of code

---

## Estimated Effort

- **Code Writing**: 30-45 minutes
- **Testing/Debugging**: 15-30 minutes (compile + manual verification)
- **Documentation**: Included in code comments
- **Total**: ~1 hour

---

## References

- **CONTEXT.md**: Execution context pattern, critical technical details
- **go-migration-plan.md**: Lines 370-441 (Kubernetes Client Strategy)
- **crd_go_modules_research.md**: Full list of CRD modules needed
- **TODO.md**: Lines 38-48 (Task definition and subtasks)
- **controller-runtime docs**: https://pkg.go.dev/sigs.k8s.io/controller-runtime

---

## Implementation Approved

All decisions confirmed:

1. ✅ File structure: `client.go` + `scheme.go`
2. ✅ Core schemes only: Kubernetes, OpenShift, OLM
3. ✅ Operator CRD schemes handled per resource file
4. ✅ Skip unit tests
5. ✅ Fail fast on missing kubeconfig
6. ✅ Use `github.com/openshift/api/console/v1`
7. ✅ Client config: timeout=30s, QPS=50, burst=100

Ready for implementation.
