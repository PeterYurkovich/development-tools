package operators

import (
	"context"
	"fmt"

	configv1 "github.com/openshift/api/config/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type IDMSConfig struct {
	Name    string
	Source  string
	Mirrors []configv1.ImageMirror
}

func CreateImageDigestMirrorSet(ctx context.Context, kubeClient client.Client, config IDMSConfig) error {
	idms := &configv1.ImageDigestMirrorSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: config.Name,
			Labels: map[string]string{
				"obstool.observability.openshift.io/managed": "true",
			},
		},
		Spec: configv1.ImageDigestMirrorSetSpec{
			ImageDigestMirrors: []configv1.ImageDigestMirrors{
				{
					Source:  config.Source,
					Mirrors: config.Mirrors,
				},
			},
		},
	}

	err := kubeClient.Create(ctx, idms)
	if err != nil {
		return fmt.Errorf("failed to create imagedigestmirrorset %s: %w", config.Name, err)
	}

	return nil
}

func GetImageDigestMirrorSet(ctx context.Context, kubeClient client.Client, name string) (*configv1.ImageDigestMirrorSet, error) {
	idms := &configv1.ImageDigestMirrorSet{}
	key := client.ObjectKey{Name: name}

	err := kubeClient.Get(ctx, key, idms)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, fmt.Errorf("imagedigestmirrorset %s not found", name)
		}
		return nil, fmt.Errorf("failed to get imagedigestmirrorset %s: %w", name, err)
	}

	return idms, nil
}

func DeleteImageDigestMirrorSet(ctx context.Context, kubeClient client.Client, name string) error {
	idms := &configv1.ImageDigestMirrorSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	err := kubeClient.Delete(ctx, idms)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete imagedigestmirrorset %s: %w", name, err)
	}

	return nil
}

func ListImageDigestMirrorSets(ctx context.Context, kubeClient client.Client) (*configv1.ImageDigestMirrorSetList, error) {
	idmsList := &configv1.ImageDigestMirrorSetList{}

	err := kubeClient.List(ctx, idmsList)
	if err != nil {
		return nil, fmt.Errorf("failed to list imagedigestmirrorsets: %w", err)
	}

	return idmsList, nil
}

func ImageDigestMirrorSetExists(ctx context.Context, kubeClient client.Client, name string) (bool, error) {
	_, err := GetImageDigestMirrorSet(ctx, kubeClient, name)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func EnsureIDMSQuay(ctx context.Context, kubeClient client.Client) error {
	exists, err := ImageDigestMirrorSetExists(ctx, kubeClient, "idms-coo-quay")
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	return CreateImageDigestMirrorSet(ctx, kubeClient, IDMSConfig{
		Name:   "idms-coo-quay",
		Source: "registry.redhat.io/cluster-observability-operator",
		Mirrors: []configv1.ImageMirror{
			"quay.io/redhat-user-workloads/cluster-observabilit-tenant/cluster-observability-operator",
		},
	})
}

func EnsureIDMSStage(ctx context.Context, kubeClient client.Client) error {
	exists, err := ImageDigestMirrorSetExists(ctx, kubeClient, "idms-coo-stage")
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	return CreateImageDigestMirrorSet(ctx, kubeClient, IDMSConfig{
		Name:   "idms-coo-stage",
		Source: "registry.redhat.io",
		Mirrors: []configv1.ImageMirror{
			"registry.stage.redhat.io",
		},
	})
}

func EnsureIDMSStageWithBrew(ctx context.Context, kubeClient client.Client) error {
	exists, err := ImageDigestMirrorSetExists(ctx, kubeClient, "idms-coo-stage")
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	idms := &configv1.ImageDigestMirrorSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: "idms-coo-stage",
			Labels: map[string]string{
				"obstool.observability.openshift.io/managed": "true",
			},
		},
		Spec: configv1.ImageDigestMirrorSetSpec{
			ImageDigestMirrors: []configv1.ImageDigestMirrors{
				{
					Source: "registry.redhat.io",
					Mirrors: []configv1.ImageMirror{
						"registry.stage.redhat.io",
					},
				},
				{
					Source: "registry-proxy.engineering.redhat.com/rh-osbs/iib",
					Mirrors: []configv1.ImageMirror{
						"brew.registry.redhat.io/rh-osbs/iib",
					},
				},
			},
		},
	}

	err = kubeClient.Create(ctx, idms)
	if err != nil {
		return fmt.Errorf("failed to create imagedigestmirrorset idms-coo-stage: %w", err)
	}

	return nil
}
