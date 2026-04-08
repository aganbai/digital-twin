#!/bin/bash
# 冒烟测试环境检查脚本
# 用途：检查 minium 和 playwright 环境，确保测试可以正常运行
# 用法：./check-smoke-env.sh [--fix] [--verbose]

set -e

# ==================== 配置区 ====================
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
LOG_FILE="/tmp/smoke-env-check-${TIMESTAMP}.log"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 检查结果
CHECK_RESULTS=()
PASS_COUNT=0
FAIL_COUNT=0
WARN_COUNT=0

# ==================== 日志函数 ====================
log() {
    echo -e "$@" | tee -a "$LOG_FILE"
}

log_info() {
    log "${BLUE}[INFO]${NC} $@"
}

log_success() {
    log "${GREEN}[✓]${NC} $@"
    PASS_COUNT=$((PASS_COUNT + 1))
    CHECK_RESULTS+=("✅ $@")
}

log_warn() {
    log "${YELLOW}[⚠]${NC} $@"
    WARN_COUNT=$((WARN_COUNT + 1))
    CHECK_RESULTS+=("⚠️ $@")
}

log_error() {
    log "${RED}[✗]${NC} $@"
    FAIL_COUNT=$((FAIL_COUNT + 1))
    CHECK_RESULTS+=("❌ $@")
}

log_section() {
    log ""
    log "${CYAN}========================================${NC}"
    log "${CYAN}$@${NC}"
    log "${CYAN}========================================${NC}"
}

# ==================== 参数解析 ====================
VERBOSE=false
AUTO_FIX=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --fix)
            AUTO_FIX=true
            shift
            ;;
        --verbose|-v)
            VERBOSE=true
            shift
            ;;
        -h|--help)
            echo "冒烟测试环境检查脚本"
            echo ""
            echo "用法: $0 [OPTIONS]"
            echo ""
            echo "选项:"
            echo "  --fix        自动修复可修复的问题"
            echo "  --verbose    显示详细输出"
            echo "  -h, --help   显示帮助信息"
            exit 0
            ;;
        *)
            log_error "未知参数: $1"
            exit 1
            ;;
    esac
done

# ==================== Python 环境检查 ====================
check_python() {
    log_section "Python 环境检查"
    
    # 检查 Python 3
    if command -v python3 &> /dev/null; then
        PYTHON_VERSION=$(python3 --version 2>&1)
        log_success "Python 已安装: $PYTHON_VERSION"
        
        # 检查 Python 版本 >= 3.8
        PYTHON_MAJOR=$(python3 -c 'import sys; print(sys.version_info.major)')
        PYTHON_MINOR=$(python3 -c 'import sys; print(sys.version_info.minor)')
        
        if [ "$PYTHON_MAJOR" -ge 3 ] && [ "$PYTHON_MINOR" -ge 8 ]; then
            log_success "Python 版本符合要求 (>= 3.8)"
        else
            log_error "Python 版本过低，需要 >= 3.8，当前: $PYTHON_VERSION"
        fi
    else
        log_error "Python 3 未安装"
        return 1
    fi
    
    # 检查 pip
    if command -v pip3 &> /dev/null || python3 -m pip --version &> /dev/null; then
        PIP_VERSION=$(python3 -m pip --version 2>&1 | cut -d' ' -f1-2)
        log_success "pip 已安装: $PIP_VERSION"
    else
        log_error "pip 未安装"
        if [ "$AUTO_FIX" = true ]; then
            log_info "正在安装 pip..."
            python3 -m ensurepip --upgrade
        fi
    fi
}

# ==================== Minium 环境检查 ====================
check_minium() {
    log_section "Minium 环境检查"
    
    # 检查 minium 包
    if python3 -c "import minium" 2>/dev/null; then
        MINIUM_VERSION=$(python3 -c "import minium; print(minium.__version__)" 2>/dev/null || echo "unknown")
        log_success "minium 包已安装: $MINIUM_VERSION"
    else
        log_error "minium 包未安装"
        if [ "$AUTO_FIX" = true ]; then
            log_info "正在安装 minium..."
            python3 -m pip install minium -i https://pypi.tuna.tsinghua.edu.cn/simple
            if python3 -c "import minium" 2>/dev/null; then
                log_success "minium 安装成功"
            else
                log_error "minium 安装失败"
            fi
        fi
        return 1
    fi
    
    # 检查 minium 依赖
    log_info "检查 minium 依赖..."
    
    MINIUM_DEPS=("requests" "pytest" "allure-pytest")
    for dep in "${MINIUM_DEPS[@]}"; do
        if python3 -c "import ${dep%%=*}" 2>/dev/null; then
            log_success "${dep} 已安装"
        else
            log_warn "${dep} 未安装"
            if [ "$AUTO_FIX" = true ]; then
                python3 -m pip install ${dep} -i https://pypi.tuna.tsinghua.edu.cn/simple
            fi
        fi
    done
    
    # 检查 Minium 连接能力
    log_info "测试 Minium 初始化能力..."
    python3 -c "
import sys
try:
    from minium import WXMinium
    print('Minium 导入成功')
except Exception as e:
    print(f'Minium 导入失败: {e}', file=sys.stderr)
    sys.exit(1)
" 2>&1 | tee -a "$LOG_FILE"
    
    if [ $? -eq 0 ]; then
        log_success "Minium 导入成功"
    else
        log_error "Minium 导入失败"
    fi
}

