package resources

import (
	"context"
	"fmt"

	otelv1beta1 "github.com/open-telemetry/opentelemetry-operator/apis/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/observability-ui/development-tools/internal/constants"
)

func CreatePlatformCollector(ctx context.Context, kubeClient client.Client, namespace string) error {
	collector := newOTelCollector(constants.PlatformCollectorName, namespace, constants.PlatformTenantName)
	err := kubeClient.Create(ctx, collector)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create platform OTel collector: %w", err)
	}
	return nil
}

func CreateUserCollector(ctx context.Context, kubeClient client.Client, namespace string) error {
	collector := newOTelCollector(constants.UserCollectorName, namespace, constants.UserTenantName)
	err := kubeClient.Create(ctx, collector)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create user OTel collector: %w", err)
	}
	return nil
}

func newOTelCollector(name, namespace, tenantName string) *otelv1beta1.OpenTelemetryCollector {
	gatewayEndpoint := fmt.Sprintf("tempo-platform-gateway.%s.svc.cluster.local:8090", namespace)

	return &otelv1beta1.OpenTelemetryCollector{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: otelv1beta1.OpenTelemetryCollectorSpec{
			Observability: otelv1beta1.ObservabilitySpec{
				Metrics: otelv1beta1.MetricsConfigSpec{
					EnableMetrics: true,
				},
			},
			Config: otelv1beta1.Config{
				Extensions: &otelv1beta1.AnyConfig{Object: map[string]any{
					"bearertokenauth": map[string]any{
						"filename": "/var/run/secrets/kubernetes.io/serviceaccount/token",
					},
				}},
				Receivers: otelv1beta1.AnyConfig{Object: map[string]any{
					"otlp": map[string]any{
						"protocols": map[string]any{
							"grpc": map[string]any{"endpoint": "0.0.0.0:4317"},
							"http": map[string]any{"endpoint": "0.0.0.0:4318"},
						},
					},
					"jaeger": map[string]any{
						"protocols": map[string]any{
							"thrift_compact": map[string]any{"endpoint": "0.0.0.0:6831"},
						},
					},
				}},
				Connectors: &otelv1beta1.AnyConfig{Object: map[string]any{
					"spanmetrics": map[string]any{
						"metrics_flush_interval": "5s",
						"dimensions": []any{
							map[string]any{"name": "k8s.namespace.name"},
						},
					},
				}},
				Processors: &otelv1beta1.AnyConfig{Object: map[string]any{
					"k8sattributes": map[string]any{},
				}},
				Exporters: otelv1beta1.AnyConfig{Object: map[string]any{
					"otlp": map[string]any{
						"endpoint": gatewayEndpoint,
						"tls": map[string]any{
							"ca_file": "/var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt",
						},
						"auth": map[string]any{
							"authenticator": "bearertokenauth",
						},
						"headers": map[string]any{
							"X-Scope-OrgID": tenantName,
						},
					},
					"prometheus": map[string]any{
						"endpoint":            "0.0.0.0:8889",
						"add_metric_suffixes": false,
						"resource_to_telemetry_conversion": map[string]any{
							"enabled": true,
						},
					},
				}},
				Service: otelv1beta1.Service{
					Extensions: []string{"bearertokenauth"},
					Pipelines: map[string]*otelv1beta1.Pipeline{
						"traces": {
							Receivers:  []string{"otlp", "jaeger"},
							Processors: []string{"k8sattributes"},
							Exporters:  []string{"otlp", "spanmetrics"},
						},
						"metrics": {
							Receivers: []string{"spanmetrics"},
							Exporters: []string{"prometheus"},
						},
					},
				},
			},
		},
	}
}
