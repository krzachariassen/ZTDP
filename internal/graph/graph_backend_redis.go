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

// Global graph persistence - the only storage mechanism
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

// Clear removes all global data (useful for testing)
func (r *redisGraph) Clear() error {
	ctx := context.Background()
	return r.client.Del(ctx, "ztgp:graph:global").Err()
}
