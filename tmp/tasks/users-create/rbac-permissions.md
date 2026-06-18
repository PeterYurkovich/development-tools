# RBAC Permissions Matrix for Test Users

**Purpose**: Define varied RBAC permissions for 6 test users to enable comprehensive testing of observability signal access controls.

**Design Goals**:
- Cover all signal types: Metrics, Logs, Traces, Dashboards
- Test both namespace-scoped and cluster-scoped permissions
- Test read-only vs read-write access
- Enable verification that RBAC is working correctly

---

## User Permissions Overview

| User | Profile | Namespaces | Access Level |
|------|---------|------------|--------------|
| user1 | Admin-like | Cluster-wide | Read + Write |
| user2 | Read-only | Cluster-wide | Read only |
| user3 | Multi-namespace editor | perses-dev, openshift-cluster-observability-operator | Read + Write |
| user4 | Single-namespace editor | perses-dev | Read + Write |
| user5 | Single-namespace viewer | perses-dev | Read only |
| user6 | Dashboards + Metrics viewer | perses-dev | Read only (specific resources) |

---

## Detailed Permissions

### User1: Admin-like (All Signals, Cluster-wide, Read+Write)

**Use Case**: Verify full access to all observability features

**Cluster-scoped Permissions:**
```yaml
ClusterRoleBindings:
  - user1-cluster-monitoring-view        # View all metrics cluster-wide
  - user1-cluster-admin                  # Full cluster admin (includes monitoring edit)
  - user1-cluster-logging-application-view  # View logs cluster-wide
  - user1-distributed-tracing-view       # View traces cluster-wide
```

**Namespace-scoped Permissions** (in specified namespace):
```yaml
RoleBindings:
  - user1-admin                          # Full admin in namespace
```

**Expected Capabilities:**
- ✅ View/edit PrometheusRules, ServiceMonitors
- ✅ View/edit ClusterLogForwarders, Logging configurations
- ✅ View/edit TempoStack, OpenTelemetry collectors
- ✅ View/edit Perses dashboards
- ✅ Access all console plugins (monitoring, logging, tracing, dashboards)

---

### User2: Read-only (All Signals, Cluster-wide, Read-only)

**Use Case**: Verify read-only access works across all signals

**Cluster-scoped Permissions:**
```yaml
ClusterRoleBindings:
  - user2-cluster-monitoring-view        # View metrics cluster-wide
  - user2-cluster-logging-application-view  # View logs cluster-wide
  - user2-distributed-tracing-view       # View traces cluster-wide
```

**Namespace-scoped Permissions** (in specified namespace):
```yaml
RoleBindings:
  - user2-view                           # View resources in namespace
```

**Expected Capabilities:**
- ✅ View PrometheusRules, ServiceMonitors (cannot edit)
- ✅ View logs from all namespaces (cannot edit forwarders)
- ✅ View traces from all namespaces (cannot edit collectors)
- ✅ View Perses dashboards (cannot edit)
- ❌ Cannot create/edit/delete any observability resources

---

### User3: Multi-Namespace Editor (2 namespaces, Read+Write)

**Use Case**: Verify read/write access across multiple namespaces

**Cluster-scoped Permissions:**
```yaml
ClusterRoleBindings:
  - None
```

**Namespace-scoped Permissions**:
```yaml
RoleBindings in perses-dev:
  - user3-edit-perses-dev                # Edit resources in perses-dev

RoleBindings in openshift-cluster-observability-operator:
  - user3-edit-observability-operator    # Edit resources in openshift-cluster-observability-operator
```

**Expected Capabilities:**
- ✅ Create/edit/delete resources in perses-dev
- ✅ Create/edit/delete resources in openshift-cluster-observability-operator
- ✅ Full access to all resource types in both namespaces
- ❌ Cannot access other namespaces
- ❌ No cluster-scoped permissions

---

### User4: Single-Namespace Editor (1 namespace, Read+Write)

**Use Case**: Verify read/write access in single namespace

