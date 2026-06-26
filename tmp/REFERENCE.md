# obstool Quick Reference

**Purpose**: Fast lookup for CRD modules, commands, troubleshooting  
**Target**: Terse reference, not tutorial

---

## CRD Modules (250 lines)

### Observability CRDs

| CRD | Module | Version |
|-----|--------|---------|
| UIPlugin | `github.com/rhobs/observability-operator/pkg/apis/uiplugin/v1alpha1` | v1alpha1 |
| LokiStack | `github.com/grafana/loki/operator/api/loki/v1` | v1 |
| TempoStack | `github.com/grafana/tempo-operator/api/tempo/v1alpha1` | v1alpha1 |
| OpenTelemetryCollector | `github.com/open-telemetry/opentelemetry-operator/apis/v1beta1` | v1beta1 |
| FlowCollector | `github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2` | v1beta2 |
| MultiClusterObservability | `github.com/stolostron/multicluster-observability-operator/apis/observability/v1beta2` | v1beta2 |
| ServiceMonitor | `github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1` | v1 |
| PrometheusRule | `github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1` | v1 |
| PersesDashboard | `github.com/perses/perses/pkg/apis/v1alpha1` | v1alpha1 |
| PersesDatasource | `github.com/perses/perses/pkg/apis/v1alpha1` | v1alpha1 |

**Imports**:
```go
import lokiv1 "github.com/grafana/loki/operator/api/loki/v1"
import tempov1alpha1 "github.com/grafana/tempo-operator/api/tempo/v1alpha1"
import otelv1beta1 "github.com/open-telemetry/opentelemetry-operator/apis/v1beta1"
import flowsv1beta2 "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
import monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
```

### OpenShift Core CRDs

| CRD | Module | Version | Version-Specific |
|-----|--------|---------|------------------|
| Console | See below | v1/v1alpha1 | **YES** |
| Route | `github.com/openshift/api/route/v1` | v1 | No |
| OAuth | `github.com/openshift/api/config/v1` | v1 | No |
| Project | `github.com/openshift/api/project/v1` | v1 | No |
| ImageDigestMirrorSet | `github.com/openshift/api/config/v1` | v1 | No (4.13+) |
| ClusterVersion | `github.com/openshift/api/config/v1` | v1 | No |
| ClusterLogForwarder | `github.com/openshift/cluster-logging-operator/apis/observability/v1` | v1 | No ⚠️ |

⚠️ **ClusterLogForwarder**: From cluster-logging-operator, NOT observability-operator

**Console Version-Specific** (critical):

| OCP Version | Module | Import | Why |
|-------------|--------|--------|-----|
| 4.19+ | `github.com/openshift/api` | `.../console/v1` | Has CSP field |
| 4.17-4.18 | `github.com/rhobs/openshift-api` | `.../console/v1` | Fork without CSP |
| <4.17 | `github.com/rhobs/openshift-api` | `.../console/v1alpha1` | Old API |

**Reason**: OCP 4.19 added ContentSecurityPolicy field that breaks 4.17/4.18

**Imports**:
```go
import configv1 "github.com/openshift/api/config/v1"
import routev1 "github.com/openshift/api/route/v1"
import observabilityv1 "github.com/openshift/cluster-logging-operator/apis/observability/v1"

// Console - version-dependent
import osv1 "github.com/openshift/api/console/v1"                    // 4.19+
import osRhobsv1 "github.com/rhobs/openshift-api/console/v1"         // 4.17-4.18
import osv1alpha1 "github.com/rhobs/openshift-api/console/v1alpha1"  // <4.17
```

### OLM CRDs

| CRD | Module | Version |
|-----|--------|---------|
| Subscription | `github.com/operator-framework/api/pkg/operators/v1alpha1` | v1alpha1 |
| OperatorGroup | `github.com/operator-framework/api/pkg/operators/v1` | v1 |
| CatalogSource | `github.com/operator-framework/api/pkg/operators/v1alpha1` | v1alpha1 |
| ClusterServiceVersion | `github.com/operator-framework/api/pkg/operators/v1alpha1` | v1alpha1 |
| InstallPlan | `github.com/operator-framework/api/pkg/operators/v1alpha1` | v1alpha1 |

**Imports**:
```go
import olmv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
import olmv1 "github.com/operator-framework/api/pkg/operators/v1"
```

### Version Detection

**Get OCP Version**:
```go
cv := &configv1.ClusterVersion{}
c.Get(ctx, client.ObjectKey{Name: "version"}, cv)
version := cv.Status.Desired.Version  // "4.17.1"
```

**Compare Versions**:
```go
import "golang.org/x/mod/semver"

func isVersionAheadOrEqual(current, target string) bool {
    if !strings.HasPrefix(current, "v") { current = "v" + current }
    if !strings.HasPrefix(target, "v") { target = "v" + target }
    return semver.Compare(current, target) >= 0
}
```

