package main

import (
	"fmt"
	"os"
	"time"

	"github.com/krzachariassen/ZTDP/internal/contracts"
	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/krzachariassen/ZTDP/internal/resources"
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

	// Check for --force flag to create new graph
	forceCreate := false
	for _, arg := range os.Args {
		if arg == "--force" || arg == "-f" {
			forceCreate = true
			break
		}
	}

	var app contracts.ApplicationContract
	if forceCreate || global.Load() != nil {
		fmt.Println("🔄 No existing global graph found, creating new one")
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
		global.AddEdge(serviceVersion.ID(), envDev.Metadata.Name, "deploy")
		global.AddEdge(serviceVersion.ID(), envProd.Metadata.Name, "deploy")

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
		global.AddEdge(workerServiceVersion.ID(), envDev.Metadata.Name, "deploy")
		global.AddEdge(workerServiceVersion.ID(), envProd.Metadata.Name, "deploy")

		// Add another application with its own resource instances for demonstration
		paymentApp := contracts.ApplicationContract{
			Metadata: contracts.Metadata{
				Name:  "payment",
				Owner: "team-y",
			},
			Spec: contracts.ApplicationSpec{
				Description: "Handles payment processing",
				Tags:        []string{"payments", "financial"},
				Lifecycle:   map[string]contracts.LifecycleDefinition{},
			},
		}
		paymentAppNode, _ := graph.ResolveContract(paymentApp)
		global.AddNode(paymentAppNode)
		paymentSvc := contracts.ServiceContract{
			Metadata: contracts.Metadata{
				Name:  "payment-api",
				Owner: "team-y",
			},
			Spec: contracts.ServiceSpec{
				Application: "payment",
				Port:        8081,
				Public:      true,
			},
		}
		paymentSvcNode, _ := graph.ResolveContract(paymentSvc)
		global.AddNode(paymentSvcNode)
		global.AddEdge(paymentApp.Metadata.Name, paymentSvc.Metadata.Name, "owns")

		// --- Resource catalog and instance setup ---
		setupResourcesWithContracts(global, app, paymentApp, svc, workerSvc, paymentSvc)

		if err := global.Save(); err != nil {
			fmt.Printf("❌ Failed to save global graph: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("✅ Loaded global graph from backend")
		if n, ok := global.Graph.Nodes["checkout"]; ok {
			contract, err := resources.LoadNode(n.Kind, n.Spec, contracts.Metadata{
				Name:  n.ID,
				Owner: n.Metadata["owner"].(string),
			})
			if err == nil {
				if loadedApp, ok := contract.(*contracts.ApplicationContract); ok {
					app = *loadedApp
				}
			}
		}
	}

	// --- Print summary of the generated demo data ---
	fmt.Println("\n📋 Global Graph Nodes:")
	if len(global.Graph.Nodes) == 0 {
		fmt.Println("   (none)")
	} else {
		nodesByKind := make(map[string][]string)
		for id, n := range global.Graph.Nodes {
			nodesByKind[n.Kind] = append(nodesByKind[n.Kind], id)
		}
		for kind, nodes := range nodesByKind {
			fmt.Printf("\n   [%s]:\n", kind)
			for _, id := range nodes {
				fmt.Printf("     - %s\n", id)
			}
		}
	}

	fmt.Println("\n🔗 Global Graph Edges:")
	empty := true
	for from, edgeList := range global.Graph.Edges {
		for _, edge := range edgeList {
			fmt.Printf("     - %s -%s-> %s\n", from, edge.Type, edge.To)
			empty = false
		}
	}
	if empty {
		fmt.Println("   (none)")
	}

	// --- Policy node and relationships setup (always run) ---
	fmt.Println("\n🔒 Adding policy node and relationships to the demo graph...")
	devBeforeProdPolicy := &graph.Node{
		ID:   "policy-dev-before-prod",
		Kind: "policy",
		Metadata: map[string]interface{}{
			"name":        "Must Deploy To Dev Before Prod",
			"description": "Requires a service version to be deployed to dev before it can be deployed to prod",
			"type":        "system",
			"status":      "active",
		},
		Spec: map[string]interface{}{
			"sourceKind":      "service_version",
			"targetKind":      "environment",
			"targetID":        "prod",
			"requiredPathIDs": []string{"dev"},
		},
	}
	global.AddNode(devBeforeProdPolicy)

	// Add an explicit edge from the policy node to the application node it enforces
	global.AddEdge(devBeforeProdPolicy.ID, "checkout", "enforces")

	apiSvcV2 := createServiceVersion("checkout", "checkout-api", "2.0.0")
	apiSvcV2Node, _ := graph.ResolveContract(apiSvcV2)
	global.AddNode(apiSvcV2Node)
	global.AddEdge("checkout-api", apiSvcV2.ID(), "has_version")
	global.Graph.AttachPolicyToTransition(apiSvcV2.ID(), "prod", graph.EdgeTypeDeploy, devBeforeProdPolicy.ID)

	checkNode := &graph.Node{
		ID:   "check-dev-deployment-" + apiSvcV2.ID(),
		Kind: graph.KindCheck,
		Metadata: map[string]interface{}{
			"name":   "Dev Deployment Verification for " + apiSvcV2.ID(),
			"type":   "deployment-verification",
			"status": graph.CheckStatusSucceeded,
		},
		Spec: map[string]interface{}{
			"serviceVersion": apiSvcV2.ID(),
			"environment":    "dev",
		},
	}
	global.AddNode(checkNode)
	global.AddEdge(checkNode.ID, devBeforeProdPolicy.ID, graph.EdgeTypeSatisfies)
	global.AddEdge(apiSvcV2.ID(), "dev", graph.EdgeTypeDeploy)
	global.AddEdge(apiSvcV2.ID(), "prod", graph.EdgeTypeDeploy)

	// --- Print summary of the policy graph ---
	fmt.Println("\n🔒 Policy Nodes:")
	for id, n := range global.Graph.Nodes {
		if n.Kind == graph.KindPolicy {
			fmt.Printf("   - %s: %s\n", id, n.Metadata["name"])
		}
	}
	fmt.Println("\n🔗 Policy Attachments:")
	for from, edgeList := range global.Graph.Edges {
		for _, edge := range edgeList {
			// Print all edges to policy nodes
			if toNode, ok := global.Graph.Nodes[edge.To]; ok && toNode.Kind == graph.KindPolicy {
				fmt.Printf("   - %s -%s-> %s\n", from, edge.Type, edge.To)
			}
		}
	}
	fmt.Println("\n✅ Policy demonstration complete.")

	// Persist the updated graph (including policy nodes/edges) to Redis
	if err := global.Save(); err != nil {
		fmt.Printf("❌ Failed to save global graph after policy integration: %v\n", err)
	} else {
		fmt.Println("✅ Saved global graph with policy nodes/edges to backend.")
	}
}

// This is a helper function to create and add resource types and instances to the demo graph
func setupResourcesWithContracts(global *graph.GlobalGraph, app, paymentApp contracts.ApplicationContract, svc, workerSvc, paymentSvc contracts.ServiceContract) {
	// Create resource catalog root node
	catalogNode := graph.Node{
		ID:   "resource-catalog",
		Kind: "resource_register",
		Metadata: map[string]interface{}{
			"name":  "resource-catalog",
			"owner": "platform-team",
		},
		Spec: map[string]interface{}{
			"description": "Root node for all resource types in the platform",
		},
	}
	global.AddNode(&catalogNode)

	fmt.Println("📚 Creating resource catalog (resource types)...")

	// Postgres Database resource type
	postgresType := contracts.ResourceTypeContract{
		Metadata: contracts.Metadata{
			Name:  "postgres",
			Owner: "platform-team",
		},
		Spec: contracts.ResourceTypeSpec{
			Version:         "15.0",
			TierOptions:     []string{"standard", "high-memory", "high-cpu"},
			DefaultTier:     "standard",
			ConfigTemplate:  "config/templates/postgres-config.yaml",
			AvailablePlans:  []string{"dev", "prod"},
			DefaultCapacity: "10GB",
			ProviderMetadata: map[string]interface{}{
				"description": "PostgreSQL database service",
			},
		},
	}
	postgresTypeNode, _ := graph.ResolveContract(postgresType)
	global.AddNode(postgresTypeNode)
	global.AddEdge(catalogNode.ID, postgresType.Metadata.Name, "owns")

	// Redis Cache resource type
	redisType := contracts.ResourceTypeContract{
		Metadata: contracts.Metadata{
			Name:  "redis",
			Owner: "platform-team",
		},
		Spec: contracts.ResourceTypeSpec{
			Version:         "7.0",
			TierOptions:     []string{"cache", "persistent"},
			DefaultTier:     "cache",
			ConfigTemplate:  "config/templates/redis-config.yaml",
			AvailablePlans:  []string{"dev", "prod"},
			DefaultCapacity: "1GB",
			ProviderMetadata: map[string]interface{}{
				"description": "Redis in-memory cache service",
			},
		},
	}
	redisTypeNode, _ := graph.ResolveContract(redisType)
	global.AddNode(redisTypeNode)
	global.AddEdge(catalogNode.ID, redisType.Metadata.Name, "owns")

	// Kafka queue resource type
	kafkaType := contracts.ResourceTypeContract{
		Metadata: contracts.Metadata{
			Name:  "kafka",
			Owner: "platform-team",
		},
		Spec: contracts.ResourceTypeSpec{
			Version:         "3.4",
			TierOptions:     []string{"standard", "high-throughput"},
			DefaultTier:     "standard",
			ConfigTemplate:  "config/templates/kafka-config.yaml",
			AvailablePlans:  []string{"dev", "prod"},
			DefaultCapacity: "10GB",
			ProviderMetadata: map[string]interface{}{
				"description": "Kafka streaming platform",
			},
		},
	}
	kafkaTypeNode, _ := graph.ResolveContract(kafkaType)
	global.AddNode(kafkaTypeNode)
	global.AddEdge(catalogNode.ID, kafkaType.Metadata.Name, "owns")

	fmt.Println("🔧 Creating application resource instances...")

	// Create resource instances for the checkout application
	checkoutPgInstance := contracts.ResourceContract{
		Metadata: contracts.Metadata{
			Name:  "checkout-postgres",
			Owner: app.Metadata.Owner,
		},
		Spec: contracts.ResourceSpec{
			Type:     "postgres",
			Version:  "15.0",
			Tier:     "standard",
			Capacity: "20GB",
			Plan:     "prod",
			ProviderConfig: map[string]interface{}{
				"config_ref": "config/checkout/postgres-db",
			},
		},
	}
	checkoutPgNode, _ := graph.ResolveContract(checkoutPgInstance)
	global.AddNode(checkoutPgNode)

	checkoutRedisInstance := contracts.ResourceContract{
		Metadata: contracts.Metadata{
			Name:  "checkout-redis",
			Owner: app.Metadata.Owner,
		},
		Spec: contracts.ResourceSpec{
			Type:     "redis",
			Version:  "7.0",
			Tier:     "cache",
			Capacity: "2GB",
			Plan:     "prod",
			ProviderConfig: map[string]interface{}{
				"config_ref": "config/checkout/redis-cache",
			},
		},
	}
	checkoutRedisNode, _ := graph.ResolveContract(checkoutRedisInstance)
	global.AddNode(checkoutRedisNode)

	checkoutKafkaInstance := contracts.ResourceContract{
		Metadata: contracts.Metadata{
			Name:  "checkout-events",
			Owner: app.Metadata.Owner,
		},
		Spec: contracts.ResourceSpec{
			Type:     "kafka",
			Version:  "3.4",
			Tier:     "standard",
			Capacity: "15GB",
			Plan:     "prod",
			ProviderConfig: map[string]interface{}{
				"config_ref": "config/checkout/kafka-events",
			},
		},
	}
	checkoutKafkaNode, _ := graph.ResolveContract(checkoutKafkaInstance)
	global.AddNode(checkoutKafkaNode)

	// Create payment application's own Redis instance
	paymentRedisInstance := contracts.ResourceContract{
		Metadata: contracts.Metadata{
			Name:  "payment-redis",
			Owner: paymentApp.Metadata.Owner,
		},
		Spec: contracts.ResourceSpec{
			Type:     "redis",
			Version:  "7.0",
			Tier:     "persistent", // Different tier than checkout's Redis
			Capacity: "5GB",        // Different capacity than checkout's Redis
			Plan:     "prod",
			ProviderConfig: map[string]interface{}{
				"config_ref": "config/payment/redis-cache",
			},
		},
	}
	paymentRedisNode, _ := graph.ResolveContract(paymentRedisInstance)
	global.AddNode(paymentRedisNode)

	fmt.Println("🔗 Linking resource instances to application and services...")

	// Create "instance_of" relationships for resources
	global.AddEdge(checkoutPgInstance.Metadata.Name, postgresType.Metadata.Name, "instance_of")
	global.AddEdge(checkoutRedisInstance.Metadata.Name, redisType.Metadata.Name, "instance_of")
	global.AddEdge(checkoutKafkaInstance.Metadata.Name, kafkaType.Metadata.Name, "instance_of")
	global.AddEdge(paymentRedisInstance.Metadata.Name, redisType.Metadata.Name, "instance_of")

	// Application owns its resource instances
	global.AddEdge(app.Metadata.Name, checkoutPgInstance.Metadata.Name, "owns")
	global.AddEdge(app.Metadata.Name, checkoutRedisInstance.Metadata.Name, "owns")
	global.AddEdge(app.Metadata.Name, checkoutKafkaInstance.Metadata.Name, "owns")
	global.AddEdge(paymentApp.Metadata.Name, paymentRedisInstance.Metadata.Name, "owns")

	// Services use specific resource instances
	global.AddEdge(svc.Metadata.Name, checkoutPgInstance.Metadata.Name, "uses")
	global.AddEdge(svc.Metadata.Name, checkoutRedisInstance.Metadata.Name, "uses")
	global.AddEdge(workerSvc.Metadata.Name, checkoutPgInstance.Metadata.Name, "uses")
	global.AddEdge(workerSvc.Metadata.Name, checkoutKafkaInstance.Metadata.Name, "uses")
	global.AddEdge(paymentSvc.Metadata.Name, paymentRedisInstance.Metadata.Name, "uses")
}

// Helper function to create a ServiceVersionContract instance
func createServiceVersion(appName, serviceName, version string) contracts.ServiceVersionContract {
	return contracts.ServiceVersionContract{
		IDValue:   serviceName + ":" + version,
		Name:      serviceName,
		Owner:     "team-x", // Assuming team-x owns all services in this demo
		Version:   version,
		ConfigRef: "default-config",
		CreatedAt: time.Now(),
	}
}
