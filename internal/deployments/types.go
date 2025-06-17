package deployments

// DeploymentResult represents the result of a deployment operation
type DeploymentResult struct {
	Application  string                   `json:"application"`
	Environment  string                   `json:"environment"`
	DeploymentID string                   `json:"deployment_id"`
	Deployments  []string                 `json:"deployments"`
	Skipped      []string                 `json:"skipped"`
	Failed       []map[string]interface{} `json:"failed"`
	Summary      DeploymentSummary        `json:"summary"`
	Status       string                   `json:"status"` // "initiated", "in_progress", "completed", "failed"
}

// DeploymentSummary provides a high-level summary of the deployment
type DeploymentSummary struct {
	TotalServices int    `json:"total_services"`
	Deployed      int    `json:"deployed"`
	Skipped       int    `json:"skipped"`
	Failed        int    `json:"failed"`
	Success       bool   `json:"success"`
	Message       string `json:"message"`
}
