# 第一迭代 API 接口规范

## 1. 通用约定

### 1.1 基础信息

| 项目 | 说明 |
|------|------|
| Base URL | `http://localhost:8080` |
| 协议 | HTTP（开发环境），HTTPS（生产环境） |
| 数据格式 | JSON（`Content-Type: application/json`） |
| 字符编码 | UTF-8 |
| 认证方式 | Bearer Token（`Authorization: Bearer <token>`） |

### 1.2 统一响应格式

```json
{
  "code": 0,
  "message": "success",
  "data": { ... }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| code | int | 业务状态码，0 表示成功 |
| message | string | 状态描述 |
| data | object/null | 响应数据，错误时为 null |

### 1.3 错误码表

| 错误码 | HTTP Status | 说明 |
|--------|-------------|------|
| 0 | 200 | 成功 |
| 40001 | 401 | 未认证 / 令牌无效 |
| 40002 | 401 | 令牌已过期 |
| 40003 | 403 | 权限不足 |
| 40004 | 400 | 请求参数校验失败 |
| 40005 | 404 | 资源不存在 |
| 40006 | 409 | 用户名已存在 |
| 50001 | 500 | 服务器内部错误 |
| 50002 | 502 | 大模型调用失败 |
| 50003 | 502 | 向量数据库错误 |
| 50004 | 504 | 管道执行超时 |

### 1.4 分页约定

分页参数通过 Query String 传递：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| page | int | 1 | 页码，从 1 开始 |
| page_size | int | 20 | 每页数量，最大 100 |

分页响应格式：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [...],
    "total": 100,
    "page": 1,
    "page_size": 20
  }
}
```

---

## 2. 认证接口

### 2.1 用户注册

**POST** `/api/auth/register`

**鉴权**：无

**请求体**：
```json
{
  "username": "teacher_wang",
  "password": "123456",
  "role": "teacher",
  "nickname": "王老师",
  "email": "wang@example.com"
}
```

| 字段 | 类型 | 必填 | 校验规则 | 说明 |
|------|------|------|----------|------|
| username | string | ✅ | 3-32字符，字母数字下划线 | 用户名 |
| password | string | ✅ | 6-64字符 | 密码（明文传输，服务端 bcrypt 加密） |
| role | string | ✅ | 枚举：teacher/student/admin | 用户角色 |
| nickname | string | ❌ | 最长64字符 | 昵称 |
| email | string | ❌ | 邮箱格式 | 邮箱 |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user_id": 1,
    "username": "teacher_wang",
    "role": "teacher",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

**错误响应**：
| 场景 | code | message |
|------|------|---------|
| 用户名已存在 | 40006 | username already exists |
| 参数校验失败 | 40004 | invalid username format |

---

### 2.2 用户登录

**POST** `/api/auth/login`

**鉴权**：无

**请求体**：
```json
{
  "username": "teacher_wang",
  "password": "123456"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | ✅ | 用户名 |
| password | string | ✅ | 密码 |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user_id": 1,
    "username": "teacher_wang",
    "role": "teacher",
    "nickname": "王老师",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2026-04-02T13:00:00Z"
  }
}
```

**错误响应**：
| 场景 | code | message |
|------|------|---------|
| 用户不存在 | 40005 | user not found |
| 密码错误 | 40001 | invalid credentials |

---

### 2.3 刷新令牌

**POST** `/api/auth/refresh`

**鉴权**：需要（Bearer Token）

**请求体**：无

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2026-04-03T13:00:00Z"
  }
}
```

**错误响应**：
| 场景 | code | message |
|------|------|---------|
| 令牌无效 | 40001 | invalid token |
| 令牌已过期 | 40002 | token expired |

---

## 3. 对话接口

### 3.1 发送对话消息

**POST** `/api/chat`

**鉴权**：需要（Bearer Token）

**说明**：走 `student_chat` 管道，依次执行认证→记忆→知识检索→对话生成。

**请求体**：
```json
{
  "message": "什么是牛顿第一定律?",
  "teacher_id": 1,
  "session_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| message | string | ✅ | 用户消息，最长 2000 字符 |
| teacher_id | int | ✅ | 目标教师ID（选择哪个数字分身） |
| session_id | string | ❌ | 会话ID，不传则自动生成新会话 |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "reply": "这是一个很好的问题！在我们讨论牛顿第一定律之前，你能先想一想：如果你在冰面上推一个冰球，它会一直滑下去吗？是什么让它最终停下来的呢？",
    "session_id": "550e8400-e29b-41d4-a716-446655440000",
    "conversation_id": 42,
    "token_usage": {
      "prompt_tokens": 850,
      "completion_tokens": 120,
      "total_tokens": 970
    },
    "pipeline_duration_ms": 1523
  }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| reply | string | AI 回复内容（苏格拉底式引导） |
