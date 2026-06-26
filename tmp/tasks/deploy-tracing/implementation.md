# Deploy Tracing — Implementation Summary

## Status: Complete

`go build ./...` passes with zero errors.

---

## What Was Implemented

### Go Module Dependencies Added

Four new modules were discovered to have non-obvious sub-module structure:

| Module added | Import path used | Notes |
|---|---|---|
| `github.com/grafana/loki/operator/api/loki` | `github.com/grafana/loki/operator/api/loki/v1` | Separate sub-module — LokiStack types are NOT in the main `loki/operator` module |
| `github.com/grafana/tempo-operator` | `github.com/grafana/tempo-operator/api/tempo/v1alpha1` | Path is `api/`, not `apis/` |
| `github.com/open-telemetry/opentelemetry-operator` | `github.com/open-telemetry/opentelemetry-operator/apis/v1beta1` | Standard path |
| `github.com/rhobs/observability-operator/pkg/apis` | `github.com/rhobs/observability-operator/pkg/apis/uiplugin/v1alpha1` | Separate sub-module — UIPlugin types are NOT in the main `observability-operator` module |

---

## Files Created

| File | Description |
|---|---|
| `pkg/resources/tracing_rbac.go` | 6 ClusterRole + 6 ClusterRoleBinding pairs for Tempo tenant readers/writers and OTel collector k8sattributes RBAC |
| `pkg/resources/otel.go` | `CreatePlatformCollector` and `CreateUserCollector` using typed `otelv1beta1.OpenTelemetryCollector` with `AnyConfig` for component sections |
| `pkg/resources/tracing_signals.go` | `CreateHotrodApp`, `CreateK6TracingApp`, `CreateTelemetrygenApp` with namespace creation |
| `tmp/tasks/deploy-tracing/plan.md` | Approved implementation plan |

---

## Files Modified

| File | Change |
|---|---|
| `go.mod` / `go.sum` | Added 4 new module dependencies |
| `pkg/k8s/scheme.go` | Registered 4 new CRD schemes (loki, tempo, otel, uiplugin) |
| `internal/constants/tracing.go` | Added signal app namespaces, collector deployment names, RBAC names, image constants; fixed `TracingUIPluginName` from `"tracing"` to `"distributed-tracing"` |
| `pkg/resources/lokistack.go` | Converted from `unstructured.Unstructured` to typed `lokiv1.LokiStack`; `GetLokiStack` now returns `*lokiv1.LokiStack` |
| `pkg/resources/tempostack.go` | Converted from `unstructured.Unstructured` to typed `tempov1alpha1.TempoStack`; `StorageSize` uses `resource.Quantity`; `GetTempoStack` now returns `*tempov1alpha1.TempoStack` |
| `pkg/resources/uiplugin.go` | Converted from `unstructured.Unstructured` to typed `uipluginv1alpha1.UIPlugin`; `LoggingConfig.LogsLimit` changed from `int` to `int32`; added `CreateTracingUIPlugin` |
| `pkg/resources/delete.go` | All deletions now use typed zero-value structs; added `DeleteTempoStack` and `DeleteOTelCollector`; `DeleteClusterLogForwarder` remains unstructured (CLF module not in go.mod) |
| `pkg/operations/tracing.go` | Added 8 new steps; expanded `DeployTracingConfig` with 4 new boolean fields; added `enableUserWorkloadMonitoring` helper |
| `cmd/deploy/tracing.go` | Added 4 new flags; refactored into `tracingConfigFromFlags` helper; `buildTracingOperationsList` builds TUI list dynamically; TUI form presents all options as `huh.NewConfirm()` questions with no pre-selected defaults |
| `tmp/REFERENCE.md` | Fixed 4 incorrect import paths in CRD Modules section |

---

## Notable Design Decisions

**`ensureNamespaceExists` in tracing_signals.go**: A private package-level function for signal app namespace creation. Does not use the `k8s.EnsureNamespaceWithLabels` helper intentionally — signal app namespaces need no special labels, keeping the call site simple.

**Signal app failures are non-fatal**: `DeploySignals` logs warnings per-app but does not abort. Image pulls for signal apps (especially `telemetrygen`) can be slow; the user may inspect later.

**User workload monitoring is non-fatal**: The ConfigMap patch may fail if the user lacks cluster-admin. The step is always marked `StatusComplete` regardless, with a warning log.

**`DeleteClusterLogForwarder` remains unstructured**: The `cluster-logging-operator` module is not a dependency of this tool. This is the one deliberate exception to the typed-CR rule.

**`AnyConfig.Object` in OTel collectors**: This is the typed API — `otelv1beta1.AnyConfig` is the operator's own type, backed by `map[string]any` with custom JSON marshaling. Using `map[string]any` values within it is not "unstructured" — it is using the typed API correctly.

---

## Dependencies

All new `go.mod` entries:

```
github.com/grafana/loki/operator/api/loki v0.0.0-20260626101003-b8dc2cf317d9
github.com/grafana/tempo-operator v0.20.0
github.com/open-telemetry/opentelemetry-operator v0.148.0
github.com/rhobs/observability-operator v1.5.0
github.com/rhobs/observability-operator/pkg/apis v0.0.0-20260624154304-d8d51802b035
```

---

## Build Status

```
go build ./...   → OK (zero errors)
```

---

## Next Steps

- `cleanup tracing` command (separate task — see TASKS.md)
- TASKS.md status update: mark Deploy Tracing subtasks `[x]`
