# obstool Implementation TODO

> **🤖 AI Agents**: This is the task list. For project context and patterns, see [CONTEXT.md](./CONTEXT.md) first.

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
- [ ] **Implement cleanup coo command**
  - [ ] Create `cmd/cleanup/coo.go`
  - [ ] Delete COO Subscription
  - [ ] Delete CSV
  - [ ] Delete CatalogSource if created
  - [ ] Delete IDMS if created
  - Blocked by: Implement cleanup command group

### Cleanup Logging
- [ ] **Implement cleanup logging command**
  - [ ] Create `cmd/cleanup/logging.go`
  - [ ] Delete UIPlugin
  - [ ] Delete ClusterLogForwarder
  - [ ] Delete LokiStack
  - [ ] Delete MinIO resources
  - Blocked by: Implement cleanup command group

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
- [ ] **Implement users create command**
  - [ ] Create `cmd/users/users.go` command group
  - [ ] Create `cmd/users/create.go`
  - [ ] Add `--count` flag (default 6)
  - [ ] Add `--password` flag (default "password")
  - Blocked by: Implement root command

  - [ ] **Implement htpasswd generation**
    - [ ] Create `pkg/users/htpasswd.go`
    - [ ] Generate htpasswd file with N users
    - [ ] Create Secret in openshift-config namespace
    - Blocked by: Implement users create command

  - [ ] **Implement OAuth configuration**
    - [ ] Create `pkg/users/oauth.go`
    - [ ] Patch OAuth CR to add htpasswd identity provider
    - Blocked by: Implement htpasswd generation

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
- [ ] **Implement deploy command group**
  - [ ] Create `cmd/deploy/deploy.go`
  - [ ] When run without subcommand, show TUI selection
  - Blocked by: Implement root command, Implement deploy selection TUI

### Deploy COO
- [ ] **Implement deploy coo command**
  - [ ] Create `cmd/deploy/coo.go`
  - [ ] Add `--method` flag (bundle, fbc, stage, operatorhub)
  - [ ] Add method-specific flags
  - Blocked by: Implement deploy command group

  - [ ] **Implement COO bundle deployment**
    - [ ] Create `pkg/operators/coo/bundle.go`
    - [ ] Use operator-sdk run bundle pattern
    - [ ] Create IDMS if needed
    - Blocked by: Implement deploy coo command

  - [ ] **Implement COO FBC deployment**
    - [ ] Create `pkg/operators/coo/fbc.go`
    - [ ] Create CatalogSource
    - [ ] Create Subscription
    - Blocked by: Implement deploy coo command

  - [ ] **Implement COO stage deployment**
    - [ ] Create `pkg/operators/coo/stage.go`
    - [ ] Use stage registry
    - [ ] Create CatalogSource for stage
    - Blocked by: Implement deploy coo command

  - [ ] **Implement COO operatorhub deployment**
    - [ ] Create `pkg/operators/coo/operatorhub.go`
    - [ ] Create Subscription to default catalog
    - Blocked by: Implement deploy coo command

### Deploy Logging
- [ ] **Implement deploy logging command**
  - [ ] Create `cmd/deploy/logging.go`
  - [ ] Add `--data-model` flag (otel, viaq)
  - [ ] Add `--namespace` flag
  - [ ] Add `--skip-ui-plugin` flag
  - Blocked by: Implement deploy command group

  - [ ] **Implement MinIO deployment for logging**
    - [ ] Create `pkg/resources/minio.go`
    - [ ] Create MinIO Deployment, Service, PVC, Secret
    - [ ] Implement storage provider interface
    - Blocked by: Implement deploy logging command

  - [ ] **Implement LokiStack deployment**
    - [ ] Create `pkg/resources/lokistack.go`
    - [ ] Function to create LokiStack CR
    - [ ] Configure with MinIO backend
    - Blocked by: Implement MinIO deployment for logging

  - [ ] **Implement ClusterLogForwarder deployment**
    - [ ] Create `pkg/resources/clusterlogforwarder.go`
    - [ ] Support OTEL and ViaQ data models
    - [ ] Forward to LokiStack
    - Blocked by: Implement LokiStack deployment

  - [ ] **Implement Logging UIPlugin deployment**
    - [ ] Create `pkg/resources/uiplugin.go`
    - [ ] Create UIPlugin CR for logging
    - Blocked by: Implement ClusterLogForwarder deployment

### Deploy Tracing
- [ ] **Implement deploy tracing command**
  - [ ] Create `cmd/deploy/tracing.go`
  - [ ] Add `--size` flag
  - [ ] Add `--namespace` flag
  - Blocked by: Implement deploy command group

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
- [ ] **Implement storage provider interface**
  - [ ] Create `pkg/storage/provider.go`
  - [ ] Define `StorageProvider` interface
  - Blocked by: Set up directory structure

  - [ ] **Implement MinIO provider**
    - [ ] Update `pkg/resources/minio.go` to implement interface
    - [ ] Mark for future deprecation in comments
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
