package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"grafana-ai-agent-platform/backend/internal/agent"
	"grafana-ai-agent-platform/backend/internal/services"
	"grafana-ai-agent-platform/backend/pkg/database"

	"github.com/gin-gonic/gin"
)

// AgentHandler handles AI agent operations
type AgentHandler struct {
	db                 *database.Database
	aiAgent            *agent.AIAgent
	clusterAnalyzer    *services.ClusterAnalyzerService
	helmService        *services.HelmService
	deploymentExecutor *services.DeploymentExecutorService
}

// NewAgentHandler creates a new agent handler
func NewAgentHandler(db *database.Database, aiAgent *agent.AIAgent) *AgentHandler {
	helmService := services.NewHelmService()
	deploymentExecutor := services.NewDeploymentExecutorService(helmService)
	clusterAnalyzer := services.NewClusterAnalyzerService()

	return &AgentHandler{
		db:                 db,
		aiAgent:            aiAgent,
		clusterAnalyzer:    clusterAnalyzer,
		helmService:        helmService,
		deploymentExecutor: deploymentExecutor,
	}
}

// QueryRequest represents a user query to the AI agent
type QueryRequest struct {
	Query     string `json:"query" binding:"required"`
	ClusterID *uint  `json:"cluster_id,omitempty"`
}

// QueryResponse represents the AI agent response
type QueryResponse struct {
	Response        string                 `json:"response"`
	DeploymentPlan  *agent.DeploymentPlan  `json:"deployment_plan,omitempty"`
	ClusterAnalysis *agent.ClusterAnalysis `json:"cluster_analysis,omitempty"`
	Status          string                 `json:"status"`
	Timestamp       string                 `json:"timestamp"`
}

// DeployRequest represents a deployment request
type DeployRequest struct {
	PlanID     string `json:"plan_id" binding:"required"`
	ClusterID  uint   `json:"cluster_id" binding:"required"`
	KubeConfig string `json:"kube_config" binding:"required"`
}

// DeployResponse represents a deployment response
type DeployResponse struct {
	ExecutionID string `json:"execution_id"`
	Status      string `json:"status"`
	Message     string `json:"message"`
}

// QueryAgent handles AI agent queries
func (h *AgentHandler) QueryAgent(c *gin.Context) {
	var req QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get cluster information if cluster ID is provided
	var clusterInfo string
	if req.ClusterID != nil {
		cluster, err := h.getClusterInfo(*req.ClusterID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to get cluster info: %v", err)})
			return
		}
		clusterInfo = cluster
	}

	// Create AI agent request
	aiReq := &agent.QueryRequest{
		Query:       req.Query,
		ClusterID:   req.ClusterID,
		ClusterInfo: clusterInfo,
	}

	// Query the AI agent
	ctx := context.Background()
	aiResp, err := h.aiAgent.Query(ctx, aiReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("AI agent query failed: %v", err)})
		return
	}

	// If this is a deployment request, create a deployment plan
	var deploymentPlan *agent.DeploymentPlan
	if h.isDeploymentQuery(req.Query) {
		plan, err := h.createDeploymentPlan(req.Query, req.ClusterID, clusterInfo)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create deployment plan: %v", err)})
			return
		}
		deploymentPlan = plan
	}

	// Create response
	response := QueryResponse{
		Response:        aiResp.Response,
		DeploymentPlan:  deploymentPlan,
		ClusterAnalysis: aiResp.ClusterAnalysis,
		Status:          aiResp.Status,
		Timestamp:       aiResp.Timestamp.Format("2006-01-02T15:04:05Z"),
	}

	// Save query to database
	h.saveQuery(c, req, response)

	c.JSON(http.StatusOK, response)
}

// DeployStack handles stack deployment requests
func (h *AgentHandler) DeployStack(c *gin.Context) {
	var req DeployRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get the deployment plan (in production, this would come from storage)
	plan, err := h.getDeploymentPlan(req.PlanID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Deployment plan not found: %v", err)})
		return
	}

	// Execute the deployment
	ctx := context.Background()
	execution, err := h.deploymentExecutor.ExecuteDeployment(ctx, plan, req.KubeConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Deployment execution failed: %v", err)})
		return
	}

	// Save deployment to database
	h.saveDeployment(c, req, execution)

	response := DeployResponse{
		ExecutionID: execution.ID,
		Status:      execution.Status,
		Message:     "Deployment started successfully",
	}

	c.JSON(http.StatusOK, response)
}

