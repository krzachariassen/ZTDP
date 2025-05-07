package policies

import "github.com/krzachariassen/ZTDP/internal/graph"

// GetAllowedEnvironmentsForApp returns a list of environment node IDs that the application is allowed to deploy to.
func GetAllowedEnvironmentsForApp(g *graph.Graph, appID string) []string {
	envs := []string{}
	for _, edge := range g.Edges[appID] {
		if edge.Type == "allowed_in" {
			envs = append(envs, edge.To)
		}
	}
	return envs
}

// IsEnvironmentAllowedForApp returns true if the environment is allowed for the application.
func IsEnvironmentAllowedForApp(g *graph.Graph, appID, envID string) bool {
	for _, edge := range g.Edges[appID] {
		if edge.Type == "allowed_in" && edge.To == envID {
			return true
		}
	}
	return false
}