# ==================== Playwright 环境检查 ====================
check_playwright() {
    log_section "Playwright 环境检查"
    
    # 检查 playwright 包
    if python3 -c "import playwright" 2>/dev/null; then
        PLAYWRIGHT_VERSION=$(python3 -c "import playwright; print(playwright.__version__)" 2>/dev/null || echo "unknown")
        log_success "playwright 包已安装: $PLAYWRIGHT_VERSION"
    else
        log_error "playwright 包未安装"
        if [ "$AUTO_FIX" = true ]; then
            log_info "正在安装 playwright..."
            python3 -m pip install playwright -i https://pypi.tuna.tsinghua.edu.cn/simple
            if python3 -c "import playwright" 2>/dev/null; then
                log_success "playwright 安装成功"
            else
                log_error "playwright 安装失败"
            fi
        fi
        return 1
    fi
    
    # 检查 playwright 浏览器
    log_info "检查 Playwright 浏览器..."
    
    BROWSERS=("chromium" "firefox" "webkit")
    for browser in "${BROWSERS[@]}"; do
        if python3 -c "from playwright.sync_api import sync_playwright; p = sync_playwright().start(); b = p.$browser.launch(headless=True); b.close(); p.stop()" 2>/dev/null; then
            log_success "$browser 浏览器已安装"
        else
            log_warn "$browser 浏览器未安装"
        fi
    done
    
    # 至少需要 chromium
    if python3 -c "from playwright.sync_api import sync_playwright; p = sync_playwright().start(); b = p.chromium.launch(headless=True); b.close(); p.stop()" 2>/dev/null; then
        log_success "chromium 浏览器可用"
    else
        log_error "chromium 浏览器不可用，需要安装"
        if [ "$AUTO_FIX" = true ]; then
            log_info "正在安装 Playwright 浏览器..."
            python3 -m playwright install chromium
        fi
    fi
    
    # 检查系统 Chrome（可选）
    if [ -f "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" ]; then
        CHROME_VERSION=$("/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" --version 2>/dev/null || echo "unknown")
        log_success "系统 Chrome 已安装: $CHROME_VERSION"
    else
        log_warn "系统 Chrome 未安装（可选）"
    fi
}

# ==================== 微信开发者工具检查 ====================
check_wechat_devtools() {
    log_section "微信开发者工具检查"
    
    # macOS 默认路径
    DEVTOOLS_PATH="/Applications/wechatwebdevtools.app"
    CLI_PATH="$DEVTOOLS_PATH/Contents/MacOS/cli"
    
    if [ -d "$DEVTOOLS_PATH" ]; then
        log_success "微信开发者工具已安装"
        
        # 检查 CLI 工具
        if [ -f "$CLI_PATH" ]; then
            log_success "CLI 工具可用: $CLI_PATH"
            
            # 检查服务端口
            log_info "检查开发者工具服务端口..."
            
            # 尝试调用 CLI
            if "$CLI_PATH" islogin &>/dev/null; then
                log_success "开发者工具服务端口已开启"
            else
                log_warn "开发者工具服务端口可能未开启"
                log_info "请在开发者工具中: 设置 → 安全 → 开启服务端口"
            fi
        else
            log_warn "CLI 工具不存在"
        fi
        
        # 检查是否正在运行
        if pgrep -f "wechatwebdevtools" > /dev/null; then
            log_success "开发者工具正在运行"
        else
            log_warn "开发者工具未运行，建议先启动"
        fi
    else
        log_error "微信开发者工具未安装"
        log_info "请从 https://developers.weixin.qq.com/miniprogram/dev/devtools/download.html 下载"
    fi
    
    # 检查项目编译产物
    log_info "检查小程序编译产物..."
    DIST_PATH="$PROJECT_ROOT/src/frontend/dist"
    
    if [ -d "$DIST_PATH" ] && [ -f "$DIST_PATH/app.json" ]; then
        log_success "小程序编译产物存在: $DIST_PATH"
    else
        log_warn "小程序编译产物不存在"
        if [ "$AUTO_FIX" = true ]; then
            log_info "正在编译小程序..."
            cd "$PROJECT_ROOT/src/frontend"
            npm run build:weapp
        else
            log_info "请执行: cd src/frontend && npm run build:weapp"
        fi
    fi
}

