# Deploy Logging Signal Generator (Chat App) — Implementation Plan

## Overview

Add a `--deploy-signals` flag to `obstool deploy logging` that deploys a chat log generator application. This gives the logging stack an application workload whose stdout logs are immediately collectible via the ClusterLogForwarder, enabling end-to-end testing of the logging pipeline.

## Source of Truth

- `perses/demo/z_logging.sh` — step 2 (chat workload deployed before operators)
- `korrel8r-manifests/openshift/config/resources/chat.yaml` — reference YAML

## Workload Specification

- **Kind**: `Deployment` (upgraded from bare Pod in bash for self-healing)
- **Namespace**: `chat`
- **Deployment name**: `chat-x`
- **Labels**: `app: chat`, `test: "true"`
- **Image**: `quay.io/libpod/alpine`
- **Command**: `sh -c 'i=1; while true; do echo "$(date) chat says hello - $i"; i=$((i + 1)); sleep 1; done'`
- **Security context** (container-level, all four fields from bash source):
  - `allowPrivilegeEscalation: false`
  - `runAsNonRoot: true`
  - `seccompProfile.type: RuntimeDefault`
  - `capabilities.drop: [ALL]`
- **Failure handling**: non-fatal — log warning and continue (matches bash `|| echo "Warning: ..."`)

## Files

### New: `pkg/resources/logging_signals.go`

```go
func CreateChatApp(ctx context.Context, kubeClient client.Client) error
```

- Creates `chat` namespace (idempotent)
- Creates `chat-x` Deployment with the spec above
- Uses `ensureNamespaceExists` from the same `resources` package (defined in `tracing_signals.go`)
- Returns error; caller is responsible for non-fatal handling

### Modified: `internal/constants/logging.go`

```go
ChatNamespace      = "chat"
ChatDeploymentName = "chat-x"
ChatImage          = "quay.io/libpod/alpine"
```

### Modified: `pkg/operations/logging.go`

- Add `DeploySignals bool` to `DeployLoggingConfig`
- Add `StepDeploySignalApps` step constant after `StepWaitForUIPlugin`
- Add step at end of `DeployLogging`: non-fatal call to `resources.CreateChatApp`

### Modified: `cmd/deploy/logging.go`

- Add `--deploy-signals` flag (default: false)
- Read flag in `runDeployLoggingCLI` → pass through config
- Add `huh.NewConfirm()` in `collectLoggingInput` for TUI
- Conditionally append `"Deploy chat signal app"` to `operationsList`

## Implementation Order

```
1. internal/constants/logging.go
2. pkg/resources/logging_signals.go
3. pkg/operations/logging.go
4. cmd/deploy/logging.go
5. go build ./...
```

## Success Criteria

- `go build ./...` passes with zero errors
- `obstool deploy logging --deploy-signals` deploys a `chat-x` Deployment in the `chat` namespace
- TUI form presents the signal option as a confirm question with no pre-selected default
- Failure to deploy the chat app does not abort the logging deployment
