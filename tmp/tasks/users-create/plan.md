# Implementation Plan: `obstool users create` Command

**Status**: Ready for Implementation  
**Created**: 2026-06-18  
**Updated**: 2026-06-18 (incorporated user feedback)  
**Implements**: [TASKS.md](../../TASKS.md) - Users Create Command

---

## Quick Summary

**What**: Create test users with htpasswd auth and varied RBAC permissions for testing  
**Users**: Minimum 6 users (user1...userN) with shared password  
**RBAC**: Simplified namespace-focused permissions (see [rbac-permissions.md](./rbac-permissions.md) or [RBAC-SUMMARY.md](./RBAC-SUMMARY.md))  
**Namespaces**: Auto-creates perses-dev, openshift-cluster-observability-operator, and --namespace value  
**Steps**: Generate htpasswd → Create Secret → Patch OAuth → Ensure Namespaces → Apply RBAC  
**Key Changes**: Simplified RBAC (15 bindings vs 22), namespace-scoped focus for users 3-6

---

## Overview

Implement `obstool users create` command to create test users for OpenShift clusters using htpasswd authentication. This replaces the bash script `users/create-htpsswd-auth.sh` with a Go-based implementation following obstool patterns.

**Key Features:**
- Varied RBAC permissions per user for comprehensive testing
- Automatic namespace creation if not found
- Idempotent operations (safe to run multiple times)
- Works in both CLI and TUI modes

---

## Requirements

### Functional Requirements

**What it does:**
1. Generates htpasswd file with N test users (minimum 6)
2. Creates/updates Secret `htpass-secret` in `openshift-config` namespace
3. Patches cluster OAuth CR to add htpasswd identity provider
4. Applies RBAC permissions to first 6 users only (mandatory, no opt-out)

**User Format:**
- Username pattern: `user1`, `user2`, `user3`, ..., `userN`
- All users share same password (testing only)
- Minimum count: 6 users
- RBAC applied to users 1-6 only (even if count > 6)

**RBAC Permissions (first 6 users):**

Each user receives different permissions to enable comprehensive RBAC testing.
See [rbac-permissions.md](./rbac-permissions.md) for complete permissions matrix.

**Summary:**
- user1: Admin-like (cluster-wide read+write)
- user2: Read-only (cluster-wide read-only)
- user3: Multi-namespace editor (perses-dev + openshift-cluster-observability-operator, read+write)
- user4: Single-namespace editor (perses-dev, read+write)
- user5: Single-namespace viewer (perses-dev, read-only)
- user6: Dashboards + Metrics viewer (perses-dev, specific resources only)

**Total RBAC resources created:** ~15 bindings (8 namespace-scoped, 7 cluster-scoped)

**Namespaces created automatically:**
- `perses-dev` (for users 3-6)
- `openshift-cluster-observability-operator` (for user3)
- Namespace from `--namespace` flag (for users 1-2)

**Behavior:**
- No wait after OAuth creation (users available within ~60s)
- No browser opening
- Idempotent (safe to run multiple times)

### Flags

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--count` | int | 6 | No | Number of users to create (minimum 6) |
| `--password` | string | "password" | No | Password for all users |
| `--namespace` | string | "openshift-monitoring" | No | Namespace for RBAC permissions |

**Validation:**
- `count` must be >= 6
- `password` cannot be empty
- `namespace` must exist (error if not found)

### Mode Support

**CLI Mode** (all flags or defaults):
```bash
obstool users create
obstool users create --count=10 --password=testpass --namespace=openshift-logging
```

**TUI Mode** (interactive prompts):
- Prompt for count (default: 6, validate >= 6)
- Prompt for password (masked input, validate not empty)
- Prompt for namespace (default: openshift-monitoring)

---

## Architecture

### Command Flow

```
cmd/users/create.go
  ↓
Mode Detection (CLI vs TUI)
  ↓
Validate inputs (count >= 6, namespace exists)
  ↓
pkg/operations/users.go::CreateUsers()
  ├─ Step 1: Generate htpasswd data (pkg/users/htpasswd.go)
  ├─ Step 2: Create/update Secret (pkg/users/oauth.go)
  ├─ Step 3: Patch OAuth CR (pkg/users/oauth.go)
  └─ Step 4: Apply RBAC to users 1-6 (pkg/users/rbac.go)
  ↓
