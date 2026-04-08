# V1.0 第三迭代 API 接口规范

## 1. 概述

本文档描述 V1.0 迭代3 中涉及的接口变更。迭代3 主要是架构补全和质量加固，**不新增 API 接口**，但对以下现有接口的内部实现和返回格式进行了改造。

### 1.1 接口变更总览

| 接口 | 变更类型 | 说明 |
|------|----------|------|
| `POST /api/chat` | 内部改造 | 从手动编排改为管道编排，返回格式不变 |
| `POST /api/auth/refresh` | 行为增强 | 支持过期 token 在宽限期内刷新 |
| `GET /api/system/health` | 返回格式调整 | 对齐 V1.0 API 规范定义的完整格式 |

---

## 2. 接口详细规范

### 2.1 对话接口（内部改造）

#### POST `/api/chat`

**变更说明**：内部实现从手动编排 3 个插件改为调用 `ExecutePipeline("student_chat", input)`，走完整管道编排。**请求和响应格式不变**。

**鉴权**：需要（Bearer Token）

**请求体**（不变）：
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
| teacher_id | int | ✅ | 目标教师ID |
| session_id | string | ❌ | 会话ID，不传则自动生成 |

**成功响应**（不变）`200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "reply": "这是一个很好的问题！...",
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

**新增错误响应**：

| 场景 | code | message | HTTP Status |
|------|------|---------|-------------|
| 管道执行超时 | 50004 | pipeline execution timeout | 504 |
| 管道不存在 | 50001 | pipeline student_chat not found | 500 |

**内部管道执行流程**：
```
HandleChat 构建 PluginInput
    ↓
ExecutePipeline("student_chat", input)
    ↓
[authentication] → 透传模式（UserContext 已填充）
    ↓
[memory-management] → 自动检索记忆，注入 Data["memories"]
    ↓
[knowledge-retrieval] → 自动语义检索，注入 Data["chunks"]
    ↓
[socratic-dialogue] → 构建 prompt，调用 LLM，保存对话，提取记忆
    ↓
