# Go Modules Research for Kubernetes CRDs

> **🤖 AI Agents**: Reference guide for CRD Go modules and version compatibility. Check [CONTEXT.md](./CONTEXT.md) for project overview first.

Based on the pattern from `rhobs/observability-operator` PR #1100, this document identifies Go modules needed for working with various Kubernetes CRDs commonly used in observability stacks.

**Purpose**: Version-specific module selection guide for 18 CRD types used in obstool

---

## Quick CRD Lookup

**Observability CRDs**:
- UIPlugin → `github.com/openshift/api/console/v1` (or rhobs fork for 4.17-4.18)
- LokiStack → `github.com/grafana/loki/operator/apis/loki/v1`
- TempoStack → `github.com/grafana/tempo-operator/apis/tempo/v1alpha1`
- PersesDashboard → `github.com/perses/perses-operator/api/v1alpha1`
- ClusterLogForwarder → `github.com/openshift/cluster-logging-operator/apis/observability/v1` ⚠️

**OpenShift Core**:
- Route → `github.com/openshift/api/route/v1`
- OAuth → `github.com/openshift/api/config/v1`
- ClusterVersion → `github.com/openshift/api/config/v1`

**OLM**:
- Subscription → `github.com/operator-framework/api/pkg/operators/v1alpha1`

---

## Executive Summary

### Key Pattern Learned from PR #1100

**Issue**: Different OpenShift versions may require different Go module packages even for the same API version.

**Example**: The `console/v1` API requires:
- `github.com/rhobs/openshift-api/console/v1` for OpenShift 4.17-4.18
- `github.com/openshift/api/console/v1` for OpenShift 4.19+

**Reason**: Newer versions of upstream modules may add fields (like `ContentSecurityPolicy` in console/v1) that aren't backward compatible with older clusters. These new fields are marshalled by default even when unset, causing reconciliation errors on older clusters.

**Solution**: 
1. Version detection at runtime
2. Conditional imports based on cluster version
3. Forked modules when necessary for backward compatibility

---

## Observability CRDs

### 1. UIPlugin (observability.openshift.io/v1alpha1)

**Go Module**: `github.com/rhobs/observability-operator/pkg/apis/ui/v1alpha1`

**Source Repository**: https://github.com/rhobs/observability-operator

**Version-Specific Modules**: No - this is a custom CRD specific to the observability operator

**Usage Example** (from go.mod):
```go
github.com/rhobs/observability-operator/pkg/apis v0.0.0-20251009091129-76135c924ed6
```

**Notes**:
- This is an internal CRD to the observability-operator
- Not a standard Kubernetes/OpenShift resource
- No upstream/forked versions needed

---

### 2. ClusterLogForwarder (observability.openshift.io/v1)

**Go Module**: `github.com/openshift/cluster-logging-operator/apis/observability/v1`

**Source**: OpenShift Cluster Logging Operator

**Source Repository**: https://github.com/openshift/cluster-logging-operator

**Version-Specific Modules**: No

**Notes**:
- **Correction**: ClusterLogForwarder is NOT internal to observability-operator
- It is part of the **OpenShift Cluster Logging Operator**
- Used to forward logs from the cluster to various outputs (including Loki)
- The observability-operator uses this CRD but doesn't own it

**Usage Pattern**:
```go
import observabilityv1 "github.com/openshift/cluster-logging-operator/apis/observability/v1"

clf := &observabilityv1.ClusterLogForwarder{
    ObjectMeta: metav1.ObjectMeta{
        Name:      "instance",
        Namespace: "openshift-logging",
    },
    Spec: observabilityv1.ClusterLogForwarderSpec{
        // Forwarder configuration
    },
}
```

---

### 3. LokiStack (loki.grafana.com/v1)

**Actual Operator**: The LokiStack CRD is provided by the **Loki Operator**, not the main Loki repository

**Go Module Path**: `github.com/grafana/loki/operator/apis/loki/v1`

**Alternative Repo**: Look for `github.com/grafana/loki-operator` or check within the Loki monorepo under `/operator`

