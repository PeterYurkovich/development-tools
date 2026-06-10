package k8s

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Client struct {
	client.Client
	config *rest.Config
}

// Priority: kubeconfigPath parameter > KUBECONFIG env > ~/.kube/config
func NewClient(ctx context.Context, kubeconfigPath string) (*Client, error) {
	config, err := getKubeConfig(kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	config.Timeout = 30 * time.Second
	config.QPS = 50
	config.Burst = 100

	scheme := runtime.NewScheme()
	if err := registerSchemes(scheme); err != nil {
		return nil, fmt.Errorf("failed to register schemes: %w", err)
	}

	kubeClient, err := client.New(config, client.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return &Client{
		Client: kubeClient,
		config: config,
	}, nil
}

func getKubeConfig(kubeconfigPath string) (*rest.Config, error) {
	if kubeconfigPath != "" {
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load kubeconfig from %s: %w", kubeconfigPath, err)
		}
		return config, nil
	}

	if kubeconfigEnv := os.Getenv("KUBECONFIG"); kubeconfigEnv != "" {
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfigEnv)
		if err != nil {
			return nil, fmt.Errorf("failed to load kubeconfig from KUBECONFIG env (%s): %w", kubeconfigEnv, err)
		}
		return config, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	defaultPath := filepath.Join(home, ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", defaultPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig from default location (%s): %w. Provide --kubeconfig flag or set KUBECONFIG env", defaultPath, err)
	}

	return config, nil
}

func (c *Client) Config() *rest.Config {
	return c.config
}
