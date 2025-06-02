package graph

import (
	"errors"
	"fmt"

	"github.com/krzachariassen/ZTDP/internal/contracts"
)

type Edge struct {
	To       string                 `json:"to"`
	Type     string                 `json:"type"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type Graph struct {
	Nodes map[string]*Node  `json:"nodes"`
	Edges map[string][]Edge `json:"edges"`
}

type Node struct {
	ID       string                 `json:"id"`
	Kind     string                 `json:"kind"`
	Metadata map[string]interface{} `json:"metadata"`
	Spec     map[string]interface{} `json:"spec"`
}

func NewGraph() *Graph {
	return &Graph{
		Nodes: make(map[string]*Node),
		Edges: make(map[string][]Edge),
	}
}

func (g *Graph) AddNode(n *Node) error {
	if _, exists := g.Nodes[n.ID]; exists {
		return fmt.Errorf("node with ID %s already exists", n.ID)
	}
	g.Nodes[n.ID] = n
	return nil
}

func (g *Graph) GetNode(id string) (*Node, error) {
	n, ok := g.Nodes[id]
	if !ok {
		return nil, fmt.Errorf("node with ID %s not found", id)
	}
	return n, nil
}

func (g *Graph) AddEdge(fromID, toID, relType string) error {
	if !IsValidEdgeType(relType) {
		return fmt.Errorf("invalid edge type: %s", relType)
	}
	if _, ok := g.Nodes[fromID]; !ok {
		return fmt.Errorf("source node %s does not exist", fromID)
	}
	if _, ok := g.Nodes[toID]; !ok {
		return fmt.Errorf("target node %s does not exist", toID)
	}
	for _, existing := range g.Edges[fromID] {
		if existing.To == toID && existing.Type == relType {
			return errors.New("edge already exists")
		}
	}

	// Create and validate edge contract
	if err := g.validateEdgeContract(fromID, toID, relType); err != nil {
		return fmt.Errorf("edge validation failed: %w", err)
	}

	// Check policy requirements for this transition
	if err := g.IsTransitionAllowed(fromID, toID, relType); err != nil {
		return fmt.Errorf("policy validation failed: %w", err)
	}

	g.Edges[fromID] = append(g.Edges[fromID], Edge{To: toID, Type: relType})
	return nil
}

// UpdateNode updates an existing node in the graph.
// If the node doesn't exist, an error is returned.
func (g *Graph) UpdateNode(node *Node) error {
	if _, exists := g.Nodes[node.ID]; !exists {
		return fmt.Errorf("node with ID %s not found", node.ID)
	}
	g.Nodes[node.ID] = node
	return nil
}

// validateEdgeContract validates an edge using the contract system
func (g *Graph) validateEdgeContract(fromID, toID, relType string) error {
	fromNode := g.Nodes[fromID]
	toNode := g.Nodes[toID]

	edgeContract := contracts.EdgeContract{
		FromID:   fromID,
		ToID:     toID,
		Type:     relType,
		FromKind: fromNode.Kind,
		ToKind:   toNode.Kind,
	}

	// Validate the basic edge contract
	if err := edgeContract.Validate(); err != nil {
		return err
	}

	// Apply special validation rules that need full node data
	return g.validateSpecialEdgeRules(fromNode, toNode, relType)
}

// validateSpecialEdgeRules applies special validation that requires full node data
func (g *Graph) validateSpecialEdgeRules(fromNode, toNode *Node, edgeType string) error {
	// Find the applicable rule
	var applicableRule *contracts.EdgeValidationRule
	for _, rule := range contracts.EdgeValidationRules {
		if rule.FromKind == fromNode.Kind && rule.ToKind == toNode.Kind {
			applicableRule = &rule
			break
		}
	}

	if applicableRule != nil && applicableRule.SpecialRules != nil {
		// Convert nodes to the expected format for special rules
		fromData := map[string]interface{}{
			"id":       fromNode.ID,
			"kind":     fromNode.Kind,
			"metadata": fromNode.Metadata,
			"spec":     fromNode.Spec,
		}
		toData := map[string]interface{}{
			"id":       toNode.ID,
			"kind":     toNode.Kind,
			"metadata": toNode.Metadata,
			"spec":     toNode.Spec,
		}

		return applicableRule.SpecialRules(fromData, toData)
	}

	return nil
}

// GetEdge retrieves an edge by constructing an edge ID from fromID, toID, and edgeType
func (g *Graph) GetEdge(edgeID string) (*Edge, bool) {
	// Try to find edge by searching through all edges
	// EdgeID format expected: "fromID-toID-type" or "fromID-toID"
	for fromID, edges := range g.Edges {
		for i := range edges {
			// Try multiple formats for edge identification
			fullID := fmt.Sprintf("%s-%s-%s", fromID, edges[i].To, edges[i].Type)
			shortID := fmt.Sprintf("%s-%s", fromID, edges[i].To)

			if fullID == edgeID || shortID == edgeID {
				return &edges[i], true
			}
		}
	}
	return nil, false
}

// UpdateEdge updates an existing edge in the graph
func (g *Graph) UpdateEdge(edge *Edge) error {
	// Find the edge by searching through all edges
	for fromID, edges := range g.Edges {
		for i := range edges {
			// Check if this is the edge we want to update
			// We'll match based on To and Type fields
			if edges[i].To == edge.To && edges[i].Type == edge.Type {
				// Update the edge
				g.Edges[fromID][i] = *edge
				return nil
			}
		}
	}
	return fmt.Errorf("edge not found for update")
}

// GetEdgeByFromToType retrieves an edge by explicit from, to, and type parameters
func (g *Graph) GetEdgeByFromToType(fromID, toID, edgeType string) (*Edge, bool) {
	edges, exists := g.Edges[fromID]
	if !exists {
		return nil, false
	}

	for i := range edges {
		if edges[i].To == toID && edges[i].Type == edgeType {
			return &edges[i], true
		}
	}
	return nil, false
}
