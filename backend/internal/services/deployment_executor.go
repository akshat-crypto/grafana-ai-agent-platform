package services

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"grafana-ai-agent-platform/backend/internal/agent"
)

// DeploymentExecutorService handles the execution of deployment plans
type DeploymentExecutorService struct {
	helmService *HelmService
}

// NewDeploymentExecutorService creates a new deployment executor service
func NewDeploymentExecutorService(helmService *HelmService) *DeploymentExecutorService {
	return &DeploymentExecutorService{
		helmService: helmService,
	}
}

// ExecuteDeployment executes a deployment plan
func (s *DeploymentExecutorService) ExecuteDeployment(ctx context.Context, plan *agent.DeploymentPlan, kubeconfig string) (*agent.DeploymentExecution, error) {
	execution := &agent.DeploymentExecution{
		ID:        fmt.Sprintf("exec-%d", time.Now().Unix()),
		PlanID:    plan.ID,
		Status:    "running",
		StartTime: time.Now(),
		Steps:     make([]agent.DeploymentStepExecution, len(plan.Steps)),
		Logs:      []string{fmt.Sprintf("Starting deployment of %s", plan.Name)},
	}

	// Initialize steps
	for i, step := range plan.Steps {
		execution.Steps[i] = agent.DeploymentStepExecution{
			StepID:    step.ID,
			Status:    "pending",
			StartTime: nil,
			EndTime:   nil,
			Logs:      []string{},
		}
	}

	// Execute steps sequentially
	for i := range execution.Steps {
		execution.Steps[i].Status = "running"
		execution.Steps[i].StartTime = &time.Time{}
		*execution.Steps[i].StartTime = time.Now()

		// Add log entry
		execution.Logs = append(execution.Logs, fmt.Sprintf("Executing step %d: %s", i+1, execution.Steps[i].StepID))

		// Execute the step
		err := s.executeStep(ctx, &execution.Steps[i], plan.Steps[i], kubeconfig)

		if err != nil {
			execution.Steps[i].Status = "failed"
			execution.Steps[i].Error = err.Error()
			execution.Logs = append(execution.Logs, fmt.Sprintf("Step %d failed: %v", i+1, err))
			execution.Status = "failed"
			execution.Error = fmt.Sprintf("Step %d failed: %v", i+1, err)
			return execution, nil
		}

		execution.Steps[i].Status = "completed"
		execution.Steps[i].EndTime = &time.Time{}
		*execution.Steps[i].EndTime = time.Now()

		execution.Logs = append(execution.Logs, fmt.Sprintf("Step %d completed successfully", i+1))
	}

	execution.Status = "completed"
	execution.EndTime = &time.Time{}
	*execution.EndTime = time.Now()
	execution.Logs = append(execution.Logs, "Deployment completed successfully")

	return execution, nil
}

// executeStep executes a single deployment step
func (s *DeploymentExecutorService) executeStep(ctx context.Context, stepExec *agent.DeploymentStepExecution, step agent.DeploymentStep, kubeconfig string) error {
	// Add step start log
	stepExec.Logs = append(stepExec.Logs, fmt.Sprintf("Starting: %s", step.Description))

	// Check if Helm is installed
	if err := s.ensureHelmInstalled(); err != nil {
		stepExec.Logs = append(stepExec.Logs, fmt.Sprintf("Helm installation check failed: %v", err))
		return fmt.Errorf("helm not available: %w", err)
	}

	// Add Helm repository if needed
	if step.Chart != nil {
		if err := s.addHelmRepository(step.Chart.Repository); err != nil {
			stepExec.Logs = append(stepExec.Logs, fmt.Sprintf("Failed to add repository: %v", err))
			return fmt.Errorf("failed to add helm repository: %w", err)
		}
		stepExec.Logs = append(stepExec.Logs, fmt.Sprintf("Added repository: %s", step.Chart.Repository))
	}

	// Execute the deployment command
	if step.Command != "" {
		if err := s.executeCommand(ctx, step.Command, stepExec); err != nil {
			return fmt.Errorf("command execution failed: %w", err)
		}
	} else if step.Chart != nil {
		// Deploy using Helm
		if err := s.deployHelmChart(ctx, step.Chart, kubeconfig, stepExec); err != nil {
			return fmt.Errorf("helm deployment failed: %w", err)
		}
	}

	stepExec.Logs = append(stepExec.Logs, fmt.Sprintf("Completed: %s", step.Description))
	return nil
}

// ensureHelmInstalled checks if Helm is installed and installs it if needed
func (s *DeploymentExecutorService) ensureHelmInstalled() error {
	// Check if helm command is available
	if _, err := exec.LookPath("helm"); err == nil {
		return nil
	}

	// Helm not found, try to install it
	return s.installHelm()
}

