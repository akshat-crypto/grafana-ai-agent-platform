package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"grafana-ai-agent-platform/backend/internal/config"
)

type AIAgent struct {
	config *config.Config
}

func NewAIAgent(cfg *config.Config) *AIAgent {
	return &AIAgent{config: cfg}
}

type OpenRouterRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenRouterResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (a *AIAgent) QueryResponse(query string) (string, error) {
	// Enhanced prompt for Kubernetes operations
	enhancedQuery := a.enhanceKubernetesPrompt(query)

	payload := OpenRouterRequest{
		Model: "openai/gpt-3.5-turbo",
		Messages: []Message{
			{
				Role: "system",
				Content: `You are a Kubernetes expert and DevOps engineer. Your role is to help users deploy and manage applications on Kubernetes clusters. 
				Provide clear, actionable responses with proper YAML manifests when needed. Always include:
				1. Clear explanation of what will be deployed
				2. Proper YAML manifests
				3. Commands to apply the manifests
				4. Verification steps
				5. Troubleshooting tips if applicable`,
			},
			{
				Role:    "user",
				Content: enhancedQuery,
			},
		},
		MaxTokens:   2000,
		Temperature: 0.7,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+a.config.OpenRouter.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var result OpenRouterResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Check for API errors
	if result.Error != nil {
		return "", fmt.Errorf("API error: %s", result.Error.Message)
	}

	// Check for empty response
	if len(result.Choices) == 0 || result.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("empty response from AI model")
	}

	return result.Choices[0].Message.Content, nil
}

func (a *AIAgent) DebugCode(code, errorMsg string) (string, error) {
	prompt := fmt.Sprintf(`Debug this Kubernetes manifest or command:

Code/Manifest:
%s

Error:
%s

Please provide:
1. Root cause analysis
2. Step-by-step resolution
3. Corrected code/manifest if applicable
4. Prevention tips`, code, errorMsg)

	return a.QueryResponse(prompt)
}

func (a *AIAgent) DeployStack(stackName string, clusterInfo string) (string, error) {
	prompt := fmt.Sprintf(`Deploy %s stack on Kubernetes cluster with the following information:

Cluster Info:
%s

Please provide:
1. Complete YAML manifests for %s
2. Step-by-step deployment instructions
3. Configuration options
4. Verification commands
5. Access information (URLs, ports, etc.)
6. Basic troubleshooting guide

Make sure the manifests are production-ready with proper resource limits, security contexts, and best practices.`, stackName, clusterInfo, stackName)

	return a.QueryResponse(prompt)
}

func (a *AIAgent) enhanceKubernetesPrompt(query string) string {
	// Add context about common Kubernetes operations
	enhanced := query

	// Detect common patterns and enhance them
	if strings.Contains(strings.ToLower(query), "grafana") {
		enhanced = "Deploy Grafana monitoring stack with Prometheus, AlertManager, and Node Exporter. " + enhanced
	}

	if strings.Contains(strings.ToLower(query), "elk") || strings.Contains(strings.ToLower(query), "elasticsearch") {
		enhanced = "Deploy ELK (Elasticsearch, Logstash, Kibana) stack for centralized logging. " + enhanced
	}

	if strings.Contains(strings.ToLower(query), "prometheus") {
		enhanced = "Deploy Prometheus monitoring stack with Grafana for visualization. " + enhanced
	}

	if strings.Contains(strings.ToLower(query), "install") || strings.Contains(strings.ToLower(query), "deploy") {
		enhanced = enhanced + " Please provide complete YAML manifests and deployment instructions."
	}

	return enhanced
}
