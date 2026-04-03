#!/usr/bin/env python3
"""
批量文档导入入口脚本
用法: python3 import_documents.py --task-id <task_id> --input-dir <dir> --db-path <db> --persona-id <id> [--knowledge-base-id <id>]

流程: 扫描目录 → 逐文件解析 → 分块 → 向量化(可选) → 写入SQLite → 更新任务状态
"""

import argparse
import json
import logging
import os
import sys
import traceback

# 添加脚本目录到路径
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from processors.txt_processor import process_txt
from processors.md_processor import process_md
from processors.pdf_processor import process_pdf
from processors.docx_processor import process_docx
from utils.chunker import chunk_text
from utils.db_writer import insert_document_with_chunks, update_task_status
from utils.vectorizer import vectorize_document, is_vector_service_available

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format="[ImportDocuments] %(asctime)s %(levelname)s %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S"
)
logger = logging.getLogger(__name__)

# 文件处理器映射
PROCESSORS = {
    '.txt': process_txt,
    '.md': process_md,
    '.pdf': process_pdf,
    '.docx': process_docx,
}


def process_file(filepath, db_path, persona_id, knowledge_base_id, vector_available):
    """
    处理单个文件：解析 → 分块 → 向量化 → 写入DB

    Args:
        filepath: 文件路径
        db_path: 数据库路径
        persona_id: 分身ID
        knowledge_base_id: 知识库ID
        vector_available: 向量化服务是否可用

    Returns:
        dict: 处理结果
    """
    ext = os.path.splitext(filepath)[1].lower()
    filename = os.path.basename(filepath)

    processor = PROCESSORS.get(ext)
    if not processor:
        return {"filename": filename, "status": "failed", "error": f"不支持的文件格式: {ext}"}

    try:
        # 步骤1: 解析文件内容
        content = processor(filepath)
        if not content or not content.strip():
            return {"filename": filename, "status": "failed", "error": "文件内容为空"}

        logger.info(f"文件解析完成: {filename}, 内容长度: {len(content)} 字符")

        # 步骤2: 文本分块
        chunks = chunk_text(content, chunk_size=800, overlap=100)
        logger.info(f"文本分块完成: {filename}, 共 {len(chunks)} 块")

        # 步骤3: 写入数据库
        doc_type = ext.lstrip('.')
        db_result = insert_document_with_chunks(
            db_path, persona_id, filename, content, doc_type, chunks, knowledge_base_id
        )
        doc_id = db_result["document_id"]
        logger.info(f"数据库写入完成: {filename}, doc_id={doc_id}")

        # 步骤4: 向量化（可选，服务不可用时跳过）
        vector_result = None
        if vector_available and chunks:
            vector_result = vectorize_document(
                doc_id, filename, content, chunks, persona_id, knowledge_base_id
            )
            if vector_result["success"]:
                logger.info(f"向量化完成: {filename}, vectors={vector_result['vector_count']}")
            else:
                logger.warning(f"向量化失败（不影响导入）: {filename} - {vector_result['message']}")

        return {
            "filename": filename,
            "status": "success",
            "document_id": doc_id,
            "chunk_count": len(chunks),
            "vector_result": vector_result,
        }

    except Exception as e:
        logger.error(f"文件处理异常: {filename} - {traceback.format_exc()}")
        return {"filename": filename, "status": "failed", "error": str(e)}


def main():
    parser = argparse.ArgumentParser(description="批量文档导入")
    parser.add_argument("--task-id", required=True, help="任务ID")
    parser.add_argument("--input-dir", required=True, help="输入目录")
    parser.add_argument("--db-path", required=True, help="数据库路径")
    parser.add_argument("--persona-id", required=True, type=int, help="分身ID")
    parser.add_argument("--knowledge-base-id", type=int, default=0, help="知识库ID")
    args = parser.parse_args()

    logger.info(f"开始处理任务: {args.task_id}")
    logger.info(f"输入目录: {args.input_dir}")
    logger.info(f"数据库: {args.db_path}")
    logger.info(f"分身ID: {args.persona_id}, 知识库ID: {args.knowledge_base_id}")

    # 检查输入目录
    if not os.path.isdir(args.input_dir):
        logger.error(f"输入目录不存在: {args.input_dir}")
        update_task_status(args.db_path, args.task_id, "failed", 0, 0,
                           json.dumps({"error": "输入目录不存在"}, ensure_ascii=False))
        sys.exit(1)

    # 扫描目录中的文件
    files = [f for f in os.listdir(args.input_dir) if os.path.isfile(os.path.join(args.input_dir, f))]
    if not files:
        logger.error("没有找到任何文件")
        update_task_status(args.db_path, args.task_id, "failed", 0, 0,
                           json.dumps({"error": "没有找到任何文件"}, ensure_ascii=False))
        sys.exit(1)

    logger.info(f"发现 {len(files)} 个文件待处理")

    # 检查向量化服务可用性
    vector_available = is_vector_service_available()
    if vector_available:
        logger.info("向量化服务可用，将进行文档向量化")
    else:
        logger.warning("向量化服务不可用，跳过向量化步骤（文档仍会正常导入）")

    # 逐个处理文件（单文件失败不影响其他文件）
    results = []
    success_count = 0
    failed_count = 0

    for filename in files:
        filepath_full = os.path.join(args.input_dir, filename)
        logger.info(f"处理文件: {filename}")
        try:
            result = process_file(filepath_full, args.db_path, args.persona_id,
                                  args.knowledge_base_id, vector_available)
            results.append(result)
            if result["status"] == "success":
                success_count += 1
            else:
                failed_count += 1
                logger.warning(f"文件处理失败: {filename} - {result.get('error', '未知错误')}")
        except Exception as e:
            failed_count += 1
            results.append({"filename": filename, "status": "failed", "error": str(e)})
            logger.error(f"文件处理异常: {filename} - {traceback.format_exc()}")

    # 确定最终状态
    total = len(files)
    if success_count == total:
        status = "success"
    elif success_count == 0:
        status = "failed"
    else:
        status = "partial_success"

    result_json = json.dumps({"results": results}, ensure_ascii=False)
    update_task_status(args.db_path, args.task_id, status, success_count, failed_count, result_json)
    logger.info(f"任务完成: {status} (成功: {success_count}, 失败: {failed_count}, 总计: {total})")


if __name__ == "__main__":
    main()
