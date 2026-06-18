package k8s

import (
	"context"
	"fmt"
	"strings"

	configv1 "github.com/openshift/api/config/v1"
	"golang.org/x/mod/semver"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type VersionInfo struct {
	OpenShiftVersion string
}

func DetectVersion(ctx context.Context, kubeClient client.Client) (*VersionInfo, error) {
	cv := &configv1.ClusterVersion{}
	key := client.ObjectKey{Name: "version"}

	err := kubeClient.Get(ctx, key, cv)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster version: %w", err)
	}

	version := cv.Status.Desired.Version
	if version == "" {
		return nil, fmt.Errorf("cluster version not available")
	}

	return &VersionInfo{
		OpenShiftVersion: version,
	}, nil
}

func (v *VersionInfo) IsOCP419OrNewer() bool {
	return compareVersion(v.OpenShiftVersion, "4.19") >= 0
}

func (v *VersionInfo) IsOCP417To418() bool {
	return compareVersion(v.OpenShiftVersion, "4.17") >= 0 &&
		compareVersion(v.OpenShiftVersion, "4.19") < 0
}

func compareVersion(current, target string) int {
	if !strings.HasPrefix(current, "v") {
		current = "v" + current
	}
	if !strings.HasPrefix(target, "v") {
		target = "v" + target
	}
	return semver.Compare(current, target)
}
