package main

import (
	"fmt"

	"github.com/krzachariassen/ZTDP/internal/contracts"
	"github.com/krzachariassen/ZTDP/internal/graph"
)

func main() {
	g := graph.NewGraph()

	app := &graph.Node{
		ID:   "checkout",
		Kind: "application",
		Metadata: contracts.Metadata{
			Name:        "checkout",
			Environment: "dev",
			Owner:       "team-x",
		},
	}

	svc := &graph.Node{
		ID:   "checkout-api",
		Kind: "service",
		Metadata: contracts.Metadata{
			Name:        "checkout-api",
			Environment: "dev",
			Owner:       "team-x",
		},
	}

	g.AddNode(app)
	g.AddNode(svc)
	g.AddEdge("checkout-api", "checkout")

	fmt.Println("Nodes:")
	for id, n := range g.Nodes {
		fmt.Printf(" - [%s] %s (%s)", n.Kind, id, n.Metadata.Environment)
	}

	fmt.Println("Edges:")
	for from, toList := range g.Edges {
		for _, to := range toList {
			fmt.Printf(" - %s --> %s", from, to)
		}
	}
}