| session_id | string | 会话ID（用于后续对话保持上下文） |
| conversation_id | int | 本轮对话记录ID |
| token_usage | object | Token 使用统计 |
| pipeline_duration_ms | int | 管道执行耗时（毫秒） |

**错误响应**：
| 场景 | code | message |
|------|------|---------|
| 教师不存在 | 40005 | teacher not found |
| 消息为空 | 40004 | message is required |
| 大模型调用失败 | 50002 | LLM service unavailable |
| 管道超时 | 50004 | pipeline execution timeout |

---

### 3.2 获取对话历史

**GET** `/api/conversations`

**鉴权**：需要（Bearer Token）

**Query 参数**：

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| teacher_id | int | ✅ | - | 教师ID |
| session_id | string | ❌ | - | 会话ID，不传则返回所有会话 |
| page | int | ❌ | 1 | 页码 |
| page_size | int | ❌ | 20 | 每页数量 |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 41,
        "session_id": "550e8400-e29b-41d4-a716-446655440000",
        "role": "user",
        "content": "什么是牛顿第一定律?",
        "created_at": "2026-04-01T10:30:00Z"
      },
      {
        "id": 42,
        "session_id": "550e8400-e29b-41d4-a716-446655440000",
        "role": "assistant",
        "content": "这是一个很好的问题！在我们讨论牛顿第一定律之前...",
        "created_at": "2026-04-01T10:30:02Z"
      }
    ],
    "total": 24,
    "page": 1,
    "page_size": 20
  }
}
```

---

## 4. 知识库接口

### 4.1 添加知识文档

**POST** `/api/documents`

**鉴权**：需要（Bearer Token，角色：teacher）

**请求体**：
```json
{
  "title": "牛顿运动定律",
  "content": "牛顿第一定律（惯性定律）：一切物体在没有受到外力作用的时候，总保持静止状态或匀速直线运动状态。这就是牛顿第一定律，也叫惯性定律。\n\n牛顿第二定律：物体加速度的大小跟作用力成正比，跟物体的质量成反比，且与物体质量的倒数成正比。\n\n牛顿第三定律：两个物体之间的作用力和反作用力总是大小相等，方向相反，作用在同一条直线上。",
  "tags": ["物理", "力学", "牛顿定律"]
}
```

| 字段 | 类型 | 必填 | 校验规则 | 说明 |
|------|------|------|----------|------|
| title | string | ✅ | 1-200字符 | 文档标题 |
| content | string | ✅ | 1-100000字符 | 文档内容（纯文本） |
| tags | []string | ❌ | 每个标签最长32字符 | 分类标签 |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "document_id": 5,
    "title": "牛顿运动定律",
    "chunks_count": 3,
    "status": "active"
  }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| document_id | int | 文档ID |
| chunks_count | int | 分块数量（文档被切分为多少个向量块） |
| status | string | 文档状态 |

**错误响应**：
| 场景 | code | message |
|------|------|---------|
| 非教师角色 | 40003 | permission denied: teacher role required |
| 标题为空 | 40004 | title is required |
| 向量化失败 | 50003 | vector database error |

---

### 4.2 获取文档列表

**GET** `/api/documents`

**鉴权**：需要（Bearer Token，角色：teacher）

**Query 参数**：

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| status | string | ❌ | active | 文档状态筛选 |
| tag | string | ❌ | - | 按标签筛选 |
| page | int | ❌ | 1 | 页码 |
| page_size | int | ❌ | 20 | 每页数量 |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 5,
        "title": "牛顿运动定律",
        "doc_type": "text",
        "tags": ["物理", "力学", "牛顿定律"],
        "status": "active",
        "created_at": "2026-04-01T09:00:00Z",
        "updated_at": "2026-04-01T09:00:00Z"
      }
    ],
    "total": 12,
    "page": 1,
    "page_size": 20
  }
}
```

