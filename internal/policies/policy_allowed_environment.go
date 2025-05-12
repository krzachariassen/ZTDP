package policies

import (
	"fmt"
	"log"
)

// AllowedEnvironmentPolicy enforces that a service version can only be deployed to environments allowed for its application.
type AllowedEnvironmentPolicy struct{}

func NewAllowedEnvironmentPolicy() *AllowedEnvironmentPolicy {
	return &AllowedEnvironmentPolicy{}
}

func (p *AllowedEnvironmentPolicy) Name() string {
	return "AllowedEnvironmentPolicy"
}

func (p *AllowedEnvironmentPolicy) Validate(g GraphView, m Mutation) error {
	log.Printf("[AllowedEnvironmentPolicy] here!")
	if m.Type != "add_edge" || m.Edge == nil {
		log.Printf("[AllowedEnvironmentPolicy] Skipping: mutation type is '%s' or edge is nil", m.Type)
		return nil // Only care about edge additions
	}
	if m.Edge.Type != "deploy" {
		log.Printf("[AllowedEnvironmentPolicy] Skipping: edge type is '%s'", m.Edge.Type)
		return nil // Only care about deployments
	}
	verNode, ok := g.Nodes[m.Edge.From]
	if !ok || verNode.Kind != "service_version" {
		log.Printf("[AllowedEnvironmentPolicy] Error: service version node not found or not a service_version (ok=%v, kind=%s)", ok, verNode.Kind)
		return fmt.Errorf("service version node not found or not a service_version")
	}
	// Find the service node (parent of service_version)
	var serviceNode *NodeView
	for _, edges := range g.Edges {
		for _, e := range edges {
			if e.To == verNode.ID && e.Type == "has_version" {
				n := g.Nodes[e.From]
				serviceNode = &n
				log.Printf("[AllowedEnvironmentPolicy] Found service node: %s", serviceNode.ID)
				break
			}
		}
	}
	if serviceNode == nil {
		log.Printf("[AllowedEnvironmentPolicy] Skipping: can't find service node for version %s", verNode.ID)
		return nil // Can't find service node, skip
	}
	// Find the application node (parent of service)
	var appNode *NodeView
	for _, edges := range g.Edges {
		for _, e := range edges {
			if e.To == serviceNode.ID && e.Type == "owns" {
				n := g.Nodes[e.From]
				appNode = &n
				log.Printf("[AllowedEnvironmentPolicy] Found app node: %s", appNode.ID)
				break
			}
		}
	}
	if appNode == nil {
		log.Printf("[AllowedEnvironmentPolicy] Skipping: can't find app node for service %s", serviceNode.ID)
		return nil // Can't find app node, skip
	}
	if !isEnvironmentAllowedForApp(g, appNode.ID, m.Edge.To) {
		log.Printf("[AllowedEnvironmentPolicy] Denied: deployment to environment '%s' is not allowed for application '%s'", m.Edge.To, appNode.ID)
		return fmt.Errorf("deployment to environment '%s' is not allowed for application '%s'", m.Edge.To, appNode.ID)
	}
	log.Printf("[AllowedEnvironmentPolicy] Allowed: deployment to environment '%s' for application '%s'", m.Edge.To, appNode.ID)
	return nil
}

func isEnvironmentAllowedForApp(g GraphView, appID, envID string) bool {
	edges := g.Edges[appID]
	for _, edge := range edges {
		if edge.Type == "allowed_in" && edge.To == envID {
			return true
		}
	}
	return false
}
