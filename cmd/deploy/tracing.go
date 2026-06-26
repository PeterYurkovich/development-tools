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

Optional components (enabled via flags or TUI):
  - MinIO storage for TempoStack backend
  - TempoStack CR (requires MinIO)
  - OpenTelemetry Collectors (platform + user, requires TempoStack)
  - Signal generators (hotrod, k6-tracing, telemetrygen)
  - Distributed Tracing UIPlugin

Both operators are installed from the redhat-operators catalog source.`,
	Example: `  # Deploy operators only (interactive TUI for optional components)
  obstool deploy tracing

  # Deploy full tracing stack non-interactively
  obstool deploy tracing --deploy-minio --deploy-tempostack --deploy-collectors --deploy-uiplugin

  # Deploy everything including signal generators
  obstool deploy tracing --deploy-minio --deploy-tempostack --deploy-collectors --deploy-signals --deploy-uiplugin

  # Deploy with specific channels
  obstool deploy tracing --tempo-channel=stable --otel-channel=stable`,
	RunE: runDeployTracing,
}

func init() {
	tracingCmd.Flags().String("tempo-channel", "stable", "Tempo Operator subscription channel")
	tracingCmd.Flags().String("otel-channel", "stable", "OpenTelemetry Operator subscription channel")
	tracingCmd.Flags().String("storage-class", "", "Storage class for MinIO (auto-detect if empty)")
	tracingCmd.Flags().Bool("deploy-minio", false, "Deploy MinIO for TempoStack storage")
	tracingCmd.Flags().Bool("deploy-tempostack", false, "Deploy TempoStack CR (requires --deploy-minio)")
	tracingCmd.Flags().Bool("enable-user-workload-monitoring", false, "Enable user workload monitoring via cluster-monitoring-config")
	tracingCmd.Flags().Bool("deploy-collectors", false, "Deploy platform and user OpenTelemetry Collectors (requires --deploy-tempostack)")
	tracingCmd.Flags().Bool("deploy-signals", false, "Deploy hotrod, k6-tracing, and telemetrygen signal generators")
	tracingCmd.Flags().Bool("deploy-uiplugin", false, "Deploy the Distributed Tracing UIPlugin")

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
	config := tracingConfigFromFlags(cmd)

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

	config, err := collectTracingInput(cmd)
	if err != nil {
		return err
	}

	operationsList := buildTracingOperationsList(config)

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

	progressModel := finalModel.(tui.ProgressModel)
	if progressModel.Error() != nil {
		return progressModel.Error()
	}

	return nil
}

func collectTracingInput(cmd *cobra.Command) (operations.DeployTracingConfig, error) {
	config := tracingConfigFromFlags(cmd)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Tempo Operator Subscription Channel").
				Value(&config.TempoChannel).
				Placeholder("stable"),
			huh.NewInput().
				Title("OpenTelemetry Operator Subscription Channel").
				Value(&config.OTelChannel).
				Placeholder("stable"),
			huh.NewInput().
				Title("Storage Class (leave empty for auto-detect)").
				Value(&config.StorageClassName).
				Placeholder("gp3-csi"),
			huh.NewConfirm().
				Title("Deploy MinIO for TempoStack storage?").
				Value(&config.DeployMinIO),
			huh.NewConfirm().
				Title("Deploy TempoStack CR?").
				Value(&config.DeployTempoStack),
			huh.NewConfirm().
				Title("Enable user workload monitoring?").
				Value(&config.EnableUserWorkloadMonitoring),
			huh.NewConfirm().
				Title("Deploy OpenTelemetry Collectors (platform + user)?").
				Value(&config.DeployCollectors),
			huh.NewConfirm().
				Title("Deploy signal generator apps (hotrod, k6-tracing, telemetrygen)?").
				Value(&config.DeploySignals),
			huh.NewConfirm().
				Title("Deploy Distributed Tracing UIPlugin?").
				Value(&config.DeployUIPlugin),
		),
	)

	if err := form.Run(); err != nil {
		return operations.DeployTracingConfig{}, err
	}

	return config, nil
}

func tracingConfigFromFlags(cmd *cobra.Command) operations.DeployTracingConfig {
	tempoChannel, _ := cmd.Flags().GetString("tempo-channel")
	otelChannel, _ := cmd.Flags().GetString("otel-channel")
	storageClass, _ := cmd.Flags().GetString("storage-class")
	deployMinIO, _ := cmd.Flags().GetBool("deploy-minio")
	deployTempoStack, _ := cmd.Flags().GetBool("deploy-tempostack")
	enableUserWorkload, _ := cmd.Flags().GetBool("enable-user-workload-monitoring")
	deployCollectors, _ := cmd.Flags().GetBool("deploy-collectors")
	deploySignals, _ := cmd.Flags().GetBool("deploy-signals")
	deployUIPlugin, _ := cmd.Flags().GetBool("deploy-uiplugin")

	if tempoChannel == "" {
		tempoChannel = "stable"
	}
	if otelChannel == "" {
		otelChannel = "stable"
	}

	return operations.DeployTracingConfig{
		TempoChannel:                tempoChannel,
		OTelChannel:                 otelChannel,
		StorageClassName:            storageClass,
		DeployMinIO:                 deployMinIO,
		DeployTempoStack:            deployTempoStack,
		EnableUserWorkloadMonitoring: enableUserWorkload,
		DeployCollectors:            deployCollectors,
		DeploySignals:               deploySignals,
		DeployUIPlugin:              deployUIPlugin,
	}
}

func buildTracingOperationsList(config operations.DeployTracingConfig) []string {
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

	if config.DeployMinIO {
		operationsList = append(operationsList, "Deploy MinIO storage")
	}

	if config.DeployTempoStack {
		operationsList = append(operationsList,
			"Deploy TempoStack",
			"Wait for TempoStack to be ready",
		)
	}

	if config.EnableUserWorkloadMonitoring {
		operationsList = append(operationsList, "Enable user workload monitoring")
	}

	if config.DeployTempoStack {
		operationsList = append(operationsList, "Create trace reader RBAC")
	}

	if config.DeployCollectors {
		operationsList = append(operationsList,
			"Deploy platform OpenTelemetry Collector",
			"Wait for platform-collector to be ready",
			"Deploy user OpenTelemetry Collector",
			"Wait for user-collector to be ready",
		)
	}

	if config.DeploySignals {
		operationsList = append(operationsList, "Deploy signal generator apps")
	}

	if config.DeployUIPlugin {
		operationsList = append(operationsList, "Deploy Distributed Tracing UIPlugin")
	}

	return operationsList
}
