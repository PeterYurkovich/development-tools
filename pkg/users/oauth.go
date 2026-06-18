package users

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	configv1 "github.com/openshift/api/config/v1"

	"github.com/observability-ui/development-tools/internal/constants"
)

func EnsureHTPasswdSecret(ctx context.Context, kubeClient client.Client, htpasswdData []byte) (bool, error) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.HTPasswdSecretName,
			Namespace: constants.OpenshiftConfigNS,
		},
		Data: map[string][]byte{
			constants.HTPasswdSecretKey: htpasswdData,
		},
		Type: corev1.SecretTypeOpaque,
	}

	err := kubeClient.Create(ctx, secret)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			err = kubeClient.Update(ctx, secret)
			if err != nil {
				return false, fmt.Errorf("failed to update existing secret: %w", err)
			}
			return false, nil
		}
		return false, fmt.Errorf("failed to create secret: %w", err)
	}

	return true, nil
}

func EnsureOAuthHTPasswdProvider(ctx context.Context, kubeClient client.Client) (bool, error) {
	oauth := &configv1.OAuth{}
	err := kubeClient.Get(ctx, client.ObjectKey{Name: constants.OAuthCRName}, oauth)
	if err != nil {
		return false, fmt.Errorf("failed to get OAuth CR: %w", err)
	}

	for _, provider := range oauth.Spec.IdentityProviders {
		if provider.Name == constants.HTPasswdProviderName {
			return false, nil
		}
	}

	htpasswdProvider := configv1.IdentityProvider{
		Name:          constants.HTPasswdProviderName,
		MappingMethod: "claim",
		IdentityProviderConfig: configv1.IdentityProviderConfig{
			Type: configv1.IdentityProviderTypeHTPasswd,
			HTPasswd: &configv1.HTPasswdIdentityProvider{
				FileData: configv1.SecretNameReference{
					Name: constants.HTPasswdSecretName,
				},
			},
		},
	}

	oauth.Spec.IdentityProviders = append(oauth.Spec.IdentityProviders, htpasswdProvider)

	err = kubeClient.Update(ctx, oauth)
	if err != nil {
		return false, fmt.Errorf("failed to update OAuth CR: %w", err)
	}

	return true, nil
}
