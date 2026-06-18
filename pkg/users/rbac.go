package users

import (
	"context"
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/observability-ui/development-tools/internal/constants"
	"github.com/observability-ui/development-tools/pkg/executor"
)

func ApplyUserRBAC(ctx context.Context, kubeClient client.Client, namespace string, exec *executor.Executor) error {
	if err := ensureUser6CustomRole(ctx, kubeClient); err != nil {
		return fmt.Errorf("failed to create custom role for user6: %w", err)
	}

	if err := applyUser1RBAC(ctx, kubeClient, namespace); err != nil {
		return fmt.Errorf("failed to apply user1 RBAC: %w", err)
	}
	exec.SendLog(4, "Applied user1 RBAC (cluster-admin + admin)")

	if err := applyUser2RBAC(ctx, kubeClient, namespace); err != nil {
		return fmt.Errorf("failed to apply user2 RBAC: %w", err)
	}
	exec.SendLog(4, "Applied user2 RBAC (cluster views + view)")

	if err := applyUser3RBAC(ctx, kubeClient); err != nil {
		return fmt.Errorf("failed to apply user3 RBAC: %w", err)
	}
	exec.SendLog(4, "Applied user3 RBAC (edit in 2 namespaces)")

	if err := applyUser4RBAC(ctx, kubeClient); err != nil {
		return fmt.Errorf("failed to apply user4 RBAC: %w", err)
	}
	exec.SendLog(4, "Applied user4 RBAC (edit in perses-dev)")

	if err := applyUser5RBAC(ctx, kubeClient); err != nil {
		return fmt.Errorf("failed to apply user5 RBAC: %w", err)
	}
	exec.SendLog(4, "Applied user5 RBAC (view in perses-dev)")

	if err := applyUser6RBAC(ctx, kubeClient); err != nil {
		return fmt.Errorf("failed to apply user6 RBAC: %w", err)
	}
	exec.SendLog(4, "Applied user6 RBAC (dashboards+metrics view)")

	return nil
}

func applyUser1RBAC(ctx context.Context, kubeClient client.Client, namespace string) error {
	if err := createClusterRoleBinding(ctx, kubeClient, "user1-cluster-admin", "cluster-admin", "user1"); err != nil {
		return err
	}
	if err := createClusterRoleBinding(ctx, kubeClient, "user1-cluster-monitoring-view", "cluster-monitoring-view", "user1"); err != nil {
		return err
	}
	if err := createClusterRoleBinding(ctx, kubeClient, "user1-cluster-logging-application-view", "cluster-logging-application-view", "user1"); err != nil {
		return err
	}
	if err := createClusterRoleBinding(ctx, kubeClient, "user1-distributed-tracing-view", "distributed-tracing-view", "user1"); err != nil {
		return err
	}
	if err := createRoleBinding(ctx, kubeClient, namespace, "user1-admin", "admin", "user1"); err != nil {
		return err
	}
	return nil
}

func applyUser2RBAC(ctx context.Context, kubeClient client.Client, namespace string) error {
	if err := createClusterRoleBinding(ctx, kubeClient, "user2-cluster-monitoring-view", "cluster-monitoring-view", "user2"); err != nil {
		return err
	}
	if err := createClusterRoleBinding(ctx, kubeClient, "user2-cluster-logging-application-view", "cluster-logging-application-view", "user2"); err != nil {
		return err
	}
	if err := createClusterRoleBinding(ctx, kubeClient, "user2-distributed-tracing-view", "distributed-tracing-view", "user2"); err != nil {
		return err
	}
	if err := createRoleBinding(ctx, kubeClient, namespace, "user2-view", "view", "user2"); err != nil {
		return err
	}
	return nil
}

func applyUser3RBAC(ctx context.Context, kubeClient client.Client) error {
	if err := createRoleBinding(ctx, kubeClient, constants.PersesDev, "user3-edit-perses-dev", "edit", "user3"); err != nil {
		return err
	}
	if err := createRoleBinding(ctx, kubeClient, constants.ObservabilityOperatorNamespace, "user3-edit-observability-operator", "edit", "user3"); err != nil {
		return err
	}
	return nil
}

func applyUser4RBAC(ctx context.Context, kubeClient client.Client) error {
	if err := createRoleBinding(ctx, kubeClient, constants.PersesDev, "user4-edit-perses-dev", "edit", "user4"); err != nil {
		return err
	}
	return nil
}

func applyUser5RBAC(ctx context.Context, kubeClient client.Client) error {
	if err := createRoleBinding(ctx, kubeClient, constants.PersesDev, "user5-view-perses-dev", "view", "user5"); err != nil {
		return err
	}
	return nil
}

func applyUser6RBAC(ctx context.Context, kubeClient client.Client) error {
	if err := createRoleBinding(ctx, kubeClient, constants.PersesDev, "user6-dashboards-metrics", "perses-dashboards-metrics-viewer", "user6"); err != nil {
		return err
	}
	return nil
}

func ensureUser6CustomRole(ctx context.Context, kubeClient client.Client) error {
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "perses-dashboards-metrics-viewer",
			Namespace: constants.PersesDev,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"perses.dev"},
				Resources: []string{"persesdashboards"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"monitoring.coreos.com"},
				Resources: []string{"servicemonitors", "prometheusrules"},
				Verbs:     []string{"get", "list", "watch"},
			},
		},
	}

	err := kubeClient.Create(ctx, role)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create custom role: %w", err)
	}
	return nil
}

func createRoleBinding(ctx context.Context, kubeClient client.Client, namespace, name, roleName, userName string) error {
	roleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     roleName,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:     "User",
				Name:     userName,
				APIGroup: "rbac.authorization.k8s.io",
			},
		},
	}

	if roleName == "perses-dashboards-metrics-viewer" {
		roleBinding.RoleRef.Kind = "Role"
	}

	err := kubeClient.Create(ctx, roleBinding)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create RoleBinding %s in namespace %s: %w", name, namespace, err)
	}
	return nil
}

func createClusterRoleBinding(ctx context.Context, kubeClient client.Client, name, roleName, userName string) error {
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     roleName,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:     "User",
				Name:     userName,
				APIGroup: "rbac.authorization.k8s.io",
			},
		},
	}

	err := kubeClient.Create(ctx, clusterRoleBinding)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create ClusterRoleBinding %s: %w", name, err)
	}
	return nil
}
