# Deploy Troubleshooting Panel — Task Plan

**Status**: In Progress
**Date**: 2026-06-26

---

## Overview

Implement `obstool deploy troubleshooting-panel` command. Deploys the Network Observability
operator, MinIO (in dedicated `minio` namespace), LokiStack `loki` for netobserv, FlowCollector
`cluster`, and the TroubleshootingPanel UIPlugin.

---

## Assumptions

- `deploy logging` already ran — loki-operator and cluster-logging operators pre-installed
- No signals flag — omitted (chat app covered by `deploy logging`)
- Bare UIPlugin — no TroubleshootingPanelConfig block
- MinIO in dedicated `minio` namespace (matching korrel8r-manifests approach)

---

## Final Design

### Flags

```
obstool deploy troubleshooting-panel [flags]

  --netobserv-channel string   Network Observability operator channel (default: "stable")
  --storage-class string       Storage class for MinIO PVC (auto-detect if empty)
  --deploy-flowcollector       Deploy MinIO, LokiStack 'loki', and FlowCollector cluster
  --deploy-uiplugin            Deploy TroubleshootingPanel UIPlugin
```

`--deploy-flowcollector` triggers all prerequisites: MinIO + LokiStack + FlowCollector.

### Executor Steps (11 max)

| Step | Always? | Description |
|------|---------|-------------|
| StepTSPEnsureNetObservNS | yes | Create namespace openshift-netobserv-operator |
| StepTSPEnsureOperatorGroup | yes | Create OperatorGroup |
| StepTSPCreateSubscription | yes | Create Subscription netobserv-operator |
| StepTSPWaitForCSV | yes | Wait for netobserv-operator CSV |
| StepTSPDeployMinIO | DeployFlowCollector | Deploy MinIO in minio namespace |
| StepTSPEnsureNetObservWorkloadNS | DeployFlowCollector | Create namespace netobserv |
| StepTSPCreateLokiStack | DeployFlowCollector | Create LokiStack loki + secret |
| StepTSPWaitForLokiStack | DeployFlowCollector | Wait for loki-gateway in netobserv |
| StepTSPCreateFlowCollector | DeployFlowCollector | Create FlowCollector cluster |
| StepTSPWaitForNetObservPlugin | DeployFlowCollector | Wait for netobserv-plugin |
| StepTSPDeployUIPlugin | DeployUIPlugin | Deploy TroubleshootingPanel UIPlugin |

---

## Files

| File | Action |
|------|--------|
| `tmp/tasks/deploy-troubleshooting-panel/plan.md` | This file |
| `internal/constants/netobserv.go` | Create |
| `pkg/resources/lokistack.go` | Update — add TenantMode to LokiStackConfig |
| `pkg/resources/netobserv.go` | Create — secret, FlowCollector, wait |
| `pkg/resources/uiplugin.go` | Update — add CreateTroubleshootingPanelUIPlugin() |
| `pkg/operators/netobserv/operatorhub.go` | Create |
| `pkg/k8s/scheme.go` | Update — register FlowCollector v1beta2 scheme |
| `pkg/operations/troubleshooting_panel.go` | Create |
| `cmd/deploy/troubleshooting_panel.go` | Create |
| `go.mod` | Update — add netobserv FlowCollector v1beta2 |

---

## Blocked By

- Deploy command group ✅
- OLM utilities ✅
- Storage provider interface ✅
- Deploy logging ✅ (assumed pre-run)
