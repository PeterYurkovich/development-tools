# development-tools
A centralized location for the openshift observability-ui team to keep development-related scripts and information.

## 🚀 obstool - New Go CLI (In Development)

We're migrating to a unified Go CLI tool called `obstool` that replaces the bash/make/just/yaml tooling.

**Documentation**: See `tmp/` directory
- **[tmp/README.md](./tmp/README.md)** - Start here (overview, status, navigation)
- **[tmp/TASKS.md](./tmp/TASKS.md)** - Implementation tracking
- **[tmp/ARCHITECTURE.md](./tmp/ARCHITECTURE.md)** - System design
- **[tmp/PATTERNS.md](./tmp/PATTERNS.md)** - Code patterns
- **[tmp/REFERENCE.md](./tmp/REFERENCE.md)** - Technical reference

**Status**: Foundation complete, commands in progress  
**Timeline**: No pressure - quality over speed, bash scripts continue to work

---

## Legacy Tooling (Current)

## Development
This repo uses a [Justfile](https://github.com/casey/just) to maintain it's commands. To pass your own image of the observability operator bundle use the flag `OPERATOR_BUNDLE`. For example:

```bash
just coo-resources OPERATOR_BUNDLE="quay.io/test/observability-operator-bundle"
```

The repo uses an up/down system for manipulating your cluster, for example:

```bash
just scale-monitoring down
```
Will scale down the Cluster Monitoring Operator and the monitoring-plugin.

```bash
just scale-monitoring up
```
Will scale the Cluster Monitoring Operator and the monitoring-plugin back up to their original number of replicas.

This repo uses the bash-language-server to lint and format files:
- [vscode](https://marketplace.visualstudio.com/items?itemName=mads-hartmann.bash-ide-vscode)
- [zed](https://github.com/bash-lsp/bash-language-server)
- [lsp](https://github.com/bash-lsp/bash-language-server)

This repo uses [bun](https://bun.sh/) to bundle create and bundle a fork of the shikijs/cli which is able to have command results piped into it.
