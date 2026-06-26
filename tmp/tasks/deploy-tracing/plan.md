# Deploy Tracing — Implementation Plan

## Overview

Complete the `obstool deploy tracing` command to achieve full feature parity with `perses/demo/z_tracing.sh`. The command already deploys the Tempo and OpenTelemetry operators; this plan extends it to cover the remaining components: trace-reader RBAC, user workload monitoring, OpenTelemetry Collectors, signal generator apps, and the Distributed Tracing UIPlugin.

This plan also migrates all existing `unstructured.Unstructured` resource definitions to typed CRs using proper Go module dependencies.

## Architecture Context

- Business logic: `pkg/operations/tracing.go` (executor pattern)
- Command layer: `cmd/deploy/tracing.go` (CLI + TUI modes)
- Resource definitions: `pkg/resources/` (flat structure, typed Go structs)
- Constants: `internal/constants/tracing.go`

Reference bash script: `perses/demo/z_tracing.sh`
Reference uninstall: `perses/demo/z_tracing_uninstall.sh`

## Critical Corrections

The following import paths in `tmp/REFERENCE.md` are incorrect and must be fixed:

| Module | REFERENCE.md (wrong) | Actual |
|--------|----------------------|--------|
| LokiStack | `github.com/grafana/loki/operator/apis/loki/v1` | `github.com/grafana/loki/operator/api/loki/v1` |
| TempoStack | `github.com/grafana/tempo-operator/apis/tempo/v1alpha1` | `github.com/grafana/tempo-operator/api/tempo/v1alpha1` |
| UIPlugin | `github.com/rhobs/observability-operator/pkg/apis/ui/v1alpha1` | `github.com/rhobs/observability-operator/pkg/apis/uiplugin/v1alpha1` |

## Scope

### In Scope
- Add 4 Go module dependencies
- Register 4 new CRD schemes
- Convert existing unstructured resources to typed CRs (LokiStack, TempoStack, UIPlugin, delete.go)
- New resource files: otel.go, tracing_rbac.go, tracing_signals.go
- Extend operations/tracing.go with 8 new steps
- Extend cmd/deploy/tracing.go with 4 new flags + TUI form updates
- Fix REFERENCE.md import paths
- Add missing constants

### Out of Scope
- `cleanup tracing` command (separate task)
- Signal generators as a separate command

## Detailed Design

### New Go Module Dependencies

```
github.com/grafana/loki/operator/api             → LokiStack v1
github.com/grafana/tempo-operator/api            → TempoStack v1alpha1
github.com/open-telemetry/opentelemetry-operator → OpenTelemetryCollector v1beta1
github.com/rhobs/observability-operator          → UIPlugin v1alpha1
```

### Scheme Registration (pkg/k8s/scheme.go)

```go
lokiv1.AddToScheme(scheme)
tempov1alpha1.AddToScheme(scheme)
otelv1beta1.AddToScheme(scheme)
uipluginv1alpha1.AddToScheme(scheme)
```

### Typed CR Key Fields

**LokiStack** (`github.com/grafana/loki/operator/api/loki/v1`):
- `SizeOneXExtraSmall` = `"1x.extra-small"`
- `ObjectStorageSchemaV12` = `"v12"` — `EffectiveDate` is type `StorageSchemaEffectiveDate` (string)
- `ObjectStorageSecretS3` = `"s3"`
- `OpenshiftLogging` = `"openshift-logging"`
- `StorageClassName` is plain `string`
- `Tenants` is `*TenantsSpec`

**TempoStack** (`github.com/grafana/tempo-operator/api/tempo/v1alpha1`):
- `ObjectStorageSecretS3` = `"s3"`
- `ModeOpenShift` = `"openshift"`
- `StorageSize` is `resource.Quantity` (use `resource.MustParse("1Gi")`)
- `Tenants` is `*TenantsSpec`
- `TempoGatewaySpec.Enabled` is plain `bool`
- `TracingConfigSpec.OTLPHttpEndpoint` and `SamplingFraction string`
- `JaegerQueryMonitor.PrometheusEndpoint` field name

**UIPlugin** (`github.com/rhobs/observability-operator/pkg/apis/uiplugin/v1alpha1`):
- `TypeLogging` = `"Logging"`
- `TypeDistributedTracing` = `"DistributedTracing"`
- `LoggingConfig.LokiStack` is `*LokiStackReference{Name, Namespace}`
- `LoggingConfig.LogsLimit` is `int32`

**OpenTelemetryCollector** (`github.com/open-telemetry/opentelemetry-operator/apis/v1beta1`):
- `Config.Receivers` and `Config.Exporters` are value types (`AnyConfig`, not pointers)
- `Config.Processors`, `Config.Connectors`, `Config.Extensions` are pointer types (`*AnyConfig`)
- `AnyConfig{Object: map[string]any{...}}` is the typed API for component sections
- `Service.Extensions` is `[]string`
- `Service.Pipelines` is `map[string]*Pipeline`

### New Constants (internal/constants/tracing.go)

