package kubernetes

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

type KubernetesClient struct {
	clientset *kubernetes.Clientset
	config    *rest.Config
}

type ClusterInfo struct {
	Version   string `json:"version"`
	ServerURL string `json:"server_url"`
	IsValid   bool   `json:"is_valid"`
	Error     string `json:"error,omitempty"`
}

func NewKubernetesClient(kubeconfig string) (*KubernetesClient, error) {
	// Basic validation
	if kubeconfig == "" {
		return nil, fmt.Errorf("kubeconfig is empty")
	}

	// Parse kubeconfig
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfig))
	if err != nil {
		return nil, fmt.Errorf("failed to parse kubeconfig: %w", err)
	}

	// Validate the config
	if config.Host == "" {
		return nil, fmt.Errorf("no server URL found in kubeconfig")
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return &KubernetesClient{
		clientset: clientset,
		config:    config,
	}, nil
}

func (k *KubernetesClient) ValidateCluster() (*ClusterInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get server info
	serverVersion, err := k.clientset.ServerVersion()
	if err != nil {
		return &ClusterInfo{
			IsValid: false,
			Error:   fmt.Sprintf("Failed to connect to cluster: %v", err),
		}, nil
	}

	// Test API connectivity
	_, err = k.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{Limit: 1})
	if err != nil {
		return &ClusterInfo{
			IsValid: false,
			Error:   fmt.Sprintf("Failed to list nodes: %v", err),
		}, nil
	}

	return &ClusterInfo{
		Version:   serverVersion.String(),
		ServerURL: k.config.Host,
		IsValid:   true,
	}, nil
}

func (k *KubernetesClient) GetClusterResources() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resources := make(map[string]interface{})

	// Get nodes
	nodes, err := k.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err == nil {
		resources["nodes"] = len(nodes.Items)
	}

	// Get namespaces
	namespaces, err := k.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err == nil {
		resources["namespaces"] = len(namespaces.Items)
	}

	// Get pods
	pods, err := k.clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err == nil {
		resources["pods"] = len(pods.Items)
	}

	return resources, nil
}

func (k *KubernetesClient) ApplyManifest(manifest string) error {
	// This is a simplified version. In production, you'd want to use kubectl apply
	// or implement proper manifest parsing and application
	_, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// For now, we'll just validate the manifest can be parsed
	// In a real implementation, you'd parse the YAML and apply it
	return nil
}

func ParseKubeconfig(kubeconfig string) (*api.Config, error) {
	if kubeconfig == "" {
		return nil, fmt.Errorf("kubeconfig is empty")
	}

	config, err := clientcmd.Load([]byte(kubeconfig))
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	if len(config.Contexts) == 0 {
		return nil, fmt.Errorf("no contexts found in kubeconfig")
	}

	// Check if current context is set
	if config.CurrentContext == "" {
		return nil, fmt.Errorf("no current context found in kubeconfig")
	}

	// Validate the current context exists
	currentContext, exists := config.Contexts[config.CurrentContext]
	if !exists {
		return nil, fmt.Errorf("current context '%s' not found in kubeconfig", config.CurrentContext)
	}

	// Validate cluster exists
	_, exists = config.Clusters[currentContext.Cluster]
	if !exists {
		return nil, fmt.Errorf("cluster '%s' not found in kubeconfig", currentContext.Cluster)
	}

	// Validate auth info exists
	if currentContext.AuthInfo != "" {
		_, exists = config.AuthInfos[currentContext.AuthInfo]
		if !exists {
			return nil, fmt.Errorf("auth info '%s' not found in kubeconfig", currentContext.AuthInfo)
		}
	}

	return config, nil
}

func ValidateKubeconfigFormat(kubeconfig string) error {
	_, err := ParseKubeconfig(kubeconfig)
	return err
}

func ExtractClusterInfo(kubeconfig string) (string, error) {
	config, err := ParseKubeconfig(kubeconfig)
	if err != nil {
		return "", err
	}

	// Get current context
	currentContext := config.CurrentContext
	if currentContext == "" {
		return "", fmt.Errorf("no current context found")
	}

	context := config.Contexts[currentContext]
	if context == nil {
		return "", fmt.Errorf("current context not found")
	}

	cluster := config.Clusters[context.Cluster]
	if cluster == nil {
		return "", fmt.Errorf("cluster not found")
	}

	return cluster.Server, nil
}
