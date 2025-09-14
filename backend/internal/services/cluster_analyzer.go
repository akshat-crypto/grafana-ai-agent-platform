package services

import (
	"context"
	"fmt"
	"strings"

	"grafana-ai-agent-platform/backend/internal/agent"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// ClusterAnalyzerService analyzes Kubernetes clusters
type ClusterAnalyzerService struct{}

// NewClusterAnalyzerService creates a new cluster analyzer service
func NewClusterAnalyzerService() *ClusterAnalyzerService {
	return &ClusterAnalyzerService{}
}

// AnalyzeCluster analyzes a Kubernetes cluster and returns detailed information
func (s *ClusterAnalyzerService) AnalyzeCluster(ctx context.Context, kubeconfig string) (*agent.ClusterAnalysis, error) {
	// Create Kubernetes client
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfig))
	if err != nil {
		return nil, fmt.Errorf("failed to create kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	// Get cluster version
	version, err := clientset.Discovery().ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get server version: %w", err)
	}

	// Get nodes
	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	// Get storage classes
	storageClasses, err := clientset.StorageV1().StorageClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list storage classes: %w", err)
	}

	// Get namespaces to check for capabilities
	namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	// Analyze nodes
	nodeInfos := s.analyzeNodes(nodes.Items)

	// Analyze cluster resources
	resources := s.analyzeClusterResources(nodes.Items)

	// Analyze cluster capabilities
	capabilities := s.analyzeClusterCapabilities(clientset, namespaces.Items)

	// Analyze security
	security := s.analyzeSecurity(clientset)

	// Get storage class names
	storageClassNames := make([]string, len(storageClasses.Items))
	for i, sc := range storageClasses.Items {
		storageClassNames[i] = sc.Name
	}

	// Create cluster analysis
	analysis := &agent.ClusterAnalysis{
		ClusterName:    "analyzed-cluster", // This could be extracted from context or config
		Version:        version.GitVersion,
		Nodes:          nodeInfos,
		Resources:      resources,
		Capabilities:   capabilities,
		StorageClasses: storageClassNames,
		NetworkPolicy:  s.detectNetworkPolicy(clientset),
		Security:       security,
	}

	return analysis, nil
}

// analyzeNodes analyzes node information
func (s *ClusterAnalyzerService) analyzeNodes(nodes []corev1.Node) []agent.NodeInfo {
	nodeInfos := make([]agent.NodeInfo, len(nodes))

	for i, node := range nodes {
		// Get node role
		role := "worker"
		if _, exists := node.Labels["node-role.kubernetes.io/control-plane"]; exists {
			role = "control-plane"
		} else if _, exists := node.Labels["node-role.kubernetes.io/master"]; exists {
			role = "master"
		}

		// Analyze CPU resources
		cpu := s.analyzeResource(node.Status.Capacity.Cpu(), node.Status.Allocatable.Cpu())

		// Analyze memory resources
		memory := s.analyzeResource(node.Status.Capacity.Memory(), node.Status.Allocatable.Memory())

		// Analyze storage resources
		storage := s.analyzeResource(node.Status.Capacity.StorageEphemeral(), node.Status.Allocatable.StorageEphemeral())

		nodeInfos[i] = agent.NodeInfo{
			Name:        node.Name,
			Role:        role,
			Status:      string(node.Status.Conditions[len(node.Status.Conditions)-1].Type),
			CPU:         cpu,
			Memory:      memory,
			Storage:     storage,
			Labels:      node.Labels,
			Annotations: node.Annotations,
		}
	}

	return nodeInfos
}

// analyzeResource analyzes a specific resource
func (s *ClusterAnalyzerService) analyzeResource(capacity, allocatable *resource.Quantity) agent.ResourceInfo {
	if capacity == nil || allocatable == nil {
		return agent.ResourceInfo{
			Capacity:    "0",
			Allocatable: "0",
			Used:        "0",
			Percentage:  0,
		}
	}

	capacityStr := capacity.String()
	allocatableStr := allocatable.String()
	
	// Calculate used resources - create a copy to avoid modifying original
	used := *capacity
	used.Sub(*allocatable)
	usedStr := used.String()
	
	// Calculate percentage
	var percentage int
	if capacity.Value() > 0 {
		percentage = int((allocatable.Value() * 100) / capacity.Value())
	}

	return agent.ResourceInfo{
		Capacity:    capacityStr,
		Allocatable: allocatableStr,
		Used:        usedStr,
		Percentage:  percentage,
	}
}

