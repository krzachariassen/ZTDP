package main

import (
	"fmt"
	"os"

	"github.com/krzachariassen/ZTDP/internal/contracts"
	"github.com/krzachariassen/ZTDP/internal/graph"
)

func main() {
	var backend graph.GraphBackend
	switch os.Getenv("ZTDP_GRAPH_BACKEND") {
	case "redis":
		backend = graph.NewRedisGraph(graph.RedisGraphConfig{})
	default:
		fmt.Println("⚙️  Using backend: Memory")
		backend = graph.NewMemoryGraph()
	}

	global := graph.NewGlobalGraph(backend)

	var app contracts.ApplicationContract
	// Try loading from backend
	if err := global.Load(); err != nil {
		fmt.Println("🔄 No existing global graph found, creating new one")
		// Setup global graph for first time
		app = contracts.ApplicationContract{
			Metadata: contracts.Metadata{
				Name:  "checkout",
				Owner: "team-x",
			},
			Spec: contracts.ApplicationSpec{
				Description:  "Handles checkout flows",
				Tags:         []string{"payments"},
				Environments: []string{"dev", "qa"},
				Lifecycle:    map[string]contracts.LifecycleDefinition{},
			},
		}

		appNode, _ := graph.ResolveContract(app)
		global.AddNode(appNode)

		svc := contracts.ServiceContract{
			Metadata: contracts.Metadata{
				Name:  "checkout-api",
				Owner: "team-x",
			},
			Spec: contracts.ServiceSpec{
				Application: "checkout",
				Port:        8080,
				Public:      true,
			},
		}

		svcNode, _ := graph.ResolveContract(svc)
		global.AddNode(svcNode)
		global.AddEdge(app.Metadata.Name, svc.Metadata.Name, "owns")

		// Save it
		if err := global.Save(); err != nil {
			fmt.Printf("❌ Failed to save global graph: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("✅ Loaded global graph from backend")
		// Find the application node in the loaded graph
		if n, ok := global.Graph.Nodes["checkout"]; ok {
			// Unmarshal node.Spec into app
			contract, err := graph.LoadNode(n.Kind, n.Spec, n.Metadata)
			if err == nil {
				if loadedApp, ok := contract.(*contracts.ApplicationContract); ok {
					app = *loadedApp
				}
			}
		}
	}

	for _, env := range app.Spec.Environments {
		fmt.Printf("\n🌐 GlobalGraph applied to [%s]\n", env)
		g, _ := global.Apply(env)

		fmt.Println("  Nodes:")
		for id, n := range g.Nodes {
			fmt.Printf("   - [%s] %s\n", n.Kind, id)
		}
	}

	// Print global edges once
	fmt.Println("\nGlobal Edges:")
	for from, edgeList := range global.Graph.Edges {
		for _, edge := range edgeList {
			fmt.Printf("   - {%s} -%s-> {%s}\n", from, edge.Type, edge.To)
		}
	}
}
