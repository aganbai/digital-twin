#!/bin/bash
# LlamaIndex 语义检索服务启动脚本
# 用法: bash scripts/start-knowledge-service.sh [--bg]
#   --bg: 后台运行模式

set -e

SCRIPT_DIR="$(dirname "$0")"
PROJECT_ROOT="$(realpath "$SCRIPT_DIR/..")"
SERVICE_DIR="$PROJECT_ROOT/src/knowledge-service"
VENV_DIR="$SERVICE_DIR/.venv"
LOG_FILE="/tmp/llamaindexer.log"

# 加载环境变量
if [ -f "$PROJECT_ROOT/.env" ]; then
    source "$PROJECT_ROOT/.env"
    echo "✅ 已加载 .env 配置"
fi

# 检查虚拟环境
if [ ! -d "$VENV_DIR" ]; then
    echo "❌ 虚拟环境不存在: $VENV_DIR"
    echo "   请先创建: python3.11 -m venv $VENV_DIR"
    echo "   然后安装依赖: $VENV_DIR/bin/pip install -r $SERVICE_DIR/requirements.txt"
    exit 1
fi

PYTHON="$VENV_DIR/bin/python"
PYTHON_VERSION=$($PYTHON --version 2>&1)
echo "🐍 Python 版本: $PYTHON_VERSION"

# 检查端口是否被占用
PORT="${KNOWLEDGE_SERVICE_PORT:-8100}"
if lsof -ti:$PORT > /dev/null 2>&1; then
    echo "⚠️  端口 $PORT 已被占用，正在停止旧进程..."
    lsof -ti:$PORT | xargs kill -9 2>/dev/null
    sleep 1
fi

echo "🚀 启动 LlamaIndex 语义检索服务 (端口: $PORT)..."

if [ "$1" = "--bg" ]; then
    # 后台运行
    nohup $PYTHON -m uvicorn app.main:app \
        --host 0.0.0.0 \
        --port $PORT \
        --log-level info \
        > "$LOG_FILE" 2>&1 &
    PID=$!
    sleep 3
    
    if curl -s "http://localhost:$PORT/api/v1/health" > /dev/null 2>&1; then
        echo "✅ 服务已启动 (PID: $PID)"
        echo "   健康检查: curl http://localhost:$PORT/api/v1/health"
        echo "   日志文件: $LOG_FILE"
    else
        echo "❌ 服务启动失败，查看日志:"
        cat "$LOG_FILE"
        exit 1
    fi
else
    # 前台运行
    exec $PYTHON -m uvicorn app.main:app \
        --host 0.0.0.0 \
        --port $PORT \
        --log-level info \
        --reload
fi