// analyzeClusterResources analyzes overall cluster resources
func (s *ClusterAnalyzerService) analyzeClusterResources(nodes []corev1.Node) agent.ClusterResources {
	var totalCPU, totalMemory, totalStorage resource.Quantity
	var availableCPU, availableMemory, availableStorage resource.Quantity

	for _, node := range nodes {
		if node.Status.Capacity.Cpu() != nil {
			totalCPU.Add(*node.Status.Capacity.Cpu())
		}
		if node.Status.Allocatable.Cpu() != nil {
			availableCPU.Add(*node.Status.Allocatable.Cpu())
		}

		if node.Status.Capacity.Memory() != nil {
			totalMemory.Add(*node.Status.Capacity.Memory())
		}
		if node.Status.Allocatable.Memory() != nil {
			availableMemory.Add(*node.Status.Allocatable.Memory())
		}

		if node.Status.Capacity.StorageEphemeral() != nil {
			totalStorage.Add(*node.Status.Capacity.StorageEphemeral())
		}
		if node.Status.Allocatable.StorageEphemeral() != nil {
			availableStorage.Add(*node.Status.Allocatable.StorageEphemeral())
		}
	}

	return agent.ClusterResources{
		TotalCPU:         totalCPU.String(),
		TotalMemory:      totalMemory.String(),
		TotalStorage:     totalStorage.String(),
		AvailableCPU:     availableCPU.String(),
		AvailableMemory:  availableMemory.String(),
		AvailableStorage: availableStorage.String(),
	}
}

// analyzeClusterCapabilities analyzes cluster capabilities
func (s *ClusterAnalyzerService) analyzeClusterCapabilities(clientset *kubernetes.Clientset, namespaces []corev1.Namespace) agent.ClusterCapabilities {
	capabilities := agent.ClusterCapabilities{
		HelmInstalled:    false,
		IngressAvailable: false,
		LoadBalancer:     false,
		PersistentVolume: false,
		RBACEnabled:      false,
		NetworkPolicy:    false,
	}

	// Check for Helm installation
	if _, err := clientset.CoreV1().Namespaces().Get(context.Background(), "kube-system", metav1.GetOptions{}); err == nil {
		// Check for Helm-related resources
		secrets, err := clientset.CoreV1().Secrets("kube-system").List(context.Background(), metav1.ListOptions{})
		if err == nil {
			for _, secret := range secrets.Items {
				if strings.Contains(secret.Name, "helm") || strings.Contains(secret.Name, "tiller") {
					capabilities.HelmInstalled = true
					break
				}
			}
		}
	}

	// Check for ingress controller
	ingresses, err := clientset.NetworkingV1().Ingresses("").List(context.Background(), metav1.ListOptions{})
	if err == nil && len(ingresses.Items) > 0 {
		capabilities.IngressAvailable = true
	}

	// Check for load balancer services
	services, err := clientset.CoreV1().Services("").List(context.Background(), metav1.ListOptions{})
	if err == nil {
		for _, service := range services.Items {
			if service.Spec.Type == corev1.ServiceTypeLoadBalancer {
				capabilities.LoadBalancer = true
				break
			}
		}
	}

	// Check for persistent volumes
	pvs, err := clientset.CoreV1().PersistentVolumes().List(context.Background(), metav1.ListOptions{})
	if err == nil && len(pvs.Items) > 0 {
		capabilities.PersistentVolume = true
	}

	// Check for RBAC
	if _, err := clientset.RbacV1().ClusterRoles().List(context.Background(), metav1.ListOptions{}); err == nil {
		capabilities.RBACEnabled = true
	}

	// Check for network policies
	if _, err := clientset.NetworkingV1().NetworkPolicies("").List(context.Background(), metav1.ListOptions{}); err == nil {
		capabilities.NetworkPolicy = true
	}

	return capabilities
}

// analyzeSecurity analyzes security features
func (s *ClusterAnalyzerService) analyzeSecurity(clientset *kubernetes.Clientset) agent.SecurityInfo {
	security := agent.SecurityInfo{
		RBACEnabled:       false,
		PodSecurityPolicy: false,
		NetworkPolicy:     false,
		SecretsEnabled:    false,
	}

	// Check RBAC
	if _, err := clientset.RbacV1().ClusterRoles().List(context.Background(), metav1.ListOptions{}); err == nil {
		security.RBACEnabled = true
	}

	// Check for pod security policies (deprecated in v1.21+)
	if _, err := clientset.PolicyV1beta1().PodSecurityPolicies().List(context.Background(), metav1.ListOptions{}); err == nil {
		security.PodSecurityPolicy = true
	}

	// Check for network policies
	if _, err := clientset.NetworkingV1().NetworkPolicies("").List(context.Background(), metav1.ListOptions{}); err == nil {
		security.NetworkPolicy = true
	}

	// Check for secrets
	if _, err := clientset.CoreV1().Secrets("").List(context.Background(), metav1.ListOptions{}); err == nil {
		security.SecretsEnabled = true
	}

	return security
}

// detectNetworkPolicy detects network policy support
func (s *ClusterAnalyzerService) detectNetworkPolicy(clientset *kubernetes.Clientset) string {
	// Check for network policy support
	if _, err := clientset.NetworkingV1().NetworkPolicies("").List(context.Background(), metav1.ListOptions{}); err == nil {
		return "supported"
	}
	return "not-supported"
}
