#!/bin/bash

#######################################################################
# 测试数据管理脚本
# 用途：为端到端冒烟测试创建和管理测试数据
# 使用方法：./scripts/manage_test_data.sh [command]
# 命令：
#   create  - 创建测试数据
#   verify  - 验证测试数据
#   clean   - 清理测试数据
#   reset   - 重置测试数据（清理后重新创建）
#######################################################################

set -e

# 配置
PROJECT_DIR="/Users/aganbai/Desktop/WorkSpace/digital-twin"
DB_PATH="$PROJECT_DIR/data/digital-twin.db"
SQL_SCRIPT="$PROJECT_DIR/scripts/insert_test_data.sql"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 创建测试数据
create_test_data() {
    log_info "开始创建测试数据..."
    
    if [ ! -f "$SQL_SCRIPT" ]; then
        log_error "SQL脚本不存在: $SQL_SCRIPT"
        exit 1
    fi
    
    sqlite3 "$DB_PATH" < "$SQL_SCRIPT"
    
    if [ $? -eq 0 ]; then
        log_success "测试数据创建完成"
    else
        log_error "测试数据创建失败"
        exit 1
    fi
}

# 验证测试数据
verify_test_data() {
    log_info "验证测试数据..."
    
    echo ""
    echo "=========================================="
    echo "测试数据验证报告"
    echo "=========================================="
    
    # 验证教师账号
    TEACHER_COUNT=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM users WHERE username = '13800138001';")
    if [ "$TEACHER_COUNT" -eq 1 ]; then
        log_success "教师账号存在"
    else
        log_error "教师账号不存在"
    fi
    
    # 验证学生账号
    STUDENT_COUNT=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM users WHERE username = '13900139001';")
    if [ "$STUDENT_COUNT" -eq 1 ]; then
        log_success "学生账号存在"
    else
        log_error "学生账号不存在"
    fi
    
    # 验证教师分身
    TEACHER_PERSONA=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM personas WHERE user_id = (SELECT id FROM users WHERE username = '13800138001');")
    if [ "$TEACHER_PERSONA" -ge 1 ]; then
        log_success "教师分身存在"
    else
        log_error "教师分身不存在"
    fi
    
    # 验证学生分身
    STUDENT_PERSONA=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM personas WHERE user_id = (SELECT id FROM users WHERE username = '13900139001');")
    if [ "$STUDENT_PERSONA" -ge 1 ]; then
        log_success "学生分身存在"
    else
        log_error "学生分身不存在"
    fi
    
    # 验证班级
    CLASS_COUNT=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM classes WHERE name = '冒烟测试班-自动化';")
    if [ "$CLASS_COUNT" -ge 1 ]; then
        log_success "测试班级存在"
    else
        log_error "测试班级不存在"
    fi
    
    # 验证师生关系
    RELATION_COUNT=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM teacher_student_relations WHERE teacher_id = (SELECT id FROM users WHERE username = '13800138001') AND student_id = (SELECT id FROM users WHERE username = '13900139001');")
    if [ "$RELATION_COUNT" -ge 1 ]; then
        log_success "师生关系已建立"
    else
        log_error "师生关系未建立"
    fi
    
    # 验证班级成员
    MEMBER_COUNT=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM class_members WHERE class_id = (SELECT id FROM classes WHERE name = '冒烟测试班-自动化');")
    if [ "$MEMBER_COUNT" -ge 1 ]; then
        log_success "学生已加入班级"
    else
        log_error "学生未加入班级"
    fi
    
    echo ""
    echo "=========================================="
    echo "详细信息"
    echo "=========================================="
    
    echo ""
    echo "教师信息："
    sqlite3 "$DB_PATH" "SELECT u.id as user_id, u.username, u.nickname, p.id as persona_id, p.nickname as persona_name FROM users u LEFT JOIN personas p ON p.user_id = u.id WHERE u.username = '13800138001';" | column -t -s '|'
    
    echo ""
    echo "学生信息："
    sqlite3 "$DB_PATH" "SELECT u.id as user_id, u.username, u.nickname, p.id as persona_id, p.nickname as persona_name FROM users u LEFT JOIN personas p ON p.user_id = u.id WHERE u.username = '13900139001';" | column -t -s '|'
    
    echo ""
    echo "班级信息："
    sqlite3 "$DB_PATH" "SELECT c.id, c.name, c.persona_id, p.nickname as teacher_name FROM classes c LEFT JOIN personas p ON c.persona_id = p.id WHERE c.name = '冒烟测试班-自动化';" | column -t -s '|'
    
    echo ""
    echo "师生关系："
    sqlite3 "$DB_PATH" "SELECT r.id, u1.username as teacher, u2.username as student, r.status FROM teacher_student_relations r LEFT JOIN users u1 ON r.teacher_id = u1.id LEFT JOIN users u2 ON r.student_id = u2.id WHERE u1.username = '13800138001' AND u2.username = '13900139001';" | column -t -s '|'
    
    echo ""
    log_success "验证完成"
}

# 清理测试数据
clean_test_data() {
    log_warning "开始清理测试数据..."
    
    sqlite3 "$DB_PATH" <<EOF
DELETE FROM class_members WHERE class_id IN (SELECT id FROM classes WHERE name = '冒烟测试班-自动化');
DELETE FROM classes WHERE name = '冒烟测试班-自动化';
DELETE FROM teacher_student_relations WHERE teacher_id IN (SELECT id FROM users WHERE username IN ('13800138001', '13900139001'));
DELETE FROM personas WHERE user_id IN (SELECT id FROM users WHERE username IN ('13800138001', '13900139001'));
DELETE FROM users WHERE username IN ('13800138001', '13900139001');
EOF
    
    log_success "测试数据清理完成"
}

# 重置测试数据
reset_test_data() {
    log_info "开始重置测试数据..."
    clean_test_data
    create_test_data
    verify_test_data
    log_success "测试数据重置完成"
}

# 主函数
main() {
    cd "$PROJECT_DIR"
    
    if [ ! -f "$DB_PATH" ]; then
        log_error "数据库文件不存在: $DB_PATH"
        exit 1
    fi
    
    case "${1:-}" in
        create)
            create_test_data
            ;;
        verify)
            verify_test_data
            ;;
        clean)
            clean_test_data
            ;;
        reset)
            reset_test_data
            ;;
        *)
            echo "使用方法: $0 {create|verify|clean|reset}"
            echo ""
            echo "命令说明："
            echo "  create  - 创建测试数据"
            echo "  verify  - 验证测试数据"
            echo "  clean   - 清理测试数据"
            echo "  reset   - 重置测试数据（清理后重新创建）"
            exit 1
            ;;
    esac
}

main "$@"