**Source Repository**: https://github.com/grafana/loki (contains the operator under `/operator` directory)

**Version-Specific Modules**: No known version-specific variants

**Compatibility Notes**:
- The Loki Operator manages LokiStack CRDs
- Check operator version compatibility with your Kubernetes/OpenShift version
- Loki repository has operator code in the `/operator` subdirectory

**Usage Pattern**:
```go
// Import from the operator API subdirectory
import lokiv1 "github.com/grafana/loki/operator/apis/loki/v1"
```

**References**:
- Operator location: https://github.com/grafana/loki/tree/main/operator
- Full repository: https://github.com/grafana/loki

---

### 4. TempoStack (tempo.grafana.com/v1alpha1)

**Go Module**: `github.com/grafana/tempo-operator/apis/tempo/v1alpha1`

**Source Repository**: https://github.com/grafana/tempo-operator

**Version-Specific Modules**: No

**Usage Example** (from go.mod):
```go
github.com/grafana/tempo-operator v0.20.0
```

**Compatibility Notes**:
- Current version in observability-operator: v0.20.0
- Active development on tempo-operator
- Supports Kubernetes 1.25+ based on go.mod

**References**:
- API Docs: https://github.com/grafana/tempo-operator/tree/main/apis
- Release Notes: https://github.com/grafana/tempo-operator/releases

---

### 5. OpenTelemetryCollector (opentelemetry.io/v1beta1)

**Go Module**: `github.com/open-telemetry/opentelemetry-operator/apis/v1beta1`

**Source Repository**: https://github.com/open-telemetry/opentelemetry-operator

**Version-Specific Modules**: No - uses semantic versioning

**Current Version** (from observability-operator):
```go
github.com/open-telemetry/opentelemetry-operator v0.148.0
```

**API Versions Available**:
- `v1alpha1` - deprecated/older features
- `v1beta1` - current stable API (OpenTelemetryCollector)

**Compatibility Notes**:
- Very active project with frequent releases
- Supports Kubernetes 1.25-1.35+ (based on kind configs in repo)
- Also provides other CRDs: `Instrumentation`, `OpAMPBridge`

**Usage Pattern**:
```go
import otelv1beta1 "github.com/open-telemetry/opentelemetry-operator/apis/v1beta1"
```

**References**:
- Documentation: https://opentelemetry.io/docs/kubernetes/operator/
- API Reference: https://github.com/open-telemetry/opentelemetry-operator/tree/main/apis

---

### 6. FlowCollector (flows.netobserv.io/v1beta2)

**Go Module**: `github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2`

**Source Repository**: https://github.com/netobserv/netobserv-operator

**Version-Specific Modules**: No - uses API versioning (v1beta2 is current)

**API Versions**:
- `v1alpha1` - deprecated
- `v1beta1` - older stable
- `v1beta2` - current stable (as of 1.11.x)

**Compatibility Notes**:
- Also provides `FlowMetrics` CRD for custom metrics
- Network observability for Kubernetes/OpenShift
- Built on eBPF for flow collection

**Usage Pattern**:
```go
import flowsv1beta2 "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
```

**Module Path Note**: Check if module is at:
- `github.com/netobserv/netobserv-operator` (repo name)
- OR possibly exported separately

---

### 7. MultiClusterObservability (observability.open-cluster-management.io/v1beta2)

**Go Module**: `github.com/stolostron/multicluster-observability-operator/apis/observability/v1beta2`

**Alternative Module Path**: `open-cluster-management.io/multicluster-observability-operator/apis/observability/v1beta2`

**Source Repository**: https://github.com/stolostron/multicluster-observability-operator

**Version-Specific Modules**: No - uses API versioning

**API Versions**:
- `v1beta1` - older API
- `v1beta2` - current API

**Compatibility Notes**:
- Part of the **stolostron** (STOrage LOcal STORage Network) / Open Cluster Management ecosystem
- Requires Advanced Cluster Management (ACM) / Open Cluster Management installed
- Integrates with Thanos for multi-cluster metric aggregation

