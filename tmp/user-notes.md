**Primary Mode**: Flags + TUI (Hybrid)
- Use flags when provided (automation)
- Swap to TUI when flags missing
- Fail early in non-interactive mode


## CRD

ClusterLogForwarder is not a internal to the observability-operator. It is a part of the openshift cluster logging operator.

## Go migration plan
### deploy vs qe
`devtool/cmd/deploy/coo` is the same as `devtools/cmd/qe`. Both look to deploy COO, just using different ways. The coo.go file should handle deploying COO through the following 4 methods:
1. bundle
2. fbc
3. staging
4. Available version from operator hub

### Upgrade
Although other modules do not need it, coo should support upgrading the version in place, similar to how `operator-sdk bundle-upgrade` works

### Cleanup
Should `devtool/cmd/cleanup` have the same structure as `devtool/cmd/deploy` with each of the different modules?

### selection
All should only be supported for the flag version of the tool, when trying to add multiple via the tui the user should just call `devtool deploy` and be met with a selection screen (with a select all) that is able to select which specific items are needed

## pkg
- wait.go is not needed like wait.sh, waiting/polling should be accomplished based on if `devtool` is running in TUI or with parameters
- under resources, another folder of dashboards should be added, since the number of dashboards is going to be substantially higher than any of the other setups
- make the storage have a provider type structure, with minio being the only implementation right now. Since minio has removed its open source licenses, we will need to deprecate it in the near future
- We will be using bubbletea and it's surrounding packages for the ui, so the current UI folder is invalid
- Remove embedded templates, we will use go code and structs for all custom resource management

## Decision points:
1. Use cobra
2. New option, use CLI only when run with all needed flags. When run without all needed flags start up a tui
3. Use `controller-runtime/pkg/client`, but try to keep an abstraction over that client where possible to enable swapping to `client-go` later if neede
4. Looks mostly fine, I don't think we need the actual ConsolePlugin code as the UIPlugin from COO provides an abstraction above it that we will use. It was provided as an example of issues to look for
5. Can we change the config file to also just be like an exposed go var instead that way we keep type safety?
6. This seems mostly fine, but we should make sure that the way that errors are bubbled up is appropriate based on if the user is operating in TUI or flag mode
7.
This is a repository used to ease development and encourage sharing scripts and setups. We want to provide as low a burden as possible for users to contribute to it. Everyone should be expected to be able to debug and fix any issues. Automated testing, like github actions, is not needed. Testing should be extremely minimal and only created for extremely critical reasons, and even then should not be e2e tests, only unit tests
8. Same as 6, this should depend on if in TUI or flag mode
9. Same and 6 and 8
10. No templates, save all files in go. Templates can be constructed as go functions

## Other
### One
**Authentication Methods Supported**:
Only kubeconfig file, this tool will not be run in cluster

### Two
```
config.Timeout = 30 * time.Second // API request timeout
config.QPS = 50                   // Queries per second
config.Burst = 100                // Burst allowance
```
What is this even referring to? The kubernetes client?

## Questions for Team Discussion
1. obstool
2. Interactive only when flags missing
3. Yes
4. Our job to worry about timeline, do not consider it
5. Continue bash scripts until Go version ready. We will rebase and then update the go files based on changes to the bash scripts when ready
