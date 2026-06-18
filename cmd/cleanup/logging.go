package cleanup

import (
	"fmt"

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

var loggingCmd = &cobra.Command{
	Use:   "logging",
	Short: "Remove logging stack",
	Long: `Remove logging stack components.

Deletes (in order):
  - UIPlugin (logging)
  - ClusterLogForwarder
  - LokiStack
  - MinIO resources (optional, default: true)
  - Collector RBAC
  - Logging operator (optional)
  - Loki operator (optional)
  - Namespaces (optional)

This command removes deployed logging resources. By default, it keeps
the operators installed so you can redeploy the stack quickly.`,
	Example: `  # Cleanup logging stack (keeps operators and namespaces)
  obstool cleanup logging

  # Cleanup logging and operators
  obstool cleanup logging --delete-operators

  # Complete cleanup including namespaces
  obstool cleanup logging --delete-operators --delete-namespaces

  # Keep MinIO for reuse
  obstool cleanup logging --delete-minio=false`,
	RunE: runCleanupLogging,
}

func init() {
	loggingCmd.Flags().Bool("delete-operators", false, "Also delete operators")
	loggingCmd.Flags().Bool("delete-namespaces", false, "Also delete namespaces")
	loggingCmd.Flags().Bool("delete-minio", true, "Delete MinIO resources")
	loggingCmd.Flags().Bool("confirm", false, "Skip confirmation prompt")

	CleanupCmd.AddCommand(loggingCmd)
}

func runCleanupLogging(cmd *cobra.Command, args []string) error {
	requiredFlags := []string{}
	useTUI, err := mode.DetermineMode(cmd, requiredFlags)
	if err != nil {
		return err
	}

	if useTUI {
		return runCleanupLoggingTUI(cmd)
	}
	return runCleanupLoggingCLI(cmd)
}

func runCleanupLoggingCLI(cmd *cobra.Command) error {
	deleteOperators, _ := cmd.Flags().GetBool("delete-operators")
	deleteNamespaces, _ := cmd.Flags().GetBool("delete-namespaces")
	deleteMinio, _ := cmd.Flags().GetBool("delete-minio")
	confirm, _ := cmd.Flags().GetBool("confirm")

	if !confirm {
		fmt.Println("This will delete the logging stack (UIPlugin, ClusterLogForwarder, LokiStack, collector RBAC).")
		if deleteMinio {
			fmt.Println("This will also delete MinIO resources.")
		}
		if deleteOperators {
			fmt.Println("This will also delete the operators (cluster-logging and loki-operator).")
		}
		if deleteNamespaces {
			fmt.Printf("WARNING: This will delete namespaces: %s, %s, minio\n", constants.LoggingNamespace, constants.LokiNamespace)
		}
		fmt.Print("Continue? (yes/no): ")
		
		var response string
		fmt.Scanln(&response)
		if response != "yes" && response != "y" {
			fmt.Println("Cleanup cancelled")
			return nil
		}
	}

	config := operations.CleanupLoggingConfig{
		DeleteOperators:  deleteOperators,
		DeleteNamespaces: deleteNamespaces,
		DeleteMinIO:      deleteMinio,
		Confirm:          true,
	}

	ctx := cmd.Context()
	ctx = execctx.WithTUI(ctx, false)

	kubeClient, err := execctx.GetClient(ctx)
	if err != nil {
		return err
	}

	handler := output.NewCLIHandler()
	exec := executor.NewExecutor()

	go operations.CleanupLogging(ctx, kubeClient, config, exec)

	for update := range exec.UpdateCh {
		if err := handler.HandleUpdate(update); err != nil {
			return err
		}
	}

	return nil
}

func runCleanupLoggingTUI(cmd *cobra.Command) error {
	ctx := cmd.Context()
	ctx = execctx.WithTUI(ctx, true)

	kubeClient, err := execctx.GetClient(ctx)
	if err != nil {
		return err
	}

	config, err := collectLoggingCleanupInput(cmd)
	if err != nil {
		return err
	}

	operationsList := []string{
		"Delete Logging UIPlugin",
		"Delete ClusterLogForwarder",
		"Delete LokiStack",
	}

	if config.DeleteMinIO {
		operationsList = append(operationsList, "Delete MinIO resources")
	}

	operationsList = append(operationsList, "Delete collector RBAC")

	if config.DeleteOperators {
		operationsList = append(operationsList,
			"Delete logging operator Subscription",
			"Delete logging operator CSV",
			"Delete loki operator Subscription",
			"Delete loki operator CSV",
		)
	}

	if config.DeleteNamespaces {
		operationsList = append(operationsList, "Delete namespaces")
	}

	model := tui.NewProgressModel("Cleaning up Logging Stack", operationsList)
	program := tea.NewProgram(model)

	exec := executor.NewExecutor()

	go operations.CleanupLogging(ctx, kubeClient, config, exec)

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

	m := finalModel.(tui.ProgressModel)
	if m.Error() != nil {
		return m.Error()
	}

	return nil
}

func collectLoggingCleanupInput(cmd *cobra.Command) (operations.CleanupLoggingConfig, error) {
	deleteOperators, _ := cmd.Flags().GetBool("delete-operators")
	deleteNamespaces, _ := cmd.Flags().GetBool("delete-namespaces")
	deleteMinio, _ := cmd.Flags().GetBool("delete-minio")

	var config operations.CleanupLoggingConfig
	config.DeleteOperators = deleteOperators
	config.DeleteNamespaces = deleteNamespaces
	config.DeleteMinIO = deleteMinio

	var confirmDelete bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Delete logging stack?").
				Description("This will remove UIPlugin, ClusterLogForwarder, LokiStack, and collector RBAC.").
				Value(&confirmDelete),
		),
	)

	if err := form.Run(); err != nil {
		return config, err
	}

	if !confirmDelete {
		return config, fmt.Errorf("cleanup cancelled by user")
	}

	config.Confirm = true
	return config, nil
}