# ==================== 服务检查 ====================
check_services() {
    log_section "服务端口检查"
    
    # 检查后端服务
    log_info "检查后端服务 (端口 8080)..."
    if curl -sf http://localhost:8080/api/system/health &>/dev/null; then
        log_success "后端服务正常 (http://localhost:8080)"
    else
        log_warn "后端服务未启动或不可访问"
        log_info "启动命令: cd src/backend && go run cmd/server/main.go"
    fi
    
    # 检查 Knowledge 服务
    log_info "检查 Knowledge 服务 (端口 8100)..."
    if curl -sf http://localhost:8100/api/v1/health &>/dev/null; then
        log_success "Knowledge 服务正常 (http://localhost:8100)"
    else
        log_warn "Knowledge 服务未启动或不可访问"
        log_info "启动命令: cd src/knowledge-service && python main.py"
    fi
    
    # 检查 H5 服务
    log_info "检查 H5 管理端服务 (端口 5173)..."
    if curl -sf http://localhost:5173 &>/dev/null; then
        log_success "H5 管理端服务正常 (http://localhost:5173)"
    else
        log_warn "H5 管理端服务未启动"
        log_info "启动命令: cd src/h5-admin && npm run dev"
    fi
    
    log_info "检查 H5 教师端服务 (端口 5174)..."
    if curl -sf http://localhost:5174 &>/dev/null; then
        log_success "H5 教师端服务正常 (http://localhost:5174)"
    else
        log_warn "H5 教师端服务未启动"
        log_info "启动命令: cd src/h5-teacher && npm run dev"
    fi
    
    log_info "检查 H5 学生端服务 (端口 5175)..."
    if curl -sf http://localhost:5175 &>/dev/null; then
        log_success "H5 学生端服务正常 (http://localhost:5175)"
    else
        log_warn "H5 学生端服务未启动"
        log_info "启动命令: cd src/h5-student && npm run dev"
    fi
}

# ==================== Node.js 环境检查 ====================
check_nodejs() {
    log_section "Node.js 环境检查"
    
    # 检查 Node.js
    if command -v node &> /dev/null; then
        NODE_VERSION=$(node --version)
        log_success "Node.js 已安装: $NODE_VERSION"
    else
        log_error "Node.js 未安装"
        log_info "建议安装 Node.js >= 16"
        return 1
    fi
    
    # 检查 npm
    if command -v npm &> /dev/null; then
        NPM_VERSION=$(npm --version)
        log_success "npm 已安装: $NPM_VERSION"
    else
        log_error "npm 未安装"
        return 1
    fi
    
    # 检查前端依赖
    log_info "检查前端依赖..."
    FRONTEND_DIR="$PROJECT_ROOT/src/frontend"
    
    if [ -d "$FRONTEND_DIR/node_modules" ]; then
        log_success "前端依赖已安装"
        
        # 检查 miniprogram-automator
        if [ -d "$FRONTEND_DIR/node_modules/miniprogram-automator" ]; then
            log_success "miniprogram-automator 已安装"
        else
            log_warn "miniprogram-automator 未安装"
        fi
    else
        log_warn "前端依赖未安装"
        if [ "$AUTO_FIX" = true ]; then
            log_info "正在安装前端依赖..."
            cd "$FRONTEND_DIR"
            npm install
        else
            log_info "请执行: cd src/frontend && npm install"
        fi
    fi
}

