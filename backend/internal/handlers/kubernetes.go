package handlers

import (
	"fmt"
	"net/http"

	"grafana-ai-agent-platform/backend/internal/models"
	"grafana-ai-agent-platform/backend/pkg/database"
	"grafana-ai-agent-platform/backend/pkg/kubernetes"

	"github.com/gin-gonic/gin"
)

type KubernetesHandler struct {
	db *database.Database
}

func NewKubernetesHandler(db *database.Database) *KubernetesHandler {
	return &KubernetesHandler{
		db: db,
	}
}

type AddClusterRequest struct {
	Name       string `json:"name" binding:"required"`
	KubeConfig string `json:"kube_config" binding:"required"`
}

type ValidateClusterRequest struct {
	KubeConfig string `json:"kube_config" binding:"required"`
}

func (h *KubernetesHandler) ValidateCluster(c *gin.Context) {
	var req ValidateClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log the request for debugging
	fmt.Printf("Validating kubeconfig for user, length: %d\n", len(req.KubeConfig))

	// Basic validation
	if req.KubeConfig == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"is_valid": false,
			"error":    "Kubeconfig is empty",
		})
		return
	}

	// Validate kubeconfig format first
	if err := kubernetes.ValidateKubeconfigFormat(req.KubeConfig); err != nil {
		fmt.Printf("Kubeconfig format validation failed: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"is_valid": false,
			"error":    fmt.Sprintf("Invalid kubeconfig format: %v", err),
		})
		return
	}

	// Create Kubernetes client
	client, err := kubernetes.NewKubernetesClient(req.KubeConfig)
	if err != nil {
		fmt.Printf("Failed to create Kubernetes client: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"is_valid": false,
			"error":    fmt.Sprintf("Failed to create Kubernetes client: %v", err),
		})
		return
	}

	// Validate cluster connection
	clusterInfo, err := client.ValidateCluster()
	if err != nil {
		fmt.Printf("Failed to validate cluster: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"is_valid": false,
			"error":    fmt.Sprintf("Failed to validate cluster: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, clusterInfo)
}

func (h *KubernetesHandler) AddCluster(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req AddClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate kubeconfig format first
	if err := kubernetes.ValidateKubeconfigFormat(req.KubeConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid kubeconfig format: %v", err),
		})
		return
	}

	// Try to create Kubernetes client and validate cluster
	var clusterInfo *kubernetes.ClusterInfo
	var clusterURL string
	var status string
	var isActive bool
	var version string

	client, err := kubernetes.NewKubernetesClient(req.KubeConfig)
	if err != nil {
		// Cluster creation failed, but we'll save it as inactive
		status = "inactive"
		isActive = false
		version = "unknown"
		clusterURL = "unknown"
	} else {
		// Try to validate the cluster
		clusterInfo, err = client.ValidateCluster()
		if err != nil {
			// Cluster validation failed, mark as inactive
			status = "inactive"
			isActive = false
			version = "unknown"
			clusterURL = "unknown"
		} else {
			// Cluster is working
			status = "active"
			isActive = true
			version = clusterInfo.Version
			clusterURL = clusterInfo.ServerURL
		}
	}

	// Create cluster record
	cluster := models.KubernetesCluster{
		UserID:     userID.(uint),
		Name:       req.Name,
		KubeConfig: req.KubeConfig,
		ClusterURL: clusterURL,
		Version:    version,
		Status:     status,
		IsActive:   isActive,
	}

	if err := h.db.DB.Create(&cluster).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save cluster"})
		return
	}

	// Return appropriate response based on cluster status
	if isActive {
		c.JSON(http.StatusCreated, gin.H{
			"message": "Cluster added successfully",
			"cluster": cluster,
		})
	} else {
		c.JSON(http.StatusCreated, gin.H{
			"message": "Cluster added but marked as inactive due to connection issues",
			"cluster": cluster,
			"warning": "Cluster could not be reached. Use the refresh button to retry connection.",
		})
	}
}

func (h *KubernetesHandler) GetClusters(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var clusters []models.KubernetesCluster
	if err := h.db.DB.Where("user_id = ?", userID).Find(&clusters).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch clusters"})
		return
	}

	// Don't return kubeconfig in response for security
	var safeClusters []models.ClusterStatus
	for _, cluster := range clusters {
		safeClusters = append(safeClusters, models.ClusterStatus{
			ID:       cluster.ID,
			Name:     cluster.Name,
			Status:   cluster.Status,
			IsActive: cluster.IsActive,
			Version:  cluster.Version,
		})
	}

	c.JSON(http.StatusOK, safeClusters)
}

func (h *KubernetesHandler) DeleteCluster(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	clusterID := c.Param("id")
	if clusterID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cluster ID required"})
		return
	}

	// Delete cluster (soft delete)
	if err := h.db.DB.Where("id = ? AND user_id = ?", clusterID, userID).Delete(&models.KubernetesCluster{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete cluster"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cluster deleted successfully"})
}

func (h *KubernetesHandler) GetClusterResources(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	clusterID := c.Param("id")
	if clusterID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cluster ID required"})
		return
	}

	// Get cluster
	var cluster models.KubernetesCluster
	if err := h.db.DB.Where("id = ? AND user_id = ?", clusterID, userID).First(&cluster).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cluster not found"})
		return
	}

	// Create Kubernetes client
	client, err := kubernetes.NewKubernetesClient(cluster.KubeConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to cluster"})
		return
	}

	// Get cluster resources
	resources, err := client.GetClusterResources()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get cluster resources"})
		return
	}

	c.JSON(http.StatusOK, resources)
}

func (h *KubernetesHandler) RefreshClusterStatus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	clusterID := c.Param("id")
	if clusterID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cluster ID required"})
		return
	}

	// Get cluster
	var cluster models.KubernetesCluster
	if err := h.db.DB.Where("id = ? AND user_id = ?", clusterID, userID).First(&cluster).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cluster not found"})
		return
	}

	// Test cluster connectivity
	client, err := kubernetes.NewKubernetesClient(cluster.KubeConfig)
	if err != nil {
		// Update cluster status to inactive
		h.db.DB.Model(&cluster).Updates(map[string]interface{}{
			"status":    "inactive",
			"is_active": false,
		})
		c.JSON(http.StatusOK, gin.H{
			"message":   "Cluster status updated",
			"status":    "inactive",
			"is_active": false,
			"error":     err.Error(),
		})
		return
	}

	// Test cluster connection
	clusterInfo, err := client.ValidateCluster()
	if err != nil {
		// Update cluster status to inactive
		h.db.DB.Model(&cluster).Updates(map[string]interface{}{
			"status":    "inactive",
			"is_active": false,
		})
		c.JSON(http.StatusOK, gin.H{
			"message":   "Cluster status updated",
			"status":    "inactive",
			"is_active": false,
			"error":     err.Error(),
		})
		return
	}

	// Update cluster status to active
	h.db.DB.Model(&cluster).Updates(map[string]interface{}{
		"status":    "active",
		"is_active": true,
		"version":   clusterInfo.Version,
	})

	c.JSON(http.StatusOK, gin.H{
		"message":   "Cluster status updated",
		"status":    "active",
		"is_active": true,
		"version":   clusterInfo.Version,
	})
}
