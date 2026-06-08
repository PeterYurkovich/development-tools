# Development Tools Repository: Go Migration Plan

**Repository**: development-tools (OpenShift Observability UI Team)  
**Date**: June 8, 2026  
**Status**: Planning Phase  
**Purpose**: Comprehensive analysis and migration plan for converting the repository to Go

---

## Executive Summary

The development-tools repository has evolved into a critical toolset for the OpenShift Observability UI team but has become fragmented across 6+ technologies. This plan proposes migrating to a unified Go-based CLI tool to improve maintainability, type safety, and team productivity.

### Key Findings

- **Current State**: 58 bash scripts, 78 YAML files, 2 Makefiles, 2 Justfiles, 3 TypeScript files
- **Fragmentation Impact**: Multiple ways to do the same thing, inconsistent error handling, difficult to extend
- **Team Alignment**: Team writes Go for operators, making Go the natural choice
- **Migration Effort**: Estimated 16 weeks (4 months) for full migration

### Decisions Made

**Proceed with Full Go Migration** using:
- **CLI Tool Name**: `obstool` (observability tooling)
- **CLI Framework**: Cobra (industry standard for Kubernetes tools)
- **Client Library**: controller-runtime with abstraction layer
- **User Interaction**: Flags + TUI (CLI when all flags present, TUI when flags missing)
- **Configuration**: Type-safe Go structs (no config files)
- **Testing**: Minimal - unit tests only for critical code
- **UI Library**: Bubble Tea for TUI
- **Resource Definitions**: Go structs (no YAML templates)

---

## Documentation Structure

This plan consists of two comprehensive documents (plus user notes):

### 1. Main Migration Plan
**File**: [`go-migration-plan.md`](./go-migration-plan.md)

**Contents**:
- Current state analysis with repository statistics
- Target architecture (proposed CLI structure)
- Decision points for team consideration
- 10 key decisions with pros/cons/recommendations:
  - CLI framework choice (Cobra vs alternatives)
  - User interaction approach (CLI vs prompts vs TUI)
  - Kubernetes client strategy
  - Version detection & conditional logic
  - Configuration management
  - Error handling strategy
  - Testing strategy
  - Timeout & retry configuration
  - Logging & output
  - Embedded resources vs external files
- 6-phase migration plan with timeline
- Go modules required (with versions)
- Connection requirements
- Risk assessment
- Success metrics
- Next steps and questions for team discussion

