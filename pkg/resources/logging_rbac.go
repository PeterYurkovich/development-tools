package resources

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CreateServiceAccountWithRBAC creates a ServiceAccount and binds it to the required ClusterRoles
func CreateServiceAccountWithRBAC(ctx context.Context, kubeClient client.Client, name, namespace string, clusterRoles []string) error {
	// Create ServiceAccount
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	err := kubeClient.Create(ctx, sa)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create ServiceAccount %s/%s: %w", namespace, name, err)
	}

	// Create ClusterRoleBindings
	for _, roleName := range clusterRoles {
		bindingName := fmt.Sprintf("z-logging-%s-%s", name, roleName)

		crb := &rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: bindingName,
			},
			RoleRef: rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "ClusterRole",
				Name:     roleName,
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      name,
					Namespace: namespace,
				},
			},
		}

		err := kubeClient.Create(ctx, crb)
		if err != nil && !errors.IsAlreadyExists(err) {
			return fmt.Errorf("failed to create ClusterRoleBinding %s: %w", bindingName, err)
		}
	}

	return nil
}

// CreateLogCollectorRBAC creates the logcollector ServiceAccount and RBAC
func CreateLogCollectorRBAC(ctx context.Context, kubeClient client.Client, namespace string) error {
	clusterRoles := []string{
		"collect-application-logs",
		"collect-infrastructure-logs",
		"logging-collector-logs-writer",
		"lokistack-tenant-logs",
	}

	return CreateServiceAccountWithRBAC(ctx, kubeClient, "logcollector", namespace, clusterRoles)
}

// CreateCollectorRBAC creates the collector ServiceAccount and RBAC
func CreateCollectorRBAC(ctx context.Context, kubeClient client.Client, namespace string) error {
	clusterRoles := []string{
		"logging-collector-logs-writer",
		"collect-application-logs",
		"collect-infrastructure-logs",
		"collect-audit-logs",
	}

	return CreateServiceAccountWithRBAC(ctx, kubeClient, "collector", namespace, clusterRoles)
}
