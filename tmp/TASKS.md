# obstool Implementation Tasks

> **🤖 AI Agents**: Task breakdown and tracking. See [README.md](./README.md) for project overview, [PATTERNS.md](./PATTERNS.md) for code standards.

**Status Legend**: `[ ]` Todo | `[~]` In Progress | `[x]` Complete

**Note**: Timeline is not a constraint. Focus is on quality and maintainability.

**How to Use This File**:
- Mark tasks `[~]` when you start working on them (only one in progress at a time per section)
- Mark tasks `[x]` when completed
- Check "Blocked by" to understand dependencies
- Subtasks can be worked in parallel if their parent task is in progress

---

## Foundation

### Project Setup
- [x] **Initialize Go module**
  - [x] Create `go.mod` with module name `github.com/observability-ui/development-tools`
  - [x] Set Go version to 1.22 or later (using 1.26.0)
  - Blocked by: None

- [x] **Set up directory structure**
  - [x] Create `cmd/`, `pkg/`, `internal/` directories
  - [x] Create basic file structure per architecture diagram
  - Blocked by: Initialize Go module

- [x] **Add core dependencies**
  - [x] Add Cobra: `github.com/spf13/cobra`
  - [x] Add controller-runtime: `sigs.k8s.io/controller-runtime`
  - [x] Add Bubble Tea: `github.com/charmbracelet/bubbletea`
  - [x] Add OpenShift API: `github.com/openshift/api`
  - Blocked by: Initialize Go module

### Kubernetes Client
- [x] **Implement k8s client package**
  - [x] Create `pkg/k8s/client.go` with Client struct
  - [x] Implement `NewClient()` function with kubeconfig loading
  - [x] Configure timeout, QPS, burst settings
  - [x] Create `pkg/k8s/scheme.go` with core scheme registration
  - [x] Register core Kubernetes types (Pod, Service, Deployment, etc.)
  - [x] Register OpenShift types (ClusterVersion, Route, Console)
  - [x] Register OLM types (Subscription, CSV, CatalogSource)
  - [x] Add `github.com/operator-framework/api` dependency
  - Blocked by: Add core dependencies
  - Note: Operator CRD schemes (Loki, Tempo, etc.) will be registered in their respective `pkg/resources/*.go` files

### Execution Context
- [x] **Create execution context package**
  - [x] Create `pkg/context/context.go`
  - [x] Implement context.WithValue pattern for Client and IsTUI
  - [x] Implement helper functions: WithClient, GetClient, WithTUI, IsTUI
  - Blocked by: Implement k8s client package

### Configuration
- [x] **Create config package**
  - [x] Create `pkg/config/config.go`
  - [x] Define `Config` struct with default values
  - [x] Expose as package-level `Default` variable
  - [x] Define nested structs: `RegistryConfig`, `NamespaceConfig`, `OperatorConfig`, `UserConfig`, `TimeoutConfig`, `StorageConfig`
  - Blocked by: Set up directory structure

### Root Command
- [x] **Implement root command**
  - [x] Create `cmd/root.go`
  - [x] Set up Cobra root command with global flags
  - [x] Add `--kubeconfig` flag
  - [x] Implement execution context creation in PersistentPreRun
  - Blocked by: Add core dependencies, Create execution context package

  - [x] **Implement version command**
    - [x] Create `cmd/version.go`
    - [x] Display obstool version
    - Blocked by: Implement root command

### TUI Framework
- [x] **Create TUI package structure**
  - [x] Create `pkg/tui/models.go` with base model interface
  - [x] Create `pkg/tui/styles.go` with lipgloss styles
  - Blocked by: Add core dependencies

  - [x] **Implement deploy selection TUI**
    - [x] Create `pkg/tui/deploy.go`
    - [x] Multi-select checkbox list for components
    - [x] "Select All" option
    - Blocked by: Create TUI package structure
    - Note: Ready for use in deploy command group

  - [x] **Implement progress TUI**
    - [x] Create `pkg/tui/progress.go`
    - [x] Display operation progress with checkmarks
    - [x] Real-time status updates
    - Blocked by: Create TUI package structure
    - Note: In use by monitoring commands

  - [x] **Implement input collection TUI**
    - [x] Integrated `huh` library for forms (paste support, validation)
    - [x] Replaced custom forms with production-ready components
    - Note: In use by `update monitoring` command

