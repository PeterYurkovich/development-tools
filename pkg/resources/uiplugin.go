package resources

import (
	"context"
	"fmt"

	uipluginv1alpha1 "github.com/rhobs/observability-operator/pkg/apis/uiplugin/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/observability-ui/development-tools/internal/constants"
)

type LoggingUIPluginConfig struct {
	Name          string
	LokiStackName string
	LogsLimit     int32
	Timeout       string
	Schema        string
}

func CreateLoggingUIPlugin(ctx context.Context, kubeClient client.Client, config LoggingUIPluginConfig) error {
	plugin := &uipluginv1alpha1.UIPlugin{
		ObjectMeta: metav1.ObjectMeta{
			Name: config.Name,
		},
		Spec: uipluginv1alpha1.UIPluginSpec{
			Type: uipluginv1alpha1.TypeLogging,
			Logging: &uipluginv1alpha1.LoggingConfig{
				LokiStack: &uipluginv1alpha1.LokiStackReference{
					Name: config.LokiStackName,
				},
				LogsLimit:            config.LogsLimit,
				Timeout:              config.Timeout,
				Schema:               config.Schema,
				ShowTimezoneSelector: true,
			},
		},
	}

	err := kubeClient.Create(ctx, plugin)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create Logging UIPlugin: %w", err)
	}
	return nil
}

func CreateTracingUIPlugin(ctx context.Context, kubeClient client.Client) error {
	plugin := &uipluginv1alpha1.UIPlugin{
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.TracingUIPluginName,
		},
		Spec: uipluginv1alpha1.UIPluginSpec{
			Type: uipluginv1alpha1.TypeDistributedTracing,
		},
	}

	err := kubeClient.Create(ctx, plugin)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create Tracing UIPlugin: %w", err)
	}
	return nil
}

type MonitoringUIPluginConfig struct {
	EnablePerses                bool
	EnableACM                   bool
	EnableClusterHealthAnalyzer bool
}

func CreateMonitoringUIPlugin(ctx context.Context, kubeClient client.Client, config MonitoringUIPluginConfig) error {
	spec := uipluginv1alpha1.UIPluginSpec{
		Type: uipluginv1alpha1.TypeMonitoring,
	}

	if config.EnablePerses || config.EnableACM || config.EnableClusterHealthAnalyzer {
		spec.Monitoring = &uipluginv1alpha1.MonitoringConfig{}

		if config.EnablePerses {
			spec.Monitoring.Perses = &uipluginv1alpha1.PersesReference{Enabled: true}
		}
		if config.EnableACM {
			spec.Monitoring.ACM = &uipluginv1alpha1.AdvancedClusterManagementReference{
				Enabled: true,
				Alertmanager:  uipluginv1alpha1.AlertmanagerReference{Url: constants.ACMAlertmanagerURL},
				ThanosQuerier: uipluginv1alpha1.ThanosQuerierReference{Url: constants.ACMThanosQuerierURL},
			}
		}
		if config.EnableClusterHealthAnalyzer {
			spec.Monitoring.ClusterHealthAnalyzer = &uipluginv1alpha1.ClusterHealthAnalyzerReference{Enabled: true}
		}
	}

	plugin := &uipluginv1alpha1.UIPlugin{
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.MonitoringUIPluginName,
		},
		Spec: spec,
	}

	err := kubeClient.Create(ctx, plugin)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create Monitoring UIPlugin: %w", err)
	}
	return nil
}

func CreateTroubleshootingPanelUIPlugin(ctx context.Context, kubeClient client.Client) error {
	plugin := &uipluginv1alpha1.UIPlugin{
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.TroubleshootingPanelUIPluginName,
		},
		Spec: uipluginv1alpha1.UIPluginSpec{
			Type: uipluginv1alpha1.TypeTroubleshootingPanel,
		},
	}

	err := kubeClient.Create(ctx, plugin)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create TroubleshootingPanel UIPlugin: %w", err)
	}
	return nil
}

func GetUIPlugin(ctx context.Context, kubeClient client.Client, name string) (*uipluginv1alpha1.UIPlugin, error) {
	plugin := &uipluginv1alpha1.UIPlugin{}
	key := client.ObjectKey{Name: name}

	err := kubeClient.Get(ctx, key, plugin)
	if err != nil {
		return nil, err
	}

	return plugin, nil
}