**Cluster-scoped Permissions:**
```yaml
ClusterRoleBindings:
  - None
```

**Namespace-scoped Permissions**:
```yaml
RoleBindings in perses-dev:
  - user4-edit-perses-dev                # Edit resources in perses-dev
```

**Expected Capabilities:**
- ✅ Create/edit/delete resources in perses-dev
- ✅ Full access to all resource types in namespace
- ❌ Cannot access other namespaces
- ❌ No cluster-scoped permissions

---

### User5: Single-Namespace Viewer (1 namespace, Read-only)

**Use Case**: Verify read-only access in single namespace

**Cluster-scoped Permissions:**
```yaml
ClusterRoleBindings:
  - None
```

**Namespace-scoped Permissions**:
```yaml
RoleBindings in perses-dev:
  - user5-view-perses-dev                # View resources in perses-dev
```

**Expected Capabilities:**
- ✅ View all resources in perses-dev (read-only)
- ❌ Cannot create/edit/delete resources
- ❌ Cannot access other namespaces
- ❌ No cluster-scoped permissions

---

### User6: Dashboards + Metrics Viewer (1 namespace, Specific resources)

**Use Case**: Verify scoped access to dashboards and metrics only

**Cluster-scoped Permissions:**
```yaml
ClusterRoleBindings:
  - None
```

**Namespace-scoped Permissions**:
```yaml
RoleBindings in perses-dev:
  - user6-perses-dashboards-view         # View PersesDashboards
  - user6-metrics-view                   # View metrics-related resources (ServiceMonitors, PrometheusRules)
```

**Expected Capabilities:**
- ✅ View PersesDashboards in perses-dev
- ✅ View ServiceMonitors, PrometheusRules in perses-dev
- ❌ Cannot view other resource types (Pods, Deployments, etc.)
- ❌ Cannot edit any resources
- ❌ Cannot access other namespaces

---

## ClusterRole Reference

**OpenShift Built-in ClusterRoles** (should exist on all clusters):
- `admin` - Full admin access to namespace resources
- `edit` - Edit most resources in namespace
- `view` - Read-only access to namespace resources
- `cluster-admin` - Full cluster admin access
- `cluster-monitoring-view` - View cluster monitoring metrics
- `monitoring-rules-edit` - Edit PrometheusRules
- `monitoring-edit` - Edit ServiceMonitors, PodMonitors

**Observability Operator ClusterRoles** (created by operators):
- `cluster-logging-application-view` - View application logs (CLO)
- `cluster-monitoring-view` - View cluster monitoring metrics
- `distributed-tracing-view` - View traces (Tempo Operator)

**Custom Roles** (may need to be created for user6):
- Custom Role for viewing PersesDashboards (if not provided by operator)
- Custom Role for viewing metrics resources (ServiceMonitors, PrometheusRules)

**Note**: If a ClusterRole doesn't exist, the RoleBinding/ClusterRoleBinding will still be created but will have no effect until the role is created by its respective operator.

---

## Implementation Notes

### Namespace Flag Behavior

The `--namespace` flag is used **only for user1 and user2** (the cluster-scoped users):
- user1: Gets `admin` role in specified namespace
- user2: Gets `view` role in specified namespace

**Users 3-6 have hard-coded namespace assignments:**
- user3: `perses-dev` + `openshift-cluster-observability-operator`
- user4: `perses-dev`
- user5: `perses-dev`
- user6: `perses-dev`

**Default namespace**: `openshift-monitoring` (for user1 and user2)

**Note**: The command will create all required namespaces if they don't exist:
- `perses-dev` (always created)
- `openshift-cluster-observability-operator` (always created)
- Namespace from `--namespace` flag (created for user1/user2)

### Testing Workflow

**Setup:**
```bash
obstool users create --namespace=openshift-monitoring
# Creates:
#   - perses-dev namespace
#   - openshift-cluster-observability-operator namespace
#   - openshift-monitoring namespace (if not exists)
```