### Mode Detection
- [x] **Implement mode detection utilities**
  - [x] Create utility to check if all required flags are present
  - [x] Create utility to check if running in terminal (isTerminal)
  - [x] Add helpers to determine CLI vs TUI mode
  - Blocked by: Implement root command
  - Note: `pkg/mode/detect.go` with `DetermineMode()` function

### Output Handling
- [x] **Create output package**
  - [x] Create `pkg/output/output.go`
  - [x] Implement `Handler` with mode-aware output
  - [x] Add `Info()`, `Success()`, `Error()` methods
  - [x] Different behavior for CLI vs TUI mode
  - [x] Add `Progress()` method
  - [x] Add `IsTerminal()` helper
  - Blocked by: Create execution context package
  - Note: In use by monitoring commands

---

## Commands - Update/Cleanup (High Priority)

### Update Monitoring
- [x] **Implement update monitoring command**
  - [x] Create `cmd/update/update.go` command group
  - [x] Create `cmd/update/monitoring.go`
  - [x] Scale down CMO deployment to 0 (to allow plugin updates)
  - [x] Add `--image` flag to update monitoring-plugin image
  - [x] Update monitoring-plugin deployment image when flag provided
  - Blocked by: Implement root command

### Cleanup Monitoring
- [x] **Implement cleanup monitoring command**
  - [x] Create `cmd/cleanup/cleanup.go` command group
  - [x] Create `cmd/cleanup/monitoring.go`
  - [x] Scale up CMO deployment to 1 (CMO reconciles and restores plugin)
  - Blocked by: Implement root command

### Update COO
- [ ] **Implement update coo command**
  - [ ] Create `cmd/update/coo.go`
  - [ ] Add `--to-version` flag
  - [ ] Implement update logic similar to operator-sdk bundle-upgrade
  - Blocked by: Implement deploy coo command

  - [ ] **Implement COO update logic**
    - [ ] Create `pkg/operators/coo/update.go`
    - [ ] Update bundle or subscription
    - [ ] Wait for new version to be installed
    - Blocked by: Implement update coo command

### Cleanup COO
- [x] **Implement cleanup coo command**
  - [x] Create `cmd/cleanup/coo.go`
  - [x] Delete COO Subscription
  - [x] Delete CSV
  - [x] Delete CatalogSource if created
  - [x] Delete IDMS if created
  - [x] Optional: Delete OperatorGroup
  - [x] Optional: Delete Namespace
  - Blocked by: Implement cleanup command group ✅
  - Implementation: [tasks/cleanup-coo-logging/plan.md](./tasks/cleanup-coo-logging/plan.md)

### Cleanup Logging
- [x] **Implement cleanup logging command**
  - [x] Create `cmd/cleanup/logging.go`
  - [x] Delete UIPlugin
  - [x] Delete ClusterLogForwarder
  - [x] Delete LokiStack
  - [x] Delete MinIO resources
  - [x] Delete collector RBAC
  - [x] Optional: Delete operators
  - [x] Optional: Delete namespaces
  - Blocked by: Implement cleanup command group ✅
  - Implementation: [tasks/cleanup-coo-logging/plan.md](./tasks/cleanup-coo-logging/plan.md)

### Cleanup Tracing
- [ ] **Implement cleanup tracing command**
  - [ ] Create `cmd/cleanup/tracing.go`
  - [ ] Delete UIPlugin
  - [ ] Delete OpenTelemetry Collectors
  - [ ] Delete TempoStack
  - [ ] Delete MinIO resources
  - Blocked by: Implement cleanup command group

### Cleanup ACM
- [ ] **Implement cleanup acm command**
  - [ ] Create `cmd/cleanup/acm.go`
  - [ ] Delete MultiClusterObservability
  - [ ] Delete MinIO resources
  - Blocked by: Implement cleanup command group

