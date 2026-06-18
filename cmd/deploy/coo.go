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

var cooCmd = &cobra.Command{
	Use:   "coo",
	Short: "Deploy Cluster Observability Operator",
	Long: `Deploy Cluster Observability Operator using one of four methods:

  bundle:      Deploy from bundle image using operator-sdk
  fbc:         Deploy from File-Based Catalog (FBC) image
  stage:       Deploy from stage registry FBC
  operatorhub: Deploy from default OperatorHub catalog

Method-specific requirements:
  bundle:      --bundle-url required, --registry-type optional (quay|stage)
  fbc:         --fbc-url required, --registry-type optional (quay|stage)
  stage:       --fbc-url required (uses stage IDMS automatically)
  operatorhub: --channel optional (default: stable)

All methods create namespace and OperatorGroup, then wait for CSV to be ready.`,
	Example: `  # Deploy from bundle
  obstool deploy coo --method=bundle --bundle-url=quay.io/rhobs/coo-bundle:v0.3.6

  # Deploy from FBC with quay registry
  obstool deploy coo --method=fbc --fbc-url=quay.io/rhobs/coo-catalog:latest

  # Deploy from stage registry
  obstool deploy coo --method=stage --fbc-url=registry.stage.redhat.io/...

  # Deploy from OperatorHub
  obstool deploy coo --method=operatorhub --channel=stable

  # Interactive mode (TUI prompts for inputs)
  obstool deploy coo`,
	RunE: runDeployCOO,
}

func init() {
	cooCmd.Flags().String("method", "", "Deployment method: bundle, fbc, stage, operatorhub")
	cooCmd.Flags().String("bundle-url", "", "Bundle image URL (bundle method)")
	cooCmd.Flags().String("fbc-url", "", "FBC image URL (fbc/stage methods)")
	cooCmd.Flags().String("channel", "stable", "Subscription channel (operatorhub method)")
	cooCmd.Flags().String("registry-type", "quay", "Registry type: quay or stage (bundle/fbc methods)")

	DeployCmd.AddCommand(cooCmd)
}

func runDeployCOO(cmd *cobra.Command, args []string) error {
	method, _ := cmd.Flags().GetString("method")

	requiredFlags := getRequiredFlagsForMethod(method)
	useTUI, err := mode.DetermineMode(cmd, requiredFlags)
	if err != nil {
		return err
	}

	if useTUI {
		return runDeployCOOTUI(cmd)
	}
	return runDeployCOOCLI(cmd)
}

func getRequiredFlagsForMethod(method string) []string {
	switch method {
	case "bundle":
		return []string{"method", "bundle-url"}
	case "fbc":
		return []string{"method", "fbc-url"}
	case "stage":
		return []string{"method", "fbc-url"}
	case "operatorhub":
		return []string{"method"}
	default:
		return []string{"method"}
	}
}

func runDeployCOOCLI(cmd *cobra.Command) error {
	config, err := getConfigFromFlags(cmd)
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	ctx = execctx.WithTUI(ctx, false)

	kubeClient, err := execctx.GetClient(ctx)
	if err != nil {
		return err
	}

	handler := output.NewCLIHandler()
	exec := executor.NewExecutor()

	go operations.DeployCOO(ctx, kubeClient, config, exec)

	for update := range exec.UpdateCh {
		if err := handler.HandleUpdate(update); err != nil {
			return err
		}
	}

	return nil
}