**Conditional Resources**:
```go
if isVersionAheadOrEqual(version, "4.19") {
    plugin := &osv1.ConsolePlugin{...}
} else if isVersionAheadOrEqual(version, "4.17") {
    plugin := &osRhobsv1.ConsolePlugin{...}
} else {
    plugin := &osv1alpha1.ConsolePlugin{...}
}
```

### Compatibility Matrix

| CRD | K8s | OCP | Status |
|-----|-----|-----|--------|
| ServiceMonitor | 1.16+ | 4.6+ | Stable |
| LokiStack | 1.21+ | 4.10+ | Stable |
| TempoStack | 1.25+ | 4.12+ | v1alpha1 |
| OpenTelemetryCollector | 1.25+ | 4.12+ | v1beta1 |
| FlowCollector | 1.23+ | 4.12+ | v1beta2 |
| MultiClusterObservability | 1.23+ | 4.11+ | v1beta2, needs ACM |
| PersesDashboard | 1.23+ | 4.12+ | v1alpha1, unstable |
| Console (v1) | - | 4.17+ | Stable |
| Route | - | 4.6+ | Stable |
| ImageDigestMirrorSet | - | 4.13+ | Newer |
| Subscription | 1.16+ | 4.6+ | Stable |
| OperatorGroup | 1.16+ | 4.6+ | Stable |

### Module Selection

**OpenShift**:
- Console → Check OCP version (see table above)
- All others → `github.com/openshift/api`

**Observability**:
- Prometheus → `prometheus-operator/prometheus-operator/pkg/apis/monitoring`
- Loki → `grafana/loki/operator/api/loki/v1`
- Tempo → `grafana/tempo-operator/api/tempo/v1alpha1`
- OTel → `open-telemetry/opentelemetry-operator/apis/v1beta1`
- NetObserv → `netobserv/network-observability-operator/apis`
- Perses → `perses/perses/pkg/apis/v1alpha1`

**OLM**: All → `operator-framework/api`

---

## Commands (150 lines)

### Structure

```
obstool <command> [subcommand] [flags]
```

**Mode**:
- All required flags → CLI mode (silent, minimal output)
- Missing flags → TUI mode (interactive forms, progress)

### Global Flags

```
--kubeconfig string    Kubeconfig path (default: $KUBECONFIG or ~/.kube/config)
--context string       Kubeconfig context
--namespace string     Default namespace
--timeout duration     Timeout (default: 30s)
--verbose              Verbose logging
--debug                Debug logging
```

### version

```bash
obstool version
```

Displays: version, Go version, commit, build date

### deploy

Deploy components to cluster

#### deploy coo

```bash
obstool deploy coo --method=<bundle|fbc|stage|operatorhub> [flags]
```

**Flags**:
```
--method string        Deployment method (required)
--bundle-url string    Bundle image URL (bundle method)
--fbc-url string       FBC image URL (fbc method)
--version string       COO version (stage/operatorhub)
--namespace string     Namespace (default: openshift-observability-operator)
--channel string       Subscription channel (operatorhub)
```

**Examples**:
```bash
obstool deploy coo --method=bundle --bundle-url=quay.io/rhobs/coo-bundle:v1.5.0
obstool deploy coo --method=fbc --fbc-url=quay.io/rhobs/coo-fbc:latest
obstool deploy coo --method=stage --version=v1.5.0
obstool deploy coo --method=operatorhub --channel=stable
```

#### deploy logging

```bash
obstool deploy logging [--data-model=<otel|viaq>] [--size=<size>] [flags]
```

**Flags**:
```
--data-model string    otel or viaq (default: otel)
--size string          1x.extra-small, 1x.small, 1x.medium (default: 1x.extra-small)
--storage-class string StorageClass
--namespace string     Namespace (default: openshift-logging)
```

#### deploy tracing

```bash
obstool deploy tracing [--size=<size>] [flags]
```

**Flags**:
```
--size string          1x.extra-small, 1x.small, 1x.medium (default: 1x.extra-small)
--storage-class string StorageClass
--namespace string     Namespace (default: openshift-tempo-operator)
```

#### deploy dashboards

```bash
obstool deploy dashboards [--dashboards=<list>] [flags]
```

**Flags**:
```
--namespace string     Namespace (default: perses)
--dashboards strings   Comma-separated names (default: all)
```

### cleanup

Remove or scale down components

#### cleanup monitoring

Scale up CMO (restore default plugin)

```bash
obstool cleanup monitoring
```

#### cleanup coo

```bash
obstool cleanup coo [--confirm] [flags]
```

**Flags**:
```
--namespace string     Namespace (default: openshift-observability-operator)
--confirm              Skip confirmation
```

#### cleanup logging

```bash
obstool cleanup logging [--confirm] [flags]
```

#### cleanup tracing

```bash
obstool cleanup tracing [--confirm] [flags]
```

#### cleanup all

```bash
obstool cleanup all --confirm=yes
```

**Note**: Flag mode only. TUI NOT available. Requires `--confirm=yes`.

### update

Update or reconfigure components

#### update monitoring

