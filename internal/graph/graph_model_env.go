package graph

// GetEnvironments returns all environment names in the graph.
func (g *Graph) GetEnvironments() []string {
	environments := []string{}
	environmentSet := make(map[string]bool)

	// Look for environment nodes
	for _, node := range g.Nodes {
		if node.Kind == KindEnvironment {
			if name, ok := node.Metadata["name"].(string); ok {
				if !environmentSet[name] {
					environments = append(environments, name)
					environmentSet[name] = true
				}
			}
		}
	}

	// If no environments found, always include "default"
	if len(environments) == 0 {
		environments = append(environments, "default")
	}

	return environments
}