### Cleanup All
- [ ] **Implement cleanup all command**
  - [ ] Create `cmd/cleanup/all.go`
  - [ ] Flag mode only (require `--confirm=yes`)
  - [ ] Run all cleanup commands in reverse order
  - Blocked by: All individual cleanup commands

---

## Commands - Users (High Priority)

### Users Create
- [x] **Implement users create command**
  - [x] Create `cmd/users/users.go` command group
  - [x] Create `cmd/users/create.go`
  - [x] Add `--count` flag (default 6)
  - [x] Add `--password` flag (default "password")
  - [x] Add `--namespace` flag (default "openshift-monitoring")
  - Blocked by: Implement root command ✅
  - Implementation: [tasks/users-create/implementation.md](./tasks/users-create/implementation.md)

  - [x] **Implement htpasswd generation**
    - [x] Create `pkg/users/htpasswd.go`
    - [x] Generate htpasswd file with N users
    - [x] Create Secret in openshift-config namespace
    - Blocked by: Implement users create command ✅

  - [x] **Implement OAuth configuration**
    - [x] Create `pkg/users/oauth.go`
    - [x] Patch OAuth CR to add htpasswd identity provider
    - Blocked by: Implement htpasswd generation ✅

  - [x] **Implement RBAC configuration**
    - [x] Create `pkg/users/rbac.go`
    - [x] Apply varied RBAC to 6 users (per plan)
    - [x] Create custom role for user6
    - [x] Create 15 bindings (8 namespace + 7 cluster)
    - Blocked by: Implement OAuth configuration ✅

### Users RBAC
- [ ] **Implement users rbac command**
  - [ ] Create `cmd/users/rbac.go`
  - [ ] Add `--scenario` flag for different RBAC scenarios
  - [ ] Support scenarios: perses-e2e, basic, etc.
  - Blocked by: Implement users create command

  - [ ] **Implement RBAC resource creation**
    - [ ] Create `pkg/resources/rbac.go`
    - [ ] Generate ClusterRole resources
    - [ ] Generate ClusterRoleBinding resources
    - [ ] Generate RoleBinding resources
    - Blocked by: Implement users rbac command

---

## Commands - Deploy

### Deploy Command Group
- [x] **Implement deploy command group**
  - [x] Create `cmd/deploy/deploy.go`
  - [x] When run without subcommand, show TUI selection
  - [x] Initially shows only COO option (others add themselves when implemented)
  - Blocked by: Implement root command ✅, Implement deploy selection TUI ✅

### Deploy COO
- [x] **Implement deploy coo command**
  - [x] Create `cmd/deploy/coo.go`
  - [x] Add `--method` flag (bundle, fbc, stage, operatorhub)
  - [x] Add method-specific flags
  - [x] CLI and TUI modes both working
  - [x] Bundle method uses operator-sdk via exec.Command
  - [x] Scheduler patching DEFERRED (see Deferred Items below)
  - [x] Perses namespace NOT created (left to dashboards deployment)
  - Blocked by: Implement deploy command group ✅

  - [x] **Implement COO bundle deployment**
    - [x] Create `pkg/operators/coo/bundle.go`
    - [x] Check for operator-sdk binary
    - [x] Use operator-sdk run bundle pattern
    - [x] Create IDMS based on registry type (quay or stage)
    - [x] Version-aware security context (OCP 4.19+)
    - Blocked by: Implement deploy coo command ✅

  - [x] **Implement COO FBC deployment**
    - [x] Create `pkg/operators/coo/fbc.go`
    - [x] Create IDMS for quay
    - [x] Create CatalogSource
    - [x] Create Subscription
    - Blocked by: Implement deploy coo command ✅

  - [x] **Implement COO stage deployment**
    - [x] Create `pkg/operators/coo/stage.go`
    - [x] Use stage registry
    - [x] Create stage IDMS with brew registry
    - [x] Create CatalogSource for stage
    - [x] Create Subscription
    - Blocked by: Implement deploy coo command ✅

  - [x] **Implement COO operatorhub deployment**
    - [x] Create `pkg/operators/coo/operatorhub.go`
    - [x] Create Subscription to default catalog (redhat-operators)
    - Blocked by: Implement deploy coo command ✅