Output results (CLI handler or TUI progress)
```

### Executor Pattern

**Step Constants:**
```go
const (
    StepGenerateHtpasswd = iota
    StepCreateSecret
    StepPatchOAuth
    StepEnsureNamespaces
    StepApplyRBAC
)
```

**Progress Updates:**
- Step 1: "Generate htpasswd data for N users"
- Step 2: "Create/update htpass-secret in openshift-config"
- Step 3: "Patch OAuth CR with htpasswd provider"
- Step 4: "Apply RBAC to 6 users in {namespace}"

---

## Implementation Details

### 1. File Structure

**New Files:**
```
cmd/users/
├── users.go              # Users command group
└── create.go             # Create command implementation

pkg/users/
├── htpasswd.go           # htpasswd generation utilities
├── oauth.go              # OAuth CR patching
└── rbac.go               # RBAC role binding creation

pkg/operations/
└── users.go              # CreateUsers business logic

internal/constants/
└── users.go              # User-related constants
```

**Modified Files:**
```
pkg/k8s/scheme.go         # Add rbacv1 scheme registration
cmd/root.go               # Register users command group
```

---

### 2. Constants (`internal/constants/users.go`)

```go
package constants

const (
    HTPasswdSecretName    = "htpass-secret"
    HTPasswdSecretKey     = "htpasswd"
    HTPasswdProviderName  = "my_htpasswd_provider"
    OAuthCRName           = "cluster"
    OpenshiftConfigNS     = "openshift-config"
    MinUserCount          = 6
    RBACUserCount         = 6
)
```

---

### 3. htpasswd Generation (`pkg/users/htpasswd.go`)

**Purpose**: Generate bcrypt-hashed htpasswd file content

**Functions:**

```go
// GenerateHtpasswdData creates htpasswd file content for N users
// Format: "user1:$2y$...\nuser2:$2y$...\n"
// Returns: []byte (htpasswd file content), error
func GenerateHtpasswdData(count int, password string) ([]byte, error)

// generateBcryptHash creates bcrypt hash compatible with htpasswd
// Matches: htpasswd -B -b flag (bcrypt)
func generateBcryptHash(password string) (string, error)
```

**Implementation Notes:**
- Use `golang.org/x/crypto/bcrypt` with `bcrypt.DefaultCost`
- Format: `user{N}:{bcrypt_hash}\n` for each user
- All users share the same password hash (performance: hash once, reuse)

**Example Output:**
```
user1:$2y$05$...hash...
user2:$2y$05$...hash...
user3:$2y$05$...hash...
```

---

### 4. OAuth Management (`pkg/users/oauth.go`)

**Purpose**: Create Secret and patch OAuth CR

**Functions:**

```go
// EnsureHTPasswdSecret creates or updates htpass-secret in openshift-config
// Returns: (created bool, error)
func EnsureHTPasswdSecret(ctx context.Context, kubeClient client.Client, 
                         htpasswdData []byte) (bool, error)

// EnsureOAuthHTPasswdProvider ensures htpasswd provider exists in OAuth CR
// Returns: (created bool, error)
func EnsureOAuthHTPasswdProvider(ctx context.Context, kubeClient client.Client) (bool, error)
```

**Implementation Notes:**

**EnsureHTPasswdSecret:**
- Check if Secret exists in `openshift-config` namespace
- If exists: Update `.data.htpasswd` field
- If not: Create new Secret
- Return `true` if created, `false` if updated

**EnsureOAuthHTPasswdProvider:**
- Get OAuth CR named "cluster"
- Check `spec.identityProviders` for existing htpasswd provider
- If found with matching name: return `false, nil` (already exists)
- If not found: append to `spec.identityProviders`
- Patch OAuth CR
- Return `true` if added, `false` if already existed

**OAuth IdentityProvider Spec:**
```go
configv1.IdentityProvider{
    Name:          constants.HTPasswdProviderName,
    MappingMethod: "claim",
    IdentityProviderConfig: configv1.IdentityProviderConfig{
        Type: configv1.IdentityProviderTypeHTPasswd,
        HTPasswd: &configv1.HTPasswdIdentityProvider{
            FileData: configv1.SecretNameReference{
                Name: constants.HTPasswdSecretName,
            },
        },
    },
}
```

---

### 5. RBAC Management (`pkg/users/rbac.go`)

**Purpose**: Apply varied role bindings to first 6 users for comprehensive testing

See [rbac-permissions.md](./rbac-permissions.md) for complete permissions matrix.

**Functions:**

```go
// ApplyUserRBAC applies observability RBAC to first 6 users
// Each user gets different permissions for testing purposes
func ApplyUserRBAC(ctx context.Context, kubeClient client.Client, 
                  namespace string, exec *executor.Executor) error

