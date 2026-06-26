package deploy

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

var acmCmd = &cobra.Command{
	Use:   "acm",
	Short: "Deploy ACM observability stack",
	Long: `Deploy Advanced Cluster Management (ACM) observability to the cluster.

Installs in order:
  1. ACM Hub operator (advanced-cluster-management) in open-cluster-management namespace
  2. MinIO storage with Thanos S3 secret in open-cluster-management-observability namespace
  3. MultiClusterHub CR (multiclusterhub) — waits up to 15 minutes for ready
  4. MultiClusterObservability CR (observability) backed by MinIO

Use --skip-acm-install if ACM Hub operator is already installed.
Use --skip-multclusterhub if MultiClusterHub already exists.`,
	Example: `  # Full deployment (interactive TUI)
  obstool deploy acm

  # Full deployment non-interactively
  obstool deploy acm --acm-channel=release-2.17

  # Skip ACM Hub install (already installed on cluster)
  obstool deploy acm --skip-acm-install

  # Skip both prerequisites (only deploy MinIO + MCO)
  obstool deploy acm --skip-acm-install --skip-multclusterhub`,
	RunE: runDeployACM,
}

func init() {
	acmCmd.Flags().String("acm-channel", constants.ACMDefaultChannel, "ACM subscription channel")
	acmCmd.Flags().String("storage-class", "", "Storage class for MinIO PVC (auto-detect if empty)")
	acmCmd.Flags().Bool("skip-acm-install", false, "Skip ACM Hub operator installation (assume pre-installed)")
	acmCmd.Flags().Bool("skip-multclusterhub", false, "Skip MultiClusterHub CR creation (assume pre-existing)")

	DeployCmd.AddCommand(acmCmd)
}

func runDeployACM(cmd *cobra.Command, args []string) error {
	requiredFlags := []string{}
	useTUI, err := mode.DetermineMode(cmd, requiredFlags)
	if err != nil {
		return err
	}

	if useTUI {
		return runDeployACMTUI(cmd)
	}
	return runDeployACMCLI(cmd)
}

func runDeployACMCLI(cmd *cobra.Command) error {
	config := acmConfigFromFlags(cmd)

	ctx := cmd.Context()
	ctx = execctx.WithTUI(ctx, false)

	kubeClient, err := execctx.GetClient(ctx)
	if err != nil {
		return err
	}

	handler := output.NewCLIHandler()
	exec := executor.NewExecutor()

	go operations.DeployACM(ctx, kubeClient, config, exec)

	for update := range exec.UpdateCh {
		if err := handler.HandleUpdate(update); err != nil {
			return err
		}
	}

	return nil
}

func runDeployACMTUI(cmd *cobra.Command) error {
	ctx := cmd.Context()
	ctx = execctx.WithTUI(ctx, true)

	kubeClient, err := execctx.GetClient(ctx)
	if err != nil {
		return err
	}

	config, err := collectACMInput(cmd)
	if err != nil {
		return err
	}

	operationsList := buildACMOperationsList(config)

	model := tui.NewProgressModel("Deploying ACM Observability Stack", operationsList)
	program := tea.NewProgram(model)

	exec := executor.NewExecutor()

	go operations.DeployACM(ctx, kubeClient, config, exec)

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

	progressModel := finalModel.(tui.ProgressModel)
	if progressModel.Error() != nil {
		return progressModel.Error()
	}

	return nil
}

func collectACMInput(cmd *cobra.Command) (operations.DeployACMConfig, error) {
	config := acmConfigFromFlags(cmd)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("ACM subscription channel").
				Value(&config.ACMChannel).
				Placeholder(constants.ACMDefaultChannel),
			huh.NewInput().
				Title("Storage class for MinIO PVC (leave empty for auto-detect)").
				Value(&config.StorageClassName).
				Placeholder("gp3-csi"),
			huh.NewConfirm().
				Title("Skip ACM Hub operator installation?").
				Description("Enable if advanced-cluster-management is already installed on this cluster").
				Value(&config.SkipACMInstall),
			huh.NewConfirm().
				Title("Skip MultiClusterHub CR creation?").
				Description("Enable if MultiClusterHub already exists on this cluster").
				Value(&config.SkipMultiClusterHub),
		),
	)

	if err := form.Run(); err != nil {
		return operations.DeployACMConfig{}, err
	}

	if config.ACMChannel == "" {
		config.ACMChannel = constants.ACMDefaultChannel
	}

	return config, nil
}

func acmConfigFromFlags(cmd *cobra.Command) operations.DeployACMConfig {
	acmChannel, _ := cmd.Flags().GetString("acm-channel")
	storageClass, _ := cmd.Flags().GetString("storage-class")
	skipACMInstall, _ := cmd.Flags().GetBool("skip-acm-install")
	skipMCH, _ := cmd.Flags().GetBool("skip-multclusterhub")

	if acmChannel == "" {
		acmChannel = constants.ACMDefaultChannel
	}

	return operations.DeployACMConfig{
		ACMChannel:          acmChannel,
		StorageClassName:    storageClass,
		SkipACMInstall:      skipACMInstall,
		SkipMultiClusterHub: skipMCH,
	}
}

func buildACMOperationsList(config operations.DeployACMConfig) []string {
	operationsList := []string{}

	if !config.SkipACMInstall {
		operationsList = append(operationsList,
			fmt.Sprintf("Create namespace %s", constants.ACMNamespace),
			"Create OperatorGroup og-global",
			"Create Subscription for advanced-cluster-management",
			"Wait for advanced-cluster-management operator to be ready",
		)
	}

	operationsList = append(operationsList,
		fmt.Sprintf("Create namespace %s", constants.ACMObservabilityNamespace),
		"Deploy MinIO storage",
	)

	if !config.SkipMultiClusterHub {
		operationsList = append(operationsList,
			"Create MultiClusterHub",
			"Wait for MultiClusterHub to be ready",
		)
	}

	operationsList = append(operationsList, "Create MultiClusterObservability")

	return operationsList
}
