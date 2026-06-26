package deploy

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	execctx "github.com/observability-ui/development-tools/pkg/context"
	"github.com/observability-ui/development-tools/pkg/executor"
	"github.com/observability-ui/development-tools/pkg/mode"
	"github.com/observability-ui/development-tools/pkg/operations"
	"github.com/observability-ui/development-tools/pkg/output"
	"github.com/observability-ui/development-tools/pkg/tui"
)

var monitoringCmd = &cobra.Command{
	Use:   "monitoring",
	Short: "Deploy Monitoring UIPlugin",
	Long: `Deploy the Monitoring console plugin via the Cluster Observability Operator.

Creates a UIPlugin CR of type Monitoring. COO reconciles it and deploys
the monitoring-console-plugin to the cluster.

Prerequisites:
  - Cluster Observability Operator must be deployed (obstool deploy coo)

Optional integrations:
  - Perses: wires deployed Perses dashboards into the monitoring plugin UI
  - ACM: proxies ACM Alertmanager and Thanos Querier for multi-cluster monitoring
  - Cluster Health Analyzer: enables cluster health analysis features`,
	Example: `  # Deploy with interactive prompts (TUI)
  obstool deploy monitoring

  # Deploy bare monitoring plugin (no optional features)
  obstool deploy monitoring --enable-perses=false --enable-acm=false --enable-cluster-health-analyzer=false

  # Deploy with Perses and Cluster Health Analyzer enabled
  obstool deploy monitoring --enable-perses --enable-cluster-health-analyzer`,
	RunE: runDeployMonitoring,
}

func init() {
	monitoringCmd.Flags().Bool("enable-perses", false, "Enable Perses dashboard integration")
	monitoringCmd.Flags().Bool("enable-acm", false, "Enable ACM observability integration")
	monitoringCmd.Flags().Bool("enable-cluster-health-analyzer", false, "Enable Cluster Health Analyzer")

	DeployCmd.AddCommand(monitoringCmd)
}

func runDeployMonitoring(cmd *cobra.Command, args []string) error {
	requiredFlags := []string{}
	useTUI, err := mode.DetermineMode(cmd, requiredFlags)
	if err != nil {
		return err
	}

	if useTUI {
		return runDeployMonitoringTUI(cmd)
	}
	return runDeployMonitoringCLI(cmd)
}

func runDeployMonitoringCLI(cmd *cobra.Command) error {
	config := monitoringConfigFromFlags(cmd)

	ctx := cmd.Context()
	ctx = execctx.WithTUI(ctx, false)

	kubeClient, err := execctx.GetClient(ctx)
	if err != nil {
		return err
	}

	handler := output.NewCLIHandler()
	exec := executor.NewExecutor()

	go operations.DeployMonitoring(ctx, kubeClient, config, exec)

	for update := range exec.UpdateCh {
		if err := handler.HandleUpdate(update); err != nil {
			return err
		}
	}

	return nil
}

func runDeployMonitoringTUI(cmd *cobra.Command) error {
	ctx := cmd.Context()
	ctx = execctx.WithTUI(ctx, true)

	kubeClient, err := execctx.GetClient(ctx)
	if err != nil {
		return err
	}

	config, err := collectMonitoringInput(cmd)
	if err != nil {
		return err
	}

	operationsList := []string{"Deploy Monitoring UIPlugin"}

	model := tui.NewProgressModel("Deploying Monitoring UIPlugin", operationsList)
	program := tea.NewProgram(model)

	exec := executor.NewExecutor()

	go operations.DeployMonitoring(ctx, kubeClient, config, exec)

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

func collectMonitoringInput(cmd *cobra.Command) (operations.DeployMonitoringConfig, error) {
	config := monitoringConfigFromFlags(cmd)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Enable Perses dashboard integration?").
				Description("Wires deployed Perses dashboards into the monitoring plugin UI").
				Value(&config.EnablePerses),
			huh.NewConfirm().
				Title("Enable ACM observability integration?").
				Description("Proxies ACM Alertmanager and Thanos Querier for multi-cluster monitoring").
				Value(&config.EnableACM),
			huh.NewConfirm().
				Title("Enable Cluster Health Analyzer?").
				Description("Enables cluster health analysis features in the monitoring plugin").
				Value(&config.EnableClusterHealthAnalyzer),
		),
	)

	if err := form.Run(); err != nil {
		return operations.DeployMonitoringConfig{}, err
	}

	return config, nil
}

func monitoringConfigFromFlags(cmd *cobra.Command) operations.DeployMonitoringConfig {
	enablePerses, _ := cmd.Flags().GetBool("enable-perses")
	enableACM, _ := cmd.Flags().GetBool("enable-acm")
	enableCHA, _ := cmd.Flags().GetBool("enable-cluster-health-analyzer")

	return operations.DeployMonitoringConfig{
		EnablePerses:                enablePerses,
		EnableACM:                   enableACM,
		EnableClusterHealthAnalyzer: enableCHA,
	}
}
