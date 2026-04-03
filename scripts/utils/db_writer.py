"""
SQLite 数据库写入工具
负责将文档和分块数据写入 documents 表
"""

import sqlite3
from datetime import datetime


def get_teacher_id_by_persona(db_path, persona_id):
    """根据 persona_id 获取对应的 teacher user_id"""
    conn = sqlite3.connect(db_path)
    try:
        row = conn.execute("SELECT user_id FROM personas WHERE id = ?", (persona_id,)).fetchone()
        return row[0] if row else 0
    finally:
        conn.close()


def insert_document(db_path, persona_id, title, content, doc_type, knowledge_base_id=0, teacher_id=None):
    """
    插入文档到 documents 表

    Args:
        db_path: 数据库路径
        persona_id: 分身ID
        title: 文档标题
        content: 文档内容
        doc_type: 文档类型 (pdf/docx/txt/md)
        knowledge_base_id: 知识库ID (可选)
        teacher_id: 教师ID (可选，为None时自动查询)

    Returns:
        int: 新插入文档的ID
    """
    if teacher_id is None:
        teacher_id = get_teacher_id_by_persona(db_path, persona_id)

    conn = sqlite3.connect(db_path)
    try:
        now = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        cursor = conn.execute(
            """INSERT INTO documents (teacher_id, title, content, doc_type, tags, status, scope, persona_id, created_at, updated_at)
               VALUES (?, ?, ?, ?, '[]', 'active', 'global', ?, ?, ?)""",
            (teacher_id, title, content, doc_type, persona_id, now, now)
        )
        conn.commit()
        return cursor.lastrowid
    finally:
        conn.close()


def insert_document_with_chunks(db_path, persona_id, title, content, doc_type, chunks, knowledge_base_id=0):
    """
    插入文档及其分块到数据库

    Args:
        db_path: 数据库路径
        persona_id: 分身ID
        title: 文档标题
        content: 完整文档内容
        doc_type: 文档类型
        chunks: 分块列表 (来自 chunker.chunk_text)
        knowledge_base_id: 知识库ID

    Returns:
        dict: {"document_id": int, "chunk_count": int}
    """
    teacher_id = get_teacher_id_by_persona(db_path, persona_id)

    # 将分块内容拼接为摘要（取前3块的前100字符）
    summary_parts = []
    for chunk in chunks[:3]:
        text = chunk["content"][:100]
        summary_parts.append(text)
    summary = "...".join(summary_parts) if summary_parts else ""

    conn = sqlite3.connect(db_path)
    try:
        now = datetime.now().strftime("%Y-%m-%d %H:%M:%S")

        # 插入主文档记录
        cursor = conn.execute(
            """INSERT INTO documents (teacher_id, title, content, doc_type, tags, status, scope, persona_id, summary, created_at, updated_at)
               VALUES (?, ?, ?, ?, '[]', 'active', 'global', ?, ?, ?, ?)""",
            (teacher_id, title, content, doc_type, persona_id, summary, now, now)
        )
        doc_id = cursor.lastrowid
        conn.commit()

        return {"document_id": doc_id, "chunk_count": len(chunks)}
    finally:
        conn.close()


def update_task_status(db_path, task_id, status, success_files, failed_files, result_json_str):
    """
    更新批量任务状态

    Args:
        db_path: 数据库路径
        task_id: 任务ID
        status: 状态 (pending/processing/success/partial_success/failed)
        success_files: 成功文件数
        failed_files: 失败文件数
        result_json_str: 结果JSON字符串
    """
    conn = sqlite3.connect(db_path)
    try:
        now = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        conn.execute(
            "UPDATE batch_tasks SET status=?, success_files=?, failed_files=?, result_json=?, updated_at=? WHERE task_id=?",
            (status, success_files, failed_files, result_json_str, now, task_id)
        )
        conn.commit()
    finally:
        conn.close()
