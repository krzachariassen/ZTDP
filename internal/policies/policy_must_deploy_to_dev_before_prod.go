package policies

import "fmt"

// MustDeployToDevBeforeProdPolicy enforces that a service version must be deployed to 'dev' before it can be deployed to 'prod'.
type MustDeployToDevBeforeProdPolicy struct{}

func NewMustDeployToDevBeforeProdPolicy() *MustDeployToDevBeforeProdPolicy {
	return &MustDeployToDevBeforeProdPolicy{}
}

func (p *MustDeployToDevBeforeProdPolicy) Name() string {
	return "MustDeployToDevBeforeProdPolicy"
}

func (p *MustDeployToDevBeforeProdPolicy) Validate(g GraphView, m Mutation) error {
	if m.Type != "add_edge" || m.Edge == nil {
		return nil // Only care about edge additions
	}
	if m.Edge.Type != "deploy" {
		return nil // Only care about deployments
	}
	// Only enforce for prod deployments
	toNode, ok := g.Nodes[m.Edge.To]
	if !ok || toNode.Kind != "environment" || toNode.Metadata["name"] != "prod" {
		return nil
	}
	verNode, ok := g.Nodes[m.Edge.From]
	if !ok || verNode.Kind != "service_version" {
		return nil
	}
	// Find the service node (parent of service_version)
	var serviceNode *NodeView
	for _, edges := range g.Edges {
		for _, e := range edges {
			if e.To == verNode.ID && e.Type == "has_version" {
				n := g.Nodes[e.From]
				serviceNode = &n
				break
			}
		}
	}
	if serviceNode == nil {
		return nil // Can't find service node, skip
	}
	// Check if this version is already deployed to 'dev'
	for _, e := range g.Edges[verNode.ID] {
		if e.Type == "deploy" {
			devNode, ok := g.Nodes[e.To]
			if ok && devNode.Kind == "environment" && devNode.Metadata["name"] == "dev" {
				return nil // Already deployed to dev, allow
			}
		}
	}
	return fmt.Errorf("must deploy service version %s to 'dev' before deploying to 'prod'", verNode.ID)
}
