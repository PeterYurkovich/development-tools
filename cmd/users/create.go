package users

import (
	"fmt"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/observability-ui/development-tools/internal/constants"
	execctx "github.com/observability-ui/development-tools/pkg/context"
	"github.com/observability-ui/development-tools/pkg/executor"
	"github.com/observability-ui/development-tools/pkg/mode"
	"github.com/observability-ui/development-tools/pkg/operations"
	"github.com/observability-ui/development-tools/pkg/output"
	"github.com/observability-ui/development-tools/pkg/tui"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create test users with htpasswd authentication",
	Long: `Creates test users for OpenShift development and testing.

Generates N users (minimum 6) with htpasswd authentication and applies
RBAC permissions to the first 6 users in specified namespaces.

Users are named: user1, user2, user3, ..., userN
All users share the same password.

RBAC permissions (first 6 users):
  user1: Admin-like (cluster-wide read+write)
  user2: Read-only (cluster-wide read-only)
  user3: Multi-namespace editor (perses-dev + openshift-cluster-observability-operator)
  user4: Single-namespace editor (perses-dev)
  user5: Single-namespace viewer (perses-dev)
  user6: Dashboards + Metrics viewer (perses-dev, specific resources only)

Namespaces created automatically:
  - perses-dev
  - openshift-cluster-observability-operator
  - Namespace from --namespace flag (for user1/user2)

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
	createCmd.Flags().String("namespace", constants.DefaultUserNamespace, "Namespace for RBAC permissions (user1/user2)")
	UsersCmd.AddCommand(createCmd)
}

func runUsersCreate(cmd *cobra.Command, args []string) error {
	useTUI, err := mode.DetermineMode(cmd, []string{})
	if err != nil {
		return err
	}

	if useTUI {
		return runUsersCreateTUI(cmd)
	}

	return runUsersCreateCLI(cmd)
}

func runUsersCreateCLI(cmd *cobra.Command) error {
	count, _ := cmd.Flags().GetInt("count")
	password, _ := cmd.Flags().GetString("password")
	namespace, _ := cmd.Flags().GetString("namespace")

	if count < constants.MinUserCount {
		return fmt.Errorf("count must be at least %d, got %d", constants.MinUserCount, count)
	}

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

func runUsersCreateTUI(cmd *cobra.Command) error {
	ctx := cmd.Context()
	ctx = execctx.WithTUI(ctx, true)

	kubeClient, err := execctx.GetClient(ctx)
	if err != nil {
		return err
	}

	count, _ := cmd.Flags().GetInt("count")
	password, _ := cmd.Flags().GetString("password")
	namespace, _ := cmd.Flags().GetString("namespace")

	config, err := collectUsersInput(count, password, namespace)
	if err != nil {
		return err
	}

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

	tuiModel := finalModel.(tui.ProgressModel)
	if tuiModel.Error() != nil {
		return tuiModel.Error()
	}

	return nil
}

func convertStatus(status executor.UpdateStatus) tui.OperationStatus {
	switch status {
	case executor.StatusPending:
		return tui.OperationPending
	case executor.StatusInProgress:
		return tui.OperationInProgress
	case executor.StatusComplete:
		return tui.OperationComplete
	case executor.StatusFailed:
		return tui.OperationFailed
	default:
		return tui.OperationPending
	}
}

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
			huh.NewInput().
				Title("Namespace for RBAC permissions (user1/user2)").
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

	if config.Password == "" {
		config.Password = "password"
	}

	return config, nil
}