// createRoleBinding creates a RoleBinding in namespace (idempotent)
func createRoleBinding(ctx context.Context, kubeClient client.Client,
                      namespace, name, roleName, userName string) error

// createClusterRoleBinding creates a ClusterRoleBinding (idempotent)
func createClusterRoleBinding(ctx context.Context, kubeClient client.Client,
                             name, roleName, userName string) error
```

**Implementation - User-specific RBAC:**

**User1 (Admin-like):**
- ClusterRoleBinding: `cluster-admin`, `cluster-monitoring-view`, `cluster-logging-application-view`, `distributed-tracing-view`
- RoleBinding (--namespace flag): `admin`

**User2 (Read-only):**
- ClusterRoleBinding: `cluster-monitoring-view`, `cluster-logging-application-view`, `distributed-tracing-view`
- RoleBinding (--namespace flag): `view`

**User3 (Multi-namespace editor):**
- RoleBinding (perses-dev): `edit`
- RoleBinding (openshift-cluster-observability-operator): `edit`

**User4 (Single-namespace editor):**
- RoleBinding (perses-dev): `edit`

**User5 (Single-namespace viewer):**
- RoleBinding (perses-dev): `view`

**User6 (Dashboards + Metrics viewer):**
- RoleBinding (perses-dev): Custom role for PersesDashboards view
- RoleBinding (perses-dev): Custom role for metrics resources view

**Total bindings:** 8 RoleBindings + 7 ClusterRoleBindings = 15 total

**Namespaces to ensure exist:**
- `perses-dev`
- `openshift-cluster-observability-operator`
- Value from `--namespace` flag (default: `openshift-monitoring`)

**Idempotency:**
- Use `client.Create()` with `errors.IsAlreadyExists()` check
- If already exists, skip silently (no log spam)
- If ClusterRole doesn't exist, still create binding (will activate when role is created)

---

### 6. Business Logic (`pkg/operations/users.go`)

**Config Struct:**
```go
type CreateUsersConfig struct {
    Count     int
    Password  string
    Namespace string
}
```

**Main Function:**
```go
func CreateUsers(ctx context.Context, kubeClient client.Client, 
                config CreateUsersConfig, exec *executor.Executor) error {
    defer exec.Close()
    
    // Step 1: Generate htpasswd data
    stepName := fmt.Sprintf("Generate htpasswd data for %d users", config.Count)
    exec.SendUpdate(StepGenerateHtpasswd, executor.StatusInProgress, stepName)
    exec.SendLog(StepGenerateHtpasswd, "Hashing password with bcrypt")
    
    htpasswdData, err := htpasswd.GenerateHtpasswdData(config.Count, config.Password)
    if err != nil {
        exec.SendUpdateWithError(StepGenerateHtpasswd, executor.StatusFailed, stepName, err)
        return err
    }
    exec.SendLog(StepGenerateHtpasswd, fmt.Sprintf("Generated credentials for users 1-%d", config.Count))
    exec.SendUpdate(StepGenerateHtpasswd, executor.StatusComplete, stepName)
    
    // Step 2: Create/update Secret
    stepName = "Create htpass-secret in openshift-config"
    exec.SendUpdate(StepCreateSecret, executor.StatusInProgress, stepName)
    
    created, err := oauth.EnsureHTPasswdSecret(ctx, kubeClient, htpasswdData)
    if err != nil {
        exec.SendUpdateWithError(StepCreateSecret, executor.StatusFailed, stepName, err)
        return err
    }
    if created {
        exec.SendLog(StepCreateSecret, "Secret created")
    } else {
        exec.SendLog(StepCreateSecret, "Secret updated (already existed)")
    }
    exec.SendUpdate(StepCreateSecret, executor.StatusComplete, stepName)
    
    // Step 3: Patch OAuth CR
    stepName = "Patch OAuth CR with htpasswd provider"
    exec.SendUpdate(StepPatchOAuth, executor.StatusInProgress, stepName)
    
    created, err = oauth.EnsureOAuthHTPasswdProvider(ctx, kubeClient)
    if err != nil {
        exec.SendUpdateWithError(StepPatchOAuth, executor.StatusFailed, stepName, err)
        return err
    }
    if created {
        exec.SendLog(StepPatchOAuth, "HTPasswd provider added to OAuth")
    } else {
        exec.SendLog(StepPatchOAuth, "HTPasswd provider already configured")
    }
    exec.SendUpdate(StepPatchOAuth, executor.StatusComplete, stepName)
    
    // Step 4: Ensure required namespaces exist
    stepName = "Ensure required namespaces exist"
    exec.SendUpdate(StepEnsureNamespaces, executor.StatusInProgress, stepName)
    
    namespaces := []string{
        "perses-dev",
        "openshift-cluster-observability-operator",
        config.Namespace, // For user1 and user2
    }
    
    for _, ns := range namespaces {
        created, err := k8s.EnsureNamespace(ctx, kubeClient, ns)
        if err != nil {
            exec.SendUpdateWithError(StepEnsureNamespaces, executor.StatusFailed, stepName, err)
            return err
        }
        if created {
            exec.SendLog(StepEnsureNamespaces, fmt.Sprintf("Created namespace: %s", ns))
        } else {
            exec.SendLog(StepEnsureNamespaces, fmt.Sprintf("Namespace already exists: %s", ns))
        }
    }
    exec.SendUpdate(StepEnsureNamespaces, executor.StatusComplete, stepName)
    
    // Step 5: Apply RBAC to first 6 users
    stepName = "Apply varied RBAC to 6 users"
    exec.SendUpdate(StepApplyRBAC, executor.StatusInProgress, stepName)
    exec.SendLog(StepApplyRBAC, "Creating role bindings with different permissions per user")
    
    err = rbac.ApplyUserRBAC(ctx, kubeClient, config.Namespace, exec)
    if err != nil {
        exec.SendUpdateWithError(StepApplyRBAC, executor.StatusFailed, stepName, err)
        return err
    }
    exec.SendLog(StepApplyRBAC, "Created 15 role bindings (8 namespace + 7 cluster)")
    exec.SendUpdate(StepApplyRBAC, executor.StatusComplete, stepName)
    
    return nil
}
```

---

### 7. Command Implementation (`cmd/users/create.go`)

**Command Definition:**
```go
var createCmd = &cobra.Command{
    Use:   "create",
    Short: "Create test users with htpasswd authentication",
    Long: `Creates test users for OpenShift development and testing.

Generates N users (minimum 6) with htpasswd authentication and applies
RBAC permissions to the first 6 users in the specified namespace.

Users are named: user1, user2, user3, ..., userN
All users share the same password.

RBAC permissions applied to users 1-6:
  - view (namespace-scoped)
  - cluster-logging-application-view (cluster-scoped)
  - monitoring-rules-edit (cluster-scoped)
  - cluster-monitoring-view (cluster-scoped)

Note: Users will be available for login within ~60 seconds after creation.`,
    Example: `  # Create 6 users with default password
  obstool users create

  # Create 10 users with custom password
  obstool users create --count=10 --password=mypass

  # Create users with RBAC in custom namespace
  obstool users create --namespace=openshift-logging`,
    RunE: runUsersCreate,
}

