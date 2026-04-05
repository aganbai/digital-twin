#!/bin/bash

#######################################################################
# H5教师端登录功能 E2E 测试脚本
# 测试范围：
#   1. 后端API服务状态
#   2. H5前端服务状态
#   3. 微信授权登录URL获取
#   4. Mock登录回调流程
# 输出文件：h5-teacher-login-test-output.log
#######################################################################

# 配置
BACKEND_URL="http://localhost:8080"
H5_URL="http://localhost:5175"
OUTPUT_FILE="/Users/aganbai/Desktop/WorkSpace/digital-twin/e2e-tests/h5-teacher-login-test-output.log"
BACKEND_PID=""
H5_PID=""

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 测试计数器
TEST_PASSED=0
TEST_FAILED=0
TEST_SKIPPED=0

# 清空输出文件
echo "" > "$OUTPUT_FILE"

# 日志函数
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$OUTPUT_FILE"
}

log_success() {
    echo -e "${GREEN}[✓]${NC} $1" | tee -a "$OUTPUT_FILE"
    TEST_PASSED=$((TEST_PASSED + 1))
}

log_error() {
    echo -e "${RED}[✗]${NC} $1" | tee -a "$OUTPUT_FILE"
    TEST_FAILED=$((TEST_FAILED + 1))
}

log_skip() {
    echo -e "${YELLOW}[⊘]${NC} $1" | tee -a "$OUTPUT_FILE"
    TEST_SKIPPED=$((TEST_SKIPPED + 1))
}

log_info() {
    echo -e "${BLUE}[i]${NC} $1" | tee -a "$OUTPUT_FILE"
}

log_section() {
    echo "" | tee -a "$OUTPUT_FILE"
    echo "========================================" | tee -a "$OUTPUT_FILE"
    echo "$1" | tee -a "$OUTPUT_FILE"
    echo "========================================" | tee -a "$OUTPUT_FILE"
}

# 清理函数
cleanup() {
    log_info "清理测试环境..."
    if [ ! -z "$BACKEND_PID" ]; then
        log_info "停止后端服务 (PID: $BACKEND_PID)..."
        kill $BACKEND_PID 2>/dev/null
        wait $BACKEND_PID 2>/dev/null
    fi
    if [ ! -z "$H5_PID" ]; then
        log_info "停止H5前端服务 (PID: $H5_PID)..."
        kill $H5_PID 2>/dev/null
        wait $H5_PID 2>/dev/null
    fi
}

# 注册清理函数
trap cleanup EXIT

#######################################################################
# 测试结果汇总
#######################################################################
print_summary() {
    log_section "测试结果汇总"
    echo -e "${GREEN}通过: $TEST_PASSED${NC}" | tee -a "$OUTPUT_FILE"
    echo -e "${RED}失败: $TEST_FAILED${NC}" | tee -a "$OUTPUT_FILE"
    echo -e "${YELLOW}跳过: $TEST_SKIPPED${NC}" | tee -a "$OUTPUT_FILE"
    echo "" | tee -a "$OUTPUT_FILE"
    
    TOTAL=$((TEST_PASSED + TEST_FAILED + TEST_SKIPPED))
    if [ $TEST_FAILED -eq 0 ]; then
        echo -e "${GREEN}所有测试通过! ($TEST_PASSED/$TOTAL)${NC}" | tee -a "$OUTPUT_FILE"
        return 0
    else
        echo -e "${RED}存在失败的测试! ($TEST_FAILED/$TOTAL 失败)${NC}" | tee -a "$OUTPUT_FILE"
        return 1
    fi
}

#######################################################################
# 步骤1: 检查后端服务状态
#######################################################################
log_section "步骤1: 检查后端服务状态 (端口 8080)"

# 检查端口是否被监听
if lsof -Pi :8080 -sTCP:LISTEN -t >/dev/null 2>&1; then
    log_success "后端服务正在运行 (端口 8080)"
    
    # 检查健康检查端点
    log_info "检查后端健康检查端点..."
    HEALTH_RESPONSE=$(curl -s -w "\n%{http_code}" "$BACKEND_URL/api/system/health" 2>&1)
    HTTP_CODE=$(echo "$HEALTH_RESPONSE" | tail -n1)
    BODY=$(echo "$HEALTH_RESPONSE" | sed '$d')
    
    if [ "$HTTP_CODE" = "200" ]; then
        log_success "后端健康检查通过"
        log_info "响应: $BODY"
    else
        log_error "后端健康检查失败 (HTTP $HTTP_CODE)"
        log_info "响应: $BODY"
    fi
else
    log_error "后端服务未运行 (端口 8080 未监听)"
    log_info "请先启动后端服务: cd src/backend && go run main.go"
    exit 1
fi

#######################################################################
# 步骤2: 检查H5前端服务状态
#######################################################################
log_section "步骤2: 检查H5前端服务状态 (端口 5175)"