### Deploy Logging
- [x] **Implement deploy logging command (OperatorHub only - both operators)**
  - [x] Create `cmd/deploy/logging.go`
  - [x] Create `pkg/operations/logging.go`
  - [x] Create `pkg/operators/logging/operatorhub.go`
  - [x] Create `pkg/operators/loki/operatorhub.go`
  - [x] Create `internal/constants/logging.go`
  - [x] Add `--logging-channel` and `--loki-channel` flags
  - [x] Installs **cluster-logging** operator in `openshift-logging` namespace
  - [x] Installs **loki-operator** in `openshift-operators-redhat` namespace
  - [x] CLI and TUI modes with proper executor pattern
  - [x] Both operators from redhat-operators catalog
  - [x] 8 deployment steps total (4 per operator)

- [ ] **Deploy logging signal generator (chat app)**
  - [ ] Create `cmd/deploy/signals-logging.go` or add to logging.go
  - [ ] Deploy chat namespace
  - [ ] Deploy chat pod with log generation
  - [ ] Container runs: `while true; do echo "$(date) chat says hello - $i"; i=$((i + 1)); sleep 1; done`
  - [ ] Add security context (runAsNonRoot, drop ALL caps, seccomp RuntimeDefault)
  - Blocked by: Deploy logging command

- [x] **Implement MinIO deployment for logging**
  - [x] Create `pkg/resources/minio.go`
  - [x] Create `internal/constants/minio.go`
  - [x] Create MinIO Deployment, Service, PVC, Secret
  - [x] MinIO secret also created in openshift-logging namespace for LokiStack
  - [x] Configurable StorageClassName
  - [x] Functions: DeployMinIO, CreateMinIOSecret, CreateMinIOService, CreateMinIOPVC, CreateMinIODeployment

- [x] **Implement LokiStack deployment**
  - [x] Create `pkg/resources/lokistack.go`
  - [x] Function to create LokiStack CR
  - [x] Configure with MinIO S3 backend
  - [x] Size: 1x.demo
  - [x] Schema: v12 with effectiveDate 2022-06-01
  - [x] Tenant mode: openshift-logging

- [x] **Implement RBAC for logging collectors**
  - [x] Create `pkg/resources/logging_rbac.go`
  - [x] CreateServiceAccountWithRBAC helper
  - [x] CreateLogCollectorRBAC (collect-application-logs, collect-infrastructure-logs, logging-collector-logs-writer, lokistack-tenant-logs)
  - [x] CreateCollectorRBAC (logging-collector-logs-writer, collect-application-logs, collect-infrastructure-logs, collect-audit-logs)

- [x] **Implement ClusterLogForwarder deployment**
  - [x] Create `pkg/resources/clusterlogforwarder.go`
  - [x] ClusterLogForwarder CR with lokiStack output type
  - [x] ServiceAccount-based authentication
  - [x] Pipelines for application and infrastructure logs
  - [x] TLS with openshift-service-ca.crt ConfigMap

- [x] **Implement Logging UIPlugin deployment**
  - [x] Create `pkg/resources/uiplugin.go`
  - [x] Create UIPlugin CR for logging
  - [x] Configure with LokiStack reference
  - [x] Settings: logsLimit=50, timeout=30s, schema=select, showTimezoneSelector=true

### Deploy Tracing
- [ ] **Implement deploy tracing command**
  - [ ] Create `cmd/deploy/tracing.go`
  - [ ] Add `--size` flag
  - [ ] Add `--namespace` flag
  - Blocked by: Implement deploy command group

