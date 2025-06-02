package deployments

import (
	"fmt"
	"time"

	"github.com/krzachariassen/ZTDP/internal/contracts"
)

// DeploymentMetadataKey is the key used to store deployment information in edge metadata
const DeploymentMetadataKey = "deployment"

// SetDeploymentStatus sets the deployment status in edge metadata
func SetDeploymentStatus(metadata map[string]interface{}, status contracts.DeploymentStatus, message string) error {
	// Validate the status
	if !status.IsValid() {
		return fmt.Errorf("invalid deployment status: %s", status)
	}

	// Validate status transition if there's existing status
	if err := ValidateStatusTransition(metadata, status); err != nil {
		return err
	}

	// Get or create the deployment metadata
	var deploymentMeta map[string]interface{}
	if existing, exists := metadata[DeploymentMetadataKey]; exists {
		if existingMap, ok := existing.(map[string]interface{}); ok {
			deploymentMeta = existingMap
		} else {
			deploymentMeta = make(map[string]interface{})
		}
	} else {
		deploymentMeta = make(map[string]interface{})
	}

	// Set the status information
	deploymentMeta["status"] = string(status)
	deploymentMeta["message"] = message
	deploymentMeta["last_updated"] = time.Now().Format(time.RFC3339Nano)

	// Store back in metadata
	metadata[DeploymentMetadataKey] = deploymentMeta

	return nil
}

// GetDeploymentStatus gets the deployment status from edge metadata
func GetDeploymentStatus(metadata map[string]interface{}) (contracts.DeploymentStatus, string, bool) {
	deploymentMeta, exists := metadata[DeploymentMetadataKey]
	if !exists {
		return "", "", false
	}

	deploymentMap, ok := deploymentMeta.(map[string]interface{})
	if !ok {
		return "", "", false
	}

	statusStr, exists := deploymentMap["status"]
	if !exists {
		return "", "", false
	}

	status, ok := statusStr.(string)
	if !ok {
		return "", "", false
	}

	message, _ := deploymentMap["message"].(string)

	return contracts.DeploymentStatus(status), message, true
}

// ValidateStatusTransition validates that a status transition is allowed
func ValidateStatusTransition(metadata map[string]interface{}, newStatus contracts.DeploymentStatus) error {
	currentStatus, _, exists := GetDeploymentStatus(metadata)
	if !exists {
		// No existing status, any valid status is allowed
		return nil
	}

	// Use the contract's transition validation
	if !currentStatus.CanTransitionTo(newStatus) {
		return fmt.Errorf("invalid status transition from %s to %s", currentStatus, newStatus)
	}

	return nil
}
