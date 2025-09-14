package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
)

// AIAgent handles AI-powered Kubernetes operations
type AIAgent struct {
	client *openai.Client
	cfg    *Config
}

// Config holds AI agent configuration
type Config struct {
	OpenAIAPIKey     string
	OpenRouterAPIKey string
	Model            string
	UseOpenRouter    bool
}

// NewAIAgent creates a new AI agent instance
func NewAIAgent(cfg *Config) *AIAgent {
	var client *openai.Client

	if cfg.UseOpenRouter {
		// Configure OpenRouter client
		clientConfig := openai.DefaultConfig(cfg.OpenRouterAPIKey)
		clientConfig.BaseURL = "https://openrouter.ai/api/v1"
		client = openai.NewClientWithConfig(clientConfig)
	} else {
		// Use OpenAI client
		client = openai.NewClient(cfg.OpenAIAPIKey)
	}

	return &AIAgent{
		client: client,
		cfg:    cfg,
	}
}

// QueryRequest represents a user query
type QueryRequest struct {
	Query       string `json:"query"`
	ClusterID   *uint  `json:"cluster_id,omitempty"`
	ClusterName string `json:"cluster_name,omitempty"`
	ClusterInfo string `json:"cluster_info,omitempty"`
}

// QueryResponse represents the AI response
type QueryResponse struct {
	Response        string           `json:"response"`
	DeploymentPlan  *DeploymentPlan  `json:"deployment_plan,omitempty"`
	ClusterAnalysis *ClusterAnalysis `json:"cluster_analysis,omitempty"`
	Status          string           `json:"status"`
	Timestamp       time.Time        `json:"timestamp"`
}

// DeploymentPlan represents a deployment strategy
type DeploymentPlan struct {
	ID             string           `json:"id"`
	Name           string           `json:"name"`
	Description    string           `json:"description"`
	Charts         []HelmChart      `json:"charts"`
	Steps          []DeploymentStep `json:"steps"`
	EstimatedTime  string           `json:"estimated_time"`
	ResourceImpact ResourceImpact   `json:"resource_impact"`
	Prerequisites  []string         `json:"prerequisites"`
	Risks          []string         `json:"risks"`
}

// HelmChart represents a Helm chart to be deployed
type HelmChart struct {
	Name        string                 `json:"name"`
	Repository  string                 `json:"repository"`
	Version     string                 `json:"version"`
	Values      map[string]interface{} `json:"values"`
	Description string                 `json:"description"`
	URL         string                 `json:"url"`
}

// DeploymentStep represents a deployment step
type DeploymentStep struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Chart       *HelmChart `json:"chart,omitempty"`
	Command     string     `json:"command,omitempty"`
	Status      string     `json:"status"` // pending, running, completed, failed
	Logs        []string   `json:"logs"`
	StartTime   *time.Time `json:"start_time,omitempty"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	Error       string     `json:"error,omitempty"`
}

// ResourceImpact represents the impact on cluster resources
type ResourceImpact struct {
	CPU     string `json:"cpu"`
	Memory  string `json:"memory"`
	Storage string `json:"storage"`
	Nodes   int    `json:"nodes"`
}

// ClusterAnalysis represents cluster information and capabilities
type ClusterAnalysis struct {
	ClusterID      uint                `json:"cluster_id"`
	ClusterName    string              `json:"cluster_name"`
	Version        string              `json:"version"`
	Nodes          []NodeInfo          `json:"nodes"`
	Resources      ClusterResources    `json:"resources"`
	Capabilities   ClusterCapabilities `json:"capabilities"`
	StorageClasses []string            `json:"storage_classes"`
	NetworkPolicy  string              `json:"network_policy"`
	Security       SecurityInfo        `json:"security"`
}

// NodeInfo represents information about a cluster node
type NodeInfo struct {
	Name        string            `json:"name"`
	Role        string            `json:"role"`
	Status      string            `json:"status"`
	CPU         ResourceInfo      `json:"cpu"`
	Memory      ResourceInfo      `json:"memory"`
	Storage     ResourceInfo      `json:"storage"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

// ResourceInfo represents resource information
type ResourceInfo struct {
	Capacity    string `json:"capacity"`
	Allocatable string `json:"allocatable"`
	Used        string `json:"used"`
	Percentage  int    `json:"percentage"`
}

// ClusterResources represents overall cluster resources
type ClusterResources struct {
	TotalCPU         string `json:"total_cpu"`
	TotalMemory      string `json:"total_memory"`
	TotalStorage     string `json:"total_storage"`
	AvailableCPU     string `json:"available_cpu"`
	AvailableMemory  string `json:"available_memory"`
	AvailableStorage string `json:"available_storage"`
}

// ClusterCapabilities represents cluster capabilities
type ClusterCapabilities struct {
	HelmInstalled    bool `json:"helm_installed"`
	IngressAvailable bool `json:"ingress_available"`
	LoadBalancer     bool `json:"load_balancer"`
	PersistentVolume bool `json:"persistent_volume"`
	RBACEnabled      bool `json:"rbac_enabled"`
	NetworkPolicy    bool `json:"network_policy"`
}

// SecurityInfo represents security information
type SecurityInfo struct {
	RBACEnabled       bool `json:"rbac_enabled"`
	PodSecurityPolicy bool `json:"pod_security_policy"`
	NetworkPolicy     bool `json:"network_policy"`
	SecretsEnabled    bool `json:"secrets_enabled"`
}

