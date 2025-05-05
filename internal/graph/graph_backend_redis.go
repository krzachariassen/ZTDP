package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisGraph struct {
	client *redis.Client
}

type RedisGraphConfig struct {
	Addr     string
	Password string
}

func NewRedisGraph(cfg RedisGraphConfig) GraphBackend {
	addr := cfg.Addr
	if addr == "" {
		addr = os.Getenv("REDIS_HOST")
	}

	password := cfg.Password
	if password == "" {
		password = os.Getenv("REDIS_PASSWORD")
	}

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
	})

	ctx := context.Background()
	var err error
	fmt.Println("⚙️  Using Redis Backend:", client.Options().Addr)
	for i := 0; i < 3; i++ {
		err = client.Ping(ctx).Err()
		if err == nil {
			return &redisGraph{client: client}
		}
		time.Sleep(2 * time.Second)
	}

	panic(fmt.Errorf("failed to connect to Redis after 3 attempts: %w", err))
}

func redisKeyNode(env, id string) string {
	return fmt.Sprintf("ztgp:%s:node:%s", env, id)
}

func redisKeyEdges(env, fromID string) string {
	return fmt.Sprintf("ztgp:%s:edges:%s", env, fromID)
}

func (r *redisGraph) AddNode(env string, node *Node) error {
	data, err := json.Marshal(node)
	if err != nil {
		return fmt.Errorf("marshal node: %w", err)
	}
	return r.client.Set(context.Background(), redisKeyNode(env, node.ID), data, 0).Err()
}

func (r *redisGraph) GetNode(env, id string) (*Node, error) {
	data, err := r.client.Get(context.Background(), redisKeyNode(env, id)).Bytes()
	if err != nil {
		return nil, fmt.Errorf("get node: %w", err)
	}
	var node Node
	if err := json.Unmarshal(data, &node); err != nil {
		return nil, fmt.Errorf("unmarshal node: %w", err)
	}
	return &node, nil
}

func (r *redisGraph) AddEdge(env, fromID, toID string) error {
	return r.client.SAdd(context.Background(), redisKeyEdges(env, fromID), toID).Err()
}

func (r *redisGraph) GetAll(env string) (*Graph, error) {
	ctx := context.Background()
	graph := NewGraph()

	iter := r.client.Scan(ctx, 0, fmt.Sprintf("ztgp:%s:node:*", env), 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		data, err := r.client.Get(ctx, key).Bytes()
		if err != nil {
			continue
		}
		var node Node
		if err := json.Unmarshal(data, &node); err == nil {
			graph.Nodes[node.ID] = &node
		}
	}

	for id := range graph.Nodes {
		edgeKey := redisKeyEdges(env, id)
		edges, err := r.client.SMembers(ctx, edgeKey).Result()
		if err != nil {
			continue
		}
		graph.Edges[id] = edges
	}

	return graph, nil
}

// Global graph persistence
func (r *redisGraph) SaveGlobal(g *Graph) error {
	data, err := json.Marshal(g)
	if err != nil {
		return fmt.Errorf("marshal global graph: %w", err)
	}
	return r.client.Set(context.Background(), "ztgp:graph:global", data, 0).Err()
}

func (r *redisGraph) LoadGlobal() (*Graph, error) {
	data, err := r.client.Get(context.Background(), "ztgp:graph:global").Bytes()
	if err != nil {
		return nil, fmt.Errorf("get global graph: %w", err)
	}
	var graph Graph
	if err := json.Unmarshal(data, &graph); err != nil {
		return nil, fmt.Errorf("unmarshal global graph: %w", err)
	}
	return &graph, nil
}
