package operations

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/observability-ui/development-tools/internal/constants"
	"github.com/observability-ui/development-tools/pkg/executor"
	"github.com/observability-ui/development-tools/pkg/k8s"
	"github.com/observability-ui/development-tools/pkg/resources/dashboards"
	"github.com/observability-ui/development-tools/pkg/resources/datasources"
)

const (
	StepDashboardsEnsureNamespace = iota
	StepDashboardsCreateThanosDatasource
	StepDashboardsCreateLokiDatasource
	StepDashboardsCreateTempoDatasource
	StepDashboardsDeployPrometheusOverview
	StepDashboardsDeployThanosCompact
	StepDashboardsDeployNodeExporter
	StepDashboardsDeployACM
	StepDashboardsDeployPersesSample
)

type DeployDashboardsConfig struct {
	Namespace string
}

func DeployDashboards(ctx context.Context, kubeClient client.Client,
	config DeployDashboardsConfig, exec *executor.Executor) error {
	defer exec.Close()

	namespace := config.Namespace
	if namespace == "" {
		namespace = constants.PersesDefaultNamespace
	}

	stepName := fmt.Sprintf("Create namespace %s", namespace)
	exec.SendUpdate(StepDashboardsEnsureNamespace, executor.StatusInProgress, stepName)
	exec.SendLog(StepDashboardsEnsureNamespace, "Ensuring Perses project namespace exists")

	created, err := k8s.EnsureNamespaceWithLabels(ctx, kubeClient, namespace, nil)
	if err != nil {
		exec.SendUpdateWithError(StepDashboardsEnsureNamespace, executor.StatusFailed, stepName, err)
		return err
	}
	if created {
		exec.SendLog(StepDashboardsEnsureNamespace, fmt.Sprintf("Namespace %s created", namespace))
	} else {
		exec.SendLog(StepDashboardsEnsureNamespace, fmt.Sprintf("Namespace %s already exists", namespace))
	}
	exec.SendUpdate(StepDashboardsEnsureNamespace, executor.StatusComplete, stepName)

	stepName = fmt.Sprintf("Create PersesGlobalDatasource: %s", constants.ThanosDatasourceName)
	exec.SendUpdate(StepDashboardsCreateThanosDatasource, executor.StatusInProgress, stepName)
	exec.SendLog(StepDashboardsCreateThanosDatasource, "Creating Thanos Querier datasource")

	created, err = datasources.EnsureThanosDatasource(ctx, kubeClient)
	if err != nil {
		exec.SendUpdateWithError(StepDashboardsCreateThanosDatasource, executor.StatusFailed, stepName, err)
		return err
	}
	if created {
		exec.SendLog(StepDashboardsCreateThanosDatasource, "Thanos datasource created")
	} else {
		exec.SendLog(StepDashboardsCreateThanosDatasource, "Thanos datasource already exists")
	}
	exec.SendUpdate(StepDashboardsCreateThanosDatasource, executor.StatusComplete, stepName)

	stepName = fmt.Sprintf("Create PersesGlobalDatasource: %s", constants.LokiDatasourceName)
	exec.SendUpdate(StepDashboardsCreateLokiDatasource, executor.StatusInProgress, stepName)
	exec.SendLog(StepDashboardsCreateLokiDatasource, "Creating Loki datasource")

	created, err = datasources.EnsureLokiDatasource(ctx, kubeClient)
	if err != nil {
		exec.SendUpdateWithError(StepDashboardsCreateLokiDatasource, executor.StatusFailed, stepName, err)
		return err
	}
	if created {
		exec.SendLog(StepDashboardsCreateLokiDatasource, "Loki datasource created")
	} else {
		exec.SendLog(StepDashboardsCreateLokiDatasource, "Loki datasource already exists")
	}
	exec.SendUpdate(StepDashboardsCreateLokiDatasource, executor.StatusComplete, stepName)

	stepName = fmt.Sprintf("Create PersesGlobalDatasource: %s", constants.TempoDatasourceName)
	exec.SendUpdate(StepDashboardsCreateTempoDatasource, executor.StatusInProgress, stepName)
	exec.SendLog(StepDashboardsCreateTempoDatasource, "Creating Tempo datasource")

	created, err = datasources.EnsureTempoDatasource(ctx, kubeClient)
	if err != nil {
		exec.SendUpdateWithError(StepDashboardsCreateTempoDatasource, executor.StatusFailed, stepName, err)
		return err
	}
	if created {
		exec.SendLog(StepDashboardsCreateTempoDatasource, "Tempo datasource created")
	} else {
		exec.SendLog(StepDashboardsCreateTempoDatasource, "Tempo datasource already exists")
	}
	exec.SendUpdate(StepDashboardsCreateTempoDatasource, executor.StatusComplete, stepName)

	stepName = "Deploy dashboard: Prometheus Overview"
	exec.SendUpdate(StepDashboardsDeployPrometheusOverview, executor.StatusInProgress, stepName)
	exec.SendLog(StepDashboardsDeployPrometheusOverview, "Creating Prometheus Overview PersesDashboard CR")

	if err := dashboards.CreatePrometheusOverviewDashboard(ctx, kubeClient, namespace); err != nil {
		exec.SendUpdateWithError(StepDashboardsDeployPrometheusOverview, executor.StatusFailed, stepName, err)
		return err
	}
	exec.SendUpdate(StepDashboardsDeployPrometheusOverview, executor.StatusComplete, stepName)

	stepName = "Deploy dashboard: Thanos Compact"
	exec.SendUpdate(StepDashboardsDeployThanosCompact, executor.StatusInProgress, stepName)
	exec.SendLog(StepDashboardsDeployThanosCompact, "Creating Thanos Compact PersesDashboard CR")

	if err := dashboards.CreateThanosCompactDashboard(ctx, kubeClient, namespace); err != nil {
		exec.SendUpdateWithError(StepDashboardsDeployThanosCompact, executor.StatusFailed, stepName, err)
		return err
	}
	exec.SendUpdate(StepDashboardsDeployThanosCompact, executor.StatusComplete, stepName)

	stepName = "Deploy dashboard: Node Exporter Full"
	exec.SendUpdate(StepDashboardsDeployNodeExporter, executor.StatusInProgress, stepName)
	exec.SendLog(StepDashboardsDeployNodeExporter, "Creating Node Exporter Full PersesDashboard CR")

	if err := dashboards.CreateNodeExporterDashboard(ctx, kubeClient, namespace); err != nil {
		exec.SendUpdateWithError(StepDashboardsDeployNodeExporter, executor.StatusFailed, stepName, err)
		return err
	}
	exec.SendUpdate(StepDashboardsDeployNodeExporter, executor.StatusComplete, stepName)

	stepName = "Deploy dashboard: ACM"
	exec.SendUpdate(StepDashboardsDeployACM, executor.StatusInProgress, stepName)
	exec.SendLog(StepDashboardsDeployACM, "Creating ACM PersesDashboard CR")

	if err := dashboards.CreateACMDashboard(ctx, kubeClient, namespace); err != nil {
		exec.SendUpdateWithError(StepDashboardsDeployACM, executor.StatusFailed, stepName, err)
		return err
	}
	exec.SendUpdate(StepDashboardsDeployACM, executor.StatusComplete, stepName)

	stepName = "Deploy dashboard: Perses Sample"
	exec.SendUpdate(StepDashboardsDeployPersesSample, executor.StatusInProgress, stepName)
	exec.SendLog(StepDashboardsDeployPersesSample, "Creating Perses Sample PersesDashboard CR")

	if err := dashboards.CreatePersesSampleDashboard(ctx, kubeClient, namespace); err != nil {
		exec.SendUpdateWithError(StepDashboardsDeployPersesSample, executor.StatusFailed, stepName, err)
		return err
	}
	exec.SendUpdate(StepDashboardsDeployPersesSample, executor.StatusComplete, stepName)

	return nil
}
