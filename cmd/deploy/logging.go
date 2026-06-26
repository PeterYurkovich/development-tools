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

var loggingCmd = &cobra.Command{
	Use:   "logging",
	Short: "Deploy Cluster Logging and Loki Operators",
	Long: `Deploy both Cluster Logging and Loki Operators from OperatorHub.

Installs:
  - cluster-logging operator in openshift-logging namespace
  - loki-operator in openshift-operators-redhat namespace

Optional components (enabled via flags or TUI):
  - MinIO storage for LokiStack backend
  - LokiStack CR (requires MinIO)
  - Collector RBAC + ClusterLogForwarder (requires LokiStack)
  - Logging UIPlugin
  - Chat signal generator app in the 'chat' namespace

Both operators are installed from the redhat-operators catalog source.`,
	Example: `  # Deploy operators only (interactive TUI for optional components)
  obstool deploy logging

  # Deploy with specific channels
  obstool deploy logging --logging-channel=stable --loki-channel=stable-6.1

  # Deploy with chat signal generator
  obstool deploy logging --deploy-signals`,
	RunE: runDeployLogging,
}

func init() {
	loggingCmd.Flags().String("logging-channel", "stable", "Cluster Logging subscription channel")
	loggingCmd.Flags().String("loki-channel", "stable", "Loki Operator subscription channel")
	loggingCmd.Flags().Bool("deploy-signals", false, "Deploy chat log generator app in the 'chat' namespace")

	DeployCmd.AddCommand(loggingCmd)
}

func runDeployLogging(cmd *cobra.Command, args []string) error {
	requiredFlags := []string{}
	useTUI, err := mode.DetermineMode(cmd, requiredFlags)
	if err != nil {
		return err
	}

	if useTUI {
		return runDeployLoggingTUI(cmd)
	}
	return runDeployLoggingCLI(cmd)
}

func runDeployLoggingCLI(cmd *cobra.Command) error {
	loggingChannel, _ := cmd.Flags().GetString("logging-channel")
	lokiChannel, _ := cmd.Flags().GetString("loki-channel")
	deploySignals, _ := cmd.Flags().GetBool("deploy-signals")

	config := operations.DeployLoggingConfig{
		LoggingChannel: loggingChannel,
		LokiChannel:    lokiChannel,
		DeploySignals:  deploySignals,
	}

	ctx := cmd.Context()
	ctx = execctx.WithTUI(ctx, false)

	kubeClient, err := execctx.GetClient(ctx)
	if err != nil {
		return err
	}

	handler := output.NewCLIHandler()
	exec := executor.NewExecutor()

	go operations.DeployLogging(ctx, kubeClient, config, exec)

	for update := range exec.UpdateCh {
		if err := handler.HandleUpdate(update); err != nil {
			return err
		}
	}

	return nil
}

func runDeployLoggingTUI(cmd *cobra.Command) error {
	ctx := cmd.Context()
	ctx = execctx.WithTUI(ctx, true)

	kubeClient, err := execctx.GetClient(ctx)
	if err != nil {
		return err
	}

	config, err := collectLoggingInput(cmd)
	if err != nil {
		return err
	}

	operationsList := buildLoggingOperationsList(config)

	model := tui.NewProgressModel("Deploying Cluster Logging and Loki Operators", operationsList)
	program := tea.NewProgram(model)

	exec := executor.NewExecutor()

	go operations.DeployLogging(ctx, kubeClient, config, exec)

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

func collectLoggingInput(cmd *cobra.Command) (operations.DeployLoggingConfig, error) {
	loggingChannel, _ := cmd.Flags().GetString("logging-channel")
	lokiChannel, _ := cmd.Flags().GetString("loki-channel")
	deploySignals, _ := cmd.Flags().GetBool("deploy-signals")

	if loggingChannel == "" {
		loggingChannel = "stable"
	}
	if lokiChannel == "" {
		lokiChannel = "stable"
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Cluster Logging Subscription Channel").
				Value(&loggingChannel).
				Placeholder("stable"),
			huh.NewInput().
				Title("Loki Operator Subscription Channel").
				Value(&lokiChannel).
				Placeholder("stable"),
			huh.NewConfirm().
				Title("Deploy chat log generator app in the 'chat' namespace?").
				Value(&deploySignals),
		),
	)

	if err := form.Run(); err != nil {
		return operations.DeployLoggingConfig{}, err
	}

	return operations.DeployLoggingConfig{
		LoggingChannel: loggingChannel,
		LokiChannel:    lokiChannel,
		DeploySignals:  deploySignals,
	}, nil
}

func buildLoggingOperationsList(config operations.DeployLoggingConfig) []string {
	operationsList := []string{
		fmt.Sprintf("Create namespace %s", constants.LoggingNamespace),
		"Create OperatorGroup",
		"Create Subscription for cluster-logging",
		fmt.Sprintf("Create namespace %s", constants.LokiNamespace),
		"Create Loki OperatorGroup",
		"Create Subscription for loki-operator",
		"Wait for cluster-logging operator to be ready",
		"Wait for loki-operator to be ready",
	}

	if config.DeployMinIO {
		operationsList = append(operationsList, "Deploy MinIO storage")
	}

	if config.DeployLokiStack {
		operationsList = append(operationsList,
			"Deploy LokiStack",
			"Wait for LokiStack to be ready",
		)
	}

	if config.DeployForwarder {
		operationsList = append(operationsList,
			"Create collector RBAC",
			"Deploy ClusterLogForwarder",
			"Wait for ClusterLogForwarder to be ready",
		)
	}

	if config.DeployUIPlugin {
		operationsList = append(operationsList,
			"Deploy Logging UIPlugin",
			"Wait for UIPlugin to be ready",
		)
	}

	if config.DeploySignals {
		operationsList = append(operationsList, "Deploy chat signal app")
	}

	return operationsList
}
