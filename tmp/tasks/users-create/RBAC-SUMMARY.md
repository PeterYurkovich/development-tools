# RBAC Summary for obstool users create

**Quick Reference** - Updated 2026-06-18

## User Permissions Matrix

| User | Namespaces | Role | Description |
|------|-----------|------|-------------|
| user1 | Cluster-wide + `--namespace` | cluster-admin + admin | Full cluster admin |
| user2 | Cluster-wide + `--namespace` | Cluster view roles + view | Read-only everywhere |
| user3 | perses-dev, openshift-cluster-observability-operator | edit | Read/write in 2 namespaces |
| user4 | perses-dev | edit | Read/write in 1 namespace |
| user5 | perses-dev | view | Read-only in 1 namespace |
| user6 | perses-dev | Custom (dashboards+metrics) | View dashboards & metrics only |

## Namespaces Created

Command: `obstool users create --namespace=X`

**Always created:**
- `perses-dev`
- `openshift-cluster-observability-operator`

**Created from flag:**
- Value of `--namespace` (default: `openshift-monitoring`)

## Total RBAC Resources

**15 bindings total:**
- 8 RoleBindings (namespace-scoped)
- 7 ClusterRoleBindings (cluster-scoped)

## Testing Quick Reference

```bash
# Create users
obstool users create

# Test user3 - multi-namespace access
oc login --username=user3 --password=password
oc get pods -n perses-dev                                      # ✅ Success
oc get pods -n openshift-cluster-observability-operator        # ✅ Success
oc get pods -n default                                         # ❌ Forbidden

# Test user4 - single namespace edit
oc login --username=user4 --password=password
oc create deployment test --image=nginx -n perses-dev          # ✅ Success
oc get pods -n openshift-cluster-observability-operator        # ❌ Forbidden

# Test user5 - single namespace view
oc login --username=user5 --password=password
oc get pods -n perses-dev                                      # ✅ Success
oc create deployment test --image=nginx -n perses-dev          # ❌ Forbidden

# Test user6 - specific resources only
oc login --username=user6 --password=password
oc get persesdashboards -n perses-dev                          # ✅ Success
oc get servicemonitors -n perses-dev                           # ✅ Success
oc get pods -n perses-dev                                      # ❌ Forbidden
```

## Implementation Notes

**user6 requires custom Role creation:**
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: perses-dashboards-metrics-viewer
  namespace: perses-dev
rules:
- apiGroups: ["perses.dev"]
  resources: ["persesdashboards"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["monitoring.coreos.com"]
  resources: ["servicemonitors", "prometheusrules"]
  verbs: ["get", "list", "watch"]
```

Then bind with:
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: user6-dashboards-metrics
  namespace: perses-dev
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: perses-dashboards-metrics-viewer
subjects:
- kind: User
  name: user6
  apiGroup: rbac.authorization.k8s.io
```
