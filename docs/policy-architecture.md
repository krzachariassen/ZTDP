# Zero Trust Developer Platform (ZTDP) - Policy Architecture

## Overview

ZTDP's policy architecture enables governance of application deployments and transitions between nodes in the platform graph. Policies enforce security, compliance, and operational constraints.

## Graph-Based Policy Model

The ZTDP policy system uses a graph-based approach where policies are represented as nodes in the directed graph, with the following characteristics:

* **Policy Nodes**: First-class entities in the graph with the `kind: policy` attribute
* **Policy Attachment**: Policies are attached to specific transitions (edges) between nodes
* **Policy Satisfaction**: Checks and approvals can satisfy policies, enabling transitions
* **Policy Enforcement**: Enforced when attempting to create edges or transitions between nodes

## Policy Enforcement Process

1. When a transition is requested (e.g., deploying a service to an environment), the system checks if there are policies attached to that transition
2. For each attached policy, the system verifies if there are satisfying checks or approvals
3. If all policies are satisfied, the transition is allowed; otherwise, it's blocked

## Policy Types

ZTDP supports several policy types:

* **Deployment Policies**: Control which environments services can be deployed to
* **Transition Policies**: Enforce ordered sequences (e.g., deploy to dev before prod)
* **Approval Policies**: Require explicit approvals before transitions
* **Check Policies**: Require automated checks to pass before transitions

## Example: Dev-Before-Prod Policy

This policy ensures services are deployed to the development environment before production:

```go
// Policy node in the graph
policyNode := &graph.Node{
    ID:   "policy-dev-before-prod",
    Kind: graph.KindPolicy,
    Metadata: map[string]interface{}{
        "name":        "Must Deploy To Dev Before Prod", 
        "description": "Requires a service version to be deployed to dev before it can be deployed to prod",
        "type":        graph.PolicyTypeSystem,
        "status":      "active",
    },
    Spec: map[string]interface{}{
        "sourceKind":      graph.KindServiceVersion,
        "targetKind":      graph.KindEnvironment,
        "targetID":        "prod",
        "requiredPathIDs": []string{"dev"},
    },
}

// Attach policy to a transition
graph.AttachPolicyToTransition(serviceVersionID, "prod", graph.EdgeTypeDeploy, policyNode.ID)
```

## Usage

### Creating a Policy

```go
// Create policy node
policyNode := &graph.Node{
    ID:   "custom-policy-id",
    Kind: graph.KindPolicy,
    Metadata: map[string]interface{}{
        "name":        "Custom Policy Name",
        "description": "Description of what the policy enforces",
        "type":        graph.PolicyTypeSystem,
        "status":      "active",
    },
    Spec: map[string]interface{}{
        // Policy-specific configuration
        "sourceKind": "service_version",
        "targetKind": "environment",
        // Additional parameters as needed
    },
}
g.AddNode(policyNode)
```

### Attaching a Policy to a Transition

```go
// Attach policy to transition between nodes
g.AttachPolicyToTransition(fromNodeID, toNodeID, edgeType, policyID)
```

### Satisfying a Policy with Checks

```go
// Create a check node
checkNode := &graph.Node{
    ID:   "check-id",
    Kind: graph.KindCheck,
    Metadata: map[string]interface{}{
        "name":   "Check Name",
        "type":   "check-type",
        "status": graph.CheckStatusSucceeded,
    },
    Spec: map[string]interface{}{
        // Check-specific parameters
    },
}
g.AddNode(checkNode)

// Link check to policy with "satisfies" relationship
g.AddEdge(checkNode.ID, policyID, graph.EdgeTypeSatisfies)
```

## API Integration

The policy system is integrated with the ZTDP API for:

* Managing policies through the API
* Enforcing policies during deployments and other transitions
* Viewing policy status and requirements

For example, when deploying a service version:

```bash
# This API call is checked against policies before execution
curl -X POST http://localhost:8080/v1/applications/checkout/services/checkout-api/versions/1.0.0/deploy \
  -H "Content-Type: application/json" \
  -d '{
    "environment": "prod"
  }'
```

If policies are not satisfied, the API will return an appropriate error explaining which policies need to be satisfied.
