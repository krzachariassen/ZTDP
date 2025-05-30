package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

const (
	baseURL    = "http://localhost:8080"
	redisDelay = 200 * time.Millisecond // Wait for Redis consistency
)

func main() {
	// Check for --force flag to start fresh
	forceCreate := false
	for _, arg := range os.Args {
		if arg == "--force" || arg == "-f" {
			forceCreate = true
			break
		}
	}

	// Wait a bit for Redis consistency if using Redis backend
	if os.Getenv("ZTDP_GRAPH_BACKEND") == "redis" {
		fmt.Println("⏳ Waiting for Redis consistency...")
		time.Sleep(redisDelay)
	}

	if forceCreate || !graphExists() {
		fmt.Println("🔄 Creating new graph via API...")
		createCompleteGraph()
	} else {
		fmt.Println("✅ Using existing graph from backend")
	}

	// Show the current graph state
	showGraphSummary()

	// Test deployment endpoints
	testDeployments()
}

// Check if graph already exists by trying to get an application
func graphExists() bool {
	resp, err := http.Get(baseURL + "/v1/applications/checkout")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// Create the complete graph structure using only API calls
func createCompleteGraph() {
	fmt.Println("📝 Creating applications...")
	createApplication("checkout", "team-x", "Handles checkout flows", []string{"payments"})
	createApplication("payment", "team-y", "Handles payment processing", []string{"payments", "financial"})

	// Redis consistency wait
	if os.Getenv("ZTDP_GRAPH_BACKEND") == "redis" {
		time.Sleep(redisDelay)
	}

	fmt.Println("🔧 Creating services...")
	createService("checkout", "checkout-api", 8080, true, "API service for checkout")
	createService("checkout", "checkout-worker", 9090, false, "Background worker for checkout")
	createService("payment", "payment-api", 8081, true, "API service for payment")

	// Redis consistency wait
	if os.Getenv("ZTDP_GRAPH_BACKEND") == "redis" {
		time.Sleep(redisDelay)
	}

	fmt.Println("🌍 Creating environments...")
	createEnvironment("dev", "platform-team", "Development environment")
	createEnvironment("prod", "platform-team", "Production environment")

	// Redis consistency wait
	if os.Getenv("ZTDP_GRAPH_BACKEND") == "redis" {
		time.Sleep(redisDelay)
	}

	fmt.Println("🔐 Setting up environment access...")
	addAllowedEnvironments("checkout", []string{"dev", "prod"})
	addAllowedEnvironments("payment", []string{"dev", "prod"})

	// Redis consistency wait
	if os.Getenv("ZTDP_GRAPH_BACKEND") == "redis" {
		time.Sleep(redisDelay)
	}

	fmt.Println("📦 Creating service versions...")
	createServiceVersion("checkout", "checkout-api", "1.0.0")
	createServiceVersion("checkout", "checkout-api", "2.0.0")
	createServiceVersion("checkout", "checkout-worker", "1.0.0")
	createServiceVersion("payment", "payment-api", "1.0.0")

	// Redis consistency wait
	if os.Getenv("ZTDP_GRAPH_BACKEND") == "redis" {
		time.Sleep(redisDelay)
	}

	fmt.Println("🏗️ Creating resource catalog...")
	createResourceCatalog()

	// Redis consistency wait
	if os.Getenv("ZTDP_GRAPH_BACKEND") == "redis" {
		time.Sleep(redisDelay)
	}

	fmt.Println("📚 Creating application resources...")
	createApplicationResources()

	fmt.Println("✅ Graph creation complete!")
}

func createApplication(name, owner, description string, tags []string) {
	app := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":  name,
			"owner": owner,
		},
		"spec": map[string]interface{}{
			"description": description,
			"tags":        tags,
			"lifecycle":   map[string]interface{}{},
		},
	}

	body, _ := json.Marshal(app)
	resp, err := http.Post(baseURL+"/v1/applications", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("❌ Failed to create application %s: %v\n", name, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		fmt.Printf("✅ Created application: %s\n", name)
	} else if resp.StatusCode == http.StatusConflict {
		fmt.Printf("ℹ️  Application already exists: %s\n", name)
	} else {
		fmt.Printf("⚠️  Failed to create application %s, status: %d\n", name, resp.StatusCode)
	}
}

