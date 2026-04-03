#!/bin/bash
# 数据备份脚本
set -e

BACKUP_DIR="${1:-./backups}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

mkdir -p "$BACKUP_DIR"

echo "📦 备份数据库..."
docker compose cp backend:/app/data/digital_twin.db "$BACKUP_DIR/digital_twin_${TIMESTAMP}.db"

echo "📦 备份向量索引..."
docker compose cp knowledge:/app/data "$BACKUP_DIR/knowledge_data_${TIMESTAMP}"

echo "📦 备份上传文件..."
docker compose cp backend:/app/uploads "$BACKUP_DIR/uploads_${TIMESTAMP}"

echo "✅ 备份完成: $BACKUP_DIR"
ls -la "$BACKUP_DIR"
