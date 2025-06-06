docker-compose up -d

export ZTDP_GRAPH_BACKEND=redis
export REDIS_PASSWORD=BVogb1sEPqA
export REDIS_HOST=localhost:6379
export OPENAI_API_KEY=

redis-cli -h localhost -a BVogb1sEPqA FLUSHALL
go run ./cmd/api

go run ./test/controlplane/graph_demo_api.go