func createService(appName, serviceName string, port int, public bool, description string) {
	service := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":  serviceName,
			"owner": "team-x", // Default owner
		},
		"spec": map[string]interface{}{
			"application": appName,
			"port":        port,
			"public":      public,
			"description": description,
			"tags":        []string{"test"},
		},
	}

	body, _ := json.Marshal(service)
	url := fmt.Sprintf("%s/v1/applications/%s/services", baseURL, appName)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("❌ Failed to create service %s: %v\n", serviceName, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		fmt.Printf("✅ Created service: %s\n", serviceName)
	} else if resp.StatusCode == http.StatusConflict {
		fmt.Printf("ℹ️  Service already exists: %s\n", serviceName)
	} else {
		fmt.Printf("⚠️  Failed to create service %s, status: %d\n", serviceName, resp.StatusCode)
	}
}

func createEnvironment(name, owner, description string) {
	env := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":  name,
			"owner": owner,
		},
		"spec": map[string]interface{}{
			"description": description,
		},
	}

	body, _ := json.Marshal(env)
	resp, err := http.Post(baseURL+"/v1/environments", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("❌ Failed to create environment %s: %v\n", name, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		fmt.Printf("✅ Created environment: %s\n", name)
	} else if resp.StatusCode == http.StatusConflict {
		fmt.Printf("ℹ️  Environment already exists: %s\n", name)
	} else {
		fmt.Printf("⚠️  Failed to create environment %s, status: %d\n", name, resp.StatusCode)
	}
}