# 检查端口是否被监听
if lsof -Pi :5175 -sTCP:LISTEN -t >/dev/null 2>&1; then
    log_success "H5前端服务正在运行 (端口 5175)"
    
    # 尝试访问H5前端
    log_info "访问H5前端首页..."
    H5_RESPONSE=$(curl -s -w "\n%{http_code}" "$H5_URL/" 2>&1)
    HTTP_CODE=$(echo "$H5_RESPONSE" | tail -n1)
    BODY=$(echo "$H5_RESPONSE" | sed '$d')
    
    if [ "$HTTP_CODE" = "200" ]; then
        log_success "H5前端页面可访问"
        # 检查是否包含Vue应用
        if echo "$BODY" | grep -q "div id=\"app\""; then
            log_success "H5前端Vue应用加载正常"
        else
            log_error "H5前端页面缺少Vue应用挂载点"
        fi
    else
        log_error "H5前端页面访问失败 (HTTP $HTTP_CODE)"
    fi
else
    log_error "H5前端服务未运行 (端口 5175 未监听)"
    log_info "请先启动H5前端服务: cd src/h5-teacher && npm run dev"
    log_info "注意: vite.config.ts 默认端口是 5174，需要改为 5175 或使用默认端口"
    exit 1
fi

#######################################################################
# 步骤3: 测试微信授权登录URL获取
#######################################################################
log_section "步骤3: 测试微信授权登录URL获取"

log_info "请求微信授权登录URL..."

# 构造回调URL
REDIRECT_URI="http://localhost:5175/login"

# 发送请求获取登录URL
LOGIN_URL_RESPONSE=$(curl -s -w "\n%{http_code}" \
    -X GET \
    -H "Content-Type: application/json" \
    -H "Origin: http://localhost:5175" \
    "$BACKEND_URL/api/auth/wx-h5-login-url?redirect_uri=$REDIRECT_URI" 2>&1)

HTTP_CODE=$(echo "$LOGIN_URL_RESPONSE" | tail -n1)
BODY=$(echo "$LOGIN_URL_RESPONSE" | sed '$d')

log_info "HTTP状态码: $HTTP_CODE"
log_info "响应内容: $BODY"

if [ "$HTTP_CODE" = "200" ]; then
    log_success "获取微信授权登录URL成功"
    
    # 检查响应是否包含login_url
    if echo "$BODY" | grep -q "login_url"; then
        log_success "响应包含login_url字段"
        
        # 提取登录URL（简单提取）
        LOGIN_URL=$(echo "$BODY" | grep -o '"login_url":"[^"]*"' | sed 's/"login_url":"//;s/"$//')
        if [ ! -z "$LOGIN_URL" ]; then
            log_info "登录URL: $LOGIN_URL"
            
            # 检查是否是mock模式的URL
            if echo "$LOGIN_URL" | grep -q "mock"; then
                log_success "检测到Mock模式登录URL"
            elif echo "$LOGIN_URL" | grep -q "open.weixin.qq.com"; then
                log_success "检测到真实微信授权URL"
            else
                log_info "登录URL格式: $LOGIN_URL"
            fi
        fi
    else
        log_error "响应缺少login_url字段"
    fi
    
    # 检查响应是否包含state
    if echo "$BODY" | grep -q "state"; then
        log_success "响应包含state字段"
    else
        log_error "响应缺少state字段"
    fi
else
    log_error "获取微信授权登录URL失败 (HTTP $HTTP_CODE)"
fi

#######################################################################
# 步骤4: 测试CORS配置
#######################################################################
log_section "步骤4: 测试CORS配置"

log_info "测试OPTIONS预检请求..."

OPTIONS_RESPONSE=$(curl -s -w "\n%{http_code}" \
    -X OPTIONS \
    -H "Origin: http://localhost:5175" \
    -H "Access-Control-Request-Method: GET" \
    -H "Access-Control-Request-Headers: Content-Type,Authorization,X-Platform" \
    "$BACKEND_URL/api/auth/wx-h5-login-url" 2>&1)

HTTP_CODE=$(echo "$OPTIONS_RESPONSE" | tail -n1)
BODY=$(echo "$OPTIONS_RESPONSE" | sed '$d')

log_info "OPTIONS响应状态码: $HTTP_CODE"

if [ "$HTTP_CODE" = "204" ] || [ "$HTTP_CODE" = "200" ]; then
    log_success "OPTIONS预检请求成功 (HTTP $HTTP_CODE)"
else
    log_error "OPTIONS预检请求失败 (HTTP $HTTP_CODE)"
fi

# 检查CORS头
HEADERS=$(curl -s -I \
    -X OPTIONS \
    -H "Origin: http://localhost:5175" \
    -H "Access-Control-Request-Method: GET" \
    "$BACKEND_URL/api/auth/wx-h5-login-url" 2>&1)

log_info "CORS响应头:"
echo "$HEADERS" | tee -a "$OUTPUT_FILE"

