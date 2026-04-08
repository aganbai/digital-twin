# V2.0 迭代6 API 接口规范

## 1. 通用约定

> 继承 V2.0 迭代5 的通用约定，以下仅列出新增和变更部分。

### 1.1 基础信息

| 项目 | 说明 |
|------|------|
| Go 后端 Base URL | `http://localhost:8080` |
| Python LlamaIndex 服务 Base URL | `http://localhost:8100`（内部调用，不对外暴露） |
| 协议 | HTTP（开发环境）/ HTTPS（生产环境，通过 Nginx） |
| 数据格式 | JSON（`Content-Type: application/json`） |
| 认证方式 | Bearer Token（`Authorization: Bearer <token>`） |

### 1.2 新增错误码

| 错误码 | HTTP Status | 说明 |
|--------|-------------|------|
| 40037 | 400 | 无效的聊天记录 JSON 格式 |
| 40038 | 400 | 聊天记录为空（解析后无有效对话） |
| 40039 | 403 | 无权操作该记忆（非该学生的教师） |
| 40040 | 400 | 无效的教学风格类型 |

---

## 2. 新增接口

### 2.1 记忆摘要合并

**POST** `/api/memories/summarize`

**鉴权**：需要（Bearer Token，仅教师角色）

**请求体**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| teacher_persona_id | int64 | ✅ | 教师分身 ID |
| student_persona_id | int64 | ✅ | 学生分身 ID |

**请求示例**：
```json
{
  "teacher_persona_id": 1,
  "student_persona_id": 5
}
```

**行为说明**：
1. 查询该学生的所有 `memory_layer = 'episodic'` 的记忆
2. 调用 LLM 将多条 episodic 记忆压缩为 1~3 条 core 记忆
3. 新生成的 core 记忆写入 memories 表（同 memory_type 只保留最近一条，有则 UPDATE）
4. 原始 episodic 记忆的 `memory_layer` 更新为 `archived`

**触发方式**：
- **手动触发**：教师通过此接口手动触发
- **定时任务**：后台每天凌晨 2:00 自动扫描所有学生的 episodic 记忆，超过 50 条的自动触发摘要合并

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "summarized_count": 8,
    "new_core_memories": [
      {
        "id": 101,
        "memory_type": "learning_progress",
        "memory_layer": "core",
        "content": "学生对数学函数有较好的理解，但在几何证明方面较薄弱",
        "importance": 0.9
      }
    ],
    "archived_count": 8
  }
}
```

**失败响应**：

| 场景 | 错误码 | 说明 |
|------|--------|------|
| 无 episodic 记忆 | 40038 | 没有可合并的情景记忆 |
| 非该学生的教师 | 40039 | 无权操作该记忆 |

---

### 2.2 编辑记忆

**PUT** `/api/memories/:id`

**鉴权**：需要（Bearer Token，仅教师角色）

**请求体**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| content | string | ✅ | 修改后的记忆内容 |
| importance | float64 | ❌ | 修改后的重要性（0.0~1.0） |
| memory_layer | string | ❌ | 修改记忆层级（core / episodic） |

**请求示例**：
```json
{
  "content": "学生对二次方程有较好的理解，能独立完成基础题",
  "importance": 0.85,
  "memory_layer": "core"
}
```

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 42,
    "memory_type": "learning_progress",
    "memory_layer": "core",
    "content": "学生对二次方程有较好的理解，能独立完成基础题",
    "importance": 0.85,
    "updated_at": "2026-03-31T10:00:00Z"
  }
}
```

**失败响应**：

| 场景 | 错误码 | 说明 |
|------|--------|------|
| 记忆不存在 | 40004 | 记忆不存在 |
| 非该学生的教师 | 40039 | 无权操作该记忆 |

---

### 2.3 删除记忆

**DELETE** `/api/memories/:id`

**鉴权**：需要（Bearer Token，仅教师角色）

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

**失败响应**：

| 场景 | 错误码 | 说明 |
|------|--------|------|
| 记忆不存在 | 40004 | 记忆不存在 |
| 非该学生的教师 | 40039 | 无权操作该记忆 |

---

### 2.4 聊天记录 JSON 导入知识库

**POST** `/api/documents/import-chat`

**鉴权**：需要（Bearer Token，仅教师角色）

