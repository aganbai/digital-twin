#!/bin/bash
set -e

echo "=== Starting deployment ==="

# 1. 检查 Docker 环境
if ! command -v docker &> /dev/null; then
    echo "Error: Docker is not installed."
    exit 1
fi

if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo "Error: Docker Compose is not installed."
    exit 1
fi

# 2. 检查 .env.production 环境变量
if [ ! -f ".env.production" ]; then
    echo "Error: .env.production file is missing in the project root."
    exit 1
fi

REQUIRED_VARS=("JWT_SECRET" "OPENAI_API_KEY")
for var in "${REQUIRED_VARS[@]}"; do
    if ! grep -q "^${var}=" .env.production; then
        echo "Error: Required variable ${var} is missing or empty in .env.production."
        exit 1
    fi
done

# 获取正确的 docker compose 命令
if docker compose version &> /dev/null; then
    DOCKER_COMPOSE_CMD="docker compose"
else
    DOCKER_COMPOSE_CMD="docker-compose"
fi

echo "Environment checks passed."

# 3. 构建镜像并启动服务
echo "Building and starting services..."
$DOCKER_COMPOSE_CMD up -d --build

# 4. 等待服务就绪
echo "Waiting for services to be ready..."
# 因为 nginx 依赖 backend (service_healthy) 并且 backend 依赖 knowledge (service_healthy)
# Nginx 容器运行说明所有健康检查已通过
for i in {1..30}; do
    # Docker compose v2 命令支持
    NGINX_STATUS=$($DOCKER_COMPOSE_CMD ps --status running --services | grep -w "nginx" || true)
    if [ ! -z "$NGINX_STATUS" ]; then
        echo "All services are up and running successfully!"
        exit 0
    fi
    echo "Waiting... ($i/30)"
    sleep 5
done

echo "Warning: Timeout waiting for services to be fully ready."
echo "Please check logs with: $DOCKER_COMPOSE_CMD logs"
exit 1