**Additional CRDs in this Operator**:
- `ObservabilityAddon` - per-cluster observability configuration
- May have internal APIs not exported

**Usage Pattern**:
```go
import mcoav1beta2 "github.com/stolostron/multicluster-observability-operator/apis/observability/v1beta2"
```

**References**:
- Architecture: https://github.com/stolostron/multicluster-observability-operator#architecture
- API Location: https://github.com/stolostron/multicluster-observability-operator/tree/main/operators

---

### 8. ServiceMonitor (monitoring.coreos.com/v1)

**Go Module**: `github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1`

**Source Repository**: https://github.com/prometheus-operator/prometheus-operator

**Version-Specific Modules**: No - mature stable API

**Current Version** (from observability-operator):
```go
github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.91.0
```

**Additional CRDs in Same Module**:
- `Prometheus`
- `Alertmanager`
- `PrometheusRule`
- `PodMonitor`
- `Probe`
- `ThanosRuler`
- `ScrapeConfig` (v1alpha1)
- `AlertmanagerConfig` (v1beta1)

**Compatibility Notes**:
- Very mature and stable project (v0.91.0 as of May 2026)
- De facto standard for Prometheus on Kubernetes
- All monitoring.coreos.com CRDs come from this operator

**Usage Pattern**:
```go
import monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
```

**Alternative/Forked Versions**:
- `github.com/rhobs/obo-prometheus-operator/pkg/apis/monitoring` - Red Hat fork with additional features
- Version from observability-operator go.mod: `v0.91.0-rhobs1`

**When to use the fork**:
- If you need Red Hat-specific extensions
- For compatibility with OpenShift operator ecosystem
- Generally the upstream version is preferred unless you have specific needs

---

### 9. PersesDashboard (perses.dev/v1alpha1, v1alpha2)

**Go Module**: `github.com/perses/perses/pkg/apis/v1alpha1`

**Alternative Operator Module**: `github.com/rhobs/perses-operator` (used in observability-operator)

**Source Repositories**: 
- Main Project: https://github.com/perses/perses
- Operator: https://github.com/rhobs/perses-operator (RHOBS fork)

**Version-Specific Modules**: Yes - API is evolving

**API Versions**:
- `v1alpha1` - initial API
- `v1alpha2` - current development (not yet stable)

**Current Usage** (from observability-operator go.mod):
```go
github.com/rhobs/perses v0.0.0-20260422074433-2c06d5cd1312  // forked perses
github.com/rhobs/perses-operator v0.1.10-0.20260518165420-4a0e166ccfca
github.com/perses/perses v0.53.1
```

**Compatibility Notes**:
- Perses is a newer project (alternative to Grafana dashboards)
- API is still in alpha - expect breaking changes
- RHOBS maintains a fork for operator integration
- Plan is to migrate NetObserv dashboards to Perses for OpenShift console integration

**Usage Pattern**:
```go
// For core Perses types
import persesv1alpha1 "github.com/perses/perses/pkg/apis/v1alpha1"

// For operator CRDs (if using the operator)
import persesopv1alpha1 "github.com/rhobs/perses-operator/api/v1alpha1"
```

**Alternative**: If not using RHOBS operator, use upstream operator when available

---

### 10. PersesDatasource (perses.dev/v1alpha1, v1alpha2)

**Go Module**: Same as PersesDashboard above

**Module**: `github.com/perses/perses/pkg/apis/v1alpha1`

**Operator Module**: `github.com/rhobs/perses-operator/api/v1alpha1`

**Source**: Same repositories as PersesDashboard

**CRDs Provided**:
- `PersesDashboard`
- `PersesDatasource` 
- `PersesGlobalDatasource` (cluster-scoped)

**Notes**:
- Same compatibility and version considerations as PersesDashboard
- Part of the same API group
- Used for configuring data sources (Prometheus, etc.) for Perses dashboards

---

## OpenShift Resources

### 11. OAuth (config.openshift.io/v1)

**Go Module**: `github.com/openshift/api/config/v1`

**Source Repository**: https://github.com/openshift/api

**Version-Specific Modules**: No standard version-specific modules, but see notes below

