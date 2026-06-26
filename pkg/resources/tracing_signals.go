package resources

import (
	"context"
	"fmt"

	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/observability-ui/development-tools/internal/constants"
)

func CreateHotrodApp(ctx context.Context, kubeClient client.Client) error {
	namespace := constants.HotrodNamespace
	userCollectorEndpoint := fmt.Sprintf("http://%s.%s:4318", constants.UserCollectorDeployment, constants.TracingNamespace)

	if err := ensureNamespaceExists(ctx, kubeClient, namespace); err != nil {
		return err
	}

	replicas := int32(1)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hotrod",
			Namespace: namespace,
			Labels:    map[string]string{"app.kubernetes.io/name": "hotrod"},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app.kubernetes.io/name": "hotrod"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app.kubernetes.io/name": "hotrod"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "hotrod",
						Image: constants.HotrodImage,
						Args:  []string{"all", "--otel-exporter=otlp"},
						Env: []corev1.EnvVar{{
							Name:  "OTEL_EXPORTER_OTLP_ENDPOINT",
							Value: userCollectorEndpoint,
						}},
						Ports: []corev1.ContainerPort{
							{Name: "frontend", ContainerPort: 8080},
							{Name: "customer", ContainerPort: 8081},
							{Name: "route", ContainerPort: 8083},
						},
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("100M"),
							},
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("100M"),
							},
						},
					}},
				},
			},
		},
	}

	if err := kubeClient.Create(ctx, deployment); err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create hotrod deployment: %w", err)
	}

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hotrod",
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app.kubernetes.io/name": "hotrod"},
			Ports: []corev1.ServicePort{{
				Name: "frontend",
				Port: 8080,
			}},
		},
	}

	if err := kubeClient.Create(ctx, svc); err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create hotrod service: %w", err)
	}

	route := &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hotrod",
			Namespace: namespace,
		},
		Spec: routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: "hotrod",
			},
		},
	}

	if err := kubeClient.Create(ctx, route); err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create hotrod route: %w", err)
	}

	return nil
}

func CreateK6TracingApp(ctx context.Context, kubeClient client.Client) error {
	namespace := constants.K6TracingNamespace
	userCollectorGRPC := fmt.Sprintf("%s.%s:4317", constants.UserCollectorDeployment, constants.TracingNamespace)

	if err := ensureNamespaceExists(ctx, kubeClient, namespace); err != nil {
		return err
	}

	replicas := int32(1)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "k6-tracing",
			Namespace: namespace,
			Labels:    map[string]string{"app.kubernetes.io/name": "k6-tracing"},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app.kubernetes.io/name": "k6-tracing"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app.kubernetes.io/name": "k6-tracing"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "k6-tracing",
						Image: constants.K6TracingImage,
						Env: []corev1.EnvVar{{
							Name:  "ENDPOINT",
							Value: userCollectorGRPC,
						}},
					}},
				},
			},
		},
	}

	if err := kubeClient.Create(ctx, deployment); err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create k6-tracing deployment: %w", err)
	}

	return nil
}

func CreateTelemetrygenApp(ctx context.Context, kubeClient client.Client) error {
	namespace := constants.TelemetrygenNamespace
	userCollectorGRPC := fmt.Sprintf("%s.%s:4317", constants.UserCollectorDeployment, constants.TracingNamespace)

	if err := ensureNamespaceExists(ctx, kubeClient, namespace); err != nil {
		return err
	}

	replicas := int32(1)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "telemetrygen",
			Namespace: namespace,
			Labels:    map[string]string{"app.kubernetes.io/name": "telemetrygen"},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app.kubernetes.io/name": "telemetrygen"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app.kubernetes.io/name": "telemetrygen"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "telemetrygen1",
							Image: constants.TelemetrygenImage,
							Args: []string{
								"traces",
								fmt.Sprintf("--otlp-endpoint=%s", userCollectorGRPC),
								"--otlp-insecure",
								"--duration=1h",
								"--service=good_service",
								"--rate=3",
								"--child-spans=2",
							},
						},
						{
							Name:  "telemetrygen2",
							Image: constants.TelemetrygenImage,
							Args: []string{
								"traces",
								fmt.Sprintf("--otlp-endpoint=%s", userCollectorGRPC),
								"--otlp-insecure",
								"--duration=1h",
								"--service=faulty_service",
								"--rate=2",
								"--child-spans=1",
								"--status-code=Error",
							},
						},
					},
				},
			},
		},
	}

	if err := kubeClient.Create(ctx, deployment); err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create telemetrygen deployment: %w", err)
	}

	return nil
}

func ensureNamespaceExists(ctx context.Context, kubeClient client.Client, name string) error {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: name},
	}
	if err := kubeClient.Create(ctx, ns); err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create namespace %s: %w", name, err)
	}
	return nil
}
