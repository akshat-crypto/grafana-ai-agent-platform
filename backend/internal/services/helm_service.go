package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"grafana-ai-agent-platform/backend/internal/agent"
)

// HelmService handles Helm chart operations
type HelmService struct {
	artifactHubClient *http.Client
}

// NewHelmService creates a new Helm service
func NewHelmService() *HelmService {
	return &HelmService{
		artifactHubClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ChartSearchResult represents a search result from Artifact Hub
type ChartSearchResult struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Repository  string   `json:"repository"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	URL         string   `json:"url"`
	HomeURL     string   `json:"home_url"`
	Keywords    []string `json:"keywords"`
	Maintainers []struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"maintainers"`
	Provider   string `json:"provider"`
	Deprecated bool   `json:"deprecated"`
}

// SearchCharts searches for Helm charts on Artifact Hub
func (s *HelmService) SearchCharts(query string) ([]ChartSearchResult, error) {
	// Artifact Hub search API
	url := fmt.Sprintf("https://artifacthub.io/api/v1/packages/search?q=%s&kind=0&limit=20", query)

	resp, err := s.artifactHubClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to search charts: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var results []ChartSearchResult
	if err := json.Unmarshal(body, &results); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return results, nil
}

// GetChartDetails gets detailed information about a specific chart
func (s *HelmService) GetChartDetails(chartID string) (*ChartDetails, error) {
	url := fmt.Sprintf("https://artifacthub.io/api/v1/packages/%s", chartID)

	resp, err := s.artifactHubClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get chart details: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get chart details with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var details ChartDetails
	if err := json.Unmarshal(body, &details); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &details, nil
}

// ChartDetails represents detailed chart information
type ChartDetails struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Repository  string   `json:"repository"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	URL         string   `json:"url"`
	HomeURL     string   `json:"home_url"`
	Keywords    []string `json:"keywords"`
	Maintainers []struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"maintainers"`
	Provider   string `json:"provider"`
	Deprecated bool   `json:"deprecated"`
	Values     string `json:"values"` // Default values.yaml content
	Readme     string `json:"readme"` // README content
}

// GenerateValues generates Helm values based on cluster analysis and requirements
func (s *HelmService) GenerateValues(chart *agent.HelmChart, clusterAnalysis *agent.ClusterAnalysis, requirements map[string]interface{}) (map[string]interface{}, error) {
	// Start with default values
	values := make(map[string]interface{})

	// Copy chart values if available
	if chart.Values != nil {
		for k, v := range chart.Values {
			values[k] = v
		}
	}

	// Apply cluster-specific customizations
	s.customizeForCluster(values, clusterAnalysis)

	// Apply user requirements
	s.applyUserRequirements(values, requirements)

	// Apply best practices
	s.applyBestPractices(values, chart.Name)

	return values, nil
}

// customizeForCluster customizes values based on cluster capabilities
func (s *HelmService) customizeForCluster(values map[string]interface{}, cluster *agent.ClusterAnalysis) {
	// Set resource limits based on cluster capacity
	if cluster.Resources.AvailableCPU != "" && cluster.Resources.AvailableMemory != "" {
		// Calculate reasonable resource limits (e.g., 20% of available resources)
		s.setResourceLimits(values, cluster.Resources)
	}

	// Configure storage based on available storage classes
	if len(cluster.StorageClasses) > 0 {
		s.configureStorage(values, cluster.StorageClasses)
	}

	// Configure ingress if available
	if cluster.Capabilities.IngressAvailable {
		s.configureIngress(values)
	}

	// Configure security settings
	if cluster.Security.RBACEnabled {
		s.configureRBAC(values)
	}
}

// setResourceLimits sets resource limits based on cluster capacity
func (s *HelmService) setResourceLimits(values map[string]interface{}, resources agent.ClusterResources) {
	// This is a simplified approach - in production, you'd want more sophisticated resource calculation
	resourceConfig := map[string]interface{}{
		"resources": map[string]interface{}{
			"limits": map[string]interface{}{
				"cpu":    "500m",
				"memory": "512Mi",
			},
			"requests": map[string]interface{}{
				"cpu":    "100m",
				"memory": "128Mi",
			},
		},
	}

	s.mergeValues(values, resourceConfig)
}

// configureStorage configures storage settings
func (s *HelmService) configureStorage(values map[string]interface{}, storageClasses []string) {
	// Use the first available storage class
	if len(storageClasses) > 0 {
		storageConfig := map[string]interface{}{
			"persistence": map[string]interface{}{
				"storageClass": storageClasses[0],
			},
		}
		s.mergeValues(values, storageConfig)
	}
}