func init() {
    createCmd.Flags().Int("count", 6, "Number of users to create (minimum 6)")
    createCmd.Flags().String("password", "password", "Password for all users")
    createCmd.Flags().String("namespace", "openshift-monitoring", "Namespace for RBAC permissions")
    UsersCmd.AddCommand(createCmd)
}
```

**Mode Detection:**
```go
func runUsersCreate(cmd *cobra.Command, args []string) error {
    // No required flags - all have defaults, but we validate count >= 6
    
    useTUI, err := mode.DetermineMode(cmd, []string{})
    if err != nil {
        return err
    }
    
    if useTUI {
        return runUsersCreateTUI(cmd)
    }
    
    return runUsersCreateCLI(cmd)
}
```

**CLI Mode:**
```go
func runUsersCreateCLI(cmd *cobra.Command) error {
    count, _ := cmd.Flags().GetInt("count")
    password, _ := cmd.Flags().GetString("password")
    namespace, _ := cmd.Flags().GetString("namespace")
    
    // Validation
    if count < constants.MinUserCount {
        return fmt.Errorf("count must be at least %d, got %d", constants.MinUserCount, count)
    }
    
    // Default password if empty
    if password == "" {
        password = "password"
    }
    
    ctx := cmd.Context()
    ctx = execctx.WithTUI(ctx, false)
    
    kubeClient, err := execctx.GetClient(ctx)
    if err != nil {
        return err
    }
    
    handler := output.NewCLIHandler()
    exec := executor.NewExecutor()
    
    go operations.CreateUsers(ctx, kubeClient, operations.CreateUsersConfig{
        Count:     count,
        Password:  password,
        Namespace: namespace,
    }, exec)
    
    for update := range exec.UpdateCh {
        if err := handler.HandleUpdate(update); err != nil {
            return err
        }
    }
    
    return nil
}
```

**TUI Mode:**
```go
func runUsersCreateTUI(cmd *cobra.Command) error {
    ctx := cmd.Context()
    ctx = execctx.WithTUI(ctx, true)
    
    kubeClient, err := execctx.GetClient(ctx)
    if err != nil {
        return err
    }
    
    // Collect inputs if not provided
    count, _ := cmd.Flags().GetInt("count")
    password, _ := cmd.Flags().GetString("password")
    namespace, _ := cmd.Flags().GetString("namespace")
    
    config, err := collectUsersInput(count, password, namespace)
    if err != nil {
        return err
    }
    
    // Note: Namespace will be created in operations.CreateUsers() if needed
    
    operationsList := []string{
        fmt.Sprintf("Generate htpasswd data for %d users", config.Count),
        "Create htpass-secret in openshift-config",
        "Patch OAuth CR with htpasswd provider",
        "Ensure required namespaces exist (perses-dev, openshift-cluster-observability-operator, etc.)",
        "Apply varied RBAC to 6 users",
    }
    
    model := tui.NewProgressModel("Creating Users", operationsList)
    program := tea.NewProgram(model)
    
    exec := executor.NewExecutor()
    
    go operations.CreateUsers(ctx, kubeClient, config, exec)
    
    go func() {
        for update := range exec.UpdateCh {
            if update.Message != "" {
                continue
            }
            
            program.Send(tui.OperationUpdateMsg{
                Index:  update.Index,
                Status: convertStatus(update.Status),
                Error:  update.Error,
            })
        }
    }()
    
    finalModel, err := program.Run()
    if err != nil {
        return err
    }
    
    model = finalModel.(tui.ProgressModel)
    if model.Error() != nil {
        return model.Error()
    }
    
    return nil
}
```

**Input Collection (TUI):**
```go
func collectUsersInput(defaultCount int, defaultPassword, defaultNamespace string) (operations.CreateUsersConfig, error) {
    var config operations.CreateUsersConfig
    config.Count = defaultCount
    config.Password = defaultPassword
    config.Namespace = defaultNamespace
    
    var countStr string
    if defaultCount > 0 {
        countStr = fmt.Sprintf("%d", defaultCount)
    }
    
    form := huh.NewForm(
        huh.NewGroup(
            huh.NewInput().
                Title("Number of users to create").
                Description("Minimum 6 required. RBAC applied to first 6 users only.").
                Placeholder("6").
                Value(&countStr).
                Validate(func(s string) error {
                    count, err := strconv.Atoi(s)
                    if err != nil {
                        return fmt.Errorf("must be a number")
                    }
                    if count < constants.MinUserCount {
                        return fmt.Errorf("minimum %d users required", constants.MinUserCount)
                    }
                    return nil
                }),
            huh.NewInput().
                Title("Password for all users").
                Description("All users will share this password").
                EchoMode(huh.EchoModePassword).
                Placeholder("password").
                Value(&config.Password),
                // Note: Empty password defaults to "password" in business logic
            huh.NewInput().
                Title("Namespace for RBAC permissions").
                Description("Will be created if it doesn't exist").
                Placeholder("openshift-monitoring").
                Value(&config.Namespace),
        ),
    )
    
    err := form.Run()
    if err != nil {
        return config, err
    }
    
    count, _ := strconv.Atoi(countStr)
    config.Count = count
    
    return config, nil
}
```

**Note**: Removed `validateNamespaceExists()` - namespace creation now handled in business logic via `EnsureNamespace()` pattern.

---

### 8. Scheme Registration (`pkg/k8s/scheme.go`)

**Add rbacv1 registration:**
```go
import (
    rbacv1 "k8s.io/api/rbac/v1"
)

