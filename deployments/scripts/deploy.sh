#!/bin/bash
# 一键部署脚本
# 用法: ./deploy.sh [up|down|restart|logs|status]

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
COMPOSE_DIR="$SCRIPT_DIR/.."

# 检查 Docker 环境
check_docker() {
    if ! command -v docker &> /dev/null; then
        echo "❌ Docker 未安装"
        exit 1
    fi
    if ! docker compose version &> /dev/null; then
        echo "❌ Docker Compose 未安装"
        exit 1
    fi
    echo "✅ Docker 环境就绪"
}

# 检查环境变量
check_env() {
    ENV_FILE="$COMPOSE_DIR/.env"
    if [ ! -f "$ENV_FILE" ]; then
        echo "⚠️ .env 文件不存在，从模板创建..."
        cp "$COMPOSE_DIR/.env.production" "$ENV_FILE"
        echo "📝 请编辑 $ENV_FILE 填写必要的配置"
        exit 1
    fi
    # 检查必要变量
    source "$ENV_FILE"
    if [ "$JWT_SECRET" = "change-me-to-a-random-string" ] || [ -z "$JWT_SECRET" ]; then
        echo "❌ 请修改 JWT_SECRET"
        exit 1
    fi
    if [ -z "$OPENAI_API_KEY" ] || [ "$OPENAI_API_KEY" = "your-api-key" ]; then
        echo "❌ 请设置 OPENAI_API_KEY"
        exit 1
    fi
    echo "✅ 环境变量检查通过"
}

# 启动服务
up() {
    check_docker
    check_env
    echo "🚀 构建并启动服务..."
    cd "$COMPOSE_DIR"
    docker compose up -d --build
    echo "⏳ 等待服务就绪..."
    sleep 10
    status
}

# 停止服务
down() {
    cd "$COMPOSE_DIR"
    docker compose down
    echo "✅ 服务已停止"
}

# 重启服务
restart() {
    down
    up
}

# 查看日志
logs() {
    cd "$COMPOSE_DIR"
    docker compose logs -f --tail=100
}

# 查看状态
status() {
    cd "$COMPOSE_DIR"
    docker compose ps
    echo ""
    echo "🔍 健康检查:"
    curl -sf http://localhost:8080/health && echo " ✅ Backend OK" || echo " ❌ Backend 未就绪"
    curl -sf http://localhost:8100/api/v1/health && echo " ✅ Knowledge OK" || echo " ❌ Knowledge 未就绪"
    curl -sf http://localhost:80/health && echo " ✅ Nginx OK" || echo " ❌ Nginx 未就绪"
}

# 主入口
case "${1:-up}" in
    up)      up ;;
    down)    down ;;
    restart) restart ;;
    logs)    logs ;;
    status)  status ;;
    *)       echo "用法: $0 [up|down|restart|logs|status]" ;;
esac
