# Implementation: Update/Cleanup Monitoring Commands

**Task ID**: Commands → Update/Cleanup → Monitoring  
**Status**: ✅ Complete  
**Date**: 2026-06-12  
**Implementation Time**: ~30 minutes

---

## Summary

Successfully implemented `obstool update monitoring` and `obstool cleanup monitoring` commands. These commands scale monitoring components up (to 1 replica) or down (to 0 replicas). **Flags-only mode** - no TUI.

---

## What Was Implemented

### Files Created

1. ✅ **`cmd/update/update.go`** (11 lines)
   - Update command group
   - Base command for all update operations

2. ✅ **`cmd/update/monitoring.go`** (67 lines)
   - Scale up monitoring components to 1 replica
   - Updates CMO and monitoring-plugin deployments

3. ✅ **`cmd/cleanup/cleanup.go`** (11 lines)
   - Cleanup command group
   - Base command for all cleanup operations

4. ✅ **`cmd/cleanup/monitoring.go`** (67 lines)
   - Scale down monitoring components to 0 replicas
   - Updates CMO and monitoring-plugin deployments

### Files Modified

5. ✅ **`cmd/root.go`** (updated)
   - Removed monitoring import
   - Added update and cleanup imports
   - Registered UpdateCmd and CleanupCmd

### Files Removed

6. ✅ **`cmd/monitoring/`** (deleted)
   - Old monitoring command group removed
   - Functionality moved to update/cleanup

---

## Commands Available

```bash
# Scale up monitoring (CMO + plugin to 1 replica)
obstool update monitoring

# Scale down monitoring (CMO + plugin to 0 replicas)
obstool cleanup monitoring

# Help
obstool update --help
obstool cleanup --help
```

---

## Implementation Details

### Resources Managed

**1. Cluster Monitoring Operator (CMO)**:
- Namespace: `openshift-monitoring`
- Deployment: `cluster-monitoring-operator`

**2. Monitoring Plugin**:
- Namespace: `openshift-monitoring`
- Deployment: `monitoring-plugin`

### Update Monitoring Logic

1. Get current deployment
2. Set replicas to 1
3. Update deployment

### Cleanup Monitoring Logic

1. Get current deployment
2. Set replicas to 0
3. Update deployment

### Simplicity

No state storage needed - always set to 1 (update) or 0 (cleanup).

---

## Code Structure

### Constants

```go
const (
    monitoringNamespace = "openshift-monitoring"
    cmoDeployment       = "cluster-monitoring-operator"
    pluginDeployment    = "monitoring-plugin"
)
```

### Command Hierarchy

```
UpdateCmd (update)
└── monitoringCmd (monitoring) - scales up to 1

CleanupCmd (cleanup)
└── monitoringCmd (monitoring) - scales down to 0
```

### Core Functions

**Update**:
- `runUpdateMonitoring(cmd, args)` - Orchestrates scaling up both deployments
- `scaleUpDeployment(ctx, client, name)` - Sets deployment replicas to 1

**Cleanup**:
- `runCleanupMonitoring(cmd, args)` - Orchestrates scaling down both deployments
- `scaleDownDeployment(ctx, client, name)` - Sets deployment replicas to 0

---

## Testing Results

### ✅ Root Command Help

```bash
$ ./obstool --help
A unified CLI tool for deploying and managing observability components on OpenShift clusters

Usage:
  obstool [command]

Available Commands:
  cleanup     Cleanup and scale down observability components
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  update      Update and scale up observability components
  version     Print the version of obstool
```

### ✅ Update Command

```bash
$ ./obstool update --help
Commands for updating and scaling up observability components

Usage:
  obstool update [command]

Available Commands:
  monitoring  Scale up monitoring components

$ ./obstool update monitoring
Scaled up cluster-monitoring-operator
Scaled up monitoring-plugin
```

### ✅ Cleanup Command

```bash
$ ./obstool cleanup --help
Commands for cleaning up and scaling down observability components

Usage:
  obstool cleanup [command]

Available Commands:
  monitoring  Scale down monitoring components

$ ./obstool cleanup monitoring
Scaled down cluster-monitoring-operator
Scaled down monitoring-plugin
```

