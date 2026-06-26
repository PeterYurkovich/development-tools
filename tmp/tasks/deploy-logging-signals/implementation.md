# Deploy Logging Signal Generator — Implementation Summary

## Status: Complete

`go build ./...` passes with zero errors.

---

## What Was Implemented

### New File

| File | Description |
|---|---|
| `pkg/resources/logging_signals.go` | `CreateChatApp` — creates `chat` namespace and a `chat-x` Deployment running the log generator with full security context |

### Modified Files

| File | Change |
|---|---|
| `internal/constants/logging.go` | Added `ChatNamespace = "chat"`, `ChatDeploymentName = "chat-x"`, `ChatImage = "quay.io/libpod/alpine"` |
| `pkg/operations/logging.go` | Added `DeploySignals bool` to `DeployLoggingConfig`; added `StepDeploySignalApps` constant; added non-fatal signal step at end of `DeployLogging` |
| `cmd/deploy/logging.go` | Added `--deploy-signals` flag; refactored TUI path to use `collectLoggingInput` returning full config struct and `buildLoggingOperationsList` for dynamic list; added `huh.NewConfirm()` for signals |

---

## Design Notes

**Deployment (not bare Pod)**: Upgraded from the bare `Pod` in `z_logging.sh` to an `appsv1.Deployment` for self-healing consistency with the tracing signal generators.

**Non-fatal**: `CreateChatApp` errors are logged as warnings and do not abort the logging deployment. The step always reaches `StatusComplete`.

**TUI refactor**: `runDeployLoggingTUI` was refactored to use `collectLoggingInput` (returning a full `DeployLoggingConfig`) and `buildLoggingOperationsList` (dynamic list based on config). This eliminates the previous hardcoded 8-item list that would drift out of sync when new optional steps were added.

**`ensureNamespaceExists` reuse**: `CreateChatApp` calls `ensureNamespaceExists` which is defined in `tracing_signals.go` within the same `resources` package — no duplication.

---

## Build Status

```
go build ./...   → OK (zero errors)
```
