package storage

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/observability-ui/development-tools/pkg/executor"
)

const (
	DefaultMinioImage = "quay.io/minio/minio:latest"
	DefaultAccessKey  = "minio"
	DefaultSecretKey  = "minio123"
	MinioPort         = 9000
	ReadyTimeout      = 5 * time.Minute
	PollInterval      = 2 * time.Second
)

const (
	StepCreateNamespace = iota
	StepCreatePVC
	StepCreateDeployment
	StepCreateService
	StepCreateSecret
	StepWaitForReady
)

const (
	StepDeleteSecret = iota
	StepDeleteService
	StepDeleteDeployment
	StepDeletePVC
)

type MinioProvider struct {
	config ProviderConfig
}

func NewMinioProvider(config ProviderConfig) *MinioProvider {
	return &MinioProvider{
		config: config,
	}
}

func (m *MinioProvider) GetType() StorageType {
	return StorageTypeMinio
}

func (m *MinioProvider) GetSecretName() string {
	if m.config.SecretFormat == SecretFormatThanos {
		return "thanos-object-storage"
	}
	return m.config.BucketName + "-object-storage"
}

func (m *MinioProvider) GetEndpoint() string {
	return fmt.Sprintf("http://minio.%s.svc:%d", m.config.Namespace, MinioPort)
}

func (m *MinioProvider) GetBucketName() string {
	return m.config.BucketName
}

func (m *MinioProvider) Deploy(ctx context.Context, k8sClient client.Client, exec *executor.Executor) (string, error) {
	stepName := fmt.Sprintf("Create namespace %s", m.config.Namespace)
	exec.SendUpdate(StepCreateNamespace, executor.StatusInProgress, stepName)
	exec.SendLog(StepCreateNamespace, fmt.Sprintf("Ensuring namespace %s exists", m.config.Namespace))

	if err := m.createOrGetNamespace(ctx, k8sClient); err != nil {
		exec.SendUpdateWithError(StepCreateNamespace, executor.StatusFailed, stepName, err)
		return "", fmt.Errorf("failed to create namespace: %w", err)
	}
	exec.SendUpdate(StepCreateNamespace, executor.StatusComplete, stepName)

	stepName = "Create PersistentVolumeClaim"
	exec.SendUpdate(StepCreatePVC, executor.StatusInProgress, stepName)
	exec.SendLog(StepCreatePVC, fmt.Sprintf("Creating PVC with size %s", m.config.StorageSize))

	if err := m.createPVC(ctx, k8sClient); err != nil {
		exec.SendUpdateWithError(StepCreatePVC, executor.StatusFailed, stepName, err)
		return "", fmt.Errorf("failed to create PVC: %w", err)
	}
	exec.SendUpdate(StepCreatePVC, executor.StatusComplete, stepName)

	stepName = "Create MinIO deployment"
	exec.SendUpdate(StepCreateDeployment, executor.StatusInProgress, stepName)
	exec.SendLog(StepCreateDeployment, fmt.Sprintf("Deploying MinIO for bucket: %s", m.config.BucketName))

	if err := m.createDeployment(ctx, k8sClient); err != nil {
		exec.SendUpdateWithError(StepCreateDeployment, executor.StatusFailed, stepName, err)
		return "", fmt.Errorf("failed to create deployment: %w", err)
	}
	exec.SendUpdate(StepCreateDeployment, executor.StatusComplete, stepName)

	stepName = "Create MinIO service"
	exec.SendUpdate(StepCreateService, executor.StatusInProgress, stepName)
	exec.SendLog(StepCreateService, fmt.Sprintf("Creating service on port %d", MinioPort))

	if err := m.createService(ctx, k8sClient); err != nil {
		exec.SendUpdateWithError(StepCreateService, executor.StatusFailed, stepName, err)
		return "", fmt.Errorf("failed to create service: %w", err)
	}
	exec.SendUpdate(StepCreateService, executor.StatusComplete, stepName)

	stepName = "Create storage credentials secret"
	exec.SendUpdate(StepCreateSecret, executor.StatusInProgress, stepName)
	exec.SendLog(StepCreateSecret, fmt.Sprintf("Creating secret: %s", m.GetSecretName()))

	if err := m.createSecret(ctx, k8sClient); err != nil {
		exec.SendUpdateWithError(StepCreateSecret, executor.StatusFailed, stepName, err)
		return "", fmt.Errorf("failed to create secret: %w", err)
	}
	exec.SendUpdate(StepCreateSecret, executor.StatusComplete, stepName)

	stepName = "Wait for MinIO to be ready"
	exec.SendUpdate(StepWaitForReady, executor.StatusInProgress, stepName)
	exec.SendLog(StepWaitForReady, fmt.Sprintf("Waiting for deployment to become ready (timeout: %s)", ReadyTimeout))

	if err := m.waitForReady(ctx, k8sClient, exec); err != nil {
		exec.SendUpdateWithError(StepWaitForReady, executor.StatusFailed, stepName, err)
		return "", fmt.Errorf("deployment not ready: %w", err)
	}
	exec.SendUpdate(StepWaitForReady, executor.StatusComplete, stepName)

	return m.GetSecretName(), nil
}

