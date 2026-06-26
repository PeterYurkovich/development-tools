package resources

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/observability-ui/development-tools/internal/constants"
)

func CreateChatApp(ctx context.Context, kubeClient client.Client) error {
	if err := ensureNamespaceExists(ctx, kubeClient, constants.ChatNamespace); err != nil {
		return err
	}

	allowPrivilegeEscalation := false
	runAsNonRoot := true
	replicas := int32(1)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.ChatDeploymentName,
			Namespace: constants.ChatNamespace,
			Labels: map[string]string{
				"app":  "chat",
				"test": "true",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "chat"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":  "chat",
						"test": "true",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:    "chat",
						Image:   constants.ChatImage,
						Command: []string{"sh", "-c"},
						Args:    []string{`i=1; while true; do echo "$(date) chat says hello - $i"; i=$((i + 1)); sleep 1; done`},
						SecurityContext: &corev1.SecurityContext{
							AllowPrivilegeEscalation: &allowPrivilegeEscalation,
							RunAsNonRoot:             &runAsNonRoot,
							SeccompProfile: &corev1.SeccompProfile{
								Type: corev1.SeccompProfileTypeRuntimeDefault,
							},
							Capabilities: &corev1.Capabilities{
								Drop: []corev1.Capability{"ALL"},
							},
						},
					}},
				},
			},
		},
	}

	if err := kubeClient.Create(ctx, deployment); err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create chat deployment: %w", err)
	}

	return nil
}