- [ ] **Deploy tracing signal generators**
  - [ ] Create `cmd/deploy/signals-tracing.go` or add to tracing.go
  - [ ] Deploy hotrod application (tracing-app-hotrod namespace)
    - [ ] Deployment with jaegertracing/example-hotrod:1.46 image
    - [ ] Service exposing port 8080
    - [ ] Route for external access
    - [ ] Configure OTLP exporter to user-collector.openshift-tracing:4318
  - [ ] Deploy k6-tracing (tracing-app-k6 namespace)
    - [ ] Deployment with ghcr.io/grafana/xk6-client-tracing:v0.0.5
    - [ ] Connect to user-collector.openshift-tracing:4317
  - [ ] Deploy telemetrygen (tracing-app-telemetrygen namespace)
    - [ ] Container 1: good_service (rate=3, child-spans=2)
    - [ ] Container 2: faulty_service (rate=2, child-spans=1, status-code=Error)
    - [ ] Both send to user-collector.openshift-tracing:4317
  - Blocked by: Deploy tracing command

- [ ] **Implement TempoStack deployment**
  - [ ] Create `pkg/resources/tempostack.go`
  - [ ] Function to create TempoStack CR
  - [ ] Configure with MinIO backend
  - Blocked by: Implement deploy tracing command

- [ ] **Implement OpenTelemetry Collector deployment**
  - [ ] Create `pkg/resources/otel.go`
  - [ ] Create platform and user collectors
  - Blocked by: Implement TempoStack deployment

- [ ] **Implement Tracing UIPlugin deployment**
  - [ ] Update `pkg/resources/uiplugin.go`
  - [ ] Add UIPlugin CR for tracing
  - Blocked by: Implement OpenTelemetry Collector deployment

### Deploy Dashboards
- [ ] **Implement deploy dashboards command**
  - [ ] Create `cmd/deploy/dashboards.go`
  - Blocked by: Implement deploy command group

  - [ ] **Implement Perses datasource deployment**
    - [ ] Create `pkg/resources/perses.go`
    - [ ] Create PersesDatasource CRs
    - [ ] Create PersesGlobalDatasource CRs
    - Blocked by: Implement deploy dashboards command

  - [ ] **Implement dashboard definitions**
    - [ ] Create `pkg/resources/dashboards/` directory
    - [ ] Create individual dashboard files (node-exporter.go, prometheus.go, etc.)
    - [ ] 30+ dashboard definitions
    - Blocked by: Implement Perses datasource deployment

  - [ ] **Implement Dashboards UIPlugin deployment**
    - [ ] Update `pkg/resources/uiplugin.go`
    - [ ] Add UIPlugin CR for dashboards
    - Blocked by: Implement dashboard definitions

### Deploy Monitoring
- [ ] **Implement deploy monitoring command**
  - [ ] Create `cmd/deploy/monitoring.go`
  - [ ] Deploy monitoring UIPlugin
  - Blocked by: Implement deploy command group

  - [ ] **Implement Monitoring UIPlugin deployment**
    - [ ] Update `pkg/resources/uiplugin.go`
    - [ ] Add UIPlugin CR for monitoring
    - Blocked by: Implement deploy monitoring command

### Deploy ACM
- [ ] **Implement deploy acm command**
  - [ ] Create `cmd/deploy/acm.go`
  - Blocked by: Implement deploy command group

  - [ ] **Implement MultiClusterObservability deployment**
    - [ ] Create `pkg/resources/acm.go`
    - [ ] Create MultiClusterObservability CR
    - [ ] Configure with MinIO backend
    - Blocked by: Implement deploy acm command

### Deploy Korrel8r
- [ ] **Implement deploy korrel8r command**
  - [ ] Create `cmd/deploy/korrel8r.go`
  - [ ] Deploy logging, network observability
  - Blocked by: Implement deploy logging command

  - [ ] **Implement FlowCollector deployment**
    - [ ] Create `pkg/resources/flowcollector.go`
    - [ ] Create FlowCollector CR for network observability
    - Blocked by: Implement deploy korrel8r command

  - [ ] **Implement dedicated LokiStack for NetObserv**
    - [ ] Update `pkg/resources/lokistack.go`
    - [ ] Support multiple LokiStack instances
    - Blocked by: Implement FlowCollector deployment

### Deploy All
- [ ] **Implement deploy all command**
  - [ ] Create `cmd/deploy/all.go`
  - [ ] Flag mode only (require all flags)
  - [ ] Deploy all components in sequence
  - Blocked by: All individual deploy commands

---

## Supporting Infrastructure