**Current Version** (from observability-operator):
```go
github.com/openshift/api v0.0.0-20260511191110-9b69e5fa27e9
```

**Version Compatibility Pattern**:
- OpenShift uses date-based pseudo-versions for the API module
- The version shown above corresponds to a specific commit
- OpenShift 4.x versions typically compatible with latest `openshift/api`

**Compatibility Notes**:
- This module contains ALL OpenShift config APIs
- Other resources in `config.openshift.io/v1`:
  - `ClusterVersion`
  - `Infrastructure`
  - `Network`
  - `Ingress`
  - Many more cluster-wide config CRDs

**Usage Pattern**:
```go
import configv1 "github.com/openshift/api/config/v1"

// Access OAuth
oauth := &configv1.OAuth{}
```

**Important**: See Console resource section below for version-specific considerations

---

### 12. Route (route.openshift.io/v1)

**Go Module**: `github.com/openshift/api/route/v1`

**Source Repository**: https://github.com/openshift/api

**Version-Specific Modules**: **YES** - Same pattern as Console

**Standard Module**:
```go
github.com/openshift/api v0.0.0-20260511191110-9b69e5fa27e9
```

**Forked Module** (if needed for older OpenShift):
```go
github.com/rhobs/openshift-api v0.0.0-20260512142436-2e89e902a420
```

**Version-Specific Considerations**:
- Routes are one of the oldest and most stable OpenShift APIs
- Generally **NO** version-specific issues unlike Console
- The upstream `github.com/openshift/api` should work for all supported versions

**Compatibility**:
- OpenShift 4.6+ - fully stable
- No known breaking changes across versions
- Safe to use latest `openshift/api` module

**Usage Pattern**:
```go
import routev1 "github.com/openshift/api/route/v1"

route := &routev1.Route{
    // Route definition
}
```

**When to Use Forked Version**:
- Only if you're using it alongside Console and need consistency
- Generally **not needed** for Route specifically

---

### 13. Project (project.openshift.io/v1)

**Go Module**: `github.com/openshift/api/project/v1`

**Source Repository**: https://github.com/openshift/api

**Version-Specific Modules**: No - stable API

**Current Version**:
```go
github.com/openshift/api v0.0.0-20260511191110-9b69e5fa27e9
```

**Compatibility Notes**:
- Projects are OpenShift's extension of Kubernetes Namespaces
- Very stable API since OpenShift 3.x
- No breaking changes in recent versions

**Related CRDs** in same API group:
- `ProjectRequest` - for requesting new projects

**Usage Pattern**:
```go
import projectv1 "github.com/openshift/api/project/v1"

project := &projectv1.Project{
    // Project definition  
}
```

---

### 14. ImageDigestMirrorSet (config.openshift.io/v1)

**Go Module**: `github.com/openshift/api/config/v1`

**Source Repository**: https://github.com/openshift/api

**Version-Specific Modules**: No

**Current Version**:
```go
github.com/openshift/api v0.0.0-20260511191110-9b69e5fa27e9
```

**Compatibility Notes**:
- Introduced in OpenShift 4.13 (relatively new)
- Replaces older `ImageContentSourcePolicy` (ICSP)
- Only available in newer OpenShift versions

**Related CRDs**:
- `ImageTagMirrorSet` - for tag-based mirroring
- `ImageContentSourcePolicy` (deprecated, use ImageDigestMirrorSet)

**Minimum OpenShift Version**: 4.13+

**Usage Pattern**:
```go
import configv1 "github.com/openshift/api/config/v1"

idms := &configv1.ImageDigestMirrorSet{
    // Mirror set definition
}
```

**Migration Note**: If supporting older OpenShift (<4.13), you may need to use `ImageContentSourcePolicy` instead

---

### 15. ClusterVersion (config.openshift.io/v1)

**Go Module**: `github.com/openshift/api/config/v1`

**Source Repository**: https://github.com/openshift/api

**Version-Specific Modules**: No

**Current Version**:
```go
github.com/openshift/api v0.0.0-20260511191110-9b69e5fa27e9
```

