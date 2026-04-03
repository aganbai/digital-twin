"""
向量化工具
调用 LlamaIndex 知识服务 HTTP API 存储文档向量
服务地址: http://localhost:8100
"""

import json
import logging
import time

try:
    import requests
    HAS_REQUESTS = True
except ImportError:
    HAS_REQUESTS = False

logger = logging.getLogger(__name__)

# LlamaIndex 知识服务地址
VECTOR_SERVICE_URL = "http://localhost:8100"
# 最大重试次数
MAX_RETRIES = 3
# 重试间隔（秒）
RETRY_DELAY = 2


def vectorize_document(doc_id, title, content, chunks, persona_id, knowledge_base_id=0):
    """
    调用 LlamaIndex 服务进行文档向量化

    Args:
        doc_id: 文档ID
        title: 文档标题
        content: 文档完整内容
        chunks: 分块列表
        persona_id: 分身ID
        knowledge_base_id: 知识库ID

    Returns:
        dict: {"success": bool, "message": str, "vector_count": int}
    """
    if not HAS_REQUESTS:
        logger.warning("requests 库未安装，跳过向量化")
        return {"success": False, "message": "requests 库未安装", "vector_count": 0}

    url = f"{VECTOR_SERVICE_URL}/vectors/documents"

    # 构建请求体
    documents = []
    for chunk in chunks:
        documents.append({
            "doc_id": doc_id,
            "title": title,
            "content": chunk["content"],
            "chunk_index": chunk["chunk_index"],
            "persona_id": persona_id,
            "knowledge_base_id": knowledge_base_id,
            "metadata": {
                "char_count": chunk["char_count"],
            }
        })

    payload = {
        "documents": documents,
        "persona_id": persona_id,
    }

    # 重试逻辑
    last_error = None
    for attempt in range(1, MAX_RETRIES + 1):
        try:
            logger.info(f"向量化请求 (尝试 {attempt}/{MAX_RETRIES}): doc_id={doc_id}, chunks={len(chunks)}")
            resp = requests.post(url, json=payload, timeout=60)

            if resp.status_code == 200:
                result = resp.json()
                return {
                    "success": True,
                    "message": "向量化成功",
                    "vector_count": len(chunks),
                }
            else:
                last_error = f"HTTP {resp.status_code}: {resp.text[:200]}"
                logger.warning(f"向量化请求失败 (尝试 {attempt}): {last_error}")

        except requests.exceptions.ConnectionError:
            last_error = "无法连接到向量化服务"
            logger.warning(f"向量化服务不可用 (尝试 {attempt}): {last_error}")
        except requests.exceptions.Timeout:
            last_error = "向量化请求超时"
            logger.warning(f"向量化请求超时 (尝试 {attempt})")
        except Exception as e:
            last_error = str(e)
            logger.warning(f"向量化异常 (尝试 {attempt}): {last_error}")

        if attempt < MAX_RETRIES:
            time.sleep(RETRY_DELAY * attempt)

    return {
        "success": False,
        "message": f"向量化失败（重试{MAX_RETRIES}次）: {last_error}",
        "vector_count": 0,
    }


def is_vector_service_available():
    """检查向量化服务是否可用"""
    if not HAS_REQUESTS:
        return False
    try:
        resp = requests.get(f"{VECTOR_SERVICE_URL}/health", timeout=5)
        return resp.status_code == 200
    except Exception:
        return False
