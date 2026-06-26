package resources

import (
	"context"
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/observability-ui/development-tools/internal/constants"
)

func CreateTracingRBAC(ctx context.Context, kubeClient client.Client, namespace string) error {
	if err := createTracingClusterRoles(ctx, kubeClient, namespace); err != nil {
		return err
	}
	return createTracingClusterRoleBindings(ctx, kubeClient, namespace)
}

func DeleteTracingRBAC(ctx context.Context, kubeClient client.Client) error {
	roleNames := []string{
		constants.TracesReaderPlatformRole,
		constants.TracesReaderUserRole,
		constants.PlatformCollectorRole,
		constants.UserCollectorRole,
		constants.TracesWriterPlatformRole,
		constants.TracesWriterUserRole,
	}

	for _, name := range roleNames {
		role := &rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: name}}
		if err := kubeClient.Delete(ctx, role); err != nil && !errors.IsNotFound(err) {
			return fmt.Errorf("failed to delete ClusterRole %s: %w", name, err)
		}

		binding := &rbacv1.ClusterRoleBinding{ObjectMeta: metav1.ObjectMeta{Name: name}}
		if err := kubeClient.Delete(ctx, binding); err != nil && !errors.IsNotFound(err) {
			return fmt.Errorf("failed to delete ClusterRoleBinding %s: %w", name, err)
		}
	}

	return nil
}

func createTracingClusterRoles(ctx context.Context, kubeClient client.Client, namespace string) error {
	clusterRoles := []*rbacv1.ClusterRole{
		{
			ObjectMeta: metav1.ObjectMeta{Name: constants.TracesReaderPlatformRole},
			Rules: []rbacv1.PolicyRule{{
				APIGroups:     []string{"tempo.grafana.com"},
				Resources:     []string{constants.PlatformTenantName},
				ResourceNames: []string{"traces"},
				Verbs:         []string{"get"},
			}},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: constants.TracesReaderUserRole},
			Rules: []rbacv1.PolicyRule{{
				APIGroups:     []string{"tempo.grafana.com"},
				Resources:     []string{constants.UserTenantName},
				ResourceNames: []string{"traces"},
				Verbs:         []string{"get"},
			}},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: constants.PlatformCollectorRole},
			Rules: []rbacv1.PolicyRule{
				{
					APIGroups: []string{""},
					Resources: []string{"namespaces", "pods"},
					Verbs:     []string{"get", "list", "watch"},
				},
				{
					APIGroups: []string{"apps"},
					Resources: []string{"replicasets"},
					Verbs:     []string{"get", "list", "watch"},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: constants.TracesWriterPlatformRole},
			Rules: []rbacv1.PolicyRule{{
				APIGroups:     []string{"tempo.grafana.com"},
				Resources:     []string{constants.PlatformTenantName},
				ResourceNames: []string{"traces"},
				Verbs:         []string{"create"},
			}},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: constants.UserCollectorRole},
			Rules: []rbacv1.PolicyRule{
				{
					APIGroups: []string{""},
					Resources: []string{"namespaces", "pods"},
					Verbs:     []string{"get", "list", "watch"},
				},
				{
					APIGroups: []string{"apps"},
					Resources: []string{"replicasets"},
					Verbs:     []string{"get", "list", "watch"},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: constants.TracesWriterUserRole},
			Rules: []rbacv1.PolicyRule{{
				APIGroups:     []string{"tempo.grafana.com"},
				Resources:     []string{constants.UserTenantName},
				ResourceNames: []string{"traces"},
				Verbs:         []string{"create"},
			}},
		},
	}

	for _, clusterRole := range clusterRoles {
		if err := kubeClient.Create(ctx, clusterRole); err != nil && !errors.IsAlreadyExists(err) {
			return fmt.Errorf("failed to create ClusterRole %s: %w", clusterRole.Name, err)
		}
	}

	return nil
}

func createTracingClusterRoleBindings(ctx context.Context, kubeClient client.Client, namespace string) error {
	authenticatedGroup := rbacv1.Subject{
		Kind:     "Group",
		APIGroup: "rbac.authorization.k8s.io",
		Name:     "system:authenticated",
	}

	platformCollectorSA := rbacv1.Subject{
		Kind:      "ServiceAccount",
		Name:      constants.PlatformCollectorDeployment,
		Namespace: namespace,
	}

	userCollectorSA := rbacv1.Subject{
		Kind:      "ServiceAccount",
		Name:      constants.UserCollectorDeployment,
		Namespace: namespace,
	}

	bindings := []*rbacv1.ClusterRoleBinding{
		{
			ObjectMeta: metav1.ObjectMeta{Name: constants.TracesReaderPlatformRole},
			RoleRef: rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "ClusterRole",
				Name:     constants.TracesReaderPlatformRole,
			},
			Subjects: []rbacv1.Subject{authenticatedGroup},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: constants.TracesReaderUserRole},
			RoleRef: rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "ClusterRole",
				Name:     constants.TracesReaderUserRole,
			},
			Subjects: []rbacv1.Subject{authenticatedGroup},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: constants.PlatformCollectorRole},
			RoleRef: rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "ClusterRole",
				Name:     constants.PlatformCollectorRole,
			},
			Subjects: []rbacv1.Subject{platformCollectorSA},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: constants.TracesWriterPlatformRole},
			RoleRef: rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "ClusterRole",
				Name:     constants.TracesWriterPlatformRole,
			},
			Subjects: []rbacv1.Subject{platformCollectorSA},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: constants.UserCollectorRole},
			RoleRef: rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "ClusterRole",
				Name:     constants.UserCollectorRole,
			},
			Subjects: []rbacv1.Subject{userCollectorSA},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: constants.TracesWriterUserRole},
			RoleRef: rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "ClusterRole",
				Name:     constants.TracesWriterUserRole,
			},
			Subjects: []rbacv1.Subject{userCollectorSA},
		},
	}

	for _, binding := range bindings {
		if err := kubeClient.Create(ctx, binding); err != nil && !errors.IsAlreadyExists(err) {
			return fmt.Errorf("failed to create ClusterRoleBinding %s: %w", binding.Name, err)
		}
	}

	return nil
}
