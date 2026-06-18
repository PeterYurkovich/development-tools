package operations

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/observability-ui/development-tools/internal/constants"
	"github.com/observability-ui/development-tools/pkg/executor"
	"github.com/observability-ui/development-tools/pkg/k8s"
	"github.com/observability-ui/development-tools/pkg/users"
)

const (
	StepGenerateHtpasswd = iota
	StepCreateSecret
	StepPatchOAuth
	StepEnsureNamespaces
	StepApplyRBAC
)

type CreateUsersConfig struct {
	Count     int
	Password  string
	Namespace string
}

func CreateUsers(ctx context.Context, kubeClient client.Client, config CreateUsersConfig, exec *executor.Executor) error {
	defer exec.Close()

	if config.Password == "" {
		config.Password = "password"
	}

	stepName := fmt.Sprintf("Generate htpasswd data for %d users", config.Count)
	exec.SendUpdate(StepGenerateHtpasswd, executor.StatusInProgress, stepName)
	exec.SendLog(StepGenerateHtpasswd, "Hashing password with bcrypt")

	htpasswdData, err := users.GenerateHtpasswdData(config.Count, config.Password)
	if err != nil {
		exec.SendUpdateWithError(StepGenerateHtpasswd, executor.StatusFailed, stepName, err)
		return err
	}
	exec.SendLog(StepGenerateHtpasswd, fmt.Sprintf("Generated credentials for users 1-%d", config.Count))
	exec.SendUpdate(StepGenerateHtpasswd, executor.StatusComplete, stepName)

	stepName = "Create htpass-secret in openshift-config"
	exec.SendUpdate(StepCreateSecret, executor.StatusInProgress, stepName)

	created, err := users.EnsureHTPasswdSecret(ctx, kubeClient, htpasswdData)
	if err != nil {
		exec.SendUpdateWithError(StepCreateSecret, executor.StatusFailed, stepName, err)
		return err
	}
	if created {
		exec.SendLog(StepCreateSecret, "Secret created")
	} else {
		exec.SendLog(StepCreateSecret, "Secret updated (already existed)")
	}
	exec.SendUpdate(StepCreateSecret, executor.StatusComplete, stepName)

	stepName = "Patch OAuth CR with htpasswd provider"
	exec.SendUpdate(StepPatchOAuth, executor.StatusInProgress, stepName)

	created, err = users.EnsureOAuthHTPasswdProvider(ctx, kubeClient)
	if err != nil {
		exec.SendUpdateWithError(StepPatchOAuth, executor.StatusFailed, stepName, err)
		return err
	}
	if created {
		exec.SendLog(StepPatchOAuth, "HTPasswd provider added to OAuth")
	} else {
		exec.SendLog(StepPatchOAuth, "HTPasswd provider already configured")
	}
	exec.SendUpdate(StepPatchOAuth, executor.StatusComplete, stepName)

	stepName = "Ensure required namespaces exist"
	exec.SendUpdate(StepEnsureNamespaces, executor.StatusInProgress, stepName)

	namespaces := []string{
		constants.PersesDev,
		constants.ObservabilityOperatorNamespace,
		config.Namespace,
	}

	for _, namespace := range namespaces {
		created, err := k8s.EnsureNamespaceWithLabels(ctx, kubeClient, namespace, nil)
		if err != nil {
			exec.SendUpdateWithError(StepEnsureNamespaces, executor.StatusFailed, stepName, err)
			return err
		}
		if created {
			exec.SendLog(StepEnsureNamespaces, fmt.Sprintf("Created namespace: %s", namespace))
		} else {
			exec.SendLog(StepEnsureNamespaces, fmt.Sprintf("Namespace already exists: %s", namespace))
		}
	}
	exec.SendUpdate(StepEnsureNamespaces, executor.StatusComplete, stepName)

	stepName = "Apply varied RBAC to 6 users"
	exec.SendUpdate(StepApplyRBAC, executor.StatusInProgress, stepName)
	exec.SendLog(StepApplyRBAC, "Creating role bindings with different permissions per user")

	err = users.ApplyUserRBAC(ctx, kubeClient, config.Namespace, exec)
	if err != nil {
		exec.SendUpdateWithError(StepApplyRBAC, executor.StatusFailed, stepName, err)
		return err
	}
	exec.SendLog(StepApplyRBAC, "Created 15 role bindings (8 namespace + 7 cluster)")
	exec.SendUpdate(StepApplyRBAC, executor.StatusComplete, stepName)

	return nil
}
