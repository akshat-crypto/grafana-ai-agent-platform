package handlers

import (
	"fmt"
	"net/http"

	"grafana-ai-agent-platform/backend/internal/agent"
	"grafana-ai-agent-platform/backend/internal/models"
	"grafana-ai-agent-platform/backend/pkg/database"

	"github.com/gin-gonic/gin"
)

type AgentHandler struct {
	db    *database.Database
	agent *agent.AIAgent
}

func NewAgentHandler(db *database.Database, aiAgent *agent.AIAgent) *AgentHandler {
	return &AgentHandler{
		db:    db,
		agent: aiAgent,
	}
}

func (h *AgentHandler) QueryAgent(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.AgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create agent query record
	agentQuery := models.AgentQuery{
		UserID:    userID.(uint),
		ClusterID: req.ClusterID,
		Query:     req.Query,
		Status:    "processing",
	}

	if err := h.db.DB.Create(&agentQuery).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create query record"})
		return
	}

	// Get cluster info if clusterID is provided
	if req.ClusterID > 0 {
		var cluster models.KubernetesCluster
		if err := h.db.DB.Where("id = ? AND user_id = ?", req.ClusterID, userID).First(&cluster).Error; err == nil {
			// Cluster info available for context
		}
	}

	// Process query with AI agent
	response, err := h.agent.QueryResponse(req.Query)
	if err != nil {
		// Update query record with error
		h.db.DB.Model(&agentQuery).Updates(map[string]interface{}{
			"response": "Error: " + err.Error(),
			"status":   "error",
		})

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to process query",
			"details": err.Error(),
		})
		return
	}

	// Update query record with response
	h.db.DB.Model(&agentQuery).Updates(map[string]interface{}{
		"response": response,
		"status":   "completed",
	})

	c.JSON(http.StatusOK, models.AgentResponse{
		Response: response,
		Status:   "completed",
	})
}

func (h *AgentHandler) DeployStack(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req struct {
		StackName string `json:"stack_name" binding:"required"`
		ClusterID uint   `json:"cluster_id" binding:"required"`
		Query     string `json:"query" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get cluster info
	var cluster models.KubernetesCluster
	if err := h.db.DB.Where("id = ? AND user_id = ?", req.ClusterID, userID).First(&cluster).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cluster not found"})
		return
	}

	// Create deployment record
	deployment := models.Deployment{
		UserID:    userID.(uint),
		ClusterID: req.ClusterID,
		StackName: req.StackName,
		Status:    "processing",
	}

	if err := h.db.DB.Create(&deployment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create deployment record"})
		return
	}

	// Prepare cluster info for AI
	clusterInfo := fmt.Sprintf("Cluster: %s (v%s)\nServer: %s",
		cluster.Name, cluster.Version, cluster.ClusterURL)

	// Get deployment instructions from AI
	manifest, err := h.agent.DeployStack(req.StackName, clusterInfo)
	if err != nil {
		// Update deployment record with error
		h.db.DB.Model(&deployment).Updates(map[string]interface{}{
			"error":  "Error: " + err.Error(),
			"status": "error",
		})

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate deployment manifest",
			"details": err.Error(),
		})
		return
	}

	// Update deployment record with manifest
	h.db.DB.Model(&deployment).Updates(map[string]interface{}{
		"manifest": manifest,
		"status":   "ready",
	})

	c.JSON(http.StatusOK, gin.H{
		"message":       "Deployment manifest generated successfully",
		"manifest":      manifest,
		"deployment_id": deployment.ID,
	})
}

func (h *AgentHandler) GetQueryHistory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var queries []models.AgentQuery
	if err := h.db.DB.Where("user_id = ?", userID).Order("created_at DESC").Limit(50).Find(&queries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch query history"})
		return
	}

	c.JSON(http.StatusOK, queries)
}

func (h *AgentHandler) GetDeploymentHistory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var deployments []models.Deployment
	if err := h.db.DB.Where("user_id = ?", userID).Order("created_at DESC").Limit(50).Find(&deployments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch deployment history"})
		return
	}

	c.JSON(http.StatusOK, deployments)
}
