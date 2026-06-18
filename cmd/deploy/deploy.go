package deploy

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var DeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy observability components",
	Long: `Deploy observability components to the cluster.

Run without subcommand for interactive component selection.
Run with subcommand to deploy specific component directly.

Available components:
  - coo: Cluster Observability Operator
  - logging: Logging stack (LokiStack + ClusterLogForwarder)
  - tracing: Tracing stack (TempoStack + OTEL)
  - dashboards: Perses dashboards
  - monitoring: Monitoring plugin
  - acm: ACM observability
  - korrel8r: Korrel8r + NetObserv`,
	RunE: runDeploy,
}

func runDeploy(cmd *cobra.Command, args []string) error {
	return runDeployTUI(cmd)
}

func runDeployTUI(cmd *cobra.Command) error {
	var selectedComponents []string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select components to deploy").
				Description("Components will be deployed sequentially in order selected").
				Options(
					huh.NewOption("Cluster Observability Operator", "coo"),
				).
				Value(&selectedComponents),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	if len(selectedComponents) == 0 {
		fmt.Println("No components selected")
		return nil
	}

	for _, component := range selectedComponents {
		switch component {
		case "coo":
			if err := runDeployCOO(cmd, []string{}); err != nil {
				return fmt.Errorf("failed to deploy COO: %w", err)
			}
		}
	}

	return nil
}