func registerSchemes(scheme *runtime.Scheme) error {
    // ... existing registrations ...
    
    if err := rbacv1.AddToScheme(scheme); err != nil {
        return fmt.Errorf("failed to register rbac/v1 scheme: %w", err)
    }
    
    return nil
}
```

### 9. Namespace Utilities (`pkg/k8s/namespace.go` or add to existing file)

**Create namespace utility following Ensure* pattern:**
```go
// EnsureNamespace creates namespace if it doesn't exist
// Returns: (created bool, error)
func EnsureNamespace(ctx context.Context, kubeClient client.Client, name string) (bool, error) {
    namespace := &corev1.Namespace{
        ObjectMeta: metav1.ObjectMeta{
            Name: name,
        },
    }
    
    err := kubeClient.Create(ctx, namespace)
    if err != nil {
        if errors.IsAlreadyExists(err) {
            return false, nil  // Already exists
        }
        return false, fmt.Errorf("failed to create namespace %s: %w", name, err)
    }
    
    return true, nil  // Created
}
```

---

### 9. Command Group Registration (`cmd/users/users.go`)

```go
package users

import (
    "github.com/spf13/cobra"
)

var UsersCmd = &cobra.Command{
    Use:   "users",
    Short: "Manage test users and RBAC",
    Long:  "Create and manage test users with htpasswd authentication and RBAC permissions",
}
```

**Add to root command (`cmd/root.go`):**
```go
import (
    "github.com/observability-ui/development-tools/cmd/users"
)

