# Plan: Implement Update/Cleanup Monitoring Commands

**Task ID**: Commands → Update/Cleanup → Monitoring  
**Status**: ✅ Complete (Refactored from monitoring scale)  
**Date**: 2026-06-12  
**Dependencies**: ✅ Root command (complete)

---

## Overview

Implement `obstool update monitoring` and `obstool cleanup monitoring` commands. These scale the Cluster Monitoring Operator (CMO) and monitoring plugin deployments up (update) or down (cleanup). **Flags-only mode** - no TUI.

---

## Goals

1. **Create monitoring command group** (`cmd/monitoring/monitoring.go`)
2. **Implement scale subcommand** (`cmd/monitoring/scale.go`)
3. **Scale down logic**: Set CMO and monitoring-plugin replicas to 0
4. **Scale up logic**: Restore original replica counts
5. **Store replica counts** when scaling down for restoration

---

## Non-Goals

- ❌ No TUI mode (flags only)
- ❌ No validation of monitoring plugin existence (fail if not found)
- ❌ No waiting/polling for scale completion (just set replicas and return)

---

## Command Structure

```bash
obstool update monitoring
obstool cleanup monitoring
```

**No flags needed** - operations are straightforward.

---

## Implementation Details

### Resources to Scale

**1. Cluster Monitoring Operator (CMO)**:
- Namespace: `openshift-monitoring`
- Deployment: `cluster-monitoring-operator`

**2. Monitoring Plugin**:
- Namespace: `openshift-monitoring`
- Deployment: `monitoring-plugin`

### Scale Down Logic

1. Get current replica count for CMO
2. Get current replica count for monitoring-plugin
3. Store counts (in ConfigMap or annotations)
4. Set both deployments to 0 replicas

### Scale Up Logic

1. Retrieve stored replica counts
2. Restore CMO to original replicas
3. Restore monitoring-plugin to original replicas

### Storage Strategy

**Use deployment annotations** to store original replica count:
- Annotation key: `obstool.observability.openshift.io/original-replicas`
- Value: Original replica count as string

This is simpler than a ConfigMap and keeps data with the resource.

---

## File Structure

```
cmd/monitoring/
├── monitoring.go    # Monitoring command group
└── scale.go         # Scale up/down subcommands
```

---

## Implementation

### File 1: `cmd/monitoring/monitoring.go`

```go
package monitoring

import (
    "github.com/spf13/cobra"
)

var MonitoringCmd = &cobra.Command{
    Use:   "monitoring",
    Short: "Manage monitoring components",
    Long:  "Commands for managing OpenShift monitoring components",
}
```

### File 2: `cmd/monitoring/scale.go`