**Key Highlights**:
- Detailed 16-week phased approach
- Architecture diagrams in code form
- Code examples for each decision
- Specific Go module versions needed
- Version detection pattern (from observability-operator PR #1100)

---

### 2. Go Modules Research
**File**: [`crd_go_modules_research.md`](./crd_go_modules_research.md)

**Contents**:
- Comprehensive research on Go modules for ALL Kubernetes CRDs used in the repository
- 18 different CRD types analyzed:
  - Observability CRDs (UIPlugin, LokiStack, TempoStack, etc.)
  - OpenShift resources (OAuth, Route, Project, Console, etc.)
  - OLM resources (Subscription, OperatorGroup, CatalogSource)
- **Critical Pattern**: Version-specific module requirements
  - Example: Console CRD requires different modules for OCP 4.17-4.18 vs 4.19+
  - Detailed analysis of why (ContentSecurityPolicy field marshalling)
- For each CRD:
  - Go module import path
  - Version-specific modules (if any)
  - Compatibility notes
  - Usage patterns
  - Code examples
- Decision trees for module selection
- Common integration patterns
- Troubleshooting guide
- Version compatibility matrix
- References to observability-operator implementation

**Key Highlights**:
- **957 lines** of detailed research
- Real-world examples from production code
- Version detection and conditional imports
- Forked vs upstream module decisions
- Type conversion patterns

---

### 3. User Notes & Decisions
**File**: [`user-notes.md`](./user-notes.md)

**Contents**:
- Team feedback and decisions on key choices
- Corrections to initial plan (e.g., ClusterLogForwarder ownership)
- Architecture refinements based on team input
- Specific implementation notes for:
  - COO deployment methods consolidation
  - Cleanup structure
  - Storage provider pattern
  - TUI vs CLI mode handling

**Key Decisions**:
- Tool name: `obstool`
- Interaction: Flags + TUI (not prompts)
- Testing: Minimal, unit tests only
- No templates: Go structs only
- Config: Type-safe Go variables

---

## Quick Reference

### Current Repository Stats

| Metric | Count | Percentage |
|--------|-------|------------|
| Shell Scripts | 58 | 73% |
| YAML Files | 78 | 17% |
| Justfiles | 2 | 5% |
| Makefiles | 2 | 3% |
| TypeScript Files | 3 | 2% |
| Go Files | 0 | 0% |

### Migration Timeline

| Phase | Duration | Deliverables |
|-------|----------|--------------|
| 1. Foundation | 2 weeks | Core framework, K8s client, version detection |
| 2. Core Operations | 3 weeks | Monitoring, users, dashboards commands |
| 3. Complex Deployments | 4 weeks | Logging, tracing, korrel8r stacks |
| 4. QE & Advanced | 3 weeks | QE workflows, demo orchestration |
| 5. Migration & Docs | 2 weeks | Documentation, team training |
| 6. Refinement | 2 weeks | Polish, remove old scripts |
| **Total** | **16 weeks** | **Complete migration** |

### Migration Strategy

| Approach | Status |
|----------|--------|
| **Full Go Migration** | ✅ **APPROVED** |
| Bash scripts | Continue until Go version ready |
| Timeline | Not a constraint - quality over speed |

---

## Key Decisions Made

All major decisions have been finalized:

### 1. Tool Name
- ✅ **`obstool`** (observability tooling)

### 2. Framework
- ✅ **Cobra** - Industry standard

### 3. User Interaction
- ✅ **Flags + TUI Hybrid**
  - CLI mode when all required flags provided
  - TUI mode when flags missing
  - No survey/promptui prompts

### 4. Client Library
- ✅ **controller-runtime** with abstraction layer
  - Allows future swap to client-go if needed

### 5. Configuration
- ✅ **Type-safe Go structs** (no config files)
  - Exposed as package-level variables

### 6. Testing
- ✅ **Minimal** - unit tests for critical code only
  - No CI/CD automation
  - No E2E tests

### 7. UI Library
- ✅ **Bubble Tea** for TUI components

### 8. Resources
- ✅ **Go structs only** (no YAML templates)

### 9. Repository
- ✅ **Same repo** (development-tools)

---

## Technology Stack

### Core Dependencies
```
CLI Framework:          github.com/spf13/cobra
TUI Library:           github.com/charmbracelet/bubbletea
                       github.com/charmbracelet/lipgloss
K8s Client:            sigs.k8s.io/controller-runtime (with abstraction)
OpenShift API:         github.com/openshift/api
                       github.com/openshift/cluster-logging-operator (for ClusterLogForwarder)
Operator Framework:    github.com/operator-framework/api
Output/Colors:         github.com/fatih/color
Testing:               Standard library testing (minimal)
Version Comparison:    golang.org/x/mod/semver
```

### CRD Dependencies (18 total)
See [crd_go_modules_research.md](./crd_go_modules_research.md) for complete list with versions.

---

## Example Commands (Target State)

### Deploy Operations
```bash
# Deploy everything (CLI mode with flags)
obstool deploy all

# Deploy specific stacks (CLI mode)
obstool deploy logging --data-model=otel --namespace=openshift-logging
obstool deploy tracing --size=1x.small
obstool deploy dashboards
obstool deploy korrel8r

# Deploy with TUI (missing flags triggers TUI)
obstool deploy logging
# TUI displays selection interface for data model, namespace, etc.

# COO deployment with method selection
obstool deploy coo --method=bundle --bundle-url=quay.io/...
obstool deploy coo --method=operatorhub
```

### COO Upgrade
```bash
# Upgrade COO in place
obstool upgrade coo --to-version=1.5.0
```

### Monitoring Management
```bash
# Scale CMO
obstool monitoring scale down
obstool monitoring scale up

# Update plugin image
obstool monitoring update-image quay.io/user/plugin:v1.2.3
```

### User Management
```bash
# Create test users
obstool users create --count=6

# Apply RBAC
obstool users rbac --scenario=perses-e2e
```

### Cleanup
```bash
# Cleanup specific components (mirrors deploy structure)
obstool cleanup coo
obstool cleanup logging
obstool cleanup tracing

# Cleanup everything (flag mode only - requires all flags)
obstool cleanup all --confirm=yes
```

**Note**: Demo workflows are achieved by combining the above commands in sequence.

---

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| **Migration takes longer** | Phased approach, parallel run of old scripts |
| **Team resistance** | Early involvement, training, feedback loops |
| **Loss of functionality** | Comprehensive requirements, feature parity testing |
| **Bugs in new implementation** | Extensive testing, gradual rollout |
| **Breaking changes** | Keep old scripts until fully validated |

---

## Success Criteria

### Technical
- ✅ 100% functional parity with bash scripts
- ✅ Minimal unit tests for critical code paths
- ✅ Fast startup time
- ✅ All CRD types supported
- ✅ Version detection working
- ✅ Both CLI and TUI modes functional

### Adoption
- ✅ Team can use new tool for daily work
- ✅ Zero bash scripts eventually (timeline flexible)
- ✅ Easy for team to contribute and debug
- ✅ Low barrier to entry for new features

---

## Next Steps

See **[TODO.md](./TODO.md)** for detailed implementation checklist with:
- Atomic tasks organized by command/feature
- Dependency tracking (blocked by)
- Status tracking: `[ ]` Todo | `[~]` In Progress | `[x]` Complete
- Parallel work opportunities clearly identified

**Migration Strategy**:
- Bash scripts continue to work during development
- Rebase Go implementation when bash scripts change
- Update Go code to match any bash script updates
- No rush - maintain quality over speed

**Recommended Starting Point**: Foundation tasks (project setup, k8s client, version detection, execution context)

---

## Document Versions

| Document | Version | Last Updated | Status |
|----------|---------|--------------|--------|
| go-migration-plan.md | 1.1 | June 8, 2026 | Updated with decisions |
| crd_go_modules_research.md | 1.1 | June 8, 2026 | Corrected ClusterLogForwarder |
| user-notes.md | 1.0 | June 8, 2026 | Team feedback captured |
| README.md (this file) | 1.1 | June 8, 2026 | Updated with decisions |

---

## References

### Internal
- Observability Operator PR #1100 (version-specific CRD pattern)
- Development-tools repository current state
- Team practices and conventions

### External
- [Cobra Framework](https://github.com/spf13/cobra) - CLI framework
- [Perses Login K8s Implementation](https://github.com/perses/perses/blob/main/internal/cli/cmd/login/k8s.go) - Connection pattern reference
- [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime) - K8s client library
- [OpenShift API](https://github.com/openshift/api) - OpenShift CRDs
- [Operator Framework API](https://github.com/operator-framework/api) - OLM CRDs

---

## Contact & Feedback

**Questions?** Open a discussion in the team channel.

**Found an issue with the plan?** File an issue or submit a PR to update.

**Ready to proceed?** Review the decision points and schedule a team meeting.

---

## Appendix: Document Map

```
📁 ./tmp/
├── 📄 README.md (this file)
│   └── Overview, decisions, and quick reference
│
├── 📄 go-migration-plan.md
│   ├── Current state analysis
│   ├── Target architecture (updated with team feedback)
│   ├── 10 key decisions (finalized)
│   ├── Implementation patterns
│   ├── Go modules required
│   ├── Connection requirements
│   └── Code examples
│
├── 📄 crd_go_modules_research.md
│   ├── 18 CRD types analyzed
│   ├── Version-specific patterns
│   ├── Module selection decision trees
│   ├── Code examples
│   ├── Compatibility matrix
│   └── Corrected ClusterLogForwarder info
│
└── 📄 user-notes.md
    ├── Team feedback on initial plan
    ├── Corrections and refinements
    ├── Architecture decisions
    └── Implementation notes
```

---

**Status**: ✅ Decisions Made, Ready for Implementation  
**Action Required**: Begin Phase 1 (Foundation) when ready  
**Timeline**: Quality over speed - no rush