Scale down CMO, update plugin image

```bash
obstool update monitoring --image=<image-url>
```

#### update coo

```bash
obstool update coo --to-version=<version> [flags]
```

**Flags**:
```
--to-version string    Target version (required)
--namespace string     Namespace (default: openshift-observability-operator)
```

### users

Manage test users and RBAC

#### users create

```bash
obstool users create [--count=<n>] [--prefix=<prefix>] [--password=<pw>]
```

**Flags**:
```
--count int            User count (default: 6)
--prefix string        Username prefix (default: "testuser")
--password string      Password (default: "password")
```

#### users rbac

```bash
obstool users rbac --scenario=<scenario>
```

**Flags**:
```
--scenario string      RBAC scenario (required)
                       Options: perses-e2e, admin-only, viewer-only
```

---

## Troubleshooting (100 lines)

### CRD Issues

**Error**: `no matches for kind ConsolePlugin in version console.openshift.io/v1`

**Fix**: Console v1 requires OCP 4.17+. Check version, use v1alpha1 if <4.17.

---

**Error**: `unknown field spec.contentSecurityPolicy`

**Fix**: Using upstream Console v1 on OCP 4.17/4.18. Use `github.com/rhobs/openshift-api`.

---

**Error**: `the server could not find the requested resource`

**Fix**:
1. Check CRD: `oc get crd <name>`
2. Check operator: `oc get csv -A | grep <operator>`
3. Verify API version

---

**Error**: Module conflicts `openshift/api` vs `rhobs/openshift-api`

**Fix**: These coexist - different import paths. Import correct one per file.

---

### Deployment Issues

**Error**: LokiStack storage errors

**Fix**:
- Check StorageClasses: `oc get storageclass`
- Use `--storage-class` flag
- Verify PVCs: `oc get pvc -n openshift-logging`

---

**Error**: COO bundle image pull errors

**Fix**:
- Test pull: `podman pull <bundle-url>`
- Check IDMS/ICSP
- Verify pull secrets: `oc get secret -n openshift-marketplace`

---

**Error**: Subscription stuck UpgradePending

**Fix**:
- Check InstallPlan: `oc get installplan -n <ns>`
- Approve: `oc patch installplan <name> --type=merge -p '{"spec":{"approved":true}}'`
- Or set `installPlanApproval: Automatic`

---

**Error**: CSV never reaches Succeeded

**Fix**:
1. Check pods: `oc get pods -n <ns>`
2. Check logs: `oc logs <pod> -n <ns>`
3. Check events: `oc get events -n <ns> --sort-by='.lastTimestamp'`
4. Verify RBAC

---

**Error**: ClusterLogForwarder not forwarding

**Fix**:
- Check LokiStack: `oc get lokistack -n openshift-logging`
- Check CLF status: `oc get clusterlogforwarder instance -n openshift-logging -o yaml`
- Check collectors: `oc get pods -n openshift-logging`

---

### Connection Issues

**Error**: `Unable to connect to the server: dial tcp: lookup`

**Fix**:
- Verify kubeconfig: `oc whoami`
- Check context: `oc config current-context`
- Use `--kubeconfig` flag

---

**Error**: `You must be logged in to the server (Unauthorized)`

**Fix**:
- Re-login: `oc login <url>`
- Check token: `oc whoami -t`

---

**Error**: Operation timeout

**Fix**:
- Increase: `--timeout=5m`
- Check cluster: `oc adm top nodes`
- Retry with `--verbose`

---

### TUI Issues

**Error**: TUI not rendering

**Fix**:
- Use CLI mode with flags
- Check: `echo $TERM`
- Resize terminal (min 80x24)

---

**Error**: Flags ignored, TUI appears

**Fix**: Missing required flags. Run `--help` to see requirements.

---

### Known Gotchas

⚠️ **ClusterLogForwarder**: `cluster-logging-operator`, NOT `observability-operator`  
⚠️ **Console**: Version-specific modules - always detect version  
⚠️ **cleanup all**: CLI mode only, requires `--confirm=yes`  
⚠️ **Dashboards**: 30+ files in `pkg/resources/dashboards/`  
⚠️ **Auth**: Kubeconfig only, not in-cluster  

### Namespace Defaults

- COO: `openshift-observability-operator`
- Logging: `openshift-logging`
- Tracing: `openshift-tempo-operator`
- Perses: `perses`

### Quick Diagnostics

```bash
# Version
oc get clusterversion version -o jsonpath='{.status.desired.version}'

# Operators
oc get csv -A

# CRDs
oc get crd <name>
oc api-resources | grep <resource>

# Deployment
oc get deployment,pods -n <ns>
oc logs -n <ns> <pod>

# OLM
oc get subscription,installplan -A

# Storage
oc get storageclass
oc get pvc -A

# Events
oc get events -n <ns> --sort-by='.lastTimestamp'
oc get events -A --sort-by='.lastTimestamp' | tail -20
```

---

**Last Updated**: 2026-06-18  
**Version**: 1.0