// configureIngress configures ingress settings
func (s *HelmService) configureIngress(values map[string]interface{}) {
	ingressConfig := map[string]interface{}{
		"ingress": map[string]interface{}{
			"enabled": true,
			"annotations": map[string]interface{}{
				"kubernetes.io/ingress.class": "nginx",
			},
		},
	}
	s.mergeValues(values, ingressConfig)
}

// configureRBAC configures RBAC settings
func (s *HelmService) configureRBAC(values map[string]interface{}) {
	rbacConfig := map[string]interface{}{
		"rbac": map[string]interface{}{
			"create": true,
		},
	}
	s.mergeValues(values, rbacConfig)
}

// applyUserRequirements applies user-specific requirements
func (s *HelmService) applyUserRequirements(values map[string]interface{}, requirements map[string]interface{}) {
	if requirements != nil {
		s.mergeValues(values, requirements)
	}
}

// applyBestPractices applies security and operational best practices
func (s *HelmService) applyBestPractices(values map[string]interface{}, chartName string) {
	// Apply security best practices
	securityConfig := map[string]interface{}{
		"securityContext": map[string]interface{}{
			"runAsNonRoot": true,
			"runAsUser":    1000,
		},
	}
	s.mergeValues(values, securityConfig)

	// Apply monitoring best practices
	if strings.Contains(strings.ToLower(chartName), "prometheus") || strings.Contains(strings.ToLower(chartName), "grafana") {
		monitoringConfig := map[string]interface{}{
			"serviceMonitor": map[string]interface{}{
				"enabled": true,
			},
		}
		s.mergeValues(values, monitoringConfig)
	}
}

// mergeValues merges configuration values
func (s *HelmService) mergeValues(target, source map[string]interface{}) {
	for key, value := range source {
		if targetValue, exists := target[key]; exists {
			// If both are maps, merge recursively
			if targetMap, ok := targetValue.(map[string]interface{}); ok {
				if sourceMap, ok := value.(map[string]interface{}); ok {
					s.mergeValues(targetMap, sourceMap)
					continue
				}
			}
		}
		// Otherwise, overwrite
		target[key] = value
	}
}

// CreateDeploymentPlan creates a deployment plan for a specific stack
func (s *HelmService) CreateDeploymentPlan(stackName string, clusterAnalysis *agent.ClusterAnalysis) (*agent.DeploymentPlan, error) {
	// Search for relevant charts
	charts, err := s.SearchCharts(stackName)
	if err != nil {
		return nil, fmt.Errorf("failed to search charts: %w", err)
	}

	if len(charts) == 0 {
		return nil, fmt.Errorf("no charts found for stack: %s", stackName)
	}

	// Create deployment plan
	plan := &agent.DeploymentPlan{
		ID:            fmt.Sprintf("plan-%s-%d", stackName, time.Now().Unix()),
		Name:          fmt.Sprintf("Deploy %s Stack", stackName),
		Description:   fmt.Sprintf("Deployment plan for %s stack", stackName),
		Charts:        make([]agent.HelmChart, 0),
		Steps:         make([]agent.DeploymentStep, 0),
		EstimatedTime: "10-15 minutes",
		ResourceImpact: agent.ResourceImpact{
			CPU:     "500m",
			Memory:  "1Gi",
			Storage: "10Gi",
			Nodes:   1,
		},
		Prerequisites: []string{
			"Kubernetes cluster with sufficient resources",
			"kubectl configured and accessible",
			"Helm 3.x installed (optional, can be installed automatically)",
		},
		Risks: []string{
			"Resource consumption may impact other workloads",
			"Configuration changes may affect existing services",
			"Rollback may be required if issues occur",
		},
	}

	// Add charts to the plan
	for i, chart := range charts[:3] { // Limit to top 3 charts
		helmChart := agent.HelmChart{
			Name:        chart.Name,
			Repository:  chart.Repository,
			Version:     chart.Version,
			Description: chart.Description,
			URL:         chart.URL,
			Values:      make(map[string]interface{}),
		}

		// Generate values for this chart
		values, err := s.GenerateValues(&helmChart, clusterAnalysis, nil)
		if err == nil {
			helmChart.Values = values
		}

		plan.Charts = append(plan.Charts, helmChart)

		// Create deployment step
		step := agent.DeploymentStep{
			ID:          fmt.Sprintf("step-%d", i+1),
			Name:        fmt.Sprintf("Deploy %s", chart.Name),
			Description: fmt.Sprintf("Deploy %s chart from %s repository", chart.Name, chart.Repository),
			Chart:       &helmChart,
			Status:      "pending",
		}
		plan.Steps = append(plan.Steps, step)
	}

	return plan, nil
}
