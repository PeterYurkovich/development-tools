# Plan: Create Config Package

**Task ID**: Foundation → Configuration  
**Status**: Planning → Implementation  
**Date**: 2026-06-10  
**Dependencies**: ✅ Set up directory structure (complete)

---

## Overview

Implement the `pkg/config/` package that provides type-safe configuration defaults for obstool. This package uses Go structs with sensible defaults (no Viper, no config files) and allows runtime overrides via Cobra flags.

---

## Goals

1. **Create type-safe configuration struct** following the architecture decision (no Viper, no config files)
2. **Define sensible defaults** based on existing bash script configurations
3. **Expose as package-level variable** for easy access throughout the application
4. **Support runtime overrides** via flags (Cobra's native flag system)
5. **Maintain simplicity** - minimal abstraction, self-documenting

---

## Non-Goals

- ❌ No config file parsing (YAML, JSON, TOML) - per architecture decision
- ❌ No Viper or other config libraries
- ❌ No environment variable parsing (flags only)
- ❌ No validation logic (validate at usage point)
- ❌ No getters/setters (direct field access)
- ❌ No complex nested hierarchies (keep it flat where possible)

---

## Architecture Context

### From go-migration-plan.md (Lines 559-614)

**Decision**: Type-safe Go struct for configuration

**Key Pattern**:
```go
var Default = Config{
    DefaultBundle:    "...",
    DefaultNamespace: "...",
    Registry: RegistryConfig{...},
    Demo: DemoConfig{...},
}
```

**Benefits**:
- Type safety at compile time
- IDE autocomplete
- Easy to test
- No runtime config parsing errors
- Can be overridden programmatically if needed

### Configuration Strategy: Flags + Defaults

**How it works**:
1. **Config package provides DEFAULT values** (compile-time constants)
2. **Commands define Cobra flags** (runtime overrides)
3. **Flag values override defaults** when provided

```go
// User provides flag: use it
obstool users create --count=10

// User omits flag: use config.Default.Users.Count (6)
obstool users create
```

### Code Style Requirements (from CONTEXT.md)

- **Comments**: Minimal to none - code should be self-documenting
- **Variable Names**: No 1-2 letter names; descriptive names only
- **Simplicity**: Prefer simple, direct implementation over abstraction

---

## Configuration Values Analysis

Based on exploration of existing bash scripts, identified these major configuration areas:

### 1. **Registry Configuration**
- Default registries: quay, stage, brew, redhat
- Default org: `openshift-observability-ui`

### 2. **Namespace Defaults**
- COO namespace: `openshift-cluster-observability-operator`
- Logging namespace: `openshift-logging`
- Tracing namespace: `openshift-tracing`
- Monitoring namespace: `openshift-monitoring`
- Marketplace namespace: `openshift-marketplace`
- Config namespace: `openshift-config`

### 3. **Operator Defaults**
- Default bundle: `quay.io/openshift-observability-ui/observability-ui-operator-bundle:latest`
- Channel: `stable`
- Install mode: `AllNamespaces`
- Approval: `Automatic`

### 4. **User/Demo Defaults**
- User count: `6`
- Default password: `password`
- Username prefix: `user`
- OAuth provider: `my_htpasswd_provider`

### 5. **Timeout Defaults**
- Operator install: `900s` (15 minutes)
- API timeout: `30s`
- Rollout timeout: `600s` (10 minutes)
- Poll interval: `5s`

### 6. **Storage Defaults**
- Default storage class: `gp3-csi`
- Auto-detect: `true`

---

## Design: Config Package Structure

### File: `pkg/config/config.go`

**Package**: `config`

**Structure**:
```
pkg/config/
└── config.go    # Single file with all config structs and Default variable
```

### Top-Level Config Struct

```go
type Config struct {
    DefaultBundle    string
    DefaultNamespace string
    Registry         RegistryConfig
    Namespaces       NamespaceConfig
    Operator         OperatorConfig
    Users            UserConfig
    Timeouts         TimeoutConfig
    Storage          StorageConfig
}
```

### Nested Structs

#### 1. RegistryConfig
```go
type RegistryConfig struct {
    Quay         string
    QuayOrg      string
    StageURL     string
    RedHatURL    string
    BrewURL      string
}
```

#### 2. NamespaceConfig
```go
type NamespaceConfig struct {
    COO           string
    Logging       string
    Tracing       string
    Monitoring    string
    Marketplace   string
    Config        string
    NetObserv     string
    ACM           string
    TempoOperator string
    OTELOperator  string
    LokiOperator  string
}
```

#### 3. OperatorConfig
```go
type OperatorConfig struct {
    Channel        string
    InstallMode    string
    Approval       string
    SecurityConfig string
}
```

#### 4. UserConfig
```go
type UserConfig struct {
    Count           int
    DefaultPassword string
    UsernamePrefix  string
    OAuthProvider   string
    HtpasswdFile    string
}
```

#### 5. TimeoutConfig
```go
type TimeoutConfig struct {
    OperatorInstall time.Duration
    APIRequest      time.Duration
    Rollout         time.Duration
    PollInterval    time.Duration
}
```

#### 6. StorageConfig
```go
type StorageConfig struct {
    DefaultClass string
    AutoDetect   bool
}
```

---

## Implementation Plan

### Complete File: `pkg/config/config.go`

**Lines of Code**: ~120 lines

**Structure**:
1. Package declaration
2. Import `time`
3. Define all struct types (6 structs)
4. Define `Default` variable with all default values

---

## Code Style Compliance

### ✅ Minimal Comments
- No field comments (names are self-documenting)
- No package documentation (simple enough to understand)
- Type definitions speak for themselves

### ✅ Descriptive Variable Names
- `DefaultBundle` not `defBundle` or `db`
- `OperatorInstall` not `opInstall` or `oi`
- `QuayOrg` not `qorg` or `qo`

### ✅ Simple Implementation
- No validation in config (validate at usage point)
- No getters/setters
- Direct field access
- Single file, no complex abstractions

---

## Usage Examples

### 1. **Access Default Config**

```go
import "github.com/observability-ui/development-tools/pkg/config"

func createCatalogSource() string {
    return config.Default.Namespaces.Marketplace
}
```

### 2. **Override at Runtime with Flags**

```go
// cmd/users/create.go
func createUsers(cmd *cobra.Command, args []string) error {
    // Flag value if provided, otherwise config default
    count, _ := cmd.Flags().GetInt("count")
    if count == 0 {
        count = config.Default.Users.Count  // Use default: 6
    }
    
    password, _ := cmd.Flags().GetString("password")
    if password == "" {
        password = config.Default.Users.DefaultPassword  // Use default: "password"
    }
    
    // ... create users
}
```

### 3. **User Experience**

```bash
# Use all defaults (from config.Default)
obstool users create
# ^ Uses count=6, password="password"

# Override with flags
obstool users create --count=10 --password=mypass

# Mix: override some, default others
obstool users create --count=10
# ^ Uses count=10 (from flag), password="password" (from config.Default)
```

---

## Testing Strategy

Per project philosophy: **Minimal testing**

**Decision**: Skip unit tests for this package.

**Rationale**:
- Simple struct definitions with default values
- No logic to test
- No error cases
- Type safety enforced at compile time
- Will be tested implicitly through command usage

---

## Dependencies

### Required Imports

**Standard Library**:
- `time` - for `time.Duration` in TimeoutConfig

**No external dependencies needed**.

### go.mod Changes

**None** - only uses stdlib `time` package.

---

## Implementation Steps

### Step 1: Create Directory
```bash
mkdir -p pkg/config
```

### Step 2: Create config.go File
```bash
touch pkg/config/config.go
```

### Step 3: Implement File
- Add package declaration
- Add import for `time`
- Define all struct types
- Define `Default` variable with all default values

### Step 4: Verify Compilation
```bash
go build ./pkg/config/...
```

### Step 5: Format Code
```bash
go fmt ./pkg/config/...
```

### Step 6: Update TODO.md
Mark task as complete

---

## Success Criteria

✅ **File Created**:
- `pkg/config/config.go` exists

✅ **Compilation**:
- `go build ./pkg/config/...` succeeds
- No compilation errors
- No import errors

✅ **Code Quality**:
- Minimal comments (type names only if needed)
- Descriptive field names (no abbreviations)
- Follows project code style
- Formatted with `go fmt`

✅ **Completeness**:
- All config structs defined
- `Default` variable populated with sensible defaults
- Covers all major configuration areas from bash scripts

✅ **Integration Ready**:
- Can be imported by other packages
- All fields exported (public)
- Ready for use in command implementations

---

## Files Summary

### New Files Created
1. `pkg/config/config.go` (~120 lines)

### Modified Files
1. `tmp/TODO.md` - Mark config package task as complete

### No Changes Needed
- No dependencies to add to go.mod
- No scheme registration needed
- No tests to create

---

## Timeline Estimate

**Total Effort**: ~20-25 minutes

- Step 1-2 (setup): 2 minutes
- Step 3 (implementation): 10-12 minutes
- Step 4-5 (verification): 3 minutes
- Step 6 (documentation): 5 minutes
- Buffer: 5 minutes

---

## Future Extensions

Possible additions (when needed, not now):

**Monitoring/Plugin Config**:
- `MonitoringConfig` - monitoring plugin settings
- `DashboardConfig` - dashboard defaults

**Deployment Config**:
- `LoggingConfig` - logging stack defaults (data model, size, etc.)
- `TracingConfig` - tracing stack defaults

**Feature Flags**:
- `FeatureConfig` - boolean toggles for features

**Not needed now** - keep it simple, add when there's a clear need.

---

## Integration Points

### 1. Command Implementations (Future)

Will use config for defaults:
```go
// cmd/users/create.go
count := config.Default.Users.Count
password := config.Default.Users.DefaultPassword
```

### 2. Resource Creation (Future)

Will use config for namespaces, classes:
```go
// pkg/resources/lokistack.go
namespace := config.Default.Namespaces.Logging
storageClass := config.Default.Storage.DefaultClass
```

### 3. Operator Deployment (Future)

Will use config for operator settings:
```go
// pkg/operators/coo/fbc.go
catalogNS := config.Default.Namespaces.Marketplace
channel := config.Default.Operator.Channel
```

---

## Decisions Made

1. **Struct Granularity**: ✅ Nested structs for logical grouping (Registry, Namespaces, etc.)
   - Makes config self-organizing
   - Easier to understand what belongs where

2. **Field Names**: ✅ Full descriptive names (no abbreviations)
   - `DefaultBundle` not `Bundle`
   - `OperatorInstall` not `OpInstall`

3. **Duration vs Int**: ✅ Use `time.Duration` for timeouts
   - Type-safe (can't accidentally pass seconds as minutes)
   - Clearer intent: `15 * time.Minute` vs `900`

4. **Namespace Consolidation**: ✅ Single `NamespaceConfig` struct
   - All namespace defaults in one place
   - Easy to see all namespaces at a glance

5. **Comments**: ✅ Minimal to none
   - Field names are self-documenting
   - Types make purpose clear

6. **Validation**: ✅ None in config package
   - Validate at usage point
   - Keeps config simple
   - Commands responsible for their validation

7. **Testing**: ✅ Skip unit tests
   - Too simple to need tests
   - No logic to test
   - Will be tested through integration

8. **Flags vs Config Files**: ✅ Flags-only (Cobra native)
   - No Viper
   - No YAML/JSON parsing
   - Flags override defaults

---

## References

### Documentation
- **TODO.md**: Lines 57-63 (config package task)
- **CONTEXT.md**: Lines 90-120 (architecture overview)
- **go-migration-plan.md**: Lines 559-622 (Configuration Management decision)

### Code References
- **pkg/k8s/client.go**: Pattern for package structure
- **pkg/context/context.go**: Pattern for simple struct definitions

### External
- Go time package: https://pkg.go.dev/time
- Cobra flags: https://github.com/spf13/cobra

---

## Implementation Approved

All decisions confirmed:

1. ✅ Single file: `pkg/config/config.go`
2. ✅ Nested structs: `RegistryConfig`, `NamespaceConfig`, etc.
3. ✅ Package-level `Default` variable
4. ✅ Use `time.Duration` for timeouts
5. ✅ No validation logic
6. ✅ Skip unit tests
7. ✅ Minimal comments
8. ✅ Descriptive field names
9. ✅ Flags-only runtime overrides (no Viper, no config files)

**Ready for implementation.**

---

**Plan Status**: ✅ Complete - proceeding with implementation  
**Blockers**: None  
**Next Task After This**: Implement root command (`cmd/root.go`)