**请求格式**：`multipart/form-data`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| file | file | ✅ | JSON 聊天记录文件 |
| title | string | ❌ | 文档标题（不传则自动生成） |
| tags | string | ❌ | 标签（JSON 数组格式，如 `["二叉树","数据结构"]`） |
| scope | string | ❌ | 范围：`global` / `class` / `student`，默认 `global` |
| scope_ids | string | ❌ | 范围 ID 列表（逗号分隔） |
| persona_id | int64 | ✅ | 教师分身 ID |

**文件限制**：

| 限制项 | 值 |
|--------|-----|
| 最大文件大小 | 5MB |
| 支持的类型 | JSON（`.json`） |

**支持的 JSON 格式**：

```json
// 格式1: OpenAI 风格（messages 数组）
{
  "messages": [
    {"role": "user", "content": "..."},
    {"role": "assistant", "content": "..."}
  ]
}

// 格式2: conversations 数组（带 sender/text）
{
  "conversations": [
    {"sender": "学生", "text": "..."},
    {"sender": "AI", "text": "..."}
  ]
}

// 格式3: 顶层数组
[
  {"role": "user", "content": "..."},
  {"role": "assistant", "content": "..."}
]
```

**解析规则**：
1. 自动识别 JSON 结构（`messages` / `conversations` / 顶层数组）
2. 提取对话对，拼接为结构化 Q&A 文本：
   ```
   Q: 什么是二叉树？
   A: 二叉树是一种树形数据结构，每个节点最多有两个子节点...

   Q: 二叉树有哪些遍历方式？
   A: 二叉树有三种基本遍历方式：前序遍历、中序遍历、后序遍历...
   ```
3. 调用现有 knowledge 插件的 add action 入库
4. `doc_type` 记为 `chat`

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "document_id": 42,
    "title": "关于二叉树的问答",
    "doc_type": "chat",
    "conversation_count": 12,
    "chunks_count": 5,
    "status": "active"
  }
}
```

**失败响应**：

| 场景 | 错误码 | 说明 |
|------|--------|------|
| 非教师角色 | 40003 | 权限不足 |
| JSON 格式无效 | 40037 | 无效的聊天记录 JSON 格式 |
| 解析后无有效对话 | 40038 | 聊天记录为空 |
| 文件过大 | 40036 | 文件大小超出限制（最大 5MB） |

---

## 3. 修改接口

### 3.1 记忆列表 - 新增 layer 筛选

**GET** `/api/memories`

**新增查询参数**：

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| layer | string | ❌ | - | 按记忆层级筛选：`core` / `episodic` / `archived`，不传则返回 core + episodic |
| page | int | ❌ | 1 | 页码（SQL 层分页） |
| page_size | int | ❌ | 20 | 每页数量（最大 50） |

**请求示例**：
```
GET /api/memories?teacher_persona_id=1&student_persona_id=5&layer=core&page=1&page_size=20
```

**响应变更**：每条记忆新增 `memory_layer` 字段

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 42,
      "memory_type": "learning_progress",
      "memory_layer": "core",
      "content": "学生对数学函数有较好的理解",
      "importance": 0.9,
      "created_at": "2026-03-20T10:00:00Z",
      "updated_at": "2026-03-31T10:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 15
  }
}
```

---

### 3.2 分享码信息 - 新增 join_status

**GET** `/api/shares/:code/info`

**响应变更**：新增 `join_status` 字段

| 新增字段 | 类型 | 说明 |
|----------|------|------|
| join_status | string | 加入状态：`can_join` / `already_joined` / `not_target` / `need_login` / `need_persona` |

**join_status 判断逻辑**：

| 条件 | join_status |
|------|-------------|
| 未登录（无 Bearer Token） | `need_login` |
| 已登录但无学生分身 | `need_persona` |
| 已是该教师的学生 | `already_joined` |
| 通用分享码（target=0） | `can_join` |
| 定向分享码 + 当前学生是目标 | `can_join` |
| 定向分享码 + 当前学生非目标 | `not_target` |

**响应示例（已登录，可加入）**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "teacher_persona_id": 1,
    "teacher_nickname": "张老师",
    "teacher_school": "XX中学",
    "teacher_description": "数学老师",
    "class_name": "高一1班",
    "target_student_persona_id": 0,
    "is_valid": true,
    "join_status": "can_join"
  }
}
```

**响应示例（非目标学生）**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "teacher_persona_id": 1,
    "teacher_nickname": "张老师",
    "teacher_school": "XX中学",
    "teacher_description": "数学老师",
    "class_name": "高一1班",
    "target_student_persona_id": 5,
    "target_student_nickname": "李四",
    "is_valid": true,
    "join_status": "not_target"
  }
}
```

