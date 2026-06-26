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

var troubleshootingPanelCmd = &cobra.Command{
	Use:   "troubleshooting-panel",
	Short: "Deploy Troubleshooting Panel (Network Observability + UIPlugin)",
	Long: `Deploy the Troubleshooting Panel precondition stack via the Cluster Observability Operator.

Always installs:
  - Network Observability operator (netobserv-operator) in openshift-netobserv-operator namespace

With --deploy-flowcollector (also deploys):
  - MinIO in dedicated 'minio' namespace as shared S3 backend
  - LokiStack 'loki' in netobserv namespace (tenant mode: openshift-network)
  - FlowCollector 'cluster' with eBPF agent pointing at LokiStack
  - Waits for netobserv-plugin deployment to be ready

With --deploy-uiplugin:
  - TroubleshootingPanel UIPlugin CR (name: troubleshooting-panel)

Prerequisites:
  - Cluster Observability Operator must be deployed (obstool deploy coo)
  - Loki Operator must be installed (obstool deploy logging installs it)`,
	Example: `  # Deploy interactively (TUI)
  obstool deploy troubleshooting-panel

  # Install operator only (no flow collection yet)
  obstool deploy troubleshooting-panel --deploy-flowcollector=false --deploy-uiplugin=false

  # Full deployment non-interactively
  obstool deploy troubleshooting-panel --deploy-flowcollector --deploy-uiplugin

  # Deploy with specific netobserv channel
  obstool deploy troubleshooting-panel --netobserv-channel=stable --deploy-flowcollector --deploy-uiplugin`,
	RunE: runDeployTroubleshootingPanel,
}

func init() {
	troubleshootingPanelCmd.Flags().String("netobserv-channel", constants.NetObservDefaultChannel,
		"Network Observability operator subscription channel")
	troubleshootingPanelCmd.Flags().String("storage-class", "",
		"Storage class for MinIO PVC (auto-detect if empty)")
	troubleshootingPanelCmd.Flags().Bool("deploy-flowcollector", false,
		"Deploy MinIO, LokiStack 'loki', and FlowCollector cluster")
	troubleshootingPanelCmd.Flags().Bool("deploy-uiplugin", false,
		"Deploy the TroubleshootingPanel UIPlugin")

	DeployCmd.AddCommand(troubleshootingPanelCmd)
}

func runDeployTroubleshootingPanel(cmd *cobra.Command, args []string) error {
	requiredFlags := []string{}
	useTUI, err := mode.DetermineMode(cmd, requiredFlags)
	if err != nil {
		return err
	}

	if useTUI {
		return runDeployTroubleshootingPanelTUI(cmd)
	}
	return runDeployTroubleshootingPanelCLI(cmd)
}

func runDeployTroubleshootingPanelCLI(cmd *cobra.Command) error {
	config := troubleshootingPanelConfigFromFlags(cmd)

	ctx := cmd.Context()
	ctx = execctx.WithTUI(ctx, false)

	kubeClient, err := execctx.GetClient(ctx)
	if err != nil {
		return err
	}

	handler := output.NewCLIHandler()
	exec := executor.NewExecutor()

	go operations.DeployTroubleshootingPanel(ctx, kubeClient, config, exec)

	for update := range exec.UpdateCh {
		if err := handler.HandleUpdate(update); err != nil {
			return err
		}
	}

	return nil
}

func runDeployTroubleshootingPanelTUI(cmd *cobra.Command) error {
	ctx := cmd.Context()
	ctx = execctx.WithTUI(ctx, true)

	kubeClient, err := execctx.GetClient(ctx)
	if err != nil {
		return err
	}

	config, err := collectTroubleshootingPanelInput(cmd)
	if err != nil {
		return err
	}

	operationsList := buildTroubleshootingPanelOperationsList(config)

	model := tui.NewProgressModel("Deploying Troubleshooting Panel Stack", operationsList)
	program := tea.NewProgram(model)

	exec := executor.NewExecutor()

	go operations.DeployTroubleshootingPanel(ctx, kubeClient, config, exec)

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

func collectTroubleshootingPanelInput(cmd *cobra.Command) (operations.DeployTroubleshootingPanelConfig, error) {
	config := troubleshootingPanelConfigFromFlags(cmd)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Network Observability operator subscription channel").
				Value(&config.NetObservChannel).
				Placeholder(constants.NetObservDefaultChannel),
			huh.NewInput().
				Title("Storage class for MinIO PVC (leave empty for auto-detect)").
				Value(&config.StorageClassName).
				Placeholder("gp3-csi"),
			huh.NewConfirm().
				Title("Deploy FlowCollector?").
				Description("Deploys MinIO, LokiStack 'loki' in netobserv, and FlowCollector cluster").
				Value(&config.DeployFlowCollector),
			huh.NewConfirm().
				Title("Deploy TroubleshootingPanel UIPlugin?").
				Description("Creates the UIPlugin CR for the Troubleshooting Panel console plugin").
				Value(&config.DeployUIPlugin),
		),
	)

	if err := form.Run(); err != nil {
		return operations.DeployTroubleshootingPanelConfig{}, err
	}

	if config.NetObservChannel == "" {
		config.NetObservChannel = constants.NetObservDefaultChannel
	}

	return config, nil
}

func troubleshootingPanelConfigFromFlags(cmd *cobra.Command) operations.DeployTroubleshootingPanelConfig {
	netobservChannel, _ := cmd.Flags().GetString("netobserv-channel")
	storageClass, _ := cmd.Flags().GetString("storage-class")
	deployFC, _ := cmd.Flags().GetBool("deploy-flowcollector")
	deployUIPlugin, _ := cmd.Flags().GetBool("deploy-uiplugin")

	if netobservChannel == "" {
		netobservChannel = constants.NetObservDefaultChannel
	}

	return operations.DeployTroubleshootingPanelConfig{
		NetObservChannel:    netobservChannel,
		StorageClassName:    storageClass,
		DeployFlowCollector: deployFC,
		DeployUIPlugin:      deployUIPlugin,
	}
}

func buildTroubleshootingPanelOperationsList(config operations.DeployTroubleshootingPanelConfig) []string {
	operationsList := []string{
		fmt.Sprintf("Create namespace %s", constants.NetObservOperatorNamespace),
		"Create OperatorGroup openshift-netobserv-operator-hack",
		"Create Subscription for netobserv-operator",
		"Wait for netobserv-operator to be ready",
	}

	if config.DeployFlowCollector {
		operationsList = append(operationsList,
			"Deploy MinIO storage",
			fmt.Sprintf("Create namespace %s", constants.NetObservNamespace),
			fmt.Sprintf("Create LokiStack %s", constants.NetObservLokiStackName),
			fmt.Sprintf("Wait for LokiStack %s to be ready", constants.NetObservLokiStackName),
			fmt.Sprintf("Deploy FlowCollector %s", constants.FlowCollectorName),
			"Wait for netobserv-plugin to be ready",
		)
	}

	if config.DeployUIPlugin {
		operationsList = append(operationsList, "Deploy TroubleshootingPanel UIPlugin")
	}

	return operationsList
}
