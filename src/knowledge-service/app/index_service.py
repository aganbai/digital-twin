"""LlamaIndex 核心逻辑：向量索引管理"""

import os
import json
import logging
from typing import List, Dict, Any, Optional
from threading import Lock

from llama_index.core import VectorStoreIndex, StorageContext, Document
from llama_index.core.storage.docstore import SimpleDocumentStore
from llama_index.core.vector_stores import SimpleVectorStore
from llama_index.core.node_parser import SentenceSplitter
from llama_index.core.schema import TextNode
from llama_index.embeddings.dashscope import (
    DashScopeEmbedding,
    DashScopeTextEmbeddingModels,
    DashScopeTextEmbeddingType,
)

from . import config

logger = logging.getLogger(__name__)


class IndexService:
    """向量索引服务：管理多个 collection 的 VectorStoreIndex"""

    def __init__(self):
        self._indices: Dict[str, VectorStoreIndex] = {}
        self._vector_stores: Dict[str, SimpleVectorStore] = {}
        self._lock = Lock()
        self._embed_model = self._create_embed_model()
        self._query_embed_model = self._create_query_embed_model()
        # 启动时加载已持久化的索引
        self._load_existing_indices()

    def _create_embed_model(self) -> DashScopeEmbedding:
        """创建 DashScope Embedding 模型（文档类型，用于索引构建）"""
        return DashScopeEmbedding(
            model_name=DashScopeTextEmbeddingModels.TEXT_EMBEDDING_V3,
            api_key=config.DASHSCOPE_API_KEY,
            text_type=DashScopeTextEmbeddingType.TEXT_TYPE_DOCUMENT,
        )

    def _create_query_embed_model(self) -> DashScopeEmbedding:
        """创建 DashScope Embedding 模型（查询类型，用于语义检索）"""
        return DashScopeEmbedding(
            model_name=DashScopeTextEmbeddingModels.TEXT_EMBEDDING_V3,
            api_key=config.DASHSCOPE_API_KEY,
            text_type=DashScopeTextEmbeddingType.TEXT_TYPE_QUERY,
        )

    def _get_collection_dir(self, collection: str) -> str:
        """获取 collection 的持久化目录"""
        collection_dir = os.path.join(config.DATA_DIR, collection)
        os.makedirs(collection_dir, exist_ok=True)
        return collection_dir

    def _load_existing_indices(self):
        """启动时加载已持久化的所有索引"""
        if not os.path.exists(config.DATA_DIR):
            os.makedirs(config.DATA_DIR, exist_ok=True)
            return

        for name in os.listdir(config.DATA_DIR):
            collection_dir = os.path.join(config.DATA_DIR, name)
            vector_store_path = os.path.join(collection_dir, "vector_store.json")
            if os.path.isdir(collection_dir) and os.path.exists(vector_store_path):
                try:
                    self._load_index(name)
                    logger.info(f"已加载索引: {name}")
                except Exception as e:
                    logger.warning(f"加载索引 {name} 失败: {e}")

    def _load_index(self, collection: str):
        """从磁盘加载指定 collection 的索引"""
        collection_dir = self._get_collection_dir(collection)
        vector_store_path = os.path.join(collection_dir, "vector_store.json")

        if not os.path.exists(vector_store_path):
            return

        vector_store = SimpleVectorStore.from_persist_path(vector_store_path)
        storage_context = StorageContext.from_defaults(vector_store=vector_store)

        index = VectorStoreIndex(
            nodes=[],
            storage_context=storage_context,
            embed_model=self._embed_model,
        )

        self._vector_stores[collection] = vector_store
        self._indices[collection] = index

    def _get_or_create_index(self, collection: str) -> VectorStoreIndex:
        """获取或创建指定 collection 的索引"""
        if collection not in self._indices:
            collection_dir = self._get_collection_dir(collection)
            vector_store_path = os.path.join(collection_dir, "vector_store.json")

            if os.path.exists(vector_store_path):
                self._load_index(collection)
            else:
                vector_store = SimpleVectorStore()
                storage_context = StorageContext.from_defaults(vector_store=vector_store)
                index = VectorStoreIndex(
                    nodes=[],
                    storage_context=storage_context,
                    embed_model=self._embed_model,
                )
                self._vector_stores[collection] = vector_store
                self._indices[collection] = index

        return self._indices[collection]

    def _persist_index(self, collection: str):
        """持久化指定 collection 的向量存储"""
        if collection in self._vector_stores:
            collection_dir = self._get_collection_dir(collection)
            vector_store_path = os.path.join(collection_dir, "vector_store.json")
            self._vector_stores[collection].persist(persist_path=vector_store_path)

    def _load_metadata(self, collection: str) -> Dict[str, Any]:
        """加载 collection 的元数据（doc_id → chunk_node_ids 映射）"""
        collection_dir = self._get_collection_dir(collection)
        metadata_path = os.path.join(collection_dir, "metadata.json")
        if os.path.exists(metadata_path):
            with open(metadata_path, "r", encoding="utf-8") as f:
                return json.load(f)
        return {}

    def _save_metadata(self, collection: str, metadata: Dict[str, Any]):
        """保存 collection 的元数据"""
        collection_dir = self._get_collection_dir(collection)
        metadata_path = os.path.join(collection_dir, "metadata.json")
        with open(metadata_path, "w", encoding="utf-8") as f:
            json.dump(metadata, f, ensure_ascii=False, indent=2)

    def add_documents(
        self,
        collection: str,
        doc_id: int,
        title: str,
        chunks: List[Dict[str, Any]],
    ) -> int:
        """
        存储文档向量

        Args:
            collection: 集合名（如 teacher_1）
            doc_id: 文档 ID
            title: 文档标题
            chunks: 已分好的文本块列表

        Returns:
            成功存储的块数量
        """
        with self._lock:
            index = self._get_or_create_index(collection)

            # 构建 TextNode 列表
            nodes = []
            for chunk in chunks:
                node = TextNode(
                    text=chunk["content"],
                    id_=chunk["id"],
                    metadata={
                        "doc_id": doc_id,
                        "title": title,
                        "chunk_index": chunk["chunk_index"],
                        "chunk_id": chunk["id"],
                    },
                    excluded_embed_metadata_keys=["doc_id", "title", "chunk_index", "chunk_id"],
                    excluded_llm_metadata_keys=["doc_id", "title", "chunk_index", "chunk_id"],
                )
                nodes.append(node)

            # 插入节点到索引
            index.insert_nodes(nodes)

            # 更新元数据：记录 doc_id → node_ids 映射
            metadata = self._load_metadata(collection)
            doc_key = str(doc_id)
            if doc_key not in metadata:
                metadata[doc_key] = {"title": title, "node_ids": []}
            metadata[doc_key]["node_ids"].extend([n.id_ for n in nodes])
            self._save_metadata(collection, metadata)

            # 持久化
            self._persist_index(collection)

            return len(nodes)

    def search(
        self,
        collection: str,
        query: str,
        top_k: int = 5,
    ) -> List[Dict[str, Any]]:
        """
        语义检索

        Args:
            collection: 集合名
            query: 检索查询文本
            top_k: 返回数量

        Returns:
            检索结果列表
        """
        with self._lock:
            if collection not in self._indices:
                # 尝试从磁盘加载
                collection_dir = self._get_collection_dir(collection)
                vector_store_path = os.path.join(collection_dir, "vector_store.json")
                if os.path.exists(vector_store_path):
                    self._load_index(collection)
                else:
                    return []

            index = self._indices.get(collection)
            if index is None:
                return []

            # 复用预创建的查询类型 embedding 实例
            retriever = index.as_retriever(
                similarity_top_k=top_k,
                embed_model=self._query_embed_model,
            )
            nodes = retriever.retrieve(query)

        results = []
        for node in nodes:
            meta = node.metadata or {}
            results.append({
                "content": node.text,
                "score": round(node.score, 4) if node.score is not None else 0.0,
                "doc_id": meta.get("doc_id", 0),
                "title": meta.get("title", ""),
                "chunk_id": meta.get("chunk_id", node.id_),
            })

        return results

    def delete_by_document_id(self, collection: str, doc_id: int) -> int:
        """
        删除指定文档的所有向量

        Args:
            collection: 集合名
            doc_id: 文档 ID

        Returns:
            删除的块数量
        """
        with self._lock:
            if collection not in self._indices:
                return 0

            index = self._indices[collection]
            metadata = self._load_metadata(collection)
            doc_key = str(doc_id)

            if doc_key not in metadata:
                return 0

            node_ids = metadata[doc_key].get("node_ids", [])
            deleted_count = 0

            for node_id in node_ids:
                try:
                    index.delete_ref_doc(node_id, delete_from_docstore=True)
                    deleted_count += 1
                except Exception:
                    # 节点可能已不存在，忽略
                    pass

            # 更新元数据
            del metadata[doc_key]
            self._save_metadata(collection, metadata)

            # 持久化
            self._persist_index(collection)

            return deleted_count

    def get_index_count(self) -> int:
        """获取当前管理的索引数量"""
        return len(self._indices)
