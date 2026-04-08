#!/bin/bash
# Digital Twin 开发环境启动脚本
# 使用方式: ./scripts/dev_start.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "🚀 Digital Twin 开发环境启动..."
echo "================================"

# 1. 加载环境变量
if [ -f "$PROJECT_DIR/.env" ]; then
    echo "📦 加载环境变量..."
    source "$PROJECT_DIR/.env"
    echo "   ✅ JWT_SECRET: 已设置 (${#JWT_SECRET} 字符)"
    echo "   ✅ APP_ENV: $APP_ENV"
    
    # LLM 模式检查
    if [ "${LLM_MODE:-mock}" = "mock" ]; then
        echo "   ℹ️  LLM 模式: mock（预设回复，无需 API Key）"
    else
        echo "   ✅ LLM 模式: api"
        echo "   ✅ LLM 模型: ${LLM_MODEL:-qwen-turbo}"
        echo "   ✅ API 地址: ${OPENAI_BASE_URL:-https://dashscope.aliyuncs.com/compatible-mode/v1}"
        if [ "$OPENAI_API_KEY" = "sk-your-dashscope-api-key-here" ] || [ -z "$OPENAI_API_KEY" ]; then
            echo "   ⚠️  OPENAI_API_KEY 未配置，请到阿里云百炼平台获取"
            echo "   获取地址: https://bailian.console.aliyun.com/"
            echo "   自动回退到 mock 模式..."
            export LLM_MODE="mock"
        else
            echo "   ✅ API Key: ${OPENAI_API_KEY:0:10}..."
        fi
    fi
else
    echo "   ❌ .env 文件不存在，请先创建"
    echo "   参考: cp .env.example .env"
    exit 1
fi

# 2. 检查 Go 环境
echo ""
echo "🔍 检查 Go 环境..."
if command -v go &> /dev/null; then
    echo "   ✅ Go: $(go version)"
else
    echo "   ❌ Go 未安装"
    exit 1
fi

# 3. 检查向量数据库
echo ""
echo "🔍 检查向量数据库..."
if [ "${VECTOR_DB_MODE:-memory}" = "memory" ]; then
    echo "   ℹ️  使用内存向量存储（开发模式）"
    echo "   提示: 设置 VECTOR_DB_MODE=chroma 可切换到 Chroma DB"
else
    if curl -s --connect-timeout 2 http://${CHROMA_HOST:-localhost}:${CHROMA_PORT:-8000}/api/v1/heartbeat > /dev/null 2>&1; then
        echo "   ✅ Chroma DB 运行中"
    else
        echo "   ⚠️  Chroma DB 未运行，回退到内存模式"
        export VECTOR_DB_MODE="memory"
    fi
fi

# 4. 创建数据目录
echo ""
echo "📁 检查数据目录..."
mkdir -p "$PROJECT_DIR/data/sqlite"
mkdir -p "$PROJECT_DIR/data/uploads"
mkdir -p "$PROJECT_DIR/data/chroma"
echo "   ✅ 数据目录就绪"

# 5. 安装 Go 依赖
echo ""
echo "📦 检查 Go 依赖..."
GOPROXY=https://goproxy.cn,direct go mod tidy 2>&1
echo "   ✅ Go 依赖就绪"

# 6. 编译并启动
echo ""
echo "🔨 编译项目..."
go build -o "$PROJECT_DIR/bin/digital-twin" "$PROJECT_DIR/src/cmd/server/main.go" 2>&1
echo "   ✅ 编译成功"

echo ""
echo "================================"
echo "🎉 启动服务 (端口: ${SERVER_PORT:-8080})..."
echo "   健康检查: http://localhost:${SERVER_PORT:-8080}/health"
echo "   API 文档: http://localhost:${SERVER_PORT:-8080}/api/v1/"
echo "================================"
echo ""

exec "$PROJECT_DIR/bin/digital-twin"
