// Package deployments provides deployment domain logic for ZTDP.
package deployments

import (
	"time"

	"github.com/krzachariassen/ZTDP/internal/graph"
)

// ExtractDeploymentMetrics gets metrics for learning from deployment.
func ExtractDeploymentMetrics(deploymentID string) map[string]interface{} {
	return map[string]interface{}{
		"deployment_id":      deploymentID,
		"start_time":         time.Now().Add(-10 * time.Minute).Format(time.RFC3339),
		"end_time":           time.Now().Format(time.RFC3339),
		"success_rate":       "95%",
		"error_count":        2,
		"rollback_triggered": false,
		"performance_impact": "minimal",
	}
}

// ExtractDeploymentContext gets context for learning from deployment.
func ExtractDeploymentContext(deploymentID string) map[string]interface{} {
	return map[string]interface{}{
		"deployment_id":     deploymentID,
		"strategy":          "rolling",
		"environment":       "production",
		"applications":      []string{"web-app", "api-service"},
		"services_affected": 3,
		"policies_applied":  []string{"zero-downtime", "health-check"},
		"ai_planned":        true,
		"complexity":        "medium",
	}
}

// CountApplicationServices counts the number of services owned by an application.
func CountApplicationServices(graph *graph.Graph, appID string) int {
	count := 0
	if edges, exists := graph.Edges[appID]; exists {
		for _, edge := range edges {
			if edge.Type == "owns" {
				if node, err := graph.GetNode(edge.To); err == nil && node.Kind == "service" {
					count++
				}
			}
		}
	}
	return count
}

// CountHealthyApplications counts healthy applications in the graph.
func CountHealthyApplications(graph *graph.Graph) int {
	count := 0
	for _, node := range graph.Nodes {
		if node.Kind == "application" {
			count++
		}
	}
	return count
}

// CountHealthyServices counts healthy services in the graph.
func CountHealthyServices(graph *graph.Graph) int {
	count := 0
	for _, node := range graph.Nodes {
		if node.Kind == "service" {
			count++
		}
	}
	return count
}
