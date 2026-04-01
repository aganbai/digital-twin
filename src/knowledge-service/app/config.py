"""配置模块：从环境变量读取配置"""

import os
from dotenv import load_dotenv

# 加载 .env 文件（从项目根目录）
load_dotenv(os.path.join(os.path.dirname(__file__), "..", "..", "..", ".env"))

# DashScope API Key（通义千问 Embedding）
DASHSCOPE_API_KEY = os.getenv("DASHSCOPE_API_KEY", os.getenv("OPENAI_API_KEY", ""))

# Embedding 模型名称
EMBEDDING_MODEL = os.getenv("EMBEDDING_MODEL", "text-embedding-v3")

# 向量存储持久化目录
DATA_DIR = os.getenv("KNOWLEDGE_DATA_DIR", os.path.join(os.path.dirname(__file__), "..", "data"))

# 服务端口
SERVICE_PORT = int(os.getenv("KNOWLEDGE_SERVICE_PORT", "8100"))

# 服务版本
SERVICE_VERSION = "1.0.0"