func (m *MinioProvider) Delete(ctx context.Context, k8sClient client.Client, exec *executor.Executor) error {
	stepName := "Delete storage credentials secret"
	exec.SendUpdate(StepDeleteSecret, executor.StatusInProgress, stepName)
	exec.SendLog(StepDeleteSecret, fmt.Sprintf("Deleting secret: %s", m.GetSecretName()))

	if err := k8sClient.Delete(ctx, m.buildSecret()); err != nil && !errors.IsNotFound(err) {
		exec.SendUpdateWithError(StepDeleteSecret, executor.StatusFailed, stepName, err)
		return fmt.Errorf("failed to delete secret: %w", err)
	}
	exec.SendUpdate(StepDeleteSecret, executor.StatusComplete, stepName)

	stepName = "Delete MinIO service"
	exec.SendUpdate(StepDeleteService, executor.StatusInProgress, stepName)
	exec.SendLog(StepDeleteService, "Deleting service")

	if err := k8sClient.Delete(ctx, m.buildService()); err != nil && !errors.IsNotFound(err) {
		exec.SendUpdateWithError(StepDeleteService, executor.StatusFailed, stepName, err)
		return fmt.Errorf("failed to delete service: %w", err)
	}
	exec.SendUpdate(StepDeleteService, executor.StatusComplete, stepName)

	stepName = "Delete MinIO deployment"
	exec.SendUpdate(StepDeleteDeployment, executor.StatusInProgress, stepName)
	exec.SendLog(StepDeleteDeployment, "Deleting deployment")

	if err := k8sClient.Delete(ctx, m.buildDeployment()); err != nil && !errors.IsNotFound(err) {
		exec.SendUpdateWithError(StepDeleteDeployment, executor.StatusFailed, stepName, err)
		return fmt.Errorf("failed to delete deployment: %w", err)
	}
	exec.SendUpdate(StepDeleteDeployment, executor.StatusComplete, stepName)

	stepName = "Delete PersistentVolumeClaim"
	exec.SendUpdate(StepDeletePVC, executor.StatusInProgress, stepName)
	exec.SendLog(StepDeletePVC, "Deleting PVC and associated storage")

	if err := k8sClient.Delete(ctx, m.buildPVC()); err != nil && !errors.IsNotFound(err) {
		exec.SendUpdateWithError(StepDeletePVC, executor.StatusFailed, stepName, err)
		return fmt.Errorf("failed to delete PVC: %w", err)
	}
	exec.SendUpdate(StepDeletePVC, executor.StatusComplete, stepName)

	return nil
}

func (m *MinioProvider) createOrGetNamespace(ctx context.Context, k8sClient client.Client) error {
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: m.config.Namespace,
		},
	}

	err := k8sClient.Create(ctx, namespace)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	return nil
}

