#!/bin/bash

# V2.0 迭代9 冒烟测试运行脚本

set -e

echo "=========================================="
echo "  V2.0 迭代9 冒烟测试"
echo "=========================================="
echo ""

# 进入前端目录
cd "$(dirname "$0")/.."

# 检查环境
echo "🔍 检查环境..."

# 检查微信开发者工具是否运行
if ! pgrep -x "wechatdevtools" > /dev/null; then
    echo "❌ 微信开发者工具未运行"
    echo "请先启动微信开发者工具"
    exit 1
fi
echo "✅ 微信开发者工具已运行"

# 检查后端服务
if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "⚠️  后端服务可能未启动（将继续执行）"
else
    echo "✅ 后端服务已启动"
fi

# 创建截图目录
mkdir -p e2e/screenshots-iter9

# 运行测试
echo ""
echo "🚀 开始执行测试..."
echo ""

NODE_ENV=test node e2e/iteration9-smoke.test.js

echo ""
echo "=========================================="
echo "  测试完成"
echo "=========================================="