---

### 4.3 删除文档

**DELETE** `/api/documents/:id`

**鉴权**：需要（Bearer Token，角色：teacher）

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 文档ID |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "document_id": 5,
    "deleted": true
  }
}
```

**错误响应**：
| 场景 | code | message |
|------|------|---------|
| 文档不存在 | 40005 | document not found |
| 非文档所有者 | 40003 | permission denied: not document owner |

---

## 5. 记忆接口

### 5.1 获取学生记忆列表

**GET** `/api/memories`

**鉴权**：需要（Bearer Token）

**Query 参数**：

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| teacher_id | int | ✅ | - | 教师ID |
| memory_type | string | ❌ | - | 记忆类型筛选（conversation/learning_progress/personality_traits） |
| page | int | ❌ | 1 | 页码 |
| page_size | int | ❌ | 20 | 每页数量 |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 1,
        "memory_type": "conversation",
        "content": "学生对牛顿第一定律有基本了解，但对惯性概念理解不够深入",
        "importance": 0.8,
        "last_accessed": "2026-04-01T10:30:00Z",
        "created_at": "2026-04-01T10:30:00Z"
      },
      {
        "id": 2,
        "memory_type": "learning_progress",
        "content": "物理-力学-牛顿定律：掌握程度60%",
        "importance": 0.9,
        "last_accessed": "2026-04-01T10:30:00Z",
        "created_at": "2026-04-01T10:25:00Z"
      }
    ],
    "total": 5,
    "page": 1,
    "page_size": 20
  }
}
```

---

## 6. 系统接口

### 6.1 健康检查

**GET** `/api/system/health`

**鉴权**：无

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "status": "running",
    "timestamp": "2026-04-01T10:00:00Z",
    "uptime_seconds": 3600,
    "plugins": {
      "total": 4,
      "healthy": 4,
      "details": {
        "authentication": "healthy",
        "memory-management": "healthy",
        "knowledge-retrieval": "healthy",
        "socratic-dialogue": "healthy"
      }
    },
    "pipelines": {
      "total": 2,
      "names": ["student_chat", "teacher_management"]
    },
    "database": "connected",
    "version": "1.1.0"
  }
}
```

---

### 6.2 插件列表

**GET** `/api/system/plugins`

**鉴权**：需要（Bearer Token，角色：admin）

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "plugins": [
      {
        "name": "authentication",
        "version": "1.0.0",
        "type": "auth",
        "enabled": true,
        "priority": 5,
        "status": "running"
      },
      {
        "name": "memory-management",
        "version": "1.0.0",
        "type": "memory",
        "enabled": true,
        "priority": 30,
        "status": "running"
      },
      {
        "name": "knowledge-retrieval",
        "version": "1.0.0",
        "type": "knowledge",
        "enabled": true,
        "priority": 10,
        "status": "running"
      },
      {
        "name": "socratic-dialogue",
        "version": "1.0.0",
        "type": "dialogue",
        "enabled": true,
        "priority": 20,
        "status": "running"
      }
    ]
  }
}
```

---

### 6.3 管道列表

**GET** `/api/system/pipelines`

**鉴权**：需要（Bearer Token，角色：admin）

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "pipelines": [
      {
        "name": "student_chat",
        "description": "学生与数字分身对话流程",
        "plugins": ["authentication", "memory-management", "knowledge-retrieval", "socratic-dialogue"],
        "timeout": "30s"
      },
      {
        "name": "teacher_management",
        "description": "老师管理知识库流程",
        "plugins": ["authentication", "knowledge-retrieval"],
        "timeout": "30s"
      }
    ]
  }
}
```

---

## 7. JWT Token 规范

### 7.1 Token 结构

```json
{
  "header": {
    "alg": "HS256",
    "typ": "JWT"
  },
  "payload": {
    "user_id": 1,
    "username": "teacher_wang",
    "role": "teacher",
    "iss": "digital-twin",
    "exp": 1743588000,
    "iat": 1743501600
  }
}
```

### 7.2 Token 传递方式

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### 7.3 Token 有效期

| 类型 | 有效期 | 说明 |
|------|--------|------|
| Access Token | 24小时 | 用于 API 鉴权 |
| Refresh Token | 7天 | 用于刷新 Access Token（V2.0 实现） |

---

## 8. 插件间数据流转规范

### 8.1 PluginInput 数据结构

```go
type PluginInput struct {
    RequestID   string                 // 请求追踪ID（UUID）
    UserContext *UserContext            // 用户上下文（认证插件填充）
    Data        map[string]interface{} // 业务数据（插件间传递）
    Params      map[string]interface{} // 插件参数（不随管道传递）
    Context     context.Context        // Go context（超时控制）
}
```

### 8.2 管道中 Data 字段的演变

```
初始 Data (Handler 构建):
{
  "message": "什么是牛顿第一定律?",
  "teacher_id": 1,
  "session_id": "uuid",
  "action": "chat",
  "token": "eyJ..."
}

