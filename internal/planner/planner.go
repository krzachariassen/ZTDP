package planner

import (
	"errors"
	"sort"

	"github.com/krzachariassen/ZTDP/internal/graph"
)

// Planner computes execution plans for deployments/operations using topological sort.
type Planner struct {
	Graph *graph.Graph
}

// NewPlanner creates a new Planner for the given graph.
func NewPlanner(g *graph.Graph) *Planner {
	return &Planner{Graph: g}
}

// Plan returns a topologically sorted list of node IDs for execution order.
// Only considers edges of type "deploy" for ordering (legacy, use PlanWithEdgeTypes).
func (p *Planner) Plan() ([]string, error) {
	return p.PlanWithEdgeTypes([]string{"deploy"})
}

// PlanWithEdgeTypes returns a topologically sorted list of node IDs for execution order.
// Only considers edges of the given types for ordering.
func (p *Planner) PlanWithEdgeTypes(edgeTypes []string) ([]string, error) {
	typeSet := map[string]struct{}{}
	for _, t := range edgeTypes {
		typeSet[t] = struct{}{}
	}
	inDegree := make(map[string]int)
	for id := range p.Graph.Nodes {
		inDegree[id] = 0
	}
	for _, edges := range p.Graph.Edges {
		for _, e := range edges {
			if _, ok := typeSet[e.Type]; ok {
				inDegree[e.To]++
			}
		}
	}
	var queue []string
	// Sort nodes to ensure deterministic order when multiple nodes have same in-degree
	var nodeIDs []string
	for id := range inDegree {
		nodeIDs = append(nodeIDs, id)
	}
	sort.Strings(nodeIDs)
	for _, id := range nodeIDs {
		if inDegree[id] == 0 {
			queue = append(queue, id)
		}
	}
	var order []string
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		order = append(order, n)

		// Collect nodes that become ready (in-degree 0) after processing current node
		var newReadyNodes []string
		for _, e := range p.Graph.Edges[n] {
			if _, ok := typeSet[e.Type]; ok {
				inDegree[e.To]--
				if inDegree[e.To] == 0 {
					newReadyNodes = append(newReadyNodes, e.To)
				}
			}
		}
		// Sort newly ready nodes to ensure deterministic order
		sort.Strings(newReadyNodes)
		queue = append(queue, newReadyNodes...)
	}
	if len(order) != len(p.Graph.Nodes) {
		return nil, errors.New("cycle detected or disconnected nodes in graph")
	}
	return order, nil
}

// ExtractApplicationSubgraph returns a subgraph containing only the nodes and edges relevant to the given application.
func ExtractApplicationSubgraph(appName string, g *graph.Graph) *graph.Graph {
	sub := graph.NewGraph()
	visited := map[string]struct{}{}
	var visit func(id string)
	visit = func(id string) {
		if _, ok := visited[id]; ok {
			return
		}
		visited[id] = struct{}{}
		if n, ok := g.Nodes[id]; ok {
			nCopy := *n
			sub.Nodes[id] = &nCopy
			for _, e := range g.Edges[id] {
				sub.Edges[id] = append(sub.Edges[id], e)
				visit(e.To)
			}
		}
	}
	visit(appName)
	return sub
}
