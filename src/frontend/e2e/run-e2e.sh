#!/bin/bash
#
# E2E 自动化测试启动脚本
#
# 使用方法：
#   ./e2e/run-e2e.sh              # 运行所有 E2E 测试
#   ./e2e/run-e2e.sh student      # 只运行学生流程
#   ./e2e/run-e2e.sh teacher      # 只运行教师流程
#

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
FRONTEND_DIR="$(dirname "$SCRIPT_DIR")"
PROJECT_ROOT="$(dirname "$(dirname "$FRONTEND_DIR")")"

echo "========================================="
echo "  AI 数字分身 - E2E 自动化测试"
echo "========================================="
echo ""

# 1. 检查微信开发者工具 CLI
CLI_PATH="/Applications/wechatwebdevtools.app/Contents/MacOS/cli"
if [ ! -f "$CLI_PATH" ]; then
  echo "❌ 未找到微信开发者工具 CLI: $CLI_PATH"
  echo "   请确认微信开发者工具已安装"
  exit 1
fi
echo "✅ 微信开发者工具 CLI 已找到"

# 2. 检查后端服务是否运行
if ! lsof -i :8080 -sTCP:LISTEN > /dev/null 2>&1; then
  echo "⚠️  后端服务未运行，正在启动..."
  cd "$PROJECT_ROOT"
  WX_MODE=mock LLM_MODE=mock go run src/cmd/server/main.go &
  BACKEND_PID=$!
  echo "   后端 PID: $BACKEND_PID"
  sleep 3
  if ! lsof -i :8080 -sTCP:LISTEN > /dev/null 2>&1; then
    echo "❌ 后端服务启动失败"
    exit 1
  fi
  echo "✅ 后端服务已启动"
else
  echo "✅ 后端服务已在运行"
fi

# 3. 用 curl 准备测试数据（注册教师 + 添加文档）
echo ""
echo "📦 准备测试数据..."

TEACHER_TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/wx-login \
  -H "Content-Type: application/json" \
  -d '{"code":"e2e_teacher_setup"}' | python3 -c "import sys,json; print(json.load(sys.stdin).get('data',{}).get('token',''))")

if [ -n "$TEACHER_TOKEN" ]; then
  # 补全教师信息（complete-profile 会返回包含角色的新 token）
  COMPLETE_RESP=$(curl -s -X POST http://localhost:8080/api/auth/complete-profile \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TEACHER_TOKEN" \
    -d '{"role":"teacher","nickname":"E2E预置教师","school":"E2E测试学校","description":"E2E预置教师描述"}')

  # 从 complete-profile 响应中提取新 token（不重新登录，模拟真实用户行为）
  NEW_TOKEN=$(echo "$COMPLETE_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('data',{}).get('token',''))" 2>/dev/null)
  if [ -n "$NEW_TOKEN" ]; then
    TEACHER_TOKEN="$NEW_TOKEN"
  fi

  # 添加测试文档
  curl -s -X POST http://localhost:8080/api/documents \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TEACHER_TOKEN" \
    -d '{"title":"E2E测试知识","content":"Python是一种高级编程语言。Go是Google开发的编程语言。JavaScript是Web开发的核心语言。","tags":"编程,测试"}' > /dev/null 2>&1

  echo "✅ 测试数据已准备"
else
  echo "⚠️  准备测试数据失败（后端可能未正常响应）"
fi

# 4. 编译小程序
echo ""
echo "🔨 编译小程序..."
cd "$FRONTEND_DIR"
npm run build:weapp 2>&1 | tail -3
echo "✅ 小程序编译完成"

# 5. 运行 E2E 测试
echo ""
echo "🚀 开始运行 E2E 测试..."
echo ""

TEST_FILE=""
case "${1:-all}" in
  student)
    TEST_FILE="e2e/student-flow.test.js"
    echo "📋 运行学生流程测试"
    ;;
  teacher)
    TEST_FILE="e2e/teacher-flow.test.js"
    echo "📋 运行教师流程测试"
    ;;
  all|*)
    TEST_FILE="e2e/"
    echo "📋 运行所有 E2E 测试"
    ;;
esac

npx jest --config e2e/jest.config.js "$TEST_FILE" --forceExit

echo ""
echo "========================================="
echo "  E2E 测试完成"
echo "========================================="