---

## Architectural Changes

### Renamed: Upgrade → Update

**Rationale**: "Update" better describes the action of scaling up and updating components.

### Removed: Monitoring Command Group

**Before**:
```
obstool monitoring scale up
obstool monitoring scale down
```

**After**:
```
obstool update monitoring
obstool cleanup monitoring
```

**Rationale**: Monitoring operations are either updates (scale up) or cleanup (scale down), which fits better into the overarching command structure.

---

## Code Quality

### ✅ Follows Code Style Guidelines

**Comments**:
- Minimal - code is self-documenting

**Variable Naming**:
- `kubeClient` not `kc`
- `deployment` not `d`

**Simplicity**:
- Direct implementation
- No unnecessary abstractions
- Clear error messages

---

## Dependencies

**No new dependencies**:
- ✅ `k8s.io/api/apps/v1` - Already in go.mod
- ✅ `sigs.k8s.io/controller-runtime/pkg/client` - Already in go.mod

---

## Success Criteria Met

✅ **Commands Created**: update monitoring and cleanup monitoring work  
✅ **Simple Logic**: Scale up to 1, scale down to 0  
✅ **Error Handling**: Clear errors for missing deployments  
✅ **Code Quality**: Minimal comments, descriptive names  
✅ **Compilation**: Builds without errors  
✅ **Help Text**: All commands have proper help  
✅ **Architecture**: Fits into update/cleanup structure  

---

## Files Summary

### New Files
1. `cmd/update/update.go` (11 lines)
2. `cmd/update/monitoring.go` (67 lines)
3. `cmd/cleanup/cleanup.go` (11 lines)
4. `cmd/cleanup/monitoring.go` (67 lines)

### Modified Files
1. `cmd/root.go` (updated imports and registration)
2. `tmp/TODO.md` (updated sections)
3. `tmp/go-migration-plan.md` (upgrade → update, monitoring structure)
4. `tmp/CONTEXT.md` (updated command structure)

### Removed Files
1. `cmd/monitoring/monitoring.go` (deleted)
2. `cmd/monitoring/scale.go` (deleted)

**Total New Code**: 156 lines

---

## Documentation Updated

### ✅ Updated Files

1. **go-migration-plan.md**:
   - Changed `upgrade/` → `update/`
   - Updated monitoring command structure
   - Updated cleanup command structure

2. **CONTEXT.md**:
   - Changed command examples
   - Updated directory structure

3. **TODO.md**:
   - Renamed "Monitoring" section → "Update" and "Cleanup"
   - Changed "Upgrade COO" → "Update COO"
   - Marked update/cleanup monitoring as complete

---

## Next Steps

### Ready to Add More Commands

**Update commands** (future):
- `obstool update coo --to-version=X` - Update COO version
- `obstool update monitoring-plugin --image=X` - Update plugin image

**Cleanup commands** (future):
- `obstool cleanup coo` - Remove COO
- `obstool cleanup logging` - Remove logging stack
- `obstool cleanup all` - Remove all components

### Pattern Established

This establishes the pattern for:
1. Update command group (scale up, update versions)
2. Cleanup command group (scale down, remove resources)
3. Clear separation of concerns
4. Consistent command structure

---

## Design Decisions

1. **Update vs Upgrade**: "Update" is more accurate for scaling and version changes
2. **Cleanup vs Delete**: "Cleanup" implies safe, reversible operations
3. **No Monitoring Group**: Operations fit better in update/cleanup
4. **Hardcoded Replicas**: Always 1 for update, 0 for cleanup
5. **Both Deployments**: Always operate on CMO and plugin together
6. **No Waiting**: Just set replicas and return
7. **Flags Only**: No TUI mode for these simple commands

---

**Implementation Status**: ✅ Complete  
**Quality**: ✅ Meets all success criteria  
**Architecture**: ✅ Clean update/cleanup structure  
**Ready For**: Additional update and cleanup commands
