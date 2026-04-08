#!/bin/bash
# 一键部署到沙盒环境脚本
# 用法: ./deploy-sandbox.sh [OPTIONS]
# 支持：编译构建 -> 上传沙盒 -> 重启服务
# 可通过编排任务自动调用

set -e

# ==================== 配置区 ====================
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
COMPOSE_DIR="$SCRIPT_DIR/.."
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
LOG_FILE="/tmp/deploy-sandbox-${TIMESTAMP}.log"

# 默认配置（可通过参数覆盖）
SANDBOX_HOST="${SANDBOX_HOST:-}"
SANDBOX_USER="${SANDBOX_USER:-root}"
SANDBOX_PORT="${SANDBOX_PORT:-22}"
SANDBOX_DEPLOY_DIR="${SANDBOX_DEPLOY_DIR:-/opt/digital-twin}"
IMAGE_REGISTRY="${IMAGE_REGISTRY:-}"
PROJECT_NAME="digital-twin"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# ==================== 日志函数 ====================
log() {
    local level=$1
    shift
    local message="$@"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo -e "${timestamp} [${level}] ${message}" | tee -a "$LOG_FILE"
}

log_info() {
    echo -e "${BLUE}[INFO]${NC} $@" | tee -a "$LOG_FILE"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $@" | tee -a "$LOG_FILE"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $@" | tee -a "$LOG_FILE"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $@" | tee -a "$LOG_FILE"
}

# ==================== 帮助信息 ====================
show_help() {
    cat << EOF
数字分身项目一键部署脚本

用法: $0 [OPTIONS] [COMMAND]

命令:
  full            完整部署流程：构建 -> 上传 -> 重启 (默认)
  build           仅构建镜像
  upload          仅上传到沙盒
  restart         仅重启远程服务
  status          查看远程服务状态
  logs            查看远程服务日志
  rollback        回滚到上一版本

选项:
  --host HOST           沙盒服务器地址 (必填或设置 SANDBOX_HOST 环境变量)
  --user USER           SSH 用户名 (默认: root)
  --port PORT           SSH 端口 (默认: 22)
  --dir DIR             沙盒部署目录 (默认: /opt/digital-twin)
  --registry REGISTRY   镜像仓库地址 (可选，用于推送镜像)
  --env-file FILE       环境变量文件 (默认: deployments/.env.production)
  --skip-tests          跳过测试
  --dry-run             仅显示执行步骤，不实际执行
  -h, --help            显示帮助信息

环境变量:
  SANDBOX_HOST          沙盒服务器地址
  SANDBOX_USER          SSH 用户名
  SANDBOX_PORT          SSH 端口
  SANDBOX_DEPLOY_DIR    沙盒部署目录
  IMAGE_REGISTRY        镜像仓库地址

示例:
  # 完整部署
  $0 --host 192.168.1.100 full

  # 仅构建镜像
  $0 build

  # 仅重启服务
  $0 --host 192.168.1.100 restart

  # 使用环境变量
  export SANDBOX_HOST=192.168.1.100
  $0 full

返回码:
  0   成功
  1   参数错误
  2   构建失败
  3   上传失败
  4   服务启动失败
  5   健康检查失败

日志文件: $LOG_FILE
EOF
}

# ==================== 参数解析 ====================
parse_args() {
    COMMAND="full"
    SKIP_TESTS=false
    DRY_RUN=false
    ENV_FILE="$COMPOSE_DIR/.env.production"
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --host)
                SANDBOX_HOST="$2"
                shift 2
                ;;
            --user)
                SANDBOX_USER="$2"
                shift 2
                ;;
            --port)
                SANDBOX_PORT="$2"
                shift 2
                ;;
            --dir)
                SANDBOX_DEPLOY_DIR="$2"
                shift 2
                ;;
            --registry)
                IMAGE_REGISTRY="$2"
                shift 2
                ;;
            --env-file)
                ENV_FILE="$2"
                shift 2
                ;;
            --skip-tests)
                SKIP_TESTS=true
                shift
                ;;
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            full|build|upload|restart|status|logs|rollback)
                COMMAND="$1"
                shift
                ;;
            *)
                log_error "未知参数: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# ==================== 前置检查 ====================
