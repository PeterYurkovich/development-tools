# Deploy ACM — Task Plan

**Status**: In Progress
**Date**: 2026-06-26

---

## Overview

Implement `obstool deploy acm` command. Deploys ACM (Advanced Cluster Management) observability stack including the ACM Hub operator, MultiClusterHub CR, MinIO storage (Thanos S3 format), and MultiClusterObservability CR.

---

## Final Design Decisions

### Command Flags

```
obstool deploy acm [flags]

  --acm-channel string   ACM subscription channel (default: "release-2.17")
  --storage-class string Storage class for MinIO PVC (auto-detect if empty)
  --skip-acm-install     Skip ACM Hub operator install (assume pre-installed)
  --skip-multclusterhub  Skip MultiClusterHub CR creation (assume pre-existing)
```

No required flags. Scheduler patching deferred — not included.

### Execution Order (9 steps total, full run)

Based on bash script split: MinIO deploys before MCH wait (doesn't depend on MCH).

| Step | Constant | Description | Conditional |
|------|----------|-------------|-------------|
| 0 | StepACMEnsureHubNamespace | Create namespace open-cluster-management | !SkipACMInstall |
| 1 | StepACMEnsureOperatorGroup | Create OperatorGroup og-global | !SkipACMInstall |
| 2 | StepACMCreateSubscription | Create Subscription advanced-cluster-management | !SkipACMInstall |
| 3 | StepACMWaitForCSV | Wait for ACM CSV Succeeded | !SkipACMInstall |
| 4 | StepACMEnsureObsNamespace | Create namespace open-cluster-management-observability | always |
| 5 | StepACMDeployMinIO | Deploy MinIO (PVC + Deployment + Service + Secret + wait) | always |
| 6 | StepACMCreateMultiClusterHub | Create MultiClusterHub CR | !SkipMultiClusterHub |
| 7 | StepACMWaitForMultiClusterHub | Wait for MCH Complete=True (up to 15m) | !SkipMultiClusterHub |
| 8 | StepACMCreateMCO | Create MultiClusterObservability CR | always |

---

## New Go Dependencies

- `github.com/stolostron/multicluster-observability-operator` — MCO v1beta2 CR
- `github.com/open-cluster-management/multiclusterhub-operator` — MultiClusterHub v1 CR

Both typed (per user decision). Fallback to `unstructured.Unstructured` if version conflicts arise.

---

## Storage Extension (SecretFormatThanos)

`ProviderConfig.SecretFormat` field added. When `SecretFormatThanos`:
- Image: `quay.io/minio/minio:RELEASE.2021-08-25T00-41-18Z`
- Env: `MINIO_ACCESS_KEY` / `MINIO_SECRET_KEY` (old-style)
- Command: `/usr/bin/minio server /storage`
- Secret name: `thanos-object-storage`
- Secret key: single `thanos.yaml` blob with Thanos S3 YAML
- Endpoint in secret: `minio:9000` (short DNS)

---

## Files to Create / Modify

| File | Action |
|------|--------|
| `tmp/tasks/deploy-acm/plan.md` | This file |
| `internal/constants/acm.go` | Create |
| `pkg/storage/provider.go` | Update — add SecretFormat type |
| `pkg/storage/minio.go` | Update — branch on SecretFormatThanos |
| `pkg/operators/acm/operatorhub.go` | Create |
| `pkg/resources/acm.go` | Create — MCH + MCO + WaitForMCH |
| `pkg/k8s/scheme.go` | Update — register MCH + MCO schemes |
| `pkg/operations/acm.go` | Create |
| `cmd/deploy/acm.go` | Create |
| `go.mod` / `go.sum` | Update |

---

## Blocked By

- Deploy command group ✅
- OLM utilities ✅
- Storage provider interface ✅