→ 经过 auth 插件后:
{
  ...原始字段,
  "authenticated": true,
  "user_id": 3,
  "role": "student"
}

→ 经过 memory 插件后:
{
  ...原始字段,
  "memories": [
    {"type": "conversation", "content": "学生对力学有基本了解", "importance": 0.8},
    {"type": "learning_progress", "content": "牛顿定律掌握60%", "importance": 0.9}
  ]
}

→ 经过 knowledge 插件后:
{
  ...原始字段,
  "chunks": [
    {"content": "牛顿第一定律：一切物体在没有受到外力...", "score": 0.92, "document_id": 5},
    {"content": "惯性是物体保持原来运动状态的性质...", "score": 0.85, "document_id": 5}
  ]
}

→ 经过 dialogue 插件后（最终输出）:
{
  "reply": "这是一个很好的问题！...",
  "conversation_id": 42,
  "token_usage": {"prompt_tokens": 850, "completion_tokens": 120, "total_tokens": 970}
}
```

---

## 9. Chroma DB 接口规范

### 9.1 集合命名

每个教师一个集合：`teacher_{teacher_id}_knowledge`

### 9.2 向量存储格式

```json
{
  "ids": ["doc_5_chunk_0", "doc_5_chunk_1", "doc_5_chunk_2"],
  "documents": ["牛顿第一定律...", "牛顿第二定律...", "牛顿第三定律..."],
  "metadatas": [
    {"document_id": 5, "teacher_id": 1, "title": "牛顿运动定律", "chunk_index": 0},
    {"document_id": 5, "teacher_id": 1, "title": "牛顿运动定律", "chunk_index": 1},
    {"document_id": 5, "teacher_id": 1, "title": "牛顿运动定律", "chunk_index": 2}
  ]
}
```

### 9.3 语义检索请求

```json
POST http://localhost:8000/api/v1/collections/{collection_id}/query
{
  "query_texts": ["什么是牛顿第一定律"],
  "n_results": 5,
  "where": {"teacher_id": 1}
}
```

---

## 10. 环境变量

| 变量名 | 必填 | 默认值 | 说明 |
|--------|------|--------|------|
| `JWT_SECRET` | ✅ | - | JWT 签名密钥（至少32字符） |
| `LLM_MODE` | ❌ | `mock` | 大模型模式：`api`(真实调用) / `mock`(预设回复) |
| `OPENAI_API_KEY` | 仅api模式 | - | 大模型 API 密钥（通义千问/DeepSeek等） |
| `OPENAI_BASE_URL` | ❌ | `https://dashscope.aliyuncs.com/compatible-mode/v1` | 大模型 API 地址（通义千问） |
| `LLM_MODEL` | ❌ | `qwen-turbo` | 模型名称（qwen-turbo免费） |
| `DB_PATH` | ❌ | `./data/digital-twin.db` | SQLite 数据库文件路径 |
| `CHROMA_HOST` | ❌ | `localhost` | Chroma DB 地址 |
| `CHROMA_PORT` | ❌ | `8000` | Chroma DB 端口 |
| `SERVER_PORT` | ❌ | `8080` | HTTP 服务端口 |
| `ENVIRONMENT` | ❌ | `development` | 运行环境 |

---

**文档版本**: v1.0.0  
**创建日期**: 2026-03-28  
**最后更新**: 2026-03-28
