package k8s

import (
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	configv1 "github.com/openshift/api/config/v1"
	consolev1 "github.com/openshift/api/console/v1"
	routev1 "github.com/openshift/api/route/v1"
	operatorsv1 "github.com/operator-framework/api/pkg/operators/v1"
	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"

	lokiv1 "github.com/grafana/loki/operator/api/loki/v1"
	tempov1alpha1 "github.com/grafana/tempo-operator/api/tempo/v1alpha1"
	otelv1beta1 "github.com/open-telemetry/opentelemetry-operator/apis/v1beta1"
	uipluginv1alpha1 "github.com/rhobs/observability-operator/pkg/apis/uiplugin/v1alpha1"
	persesv1alpha2 "github.com/rhobs/perses-operator/api/v1alpha2"
)

func registerSchemes(scheme *runtime.Scheme) error {
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return fmt.Errorf("failed to register core kubernetes scheme: %w", err)
	}
	if err := configv1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("failed to register openshift config/v1 scheme: %w", err)
	}
	if err := routev1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("failed to register openshift route/v1 scheme: %w", err)
	}
	if err := consolev1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("failed to register openshift console/v1 scheme: %w", err)
	}
	if err := operatorsv1alpha1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("failed to register olm operators/v1alpha1 scheme: %w", err)
	}
	if err := operatorsv1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("failed to register olm operators/v1 scheme: %w", err)
	}
	if err := rbacv1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("failed to register rbac/v1 scheme: %w", err)
	}
	if err := lokiv1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("failed to register loki/v1 scheme: %w", err)
	}
	if err := tempov1alpha1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("failed to register tempo/v1alpha1 scheme: %w", err)
	}
	if err := otelv1beta1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("failed to register opentelemetry/v1beta1 scheme: %w", err)
	}
	if err := uipluginv1alpha1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("failed to register uiplugin/v1alpha1 scheme: %w", err)
	}
	if err := persesv1alpha2.AddToScheme(scheme); err != nil {
		return fmt.Errorf("failed to register perses-operator/v1alpha2 scheme: %w", err)
	}
	return nil
}
