#!/bin/bash

#######################################################################
# 快速验证测试数据脚本
# 用途：快速验证测试数据是否可用于API调用
# 执行时机：在运行集成测试前
#######################################################################

set -e

PROJECT_DIR="/Users/aganbai/Desktop/WorkSpace/digital-twin"
DB_PATH="$PROJECT_DIR/data/digital-twin.db"

echo "=========================================="
echo "测试数据快速验证"
echo "=========================================="
echo ""

# 检查数据库文件
if [ ! -f "$DB_PATH" ]; then
    echo "❌ 数据库文件不存在"
    exit 1
fi

echo "✅ 数据库文件存在"

# 验证关键数据
TEACHER_EXISTS=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM users WHERE username = '13800138001';")
STUDENT_EXISTS=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM users WHERE username = '13900139001';")
RELATION_EXISTS=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM teacher_student_relations WHERE teacher_id = (SELECT id FROM users WHERE username = '13800138001') AND student_id = (SELECT id FROM users WHERE username = '13900139001');")
CLASS_EXISTS=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM classes WHERE name = '冒烟测试班-自动化';")

if [ "$TEACHER_EXISTS" -eq 1 ]; then
    echo "✅ 教师账号就绪"
else
    echo "❌ 教师账号缺失"
    exit 1
fi

if [ "$STUDENT_EXISTS" -eq 1 ]; then
    echo "✅ 学生账号就绪"
else
    echo "❌ 学生账号缺失"
    exit 1
fi

if [ "$RELATION_EXISTS" -ge 1 ]; then
    echo "✅ 师生关系就绪"
else
    echo "❌ 师生关系缺失"
    exit 1
fi

if [ "$CLASS_EXISTS" -ge 1 ]; then
    echo "✅ 测试班级就绪"
else
    echo "❌ 测试班级缺失"
    exit 1
fi

echo ""
echo "=========================================="
echo "测试数据摘要"
echo "=========================================="
echo ""
echo "教师账号：13800138001 (密码: test123456)"
echo "学生账号：13900139001 (密码: test123456)"
echo ""
echo "教师分身ID：$(sqlite3 "$DB_PATH" "SELECT id FROM personas WHERE user_id = (SELECT id FROM users WHERE username = '13800138001') LIMIT 1;")"
echo "学生分身ID：$(sqlite3 "$DB_PATH" "SELECT id FROM personas WHERE user_id = (SELECT id FROM users WHERE username = '13900139001') LIMIT 1;")"
echo "班级ID：$(sqlite3 "$DB_PATH" "SELECT id FROM classes WHERE name = '冒烟测试班-自动化' LIMIT 1;")"
echo ""
echo "✅ 所有测试数据验证通过，可以开始测试！"
