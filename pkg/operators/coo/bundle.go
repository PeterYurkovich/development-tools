package coo

import (
	"context"
	"fmt"
	"os/exec"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/observability-ui/development-tools/internal/constants"
	execpkg "github.com/observability-ui/development-tools/pkg/executor"
	"github.com/observability-ui/development-tools/pkg/k8s"
	"github.com/observability-ui/development-tools/pkg/operators"
)

const (
	StepBundleCheckOperatorSDK = iota
	StepBundleCreateIDMS
	StepBundleDetectVersion
	StepBundleRunCommand
)

func DeployBundle(ctx context.Context, kubeClient client.Client,
	bundleURL string, registryType string, executor *execpkg.Executor, baseStep int) error {

	stepName := "Check for operator-sdk"
	executor.SendUpdate(baseStep+StepBundleCheckOperatorSDK, execpkg.StatusInProgress, stepName)
	executor.SendLog(baseStep+StepBundleCheckOperatorSDK, "Verifying operator-sdk binary is available")

	_, err := exec.LookPath("operator-sdk")
	if err != nil {
		executor.SendUpdateWithError(baseStep+StepBundleCheckOperatorSDK, execpkg.StatusFailed, stepName, err)
		return fmt.Errorf("operator-sdk binary not found in PATH. Install from: https://sdk.operatorframework.io/docs/installation/")
	}

	executor.SendUpdate(baseStep+StepBundleCheckOperatorSDK, execpkg.StatusComplete, stepName)

	stepName = "Create ImageDigestMirrorSet"
	executor.SendUpdate(baseStep+StepBundleCreateIDMS, execpkg.StatusInProgress, stepName)
	executor.SendLog(baseStep+StepBundleCreateIDMS, fmt.Sprintf("Creating IDMS for registry type: %s", registryType))

	err = createIDMS(ctx, kubeClient, registryType)
	if err != nil {
		executor.SendUpdateWithError(baseStep+StepBundleCreateIDMS, execpkg.StatusFailed, stepName, err)
		return fmt.Errorf("failed to create IDMS: %w", err)
	}

	executor.SendUpdate(baseStep+StepBundleCreateIDMS, execpkg.StatusComplete, stepName)

	stepName = "Detect OCP version"
	executor.SendUpdate(baseStep+StepBundleDetectVersion, execpkg.StatusInProgress, stepName)
	executor.SendLog(baseStep+StepBundleDetectVersion, "Detecting cluster version for security context")

	version, err := k8s.DetectVersion(ctx, kubeClient)
	if err != nil {
		executor.SendUpdateWithError(baseStep+StepBundleDetectVersion, execpkg.StatusFailed, stepName, err)
		return fmt.Errorf("failed to detect OCP version: %w", err)
	}

	executor.SendLog(baseStep+StepBundleDetectVersion, fmt.Sprintf("Detected OCP version: %s", version.OpenShiftVersion))
	executor.SendUpdate(baseStep+StepBundleDetectVersion, execpkg.StatusComplete, stepName)

	stepName = "Run operator-sdk bundle"
	executor.SendUpdate(baseStep+StepBundleRunCommand, execpkg.StatusInProgress, stepName)
	executor.SendLog(baseStep+StepBundleRunCommand, fmt.Sprintf("Deploying bundle: %s", bundleURL))

	args := []string{"run", "bundle", bundleURL,
		"--install-mode", "AllNamespaces",
		"--namespace", constants.COONamespace}

	if version.IsOCP419OrNewer() {
		args = append(args, "--security-context-config", "restricted")
		executor.SendLog(baseStep+StepBundleRunCommand, "Using restricted security context (OCP 4.19+)")
	} else {
		executor.SendLog(baseStep+StepBundleRunCommand, "Using default security context (OCP <4.19)")
	}

	cmd := exec.CommandContext(ctx, "operator-sdk", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		executor.SendUpdateWithError(baseStep+StepBundleRunCommand, execpkg.StatusFailed, stepName, err)
		return fmt.Errorf("operator-sdk run bundle failed: %w\nOutput: %s", err, string(output))
	}

	executor.SendLog(baseStep+StepBundleRunCommand, "Bundle deployment initiated successfully")
	executor.SendUpdate(baseStep+StepBundleRunCommand, execpkg.StatusComplete, stepName)
	return nil
}

func createIDMS(ctx context.Context, kubeClient client.Client, registryType string) error {
	switch registryType {
	case "quay":
		return operators.EnsureIDMSQuay(ctx, kubeClient)
	case "stage":
		return operators.EnsureIDMSStage(ctx, kubeClient)
	default:
		return fmt.Errorf("unknown registry type: %s (must be quay or stage)", registryType)
	}
}
