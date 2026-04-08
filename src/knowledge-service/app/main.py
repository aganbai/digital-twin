"""FastAPI 入口：LlamaIndex 语义检索服务"""

import logging
from typing import List, Optional
from contextlib import asynccontextmanager

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field

from . import config
from .index_service import IndexService

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(name)s: %(message)s",
)
logger = logging.getLogger(__name__)

# 全局索引服务实例
index_service: Optional[IndexService] = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    """应用生命周期管理"""
    global index_service
    logger.info("正在初始化 LlamaIndex 索引服务...")
    index_service = IndexService()
    logger.info(f"索引服务已就绪，已加载 {index_service.get_index_count()} 个索引")
    yield
    logger.info("索引服务已关闭")


app = FastAPI(
    title="Knowledge Service",
    description="LlamaIndex 语义检索服务",
    version=config.SERVICE_VERSION,
    lifespan=lifespan,
)


# ======================== 请求/响应模型 ========================


class ChunkItem(BaseModel):
    """文本块"""
    id: str = Field(..., description="块 ID，格式 doc_{doc_id}_chunk_{index}")
    content: str = Field(..., description="块文本内容")
    chunk_index: int = Field(..., description="块序号（从 0 开始）")


class AddDocumentsRequest(BaseModel):
    """存储文档向量请求"""
    collection: str = Field(..., description="集合名，格式 teacher_{persona_id}")
    doc_id: int = Field(..., description="文档 ID")
    title: str = Field(..., description="文档标题")
    chunks: List[ChunkItem] = Field(..., description="已分好的文本块数组")


class AddDocumentsResponse(BaseModel):
    """存储文档向量响应"""
    success: bool
    chunks_count: int = 0
    error: Optional[str] = None


class SearchRequest(BaseModel):
    """语义检索请求"""
    collection: str = Field(..., description="集合名")
    query: str = Field(..., description="检索查询文本")
    top_k: int = Field(default=5, ge=1, le=100, description="返回数量（最大100）")


class SearchResultItem(BaseModel):
    """检索结果项"""
    content: str
    score: float
    doc_id: int
    title: str
    chunk_id: str


class SearchResponse(BaseModel):
    """语义检索响应"""
    results: List[SearchResultItem]


class DeleteResponse(BaseModel):
    """删除文档向量响应"""
    success: bool
    deleted_chunks: int = 0
    error: Optional[str] = None


class HealthResponse(BaseModel):
    """健康检查响应"""
    status: str
    version: str
    index_count: int


# ======================== API 接口 ========================


@app.post("/api/v1/vectors/documents", response_model=AddDocumentsResponse)
async def add_documents(req: AddDocumentsRequest):
    """存储文档向量：接收已分好的 chunks，做 embedding 后存入 SimpleVectorStore"""
    try:
        chunks_data = [
            {
                "id": chunk.id,
                "content": chunk.content,
                "chunk_index": chunk.chunk_index,
            }
            for chunk in req.chunks
        ]
        count = index_service.add_documents(
            collection=req.collection,
            doc_id=req.doc_id,
            title=req.title,
            chunks=chunks_data,
        )
        logger.info(f"文档 {req.doc_id} 已存储 {count} 个向量块到 {req.collection}")
        return AddDocumentsResponse(success=True, chunks_count=count)
    except Exception as e:
        logger.error(f"存储文档向量失败: {e}")
        return AddDocumentsResponse(success=False, error=str(e))


@app.post("/api/v1/vectors/search", response_model=SearchResponse)
async def search_vectors(req: SearchRequest):
    """语义检索：返回 top-k 相似文档块"""
    try:
        results = index_service.search(
            collection=req.collection,
            query=req.query,
            top_k=req.top_k,
        )
        return SearchResponse(
            results=[SearchResultItem(**r) for r in results]
        )
    except Exception as e:
        logger.error(f"语义检索失败: {e}")
        return SearchResponse(results=[])


@app.delete("/api/v1/vectors/documents/{doc_id}", response_model=DeleteResponse)
async def delete_document_vectors(doc_id: int, collection: str):
    """删除指定文档的所有向量"""
    try:
        deleted = index_service.delete_by_document_id(
            collection=collection,
            doc_id=doc_id,
        )
        logger.info(f"已删除文档 {doc_id} 的 {deleted} 个向量块（集合: {collection}）")
        return DeleteResponse(success=True, deleted_chunks=deleted)
    except Exception as e:
        logger.error(f"删除文档向量失败: {e}")
        return DeleteResponse(success=False, error=str(e))


@app.get("/api/v1/health", response_model=HealthResponse)
async def health_check():
    """健康检查"""
    return HealthResponse(
        status="running",
        version=config.SERVICE_VERSION,
        index_count=index_service.get_index_count() if index_service else 0,
    )


# ======================== 启动入口 ========================

if __name__ == "__main__":
    import uvicorn

    uvicorn.run(
        "app.main:app",
        host="0.0.0.0",
        port=config.SERVICE_PORT,
        reload=True,
    )