返回最终 PluginOutput
```

**副作用变更**：
- 🆕 对话完成后，对话插件会**异步**调用 LLM 提取记忆并存储到 `memories` 表
- 记忆提取失败不影响对话主流程的返回

---

### 2.2 令牌刷新接口（行为增强）

#### POST `/api/auth/refresh`

**变更说明**：支持过期 token 在宽限期（7天）内刷新。

**鉴权**：需要（Bearer Token，允许过期但在宽限期内的 token）

**请求体**：无

**成功响应**（不变）`200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user_id": 1,
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "role": "teacher",
    "expires_at": "2026-05-03T13:00:00Z"
  }
}
```

**刷新规则**：

| Token 状态 | 行为 | 响应 |
|------------|------|------|
| 未过期 | 正常刷新 | 200, code=0, 返回新 token |
| 过期 ≤ 7 天 | 宽限期刷新 | 200, code=0, 返回新 token |
| 过期 > 7 天 | 拒绝刷新 | 401, code=40002, "令牌已过期且超过刷新宽限期" |
| 签名无效 | 拒绝刷新 | 401, code=40001, "令牌无效，无法刷新" |

**错误响应**：

| 场景 | code | message | HTTP Status |
|------|------|---------|-------------|
| 缺少 token | 40001 | 缺少认证令牌 | 401 |
| 签名无效 | 40001 | 令牌无效，无法刷新 | 401 |
| 过期超过宽限期 | 40002 | 令牌已过期且超过刷新宽限期 | 401 |

---

### 2.3 健康检查接口（返回格式调整）

#### GET `/api/system/health`

**变更说明**：返回格式调整为与 V1.0 API 规范完全一致的结构。

**鉴权**：无

**成功响应** `200`：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "status": "running",
    "timestamp": "2026-05-01T10:00:00Z",
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

**字段说明**：

| 字段 | 类型 | 说明 |
|------|------|------|
| status | string | 管理器状态：`running` / `stopped` / `starting` |
| timestamp | string | 当前时间（RFC3339 格式） |
| uptime_seconds | int | 系统运行时长（秒） |
| plugins.total | int | 已注册插件总数 |
| plugins.healthy | int | 健康插件数量 |
| plugins.details | object | 每个插件的健康状态（`healthy` / 错误信息） |
| pipelines.total | int | 已注册管道总数 |
| pipelines.names | []string | 管道名称列表 |
| database | string | 数据库连接状态：`connected` / `disconnected` |
| version | string | 系统版本号（从配置读取） |

**与旧格式的差异**：

| 字段 | 旧格式 | 新格式 |
|------|--------|--------|
| plugins | `int`（数量） | `object`（含 total/healthy/details） |
| pipelines | `int`（数量） | `object`（含 total/names） |
| uptime_seconds | 无 | 🆕 新增 |
| database | 无 | 🆕 新增 |
| version | 无 | 🆕 新增 |
| metrics | `object` | 移除（不在 API 规范中） |
| plugin_health | `object` | 合并到 `plugins.details` |

---

## 3. 插件管道模式接口规范

迭代3 新增了插件在管道中的"管道模式"行为，以下是各插件在管道模式下的接口规范。

### 3.1 认证插件 - 透传模式

**触发条件**：`input.Data["action"]` 为空 且 `input.UserContext` 已填充（非 nil 且 UserID 非空）

**行为**：
- 不执行任何认证逻辑
- 直接将 `input.Data` 原样传递到 `output.Data`
- 返回 `Success: true`

**输入 Data**：原样透传  
**输出 Data**：与输入相同

---

### 3.2 知识库插件 - 管道模式

**触发条件**：`input.Data["action"]` 为空

**行为**：
- 从 `input.Data["message"]` 获取检索 query
- 从 `input.Data["teacher_id"]` 获取教师 ID
- 执行语义检索，返回 Top-K 相关文档片段
- 将结果注入到 `output.Data["chunks"]`

**输入 Data**：
```json
{
  "message": "什么是牛顿第一定律?",
  "teacher_id": 1,
  "memories": [...]  // 上游记忆插件注入
}
```

**输出 Data**：
```json
{
  "message": "什么是牛顿第一定律?",
  "teacher_id": 1,
  "memories": [...],
  "chunks": [
    {"content": "牛顿第一定律...", "score": 0.92, "document_id": 5, "title": "牛顿运动定律"},
    {"content": "惯性是物体...", "score": 0.85, "document_id": 5, "title": "牛顿运动定律"}
  ]
}
```

**容错**：检索失败时 `chunks` 为空数组，不阻断管道。

---

### 3.3 记忆插件 - 管道模式（已有，无变更）

**触发条件**：`input.Data["action"]` 为空

**行为**：
- 从 `input.UserContext.UserID` 获取学生 ID
- 从 `input.Data["teacher_id"]` 获取教师 ID
- 检索相关记忆，注入到 `output.Data["memories"]`

---

### 3.4 对话插件 - 管道模式增强

**触发条件**：`input.Data["action"] == "chat"`（由 Handler 在构建 input 时注入）

**行为变更**：
- 对话完成后，**异步**调用 LLM 提取记忆
- 提取的记忆存储到 `memories` 表
- 记忆提取失败不影响对话返回

**记忆提取流程**：
```
对话完成 → 保存对话记录
    ↓
启动 goroutine:
  → 构建记忆提取 prompt
  → 调用 LLM 提取记忆（JSON 格式）
  → 解析 LLM 返回的记忆数组
  → 逐条存储到 memories 表
  → 失败时仅记录日志，不影响主流程
```

---

## 4. 配置校验规则

迭代3 增强了配置文件的校验规则，以下是完整的校验清单：

### 4.1 系统配置校验

| 字段 | 规则 | 级别 |
|------|------|------|
| `system.name` | 不能为空 | ERROR |
| `system.version` | 不能为空 | ERROR |

### 4.2 插件配置校验

| 字段 | 规则 | 级别 |
|------|------|------|
| `plugins.{name}.type` | 不能为空 | ERROR |
| `authentication.config.jwt.secret` | 启用时不能为空或为 "default-secret-key" | ERROR |
| `socratic-dialogue.config.llm_provider.api_key` | 启用且 mode=api 时不能为空 | ERROR |

### 4.3 管道配置校验

| 字段 | 规则 | 级别 |
|------|------|------|
| `pipelines.{name}.plugins` | 不能为空数组 | ERROR |
| `pipelines.{name}.plugins[i]` | 引用的插件必须在 plugins 中定义且 enabled | ERROR |
| `pipelines.{name}.timeout` | 如果设置，必须是合法的 Go Duration 格式 | ERROR |

---

## 5. 环境变量（无新增）

本迭代不新增环境变量，沿用迭代1/2 的环境变量配置。

---

## 6. 错误码（新增）

| 错误码 | 说明 | HTTP Status | 新增/变更 |
|--------|------|-------------|-----------|
| 40002 | 令牌已过期且超过刷新宽限期 | 401 | 🔧 含义扩展 |
| 50004 | 管道执行超时 | 504 | 已有，本迭代实际启用 |

---

**文档版本**: v1.0.0  
**创建日期**: 2026-03-28  
**最后更新**: 2026-03-28
