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

var dashboardsCmd = &cobra.Command{
	Use:   "dashboards",
	Short: "Deploy Perses dashboards",
	Long: `Deploy Perses observability dashboards to the cluster.

Creates the following resources:
  - Namespace for the Perses project (default: perses-dev)
  - PersesGlobalDatasource CRs: thanos-querier, loki, tempo
  - PersesDashboard CRs: Prometheus Overview, Thanos Compact, Node Exporter Full, ACM, Sample

The perses-operator reconciles the Perses project automatically when dashboards
are deployed into the namespace.

Dashboards are defined using the Perses Go SDK (Dashboards as Code).`,
	Example: `  # Deploy dashboards interactively (TUI)
  obstool deploy dashboards

  # Deploy dashboards to default namespace (perses-dev)
  obstool deploy dashboards --namespace=perses-dev

  # Deploy dashboards to a custom namespace
  obstool deploy dashboards --namespace=my-perses-project`,
	RunE: runDeployDashboards,
}

func init() {
	dashboardsCmd.Flags().String("namespace", constants.PersesDefaultNamespace,
		"Namespace for PersesDashboard CRs (serves as the Perses project name)")

	DeployCmd.AddCommand(dashboardsCmd)
}

func runDeployDashboards(cmd *cobra.Command, args []string) error {
	requiredFlags := []string{}
	useTUI, err := mode.DetermineMode(cmd, requiredFlags)
	if err != nil {
		return err
	}

	if useTUI {
		return runDeployDashboardsTUI(cmd)
	}
	return runDeployDashboardsCLI(cmd)
}

func runDeployDashboardsCLI(cmd *cobra.Command) error {
	config := dashboardsConfigFromFlags(cmd)

	ctx := cmd.Context()
	ctx = execctx.WithTUI(ctx, false)

	kubeClient, err := execctx.GetClient(ctx)
	if err != nil {
		return err
	}

	handler := output.NewCLIHandler()
	exec := executor.NewExecutor()

	go operations.DeployDashboards(ctx, kubeClient, config, exec)

	for update := range exec.UpdateCh {
		if err := handler.HandleUpdate(update); err != nil {
			return err
		}
	}

	return nil
}

func runDeployDashboardsTUI(cmd *cobra.Command) error {
	ctx := cmd.Context()
	ctx = execctx.WithTUI(ctx, true)

	kubeClient, err := execctx.GetClient(ctx)
	if err != nil {
		return err
	}

	config, err := collectDashboardsInput(cmd)
	if err != nil {
		return err
	}

	operationsList := buildDashboardsOperationsList(config)

	model := tui.NewProgressModel("Deploying Perses Dashboards", operationsList)
	program := tea.NewProgram(model)

	exec := executor.NewExecutor()

	go operations.DeployDashboards(ctx, kubeClient, config, exec)

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

func collectDashboardsInput(cmd *cobra.Command) (operations.DeployDashboardsConfig, error) {
	config := dashboardsConfigFromFlags(cmd)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Perses project namespace").
				Description("Namespace for PersesDashboard CRs (serves as the Perses project name)").
				Value(&config.Namespace).
				Placeholder(constants.PersesDefaultNamespace),
		),
	)

	if err := form.Run(); err != nil {
		return operations.DeployDashboardsConfig{}, err
	}

	return config, nil
}

func dashboardsConfigFromFlags(cmd *cobra.Command) operations.DeployDashboardsConfig {
	namespace, _ := cmd.Flags().GetString("namespace")
	if namespace == "" {
		namespace = constants.PersesDefaultNamespace
	}

	return operations.DeployDashboardsConfig{
		Namespace: namespace,
	}
}

func buildDashboardsOperationsList(config operations.DeployDashboardsConfig) []string {
	return []string{
		fmt.Sprintf("Create namespace %s", config.Namespace),
		fmt.Sprintf("Create PersesGlobalDatasource: %s", constants.ThanosDatasourceName),
		fmt.Sprintf("Create PersesGlobalDatasource: %s", constants.LokiDatasourceName),
		fmt.Sprintf("Create PersesGlobalDatasource: %s", constants.TempoDatasourceName),
		"Deploy dashboard: Prometheus Overview",
		"Deploy dashboard: Thanos Compact",
		"Deploy dashboard: Node Exporter Full",
		"Deploy dashboard: ACM",
		"Deploy dashboard: Perses Sample",
	}
}