if echo "$HEADERS" | grep -qi "Access-Control-Allow-Origin"; then
    log_success "CORS: Access-Control-Allow-Origin 已设置"
else
    log_error "CORS: 缺少 Access-Control-Allow-Origin 头"
fi

if echo "$HEADERS" | grep -qi "Access-Control-Allow-Methods"; then
    log_success "CORS: Access-Control-Allow-Methods 已设置"
else
    log_error "CORS: 缺少 Access-Control-Allow-Methods 头"
fi

#######################################################################
# 步骤5: 测试Mock登录回调
#######################################################################
log_section "步骤5: 测试Mock登录回调"

log_info "测试微信H5登录回调接口..."

# 使用mock模式的code和state进行测试
MOCK_CODE="mock_code_$(date +%s)"
MOCK_STATE="mock_state_$(date +%s)"

CALLBACK_RESPONSE=$(curl -s -w "\n%{http_code}" \
    -X POST \
    -H "Content-Type: application/json" \
    -H "Origin: http://localhost:5175" \
    -d "{\"code\":\"$MOCK_CODE\",\"state\":\"$MOCK_STATE\"}" \
    "$BACKEND_URL/api/auth/wx-h5-callback" 2>&1)

HTTP_CODE=$(echo "$CALLBACK_RESPONSE" | tail -n1)
BODY=$(echo "$CALLBACK_RESPONSE" | sed '$d')

log_info "HTTP状态码: $HTTP_CODE"
log_info "响应内容: $BODY"

if [ "$HTTP_CODE" = "200" ]; then
    log_success "登录回调请求成功"
    
    # 检查必要的字段
    if echo "$BODY" | grep -q "token"; then
        log_success "响应包含token字段"
    else
        log_error "响应缺少token字段"
    fi
    
    if echo "$BODY" | grep -q "user_id"; then
        log_success "响应包含user_id字段"
    else
        log_error "响应缺少user_id字段"
    fi
    
    if echo "$BODY" | grep -q "role"; then
        log_success "响应包含role字段"
    else
        log_error "响应缺少role字段"
    fi
else
    log_error "登录回调请求失败 (HTTP $HTTP_CODE)"
    log_info "注意: 这可能是因为后端不是Mock模式，需要真实的微信授权"
fi

#######################################################################
# 步骤6: 测试登录页面资源加载
#######################################################################
log_section "步骤6: 测试登录页面资源加载"

log_info "访问登录页面..."
LOGIN_PAGE_RESPONSE=$(curl -s -w "\n%{http_code}" "$H5_URL/login" 2>&1)
HTTP_CODE=$(echo "$LOGIN_PAGE_RESPONSE" | tail -n1)
BODY=$(echo "$LOGIN_PAGE_RESPONSE" | sed '$d')

if [ "$HTTP_CODE" = "200" ]; then
    log_success "登录页面可访问 (HTTP 200)"
    
    # 检查页面内容
    if echo "$BODY" | grep -q "数字分身教师端"; then
        log_success "登录页面标题正确"
    else
        log_info "登录页面内容检查: 未找到预期标题"
    fi
    
    if echo "$BODY" | grep -q "微信授权登录"; then
        log_success "登录页面包含微信授权登录按钮文本"
    else
        log_info "登录页面内容检查: 未找到微信授权登录按钮文本"
    fi
else
    log_error "登录页面访问失败 (HTTP $HTTP_CODE)"
fi

#######################################################################
# 步骤7: 检查静态资源加载
#######################################################################
log_section "步骤7: 检查静态资源加载"

# 获取页面中的JS资源
log_info "检查JavaScript资源..."
JS_FILES=$(curl -s "$H5_URL/login" | grep -o 'src="[^"]*\.js"' | sed 's/src="//;s/"$//')

if [ ! -z "$JS_FILES" ]; then
    log_info "发现JS文件:"
    echo "$JS_FILES" | while read js_file; do
        log_info "  - $js_file"
    done
    
    # 测试第一个JS文件是否可访问
    FIRST_JS=$(echo "$JS_FILES" | head -n1)
    if [ ! -z "$FIRST_JS" ]; then
        # 处理相对路径
        if [[ "$FIRST_JS" == /* ]]; then
            JS_URL="${H5_URL}${FIRST_JS}"
        else
            JS_URL="${H5_URL}/${FIRST_JS}"
        fi
        
        JS_RESPONSE=$(curl -s -w "%{http_code}" -o /dev/null "$JS_URL" 2>&1)
        if [ "$JS_RESPONSE" = "200" ]; then
            log_success "JS资源可访问: $FIRST_JS"
        else
            log_error "JS资源访问失败: $FIRST_JS (HTTP $JS_RESPONSE)"
        fi
    fi
else
    log_info "未发现外部JS文件（可能是内联脚本）"
fi

#######################################################################
# 打印测试汇总
#######################################################################
print_summary
exit $?
