# Corrected: Update/Cleanup Monitoring Commands

**Date**: 2026-06-12  
**Status**: ✅ Logic Corrected

---

## Correction Summary

The initial implementation had the update/cleanup logic reversed. This has been corrected.

### Incorrect Implementation (Before)

```bash
# WRONG: update scaled UP
obstool update monitoring
# Scaled up CMO to 1 replica
# Scaled up monitoring-plugin to 1 replica

# WRONG: cleanup scaled DOWN
obstool cleanup monitoring  
# Scaled down CMO to 0 replicas
# Scaled down monitoring-plugin to 0 replicas
```

### Correct Implementation (After)

```bash
# CORRECT: update scales DOWN CMO to allow manual plugin updates
obstool update monitoring --image=quay.io/my-org/monitoring-plugin:v1.2.3
# Scales down CMO to 0 replicas
# Updates monitoring-plugin deployment image

# CORRECT: cleanup scales UP CMO to restore managed state
obstool cleanup monitoring
# Scales up CMO to 1 replica
# CMO reconciles and restores monitoring-plugin to proper state
```

---

## Rationale

### Update Command

**Purpose**: Allow manual updates to the monitoring plugin

**Flow**:
1. Scale down CMO (to 0 replicas)
2. CMO can no longer reconcile/overwrite changes
3. Manually update monitoring-plugin image (optional with `--image` flag)
4. Plugin runs with custom image

### Cleanup Command

**Purpose**: Restore monitoring to normal, managed state

**Flow**:
1. Scale up CMO (to 1 replica)
2. CMO starts reconciling
3. CMO restores monitoring-plugin to its managed configuration
4. System returns to normal state

---

## Updated Command Behavior

### `obstool update monitoring`

**Flags**:
- `--image` (optional) - Monitoring plugin image to use

**Actions**:
1. Scales down `cluster-monitoring-operator` to 0
2. If `--image` provided: Updates `monitoring-plugin` deployment image

**Output**:
```
Scaled down cluster-monitoring-operator
Updated monitoring-plugin image to quay.io/my-org/monitoring-plugin:v1.2.3
```

**Use Case**: Testing custom monitoring plugin builds

### `obstool cleanup monitoring`

**Flags**: None

**Actions**:
1. Scales up `cluster-monitoring-operator` to 1

**Output**:
```
Scaled up cluster-monitoring-operator (will reconcile monitoring plugin)
```

**Use Case**: Return to normal state after testing

---

## Code Changes

### cmd/update/monitoring.go

**Added**:
- `--image` flag for plugin image
- `scaleDownDeployment()` - scales deployment to 0
- `updatePluginImage()` - updates plugin container image

**Flow**:
```go
1. Scale down CMO
2. If --image provided: Update plugin image
```

### cmd/cleanup/monitoring.go

**Added**:
- `scaleUpDeployment()` - scales deployment to 1

**Flow**:
```go
1. Scale up CMO (CMO will reconcile everything)
```

---

## Testing

### Test Scenario 1: Update Plugin Image

```bash
# 1. Update to custom image
$ obstool update monitoring --image=quay.io/myorg/monitoring-plugin:custom
Scaled down cluster-monitoring-operator
Updated monitoring-plugin image to quay.io/myorg/monitoring-plugin:custom

# 2. Verify plugin is running with custom image
$ oc get deployment monitoring-plugin -n openshift-monitoring -o jsonpath='{.spec.template.spec.containers[0].image}'
quay.io/myorg/monitoring-plugin:custom

# 3. Cleanup (restore to normal)
$ obstool cleanup monitoring
Scaled up cluster-monitoring-operator (will reconcile monitoring plugin)

# 4. Verify CMO restored the managed image
$ oc get deployment monitoring-plugin -n openshift-monitoring -o jsonpath='{.spec.template.spec.containers[0].image}'
# Shows the managed image from CMO
```

### Test Scenario 2: Just Scale Down CMO

```bash
# Scale down CMO without changing image
$ obstool update monitoring
Scaled down cluster-monitoring-operator

# CMO is now stopped, plugin continues with current image
# Manual changes to plugin won't be reverted

# Restore
$ obstool cleanup monitoring
Scaled up cluster-monitoring-operator (will reconcile monitoring plugin)
```

---

## Documentation Updated

✅ **TODO.md**: Updated task descriptions  
✅ **CONTEXT.md**: Updated command examples  
✅ **Code**: Corrected logic and help text  

---

**Status**: ✅ Corrected and Working
