package orchestrator

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/krzachariassen/ZTDP/internal/graph"
)

// getPlatformState gets current platform state with detailed information
func (o *Orchestrator) getPlatformState() string {
	if o.graph == nil {
		return "Platform state: Not available"
	}

	// Get the current graph
	currentGraph, err := o.graph.Graph()
	if err != nil {
		return "Platform state: Error loading graph"
	}

	// Get detailed lists
	applications := o.getNodesByKind(currentGraph.Nodes, "application")
	services := o.getNodesByKind(currentGraph.Nodes, "service")
	environments := o.getNodesByKind(currentGraph.Nodes, "environment")
	resources := o.getNodesByKind(currentGraph.Nodes, "resource")

	state := fmt.Sprintf(`Platform State:
- Total nodes: %d

APPLICATIONS (%d):`, len(currentGraph.Nodes), len(applications))

	if len(applications) == 0 {
		state += "\n  (No applications created yet)"
	} else {
		for _, app := range applications {
			name := o.getNodeName(app)
			state += fmt.Sprintf("\n  - %s", name)
		}
	}

	state += fmt.Sprintf("\n\nSERVICES (%d):", len(services))
	if len(services) == 0 {
		state += "\n  (No services created yet)"
	} else {
		for _, service := range services {
			name := o.getNodeName(service)
			state += fmt.Sprintf("\n  - %s", name)
		}
	}

	state += fmt.Sprintf("\n\nENVIRONMENTS (%d):", len(environments))
	if len(environments) == 0 {
		state += "\n  (No environments created yet)"
	} else {
		for _, env := range environments {
			name := o.getNodeName(env)
			state += fmt.Sprintf("\n  - %s", name)
		}
	}

	state += fmt.Sprintf("\n\nRESOURCES (%d):", len(resources))
	if len(resources) == 0 {
		state += "\n  (No resources created yet)"
	} else {
		for _, resource := range resources {
			name := o.getNodeName(resource)
			state += fmt.Sprintf("\n  - %s", name)
		}
	}

	return state
}

// getNodeName extracts the name from a node's metadata
func (o *Orchestrator) getNodeName(node *graph.Node) string {
	if node.Metadata != nil {
		if name, ok := node.Metadata["name"].(string); ok {
			return name
		}
	}
	return node.ID // fallback to ID if no name found
}

// getNodesByKind returns all nodes of a specific kind
func (o *Orchestrator) getNodesByKind(nodes map[string]*graph.Node, kind string) []*graph.Node {
	var result []*graph.Node
	for _, node := range nodes {
		if node.Kind == kind {
			result = append(result, node)
		}
	}

	return result
}

// loadAllContracts dynamically loads all contract definitions
func (o *Orchestrator) loadAllContracts() string {
	contractsDir := "/mnt/c/Work/git/ztdp/internal/contracts"

	contracts := ""
	contractFiles := []string{"application.go", "service.go", "environment.go", "resource.go"}

	for _, file := range contractFiles {
		if content, err := os.ReadFile(filepath.Join(contractsDir, file)); err == nil {
			contracts += fmt.Sprintf("\n// %s\n%s\n", file, string(content))
		}
	}

	return contracts
}