**Compatibility Notes**:
- Core OpenShift resource for cluster version management
- Stable API across OpenShift versions
- Used for determining cluster version (as seen in PR #1100)

**Usage Pattern** (from PR #1100):
```go
import configv1 "github.com/openshift/api/config/v1"

clusterVersion := &configv1.ClusterVersion{}
key := client.ObjectKey{Name: "version"}
err := client.Get(ctx, key, clusterVersion)

// Access version info
version := clusterVersion.Status.Desired.Version  // e.g., "4.17.1"
```

**Important Fields**:
- `Status.Desired.Version` - target/desired OpenShift version
- `Status.History` - version update history
- `Status.Conditions` - upgrade status

---

## OpenShift Version-Specific Module Pattern (Console Example)

### Console (console.openshift.io/v1)

This is the **key example** from PR #1100 showing version-specific modules:

**For OpenShift 4.19+**:
```go
github.com/openshift/api v0.0.0-20260511191110-9b69e5fa27e9
import osv1 "github.com/openshift/api/console/v1"
```

**For OpenShift 4.17-4.18**:
```go
github.com/rhobs/openshift-api v0.0.0-20260512142436-2e89e902a420
import osRhobsv1 "github.com/rhobs/openshift-api/console/v1"
```

**For OpenShift < 4.17**:
```go
github.com/rhobs/openshift-api v0.0.0-20260512142436-2e89e902a420
import osv1alpha1 "github.com/rhobs/openshift-api/console/v1alpha1"
```

**Why Three Different Imports?**

1. **The Problem**: OpenShift 4.19 added a new field `ConsolePlugin.Spec.ContentSecurityPolicy` to the console/v1 API
   
2. **The Marshalling Issue**: The upstream `openshift/api` marshals this field by default, even when nil/unset
   
3. **The Failure**: When this CR is sent to OpenShift 4.17/4.18 clusters, the API server rejects it because the field doesn't exist

4. **The Solution**: RHOBS forked `openshift/api` at a commit before the CSP field was added
   - This fork provides console/v1 WITHOUT the CSP field for 4.17-4.18
   - It also retains console/v1alpha1 for <4.17 support

**Implementation Pattern** (from PR #1100):

```go
// Runtime version detection
func IsVersionAheadOrEqual(currentVersion, targetVersion string) bool {
    // Semantic version comparison
    return semver.Compare(current, target) >= 0
}

// Conditional resource creation
if IsVersionAheadOrEqual(clusterVersion, "4.19") {
    // Use upstream console/v1 with CSP support
    plugin := newConsolePlugin(...)  // osv1.ConsolePlugin
} else if IsVersionAheadOrEqual(clusterVersion, "4.17") {
    // Use forked console/v1 without CSP
    plugin := newRhobsConsolePlugin(...)  // osRhobsv1.ConsolePlugin
} else {
    // Use v1alpha1 for older versions
    plugin := newLegacyConsolePlugin(...)  // osv1alpha1.ConsolePlugin
}
```

**Files Modified in PR #1100**:
- `pkg/operator/scheme.go` - Register different schemes based on version
- `pkg/controllers/uiplugin/components.go` - Create different CR versions
- `pkg/controllers/uiplugin/proxy.go` - Conversion between proxy types
- `DEPENDENCY_CONSTRAINTS.md` - Documentation

**Key Takeaway**: When a module adds new required/default fields that break older clusters, you need:
1. Forked module at an older commit
2. Runtime version detection
3. Conditional code paths for each version

---

## OLM (Operator Lifecycle Manager) Resources

### 16. Subscription (operators.coreos.com/v1alpha1)

**Go Module**: `github.com/operator-framework/api/pkg/operators/v1alpha1`

**Source Repository**: https://github.com/operator-framework/api

**Version-Specific Modules**: No - stable v1alpha1

**Current Version** (from observability-operator):
```go
github.com/operator-framework/api v0.42.0
```

**Compatibility Notes**:
- Part of Operator Lifecycle Manager (OLM)
- Stable API despite v1alpha1 designation
- Standard across all Kubernetes distros with OLM

**Other CRDs in this module**:
- `ClusterServiceVersion (CSV)`
- `InstallPlan`
- `CatalogSource`
- `OperatorGroup`

**Usage Pattern**:
```go
import olmv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"

sub := &olmv1alpha1.Subscription{
    Spec: olmv1alpha1.SubscriptionSpec{
        Package: "observability-operator",
        Channel: "stable",
    },
}
```

---

### 17. OperatorGroup (operators.coreos.com/v1)

**Go Module**: `github.com/operator-framework/api/pkg/operators/v1`

**Source Repository**: https://github.com/operator-framework/api

**Version-Specific Modules**: No

**Current Version**:
```go
github.com/operator-framework/api v0.42.0
```

**Compatibility Notes**:
- OperatorGroup is v1 (more stable than Subscription's v1alpha1)
- Defines the deployment scope for operators
- Required for multi-tenant operator installations

**API Structure**:
```go
import olmv1 "github.com/operator-framework/api/pkg/operators/v1"

og := &olmv1.OperatorGroup{
    Spec: olmv1.OperatorGroupSpec{
        TargetNamespaces: []string{"namespace1", "namespace2"},
    },
}
```

**Note**: Different API version (v1) than Subscription (v1alpha1) but same module

---

### 18. CatalogSource (operators.coreos.com/v1alpha1)

**Go Module**: `github.com/operator-framework/api/pkg/operators/v1alpha1`

**Source Repository**: https://github.com/operator-framework/api

**Version-Specific Modules**: No

**Current Version**:
```go
github.com/operator-framework/api v0.42.0
```

**Compatibility Notes**:
- Defines operator catalogs/indexes
- Same module as Subscription
- Stable despite v1alpha1 version

**Usage Pattern**:
```go
import olmv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"

cs := &olmv1alpha1.CatalogSource{
    Spec: olmv1alpha1.CatalogSourceSpec{
        SourceType: olmv1alpha1.SourceTypeGrpc,
        Image: "quay.io/my-org/my-catalog:latest",
    },
}
```

**Types of CatalogSources**:
- `grpc` - OCI image containing operator bundles
- `configmap` - ConfigMap containing package manifests (deprecated)
- `address` - External gRPC server

---

## Summary: Module Selection Decision Tree

### For OpenShift Resources (config.openshift.io, route.openshift.io, etc.)

```
START
  |
  ├─ Is it Console (console.openshift.io)?
  │   ├─ YES → Check target OpenShift versions
  │   │   ├─ Only 4.19+ → Use github.com/openshift/api
  │   │   ├─ Only 4.17-4.18 → Use github.com/rhobs/openshift-api (console/v1)
  │   │   ├─ Only <4.17 → Use github.com/rhobs/openshift-api (console/v1alpha1)
  │   │   └─ Mixed versions → Use runtime detection + all three imports
  │   │
  │   └─ NO → Use github.com/openshift/api (standard)
```

### For Observability Operators

```
START
  |
  ├─ Prometheus-related CRDs?
  │   └─ YES → github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring
  │       └─ Alternative: github.com/rhobs/obo-prometheus-operator (if using RHOBS fork)
  |
  ├─ Grafana Loki?
  │   └─ YES → github.com/grafana/loki/operator/apis/loki/v1
  |
  ├─ Grafana Tempo?
  │   └─ YES → github.com/grafana/tempo-operator/apis/tempo/v1alpha1
  |
  ├─ OpenTelemetry?
  │   └─ YES → github.com/open-telemetry/opentelemetry-operator/apis/v1beta1
  |
  ├─ Network Observability?
  │   └─ YES → github.com/netobserv/network-observability-operator/apis
  |
  ├─ Multi-Cluster (ACM)?
  │   └─ YES → github.com/stolostron/multicluster-observability-operator/apis
  |
  └─ Perses Dashboards?
      └─ YES → Check if using operator
          ├─ With operator → github.com/rhobs/perses-operator
          └─ Core types → github.com/perses/perses/pkg/apis/v1alpha1
```

### For OLM Resources

```
All OLM CRDs → github.com/operator-framework/api
  ├─ Subscription → v1alpha1
  ├─ CatalogSource → v1alpha1  
  ├─ OperatorGroup → v1
  └─ InstallPlan → v1alpha1
```

---

## Common Integration Patterns

### Pattern 1: Version Detection (OpenShift)

```go
import (
    "context"
    configv1 "github.com/openshift/api/config/v1"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "golang.org/x/mod/semver"
)

func getOpenShiftVersion(ctx context.Context, c client.Client) (string, error) {
    cv := &configv1.ClusterVersion{}
    key := client.ObjectKey{Name: "version"}
    if err := c.Get(ctx, key, cv); err != nil {
        return "", err
    }
    return cv.Status.Desired.Version, nil
}

func isVersionAheadOrEqual(current, target string) bool {
    if !strings.HasPrefix(current, "v") {
        current = "v" + current
    }
    if !strings.HasPrefix(target, "v") {
        target = "v" + target  
    }
    return semver.Compare(current, target) >= 0
}
```

### Pattern 2: Conditional Scheme Registration

```go
import (
    osv1 "github.com/openshift/api/console/v1"
    osRhobsv1 "github.com/rhobs/openshift-api/console/v1"
    osv1alpha1 "github.com/rhobs/openshift-api/console/v1alpha1"
    "k8s.io/apimachinery/pkg/runtime"
)

func registerConsoleScheme(scheme *runtime.Scheme, version string) error {
    if isVersionAheadOrEqual(version, "4.19") {
        return osv1.AddToScheme(scheme)
    } else if isVersionAheadOrEqual(version, "4.17") {
        return osRhobsv1.AddToScheme(scheme)
    } else {
        return osv1alpha1.AddToScheme(scheme)
    }
}
```

### Pattern 3: Type Conversion/Adaptation

```go
// Define an adapter interface
type ProxyAdapter interface {
    ToV1() osv1.ConsolePluginProxy
    ToRhobsV1() osRhobsv1.ConsolePluginProxy
    ToV1Alpha1() osv1alpha1.ConsolePluginProxy
}

// Internal representation
type PluginProxy struct {
    Alias            string
    ServiceName      string
    ServiceNamespace string
    ServicePort      int32
    Authorize        bool
}

func (p PluginProxy) ToV1() osv1.ConsolePluginProxy {
    // Convert to upstream v1 format
}

func (p PluginProxy) ToRhobsV1() osRhobsv1.ConsolePluginProxy {
    // Convert to forked v1 format
}
```

---

## Troubleshooting Common Issues

### Issue 1: "no matches for kind ConsolePlugin in version console.openshift.io/v1"

**Symptom**: Trying to create console/v1 ConsolePlugin on OpenShift < 4.17

**Solution**: Use console/v1alpha1 for OpenShift <4.17, or console/v1 for >=4.17

### Issue 2: "unknown field spec.contentSecurityPolicy" 

**Symptom**: Creating ConsolePlugin with CSP field on OpenShift 4.17/4.18

**Solution**: Use `github.com/rhobs/openshift-api` instead of `github.com/openshift/api` for these versions

### Issue 3: Module version conflicts

**Symptom**: 
```
go: github.com/openshift/api@v0.0.0-xyz
conflicts with github.com/rhobs/openshift-api@v0.0.0-abc
```

**Solution**: These are meant to be used together - they have different import paths. Check that you're importing from the correct one in each file.

### Issue 4: CRD not found

**Symptom**: `the server could not find the requested resource`

**Solution**:
1. Ensure the operator providing the CRD is installed
2. Check CRD installation: `kubectl get crd | grep <resource>`
3. Verify API version matches your cluster version

---

## Version Compatibility Matrix

| CRD | Module | Kubernetes | OpenShift | Notes |
|-----|---------|------------|-----------|-------|
| **Observability CRDs** |
| ServiceMonitor | prometheus-operator/prometheus-operator | 1.16+ | 4.6+ | Stable |
| PrometheusRule | prometheus-operator/prometheus-operator | 1.16+ | 4.6+ | Stable |
| LokiStack | grafana/loki/operator | 1.21+ | 4.10+ | Requires Loki Operator |
| TempoStack | grafana/tempo-operator | 1.25+ | 4.12+ | v1alpha1 |
| OpenTelemetryCollector | open-telemetry/opentelemetry-operator | 1.25+ | 4.12+ | v1beta1 |
| FlowCollector | netobserv/network-observability-operator | 1.23+ | 4.12+ | v1beta2 |
| MultiClusterObservability | stolostron/multicluster-observability-operator | 1.23+ | 4.11+ | Requires ACM |
| PersesDashboard | perses/perses | 1.23+ | 4.12+ | v1alpha1 (unstable) |
| **OpenShift Resources** |
| Console (v1) | openshift/api OR rhobs/openshift-api | N/A | 4.17+ | See version notes |
| Console (v1alpha1) | rhobs/openshift-api | N/A | <4.17 | Deprecated |
| Route | openshift/api | N/A | 4.6+ | Stable |
| OAuth | openshift/api | N/A | 4.6+ | Stable |
| Project | openshift/api | N/A | 4.6+ | Stable |
| ImageDigestMirrorSet | openshift/api | N/A | 4.13+ | Newer feature |
| ClusterVersion | openshift/api | N/A | 4.6+ | Stable |
| **OLM Resources** |
| Subscription | operator-framework/api | 1.16+ | 4.6+ | v1alpha1 (stable) |
| OperatorGroup | operator-framework/api | 1.16+ | 4.6+ | v1 (stable) |
| CatalogSource | operator-framework/api | 1.16+ | 4.6+ | v1alpha1 (stable) |

---

## References

1. **Original PR #1100**: https://github.com/rhobs/observability-operator/pull/1100
   - Demonstrates version-specific module usage pattern
   - Shows runtime version detection implementation
   
2. **Observability Operator**: https://github.com/rhobs/observability-operator
   - Contains go.mod with all dependencies
   - Reference implementation for integration patterns

3. **OpenShift API**: https://github.com/openshift/api
   - Upstream OpenShift API definitions
   - Latest API versions

4. **RHOBS OpenShift API Fork**: https://github.com/rhobs/openshift-api
   - Forked for backward compatibility
   - Preserves v1alpha1 and pre-CSP console/v1

5. **Operator Framework API**: https://github.com/operator-framework/api
   - OLM CRD definitions

6. **Prometheus Operator**: https://github.com/prometheus-operator/prometheus-operator
   - Monitoring CRDs

7. **Documentation**:
   - [OLM Architecture](https://olm.operatorframework.io/docs/)
   - [Prometheus Operator Docs](https://prometheus-operator.dev/)
   - [OpenShift API Docs](https://docs.openshift.com/)

---

## Recommended Approach for New Projects

1. **Start with Latest Stable Versions**
   - Use latest stable versions of operator modules
   - Check go.mod in reference projects (like observability-operator)

2. **Plan for Multi-Version Support**
   - If targeting multiple cluster versions, plan version detection early
   - Consider using forked modules only when necessary

3. **Test Across Versions**
   - Test on minimum and maximum supported versions
   - Verify CRD compatibility in each target environment

4. **Monitor Upstream Changes**
   - Subscribe to release notifications for critical dependencies
   - Watch for API version deprecations/migrations

5. **Document Version Requirements**
   - Clearly document minimum cluster versions
   - Note any version-specific behavior
   - Create a DEPENDENCY_CONSTRAINTS.md (like observability-operator)

---

## Conclusion

The key lesson from PR #1100 is that **API version numbers alone don't guarantee backward compatibility**. When upstream modules add new fields with default marshalling behavior, you may need:

1. **Forked modules** at specific commits for older versions
2. **Runtime version detection** to choose the appropriate code path
3. **Multiple import paths** for the same logical resource at different versions

This pattern is most critical for OpenShift Console CRDs but may apply to other rapidly-evolving APIs. Always check the actual marshalled output and test against your target cluster versions.