func addAllowedEnvironments(appName string, envs []string) {
	for _, env := range envs {
		url := fmt.Sprintf("%s/v1/applications/%s/environments/%s/allowed", baseURL, appName, env)
		resp, err := http.Post(url, "application/json", nil)
		if err != nil {
			fmt.Printf("❌ Failed to add environment %s to app %s: %v\n", env, appName, err)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode == http.StatusCreated {
			fmt.Printf("✅ Added %s environment access to %s\n", env, appName)
		} else if resp.StatusCode == http.StatusConflict {
			fmt.Printf("ℹ️  Environment access already exists: %s -> %s\n", appName, env)
		} else {
			fmt.Printf("⚠️  Failed to add environment access %s -> %s, status: %d\n", appName, env, resp.StatusCode)
		}
	}
}

func createServiceVersion(appName, serviceName, version string) {
	serviceVersion := map[string]interface{}{
		"version":    version,
		"config_ref": "default-config",
	}

	body, _ := json.Marshal(serviceVersion)
	url := fmt.Sprintf("%s/v1/applications/%s/services/%s/versions", baseURL, appName, serviceName)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("❌ Failed to create service version %s:%s: %v\n", serviceName, version, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		fmt.Printf("✅ Created service version: %s:%s\n", serviceName, version)
	} else if resp.StatusCode == http.StatusOK {
		fmt.Printf("ℹ️  Service version already exists: %s:%s\n", serviceName, version)
	} else {
		fmt.Printf("⚠️  Failed to create service version %s:%s, status: %d\n", serviceName, version, resp.StatusCode)
	}
}

func createResourceCatalog() {
	// Create PostgreSQL resource type
	createResourceType("postgres", "platform-team", "15.0", []string{"standard", "high-memory", "high-cpu"}, "PostgreSQL database service")

	// Create Redis resource type
	createResourceType("redis", "platform-team", "7.0", []string{"cache", "persistent"}, "Redis in-memory cache service")

	// Create Kafka resource type
	createResourceType("kafka", "platform-team", "3.4", []string{"standard", "high-throughput"}, "Kafka streaming platform")
}

func createResourceType(name, owner, version string, tiers []string, description string) {
	resourceType := map[string]interface{}{
		"kind": "resource_type",
		"metadata": map[string]interface{}{
			"name":  name,
			"owner": owner,
		},
		"spec": map[string]interface{}{
			"version":          version,
			"tier_options":     tiers,
			"default_tier":     tiers[0],
			"config_template":  fmt.Sprintf("config/templates/%s-config.yaml", name),
			"available_plans":  []string{"dev", "prod"},
			"default_capacity": "10GB",
			"provider_metadata": map[string]interface{}{
				"description": description,
			},
		},
	}

	body, _ := json.Marshal(resourceType)
	resp, err := http.Post(baseURL+"/v1/resources", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("❌ Failed to create resource type %s: %v\n", name, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		fmt.Printf("✅ Created resource type: %s\n", name)
	} else if resp.StatusCode == http.StatusConflict {
		fmt.Printf("ℹ️  Resource type already exists: %s\n", name)
	} else {
		fmt.Printf("⚠️  Failed to create resource type %s, status: %d\n", name, resp.StatusCode)
	}
}

func createApplicationResources() {
	// Create generic catalog resources (reusable templates)
	createResource("pg-db", "platform-team", "postgres", "15.0", "standard", "20GB", "prod")
	createResource("redis-cache", "platform-team", "redis", "7.0", "cache", "2GB", "prod")
	createResource("event-bus", "platform-team", "kafka", "3.4", "standard", "15GB", "prod")
	createResource("redis-persistent", "platform-team", "redis", "7.0", "persistent", "5GB", "prod")

	// Link catalog resources to applications (this creates app-specific instances)
	linkResourceToApplication("checkout", "pg-db")
	linkResourceToApplication("checkout", "redis-cache")
	linkResourceToApplication("checkout", "event-bus")
	linkResourceToApplication("payment", "redis-persistent")

	// Redis consistency wait
	if os.Getenv("ZTDP_GRAPH_BACKEND") == "redis" {
		time.Sleep(redisDelay)
	}

	// Link services to their resources (using catalog resource names, not app-specific names)
	linkServiceToResource("checkout", "checkout-api", "pg-db")
	linkServiceToResource("checkout", "checkout-api", "redis-cache")
	linkServiceToResource("checkout", "checkout-worker", "pg-db")
	linkServiceToResource("checkout", "checkout-worker", "event-bus")
	linkServiceToResource("payment", "payment-api", "redis-persistent")
}

func createResource(name, owner, resourceType, version, tier, capacity, plan string) {
	resource := map[string]interface{}{
		"kind": "resource",
		"metadata": map[string]interface{}{
			"name":  name,
			"owner": owner,
		},
		"spec": map[string]interface{}{
			"type":     resourceType,
			"version":  version,
			"tier":     tier,
			"capacity": capacity,
			"plan":     plan,
			"provider_config": map[string]interface{}{
				"config_ref": fmt.Sprintf("config/%s/%s", owner, name),
			},
		},
	}

	body, _ := json.Marshal(resource)
	resp, err := http.Post(baseURL+"/v1/resources", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("❌ Failed to create resource %s: %v\n", name, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		fmt.Printf("✅ Created resource: %s\n", name)
	} else if resp.StatusCode == http.StatusConflict {
		fmt.Printf("ℹ️  Resource already exists: %s\n", name)
	} else {
		fmt.Printf("⚠️  Failed to create resource %s, status: %d\n", name, resp.StatusCode)
	}
}

func linkResourceToApplication(appName, resourceName string) {
	url := fmt.Sprintf("%s/v1/applications/%s/resources/%s", baseURL, appName, resourceName)
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		fmt.Printf("❌ Failed to link resource %s to app %s: %v\n", resourceName, appName, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		fmt.Printf("✅ Linked resource %s to application %s\n", resourceName, appName)
	} else if resp.StatusCode == http.StatusOK {
		fmt.Printf("ℹ️  Resource already linked: %s -> %s\n", appName, resourceName)
	} else if resp.StatusCode == http.StatusConflict {
		fmt.Printf("ℹ️  Resource already linked: %s -> %s\n", appName, resourceName)
	} else {
		fmt.Printf("⚠️  Failed to link resource %s to app %s, status: %d\n", resourceName, appName, resp.StatusCode)
	}
}

func linkServiceToResource(appName, serviceName, resourceName string) {
	url := fmt.Sprintf("%s/v1/applications/%s/services/%s/resources/%s", baseURL, appName, serviceName, resourceName)
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		fmt.Printf("❌ Failed to link service %s to resource %s: %v\n", serviceName, resourceName, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		fmt.Printf("✅ Linked service %s to resource %s\n", serviceName, resourceName)
	} else if resp.StatusCode == http.StatusOK {
		fmt.Printf("ℹ️  Service already linked to resource: %s -> %s\n", serviceName, resourceName)
	} else if resp.StatusCode == http.StatusConflict {
		fmt.Printf("ℹ️  Service already linked to resource: %s -> %s\n", serviceName, resourceName)
	} else {
		fmt.Printf("⚠️  Failed to link service %s to resource %s, status: %d\n", serviceName, resourceName, resp.StatusCode)
	}
}

func showGraphSummary() {
	fmt.Println("\n📋 Graph Summary:")

	// Get applications
	resp, err := http.Get(baseURL + "/v1/applications")
	if err != nil {
		fmt.Printf("❌ Failed to get applications: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var apps []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&apps)
	fmt.Printf("   Applications: %d\n", len(apps))
	for _, app := range apps {
		if metadata, ok := app["metadata"].(map[string]interface{}); ok {
			fmt.Printf("     - %s\n", metadata["name"])
		}
	}

	// Get environments
	resp2, err := http.Get(baseURL + "/v1/environments")
	if err != nil {
		fmt.Printf("❌ Failed to get environments: %v\n", err)
		return
	}
	defer resp2.Body.Close()

	var envs []map[string]interface{}
	json.NewDecoder(resp2.Body).Decode(&envs)
	fmt.Printf("   Environments: %d\n", len(envs))
	for _, env := range envs {
		if metadata, ok := env["metadata"].(map[string]interface{}); ok {
			fmt.Printf("     - %s\n", metadata["name"])
		}
	}

	// Get services for checkout
	resp3, err := http.Get(baseURL + "/v1/applications/checkout/services")
	if err == nil {
		defer resp3.Body.Close()
		var services []map[string]interface{}
		json.NewDecoder(resp3.Body).Decode(&services)
		fmt.Printf("   Checkout Services: %d\n", len(services))
		for _, svc := range services {
			if metadata, ok := svc["metadata"].(map[string]interface{}); ok {
				fmt.Printf("     - %s\n", metadata["name"])
			}
		}
	}
}

func testDeployments() {
	fmt.Println("\n🚀 Testing Application Deployment...")

	// Redis consistency wait before deployment
	if os.Getenv("ZTDP_GRAPH_BACKEND") == "redis" {
		fmt.Println("⏳ Waiting for Redis consistency before deployment...")
		time.Sleep(redisDelay * 3) // Longer wait for deployment
	}

	// Test deployment to dev first
	fmt.Println("\n📝 Testing deployment to dev...")
	deployApplication("checkout", "dev")

	// Redis consistency wait between deployments
	if os.Getenv("ZTDP_GRAPH_BACKEND") == "redis" {
		time.Sleep(redisDelay)
	}

	// Test deployment to prod (should work due to policy)
	fmt.Println("\n📝 Testing deployment to prod...")
	deployApplication("checkout", "prod")

	fmt.Println("\n✅ Deployment testing complete!")
}

func deployApplication(appName, environment string) {
	payload := map[string]interface{}{
		"environment": environment,
	}

	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/v1/applications/%s/deploy", baseURL, appName)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("❌ Failed to deploy %s to %s: %v\n", appName, environment, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			fmt.Printf("❌ Failed to decode deployment response: %v\n", err)
			return
		}
		fmt.Printf("✅ Successfully deployed %s to %s\n", result["application"], result["environment"])
		if summary, ok := result["summary"].(map[string]interface{}); ok {
			if deployed, ok := summary["deployed"].(float64); ok {
				fmt.Printf("   📊 Services deployed: %.0f\n", deployed)
			}
		}
	} else {
		fmt.Printf("⚠️  Deployment failed with status %d\n", resp.StatusCode)

		// Try to decode error response
		var errorResult map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorResult); err == nil {
			if message, ok := errorResult["error"].(string); ok {
				fmt.Printf("   📝 Error: %s\n", message)
			}
		}
	}
}