```go
TracingUIPluginName         = "distributed-tracing"   // fix from "tracing"
HotrodNamespace             = "tracing-app-hotrod"
K6TracingNamespace          = "tracing-app-k6"
TelemetrygenNamespace       = "tracing-app-telemetrygen"
PlatformCollectorDeployment = "platform-collector"
UserCollectorDeployment     = "user-collector"
TracesReaderPlatformRole    = "traces-reader-platform"
TracesReaderUserRole        = "traces-reader-user"
PlatformCollectorRole       = "openshift-tracing-platform-collector"
UserCollectorRole           = "openshift-tracing-user-collector"
TracesWriterPlatformRole    = "traces-writer-platform"
TracesWriterUserRole        = "traces-writer-user"
HotrodImage                 = "jaegertracing/example-hotrod:1.46"
K6TracingImage              = "ghcr.io/grafana/xk6-client-tracing:v0.0.5"
TelemetrygenImage           = "ghcr.io/open-telemetry/opentelemetry-collector-contrib/telemetrygen:v0.105.0"
```

### New Steps in pkg/operations/tracing.go

Append after `StepTracingWaitForTempoStack`:

```
StepTracingEnableUserWorkloadMonitoring  // patch cluster-monitoring-config
StepTracingCreateTraceReaderRBAC         // ClusterRole/Binding for traces-reader-*
StepTracingDeployPlatformCollector       // OTelCollector CR + RBAC
StepTracingWaitForPlatformCollector      // WaitForDeploymentReady
StepTracingDeployUserCollector           // OTelCollector CR + RBAC
StepTracingWaitForUserCollector          // WaitForDeploymentReady
StepTracingDeploySignalApps             // hotrod, k6, telemetrygen
StepTracingDeployUIPlugin               // UIPlugin distributed-tracing
```

Updated `DeployTracingConfig`:
```go
type DeployTracingConfig struct {
    TempoChannel                 string
    OTelChannel                  string
    StorageClassName             string
    DeployMinIO                  bool
    DeployTempoStack             bool
    EnableUserWorkloadMonitoring bool  // new
    DeployCollectors             bool  // new
    DeploySignals                bool  // new
    DeployUIPlugin               bool  // new
}
```

New step conditions:
- `EnableUserWorkloadMonitoring` → patch ConfigMap; non-fatal on RBAC error
- `DeployTempoStack` gates trace reader RBAC (requires TempoStack to be deployed)
- `DeployCollectors` → platform + user collectors with waits
- `DeploySignals` → 3 signal apps (non-fatal per app)
- `DeployUIPlugin` → distributed-tracing UIPlugin

### New Flags in cmd/deploy/tracing.go

```
--enable-user-workload-monitoring   (default: false)
--deploy-collectors                 (default: false)
--deploy-signals                    (default: false)
--deploy-uiplugin                   (default: false)
```

TUI: all flags presented as `huh.NewConfirm()` questions in `collectTracingInput`. No pre-selected defaults — user explicitly chooses. `operationsList` is built dynamically based on chosen options.

### New Resource Files

**pkg/resources/tracing_rbac.go**
- `CreateTracingRBAC(ctx, kubeClient, namespace)` — 6 ClusterRole + 6 ClusterRoleBinding
- `DeleteTracingRBAC(ctx, kubeClient)` — deletes by name list

**pkg/resources/otel.go**
- `CreatePlatformCollector(ctx, kubeClient, namespace)` — named "platform"
- `CreateUserCollector(ctx, kubeClient, namespace)` — named "user"
- Private helper to build shared collector config, parameterized by tenant name

**pkg/resources/tracing_signals.go**
- `CreateHotrodApp(ctx, kubeClient)` — Deployment + Service + Route in tracing-app-hotrod
- `CreateK6TracingApp(ctx, kubeClient)` — Deployment in tracing-app-k6
- `CreateTelemetrygenApp(ctx, kubeClient)` — Deployment (2 containers) in tracing-app-telemetrygen

## Implementation Order

```
1.  go get (4 modules)
2.  pkg/k8s/scheme.go                  → 4 new scheme registrations
3.  internal/constants/tracing.go      → add constants, fix UIPlugin name
4.  pkg/resources/lokistack.go         → typed lokiv1.LokiStack
5.  pkg/resources/tempostack.go        → typed tempov1alpha1.TempoStack
6.  pkg/resources/uiplugin.go          → typed UIPlugin + CreateTracingUIPlugin
7.  pkg/resources/delete.go            → typed deletion + DeleteTempoStack + DeleteOTelCollector
8.  pkg/resources/tracing_rbac.go      → new file
9.  pkg/resources/otel.go              → new file
10. pkg/resources/tracing_signals.go   → new file
11. pkg/operations/tracing.go          → new steps + updated config struct
12. cmd/deploy/tracing.go              → new flags + TUI updates
13. tmp/REFERENCE.md                   → fix import paths
14. go build ./... (verify)
```

## Success Criteria

- `go build ./...` compiles without errors
- `obstool deploy tracing` deploys operators only (no flags)
- `obstool deploy tracing --deploy-minio --deploy-tempostack --deploy-collectors --deploy-uiplugin` deploys full stack
- `--deploy-signals` deploys hotrod, k6-tracing, telemetrygen in separate namespaces
- All resource files use typed CRs (no `unstructured.Unstructured` in resource definitions)
- TUI mode presents all options when no flags are provided

## References

- Bash script: `perses/demo/z_tracing.sh`
- Uninstall script: `perses/demo/z_tracing_uninstall.sh`
- Manifests: `tracing-manifests/base/`
- QE test: `qe-release-testing/uiplugins/distributed_tracing.sh`
- Patterns: `tmp/PATTERNS.md`
