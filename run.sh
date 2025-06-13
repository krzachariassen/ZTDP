docker-compose up -d

# Load environment variables from .env file if it exists
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Set defaults if not already set
export ZTDP_GRAPH_BACKEND=${ZTDP_GRAPH_BACKEND:-redis}
export REDIS_HOST=${REDIS_HOST:-localhost:6379}

# Validate required environment variables
if [ -z "$REDIS_PASSWORD" ]; then
    echo "❌ REDIS_PASSWORD environment variable is required"
    echo "💡 Create a .env file with: REDIS_PASSWORD=your_redis_password"
    exit 1
fi

if [ -z "$OPENAI_API_KEY" ]; then
    echo "⚠️  OPENAI_API_KEY not set - AI features will be disabled"
fi

redis-cli -h localhost -a "$REDIS_PASSWORD" FLUSHALL
go run ./cmd/api

go run ./test/controlplane/graph_demo_api.go

