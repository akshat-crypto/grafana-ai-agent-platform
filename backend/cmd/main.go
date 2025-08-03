package main

import (
	"fmt"
	"log"
	"net/http"

	"grafana-ai-agent-platform/backend/internal/agent"
	"grafana-ai-agent-platform/backend/internal/config"
	"grafana-ai-agent-platform/backend/internal/handlers"
	"grafana-ai-agent-platform/backend/internal/middleware"
	"grafana-ai-agent-platform/backend/pkg/database"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database
	db, err := database.NewDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize AI agent
	aiAgent := agent.NewAIAgent(cfg)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db, cfg)
	kubernetesHandler := handlers.NewKubernetesHandler(db)
	agentHandler := handlers.NewAgentHandler(db, aiAgent)

	// Setup Gin router
	router := gin.Default()

	// Add CORS middleware
	router.Use(middleware.CORSMiddleware())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "kubernetes-ai-agent-platform",
		})
	})

	// API routes
	api := router.Group("/api")
	{
		// Authentication routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", authHandler.Logout)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
		{
			// User profile
			protected.GET("/profile", authHandler.GetProfile)

			// Kubernetes routes
			kubernetes := protected.Group("/kubernetes")
			{
				kubernetes.POST("/validate", kubernetesHandler.ValidateCluster)
				kubernetes.POST("/clusters", kubernetesHandler.AddCluster)
				kubernetes.GET("/clusters", kubernetesHandler.GetClusters)
				kubernetes.DELETE("/clusters/:id", kubernetesHandler.DeleteCluster)
				kubernetes.GET("/clusters/:id/resources", kubernetesHandler.GetClusterResources)
			}

			// AI Agent routes
			agent := protected.Group("/agent")
			{
				agent.POST("/query", agentHandler.QueryAgent)
				agent.POST("/deploy", agentHandler.DeployStack)
				agent.GET("/queries", agentHandler.GetQueryHistory)
				agent.GET("/deployments", agentHandler.GetDeploymentHistory)
			}
		}
	}

	// Start server
	serverAddr := fmt.Sprintf("0.0.0.0:%s", cfg.Server.Port)
	log.Printf("Server starting on %s", serverAddr)

	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
