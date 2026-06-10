package k8s

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	configv1 "github.com/openshift/api/config/v1"
	consolev1 "github.com/openshift/api/console/v1"
	routev1 "github.com/openshift/api/route/v1"
	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
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
	return nil
}