### Storage Provider
- [x] **Implement storage provider interface**
  - [x] Create `pkg/storage/provider.go`
  - [x] Define `StorageProvider` interface
  - Blocked by: Set up directory structure

  - [x] **Implement MinIO provider**
    - [x] Create `pkg/storage/minio.go` to implement interface
    - [x] Creates namespace if missing
    - [x] Waits for deployment to be ready
    - [x] Generic S3-compatible secret format
    - Blocked by: Implement storage provider interface

### OLM Utilities
- [x] **Implement OLM utilities**
  - [x] Create `pkg/operators/olm.go`
  - [x] Create `pkg/operators/subscription.go`
  - [x] Create `pkg/operators/catalogsource.go`
  - [x] Create `pkg/operators/operatorgroup.go`
  - [x] Create `pkg/operators/idms.go`
  - [x] Helper functions for Subscription creation
  - [x] Helper functions for CatalogSource management
  - [x] Helper functions for waiting on CSV installation (with TUI countdown)
  - [x] Helper functions for OperatorGroup management
  - [x] Helper functions for IDMS management
  - [x] Executor integration for progress reporting
  - [x] Update scheme registration for operatorsv1
  - Blocked by: Implement k8s client package (✅ Complete)
  - Implementation: [tmp/tasks/olm-utilities/implementation.md](./tasks/olm-utilities/implementation.md)

---

## Documentation

- [ ] **Create main README for obstool**
  - [ ] Usage examples
  - [ ] Installation instructions
  - [ ] Link to migration plan
  - Blocked by: Implement root command

- [ ] **Add command help text**
  - [ ] Ensure all commands have Short and Long descriptions
  - [ ] Add examples to help text
  - Blocked by: Implement individual commands

- [ ] **Add contribution guide**
  - [ ] How to add new commands
  - [ ] How to add new resources
  - [ ] Coding conventions (minimal comments, no 1-2 letter variables)
  - Blocked by: Create main README for obstool

---

## Build & Release

- [ ] **Set up Makefile**
  - [ ] `make build` - Build binary
  - [ ] `make install` - Install to PATH
  - [ ] `make clean` - Clean build artifacts
  - Blocked by: Initialize Go module

- [ ] **Add shell completion generation**
  - [ ] Generate bash completion
  - [ ] Generate zsh completion
  - Blocked by: Implement root command

- [ ] **Create release process**
  - [ ] Version bumping strategy
  - [ ] Build for multiple platforms
  - [ ] Distribution method
  - Blocked by: Set up Makefile

---

## Deferred Items

### Scheduler Patching
- **Decision**: Removed from initial COO deployment implementation
- **Reason**: Unclear if this should be:
  - Part of COO deployment specifically
  - Global cluster preparation step (separate command?)
  - Optional via flag
  - Part of all operator deployments
- **Action Required**: Determine correct placement and implement
- **Reference**: Bash scripts patch with `kubectl patch Scheduler cluster --type='json' -p '[{ "op": "replace", "path": "/spec/mastersSchedulable", "value": true }]'`
- **Affects**: COO bundle deployment (and possibly others)
- **Status**: [ ] Not implemented

### Binary Dependencies Check
- [ ] **Implement binary dependency checker**
  - [ ] Create `cmd/doctor.go` command or similar
  - [ ] Check for operator-sdk (required for bundle method)
  - [ ] Check for oc/kubectl (general requirement)
  - [ ] Provide installation instructions
  - [ ] Blocked by: None (can implement anytime)
  - [ ] Note: For current implementation, assume operator-sdk installed and error if not found

---

## Migration

- [ ] **Document bash script equivalents**
  - [ ] Map each bash script to obstool command
  - [ ] Create migration guide for team
  - Blocked by: All commands implemented

- [ ] **Deprecate bash scripts**
  - [ ] Add deprecation notices to bash scripts
  - [ ] Point to obstool equivalents
  - Blocked by: Document bash script equivalents

- [ ] **Remove bash scripts**
  - [ ] Remove after team validates obstool
  - [ ] Timeline flexible - quality over speed
  - Blocked by: Deprecate bash scripts
