# Deploy Dashboards — Task Plan

**Status**: In Progress  
**Date**: 2026-06-26

---

## Overview

Implement `obstool deploy dashboards` command. Deploys Perses observability dashboards to an OCP cluster using the Perses Go SDK for dashboard definitions (Dashboards as Code) and the perses-operator CRDs for `PersesGlobalDatasource` resources.

---

## Final Design Decisions

### Command Flags

```
obstool deploy dashboards [flags]

  --namespace string    Namespace for PersesDashboard CRs (default: perses-dev)
                        Created if it doesn't exist.
                        perses-operator reconciles the Perses project automatically.
```

- One flag only. Namespace serves as both the K8s namespace and the Perses project.
- No `--datasource` flag — datasource names are hardcoded per dashboard.
- No `--dashboards` flag — all 5 dashboards are always deployed.
- No `--deploy-uiplugin` flag — covered by monitoring-plugin tasks.

### Dashboard Definitions

Written using the Perses Go SDK (`github.com/perses/perses/go-sdk/dashboard`, panels, queries).  
Community-mixins is used as a reference/inspiration only — not imported as a dependency.

| Dashboard | File | Source |
|-----------|------|--------|
| Prometheus Overview | `pkg/resources/dashboards/prometheus_overview.go` | Go SDK (mirrors community-mixins pattern) |
| Thanos Compact | `pkg/resources/dashboards/thanos_compact.go` | Go SDK (mirrors community-mixins pattern) |
| Node Exporter Full | `pkg/resources/dashboards/node_exporter.go` | Go SDK (mirrors community-mixins pattern) |
| ACM Dashboard | `pkg/resources/dashboards/acm.go` | Go SDK — team-specific |
| Perses Sample | `pkg/resources/dashboards/sample.go` | Go SDK — team-specific |

### Datasources

Created as `PersesGlobalDatasource` CRs (cluster-scoped) using typed structs from `rhobs/perses-operator` v1alpha2.

| CR Name | Type | URL |
|---------|------|-----|
| `thanos-querier-datasource` | PrometheusDatasource | `https://thanos-querier.openshift-monitoring.svc.cluster.local:9091` |
| `loki-datasource` | LokiDatasource | `https://logging-loki-gateway-http.openshift-logging.svc.cluster.local:8080/api/logs/v1/application` |
| `tempo-platform` | TempoDatasource | `https://tempo-platform-gateway.openshift-tracing.svc.cluster.local:8080/api/traces/v1/platform/tempo` |

Secrets (`thanos-querier-datasource-secret`, `loki-datasource-secret`, `tempo-platform-secret`) are created by the perses-operator during reconciliation — not created by obstool.

### Perses Project

- K8s namespace `perses-dev` (default) is created if it doesn't exist
- perses-operator automatically reconciles the Perses project when a `PersesDashboard` CR is deployed into the namespace
- No `PersesProject` CR needs to be manually created

---

## New Dependencies

| Module | Version | Purpose |
|--------|---------|---------|
| `github.com/rhobs/perses-operator` | v0.1.10-0.20260518165420-4a0e166ccfca | `PersesDashboard`, `PersesGlobalDatasource` CR types (v1alpha2) |
| `github.com/perses/perses` | latest compatible | Go SDK dashboard/panel/query builders |
| `github.com/perses/plugins` | latest compatible | Panel/query plugin SDKs (prometheus, timeserieschart, statchart, etc.) |

---

## Architecture Note: Go SDK → CR Spec Bridge

The Perses Go SDK builder produces dashboard specs from `github.com/perses/perses/go-sdk`.  
The `PersesDashboard.Spec.Config` is `v1alpha2.Dashboard` embedding `github.com/perses/spec/go/dashboard.Spec`.

If types are not directly assignable, bridge via JSON round-trip:
```go
jsonBytes, _ := json.Marshal(builder.Dashboard.Spec)
var dashSpec persesv1alpha2.Dashboard
json.Unmarshal(jsonBytes, &dashSpec)
```
Both sides produce identical JSON (same Perses schema). Verify with `go mod tidy` during implementation.

---

## Executor Steps

```go
const (
    StepDashboardsEnsureNamespace = iota
    StepDashboardsCreateThanosDatasource
    StepDashboardsCreateLokiDatasource
    StepDashboardsCreateTempoDatasource
    StepDashboardsDeployPrometheusOverview
    StepDashboardsDeployThanosCompact
    StepDashboardsDeployNodeExporter
    StepDashboardsDeployACM
    StepDashboardsDeployPersesSample
)
```

TUI operations list:
```
"Create namespace <namespace>"
"Create PersesGlobalDatasource: thanos-querier-datasource"
"Create PersesGlobalDatasource: loki-datasource"
"Create PersesGlobalDatasource: tempo-platform"
"Deploy dashboard: Prometheus Overview"
"Deploy dashboard: Thanos Compact"
"Deploy dashboard: Node Exporter Full"
"Deploy dashboard: ACM"
"Deploy dashboard: Perses Sample"
```

---

## Files to Create / Modify

| File | Action |
|------|--------|
| `tmp/tasks/deploy-dashboards/plan.md` | This file |
| `internal/constants/dashboards.go` | Create |
| `pkg/k8s/scheme.go` | Update (register perses-operator v1alpha2 scheme) |
| `pkg/resources/perses.go` | Create |
| `pkg/resources/dashboards/prometheus_overview.go` | Create |
| `pkg/resources/dashboards/thanos_compact.go` | Create |
| `pkg/resources/dashboards/node_exporter.go` | Create |
| `pkg/resources/dashboards/acm.go` | Create |
| `pkg/resources/dashboards/sample.go` | Create |
| `pkg/operations/dashboards.go` | Create |
| `cmd/deploy/dashboards.go` | Create |
| `go.mod` / `go.sum` | Update |

---

## Blocked By

- Deploy command group ✅ (already complete)