```go
package monitoring

import (
    "context"
    "fmt"
    "strconv"

    "github.com/spf13/cobra"
    appsv1 "k8s.io/api/apps/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "sigs.k8s.io/controller-runtime/pkg/client"

    execctx "github.com/observability-ui/development-tools/pkg/context"
)

const (
    monitoringNamespace = "openshift-monitoring"
    cmoDeployment      = "cluster-monitoring-operator"
    pluginDeployment   = "monitoring-plugin"
    replicaAnnotation  = "obstool.observability.openshift.io/original-replicas"
)

var scaleCmd = &cobra.Command{
    Use:   "scale",
    Short: "Scale monitoring components",
}

var scaleUpCmd = &cobra.Command{
    Use:   "up",
    Short: "Scale monitoring components up",
    RunE:  runScaleUp,
}

var scaleDownCmd = &cobra.Command{
    Use:   "down",
    Short: "Scale monitoring components down",
    RunE:  runScaleDown,
}

func init() {
    scaleCmd.AddCommand(scaleUpCmd)
    scaleCmd.AddCommand(scaleDownCmd)
    MonitoringCmd.AddCommand(scaleCmd)
}

func runScaleDown(cmd *cobra.Command, args []string) error {
    ctx := cmd.Context()
    kubeClient, err := execctx.GetClient(ctx)
    if err != nil {
        return err
    }

    // Scale down CMO
    if err := scaleDownDeployment(ctx, kubeClient, cmoDeployment); err != nil {
        return fmt.Errorf("failed to scale down CMO: %w", err)
    }
    fmt.Printf("Scaled down %s\n", cmoDeployment)

    // Scale down monitoring plugin
    if err := scaleDownDeployment(ctx, kubeClient, pluginDeployment); err != nil {
        return fmt.Errorf("failed to scale down monitoring plugin: %w", err)
    }
    fmt.Printf("Scaled down %s\n", pluginDeployment)

    return nil
}

func runScaleUp(cmd *cobra.Command, args []string) error {
    ctx := cmd.Context()
    kubeClient, err := execctx.GetClient(ctx)
    if err != nil {
        return err
    }

    // Scale up CMO
    if err := scaleUpDeployment(ctx, kubeClient, cmoDeployment); err != nil {
        return fmt.Errorf("failed to scale up CMO: %w", err)
    }
    fmt.Printf("Scaled up %s\n", cmoDeployment)

    // Scale up monitoring plugin
    if err := scaleUpDeployment(ctx, kubeClient, pluginDeployment); err != nil {
        return fmt.Errorf("failed to scale up monitoring plugin: %w", err)
    }
    fmt.Printf("Scaled up %s\n", pluginDeployment)

    return nil
}

func scaleDownDeployment(ctx context.Context, kubeClient client.Client, name string) error {
    deployment := &appsv1.Deployment{}
    key := client.ObjectKey{Namespace: monitoringNamespace, Name: name}
    
    if err := kubeClient.Get(ctx, key, deployment); err != nil {
        return err
    }

    // Store original replica count
    if deployment.Spec.Replicas != nil {
        originalReplicas := strconv.Itoa(int(*deployment.Spec.Replicas))
        if deployment.Annotations == nil {
            deployment.Annotations = make(map[string]string)
        }
        deployment.Annotations[replicaAnnotation] = originalReplicas
    }

    // Set replicas to 0
    replicas := int32(0)
    deployment.Spec.Replicas = &replicas

    return kubeClient.Update(ctx, deployment)
}

func scaleUpDeployment(ctx context.Context, kubeClient client.Client, name string) error {
    deployment := &appsv1.Deployment{}
    key := client.ObjectKey{Namespace: monitoringNamespace, Name: name}
    
    if err := kubeClient.Get(ctx, key, deployment); err != nil {
        return err
    }

    // Retrieve stored replica count
    originalReplicasStr, ok := deployment.Annotations[replicaAnnotation]
    if !ok {
        return fmt.Errorf("no stored replica count found for %s", name)
    }

    originalReplicas, err := strconv.Atoi(originalReplicasStr)
    if err != nil {
        return fmt.Errorf("invalid stored replica count: %w", err)
    }

    // Restore original replicas
    replicas := int32(originalReplicas)
    deployment.Spec.Replicas = &replicas

    // Remove annotation
    delete(deployment.Annotations, replicaAnnotation)

    return kubeClient.Update(ctx, deployment)
}
```

### File 3: Update `cmd/root.go`

Register the monitoring command:
```go
import (
    "github.com/observability-ui/development-tools/cmd/monitoring"
)

func init() {
    // ... existing code ...
    rootCmd.AddCommand(monitoring.MonitoringCmd)
}
```

---

## Dependencies

**Already in go.mod**:
- ✅ `k8s.io/api` (apps/v1 for Deployment)
- ✅ `k8s.io/apimachinery` (metav1)
- ✅ `sigs.k8s.io/controller-runtime/pkg/client`

**No new dependencies needed**.

---

## Testing

**Manual testing**:
```bash
# Scale down
./obstool monitoring scale down

# Verify pods are gone
oc get pods -n openshift-monitoring | grep -E "cluster-monitoring-operator|monitoring-plugin"

# Scale up
./obstool monitoring scale up

# Verify pods are back
oc get pods -n openshift-monitoring | grep -E "cluster-monitoring-operator|monitoring-plugin"
```

---

## Success Criteria

✅ **Commands work**:
- `obstool monitoring scale down` - scales to 0
- `obstool monitoring scale up` - restores replicas

✅ **Replica counts preserved**:
- Original counts stored in annotations
- Restored correctly on scale up

✅ **Error handling**:
- Clear errors if deployments not found
- Error if no stored replica count on scale up

✅ **Code quality**:
- Minimal comments
- Descriptive names
- Follows patterns

---

## Timeline

**Total**: ~30 minutes
- File creation: 20 minutes
- Testing: 10 minutes

---

## References

- **TODO.md**: Lines 113-132 (monitoring scale tasks)
- **bash scripts**: `monitoring-plugin/scale.sh` (for reference)

---

**Ready to implement!**
