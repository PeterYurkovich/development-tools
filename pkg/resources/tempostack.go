package resources

import (
	"context"
	"fmt"

	tempov1alpha1 "github.com/grafana/tempo-operator/api/tempo/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/observability-ui/development-tools/internal/constants"
)

type TempoStackConfig struct {
	Name            string
	Namespace       string
	StorageSize     string
	SecretName      string
	SourceNamespace string
}

func CreateTempoStack(ctx context.Context, kubeClient client.Client, config TempoStackConfig) error {
	if config.SourceNamespace != "" && config.SourceNamespace != config.Namespace {
		if err := CopySecretToNamespace(ctx, kubeClient, config.SecretName, config.SourceNamespace, config.Namespace); err != nil {
			return err
		}
	}

	storageSize, err := resource.ParseQuantity(config.StorageSize)
	if err != nil {
		return fmt.Errorf("failed to parse storage size %q: %w", config.StorageSize, err)
	}

	prometheusEndpoint := "https://thanos-querier.openshift-monitoring.svc.cluster.local:9092"
	otlpEndpoint := fmt.Sprintf("http://%s.%s:4318", constants.PlatformCollectorName, constants.TracingNamespace)

	tempoStack := &tempov1alpha1.TempoStack{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.Name,
			Namespace: config.Namespace,
		},
		Spec: tempov1alpha1.TempoStackSpec{
			Storage: tempov1alpha1.ObjectStorageSpec{
				Secret: tempov1alpha1.ObjectStorageSecretSpec{
					Name: config.SecretName,
					Type: tempov1alpha1.ObjectStorageSecretS3,
				},
			},
			StorageSize: storageSize,
			Tenants: &tempov1alpha1.TenantsSpec{
				Mode: tempov1alpha1.ModeOpenShift,
				Authentication: []tempov1alpha1.AuthenticationSpec{
					{TenantName: constants.PlatformTenantName, TenantID: constants.PlatformTenantID},
					{TenantName: constants.UserTenantName, TenantID: constants.UserTenantID},
				},
			},
			Observability: tempov1alpha1.ObservabilitySpec{
				Tracing: tempov1alpha1.TracingConfigSpec{
					OTLPHttpEndpoint: otlpEndpoint,
					SamplingFraction: "1",
				},
			},
			Template: tempov1alpha1.TempoTemplateSpec{
				Gateway: tempov1alpha1.TempoGatewaySpec{
					Enabled: true,
				},
				QueryFrontend: tempov1alpha1.TempoQueryFrontendSpec{
					JaegerQuery: tempov1alpha1.JaegerQuerySpec{
						Enabled: true,
						MonitorTab: tempov1alpha1.JaegerQueryMonitor{
							Enabled:            true,
							PrometheusEndpoint: prometheusEndpoint,
						},
					},
				},
			},
		},
	}

	err = kubeClient.Create(ctx, tempoStack)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create TempoStack: %w", err)
	}
	return nil
}

func GetTempoStack(ctx context.Context, kubeClient client.Client, name, namespace string) (*tempov1alpha1.TempoStack, error) {
	tempoStack := &tempov1alpha1.TempoStack{}
	key := client.ObjectKey{Name: name, Namespace: namespace}

	err := kubeClient.Get(ctx, key, tempoStack)
	if err != nil {
		return nil, err
	}

	return tempoStack, nil
}
