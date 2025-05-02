package main

import (
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"

	"github.com/krzachariassen/ZTDP/internal/contracts"
	"github.com/krzachariassen/ZTDP/internal/graph"
)

func main() {
	var backend graph.GraphBackend
	switch os.Getenv("ZTDP_GRAPH_BACKEND") {
	case "redis":
		rdb := redis.NewClient(&redis.Options{
			Addr:     os.Getenv("REDIS_HOST"), // or your in-cluster Redis DNS
			Password: os.Getenv("REDIS_PASSWORD"),
		})
		backend = graph.NewRedisGraph(rdb)
	default:
		backend = graph.NewMemoryGraph()
	}
	store := graph.NewGraphStore(backend)

	envs := []string{"dev", "qa"}

	app := contracts.ApplicationContract{
		Metadata: contracts.Metadata{
			Name:  "checkout",
			Owner: "team-x",
		},
		Spec: contracts.ApplicationSpec{
			Description:  "Handles checkout flows",
			Tags:         []string{"payments"},
			Environments: envs,
			Lifecycle:    map[string]contracts.LifecycleDefinition{},
		},
	}

	// Create app and service per env
	for _, env := range envs {
		appNode, _ := graph.ResolveContract(app)
		if err := store.AddNode(env, appNode); err != nil {
			fmt.Printf("Failed to add node [%s]: %v\n", appNode.ID, err)
		}

		svc := contracts.ServiceContract{
			Metadata: contracts.Metadata{
				Name:  "checkout-api",
				Owner: "team-x",
			},
			Spec: struct {
				Application string `json:"application"`
				Port        int    `json:"port"`
				Public      bool   `json:"public"`
			}{
				Application: "checkout",
				Port:        8080,
				Public:      true,
			},
		}

		svcNode, _ := graph.ResolveContract(svc)
		if err := store.AddNode(env, svcNode); err != nil {
			fmt.Printf("Failed to add node [%s]: %v\n", appNode.ID, err)
		}

		if err := store.AddEdge(env, "checkout-api", "checkout"); err != nil {
			fmt.Printf("Failed to add edge: %v\n", err)
		}
	}

	// Print results
	for _, env := range envs {
		fmt.Printf("\n📦 Env: %s\n", env)
		g, _ := store.GetGraph(env)

		fmt.Println("  Nodes:")
		for id, n := range g.Nodes {
			fmt.Printf("   - [%s] %s\n", n.Kind, id)
		}
		fmt.Println("  Edges:")
		for from, toList := range g.Edges {
			for _, to := range toList {
				fmt.Printf("   - %s --> %s\n", from, to)
			}
		}
	}
}
