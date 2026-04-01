# V2.0 迭代5 API 接口规范

## 1. 通用约定

> 继承 V2.0 迭代4 的通用约定，以下仅列出新增和变更部分。

### 1.1 基础信息

| 项目 | 说明 |
|------|------|
| Go 后端 Base URL | `http://localhost:8080` |
| Python LlamaIndex 服务 Base URL | `http://localhost:8100`（内部调用，不对外暴露） |
| 协议 | HTTP（开发环境） |
| 数据格式 | JSON（`Content-Type: application/json`） |
| 认证方式 | Bearer Token（`Authorization: Bearer <token>`）— 仅 Go 后端接口 |

### 1.2 新增错误码

| 错误码 | HTTP Status | 说明 |
|--------|-------------|------|
| 40034 | 503 | 知识检索服务不可用（Python 服务未启动或连接失败） |
| 40035 | 400 | 不支持的文件类型 |
| 40036 | 400 | 文件大小超出限制 |

---

## 2. Python LlamaIndex 服务接口（内部调用）

> 以下接口由 Go 后端内部调用，不对微信小程序暴露。

### 2.1 存储文档向量

**POST** `/api/v1/vectors/documents`

**请求体**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| collection | string | ✅ | 集合名，格式 `teacher_{persona_id}`，如 `teacher_1` |
| doc_id | int64 | ✅ | 文档 ID |
| title | string | ✅ | 文档标题 |
| chunks | array | ✅ | 已分好的文本块数组 |
| chunks[].id | string | ✅ | 块 ID，格式 `doc_{doc_id}_chunk_{index}`，如 `doc_42_chunk_0` |
| chunks[].content | string | ✅ | 块文本内容 |
| chunks[].chunk_index | int | ✅ | 块序号（从 0 开始） |

**请求示例**：
```json
{
  "collection": "teacher_1",
  "doc_id": 42,
  "title": "二次方程教案",
  "chunks": [
    {"id": "doc_42_chunk_0", "content": "一元二次方程的一般形式为...", "chunk_index": 0},
    {"id": "doc_42_chunk_1", "content": "求根公式的推导过程...", "chunk_index": 1}
  ]
}
```

**成功响应** `200`：
```json
{
  "success": true,
  "chunks_count": 2
}
```

**失败响应** `500`：
```json
{
  "success": false,
  "error": "embedding API call failed: timeout"
}
```

---

### 2.2 语义检索

**POST** `/api/v1/vectors/search`

**请求体**：

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| collection | string | ✅ | - | 集合名 |
| query | string | ✅ | - | 检索查询文本 |
| top_k | int | ❌ | 5 | 返回数量（1~20） |

**请求示例**：
```json
{
  "collection": "teacher_1",
  "query": "怎么解二次方程",
  "top_k": 5
}
```

**成功响应** `200`：
```json
{
  "results": [
    {
      "content": "一元二次方程的求根公式为...",
      "score": 0.92,
      "doc_id": 42,
      "title": "二次方程教案",
      "chunk_id": "doc_42_chunk_0"
    },
    {
      "content": "求根公式的推导过程...",
      "score": 0.85,
      "doc_id": 42,
      "title": "二次方程教案",
      "chunk_id": "doc_42_chunk_1"
    }
  ]
}
```

**空结果响应** `200`：
```json
{
  "results": []
}
```

---

### 2.3 删除文档向量

**DELETE** `/api/v1/vectors/documents/{doc_id}`

**查询参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| collection | string | ✅ | 集合名 |

**请求示例**：
```
DELETE /api/v1/vectors/documents/42?collection=teacher_1
```

**成功响应** `200`：
```json
{
  "success": true,
  "deleted_chunks": 3
}
```

---

### 2.4 健康检查

**GET** `/api/v1/health`

**成功响应** `200`：
```json
{
  "status": "running",
  "version": "1.0.0",
  "index_count": 5
}
```

---

## 3. Go 后端新增接口

### 3.1 通用文件上传

**POST** `/api/upload`

**鉴权**：需要（Bearer Token）

