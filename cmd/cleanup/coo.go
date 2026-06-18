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

var cooCmd = &cobra.Command{
	Use:   "coo",
	Short: "Remove Cluster Observability Operator",
	Long: `Remove Cluster Observability Operator and related resources.

Deletes (in order):
  - Subscription
  - ClusterServiceVersion (CSV)
  - CatalogSource (if created by obstool)
  - ImageDigestMirrorSet (if created by obstool)
  - OperatorGroup (optional)
  - Namespace (optional)

This command removes the operator deployment but does not affect
UIPlugins or other observability components.`,
	Example: `  # Cleanup COO operator (keeps namespace and operatorgroup)
  obstool cleanup coo

  # Cleanup COO and delete namespace
  obstool cleanup coo --delete-namespace

  # Cleanup everything including operatorgroup
  obstool cleanup coo --delete-namespace --delete-operatorgroup`,
	RunE: runCleanupCOO,
}

func init() {
	cooCmd.Flags().Bool("delete-namespace", false, "Also delete namespace")
	cooCmd.Flags().Bool("delete-operatorgroup", false, "Also delete OperatorGroup")
	cooCmd.Flags().Bool("confirm", false, "Skip confirmation prompt")

	CleanupCmd.AddCommand(cooCmd)
}

func runCleanupCOO(cmd *cobra.Command, args []string) error {
	requiredFlags := []string{}
	useTUI, err := mode.DetermineMode(cmd, requiredFlags)
	if err != nil {
		return err
	}

	if useTUI {
		return runCleanupCOOTUI(cmd)
	}
	return runCleanupCOOCLI(cmd)
}

func runCleanupCOOCLI(cmd *cobra.Command) error {
	deleteNamespace, _ := cmd.Flags().GetBool("delete-namespace")
	deleteOperatorGroup, _ := cmd.Flags().GetBool("delete-operatorgroup")
	confirm, _ := cmd.Flags().GetBool("confirm")

	if !confirm {
		fmt.Println("This will delete the Cluster Observability Operator.")
		if deleteNamespace {
			fmt.Printf("WARNING: This will also delete the namespace: %s\n", constants.COONamespace)
		}
		if deleteOperatorGroup {
			fmt.Println("This will also delete the OperatorGroup.")
		}
		fmt.Print("Continue? (yes/no): ")
		
		var response string
		fmt.Scanln(&response)
		if response != "yes" && response != "y" {
			fmt.Println("Cleanup cancelled")
			return nil
		}
	}

	config := operations.CleanupCOOConfig{
		DeleteNamespace:     deleteNamespace,
		DeleteOperatorGroup: deleteOperatorGroup,
		Confirm:             true,
	}

	ctx := cmd.Context()
	ctx = execctx.WithTUI(ctx, false)

	kubeClient, err := execctx.GetClient(ctx)
	if err != nil {
		return err
	}

	handler := output.NewCLIHandler()
	exec := executor.NewExecutor()

	go operations.CleanupCOO(ctx, kubeClient, config, exec)

	for update := range exec.UpdateCh {
		if err := handler.HandleUpdate(update); err != nil {
			return err
		}
	}

	return nil
}

func runCleanupCOOTUI(cmd *cobra.Command) error {
	ctx := cmd.Context()
	ctx = execctx.WithTUI(ctx, true)

	kubeClient, err := execctx.GetClient(ctx)
	if err != nil {
		return err
	}

	config, err := collectCOOCleanupInput(cmd)
	if err != nil {
		return err
	}

	operationsList := []string{
		"Delete Subscription",
		"Delete ClusterServiceVersion",
		"Delete CatalogSource",
		"Delete ImageDigestMirrorSets",
	}

	if config.DeleteOperatorGroup {
		operationsList = append(operationsList, "Delete OperatorGroup")
	}

	if config.DeleteNamespace {
		operationsList = append(operationsList, fmt.Sprintf("Delete namespace %s", constants.COONamespace))
	}

	model := tui.NewProgressModel("Cleaning up Cluster Observability Operator", operationsList)
	program := tea.NewProgram(model)

	exec := executor.NewExecutor()

	go operations.CleanupCOO(ctx, kubeClient, config, exec)

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

func collectCOOCleanupInput(cmd *cobra.Command) (operations.CleanupCOOConfig, error) {
	deleteNamespace, _ := cmd.Flags().GetBool("delete-namespace")
	deleteOperatorGroup, _ := cmd.Flags().GetBool("delete-operatorgroup")

	var config operations.CleanupCOOConfig
	config.DeleteNamespace = deleteNamespace
	config.DeleteOperatorGroup = deleteOperatorGroup

	var confirmDelete bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Delete Cluster Observability Operator?").
				Description("This will remove the operator deployment.").
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