---

### 3.3 分享码加入 - 非目标学生友好引导

**POST** `/api/shares/:code/join`

**行为变更**：非目标学生不再返回 `40029` 错误，改为返回引导信息。

**非目标学生响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "join_status": "not_target",
    "teacher_persona_id": 1,
    "teacher_nickname": "张老师",
    "message": "该邀请码是老师专门发给特定同学的，你可以向老师发起申请",
    "can_apply": true
  }
}
```

> 前端根据 `join_status = "not_target"` 和 `can_apply = true` 显示"向老师申请"按钮，调用 `POST /api/relations` 发起申请。

---

### 3.4 风格配置 - 新增 teaching_style

**PUT** `/api/styles` 和 **GET** `/api/styles`

**StyleConfig JSON 变更**：新增 `teaching_style` 字段

| 新增字段 | 类型 | 必填 | 默认值 | 说明 |
|----------|------|------|--------|------|
| teaching_style | string | ❌ | `socratic` | 教学风格模板 |

**可选值**：

| 值 | 名称 | 说明 |
|----|------|------|
| `socratic` | 苏格拉底式提问 | 不直接给答案，通过提问引导思考 |
| `explanatory` | 讲解式教学 | 详细讲解知识点，配合举例说明 |
| `encouraging` | 鼓励式教学 | 多用肯定语言，循序渐进引导 |
| `strict` | 严格式教学 | 严格要求，注重准确性和规范性 |
| `companion` | 陪伴式学习 | 像朋友一样陪伴学习，轻松氛围 |
| `custom` | 自定义 | 完全由 `style_prompt` 决定 |

**请求示例**：
```json
{
  "teacher_persona_id": 1,
  "student_persona_id": 5,
  "style_config": {
    "temperature": 0.7,
    "guidance_level": "medium",
    "teaching_style": "encouraging",
    "style_prompt": "",
    "max_turns_per_topic": 10
  }
}
```

**向后兼容**：`teaching_style` 为空或不传时，默认使用 `socratic`。

**失败响应**：

| 场景 | 错误码 | 说明 |
|------|--------|------|
| 无效的风格类型 | 40040 | teaching_style 值不在可选范围内 |

---

## 4. 接口变更汇总

### 4.1 新增接口（4 个）

| 编号 | 服务 | 接口 | 方法 | 说明 |
|------|------|------|------|------|
| API-57 | Go | `/api/memories/summarize` | POST | 记忆摘要合并 |
| API-58 | Go | `/api/memories/:id` | PUT | 编辑记忆 |
| API-59 | Go | `/api/memories/:id` | DELETE | 删除记忆 |
| API-60 | Go | `/api/documents/import-chat` | POST | 聊天记录导入知识库 |

### 4.2 修改接口（5 个）

| 编号 | 接口 | 变更 |
|------|------|------|
| API-M4 | `GET /api/memories` | 新增 `layer` / `page` / `page_size` 参数，响应新增 `memory_layer` 字段 + `pagination` |
| API-M5 | `GET /api/shares/:code/info` | 新增 `join_status` 字段 |
| API-M6 | `POST /api/shares/:code/join` | 非目标学生返回引导而非 40029 错误 |
| API-M7 | `PUT /api/styles` | StyleConfig 新增 `teaching_style` 字段 |
| API-M8 | `GET /api/styles` | StyleConfig 新增 `teaching_style` 字段 |

### 4.3 内部改造（无接口变更）

| 改造点 | 说明 |
|--------|------|
| `memory_plugin.go` handleStore | 新增 LLM 层级判断 + core 记忆更新覆盖 |
| `memory_plugin.go` handleList | SQL 层 WHERE + LIMIT/OFFSET + 索引 |
| `prompt.go` systemPromptTemplate | 改为 teachingStyleTemplates map，根据 teaching_style 动态选择 |
| `prompt.go` buildStyleText | guidance_level 描述去掉苏格拉底绑定 |

### 4.4 数据库变更

| 变更 | 说明 |
|------|------|
| `memories.memory_layer` | 新增字段，TEXT NOT NULL DEFAULT 'episodic' |
| `idx_memories_layer` | 新增索引 (teacher_persona_id, student_persona_id, memory_layer) |
| `idx_memories_type_layer` | 新增索引 (teacher_persona_id, student_persona_id, memory_type, memory_layer) |
| `documents.source_session_id` | 新增可选字段，TEXT DEFAULT '' |