**Test User1 (Admin):**
```bash
oc login --username=user1 --password=password
oc create -f prometheus-rule.yaml -n openshift-monitoring  # Should succeed
oc delete prometheusrule test -n openshift-monitoring       # Should succeed
oc get pods -A                                               # Should succeed (cluster-admin)
```

**Test User2 (Read-only):**
```bash
oc login --username=user2 --password=password
oc get prometheusrules -A                                   # Should succeed
oc get pods -n openshift-monitoring                         # Should succeed
oc create -f prometheus-rule.yaml -n openshift-monitoring  # Should fail (forbidden)
```

**Test User3 (Logging specialist):**
```bash
oc login --username=user3 --password=password
oc get clusterlogforwarder -n openshift-monitoring         # Should succeed
oc create -f clusterlogforwarder.yaml                      # Should succeed
oc get prometheusrules                                     # Should fail (no permission)
```

**Test User4 (Monitoring specialist):**
```bash
oc login --username=user4 --password=password
oc create -f prometheus-rule.yaml -n openshift-monitoring  # Should succeed
oc get clusterlogforwarder -A                              # Should fail (no permission)
```

**Test User5 (Tracing specialist):**
```bash
oc login --username=user5 --password=password
oc create -f tempostack.yaml -n openshift-monitoring       # Should succeed
oc get prometheusrules                                     # Should fail (no permission)
```

**Test User6 (Metrics + Dashboards read-only):**
```bash
oc login --username=user6 --password=password
oc get prometheusrules -A                                  # Should succeed (view)
oc get persesdashboards -n openshift-monitoring            # Should succeed (view)
oc create -f prometheus-rule.yaml                          # Should fail (read-only)
```

---

## RBAC Resources Summary

**Total RoleBindings per namespace**: ~12-15 bindings  
**Total ClusterRoleBindings**: ~18-20 bindings

### Per-User Breakdown

| User | Namespace RoleBindings | ClusterRoleBindings | Total |
|------|----------------------|---------------------|-------|
| user1 | 1 (admin in default namespace) | 4 | 5 |
| user2 | 1 (view in default namespace) | 3 | 4 |
| user3 | 2 (edit in perses-dev + edit in openshift-cluster-observability-operator) | 0 | 2 |
| user4 | 1 (edit in perses-dev) | 0 | 1 |
| user5 | 1 (view in perses-dev) | 0 | 1 |
| user6 | 2 (dashboards-view + metrics-view in perses-dev) | 0 | 2 |
| **Total** | **8** | **7** | **15** |

---

## ClusterRole Existence Handling

**Question**: Can RoleBinding reference non-existent ClusterRole?

**Answer**: YES - RoleBindings can reference ClusterRoles that don't exist yet. The binding is created successfully but has no effect until the ClusterRole is created.

**Implication for obstool**:
- ✅ Create all RoleBindings/ClusterRoleBindings regardless of role existence
- ✅ Log informational message if role doesn't exist (not an error)
- ✅ Permissions will automatically activate when operator installs the role

**Example**:
```bash
# Create binding before role exists
oc create clusterrolebinding test --clusterrole=does-not-exist --user=user1
# Success - binding created

# Later, when role is created
oc create clusterrole does-not-exist --verb=get --resource=pods
# Binding now grants permission automatically
```

---

## Future Enhancements

**Additional user scenarios** (for later implementation):
- User7: Network observability (NetObserv/FlowCollector)
- User8: ACM observability (MultiClusterObservability)
- User9: Custom dashboard creator (Perses edit + view metrics)
- User10: Incident responder (read all signals, limited write)

**RBAC scenarios command** (`obstool users rbac --scenario=X`):
- `perses-e2e` - Perses E2E testing scenario
- `full-read` - All users get read-only to all signals
- `signal-specialists` - Current 6-user setup
- `custom` - Load from YAML file

---

**Last Updated**: 2026-06-18  
**Status**: Proposed - Awaiting approval  
**Implementation**: Part of `obstool users create` command
