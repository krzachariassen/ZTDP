package main

import (
	"fmt"
	"os"
	"time"

	"github.com/krzachariassen/ZTDP/internal/contracts"
	"github.com/krzachariassen/ZTDP/internal/graph"
)

func MapToMetadata(m map[string]interface{}) contracts.Metadata {
	return contracts.Metadata{
		Name:  m["name"].(string),
		Owner: m["owner"].(string),
	}
}

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
		// Remove Environments from ApplicationSpec and use only the graph for environment relationships
		app = contracts.ApplicationContract{
			Metadata: contracts.Metadata{
				Name:  "checkout",
				Owner: "team-x",
			},
			Spec: contracts.ApplicationSpec{
				Description: "Handles checkout flows",
				Tags:        []string{"payments"},
				Lifecycle:   map[string]contracts.LifecycleDefinition{},
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

		// Add environment nodes
		envDev := contracts.EnvironmentContract{
			Metadata: contracts.Metadata{
				Name:  "dev",
				Owner: "platform-team",
			},
			Spec: contracts.EnvironmentSpec{
				Description: "Development environment",
			},
		}
		envProd := contracts.EnvironmentContract{
			Metadata: contracts.Metadata{
				Name:  "prod",
				Owner: "platform-team",
			},
			Spec: contracts.EnvironmentSpec{
				Description: "Production environment",
			},
		}
		envDevNode, _ := graph.ResolveContract(envDev)
		global.AddNode(envDevNode)
		envProdNode, _ := graph.ResolveContract(envProd)
		global.AddNode(envProdNode)

		// Add second service
		workerSvc := contracts.ServiceContract{
			Metadata: contracts.Metadata{
				Name:  "checkout-worker",
				Owner: "team-x",
			},
			Spec: contracts.ServiceSpec{
				Application: "checkout",
				Port:        9090,
				Public:      false,
			},
		}
		workerSvcNode, _ := graph.ResolveContract(workerSvc)
		global.AddNode(workerSvcNode)
		global.AddEdge(app.Metadata.Name, workerSvc.Metadata.Name, "owns")

		// Link services to environments (deployed_in)
		// global.AddEdge(svc.Metadata.Name, envDev.Metadata.Name, "deployed_in")
		// global.AddEdge(workerSvc.Metadata.Name, envDev.Metadata.Name, "deployed_in")
		// global.AddEdge(svc.Metadata.Name, envProd.Metadata.Name, "deployed_in")

		// Add service version for checkout-api
		version := "1.0.0"
		serviceVersion := contracts.ServiceVersionContract{
			IDValue:   svc.Metadata.Name + ":" + version,
			Name:      svc.Metadata.Name,
			Owner:     svc.Metadata.Owner,
			Version:   version,
			ConfigRef: "default-config",
			CreatedAt: time.Now(),
		}
		serviceVersionNode, _ := graph.ResolveContract(serviceVersion)
		global.AddNode(serviceVersionNode)
		global.AddEdge(svc.Metadata.Name, serviceVersion.ID(), "has_version")

		// Deploy service version to environments
		global.AddEdge(serviceVersion.ID(), envDev.Metadata.Name, "deployed_in")
		global.AddEdge(serviceVersion.ID(), envProd.Metadata.Name, "deployed_in")

		// Add service version for checkout-worker
		workerVersion := "1.0.0"
		workerServiceVersion := contracts.ServiceVersionContract{
			IDValue:   workerSvc.Metadata.Name + ":" + workerVersion,
			Name:      workerSvc.Metadata.Name,
			Owner:     workerSvc.Metadata.Owner,
			Version:   workerVersion,
			ConfigRef: "default-config",
			CreatedAt: time.Now(),
		}
		workerServiceVersionNode, _ := graph.ResolveContract(workerServiceVersion)
		global.AddNode(workerServiceVersionNode)
		global.AddEdge(workerSvc.Metadata.Name, workerServiceVersion.ID(), "has_version")
		global.AddEdge(workerServiceVersion.ID(), envDev.Metadata.Name, "deployed_in")
		global.AddEdge(workerServiceVersion.ID(), envProd.Metadata.Name, "deployed_in")

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
			contract, err := graph.LoadNode(n.Kind, n.Spec, MapToMetadata(n.Metadata))
			if err == nil {
				if loadedApp, ok := contract.(*contracts.ApplicationContract); ok {
					app = *loadedApp
				}
			}
		}
	}

	fmt.Println("\nGlobal Nodes:")
	if len(global.Graph.Nodes) == 0 {
		fmt.Println("   (none)")
	} else {
		for id, n := range global.Graph.Nodes {
			fmt.Printf("   - [%s] %s\n", n.Kind, id)
		}
	}
	fmt.Println("\nGlobal Edges:")
	empty := true
	for from, edgeList := range global.Graph.Edges {
		for _, edge := range edgeList {
			fmt.Printf("   - {%s} -%s-> {%s}\n", from, edge.Type, edge.To)
			empty = false
		}
	}
	if empty {
		fmt.Println("   (none)")
	}
}