// installHelm installs Helm using the official installation script
func (s *DeploymentExecutorService) installHelm() error {
	// Download and install Helm
	installCmd := exec.Command("curl", "https://get.helm.sh/helm-v3.15.0-linux-amd64.tar.gz", "-o", "/tmp/helm.tar.gz")
	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("failed to download helm: %w", err)
	}

	// Extract and install
	extractCmd := exec.Command("tar", "-xzf", "/tmp/helm.tar.gz", "-C", "/tmp")
	if err := extractCmd.Run(); err != nil {
		return fmt.Errorf("failed to extract helm: %w", err)
	}

	moveCmd := exec.Command("sudo", "mv", "/tmp/linux-amd64/helm", "/usr/local/bin/helm")
	if err := moveCmd.Run(); err != nil {
		return fmt.Errorf("failed to move helm: %w", err)
	}

	return nil
}

// addHelmRepository adds a Helm repository
func (s *DeploymentExecutorService) addHelmRepository(repoURL string) error {
	// Check if repository already exists
	checkCmd := exec.Command("helm", "repo", "list")
	output, err := checkCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check helm repos: %w", err)
	}

	if strings.Contains(string(output), repoURL) {
		return nil // Repository already exists
	}

	// Add repository
	repoName := s.extractRepoName(repoURL)
	addCmd := exec.Command("helm", "repo", "add", repoName, repoURL)
	if err := addCmd.Run(); err != nil {
		return fmt.Errorf("failed to add helm repository: %w", err)
	}

	// Update repositories
	updateCmd := exec.Command("helm", "repo", "update")
	if err := updateCmd.Run(); err != nil {
		return fmt.Errorf("failed to update helm repos: %w", err)
	}

	return nil
}

// extractRepoName extracts a repository name from URL
func (s *DeploymentExecutorService) extractRepoName(repoURL string) string {
	// Simple extraction - in production, you might want more sophisticated logic
	if strings.Contains(repoURL, "github.com") {
		parts := strings.Split(repoURL, "/")
		if len(parts) >= 3 {
			return parts[len(parts)-1]
		}
	}
	return "repo"
}

// deployHelmChart deploys a Helm chart
func (s *DeploymentExecutorService) deployHelmChart(ctx context.Context, chart *agent.HelmChart, kubeconfig string, stepExec *agent.DeploymentStepExecution) error {
	// Create temporary values file
	valuesFile, err := s.createValuesFile(chart.Values)
	if err != nil {
		return fmt.Errorf("failed to create values file: %w", err)
	}
	defer s.cleanupValuesFile(valuesFile)

	// Set KUBECONFIG environment variable
	env := []string{fmt.Sprintf("KUBECONFIG=%s", kubeconfig)}

	// Execute helm install command
	installCmd := exec.CommandContext(ctx, "helm", "install", chart.Name, chart.Repository+"/"+chart.Name,
		"--values", valuesFile, "--wait", "--timeout", "10m")
	installCmd.Env = env

	stepExec.Logs = append(stepExec.Logs, fmt.Sprintf("Installing chart: %s from %s", chart.Name, chart.Repository))

	output, err := installCmd.CombinedOutput()
	if err != nil {
		stepExec.Logs = append(stepExec.Logs, fmt.Sprintf("Helm install failed: %v", string(output)))
		return fmt.Errorf("helm install failed: %w", err)
	}

	stepExec.Logs = append(stepExec.Logs, fmt.Sprintf("Chart installed successfully: %s", string(output)))
	return nil
}

// executeCommand executes a shell command
func (s *DeploymentExecutorService) executeCommand(ctx context.Context, command string, stepExec *agent.DeploymentStepExecution) error {
	stepExec.Logs = append(stepExec.Logs, fmt.Sprintf("Executing command: %s", command))

	// Split command into parts
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		stepExec.Logs = append(stepExec.Logs, fmt.Sprintf("Command failed: %v", string(output)))
		return fmt.Errorf("command execution failed: %w", err)
	}

	stepExec.Logs = append(stepExec.Logs, fmt.Sprintf("Command output: %s", string(output)))
	return nil
}

// createValuesFile creates a temporary values file
func (s *DeploymentExecutorService) createValuesFile(values map[string]interface{}) (string, error) {
	// For now, create a simple values file
	// In production, you'd want to use a proper YAML library
	content := "# Generated values file\n"

	// Add some basic values
	if values != nil {
		for key, value := range values {
			content += fmt.Sprintf("%s: %v\n", key, value)
		}
	}

	// Create temporary file
	filename := fmt.Sprintf("/tmp/values-%d.yaml", time.Now().Unix())
	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write values file: %w", err)
	}

	return filename, nil
}

// cleanupValuesFile removes the temporary values file
func (s *DeploymentExecutorService) cleanupValuesFile(filename string) {
	os.Remove(filename)
}

// AbortDeployment aborts a running deployment
func (s *DeploymentExecutorService) AbortDeployment(ctx context.Context, executionID string) error {
	// This would implement deployment abortion logic
	// For now, we'll just return success
	return nil
}

// GetDeploymentStatus gets the current status of a deployment
func (s *DeploymentExecutorService) GetDeploymentStatus(executionID string) (*agent.DeploymentExecution, error) {
	// This would retrieve deployment status from storage
	// For now, return nil
	return nil, nil
}