func (m *MinioProvider) createPVC(ctx context.Context, k8sClient client.Client) error {
	pvc := m.buildPVC()
	err := k8sClient.Create(ctx, pvc)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func (m *MinioProvider) createDeployment(ctx context.Context, k8sClient client.Client) error {
	deployment := m.buildDeployment()
	err := k8sClient.Create(ctx, deployment)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func (m *MinioProvider) createService(ctx context.Context, k8sClient client.Client) error {
	service := m.buildService()
	err := k8sClient.Create(ctx, service)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func (m *MinioProvider) createSecret(ctx context.Context, k8sClient client.Client) error {
	secret := m.buildSecret()
	err := k8sClient.Create(ctx, secret)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func (m *MinioProvider) waitForReady(ctx context.Context, k8sClient client.Client, exec *executor.Executor) error {
	deployment := &appsv1.Deployment{}
	key := types.NamespacedName{Name: "minio", Namespace: m.config.Namespace}

	attempts := 0
	return wait.PollImmediate(PollInterval, ReadyTimeout, func() (bool, error) {
		attempts++
		if err := k8sClient.Get(ctx, key, deployment); err != nil {
			if errors.IsNotFound(err) {
				return false, nil
			}
			return false, err
		}

		if deployment.Status.ReadyReplicas == 1 {
			exec.SendLog(StepWaitForReady, "MinIO deployment is ready")
			return true, nil
		}

		if attempts%5 == 0 {
			exec.SendLog(StepWaitForReady, fmt.Sprintf("Still waiting... (ready: %d/1)", deployment.Status.ReadyReplicas))
		}

		return false, nil
	})
}

func (m *MinioProvider) buildPVC() *corev1.PersistentVolumeClaim {
	var storageClassName *string
	if !m.config.UseDefaultStorageClass && m.config.StorageClass != "" {
		storageClassName = &m.config.StorageClass
	}

	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "minio",
			Namespace: m.config.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name": "minio",
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			StorageClassName: storageClassName,
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(m.config.StorageSize),
				},
			},
		},
	}
}

func (m *MinioProvider) buildDeployment() *appsv1.Deployment {
	replicas := int32(1)

	image := DefaultMinioImage
	envVars := []corev1.EnvVar{
		{Name: "MINIO_ROOT_USER", Value: DefaultAccessKey},
		{Name: "MINIO_ROOT_PASSWORD", Value: DefaultSecretKey},
	}
	command := fmt.Sprintf("mkdir -p /storage/%s && minio server /storage", m.config.BucketName)

	if m.config.SecretFormat == SecretFormatThanos {
		image = "quay.io/minio/minio:RELEASE.2021-08-25T00-41-18Z"
		envVars = []corev1.EnvVar{
			{Name: "MINIO_ACCESS_KEY", Value: DefaultAccessKey},
			{Name: "MINIO_SECRET_KEY", Value: DefaultSecretKey},
		}
		command = fmt.Sprintf("mkdir -p /storage/%s && /usr/bin/minio server /storage", m.config.BucketName)
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "minio",
			Namespace: m.config.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name": "minio",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/name": "minio",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app.kubernetes.io/name": "minio",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "minio",
							Image:   image,
							Command: []string{"/bin/sh", "-c", command},
							Env:     envVars,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: MinioPort,
									Name:          "http",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "storage",
									MountPath: "/storage",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "storage",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "minio",
								},
							},
						},
					},
				},
			},
		},
	}
}

func (m *MinioProvider) buildService() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "minio",
			Namespace: m.config.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name": "minio",
			},
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Selector: map[string]string{
				"app.kubernetes.io/name": "minio",
			},
			Ports: []corev1.ServicePort{
				{
					Port:     MinioPort,
					Protocol: corev1.ProtocolTCP,
					Name:     "http",
				},
			},
		},
	}
}

func (m *MinioProvider) buildSecret() *corev1.Secret {
	if m.config.SecretFormat == SecretFormatThanos {
		thanosYAML := fmt.Sprintf(`type: s3
config:
  bucket: "%s"
  endpoint: "minio:9000"
  insecure: true
  access_key: "%s"
  secret_key: "%s"
`, m.config.BucketName, DefaultAccessKey, DefaultSecretKey)

		return &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      m.GetSecretName(),
				Namespace: m.config.Namespace,
			},
			Data: map[string][]byte{
				"thanos.yaml": []byte(thanosYAML),
			},
			Type: corev1.SecretTypeOpaque,
		}
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.GetSecretName(),
			Namespace: m.config.Namespace,
		},
		Data: map[string][]byte{
			"access_key_id":     []byte(DefaultAccessKey),
			"access_key_secret": []byte(DefaultSecretKey),
			"bucketname":        []byte(m.config.BucketName),
			"endpoint":          []byte(m.GetEndpoint()),
		},
		Type: corev1.SecretTypeOpaque,
	}
}