func runDeployCOOTUI(cmd *cobra.Command) error {
	ctx := cmd.Context()
	ctx = execctx.WithTUI(ctx, true)

	kubeClient, err := execctx.GetClient(ctx)
	if err != nil {
		return err
	}

	config, err := collectCOOInput(cmd)
	if err != nil {
		return err
	}

	operationsList := []string{
		fmt.Sprintf("Create namespace %s", constants.COONamespace),
		"Create OperatorGroup",
	}

	switch config.Method {
	case "bundle":
		operationsList = append(operationsList,
			"Check for operator-sdk",
			"Create ImageDigestMirrorSet",
			"Detect OCP version",
			"Run operator-sdk bundle",
		)
	case "fbc":
		operationsList = append(operationsList,
			"Create ImageDigestMirrorSet for Quay",
			"Create CatalogSource",
			"Create Subscription",
		)
	case "stage":
		operationsList = append(operationsList,
			"Create ImageDigestMirrorSet for Stage",
			"Create CatalogSource",
			"Create Subscription",
		)
	case "operatorhub":
		operationsList = append(operationsList,
			"Create Subscription to OperatorHub",
		)
	}

	operationsList = append(operationsList, "Wait for operator to be ready")

	model := tui.NewProgressModel("Deploying Cluster Observability Operator", operationsList)
	program := tea.NewProgram(model)

	exec := executor.NewExecutor()

	go operations.DeployCOO(ctx, kubeClient, config, exec)

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

func getConfigFromFlags(cmd *cobra.Command) (operations.DeployCOOConfig, error) {
	method, _ := cmd.Flags().GetString("method")
	if method == "" {
		return operations.DeployCOOConfig{}, fmt.Errorf("--method is required")
	}

	config := operations.DeployCOOConfig{
		Method:       method,
		Channel:      cmd.Flags().Lookup("channel").Value.String(),
		RegistryType: cmd.Flags().Lookup("registry-type").Value.String(),
	}

	switch method {
	case "bundle":
		bundleURL, _ := cmd.Flags().GetString("bundle-url")
		if bundleURL == "" {
			return config, fmt.Errorf("--bundle-url is required for bundle method")
		}
		config.BundleURL = bundleURL
	case "fbc":
		fbcURL, _ := cmd.Flags().GetString("fbc-url")
		if fbcURL == "" {
			return config, fmt.Errorf("--fbc-url is required for fbc method")
		}
		config.FBCURL = fbcURL
	case "stage":
		fbcURL, _ := cmd.Flags().GetString("fbc-url")
		if fbcURL == "" {
			return config, fmt.Errorf("--fbc-url is required for stage method")
		}
		config.FBCURL = fbcURL
		config.RegistryType = "stage"
	case "operatorhub":
		// No additional required flags
	default:
		return config, fmt.Errorf("unknown method: %s (must be: bundle, fbc, stage, operatorhub)", method)
	}

	return config, nil
}

func collectCOOInput(cmd *cobra.Command) (operations.DeployCOOConfig, error) {
	var config operations.DeployCOOConfig

	methodOptions := []huh.Option[string]{
		huh.NewOption("Bundle (operator-sdk)", "bundle"),
		huh.NewOption("FBC (File-Based Catalog)", "fbc"),
		huh.NewOption("Stage Registry", "stage"),
		huh.NewOption("OperatorHub (default catalog)", "operatorhub"),
	}

	methodForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select deployment method").
				Description("Choose how to deploy Cluster Observability Operator").
				Options(methodOptions...).
				Value(&config.Method),
		),
	)

	if err := methodForm.Run(); err != nil {
		return config, err
	}

	switch config.Method {
	case "bundle":
		bundleForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Bundle image URL").
					Placeholder("quay.io/rhobs/coo-bundle:v0.3.6").
					Value(&config.BundleURL).
					Validate(func(s string) error {
						if s == "" {
							return fmt.Errorf("bundle URL cannot be empty")
						}
						return nil
					}),
				huh.NewSelect[string]().
					Title("Registry type").
					Description("Affects IDMS configuration").
					Options(
						huh.NewOption("Quay", "quay"),
						huh.NewOption("Stage", "stage"),
					).
					Value(&config.RegistryType),
			),
		)
		if err := bundleForm.Run(); err != nil {
			return config, err
		}

	case "fbc":
		fbcForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("FBC image URL").
					Placeholder("quay.io/rhobs/coo-catalog:latest").
					Value(&config.FBCURL).
					Validate(func(s string) error {
						if s == "" {
							return fmt.Errorf("FBC URL cannot be empty")
						}
						return nil
					}),
				huh.NewSelect[string]().
					Title("Registry type").
					Options(
						huh.NewOption("Quay", "quay"),
						huh.NewOption("Stage", "stage"),
					).
					Value(&config.RegistryType),
			),
		)
		if err := fbcForm.Run(); err != nil {
			return config, err
		}

	case "stage":
		stageForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Stage FBC image URL").
					Placeholder("registry.stage.redhat.io/...").
					Value(&config.FBCURL).
					Validate(func(s string) error {
						if s == "" {
							return fmt.Errorf("FBC URL cannot be empty")
						}
						return nil
					}),
			),
		)
		if err := stageForm.Run(); err != nil {
			return config, err
		}
		config.RegistryType = "stage"

	case "operatorhub":
		config.Channel = "stable"
		channelForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Subscription channel").
					Placeholder("stable").
					Value(&config.Channel).
					Validate(func(s string) error {
						if s == "" {
							return fmt.Errorf("channel cannot be empty")
						}
						return nil
					}),
			),
		)
		if err := channelForm.Run(); err != nil {
			return config, err
		}
	}

	return config, nil
}

func convertStatus(status executor.UpdateStatus) tui.OperationStatus {
	switch status {
	case executor.StatusPending:
		return tui.OperationPending
	case executor.StatusInProgress:
		return tui.OperationInProgress
	case executor.StatusComplete:
		return tui.OperationComplete
	case executor.StatusFailed:
		return tui.OperationFailed
	default:
		return tui.OperationPending
	}
}