func init() {
    rootCmd.AddCommand(users.UsersCmd)
}
```

---

## Error Handling

### Validation Errors

| Scenario | Error Message | Action |
|----------|---------------|--------|
| Count < 6 | "count must be at least 6, got N" | Return error, exit |
| Empty password | (none - defaults to "password") | Auto-default, continue |
| Namespace doesn't exist | (none - created automatically) | Create namespace, continue |

### Runtime Errors

| Scenario | Handling | User Impact |
|----------|----------|-------------|
| Secret create fails | Fail at Step 2, return error | Command fails, no OAuth/RBAC changes |
| OAuth patch fails | Fail at Step 3, return error | Secret exists, but OAuth not configured |
| RBAC creation fails | Log error, continue to next binding | Partial RBAC (some users may work) |
| All RBAC fails | Fail at Step 4, return error | Users exist but no permissions |

### Edge Cases

**Secret already exists:**
- Update existing secret with new htpasswd data
- Log: "Secret updated (already existed)"
- Success

**OAuth already has htpasswd provider:**
- Skip patching (idempotent)
- Log: "HTPasswd provider already configured"
- Success

**RoleBinding already exists:**
- Skip creation for that binding
- No log (too verbose)
- Continue to next binding

**ClusterRole doesn't exist:**
- Create RoleBinding anyway (binding will activate when role is created by operator)
- No warning needed (RoleBindings can reference non-existent ClusterRoles)
- This is normal for operator-provided roles that may not be installed yet

---

## Testing Plan

### Manual Testing

**Test 1: Default values (CLI mode)**
```bash
obstool users create
```
Expected:
- Creates 6 users (user1...user6)
- Password: "password"
- RBAC in openshift-monitoring
- Success message

**Test 2: Custom values (CLI mode)**
```bash
obstool users create --count=10 --password=testpass --namespace=openshift-logging
```
Expected:
- Creates 10 users (user1...user10)
- RBAC only for users 1-6
- RBAC in openshift-logging
- Success message

**Test 3: TUI mode**
```bash
obstool users create
# (in terminal without all flags, should prompt)
```
Expected:
- Interactive form appears
- Validates inputs
- Shows progress
- Success

**Test 4: Idempotency**
```bash
obstool users create
obstool users create  # Run again
```
Expected:
- Both runs succeed
- Second run: "Secret updated", "Provider already configured"
- No errors

**Test 5: Minimum count validation**
```bash
obstool users create --count=3
```
Expected:
- Error: "count must be at least 6, got 3"
- No changes made

**Test 6: Non-existent namespace**
```bash
obstool users create --namespace=test-new-namespace
```
Expected:
- Namespace `test-new-namespace` created
- Users created successfully
- RBAC applied in new namespace
- Success message

**Test 7: Empty password**
```bash
obstool users create --password=""
```
Expected:
- Password defaults to "password"
- Users created successfully
- Can login with `oc login --username=user1 --password=password`

**Test 8: RBAC verification per user**
```bash
obstool users create