// Query handles user queries and generates responses
func (a *AIAgent) Query(ctx context.Context, req *QueryRequest) (*QueryResponse, error) {
	// Build the system prompt based on the query type
	systemPrompt := a.buildSystemPrompt(req)

	// Create the user message
	userMessage := fmt.Sprintf("Query: %s", req.Query)
	if req.ClusterInfo != "" {
		userMessage += fmt.Sprintf("\n\nCluster Information:\n%s", req.ClusterInfo)
	}

	// Call OpenAI API
	resp, err := a.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: a.cfg.Model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: userMessage,
				},
			},
			Temperature: 0.7,
			MaxTokens:   4000,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat completion: %w", err)
	}

	// Parse the response
	response := resp.Choices[0].Message.Content

	// Try to extract structured data from the response
	deploymentPlan, clusterAnalysis := a.extractStructuredData(response)

	return &QueryResponse{
		Response:        response,
		DeploymentPlan:  deploymentPlan,
		ClusterAnalysis: clusterAnalysis,
		Status:          "completed",
		Timestamp:       time.Now(),
	}, nil
}

// buildSystemPrompt creates a system prompt based on the query type
func (a *AIAgent) buildSystemPrompt(req *QueryRequest) string {
	basePrompt := `You are an expert Kubernetes and DevOps engineer AI assistant. Your role is to help users deploy and manage applications on Kubernetes clusters.

You have access to:
1. Cluster information (nodes, resources, capabilities)
2. Helm charts from various repositories
3. Best practices for Kubernetes deployments

When responding to queries:
1. Always provide clear, actionable advice
2. If a deployment plan is requested, create a detailed plan with:
   - Required Helm charts and versions
   - Customized values files
   - Step-by-step deployment instructions
   - Resource requirements and impact assessment
   - Prerequisites and risk assessment
3. Consider the cluster's capabilities and resources
4. Provide security best practices
5. Include troubleshooting tips

Format your responses in a clear, structured manner. If you're creating a deployment plan, structure it as JSON that can be parsed.`

	// Add specific context based on query type
	if strings.Contains(strings.ToLower(req.Query), "grafana") || strings.Contains(strings.ToLower(req.Query), "prometheus") {
		basePrompt += `

SPECIFIC INSTRUCTIONS FOR MONITORING STACKS:
- Recommend Prometheus Operator for production use
- Include Grafana dashboards and alerting rules
- Consider resource requirements for monitoring
- Include persistent storage configuration
- Provide ingress configuration for web access`
	}

	if strings.Contains(strings.ToLower(req.Query), "elk") || strings.Contains(strings.ToLower(req.Query), "logging") {
		basePrompt += `

SPECIFIC INSTRUCTIONS FOR LOGGING STACKS:
- Recommend Elasticsearch with proper resource limits
- Include Logstash or Fluentd for log collection
- Configure Kibana with security best practices
- Consider using Elasticsearch Operator for production
- Include persistent storage and backup strategies`
	}

	return basePrompt
}

// extractStructuredData attempts to extract structured data from AI response
func (a *AIAgent) extractStructuredData(response string) (*DeploymentPlan, *ClusterAnalysis) {
	// Look for JSON blocks in the response
	// This is a simple extraction - in production, you might want more sophisticated parsing

	// For now, return nil as we'll implement this in the deployment handler
	return nil, nil
}

// DeployStack executes a deployment plan
func (a *AIAgent) DeployStack(ctx context.Context, plan *DeploymentPlan) (*DeploymentExecution, error) {
	execution := &DeploymentExecution{
		ID:        fmt.Sprintf("exec-%d", time.Now().Unix()),
		PlanID:    plan.ID,
		Status:    "running",
		StartTime: time.Now(),
		Steps:     make([]DeploymentStepExecution, len(plan.Steps)),
	}

	// Initialize steps
	for i, step := range plan.Steps {
		execution.Steps[i] = DeploymentStepExecution{
			StepID:    step.ID,
			Status:    "pending",
			StartTime: nil,
			EndTime:   nil,
		}
	}

	// Execute steps sequentially
	for i := range execution.Steps {
		execution.Steps[i].Status = "running"
		execution.Steps[i].StartTime = &time.Time{}
		*execution.Steps[i].StartTime = time.Now()

		// Execute the step (this would integrate with actual Kubernetes operations)
		// For now, we'll simulate execution
		time.Sleep(2 * time.Second) // Simulate execution time

		execution.Steps[i].Status = "completed"
		execution.Steps[i].EndTime = &time.Time{}
		*execution.Steps[i].EndTime = time.Now()
	}

	execution.Status = "completed"
	execution.EndTime = &time.Time{}
	*execution.EndTime = time.Now()

	return execution, nil
}

// DeploymentExecution represents the execution of a deployment plan
type DeploymentExecution struct {
	ID        string                    `json:"id"`
	PlanID    string                    `json:"plan_id"`
	Status    string                    `json:"status"` // running, completed, failed, aborted
	StartTime time.Time                 `json:"start_time"`
	EndTime   *time.Time                `json:"end_time,omitempty"`
	Steps     []DeploymentStepExecution `json:"steps"`
	Logs      []string                  `json:"logs"`
	Error     string                    `json:"error,omitempty"`
}

// DeploymentStepExecution represents the execution of a deployment step
type DeploymentStepExecution struct {
	StepID    string     `json:"step_id"`
	Status    string     `json:"status"` // pending, running, completed, failed
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Logs      []string   `json:"logs"`
	Error     string     `json:"error,omitempty"`
}