// GetQueryHistory returns the history of AI agent queries
func (h *AgentHandler) GetQueryHistory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get queries from database (implement this based on your database schema)
	// TODO: Implement actual database query using userID
	queries := []map[string]interface{}{
		{
			"id":        1,
			"user_id":   userID,
			"query":     "Install Grafana and Prometheus stack",
			"response":  "AI response here",
			"timestamp": "2025-08-17T09:00:00Z",
		},
	}

	c.JSON(http.StatusOK, queries)
}

// GetDeploymentHistory returns the history of deployments
func (h *AgentHandler) GetDeploymentHistory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get deployments from database (implement this based on your database schema)
	// TODO: Implement actual database query using userID
	deployments := []map[string]interface{}{
		{
			"id":         1,
			"user_id":    userID,
			"plan_id":    "plan-1",
			"status":     "completed",
			"start_time": "2025-08-17T09:00:00Z",
			"end_time":   "2025-08-17T09:15:00Z",
		},
	}

	c.JSON(http.StatusOK, deployments)
}

// Helper methods

// isDeploymentQuery checks if a query is requesting a deployment
func (h *AgentHandler) isDeploymentQuery(query string) bool {
	deploymentKeywords := []string{
		"install", "deploy", "setup", "create", "add", "enable",
		"grafana", "prometheus", "elk", "elasticsearch", "kibana",
		"monitoring", "logging", "observability",
	}

	queryLower := strings.ToLower(query)
	for _, keyword := range deploymentKeywords {
		if strings.Contains(queryLower, keyword) {
			return true
		}
	}
	return false
}

// createDeploymentPlan creates a deployment plan for the given query
func (h *AgentHandler) createDeploymentPlan(query string, clusterID *uint, clusterInfo string) (*agent.DeploymentPlan, error) {
	// Analyze cluster if cluster ID is provided
	var clusterAnalysis *agent.ClusterAnalysis
	if clusterID != nil && clusterInfo != "" {
		// Parse cluster info and create analysis
		// This is a simplified version - in production, you'd want more sophisticated parsing
		clusterAnalysis = &agent.ClusterAnalysis{
			ClusterName: fmt.Sprintf("cluster-%d", *clusterID),
			Version:     "v1.28.0",
			Capabilities: agent.ClusterCapabilities{
				HelmInstalled:    true,
				IngressAvailable: true,
				LoadBalancer:     true,
				PersistentVolume: true,
				RBACEnabled:      true,
				NetworkPolicy:    true,
			},
		}
	}

	// Create deployment plan using Helm service
	plan, err := h.helmService.CreateDeploymentPlan(query, clusterAnalysis)
	if err != nil {
		return nil, fmt.Errorf("failed to create deployment plan: %w", err)
	}

	return plan, nil
}

// getDeploymentPlan retrieves a deployment plan (placeholder implementation)
func (h *AgentHandler) getDeploymentPlan(planID string) (*agent.DeploymentPlan, error) {
	// In production, this would retrieve the plan from storage
	// For now, return a placeholder plan
	return &agent.DeploymentPlan{
		ID:          planID,
		Name:        "Sample Deployment Plan",
		Description: "A sample deployment plan",
		Charts: []agent.HelmChart{
			{
				Name:        "prometheus",
				Repository:  "prometheus-community",
				Version:     "25.0.0",
				Description: "Prometheus monitoring stack",
			},
		},
		Steps: []agent.DeploymentStep{
			{
				ID:          "step-1",
				Name:        "Deploy Prometheus",
				Description: "Deploy Prometheus monitoring stack",
				Status:      "pending",
			},
		},
	}, nil
}

// getClusterInfo retrieves cluster information
func (h *AgentHandler) getClusterInfo(clusterID uint) (string, error) {
	// In production, this would retrieve cluster info from the database
	// For now, return placeholder info
	return fmt.Sprintf("Cluster ID: %d\nVersion: v1.28.0\nNodes: 3\nResources: Available", clusterID), nil
}

// saveQuery saves a query to the database
func (h *AgentHandler) saveQuery(c *gin.Context, req QueryRequest, resp QueryResponse) {
	// Implement database save logic here
	// This would save the query and response for history tracking
}

// saveDeployment saves a deployment to the database
func (h *AgentHandler) saveDeployment(c *gin.Context, req DeployRequest, execution *agent.DeploymentExecution) {
	// Implement database save logic here
	// This would save the deployment execution for history tracking
}