# ==================== 测试文件检查 ====================
check_test_files() {
    log_section "测试文件检查"
    
    # 检查 Minium 测试脚本
    MINIUM_TESTS=(
        "$PROJECT_ROOT/tests/e2e/smoke_v12_minium.py"
        "$PROJECT_ROOT/src/frontend/e2e/smoke_v12_minium.py"
    )
    
    for test_file in "${MINIUM_TESTS[@]}"; do
        if [ -f "$test_file" ]; then
            log_success "Minium 测试脚本存在: $(basename $test_file)"
        else
            log_warn "Minium 测试脚本不存在: $test_file"
        fi
    done
    
    # 检查 Playwright 测试脚本
    PLAYWRIGHT_TESTS=(
        "$PROJECT_ROOT/tests/e2e/smoke_playwright.py"
        "$PROJECT_ROOT/src/frontend/e2e/smoke_h5_e2e.py"
    )
    
    for test_file in "${PLAYWRIGHT_TESTS[@]}"; do
        if [ -f "$test_file" ]; then
            log_success "Playwright 测试脚本存在: $(basename $test_file)"
        else
            log_warn "Playwright 测试脚本不存在: $test_file"
        fi
    done
    
    # 检查测试配置
    if [ -f "$PROJECT_ROOT/docs/smoke-test-cases.md" ]; then
        log_success "测试用例文档存在"
    else
        log_warn "测试用例文档不存在"
    fi
}

# ==================== 环境变量检查 ====================
check_env_vars() {
    log_section "环境变量检查"
    
    ENV_FILE="$PROJECT_ROOT/.env"
    
    if [ -f "$ENV_FILE" ]; then
        log_success ".env 文件存在"
        
        # 检查关键变量
        source "$ENV_FILE" 2>/dev/null || true
        
        if [ -n "$JWT_SECRET" ] && [ "$JWT_SECRET" != "change-me-to-a-random-string" ]; then
            log_success "JWT_SECRET 已配置"
        else
            log_warn "JWT_SECRET 未正确配置"
        fi
        
        if [ -n "$OPENAI_API_KEY" ] && [ "$OPENAI_API_KEY" != "your-api-key" ]; then
            log_success "OPENAI_API_KEY 已配置"
        else
            log_warn "OPENAI_API_KEY 未正确配置"
        fi
    else
        log_warn ".env 文件不存在"
        log_info "请创建 .env 文件并配置必要的环境变量"
    fi
}

# ==================== 生成报告 ====================
generate_report() {
    log_section "环境检查报告"
    
    log ""
    log "检查时间: $(date '+%Y-%m-%d %H:%M:%S')"
    log "日志文件: $LOG_FILE"
    log ""
    log "检查结果统计:"
    log "  ✅ 通过: $PASS_COUNT"
    log "  ⚠️  警告: $WARN_COUNT"
    log "  ❌ 失败: $FAIL_COUNT"
    log ""
    
    log "详细结果:"
    for result in "${CHECK_RESULTS[@]}"; do
        log "  $result"
    done
    
    log ""
    
    if [ $FAIL_COUNT -eq 0 ]; then
        if [ $WARN_COUNT -eq 0 ]; then
            log "${GREEN}========================================${NC}"
            log "${GREEN}✅ 环境检查完全通过！${NC}"
            log "${GREEN}可以开始执行冒烟测试${NC}"
            log "${GREEN}========================================${NC}"
        else
            log "${YELLOW}========================================${NC}"
            log "${YELLOW}⚠️  环境检查通过，但有 $WARN_COUNT 个警告${NC}"
            log "${YELLOW}建议处理警告后再执行测试${NC}"
            log "${YELLOW}========================================${NC}"
        fi
        
        log ""
        log "执行冒烟测试命令:"
        log "  # Minium 测试（小程序）"
        log "  python3 tests/e2e/smoke_v12_minium.py"
        log ""
        log "  # Playwright 测试（H5）"
        log "  python3 tests/e2e/smoke_playwright.py"
        
        return 0
    else
        log "${RED}========================================${NC}"
        log "${RED}❌ 环境检查失败，发现 $FAIL_COUNT 个错误${NC}"
        log "${RED}请修复错误后再执行测试${NC}"
        log "${RED}========================================${NC}"
        
        log ""
        log "修复建议:"
        log "  1. 使用 --fix 参数自动修复: $0 --fix"
        log "  2. 手动安装缺失的依赖"
        log "  3. 启动必要的服务"
        log "  4. 检查日志文件获取详细信息: $LOG_FILE"
        
        return 1
    fi
}

# ==================== 主流程 ====================
main() {
    log "${CYAN}========================================${NC}"
    log "${CYAN}冒烟测试环境检查${NC}"
    log "${CYAN}========================================${NC}"
    log ""
    
    # 执行所有检查
    check_python
    check_minium
    check_playwright
    check_wechat_devtools
    check_nodejs
    check_services
    check_test_files
    check_env_vars
    
    # 生成报告
    generate_report
}

# 执行主流程
main
