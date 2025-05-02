package graph

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type redisGraph struct {
	client *redis.Client
}

func NewRedisGraph(client *redis.Client) GraphBackend {
	return &redisGraph{client: client}
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
	// Simple scan implementation to fetch all nodes and edges
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