**请求格式**：`multipart/form-data`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| file | file | ✅ | 上传的文件 |
| type | string | ❌ | 用途标识：`assignment`（作业）/ `document`（知识库）/ `general`（通用），默认 `general` |

**文件限制**：

| 限制项 | 值 |
|--------|-----|
| 最大文件大小 | 10MB |
| 支持的类型 | PDF、DOCX、TXT、MD、JPG、JPEG、PNG |
| 存储路径 | `uploads/{year}/{month}/{filename}` |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "url": "/uploads/2026/03/essay_abc123.pdf",
    "filename": "essay_abc123.pdf",
    "original_name": "我的作文.pdf",
    "size": 102400,
    "mime_type": "application/pdf"
  }
}
```

**失败响应**：

| 场景 | 错误码 | 说明 |
|------|--------|------|
| 文件类型不支持 | 40035 | 不支持的文件类型 |
| 文件过大 | 40036 | 文件大小超出限制（最大 10MB） |

---

## 4. Go 后端修改接口

### 4.1 对话接口新增附件支持

**POST** `/api/chat` 和 **POST** `/api/chat/stream`

**新增请求字段**（在原有基础上扩展）：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| attachment_url | string | ❌ | 附件文件路径（上传后返回的 url） |
| attachment_type | string | ❌ | 附件类型：`pdf` / `docx` / `txt` / `image` |
| attachment_name | string | ❌ | 附件原始文件名 |

**请求示例（带附件）**：
```json
{
  "message": "老师，帮我看看这篇作文",
  "teacher_persona_id": 1,
  "session_id": "sess_abc123",
  "attachment_url": "/uploads/2026/03/essay_abc123.pdf",
  "attachment_type": "pdf",
  "attachment_name": "我的作文.pdf"
}
```

**行为变化**：
1. 有附件时，后端使用已有的 `FileParser` 解析附件内容
2. 将附件内容拼接到 prompt 中（格式：`[附件: {filename}]\n{content}\n[/附件]`）
3. AI 自动识别为作业提交，给出点评回复
4. 同时在 `assignments` 表创建一条记录（自动关联 session_id）
5. 无附件时行为完全不变（向后兼容）

**响应格式不变**，与现有 chat/chat/stream 接口一致。

---

### 4.2 评语接口权限调整

**GET** `/api/comments`

**变更**：
- 学生角色调用时，返回空列表（`data: []`），不返回错误
- 教师角色调用时，行为不变

**学生调用响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": []
}
```

---

## 5. 接口变更汇总

### 5.1 新增接口（5 个）

| 编号 | 服务 | 接口 | 方法 | 说明 |
|------|------|------|------|------|
| API-52 | Python | `/api/v1/vectors/documents` | POST | 存储文档向量 |
| API-53 | Python | `/api/v1/vectors/search` | POST | 语义检索 |
| API-54 | Python | `/api/v1/vectors/documents/{doc_id}` | DELETE | 删除文档向量 |
| API-55 | Python | `/api/v1/health` | GET | 健康检查 |
| API-56 | Go | `/api/upload` | POST | 通用文件上传 |

### 5.2 修改接口（3 个）

| 编号 | 接口 | 变更 |
|------|------|------|
| API-M1 | `POST /api/chat` | 新增 attachment 参数 |
| API-M2 | `POST /api/chat/stream` | 新增 attachment 参数 |
| API-M3 | `GET /api/comments` | 学生角色返回空列表 |

### 5.3 内部改造（无接口变更）

| 改造点 | 说明 |
|--------|------|
| `knowledge_plugin.go` handleAdd | `vectorStore.AddDocuments` → HTTP POST Python 服务 |
| `knowledge_plugin.go` handleSearch | `vectorStore.Search` → HTTP POST Python 服务 |
| `knowledge_plugin.go` handlePipeline | `vectorStore.Search` → HTTP POST Python 服务 |
| `knowledge_plugin.go` handleDelete | `vectorStore.DeleteByDocumentID` → HTTP DELETE Python 服务 |
