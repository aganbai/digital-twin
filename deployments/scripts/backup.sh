#!/bin/bash
set -e

BACKUP_DIR="./backups/$(date +%Y%m%d_%H%M%S)"
mkdir -p "$BACKUP_DIR"

echo "=== Starting data backup ==="

echo "1. Backing up SQLite database..."
if [ -d "./data/sqlite" ]; then
    cp -r ./data/sqlite "$BACKUP_DIR/"
else
    echo "No SQLite data found, skipping..."
fi

echo "2. Backing up uploaded files..."
if [ -d "./uploads" ]; then
    cp -r ./uploads "$BACKUP_DIR/"
else
    echo "No uploads directory found, skipping..."
fi

echo "3. Backing up vector index..."
if [ -d "./data/vector_store" ]; then
    cp -r ./data/vector_store "$BACKUP_DIR/"
else
    echo "No vector store found, skipping..."
fi

echo "4. Creating compressed tarball..."
tar -czf "${BACKUP_DIR}.tar.gz" -C "./backups" "$(basename "$BACKUP_DIR")"
rm -rf "$BACKUP_DIR"

echo "=== Backup completed: ${BACKUP_DIR}.tar.gz ==="