# Test user1 (admin-like)
oc login --username=user1 --password=password
oc create -f test-prometheusrule.yaml -n openshift-monitoring  # Should succeed
oc delete prometheusrule test -n openshift-monitoring          # Should succeed

# Test user2 (read-only)
oc login --username=user2 --password=password
oc get prometheusrules -A                                      # Should succeed
oc create -f test-prometheusrule.yaml                          # Should fail (forbidden)

# Test user3 (logging specialist)
oc login --username=user3 --password=password
oc get clusterlogforwarder -n openshift-monitoring            # Should succeed
oc create -f test-prometheusrule.yaml                         # Should fail (no permission)

# Test user4 (monitoring specialist)
oc login --username=user4 --password=password
oc create -f test-prometheusrule.yaml                         # Should succeed
oc get clusterlogforwarder -A                                 # Should fail (no permission)

# Test user5 (tracing specialist)
oc login --username=user5 --password=password
oc get tempostack -n openshift-monitoring                     # Should succeed (if operator installed)
oc get prometheusrules                                        # Should fail (no permission)

# Test user6 (metrics+dashboards read-only)
oc login --username=user6 --password=password
oc get prometheusrules -A                                     # Should succeed
oc create -f test-prometheusrule.yaml                         # Should fail (read-only)
```
Expected:
- Each user's permissions work as defined in rbac-permissions.md
- Appropriate access granted/denied

**Test 9: Login verification (basic)**
```bash
obstool users create
# Wait 60 seconds
oc login --username=user1 --password=password
oc get pods -n openshift-monitoring  # Should work (view permission)
```
Expected:
- Login succeeds
- Can view pods in namespace
- Cannot create pods (view only)

### Unit Tests (Optional)

**Test: htpasswd generation**
```go
func TestGenerateHtpasswdData(t *testing.T) {
    data, err := GenerateHtpasswdData(3, "testpass")
    // Verify: 3 lines, format "userN:$2y$..."
    // Verify: bcrypt hash validates
}
```

**Test: bcrypt hash validation**
```go
func TestGenerateBcryptHash(t *testing.T) {
    hash, err := generateBcryptHash("password")
    // Verify: starts with "$2y$" or "$2a$"
    // Verify: bcrypt.CompareHashAndPassword succeeds
}
```

---

## Dependencies

### Go Modules (New)

```bash
go get golang.org/x/crypto/bcrypt
```

### Go Modules (Existing)
- `github.com/openshift/api/config/v1` ✅
- `k8s.io/api/core/v1` ✅
- `k8s.io/api/rbac/v1` ✅
- `sigs.k8s.io/controller-runtime/pkg/client` ✅

### Blocked By
- Nothing! All foundation work complete.

---

## Implementation Checklist

### Phase 1: Utilities & Constants
- [ ] Create `internal/constants/users.go`
- [ ] Create `pkg/users/htpasswd.go`
  - [ ] Implement `GenerateHtpasswdData()`
  - [ ] Implement `generateBcryptHash()`
  - [ ] Test hash format manually
  - [ ] Ensure no htpasswd file written to disk (in-memory only)
- [ ] Create `pkg/users/oauth.go`
  - [ ] Implement `EnsureHTPasswdSecret()` - returns (bool, error)
  - [ ] Implement `EnsureOAuthHTPasswdProvider()` - returns (bool, error)
- [ ] Create `pkg/users/rbac.go`
  - [ ] Implement `ApplyUserRBAC()` - varied permissions per user
  - [ ] Implement `createRoleBinding()` helper
  - [ ] Implement `createClusterRoleBinding()` helper
  - [ ] Handle 9 namespace + 13 cluster = 22 bindings
  - [ ] Reference [rbac-permissions.md](./rbac-permissions.md) for permission matrix

### Phase 2: Business Logic
- [ ] Add `EnsureNamespace()` to `pkg/k8s/` (returns (bool, error))
- [ ] Create `pkg/operations/users.go`
  - [ ] Define step constants (5 steps including EnsureNamespace)
  - [ ] Implement `CreateUsers()` with executor pattern
  - [ ] Default empty password to "password"
  - [ ] Test progress updates

### Phase 3: Command Implementation
- [ ] Create `cmd/users/users.go` (command group)
- [ ] Create `cmd/users/create.go`
  - [ ] Define flags (count, password with "password" default, namespace)
  - [ ] Implement `runUsersCreate()` with mode detection
  - [ ] Implement `runUsersCreateCLI()` - default empty password to "password"
  - [ ] Implement `runUsersCreateTUI()` - 5 operation steps
  - [ ] Implement `collectUsersInput()` form - no password validation
  - [ ] Remove namespace validation (created in business logic)
- [ ] Update `cmd/root.go` to register users command
- [ ] Update `pkg/k8s/scheme.go` to register rbacv1

### Phase 4: Testing & Documentation
- [ ] Manual test: CLI mode with defaults
- [ ] Manual test: CLI mode with custom values
- [ ] Manual test: TUI mode
- [ ] Manual test: Idempotency
- [ ] Manual test: Count validation (< 6 should error)
- [ ] Manual test: Empty password defaults to "password"
- [ ] Manual test: Non-existent namespace gets created
- [ ] Manual test: Login as each user (user1-user6)
- [ ] Manual test: RBAC verification per user (see Test 8 above)
- [ ] Update TASKS.md: Mark users create as complete
- [ ] Verify help text: `obstool users create --help`
- [ ] Document implementation in `tasks/users-create/implementation.md`

---

## Success Criteria

### Functional ✅
- [ ] Creates N users (minimum 6) with htpasswd auth
- [ ] Secret created/updated in openshift-config
- [ ] OAuth CR patched with htpasswd provider
- [ ] RBAC applied to first 6 users with varied permissions (per rbac-permissions.md)
- [ ] Namespace created if doesn't exist
- [ ] Works in both CLI and TUI modes
- [ ] Idempotent (safe to run multiple times)
- [ ] Validates count >= 6
- [ ] Empty password defaults to "password"
- [ ] No htpasswd file written to disk (in-memory only)

### Quality ✅
- [ ] Follows executor pattern for progress tracking
- [ ] Proper error handling with descriptive messages
- [ ] No 1-2 letter variable names (except err, ctx, ok)
- [ ] Minimal comments (code self-documenting)
- [ ] Consistent with existing command patterns

### Documentation ✅
- [ ] Command help text complete and clear
- [ ] Examples in `--help` output
- [ ] Long description explains RBAC behavior

---

## Changes from Original Plan

**User feedback incorporated:**
1. ✅ Password validation: Empty → defaults to "password" (not an error)
2. ✅ Namespace handling: Create if not found using `EnsureNamespace()` pattern
3. ✅ RBAC permissions: Varied per user (see rbac-permissions.md) not uniform
4. ✅ No htpasswd file on disk: In-memory generation only (or write to /tmp, gitignored)
5. ✅ ClusterRole existence: Can create RoleBinding even if ClusterRole doesn't exist yet
6. ✅ Namespace flag: Still useful as target namespace for RBAC resources

## Requirements Summary

**Confirmed by user:**
1. ✅ Username format: `userX` (user1, user2, etc.)
2. ✅ RBAC: First 6 users get varied permissions, minimum count 6
3. ✅ No wait after OAuth creation
4. ✅ No browser opening
5. ✅ Same password for all users (testing only)

---

## Timeline Estimate

**Total**: ~4-6 hours development + 1-2 hours testing

- Phase 1 (Utilities): 2-3 hours
- Phase 2 (Business Logic): 1 hour  
- Phase 3 (Command): 1-2 hours
- Phase 4 (Testing): 1-2 hours

**Note**: Timeline not a constraint - focus on quality and correctness.

---

## Post-Implementation

**After completion:**
1. Update TASKS.md: `[x] Implement users create command`
2. Create `tasks/users-create/implementation.md` documenting actual implementation
3. Consider follow-up: `obstool users rbac` for custom RBAC scenarios
4. Consider follow-up: `obstool users delete` to remove htpasswd users

---

**Ready for Implementation**: YES ✅  
**Approval Required**: User to review this plan before proceeding