check_dependencies() {
    log_info "检查依赖工具..."
    
    local missing_deps=()
    
    # 检查 Docker
    if ! command -v docker &> /dev/null; then
        missing_deps+=("docker")
    fi
    
    # 检查 Docker Compose
    if ! docker compose version &> /dev/null 2>&1; then
        missing_deps+=("docker-compose")
    fi
    
    # 如果需要远程部署，检查 SSH
    if [[ "$COMMAND" != "build" && -n "$SANDBOX_HOST" ]]; then
        if ! command -v ssh &> /dev/null; then
            missing_deps+=("openssh-client")
        fi
        if ! command -v scp &> /dev/null; then
            missing_deps+=("openssh-client")
        fi
    fi
    
    if [ ${#missing_deps[@]} -gt 0 ]; then
        log_error "缺少以下依赖工具: ${missing_deps[*]}"
        log_error "请先安装这些工具后再执行部署"
        exit 1
    fi
    
    log_success "依赖工具检查通过"
}

check_env_config() {
    log_info "检查环境配置..."
    
    if [ ! -f "$ENV_FILE" ]; then
        log_error "环境变量文件不存在: $ENV_FILE"
        log_info "请先创建环境变量文件，可参考 deployments/.env.production"
        exit 1
    fi
    
    # 加载环境变量
    set -a
    source "$ENV_FILE"
    set +a
    
    # 检查必要变量
    local missing_vars=()
    
    if [ -z "$JWT_SECRET" ] || [ "$JWT_SECRET" = "change-me-to-a-random-string" ]; then
        missing_vars+=("JWT_SECRET")
    fi
    
    if [ -z "$OPENAI_API_KEY" ] || [ "$OPENAI_API_KEY" = "your-api-key" ]; then
        missing_vars+=("OPENAI_API_KEY")
    fi
    
    if [ ${#missing_vars[@]} -gt 0 ]; then
        log_error "以下环境变量未正确配置: ${missing_vars[*]}"
        log_info "请编辑 $ENV_FILE 文件填写正确的配置"
        exit 1
    fi
    
    log_success "环境配置检查通过"
}

check_sandbox_connection() {
    if [ -z "$SANDBOX_HOST" ]; then
        log_error "未指定沙盒服务器地址"
        log_info "请通过 --host 参数或 SANDBOX_HOST 环境变量指定"
        exit 1
    fi
    
    log_info "测试沙盒服务器连接: $SANDBOX_HOST:$SANDBOX_PORT"
    
    if [ "$DRY_RUN" = true ]; then
        log_info "[DRY-RUN] 跳过连接测试"
        return 0
    fi
    
    # 测试 SSH 连接
    if ! ssh -o ConnectTimeout=10 -o StrictHostKeyChecking=no \
         -p "$SANDBOX_PORT" "${SANDBOX_USER}@${SANDBOX_HOST}" "echo 'Connection OK'" &> /dev/null; then
        log_error "无法连接到沙盒服务器: ${SANDBOX_USER}@${SANDBOX_HOST}:${SANDBOX_PORT}"
        log_error "请检查："
        log_error "  1. 服务器地址和端口是否正确"
        log_error "  2. SSH 密钥是否已配置"
        log_error "  3. 用户是否有登录权限"
        exit 1
    fi
    
    log_success "沙盒服务器连接正常"
}

# ==================== 构建阶段 ====================
build_images() {
    log_info "========== 开始构建镜像 =========="
    
    if [ "$DRY_RUN" = true ]; then
        log_info "[DRY-RUN] 将执行以下构建步骤:"
        log_info "  1. 构建后端镜像 (backend)"
        log_info "  2. 构建知识服务镜像 (knowledge)"
        log_info "  3. 拉取 Nginx 镜像"
        return 0
    fi
    
    cd "$COMPOSE_DIR"
    
    # 构建所有镜像
    log_info "构建 Docker 镜像..."
    if ! docker compose build --no-cache; then
        log_error "镜像构建失败"
        exit 2
    fi
    
    # 显示镜像信息
    log_info "构建完成的镜像:"
    docker images | grep "digital-twin" | head -5 | tee -a "$LOG_FILE"
    
    log_success "镜像构建完成"
}

run_tests() {
    if [ "$SKIP_TESTS" = true ]; then
        log_info "跳过测试 (--skip-tests)"
        return 0
    fi
    
    log_info "========== 运行测试 =========="
    
    if [ "$DRY_RUN" = true ]; then
        log_info "[DRY-RUN] 将运行测试"
        return 0
    fi
    
    cd "$PROJECT_ROOT"
    
    # 运行后端测试
    if [ -d "backend" ]; then
        log_info "运行后端测试..."
        cd "$PROJECT_ROOT/backend"
        if command -v go &> /dev/null; then
            go test ./... -v 2>&1 | tee -a "$LOG_FILE" || log_warn "后端测试存在失败"
        fi
        cd "$PROJECT_ROOT"
    fi
    
    log_success "测试执行完成"
}

push_images() {
    if [ -z "$IMAGE_REGISTRY" ]; then
        log_info "未配置镜像仓库，跳过镜像推送"
        return 0
    fi
    
    log_info "========== 推送镜像到仓库 =========="
    
    if [ "$DRY_RUN" = true ]; then
        log_info "[DRY-RUN] 将推送镜像到: $IMAGE_REGISTRY"
        return 0
    fi
    
    # 标记并推送镜像
    local images=("digital-twin_backend" "digital-twin_knowledge")
    
    for img in "${images[@]}"; do
        log_info "推送镜像: $img"
        docker tag "$img:latest" "${IMAGE_REGISTRY}/${img}:latest"
        docker push "${IMAGE_REGISTRY}/${img}:latest"
    done
    
    log_success "镜像推送完成"
}

# ==================== 上传阶段 ====================
upload_to_sandbox() {
    log_info "========== 上传到沙盒 =========="
    
    if [ "$DRY_RUN" = true ]; then
        log_info "[DRY-RUN] 将上传以下内容到沙盒:"
        log_info "  目标: ${SANDBOX_USER}@${SANDBOX_HOST}:${SANDBOX_DEPLOY_DIR}"
        log_info "  文件: docker-compose.yml, Dockerfile, nginx.conf, .env.production"
        return 0
    fi
    
    # 创建远程目录
    log_info "创建远程部署目录..."
    ssh -p "$SANDBOX_PORT" "${SANDBOX_USER}@${SANDBOX_HOST}" \
        "mkdir -p ${SANDBOX_DEPLOY_DIR}/{deployments/{docker,nginx,scripts},data,uploads,configs}"
    
    # 上传配置文件
    log_info "上传配置文件..."
    scp -P "$SANDBOX_PORT" "$COMPOSE_DIR/docker-compose.yml" \
        "${SANDBOX_USER}@${SANDBOX_HOST}:${SANDBOX_DEPLOY_DIR}/"
    
    scp -P "$SANDBOX_PORT" "$COMPOSE_DIR/.env.production" \
        "${SANDBOX_USER}@${SANDBOX_HOST}:${SANDBOX_DEPLOY_DIR}/"
    
    # 上传 Dockerfile
    scp -P "$SANDBOX_PORT" -r "$COMPOSE_DIR/docker/"* \
        "${SANDBOX_USER}@${SANDBOX_HOST}:${SANDBOX_DEPLOY_DIR}/deployments/docker/"
    
    # 上传 Nginx 配置
    scp -P "$SANDBOX_PORT" "$COMPOSE_DIR/nginx/nginx.conf" \
        "${SANDBOX_USER}@${SANDBOX_HOST}:${SANDBOX_DEPLOY_DIR}/deployments/nginx/"
    
    # 上传部署脚本
    scp -P "$SANDBOX_PORT" "$SCRIPT_DIR/deploy.sh" \
        "${SANDBOX_USER}@${SANDBOX_HOST}:${SANDBOX_DEPLOY_DIR}/deployments/scripts/"
    
    # 导出并上传镜像
    log_info "导出镜像文件..."
    local image_file="/tmp/digital-twin-images-${TIMESTAMP}.tar"
    
    docker save -o "$image_file" \
        digital-twin_backend:latest \
        digital-twin_knowledge:latest \
        nginx:alpine
    
    log_info "上传镜像文件到沙盒 (大小: $(du -h "$image_file" | cut -f1))..."
    scp -P "$SANDBOX_PORT" "$image_file" \
        "${SANDBOX_USER}@${SANDBOX_HOST}:${SANDBOX_DEPLOY_DIR}/images.tar"
    
    # 清理本地临时文件
    rm -f "$image_file"
    
    # 在远程加载镜像
    log_info "在沙盒服务器加载镜像..."
    ssh -p "$SANDBOX_PORT" "${SANDBOX_USER}@${SANDBOX_HOST}" \
        "cd ${SANDBOX_DEPLOY_DIR} && docker load -i images.tar && rm -f images.tar"
    
    log_success "上传完成"
}

# ==================== 重启服务阶段 ====================
restart_services() {
    log_info "========== 重启沙盒服务 =========="
    
    if [ "$DRY_RUN" = true ]; then
        log_info "[DRY-RUN] 将在沙盒服务器执行:"
        log_info "  1. 停止旧服务"
        log_info "  2. 启动新服务"
        log_info "  3. 等待服务就绪"
        return 0
    fi
    
    # 停止旧服务
    log_info "停止旧服务..."
    ssh -p "$SANDBOX_PORT" "${SANDBOX_USER}@${SANDBOX_HOST}" \
        "cd ${SANDBOX_DEPLOY_DIR} && docker compose down || true"
    
    # 启动新服务
    log_info "启动新服务..."
    if ! ssh -p "$SANDBOX_PORT" "${SANDBOX_USER}@${SANDBOX_HOST}" \
        "cd ${SANDBOX_DEPLOY_DIR} && docker compose up -d"; then
        log_error "服务启动失败"
        exit 4
    fi
    
    # 等待服务就绪
    log_info "等待服务就绪..."
    sleep 15
    
    # 健康检查
    health_check
    
    log_success "服务重启完成"
}

health_check() {
    log_info "========== 健康检查 =========="
    
    local max_retries=30
    local retry_interval=10
    local retry_count=0
    
    while [ $retry_count -lt $max_retries ]; do
        log_info "健康检查尝试 $((retry_count + 1))/$max_retries..."
        
        # 检查服务状态
        local backend_ok=false
        local knowledge_ok=false
        local nginx_ok=false
        
        # 检查 Backend
        if ssh -p "$SANDBOX_PORT" "${SANDBOX_USER}@${SANDBOX_HOST}" \
            "curl -sf http://localhost:8080/api/system/health" &> /dev/null; then
            backend_ok=true
        fi
        
        # 检查 Knowledge
        if ssh -p "$SANDBOX_PORT" "${SANDBOX_USER}@${SANDBOX_HOST}" \
            "curl -sf http://localhost:8100/api/v1/health" &> /dev/null; then
            knowledge_ok=true
        fi
        
        # 检查 Nginx
        if ssh -p "$SANDBOX_PORT" "${SANDBOX_USER}@${SANDBOX_HOST}" \
            "curl -sf http://localhost:80/health" &> /dev/null; then
            nginx_ok=true
        fi
        
        if [ "$backend_ok" = true ] && [ "$knowledge_ok" = true ] && [ "$nginx_ok" = true ]; then
            log_success "所有服务健康检查通过 ✓"
            log_info "  - Backend:  ✓"
            log_info "  - Knowledge: ✓"
            log_info "  - Nginx:    ✓"
            return 0
        fi
        
        log_warn "服务状态: Backend=${backend_ok} Knowledge=${knowledge_ok} Nginx=${nginx_ok}"
        
        retry_count=$((retry_count + 1))
        sleep $retry_interval
    done
    
    log_error "健康检查超时，服务可能未正常启动"
    
    # 输出日志帮助调试
    log_info "最近的容器日志:"
    ssh -p "$SANDBOX_PORT" "${SANDBOX_USER}@${SANDBOX_HOST}" \
        "cd ${SANDBOX_DEPLOY_DIR} && docker compose logs --tail=50"
    
    exit 5
}

# ==================== 查询命令 ====================
get_status() {
    log_info "查询服务状态..."
    ssh -p "$SANDBOX_PORT" "${SANDBOX_USER}@${SANDBOX_HOST}" \
        "cd ${SANDBOX_DEPLOY_DIR} && docker compose ps && echo '' && \
         echo '健康检查:' && \
         curl -sf http://localhost:8080/api/system/health && echo ' ✅ Backend' || echo ' ❌ Backend' && \
         curl -sf http://localhost:8100/api/v1/health && echo ' ✅ Knowledge' || echo ' ❌ Knowledge' && \
         curl -sf http://localhost:80/health && echo ' ✅ Nginx' || echo ' ❌ Nginx'"
}

get_logs() {
    log_info "获取服务日志..."
    ssh -p "$SANDBOX_PORT" "${SANDBOX_USER}@${SANDBOX_HOST}" \
        "cd ${SANDBOX_DEPLOY_DIR} && docker compose logs -f --tail=100"
}

rollback() {
    log_info "回滚到上一版本..."
    ssh -p "$SANDBOX_PORT" "${SANDBOX_USER}@${SANDBOX_HOST}" \
        "cd ${SANDBOX_DEPLOY_DIR} && docker compose down && \
         if [ -f images.backup.tar ]; then \
             docker load -i images.backup.tar && \
             docker compose up -d; \
         else \
             echo '未找到备份镜像'; \
             exit 1; \
         fi"
}

# ==================== 主流程 ====================
main() {
    log_info "============================================"
    log_info "数字分身项目一键部署脚本"
    log_info "执行时间: $(date '+%Y-%m-%d %H:%M:%S')"
    log_info "============================================"
    log_info ""
    
    parse_args "$@"
    
    # 根据命令执行不同流程
    case "$COMMAND" in
        build)
            check_dependencies
            check_env_config
            build_images
            run_tests
            push_images
            ;;
        upload)
            check_dependencies
            check_sandbox_connection
            upload_to_sandbox
            ;;
        restart)
            check_dependencies
            check_sandbox_connection
            restart_services
            ;;
        status)
            check_sandbox_connection
            get_status
            ;;
        logs)
            check_sandbox_connection
            get_logs
            ;;
        rollback)
            check_sandbox_connection
            rollback
            ;;
        full)
            check_dependencies
            check_env_config
            check_sandbox_connection
            build_images
            run_tests
            push_images
            upload_to_sandbox
            restart_services
            ;;
        *)
            log_error "未知命令: $COMMAND"
            show_help
            exit 1
            ;;
    esac
    
    log_info ""
    log_success "============================================"
    log_success "部署完成！"
    log_success "日志文件: $LOG_FILE"
    log_success "访问地址: http://${SANDBOX_HOST}"
    log_success "============================================"
    
    exit 0
}

# 执行主流程
main "$@"
