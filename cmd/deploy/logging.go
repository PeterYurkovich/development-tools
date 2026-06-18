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

Both operators are installed from the redhat-operators catalog source.`,
	Example: `  # Deploy with default channels
  obstool deploy logging

  # Deploy with specific channels
  obstool deploy logging --logging-channel=stable --loki-channel=stable-6.1`,
	RunE: runDeployLogging,
}

func init() {
	loggingCmd.Flags().String("logging-channel", "stable", "Cluster Logging subscription channel")
	loggingCmd.Flags().String("loki-channel", "stable", "Loki Operator subscription channel")

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

	config := operations.DeployLoggingConfig{
		LoggingChannel: loggingChannel,
		LokiChannel:    lokiChannel,
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

	loggingChannel, lokiChannel, err := collectLoggingInput(cmd)
	if err != nil {
		return err
	}

	config := operations.DeployLoggingConfig{
		LoggingChannel: loggingChannel,
		LokiChannel:    lokiChannel,
	}

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

	m := finalModel.(tui.ProgressModel)
	if m.Error() != nil {
		return m.Error()
	}

	return nil
}

func collectLoggingInput(cmd *cobra.Command) (string, string, error) {
	loggingChannel, _ := cmd.Flags().GetString("logging-channel")
	lokiChannel, _ := cmd.Flags().GetString("loki-channel")

	if loggingChannel != "" && lokiChannel != "" {
		return loggingChannel, lokiChannel, nil
	}

	// Set defaults if not provided
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
		),
	)

	if err := form.Run(); err != nil {
		return "", "", err
	}

	return loggingChannel, lokiChannel, nil
}

