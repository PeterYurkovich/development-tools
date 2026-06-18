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

var tracingCmd = &cobra.Command{
	Use:   "tracing",
	Short: "Deploy Distributed Tracing (Tempo and OpenTelemetry Operators)",
	Long: `Deploy distributed tracing stack with Tempo and OpenTelemetry operators.

Installs:
  - tempo-product operator in openshift-tempo-operator namespace
  - opentelemetry-product operator in openshift-opentelemetry-operator namespace
  - openshift-tracing namespace for TempoStack and collectors

Both operators are installed from the redhat-operators catalog source.`,
	Example: `  # Deploy with default channels
  obstool deploy tracing

  # Deploy with specific channels
  obstool deploy tracing --tempo-channel=stable --otel-channel=stable`,
	RunE: runDeployTracing,
}

func init() {
	tracingCmd.Flags().String("tempo-channel", "stable", "Tempo Operator subscription channel")
	tracingCmd.Flags().String("otel-channel", "stable", "OpenTelemetry Operator subscription channel")
	tracingCmd.Flags().String("storage-class", "", "Storage class for MinIO (auto-detect if empty)")
	tracingCmd.Flags().Bool("deploy-minio", false, "Deploy MinIO for TempoStack storage")
	tracingCmd.Flags().Bool("deploy-tempostack", false, "Deploy TempoStack CR")

	DeployCmd.AddCommand(tracingCmd)
}

func runDeployTracing(cmd *cobra.Command, args []string) error {
	requiredFlags := []string{}
	useTUI, err := mode.DetermineMode(cmd, requiredFlags)
	if err != nil {
		return err
	}

	if useTUI {
		return runDeployTracingTUI(cmd)
	}
	return runDeployTracingCLI(cmd)
}

func runDeployTracingCLI(cmd *cobra.Command) error {
	tempoChannel, _ := cmd.Flags().GetString("tempo-channel")
	otelChannel, _ := cmd.Flags().GetString("otel-channel")
	storageClass, _ := cmd.Flags().GetString("storage-class")
	deployMinIO, _ := cmd.Flags().GetBool("deploy-minio")
	deployTempoStack, _ := cmd.Flags().GetBool("deploy-tempostack")

	config := operations.DeployTracingConfig{
		TempoChannel:     tempoChannel,
		OTelChannel:      otelChannel,
		StorageClassName: storageClass,
		DeployMinIO:      deployMinIO,
		DeployTempoStack: deployTempoStack,
	}

	ctx := cmd.Context()
	ctx = execctx.WithTUI(ctx, false)

	kubeClient, err := execctx.GetClient(ctx)
	if err != nil {
		return err
	}

	handler := output.NewCLIHandler()
	exec := executor.NewExecutor()

	go operations.DeployTracing(ctx, kubeClient, config, exec)

	for update := range exec.UpdateCh {
		if err := handler.HandleUpdate(update); err != nil {
			return err
		}
	}

	return nil
}

func runDeployTracingTUI(cmd *cobra.Command) error {
	ctx := cmd.Context()
	ctx = execctx.WithTUI(ctx, true)

	kubeClient, err := execctx.GetClient(ctx)
	if err != nil {
		return err
	}

	tempoChannel, otelChannel, storageClass, deployMinIO, deployTempoStack, err := collectTracingInput(cmd)
	if err != nil {
		return err
	}

	config := operations.DeployTracingConfig{
		TempoChannel:     tempoChannel,
		OTelChannel:      otelChannel,
		StorageClassName: storageClass,
		DeployMinIO:      deployMinIO,
		DeployTempoStack: deployTempoStack,
	}

	operationsList := []string{
		fmt.Sprintf("Create namespace %s", constants.TracingNamespace),
		fmt.Sprintf("Create namespace %s", constants.TempoOperatorNS),
		"Create Tempo OperatorGroup",
		"Create Subscription for tempo-product",
		fmt.Sprintf("Create namespace %s", constants.OTelOperatorNS),
		"Create OpenTelemetry OperatorGroup",
		"Create Subscription for opentelemetry-product",
		"Wait for tempo-product operator to be ready",
		"Wait for opentelemetry-product operator to be ready",
	}

	if deployMinIO {
		operationsList = append(operationsList, "Deploy MinIO storage")
	}

	if deployTempoStack {
		operationsList = append(operationsList,
			"Deploy TempoStack",
			"Wait for TempoStack to be ready",
		)
	}

	model := tui.NewProgressModel("Deploying Distributed Tracing Stack", operationsList)
	program := tea.NewProgram(model)

	exec := executor.NewExecutor()

	go operations.DeployTracing(ctx, kubeClient, config, exec)

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

func collectTracingInput(cmd *cobra.Command) (string, string, string, bool, bool, error) {
	tempoChannel, _ := cmd.Flags().GetString("tempo-channel")
	otelChannel, _ := cmd.Flags().GetString("otel-channel")
	storageClass, _ := cmd.Flags().GetString("storage-class")
	deployMinIO, _ := cmd.Flags().GetBool("deploy-minio")
	deployTempoStack, _ := cmd.Flags().GetBool("deploy-tempostack")

	// Set defaults if not provided
	if tempoChannel == "" {
		tempoChannel = "stable"
	}
	if otelChannel == "" {
		otelChannel = "stable"
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Tempo Operator Subscription Channel").
				Value(&tempoChannel).
				Placeholder("stable"),
			huh.NewInput().
				Title("OpenTelemetry Operator Subscription Channel").
				Value(&otelChannel).
				Placeholder("stable"),
			huh.NewInput().
				Title("Storage Class (leave empty for auto-detect)").
				Value(&storageClass).
				Placeholder("gp3-csi"),
			huh.NewConfirm().
				Title("Deploy MinIO for storage?").
				Value(&deployMinIO),
			huh.NewConfirm().
				Title("Deploy TempoStack CR?").
				Value(&deployTempoStack),
		),
	)

	if err := form.Run(); err != nil {
		return "", "", "", false, false, err
	}

	return tempoChannel, otelChannel, storageClass, deployMinIO, deployTempoStack, nil
}
